<template>
  <section class="rounded-[28px] border border-gray-200/80 bg-white/95 p-5 shadow-card dark:border-dark-800 dark:bg-dark-900/90">
    <div class="flex items-center justify-between gap-4">
      <div class="flex items-center gap-2">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('admin.risk.visuals.leaderboardTitle') }}</h3>
        <HelpTooltip :content="t('admin.risk.visuals.leaderboardHint')" />
      </div>
      <span class="rounded-full bg-gray-100 px-3 py-1 text-xs font-medium text-gray-600 dark:bg-dark-800 dark:text-dark-300">TOP {{ topRows.length }}</span>
    </div>

    <div v-if="topRows.length === 0" class="mt-8 rounded-2xl border border-dashed border-gray-200 px-4 py-10 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
      {{ t('common.noData') }}
    </div>

    <div v-else class="mt-6 space-y-3">
      <button
        v-for="(row, index) in topRows"
        :key="row.user.id"
        type="button"
        class="flex w-full items-center justify-between gap-4 rounded-2xl border border-gray-100 px-4 py-3 text-left transition-colors hover:border-primary-200 hover:bg-primary-50/40 dark:border-dark-800 dark:hover:border-primary-900/50 dark:hover:bg-primary-900/10"
        @click="$emit('open-user', row.user)"
      >
        <div class="flex min-w-0 items-center gap-3">
          <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-slate-100 text-sm font-semibold text-slate-700 dark:bg-dark-800 dark:text-slate-200">
            {{ index + 1 }}
          </div>
          <div class="min-w-0">
            <p class="truncate font-medium text-gray-900 dark:text-white">{{ row.user.email }}</p>
            <p class="mt-0.5 truncate text-xs text-gray-500 dark:text-gray-400">{{ row.summary.decision_label }}</p>
          </div>
        </div>
        <div class="text-right">
          <p class="text-lg font-semibold" :class="scoreClass(row.summary.risk_score)">{{ row.summary.risk_score }}</p>
          <p class="text-xs text-gray-400 dark:text-gray-500">{{ row.rule_hit_count }} {{ t('admin.risk.columns.ruleHits') }}</p>
        </div>
      </button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import HelpTooltip from '@/components/common/HelpTooltip.vue'
import type { AdminUser } from '@/types'
import type { UserRiskSummaryItem, UserRiskSettings } from '@/api/admin/risk'

interface RiskRow extends UserRiskSummaryItem {
  user: AdminUser
}

const props = defineProps<{ rows: RiskRow[]; settings: UserRiskSettings }>()
defineEmits<{ (e: 'open-user', user: AdminUser): void }>()
const { t } = useI18n()

const topRows = computed(() => props.rows.slice(0, 5))

function scoreClass(score: number) {
  if (score >= props.settings.freeze_threshold) return 'text-red-600 dark:text-red-300'
  if (score >= props.settings.throttle_threshold) return 'text-orange-600 dark:text-orange-300'
  if (score >= props.settings.review_threshold) return 'text-amber-600 dark:text-amber-300'
  return 'text-emerald-600 dark:text-emerald-300'
}
</script>
