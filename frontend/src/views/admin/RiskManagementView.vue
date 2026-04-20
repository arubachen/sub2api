<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="rounded-[28px] border border-gray-200/80 bg-white/95 p-6 shadow-card dark:border-dark-800 dark:bg-dark-900/90">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <h1 class="text-2xl font-semibold tracking-tight text-gray-900 dark:text-white">{{ t('admin.risk.title') }}</h1>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.risk.description') }}</p>
          </div>
          <div class="flex flex-wrap items-center gap-3">
            <button type="button" class="btn btn-secondary px-3 py-2 text-sm" @click="showSettingsDialog = true">
              {{ t('common.settings') }}
            </button>
            <button type="button" class="btn btn-primary px-3 py-2 text-sm" :disabled="loading" @click="loadPage">
              {{ loading ? t('common.loading') : t('common.refresh') }}
            </button>
          </div>
        </div>

        <div class="mt-5 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <div class="rounded-2xl border border-gray-200 px-4 py-3 dark:border-dark-700">
            <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">{{ t('admin.risk.cards.currentPageUsers') }}</p>
            <p class="mt-1 text-2xl font-semibold text-gray-900 dark:text-white">{{ riskRows.length }}</p>
          </div>
          <div class="rounded-2xl border border-gray-200 px-4 py-3 dark:border-dark-700">
            <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">{{ t('admin.risk.cards.reviewOrHigher') }}</p>
            <p class="mt-1 text-2xl font-semibold text-amber-600 dark:text-amber-300">{{ reviewOrHigherCount }}</p>
          </div>
          <div class="rounded-2xl border border-gray-200 px-4 py-3 dark:border-dark-700">
            <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">{{ t('admin.risk.cards.freezeReview') }}</p>
            <p class="mt-1 text-2xl font-semibold text-red-600 dark:text-red-300">{{ freezeReviewCount }}</p>
          </div>
          <div class="rounded-2xl border border-gray-200 px-4 py-3 dark:border-dark-700">
            <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">{{ t('admin.risk.cards.ipIntel') }}</p>
            <p class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">
              {{ riskSettings.ip_intel_enabled ? t('admin.risk.ipIntelEnabled') : t('admin.risk.ipIntelDisabled') }}
            </p>
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ riskSettings.ip_intel_provider }}</p>
          </div>
        </div>
      </div>

      <TablePageLayout>
        <template #filters>
          <div class="rounded-[28px] border border-gray-200/80 bg-white/95 p-4 shadow-card dark:border-dark-800 dark:bg-dark-900/90">
            <div class="grid gap-3 lg:grid-cols-[minmax(0,1fr)_200px_160px_auto] lg:items-center">
              <input
                v-model="searchQuery"
                type="text"
                class="input"
                :placeholder="t('admin.risk.searchPlaceholder')"
                @input="handleSearch"
              />
              <Select v-model="statusFilter" :options="statusOptions" @change="applyFilters" />
              <Select v-model="decisionFilter" :options="decisionOptions" @change="applyFilters" />
              <div class="text-sm text-gray-500 dark:text-gray-400">
                {{ t('admin.risk.windowLabel') }}: 24h
              </div>
            </div>
          </div>
        </template>

        <template #table>
          <DataTable
            :columns="columns"
            :data="filteredRows"
            :loading="loading"
            :actions-count="1"
          >
            <template #cell-user="{ row }">
              <button type="button" class="text-left" @click="openRiskDetail(row.user)">
                <p class="font-medium text-primary-600 dark:text-primary-400">{{ row.user.email }}</p>
                <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">{{ row.user.username || '-' }}</p>
              </button>
            </template>

            <template #cell-risk_score="{ row }">
              <span :class="scoreClass(row.summary.risk_score)" class="font-semibold">{{ row.summary.risk_score }}</span>
            </template>

            <template #cell-decision="{ row }">
              <span :class="decisionClass(row.summary.decision)" class="rounded-full px-2 py-0.5 text-xs font-medium">
                {{ row.summary.decision_label }}
              </span>
            </template>

            <template #cell-actual_cost_24h="{ row }">
              ${{ row.metrics.actual_cost_24h.toFixed(4) }}
            </template>

            <template #cell-historical_ip_count="{ row }">
              <div>
                <p>{{ row.metrics.historical_ip_count }}</p>
                <p class="text-xs text-gray-400 dark:text-gray-500">{{ row.metrics.first_ip || '-' }}</p>
              </div>
            </template>

            <template #cell-ua_24h_count="{ row }">
              <div>
                <p>{{ row.metrics.ua_24h_count }}</p>
                <p class="max-w-[240px] truncate text-xs text-gray-400 dark:text-gray-500">{{ row.top_user_agent || '-' }}</p>
              </div>
            </template>

            <template #cell-active_hours_count="{ row }">
              <div>
                <p>{{ row.metrics.active_hours_count }}</p>
                <p class="text-xs text-gray-400 dark:text-gray-500">{{ row.metrics.all_day_active ? t('common.yes') : t('common.no') }}</p>
              </div>
            </template>

            <template #cell-longest_silence_hours="{ row }">
              {{ row.metrics.longest_silence_hours.toFixed(2) }}h
            </template>

            <template #cell-rule_hit_count="{ row }">
              {{ row.rule_hit_count }}
            </template>

            <template #cell-actions="{ row }">
              <button type="button" class="btn btn-secondary px-3 py-1.5 text-xs" @click="openRiskDetail(row.user)">
                {{ t('common.details') }}
              </button>
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
import RiskSettingsDialog from '@/components/admin/risk/RiskSettingsDialog.vue'
import UserRiskDetailModal from '@/components/admin/user/UserRiskDetailModal.vue'
import { adminAPI } from '@/api/admin'
import type { UserRiskSummaryItem, UserRiskSettings } from '@/api/admin/risk'
import type { AdminUser } from '@/types'
import type { Column } from '@/components/common/types'
import { useAppStore } from '@/stores'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'

const { t } = useI18n()
const appStore = useAppStore()

interface RiskRow extends UserRiskSummaryItem {
  user: AdminUser
}

const users = ref<AdminUser[]>([])
const summaries = ref<UserRiskSummaryItem[]>([])
const loading = ref(false)
const showSettingsDialog = ref(false)
const showRiskModal = ref(false)
const selectedUser = ref<AdminUser | null>(null)
const searchQuery = ref('')
const statusFilter = ref('')
const decisionFilter = ref('')
let searchTimer: ReturnType<typeof setTimeout> | null = null

const riskSettings = ref<UserRiskSettings>({
  ip_intel_enabled: false,
  ip_intel_provider: 'ipinfo_lite',
  ip_intel_token_configured: false,
  ip_intel_docs_url: 'https://ipinfo.io/developers/lite-api',
  review_threshold: 50,
  throttle_threshold: 80,
  freeze_threshold: 120
})

const pagination = ref({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0,
  pages: 0
})

const statusOptions = computed(() => [
  { value: '', label: t('admin.risk.filters.allStatus') },
  { value: 'active', label: t('common.active') },
  { value: 'disabled', label: t('admin.users.disabled') }
])

const decisionOptions = computed(() => [
  { value: '', label: t('admin.risk.filters.allDecisions') },
  { value: 'review', label: t('admin.risk.decisions.review') },
  { value: 'throttle', label: t('admin.risk.decisions.throttle') },
  { value: 'freeze_review', label: t('admin.risk.decisions.freeze_review') }
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
  return riskRows.value
    .filter((row) => {
      if (decisionFilter.value && row.summary.decision !== decisionFilter.value) return false
      return true
    })
    .sort((left, right) => {
      if (left.summary.risk_score !== right.summary.risk_score) {
        return right.summary.risk_score - left.summary.risk_score
      }
      if (left.metrics.actual_cost_24h !== right.metrics.actual_cost_24h) {
        return right.metrics.actual_cost_24h - left.metrics.actual_cost_24h
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
