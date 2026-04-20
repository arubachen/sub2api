<template>
  <section class="rounded-[28px] border border-gray-200/80 bg-white/95 p-5 shadow-card dark:border-dark-800 dark:bg-dark-900/90">
    <div class="flex items-start justify-between gap-4">
      <div>
        <div class="flex items-center gap-2">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('admin.risk.visuals.activityTitle') }}</h3>
          <HelpTooltip :content="t('admin.risk.visuals.activityHint')">
            <template #trigger>
              <span class="inline-flex rounded-full text-gray-400 hover:text-primary-500 dark:text-gray-500 dark:hover:text-primary-400">
                <Icon name="questionCircle" size="sm" />
              </span>
            </template>
          </HelpTooltip>
        </div>
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
          <p class="text-xs tracking-[0.16em] text-gray-400 dark:text-dark-500">{{ t('admin.risk.visuals.activeWindowCoverage') }}</p>
          <p class="mt-2 text-lg font-semibold text-gray-900 dark:text-white">{{ avgActiveWindow.toFixed(1) }} / 24</p>
        </div>
        <div class="rounded-2xl border border-gray-100 p-3 dark:border-dark-800">
          <p class="text-xs tracking-[0.16em] text-gray-400 dark:text-dark-500">{{ t('admin.risk.visuals.activityPeak') }}</p>
          <p class="mt-2 text-lg font-semibold text-gray-900 dark:text-white">{{ peakHourLabel }}</p>
        </div>
        <div class="rounded-2xl border border-gray-100 p-3 dark:border-dark-800">
          <p class="text-xs tracking-[0.16em] text-gray-400 dark:text-dark-500">{{ t('admin.risk.visuals.avgConcentration') }}</p>
          <p class="mt-2 text-lg font-semibold text-gray-900 dark:text-white">{{ avgConcentration.toFixed(3) }}</p>
        </div>
      </div>

      <div class="mt-6 rounded-2xl border border-gray-100 p-4 dark:border-dark-800">
        <div class="flex items-center justify-between gap-3">
          <div>
            <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.risk.visuals.activityTimeline') }}</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">00:00 - 23:00</p>
          </div>
          <p class="text-sm font-medium text-gray-600 dark:text-gray-300">{{ peakHourLabel }}</p>
        </div>

        <div class="mt-4 rounded-2xl bg-slate-50 px-4 py-4 dark:bg-dark-900/70">
          <div class="overflow-x-auto pb-1">
            <div class="relative min-w-[960px]">
              <svg class="pointer-events-none absolute inset-0 h-full w-full" viewBox="0 0 240 40" preserveAspectRatio="none">
                <polyline
                  fill="none"
                  stroke="url(#riskActivityGradient)"
                  stroke-width="2.6"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  :points="trendPoints"
                />
                <defs>
                  <linearGradient id="riskActivityGradient" x1="0%" x2="100%" y1="0%" y2="0%">
                    <stop offset="0%" stop-color="#38bdf8" />
                    <stop offset="100%" stop-color="#8b5cf6" />
                  </linearGradient>
                </defs>
              </svg>

              <div class="relative grid gap-2" :style="{ gridTemplateColumns: 'repeat(24, minmax(0, 1fr))' }">
                <div v-for="point in timeline" :key="point.hour" class="flex flex-col items-center gap-2">
                  <div class="h-16 w-1 rounded-full bg-gray-200 dark:bg-dark-800">
                    <div class="w-full rounded-full bg-gradient-to-t from-violet-500 to-sky-500 transition-all duration-300" :style="{ height: `${point.height}%`, marginTop: `${100 - point.height}%` }"></div>
                  </div>
                  <div class="text-center">
                    <p class="text-[11px] font-medium text-gray-500 dark:text-dark-400">{{ point.label }}</p>
                    <p class="text-[11px] text-gray-400 dark:text-dark-500">{{ point.count }}</p>
                  </div>
                </div>
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
import HelpTooltip from '@/components/common/HelpTooltip.vue'
import Icon from '@/components/icons/Icon.vue'

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

const timeline = computed(() => {
  return countsByHour.value.map((count, hour) => {
    return {
      hour,
      count,
      label: String(hour).padStart(2, '0'),
      height: maxHourCount.value ? Math.max(8, (count / maxHourCount.value) * 100) : 8
    }
  })
})

const allDayUsers = computed(() => props.rows.filter((row) => row.metrics.all_day_active).length)
const avgConcentration = computed(() => {
  if (!props.rows.length) return 0
  return props.rows.reduce((sum, row) => sum + row.metrics.hour_concentration, 0) / props.rows.length
})

const avgActiveWindow = computed(() => {
  if (!props.rows.length) return 0
  return props.rows.reduce((sum, row) => sum + (row.metrics.active_hours || []).length, 0) / props.rows.length
})

const peakHourLabel = computed(() => {
  const peak = timeline.value.reduce((best, current) => (current.count > best.count ? current : best), timeline.value[0])
  return `${String(peak.hour).padStart(2, '0')}:00 · ${peak.count}`
})

const trendPoints = computed(() => {
  return timeline.value
    .map((point, index) => {
      const x = timeline.value.length === 1 ? 120 : (index / (timeline.value.length - 1)) * 240
      const y = 36 - ((point.count / maxHourCount.value) * 28)
      return `${x},${y}`
    })
    .join(' ')
})
</script>
