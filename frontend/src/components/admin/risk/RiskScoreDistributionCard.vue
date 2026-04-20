<template>
  <section class="rounded-[28px] border border-gray-200/80 bg-white/95 p-5 shadow-card dark:border-dark-800 dark:bg-dark-900/90">
    <div class="flex items-start justify-between gap-4">
      <div>
        <div class="flex items-center gap-2">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('admin.risk.visuals.scoreTitle') }}</h3>
          <HelpTooltip :content="t('admin.risk.visuals.scoreHint')">
            <template #trigger>
              <span class="inline-flex rounded-full text-gray-400 hover:text-primary-500 dark:text-gray-500 dark:hover:text-primary-400">
                <Icon name="questionCircle" size="sm" />
              </span>
            </template>
          </HelpTooltip>
        </div>
      </div>
      <div class="text-right text-sm text-gray-500 dark:text-gray-400">
        <p>{{ t('admin.risk.visuals.maxScore') }}</p>
        <p class="mt-1 text-xl font-semibold text-gray-900 dark:text-white">{{ maxScore }}</p>
      </div>
    </div>

    <div v-if="rows.length === 0" class="mt-8 rounded-2xl border border-dashed border-gray-200 px-4 py-10 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
      {{ t('common.noData') }}
    </div>

    <div v-else class="mt-6 space-y-4">
      <div v-for="bucket in buckets" :key="bucket.key" class="grid gap-2 sm:grid-cols-[108px_minmax(0,1fr)_56px] sm:items-center">
        <div>
          <p class="text-sm font-medium text-gray-900 dark:text-white">{{ bucket.label }}</p>
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ bucket.range }}</p>
        </div>
        <div class="h-3 overflow-hidden rounded-full bg-gray-100 dark:bg-dark-800">
          <div class="h-full rounded-full transition-all duration-300" :style="{ width: `${bucket.percent}%`, background: bucket.gradient }"></div>
        </div>
        <div class="text-right text-sm font-semibold text-gray-700 dark:text-gray-200">{{ bucket.count }}</div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { UserRiskSettings } from '@/api/admin/risk'
import HelpTooltip from '@/components/common/HelpTooltip.vue'
import Icon from '@/components/icons/Icon.vue'

interface RowLike {
  summary: {
    risk_score: number
  }
}

const props = defineProps<{ rows: RowLike[]; settings: UserRiskSettings }>()
const { t } = useI18n()

const maxScore = computed(() => props.rows.reduce((max, row) => Math.max(max, row.summary.risk_score), 0))

const observeThreshold = computed(() => Math.max(25, Math.floor(props.settings.review_threshold / 2)))

const buckets = computed(() => {
  const definitions = [
    {
      key: 'normal',
      label: t('admin.users.riskConfigStatusNormal'),
      range: `0-${observeThreshold.value - 1}`,
      match: (score: number) => score < observeThreshold.value,
      gradient: 'linear-gradient(90deg, #10b981 0%, #34d399 100%)'
    },
    {
      key: 'observe',
      label: t('admin.risk.decisions.observe'),
      range: `${observeThreshold.value}-${props.settings.review_threshold - 1}`,
      match: (score: number) => score >= observeThreshold.value && score < props.settings.review_threshold,
      gradient: 'linear-gradient(90deg, #facc15 0%, #fde047 100%)'
    },
    {
      key: 'review',
      label: t('admin.risk.decisions.review'),
      range: `${props.settings.review_threshold}-${props.settings.throttle_threshold - 1}`,
      match: (score: number) => score >= props.settings.review_threshold && score < props.settings.throttle_threshold,
      gradient: 'linear-gradient(90deg, #f59e0b 0%, #fbbf24 100%)'
    },
    {
      key: 'throttle',
      label: t('admin.risk.decisions.throttle'),
      range: `${props.settings.throttle_threshold}-${props.settings.freeze_threshold - 1}`,
      match: (score: number) => score >= props.settings.throttle_threshold && score < props.settings.freeze_threshold,
      gradient: 'linear-gradient(90deg, #f97316 0%, #fb923c 100%)'
    },
    {
      key: 'freeze_review',
      label: t('admin.risk.decisions.freeze_review'),
      range: `${props.settings.freeze_threshold}+`,
      match: (score: number) => score >= props.settings.freeze_threshold,
      gradient: 'linear-gradient(90deg, #ef4444 0%, #f87171 100%)'
    }
  ]

  return definitions.map((bucket) => {
    const count = props.rows.filter((row) => bucket.match(row.summary.risk_score)).length
    return {
      ...bucket,
      count,
      percent: props.rows.length ? (count / props.rows.length) * 100 : 0
    }
  })
})
</script>
