<template>
  <BaseDialog
    :show="show"
    :title="t('admin.users.riskDetailTitle')"
    width="full"
    :close-on-click-outside="true"
    :z-index="40"
    @close="$emit('close')"
  >
    <div v-if="user" class="space-y-4">
      <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-700">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-3">
              <div class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-full bg-primary-100 dark:bg-primary-900/30">
                <span class="text-lg font-medium text-primary-700 dark:text-primary-300">
                  {{ user.email.charAt(0).toUpperCase() }}
                </span>
              </div>
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <p class="truncate font-medium text-gray-900 dark:text-white">{{ user.email }}</p>
                  <span v-if="user.username" class="rounded bg-primary-50 px-1.5 py-0.5 text-xs text-primary-600 dark:bg-primary-900/20 dark:text-primary-400">
                    {{ user.username }}
                  </span>
                  <span :class="statusBadgeClass" class="rounded px-2 py-0.5 text-xs font-medium">
                    {{ user.status }}
                  </span>
                </div>
                <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
                  {{ t('admin.users.riskWindow24h') }}: {{ detail ? formatDateTime(detail.window.start_at) : '-' }} → {{ detail ? formatDateTime(detail.window.end_at) : '-' }}
                  <span v-if="detail">({{ detail.window.timezone }})</span>
                </p>
              </div>
            </div>
            <div class="mt-3 flex flex-wrap items-center gap-4 text-sm text-gray-600 dark:text-gray-300">
              <span>{{ t('admin.users.currentBalance') }}: <strong class="text-gray-900 dark:text-white">${{ user.balance.toFixed(2) }}</strong></span>
              <span v-if="detail">{{ t('admin.users.riskComputedAt') }}: {{ formatDateTime(detail.summary.computed_at) }}</span>
            </div>
          </div>

          <div v-if="detail" class="grid min-w-[220px] gap-3 sm:grid-cols-2 lg:grid-cols-1">
            <div class="rounded-xl border border-gray-200 bg-white px-4 py-3 dark:border-dark-600 dark:bg-dark-800">
              <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('admin.users.riskScore') }}</p>
              <p :class="riskScoreClass" class="mt-1 text-3xl font-bold">{{ detail.summary.risk_score }}</p>
            </div>
            <div class="rounded-xl border border-gray-200 bg-white px-4 py-3 dark:border-dark-600 dark:bg-dark-800">
              <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('admin.users.riskDecision') }}</p>
              <p :class="decisionBadgeClass" class="mt-1 inline-flex rounded-full px-2.5 py-1 text-sm font-semibold">
                {{ detail.summary.decision_label }}
              </p>
            </div>
          </div>
        </div>
      </div>

      <div class="flex items-center justify-between gap-3">
        <p class="text-sm text-gray-500 dark:text-dark-400">
          {{ t('admin.users.riskDerivedOnly') }}
        </p>
        <button
          type="button"
          class="btn btn-secondary px-3 py-2 text-sm"
          :disabled="loading"
          @click="loadDetail"
        >
          {{ loading ? t('common.loading') : t('common.refresh') }}
        </button>
      </div>

      <div v-if="error" class="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-600 dark:border-red-900/60 dark:bg-red-950/30 dark:text-red-300">
        {{ error }}
      </div>

      <div v-if="loading" class="flex justify-center py-12">
        <svg class="h-8 w-8 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
        </svg>
      </div>

      <template v-else-if="detail">
        <section class="space-y-3">
          <div class="flex items-center justify-between">
            <h4 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.users.riskMetricsTitle') }}</h4>
            <span class="text-xs text-gray-400 dark:text-dark-500">24h + history</span>
          </div>
          <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
            <div
              v-for="card in metricCards"
              :key="card.label"
              class="rounded-xl border border-gray-200 bg-white px-4 py-3 dark:border-dark-600 dark:bg-dark-800"
            >
              <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ card.label }}</p>
              <p class="mt-1 break-words text-sm font-semibold text-gray-900 dark:text-white">{{ card.value }}</p>
            </div>
          </div>
        </section>

        <section class="space-y-3">
          <h4 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.users.riskRuleHitsTitle') }}</h4>
          <div v-if="detail.rule_hits.length === 0" class="rounded-xl border border-gray-200 bg-white px-4 py-6 text-sm text-gray-500 dark:border-dark-600 dark:bg-dark-800 dark:text-dark-400">
            {{ t('admin.users.riskNoRuleHits') }}
          </div>
          <div v-else class="grid gap-3 lg:grid-cols-2">
            <div
              v-for="rule in detail.rule_hits"
              :key="rule.code"
              class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-800"
            >
              <div class="flex items-start justify-between gap-3">
                <div>
                  <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ rule.code }} · {{ rule.label }}</p>
                  <p class="mt-1 text-sm text-gray-600 dark:text-gray-300">{{ rule.description }}</p>
                </div>
                <span class="rounded-full bg-red-50 px-2 py-0.5 text-sm font-semibold text-red-600 dark:bg-red-900/20 dark:text-red-300">+{{ rule.score }}</span>
              </div>
            </div>
          </div>
        </section>

        <section class="grid gap-4 xl:grid-cols-2">
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <h4 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.users.riskIpDetailsTitle') }}</h4>
              <span class="text-xs text-gray-400 dark:text-dark-500">24h</span>
            </div>
            <p class="text-xs text-amber-600 dark:text-amber-300">{{ t('admin.users.riskIpIntelNote') }}</p>
            <div class="overflow-x-auto rounded-xl border border-gray-200 bg-white dark:border-dark-600 dark:bg-dark-800">
              <table class="min-w-full divide-y divide-gray-200 text-sm dark:divide-dark-600">
                <thead class="bg-gray-50 dark:bg-dark-700">
                  <tr>
                    <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-dark-300">IP</th>
                    <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-dark-300">{{ t('admin.users.riskRequestsShort') }}</th>
                    <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-dark-300">{{ t('admin.users.riskIpType') }}</th>
                    <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-dark-300">{{ t('admin.users.riskIpLabel') }}</th>
                    <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-dark-300">ASN</th>
                    <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-dark-300">{{ t('admin.users.riskOrganization') }}</th>
                    <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-dark-300">{{ t('admin.users.riskRegion') }}</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                  <tr v-if="detail.ip_details.length === 0">
                    <td colspan="7" class="px-4 py-6 text-center text-sm text-gray-500 dark:text-dark-400">{{ t('admin.users.riskNoIpDetails') }}</td>
                  </tr>
                  <tr v-for="item in detail.ip_details" :key="item.ip_address">
                    <td class="px-4 py-3 font-mono text-xs text-gray-700 dark:text-gray-200">{{ item.ip_address }}</td>
                    <td class="px-4 py-3 text-gray-700 dark:text-gray-200">{{ item.requests }}</td>
                    <td class="px-4 py-3 text-gray-500 dark:text-dark-300">{{ item.ip_type }}</td>
                    <td class="px-4 py-3 text-gray-500 dark:text-dark-300">{{ item.label }}</td>
                    <td class="px-4 py-3 text-gray-500 dark:text-dark-300">{{ item.asn || '-' }}</td>
                    <td class="px-4 py-3 text-gray-500 dark:text-dark-300">{{ item.organization || item.domain || '-' }}</td>
                    <td class="px-4 py-3 text-gray-500 dark:text-dark-300">{{ item.country || item.country_code || '-' }}<span v-if="item.continent"> / {{ item.continent }}</span></td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <h4 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.users.riskUaDetailsTitle') }}</h4>
              <span class="text-xs text-gray-400 dark:text-dark-500">24h</span>
            </div>
            <div class="overflow-x-auto rounded-xl border border-gray-200 bg-white dark:border-dark-600 dark:bg-dark-800">
              <table class="min-w-full divide-y divide-gray-200 text-sm dark:divide-dark-600">
                <thead class="bg-gray-50 dark:bg-dark-700">
                  <tr>
                    <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-dark-300">UA</th>
                    <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-dark-300">{{ t('admin.users.riskRequestsShort') }}</th>
                    <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-dark-300">{{ t('admin.users.riskCategory') }}</th>
                    <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-dark-300">{{ t('admin.users.riskBaseScore') }}</th>
                    <th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-dark-300">{{ t('admin.users.riskConfigStatus') }}</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                  <tr v-if="detail.ua_details.length === 0">
                    <td colspan="5" class="px-4 py-6 text-center text-sm text-gray-500 dark:text-dark-400">{{ t('admin.users.riskNoUaDetails') }}</td>
                  </tr>
                  <tr v-for="item in detail.ua_details" :key="item.user_agent">
                    <td class="px-4 py-3">
                      <div class="max-w-md">
                        <p class="break-all text-xs text-gray-700 dark:text-gray-200">{{ item.user_agent }}</p>
                        <p class="mt-1 text-xs text-gray-400 dark:text-dark-500">{{ item.description }}</p>
                        <p v-if="item.hit_rule" class="mt-1 text-[11px] text-primary-600 dark:text-primary-400">{{ item.hit_rule }}</p>
                      </div>
                    </td>
                    <td class="px-4 py-3 text-gray-700 dark:text-gray-200">{{ item.requests }}</td>
                    <td class="px-4 py-3 text-gray-500 dark:text-dark-300">{{ item.category }}</td>
                    <td class="px-4 py-3 text-gray-700 dark:text-gray-200">{{ item.base_score }}</td>
                    <td class="px-4 py-3">
                      <span :class="configStatusClass(item.config_status)" class="rounded-full px-2 py-0.5 text-xs font-medium">
                        {{ configStatusLabel(item.config_status) }}
                      </span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </section>
      </template>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI, type UserRiskDetail } from '@/api/admin'
import type { AdminUser } from '@/types'
import { formatDateTime } from '@/utils/format'
import BaseDialog from '@/components/common/BaseDialog.vue'

const props = defineProps<{ show: boolean; user: AdminUser | null }>()
defineEmits(['close'])
const { t } = useI18n()

const detail = ref<UserRiskDetail | null>(null)
const loading = ref(false)
const error = ref('')

const metricCards = computed(() => {
  if (!detail.value) return []
  return [
    { label: t('admin.users.riskRequests24h'), value: String(detail.value.metrics.request_count_24h) },
    { label: t('admin.users.riskCost24h'), value: `$${detail.value.metrics.actual_cost_24h.toFixed(4)}` },
    { label: t('admin.users.riskKeyCount'), value: String(detail.value.metrics.key_count) },
    { label: t('admin.users.riskFirstIp'), value: detail.value.metrics.first_ip || '-' },
    { label: t('admin.users.riskHistoricalIpCount'), value: String(detail.value.metrics.historical_ip_count) },
    { label: t('admin.users.riskUa24hCount'), value: String(detail.value.metrics.ua_24h_count) },
    { label: t('admin.users.riskActiveHoursCount'), value: String(detail.value.metrics.active_hours_count) },
    { label: t('admin.users.riskLongestSilence'), value: `${detail.value.metrics.longest_silence_hours.toFixed(2)}h` },
    { label: t('admin.users.riskConcurrentMinutes'), value: String(detail.value.metrics.concurrent_multi_ip_ua_minutes_24h) },
    { label: t('admin.users.riskHourConcentration'), value: detail.value.metrics.hour_concentration.toFixed(4) },
    { label: t('admin.users.riskAllDayActive'), value: detail.value.metrics.all_day_active ? t('common.yes') : t('common.no') },
    { label: t('admin.users.riskActiveHours'), value: detail.value.metrics.active_hours.length ? detail.value.metrics.active_hours.join(', ') : '-' }
  ]
})

const statusBadgeClass = computed(() => {
  if (!props.user) return 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-300'
  return props.user.status === 'active'
    ? 'bg-emerald-50 text-emerald-600 dark:bg-emerald-900/20 dark:text-emerald-300'
    : 'bg-red-50 text-red-600 dark:bg-red-900/20 dark:text-red-300'
})

const riskScoreClass = computed(() => {
  const score = detail.value?.summary.risk_score ?? 0
  if (score >= 120) return 'text-red-600 dark:text-red-300'
  if (score >= 80) return 'text-orange-600 dark:text-orange-300'
  if (score >= 50) return 'text-amber-600 dark:text-amber-300'
  if (score >= 25) return 'text-yellow-600 dark:text-yellow-300'
  return 'text-emerald-600 dark:text-emerald-300'
})

const decisionBadgeClass = computed(() => {
  const score = detail.value?.summary.risk_score ?? 0
  if (score >= 120) return 'bg-red-50 text-red-600 dark:bg-red-900/20 dark:text-red-300'
  if (score >= 80) return 'bg-orange-50 text-orange-600 dark:bg-orange-900/20 dark:text-orange-300'
  if (score >= 50) return 'bg-amber-50 text-amber-600 dark:bg-amber-900/20 dark:text-amber-300'
  if (score >= 25) return 'bg-yellow-50 text-yellow-700 dark:bg-yellow-900/20 dark:text-yellow-300'
  return 'bg-emerald-50 text-emerald-600 dark:bg-emerald-900/20 dark:text-emerald-300'
})

watch(
  () => [props.show, props.user?.id] as const,
  ([show]) => {
    if (show && props.user) {
      void loadDetail()
    }
  },
  { immediate: false }
)

const loadDetail = async () => {
  if (!props.user) return
  loading.value = true
  error.value = ''
  try {
    const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC'
    detail.value = await adminAPI.risk.getUserDetail(props.user.id, timezone)
  } catch (err: any) {
    console.error('Failed to load user risk detail:', err)
    error.value = err?.response?.data?.detail || err?.message || t('admin.users.riskLoadFailed')
  } finally {
    loading.value = false
  }
}

const configStatusClass = (status: string) => {
  switch (status) {
    case 'normal':
      return 'bg-emerald-50 text-emerald-600 dark:bg-emerald-900/20 dark:text-emerald-300'
    case 'abnormal':
      return 'bg-red-50 text-red-600 dark:bg-red-900/20 dark:text-red-300'
    default:
      return 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-300'
  }
}

const configStatusLabel = (status: string) => {
  switch (status) {
    case 'normal':
      return t('admin.users.riskConfigStatusNormal')
    case 'abnormal':
      return t('admin.users.riskConfigStatusAbnormal')
    default:
      return t('admin.users.riskConfigStatusUnconfigured')
  }
}
</script>
