<template>
  <section class="rounded-[28px] border border-gray-200/80 bg-white/95 p-5 shadow-card dark:border-dark-800 dark:bg-dark-900/90">
    <div class="flex items-start justify-between gap-4">
      <div>
        <p class="text-xs font-semibold tracking-[0.18em] text-gray-400 dark:text-dark-500">{{ t('admin.risk.visuals.activityEyebrow') }}</p>
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
      <div class="mt-6 grid gap-4 sm:grid-cols-3">
        <div class="rounded-2xl border border-gray-100 p-3 dark:border-dark-800">
          <p class="text-xs tracking-[0.16em] text-gray-400 dark:text-dark-500">{{ t('admin.risk.visuals.dayCoverage') }}</p>
          <p class="mt-2 text-lg font-semibold text-gray-900 dark:text-white">{{ avgDayCoverage.toFixed(1) }} / 12</p>
        </div>
        <div class="rounded-2xl border border-gray-100 p-3 dark:border-dark-800">
          <p class="text-xs tracking-[0.16em] text-gray-400 dark:text-dark-500">{{ t('admin.risk.visuals.nightCoverage') }}</p>
          <p class="mt-2 text-lg font-semibold text-gray-900 dark:text-white">{{ avgNightCoverage.toFixed(1) }} / 12</p>
        </div>
        <div class="rounded-2xl border border-gray-100 p-3 dark:border-dark-800">
          <p class="text-xs tracking-[0.16em] text-gray-400 dark:text-dark-500">{{ t('admin.risk.visuals.avgConcentration') }}</p>
          <p class="mt-2 text-lg font-semibold text-gray-900 dark:text-white">{{ avgConcentration.toFixed(3) }}</p>
        </div>
      </div>

      <div class="mt-6 space-y-4">
        <div class="rounded-2xl border border-gray-100 p-4 dark:border-dark-800">
          <div class="flex items-center justify-between gap-3">
            <div>
              <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.risk.visuals.dayTimeline') }}</p>
              <p class="text-xs text-gray-500 dark:text-gray-400">08:00 - 19:00</p>
            </div>
            <p class="text-sm font-medium text-gray-600 dark:text-gray-300">{{ dayPeakLabel }}</p>
          </div>
          <div class="mt-4 grid grid-cols-12 gap-2">
            <div v-for="point in dayTimeline" :key="`day-${point.hour}`" class="flex flex-col items-center gap-2">
              <div class="h-16 w-1 rounded-full bg-gray-100 dark:bg-dark-800">
                <div class="w-full rounded-full bg-sky-500 transition-all duration-300" :style="{ height: `${point.height}%`, marginTop: `${100 - point.height}%` }"></div>
              </div>
              <div class="text-center">
                <p class="text-[11px] font-medium text-gray-500 dark:text-dark-400">{{ point.label }}</p>
                <p class="text-[11px] text-gray-400 dark:text-dark-500">{{ point.count }}</p>
              </div>
            </div>
          </div>
        </div>

        <div class="rounded-2xl border border-gray-100 p-4 dark:border-dark-800">
          <div class="flex items-center justify-between gap-3">
            <div>
              <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.risk.visuals.nightTimeline') }}</p>
              <p class="text-xs text-gray-500 dark:text-gray-400">20:00 - 07:00</p>
            </div>
            <p class="text-sm font-medium text-gray-600 dark:text-gray-300">{{ nightPeakLabel }}</p>
          </div>
          <div class="mt-4 grid grid-cols-12 gap-2">
            <div v-for="point in nightTimeline" :key="`night-${point.hour}`" class="flex flex-col items-center gap-2">
              <div class="h-16 w-1 rounded-full bg-gray-100 dark:bg-dark-800">
                <div class="w-full rounded-full bg-violet-500 transition-all duration-300" :style="{ height: `${point.height}%`, marginTop: `${100 - point.height}%` }"></div>
              </div>
              <div class="text-center">
                <p class="text-[11px] font-medium text-gray-500 dark:text-dark-400">{{ point.label }}</p>
                <p class="text-[11px] text-gray-400 dark:text-dark-500">{{ point.count }}</p>
              </div>
            </div>
          </div>
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
const dayHours = [8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19]
const nightHours = [20, 21, 22, 23, 0, 1, 2, 3, 4, 5, 6, 7]

function buildTimeline(hours: number[]) {
  return hours.map((hour) => {
    const count = countsByHour.value[hour]
    return {
      hour,
      count,
      label: String(hour).padStart(2, '0'),
      height: maxHourCount.value ? Math.max(8, (count / maxHourCount.value) * 100) : 8
    }
  })
}

const dayTimeline = computed(() => buildTimeline(dayHours))
const nightTimeline = computed(() => buildTimeline(nightHours))

const allDayUsers = computed(() => props.rows.filter((row) => row.metrics.all_day_active).length)
const avgConcentration = computed(() => {
  if (!props.rows.length) return 0
  return props.rows.reduce((sum, row) => sum + row.metrics.hour_concentration, 0) / props.rows.length
})

const avgDayCoverage = computed(() => {
  if (!props.rows.length) return 0
  return props.rows.reduce((sum, row) => sum + (row.metrics.active_hours || []).filter(hour => dayHours.includes(hour)).length, 0) / props.rows.length
})

const avgNightCoverage = computed(() => {
  if (!props.rows.length) return 0
  return props.rows.reduce((sum, row) => sum + (row.metrics.active_hours || []).filter(hour => nightHours.includes(hour)).length, 0) / props.rows.length
})

function peakLabel(points: { hour: number; count: number }[], labelPrefix: string) {
  const peak = points.reduce((best, current) => (current.count > best.count ? current : best), points[0])
  return `${labelPrefix} ${String(peak.hour).padStart(2, '0')}:00 · ${peak.count}`
}

const dayPeakLabel = computed(() => peakLabel(dayTimeline.value, '峰值'))
const nightPeakLabel = computed(() => peakLabel(nightTimeline.value, '峰值'))
</script>
