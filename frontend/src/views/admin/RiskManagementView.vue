<template>
  <AppLayout>
    <div class="space-y-6 pb-12">
      <section class="overflow-hidden rounded-[32px] border border-slate-200/80 bg-[radial-gradient(circle_at_top_left,rgba(14,165,233,0.15),transparent_42%),linear-gradient(135deg,#f8fafc,rgba(255,255,255,0.95))] p-6 shadow-card dark:border-dark-800 dark:bg-[radial-gradient(circle_at_top_left,rgba(34,211,238,0.14),transparent_34%),linear-gradient(135deg,rgba(15,23,42,0.92),rgba(2,6,23,0.92))] lg:p-8">
        <div class="grid gap-6 xl:grid-cols-[minmax(0,1.3fr)_minmax(360px,0.7fr)] xl:items-start">
          <div>
            <p class="text-xs font-semibold tracking-[0.18em] text-sky-600/80 dark:text-cyan-300/80">{{ t('admin.risk.visuals.heroEyebrow') }}</p>
            <h1 class="mt-3 max-w-3xl text-3xl font-semibold tracking-tight text-slate-950 dark:text-white sm:text-4xl">
              {{ t('admin.risk.title') }}
            </h1>
            <p class="mt-3 max-w-2xl text-sm leading-6 text-slate-600 dark:text-slate-300">
              {{ t('admin.risk.description') }}
            </p>

            <div class="mt-5 flex flex-wrap gap-2">
              <span class="rounded-full border border-sky-200/80 bg-white/80 px-3 py-1 text-xs font-medium text-slate-700 backdrop-blur dark:border-cyan-900/60 dark:bg-dark-900/70 dark:text-slate-200">
                {{ t('admin.risk.windowLabel') }}: 24h
              </span>
              <span class="rounded-full border border-sky-200/80 bg-white/80 px-3 py-1 text-xs font-medium text-slate-700 backdrop-blur dark:border-cyan-900/60 dark:bg-dark-900/70 dark:text-slate-200">
                {{ riskSettings.ip_intel_enabled ? t('admin.risk.ipIntelEnabled') : t('admin.risk.ipIntelDisabled') }}
              </span>
              <span class="rounded-full border border-sky-200/80 bg-white/80 px-3 py-1 text-xs font-medium text-slate-700 backdrop-blur dark:border-cyan-900/60 dark:bg-dark-900/70 dark:text-slate-200">
                {{ t('admin.risk.visuals.thresholdSummary', { review: riskSettings.review_threshold, throttle: riskSettings.throttle_threshold, freeze: riskSettings.freeze_threshold }) }}
              </span>
            </div>

            <div class="mt-6 flex flex-wrap gap-3">
              <button type="button" class="btn btn-primary px-4 py-2 text-sm" :disabled="loading" @click="loadPage">
                {{ loading ? t('common.loading') : t('common.refresh') }}
              </button>
              <button type="button" class="btn btn-secondary px-4 py-2 text-sm" @click="showSettingsDialog = true">
                {{ t('common.settings') }}
              </button>
            </div>
          </div>

          <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
            <button
              v-for="card in heroCards"
              :key="card.label"
              type="button"
              class="rounded-3xl border border-white/70 bg-white/80 p-4 text-left shadow-sm backdrop-blur transition-all dark:border-white/5 dark:bg-dark-900/70"
              :class="[
                card.clickable ? 'hover:-translate-y-0.5 hover:shadow-md' : 'cursor-default',
                focusMode === card.key ? 'ring-2 ring-primary-500/70 dark:ring-primary-400/70' : ''
              ]"
              @click="card.clickable ? applyHeroCardFilter(card.key as HeroFocusMode) : undefined"
            >
              <div class="flex items-start justify-between gap-3">
                <div class="flex items-center gap-2">
                  <span
                    v-if="card.status !== undefined"
                    class="h-2.5 w-2.5 rounded-full"
                    :class="card.status ? 'bg-emerald-500 shadow-[0_0_0_4px_rgba(16,185,129,0.16)]' : 'bg-red-500 shadow-[0_0_0_4px_rgba(239,68,68,0.12)]'"
                  ></span>
                  <p class="text-sm font-medium text-slate-600 dark:text-slate-300">{{ card.label }}</p>
                </div>
                <HelpTooltip :content="card.hint" />
              </div>
              <p class="mt-2 text-2xl font-semibold text-slate-950 dark:text-white" :class="card.emphasisClass">{{ card.value }}</p>
            </button>
          </div>
        </div>
      </section>

      <div class="grid gap-6 xl:grid-cols-2">
        <RiskDecisionBreakdownCard :rows="filteredRows" />
        <RiskScoreDistributionCard :rows="filteredRows" :settings="riskSettings" />
      </div>
      <div class="grid gap-6">
        <RiskActivityHeatmapCard :rows="filteredRows" />
      </div>
      <div class="grid gap-6 xl:grid-cols-2">
        <RiskLeaderboardCard :rows="orderedRows" :settings="riskSettings" @open-user="openRiskDetail" />
      </div>

      <section class="rounded-[28px] border border-gray-200/80 bg-white/95 p-4 shadow-card dark:border-dark-800 dark:bg-dark-900/90">
        <div class="grid gap-3 xl:grid-cols-[minmax(0,1fr)_180px_180px_180px_auto] xl:items-center">
          <div>
            <label class="sr-only" for="risk-search">{{ t('common.search') }}</label>
            <input
              id="risk-search"
              v-model="searchQuery"
              type="text"
              class="input"
              :placeholder="t('admin.risk.searchPlaceholder')"
              @input="handleSearch"
            />
          </div>
          <Select v-model="statusFilter" :options="statusOptions" @change="applyFilters" />
          <Select v-model="decisionFilter" :options="decisionOptions" @change="applyFilters" />
          <Select v-model="sortMode" :options="sortOptions" @change="applyFilters" />
          <div class="text-right text-xs text-gray-500 dark:text-gray-400">
            {{ t('admin.risk.visuals.showingUsers', { count: filteredRows.length, total: riskRows.length }) }}
          </div>
        </div>
      </section>

      <TablePageLayout>
        <template #table>
          <DataTable :columns="columns" :data="orderedRows" :loading="loading" :actions-count="4">
            <template #cell-user="{ row }">
              <button type="button" class="group text-left" @click="openRiskDetail(row.user)">
                <p class="font-medium text-slate-900 transition-colors group-hover:text-primary-600 dark:text-white dark:group-hover:text-primary-400">{{ row.user.email }}</p>
                <div class="mt-1 flex flex-wrap items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
                  <span>{{ row.user.username || '-' }}</span>
                  <span class="rounded-full px-2 py-0.5" :class="row.user.status === 'active' ? 'bg-emerald-50 text-emerald-600 dark:bg-emerald-900/20 dark:text-emerald-300' : 'bg-red-50 text-red-600 dark:bg-red-900/20 dark:text-red-300'">
                    {{ row.user.status }}
                  </span>
                </div>
              </button>
            </template>

            <template #cell-risk_score="{ row }">
              <div>
                <p class="text-lg font-semibold" :class="scoreClass(row.summary.risk_score)">{{ row.summary.risk_score }}</p>
                <p class="text-xs text-gray-400 dark:text-gray-500">{{ row.summary.computed_at ? formatDateTime(row.summary.computed_at) : '-' }}</p>
              </div>
            </template>

            <template #cell-decision="{ row }">
              <span :class="decisionClass(row.summary.decision)" class="inline-flex rounded-full px-2.5 py-1 text-xs font-semibold">
                {{ row.summary.decision_label }}
              </span>
            </template>

            <template #cell-actual_cost_24h="{ row }">
              <div>
                <p class="font-medium text-slate-900 dark:text-white">${{ row.metrics.actual_cost_24h.toFixed(4) }}</p>
                <p class="text-xs text-gray-400 dark:text-gray-500">{{ row.metrics.request_count_24h }} 次请求</p>
              </div>
            </template>

            <template #cell-historical_ip_count="{ row }">
              <div>
                <p class="font-medium text-slate-900 dark:text-white">{{ row.metrics.historical_ip_count }}</p>
                <p class="text-xs text-gray-400 dark:text-gray-500">{{ row.metrics.first_ip || '-' }}</p>
              </div>
            </template>

            <template #cell-ua_24h_count="{ row }">
              <div>
                <p class="font-medium text-slate-900 dark:text-white">{{ row.metrics.ua_24h_count }}</p>
                <p class="max-w-[220px] truncate text-xs text-gray-400 dark:text-gray-500">{{ row.top_user_agent || '-' }}</p>
              </div>
            </template>

            <template #cell-active_hours_count="{ row }">
              <div>
                <p class="font-medium text-slate-900 dark:text-white">{{ row.metrics.active_hours_count }}</p>
                <p class="text-xs text-gray-400 dark:text-gray-500">{{ row.metrics.all_day_active ? t('common.yes') : t('common.no') }}</p>
              </div>
            </template>

            <template #cell-longest_silence_hours="{ row }">
              <div>
                <p class="font-medium text-slate-900 dark:text-white">{{ row.metrics.longest_silence_hours.toFixed(2) }}h</p>
                <p class="text-xs text-gray-400 dark:text-gray-500">集中度 {{ row.metrics.hour_concentration.toFixed(3) }}</p>
              </div>
            </template>

            <template #cell-rule_hit_count="{ row }">
              <div>
                <p class="font-medium text-slate-900 dark:text-white">{{ row.rule_hit_count }}</p>
                <p class="text-xs text-gray-400 dark:text-gray-500">{{ row.metrics.concurrent_multi_ip_ua_minutes_24h }} 分钟并发</p>
              </div>
            </template>

            <template #cell-actions="{ row }">
              <div class="flex flex-wrap gap-2">
                <button type="button" class="btn btn-secondary px-3 py-1.5 text-xs" @click="openRiskDetail(row.user)">
                  {{ t('common.details') }}
                </button>
                <button
                  type="button"
                  class="btn btn-secondary px-3 py-1.5 text-xs"
                  :disabled="row.user.concurrency <= 1 || row.user.status !== 'active'"
                  @click="applyThrottle(row.user)"
                >
                  {{ t('admin.risk.actions.throttle') }}
                </button>
                <button
                  v-if="row.user.status === 'active'"
                  type="button"
                  class="btn btn-secondary px-3 py-1.5 text-xs text-red-600 dark:text-red-300"
                  @click="freezeUser(row.user)"
                >
                  {{ t('admin.risk.actions.freeze') }}
                </button>
                <button
                  v-else
                  type="button"
                  class="btn btn-secondary px-3 py-1.5 text-xs text-emerald-600 dark:text-emerald-300"
                  @click="unfreezeUser(row.user)"
                >
                  {{ t('admin.risk.actions.unfreeze') }}
                </button>
              </div>
            </template>
          </DataTable>
        </template>

        <template #pagination>
          <Pagination
            v-if="pagination.total > 0"
            :page="pagination.page"
            :total="pagination.total"
            :page-size="pagination.page_size"
            @update:page="handlePageChange"
            @update:pageSize="handlePageSizeChange"
          />
        </template>
      </TablePageLayout>
    </div>

    <RiskSettingsDialog :show="showSettingsDialog" @close="showSettingsDialog = false" @saved="handleSettingsSaved" />
    <UserRiskDetailModal :show="showRiskModal" :user="selectedUser" @close="closeRiskModal" />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'
import RiskDecisionBreakdownCard from '@/components/admin/risk/RiskDecisionBreakdownCard.vue'
import RiskScoreDistributionCard from '@/components/admin/risk/RiskScoreDistributionCard.vue'
import RiskActivityHeatmapCard from '@/components/admin/risk/RiskActivityHeatmapCard.vue'
import RiskLeaderboardCard from '@/components/admin/risk/RiskLeaderboardCard.vue'
import RiskSettingsDialog from '@/components/admin/risk/RiskSettingsDialog.vue'
import UserRiskDetailModal from '@/components/admin/user/UserRiskDetailModal.vue'
import { adminAPI } from '@/api/admin'
import type { UserRiskSummaryItem, UserRiskSettings } from '@/api/admin/risk'
import type { AdminUser } from '@/types'
import type { Column } from '@/components/common/types'
import { useAppStore } from '@/stores'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const appStore = useAppStore()

interface RiskRow extends UserRiskSummaryItem {
  user: AdminUser
}

type HeroFocusMode = 'all' | 'review_plus' | 'freeze_review'

const users = ref<AdminUser[]>([])
const summaries = ref<UserRiskSummaryItem[]>([])
const loading = ref(false)
const showSettingsDialog = ref(false)
const showRiskModal = ref(false)
const selectedUser = ref<AdminUser | null>(null)
const searchQuery = ref('')
const statusFilter = ref('')
const decisionFilter = ref('')
const sortMode = ref<'risk_desc' | 'spend_desc' | 'silence_asc'>('risk_desc')
const focusMode = ref<HeroFocusMode>('all')
let searchTimer: ReturnType<typeof setTimeout> | null = null

const riskSettings = ref<UserRiskSettings>({
  ip_intel_enabled: false,
  ip_intel_provider: 'ipinfo_lite',
  ip_intel_token_configured: false,
  ip_intel_docs_url: 'https://ipinfo.io/developers/lite-api',
  review_threshold: 50,
  throttle_threshold: 80,
  freeze_threshold: 120,
  auto_enabled: false,
  auto_throttle: false,
  auto_freeze: false,
  auto_throttle_concurrency_cap: 1
})

const pagination = ref({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0,
  pages: 0
})

const heroCards = computed<Array<{ key: HeroFocusMode | 'ip_intel' | 'automation'; label: string; value: string; hint: string; emphasisClass: string; clickable: boolean; status?: boolean }>>(() => [
  {
    key: 'all',
    label: t('admin.risk.cards.currentPageUsers'),
    value: String(riskRows.value.length),
    hint: t('admin.risk.visuals.currentPageHint'),
    emphasisClass: 'text-slate-950 dark:text-white',
    clickable: true
  },
  {
    key: 'review_plus',
    label: t('admin.risk.cards.reviewOrHigher'),
    value: String(reviewOrHigherCount.value),
    hint: t('admin.risk.visuals.reviewHint'),
    emphasisClass: 'text-amber-600 dark:text-amber-300',
    clickable: true
  },
  {
    key: 'freeze_review',
    label: t('admin.risk.cards.freezeReview'),
    value: String(freezeReviewCount.value),
    hint: t('admin.risk.visuals.freezeHint'),
    emphasisClass: 'text-red-600 dark:text-red-300',
    clickable: true
  },
  {
    key: 'ip_intel',
    label: t('admin.risk.cards.ipIntel'),
    value: riskSettings.value.ip_intel_enabled ? t('admin.risk.ipIntelEnabled') : t('admin.risk.ipIntelDisabled'),
    hint: riskSettings.value.ip_intel_provider,
    emphasisClass: 'text-slate-950 dark:text-white text-base',
    clickable: false,
    status: riskSettings.value.ip_intel_enabled
  },
  {
    key: 'automation',
    label: t('admin.risk.cards.automation'),
    value: riskSettings.value.auto_enabled ? t('admin.risk.automationEnabled') : t('admin.risk.automationDisabled'),
    hint: t('admin.risk.cards.automationHint', {
      throttle: riskSettings.value.auto_throttle ? t('common.enabled') : t('common.disabled'),
      freeze: riskSettings.value.auto_freeze ? t('common.enabled') : t('common.disabled')
    }),
    emphasisClass: 'text-slate-950 dark:text-white text-base',
    clickable: false,
    status: riskSettings.value.auto_enabled
  }
])

const statusOptions = computed(() => [
  { value: '', label: t('admin.risk.filters.allStatus') },
  { value: 'active', label: t('common.active') },
  { value: 'disabled', label: t('admin.users.disabled') }
])

const decisionOptions = computed(() => [
  { value: '', label: t('admin.risk.filters.allDecisions') },
  { value: 'observe', label: t('admin.risk.decisions.observe') },
  { value: 'review', label: t('admin.risk.decisions.review') },
  { value: 'throttle', label: t('admin.risk.decisions.throttle') },
  { value: 'freeze_review', label: t('admin.risk.decisions.freeze_review') }
])

const sortOptions = computed(() => [
  { value: 'risk_desc', label: t('admin.risk.sort.riskDesc') },
  { value: 'spend_desc', label: t('admin.risk.sort.spendDesc') },
  { value: 'silence_asc', label: t('admin.risk.sort.silenceAsc') }
])

const columns = computed<Column[]>(() => [
  { key: 'user', label: t('admin.risk.columns.user'), sortable: false },
  { key: 'risk_score', label: t('admin.risk.columns.riskScore'), sortable: false },
  { key: 'decision', label: t('admin.risk.columns.decision'), sortable: false },
  { key: 'actual_cost_24h', label: t('admin.risk.columns.cost24h'), sortable: false },
  { key: 'historical_ip_count', label: t('admin.risk.columns.ipHistory'), sortable: false },
  { key: 'ua_24h_count', label: t('admin.risk.columns.ua24h'), sortable: false },
  { key: 'active_hours_count', label: t('admin.risk.columns.activeHours'), sortable: false },
  { key: 'longest_silence_hours', label: t('admin.risk.columns.longestSilence'), sortable: false },
  { key: 'rule_hit_count', label: t('admin.risk.columns.ruleHits'), sortable: false },
  { key: 'actions', label: t('admin.users.columns.actions'), sortable: false }
])

const riskRows = computed<RiskRow[]>(() => {
  const summaryMap = new Map(summaries.value.map(item => [item.user_id, item]))
  return users.value
    .map((user) => {
      const summary = summaryMap.get(user.id)
      if (!summary) return null
      return { ...summary, user }
    })
    .filter((row): row is RiskRow => Boolean(row))
})

const filteredRows = computed(() => {
  return riskRows.value.filter((row) => {
    if (decisionFilter.value && row.summary.decision !== decisionFilter.value) return false
    if (focusMode.value === 'review_plus' && row.summary.risk_score < riskSettings.value.review_threshold) return false
    if (focusMode.value === 'freeze_review' && row.summary.decision !== 'freeze_review') return false
    return true
  })
})

const orderedRows = computed(() => {
  return [...filteredRows.value].sort((left, right) => {
    if (sortMode.value === 'spend_desc') {
      if (left.metrics.actual_cost_24h !== right.metrics.actual_cost_24h) {
        return right.metrics.actual_cost_24h - left.metrics.actual_cost_24h
      }
    } else if (sortMode.value === 'silence_asc') {
      if (left.metrics.longest_silence_hours !== right.metrics.longest_silence_hours) {
        return left.metrics.longest_silence_hours - right.metrics.longest_silence_hours
      }
    } else {
      if (left.summary.risk_score !== right.summary.risk_score) {
        return right.summary.risk_score - left.summary.risk_score
      }
    }
    if (left.summary.risk_score !== right.summary.risk_score) {
      return right.summary.risk_score - left.summary.risk_score
    }
    return left.user.id - right.user.id
  })
})

const reviewOrHigherCount = computed(() => filteredRows.value.filter(row => row.summary.risk_score >= riskSettings.value.review_threshold).length)
const freezeReviewCount = computed(() => filteredRows.value.filter(row => row.summary.decision === 'freeze_review').length)

function scoreClass(score: number) {
  if (score >= riskSettings.value.freeze_threshold) return 'text-red-600 dark:text-red-300'
  if (score >= riskSettings.value.throttle_threshold) return 'text-orange-600 dark:text-orange-300'
  if (score >= riskSettings.value.review_threshold) return 'text-amber-600 dark:text-amber-300'
  if (score >= Math.max(25, Math.floor(riskSettings.value.review_threshold / 2))) return 'text-yellow-600 dark:text-yellow-300'
  return 'text-emerald-600 dark:text-emerald-300'
}

function decisionClass(decision: string) {
  switch (decision) {
    case 'freeze_review':
      return 'bg-red-50 text-red-600 dark:bg-red-900/20 dark:text-red-300'
    case 'throttle':
      return 'bg-orange-50 text-orange-600 dark:bg-orange-900/20 dark:text-orange-300'
    case 'review':
      return 'bg-amber-50 text-amber-700 dark:bg-amber-900/20 dark:text-amber-300'
    case 'observe':
      return 'bg-yellow-50 text-yellow-700 dark:bg-yellow-900/20 dark:text-yellow-300'
    default:
      return 'bg-emerald-50 text-emerald-600 dark:bg-emerald-900/20 dark:text-emerald-300'
  }
}

async function loadSettings() {
  try {
    riskSettings.value = await adminAPI.risk.getSettings()
  } catch (err) {
    console.error('Failed to load risk settings:', err)
  }
}

async function loadPage() {
  loading.value = true
  try {
    const response = await adminAPI.users.list(pagination.value.page, pagination.value.page_size, {
      search: searchQuery.value || undefined,
      status: (statusFilter.value || undefined) as 'active' | 'disabled' | undefined,
      include_subscriptions: false,
      sort_by: 'created_at',
      sort_order: 'desc'
    })
    users.value = response.items
    pagination.value = {
      page: response.page,
      page_size: response.page_size,
      total: response.total,
      pages: response.pages
    }
    if (response.items.length === 0) {
      summaries.value = []
      return
    }
    summaries.value = await adminAPI.risk.getSummaries(response.items.map(item => item.id))
  } catch (err: any) {
    console.error('Failed to load risk management data:', err)
    appStore.showError(err?.response?.data?.detail || err?.message || t('admin.risk.loadFailed'))
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    pagination.value.page = 1
    void loadPage()
  }, 300)
}

function applyFilters() {
  pagination.value.page = 1
  void loadPage()
}

function handlePageChange(page: number) {
  pagination.value.page = page
  void loadPage()
}

function handlePageSizeChange(pageSize: number) {
  pagination.value.page_size = pageSize
  pagination.value.page = 1
  void loadPage()
}

function openRiskDetail(user: AdminUser) {
  selectedUser.value = user
  showRiskModal.value = true
}

function applyHeroCardFilter(mode: HeroFocusMode) {
  focusMode.value = mode
  if (mode === 'freeze_review') {
    decisionFilter.value = 'freeze_review'
  } else if (mode === 'all') {
    decisionFilter.value = ''
  } else {
    decisionFilter.value = ''
  }
}

async function applyThrottle(user: AdminUser) {
  try {
    await adminAPI.users.updateConcurrency(user.id, 1)
    appStore.showSuccess(t('admin.risk.actions.throttleSuccess'))
    await loadPage()
  } catch (err: any) {
    console.error('Failed to throttle user:', err)
    appStore.showError(err?.response?.data?.detail || err?.message || t('admin.risk.actions.throttleFailed'))
  }
}

async function freezeUser(user: AdminUser) {
  try {
    await adminAPI.users.toggleStatus(user.id, 'disabled')
    appStore.showSuccess(t('admin.risk.actions.freezeSuccess'))
    await loadPage()
  } catch (err: any) {
    console.error('Failed to freeze user:', err)
    appStore.showError(err?.response?.data?.detail || err?.message || t('admin.risk.actions.freezeFailed'))
  }
}

async function unfreezeUser(user: AdminUser) {
  try {
    await adminAPI.users.toggleStatus(user.id, 'active')
    appStore.showSuccess(t('admin.risk.actions.unfreezeSuccess'))
    await loadPage()
  } catch (err: any) {
    console.error('Failed to unfreeze user:', err)
    appStore.showError(err?.response?.data?.detail || err?.message || t('admin.risk.actions.unfreezeFailed'))
  }
}

function closeRiskModal() {
  selectedUser.value = null
  showRiskModal.value = false
}

function handleSettingsSaved(settings: UserRiskSettings) {
  riskSettings.value = settings
  showSettingsDialog.value = false
  void loadPage()
}

onMounted(async () => {
  await loadSettings()
  await loadPage()
})

onUnmounted(() => {
  if (searchTimer) clearTimeout(searchTimer)
})
</script>
