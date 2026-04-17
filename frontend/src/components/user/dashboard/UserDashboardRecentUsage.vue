<template>
  <div class="card overflow-hidden">
    <div class="flex items-center justify-between border-b border-gray-100 px-6 py-5 dark:border-dark-800">
      <div>
        <p class="text-xs font-semibold uppercase tracking-[0.24em] text-gray-400 dark:text-dark-400">Recent activity</p>
        <h2 class="mt-2 text-lg font-semibold text-gray-900 dark:text-white">{{ t('dashboard.recentUsage') }}</h2>
      </div>
      <span class="badge badge-gray">{{ t('dashboard.last7Days') }}</span>
    </div>
    <div class="p-6">
      <div v-if="loading" class="flex items-center justify-center py-12">
        <LoadingSpinner size="lg" />
      </div>
      <div v-else-if="data.length === 0" class="py-8">
        <EmptyState :title="t('dashboard.noUsageRecords')" :description="t('dashboard.startUsingApi')" />
      </div>
      <div v-else class="space-y-3">
        <div
          v-for="log in data"
          :key="log.id"
          class="group flex items-center justify-between rounded-[24px] border border-gray-200/80 bg-white/90 p-4 transition-all duration-200 hover:border-primary-100 hover:bg-primary-50/45 dark:border-dark-800 dark:bg-dark-950/70 dark:hover:border-primary-900/40 dark:hover:bg-primary-950/15"
        >
          <div class="flex min-w-0 items-center gap-4">
            <div class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-2xl bg-primary-50 ring-1 ring-primary-100 dark:bg-primary-950/30 dark:ring-primary-900/30">
              <Icon name="beaker" size="md" class="text-primary-600 dark:text-primary-400" />
            </div>
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <p class="truncate text-sm font-semibold text-gray-900 dark:text-white">{{ log.model }}</p>
                <span class="badge badge-gray">{{ formatTokens((log.input_tokens || 0) + (log.output_tokens || 0)) }}</span>
              </div>
              <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ formatDateTime(log.created_at) }}</p>
            </div>
          </div>
          <div class="ml-4 text-right">
            <div class="rounded-full border border-primary-100 bg-primary-50/80 px-3 py-1.5 dark:border-primary-900/30 dark:bg-primary-950/20">
              <p class="text-sm font-semibold text-primary-700 dark:text-primary-300">${{ formatCost(log.actual_cost) }}</p>
              <p class="mt-0.5 text-[11px] text-gray-500 dark:text-dark-400">{{ t('dashboard.standard') }} ${{ formatCost(log.total_cost) }}</p>
            </div>
          </div>
        </div>

        <router-link to="/usage" class="flex items-center justify-center gap-2 py-3 text-sm font-medium text-primary-600 transition-colors hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300">
          {{ t('dashboard.viewAllUsage') }}
          <Icon name="arrowRight" size="sm" />
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime } from '@/utils/format'
import type { UsageLog } from '@/types'

defineProps<{
  data: UsageLog[]
  loading: boolean
}>()
const { t } = useI18n()
const formatCost = (cost: number) => cost.toFixed(4)
const formatTokens = (value: number) => {
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M tokens`
  if (value >= 1000) return `${(value / 1000).toFixed(1)}K tokens`
  return `${value} tokens`
}
</script>
