<template>
  <section class="rounded-[28px] border border-gray-200/80 bg-white/95 p-5 shadow-card dark:border-dark-800 dark:bg-dark-900/90">
    <div class="flex items-start justify-between gap-4">
      <div>
        <p class="text-xs font-semibold uppercase tracking-[0.24em] text-gray-400 dark:text-dark-500">{{ t('admin.risk.visuals.activityEyebrow') }}</p>
        <h3 class="mt-2 text-lg font-semibold text-gray-900 dark:text-white">{{ t('admin.risk.visuals.activityTitle') }}</h3>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.risk.visuals.activityHint') }}</p>
      </div>
      <div class="text-right text-sm text-gray-500 dark:text-gray-400">
        <p>{{ t('admin.users.riskAllDayActive') }}</p>
        <p class="mt-1 text-xl font-semibold text-gray-900 dark:text-white">{{ allDayUsers }}</p>
      </div>
    </div>

    <div v-if="rows.length === 0" class="mt-8 rounded-2xl border border-dashed border-gray-200 px-4 py-10 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
      {{ t('common.noData') }}
    </div>

    <template v-else>
      <div class="mt-6 grid grid-cols-6 gap-2 sm:grid-cols-8 xl:grid-cols-12">
        <div v-for="cell in heatmap" :key="cell.hour" class="rounded-2xl border border-gray-100 p-2 text-center dark:border-dark-800" :style="cell.style">
          <p class="text-[11px] font-medium text-gray-500 dark:text-dark-400">{{ cell.hourLabel }}</p>
          <p class="mt-2 text-sm font-semibold text-gray-900 dark:text-white">{{ cell.count }}</p>
        </div>
      </div>

      <div class="mt-6 grid gap-3 sm:grid-cols-3">
        <div class="rounded-2xl border border-gray-100 p-3 dark:border-dark-800">
          <p class="text-xs uppercase tracking-wide text-gray-400 dark:text-dark-500">{{ t('admin.risk.visuals.avgActiveHours') }}</p>
          <p class="mt-2 text-lg font-semibold text-gray-900 dark:text-white">{{ avgActiveHours.toFixed(1) }}</p>
        </div>
        <div class="rounded-2xl border border-gray-100 p-3 dark:border-dark-800">
          <p class="text-xs uppercase tracking-wide text-gray-400 dark:text-dark-500">{{ t('admin.users.riskConcurrentMinutes') }}</p>
          <p class="mt-2 text-lg font-semibold text-gray-900 dark:text-white">{{ concurrentUsers }}</p>
        </div>
        <div class="rounded-2xl border border-gray-100 p-3 dark:border-dark-800">
          <p class="text-xs uppercase tracking-wide text-gray-400 dark:text-dark-500">{{ t('admin.risk.visuals.avgConcentration') }}</p>
          <p class="mt-2 text-lg font-semibold text-gray-900 dark:text-white">{{ avgConcentration.toFixed(3) }}</p>
        </div>
      </div>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

interface RowLike {
  metrics: {
    active_hours: number[]
    active_hours_count: number
    all_day_active: boolean
    concurrent_multi_ip_ua_minutes_24h: number
    hour_concentration: number
  }
}

const props = defineProps<{ rows: RowLike[] }>()
const { t } = useI18n()

const countsByHour = computed(() => {
  const counts = Array.from({ length: 24 }, () => 0)
  for (const row of props.rows) {
    for (const hour of row.metrics.active_hours || []) {
      if (hour >= 0 && hour < 24) counts[hour] += 1
    }
  }
  return counts
})

const maxHourCount = computed(() => Math.max(...countsByHour.value, 1))

const heatmap = computed(() => {
  return countsByHour.value.map((count, hour) => {
    const intensity = maxHourCount.value ? count / maxHourCount.value : 0
    const bg = intensity === 0 ? 'rgba(148,163,184,0.08)' : `rgba(56,189,248,${0.14 + intensity * 0.5})`
    return {
      hour,
      count,
      hourLabel: `${String(hour).padStart(2, '0')}:00`,
      style: { backgroundColor: bg }
    }
  })
})

const avgActiveHours = computed(() => {
  if (!props.rows.length) return 0
  return props.rows.reduce((sum, row) => sum + row.metrics.active_hours_count, 0) / props.rows.length
})

const allDayUsers = computed(() => props.rows.filter((row) => row.metrics.all_day_active).length)
const concurrentUsers = computed(() => props.rows.filter((row) => row.metrics.concurrent_multi_ip_ua_minutes_24h > 0).length)
const avgConcentration = computed(() => {
  if (!props.rows.length) return 0
  return props.rows.reduce((sum, row) => sum + row.metrics.hour_concentration, 0) / props.rows.length
})
</script>
