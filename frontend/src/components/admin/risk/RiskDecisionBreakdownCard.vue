<template>
  <section class="rounded-[28px] border border-gray-200/80 bg-white/95 p-5 shadow-card dark:border-dark-800 dark:bg-dark-900/90">
    <div class="flex items-start justify-between gap-4">
      <div>
        <p class="text-xs font-semibold uppercase tracking-[0.24em] text-gray-400 dark:text-dark-500">{{ t('admin.risk.visuals.decisionEyebrow') }}</p>
        <h3 class="mt-2 text-lg font-semibold text-gray-900 dark:text-white">{{ t('admin.risk.visuals.decisionTitle') }}</h3>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.risk.visuals.decisionHint') }}</p>
      </div>
      <span class="rounded-full bg-gray-100 px-3 py-1 text-xs font-medium text-gray-600 dark:bg-dark-800 dark:text-dark-300">
        {{ rows.length }}
      </span>
    </div>

    <div v-if="rows.length === 0" class="mt-8 rounded-2xl border border-dashed border-gray-200 px-4 py-10 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
      {{ t('common.noData') }}
    </div>

    <div v-else class="mt-6 grid gap-6 lg:grid-cols-[180px_minmax(0,1fr)] lg:items-center">
      <div class="mx-auto flex flex-col items-center gap-3">
        <div class="relative flex h-40 w-40 items-center justify-center rounded-full" :style="ringStyle">
          <div class="flex h-[112px] w-[112px] flex-col items-center justify-center rounded-full bg-white text-center dark:bg-dark-900">
            <span class="text-xs font-medium uppercase tracking-[0.18em] text-gray-400 dark:text-dark-500">{{ t('admin.risk.visuals.dominantDecision') }}</span>
            <span class="mt-2 text-sm font-semibold text-gray-900 dark:text-white">{{ dominant.label }}</span>
            <span class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ dominant.count }} / {{ rows.length }}</span>
          </div>
        </div>
      </div>

      <div class="space-y-3">
        <div v-for="item in breakdown" :key="item.key" class="rounded-2xl border border-gray-100 p-3 dark:border-dark-800">
          <div class="flex items-center justify-between gap-3">
            <div class="flex items-center gap-3">
              <span class="h-3 w-3 rounded-full" :style="{ backgroundColor: item.color }"></span>
              <div>
                <p class="text-sm font-medium text-gray-900 dark:text-white">{{ item.label }}</p>
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.risk.visuals.avgScore') }} {{ item.avgScore.toFixed(1) }}</p>
              </div>
            </div>
            <div class="text-right">
              <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ item.count }}</p>
              <p class="text-xs text-gray-500 dark:text-gray-400">{{ item.percent.toFixed(0) }}%</p>
            </div>
          </div>
          <div class="mt-3 h-2 overflow-hidden rounded-full bg-gray-100 dark:bg-dark-800">
            <div class="h-full rounded-full transition-all duration-300" :style="{ width: `${item.percent}%`, backgroundColor: item.color }"></div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

interface RowLike {
  summary: {
    decision: string
    decision_label: string
    risk_score: number
  }
}

const props = defineProps<{ rows: RowLike[] }>()
const { t } = useI18n()

const decisionColors: Record<string, string> = {
  freeze_review: '#ef4444',
  throttle: '#f97316',
  review: '#f59e0b',
  observe: '#eab308',
  normal: '#10b981'
}

const breakdown = computed(() => {
  const order = ['freeze_review', 'throttle', 'review', 'observe', 'normal']
  const counts = new Map<string, { count: number; totalScore: number; label: string }>()
  for (const row of props.rows) {
    const existing = counts.get(row.summary.decision) || { count: 0, totalScore: 0, label: row.summary.decision_label }
    existing.count += 1
    existing.totalScore += row.summary.risk_score
    existing.label = row.summary.decision_label || existing.label
    counts.set(row.summary.decision, existing)
  }
  return order
    .filter((key) => counts.has(key))
    .map((key) => {
      const item = counts.get(key)!
      const count = item.count
      const percent = props.rows.length ? (count / props.rows.length) * 100 : 0
      return {
        key,
        label: item.label,
        count,
        percent,
        avgScore: count > 0 ? item.totalScore / count : 0,
        color: decisionColors[key] || '#94a3b8'
      }
    })
    .sort((left, right) => {
      if (left.count !== right.count) return right.count - left.count
      return right.avgScore - left.avgScore
    })
})

const dominant = computed(() => breakdown.value[0] || { label: t('common.noData'), count: 0 })

const ringStyle = computed(() => {
  if (!breakdown.value.length) return {}
  const segments: string[] = []
  let offset = 0
  for (const item of breakdown.value) {
    const next = offset + item.percent
    segments.push(`${item.color} ${offset}% ${next}%`)
    offset = next
  }
  return {
    background: `conic-gradient(${segments.join(', ')})`
  }
})
</script>
