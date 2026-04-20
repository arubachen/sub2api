<template>
  <div class="space-y-6">
    <div class="card overflow-hidden">
      <div class="flex flex-wrap items-center gap-4 px-6 py-5">
        <div>
          <p class="text-xs font-semibold uppercase tracking-[0.24em] text-gray-400 dark:text-dark-400">Filters</p>
          <h3 class="mt-2 text-lg font-semibold text-gray-900 dark:text-white">{{ t('dashboard.timeRange') }}</h3>
        </div>
        <button @click="$emit('refresh')" :disabled="loading" class="btn btn-secondary">
          {{ t('common.refresh') }}
        </button>
        <div class="flex flex-wrap items-center gap-4 lg:ml-auto">
          <div class="flex items-center gap-2">
            <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('dashboard.timeRange') }}:</span>
            <DateRangePicker :start-date="startDate" :end-date="endDate" @update:startDate="$emit('update:startDate', $event)" @update:endDate="$emit('update:endDate', $event)" @change="$emit('dateRangeChange', $event)" />
          </div>
          <div class="flex items-center gap-2">
            <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('dashboard.granularity') }}:</span>
            <div class="w-28">
              <Select :model-value="granularity" :options="[{ value: 'day', label: t('dashboard.day') }, { value: 'hour', label: t('dashboard.hour') }]" @update:model-value="$emit('update:granularity', $event)" @change="$emit('granularityChange')" />
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
      <div class="card relative overflow-hidden p-6">
        <div v-if="loading" class="absolute inset-0 z-10 flex items-center justify-center bg-white/55 backdrop-blur-sm dark:bg-dark-950/65">
          <LoadingSpinner size="md" />
        </div>
        <div class="mb-5 flex items-start justify-between gap-4">
          <div>
            <p class="text-xs font-semibold uppercase tracking-[0.24em] text-gray-400 dark:text-dark-400">Distribution</p>
            <h3 class="mt-2 text-lg font-semibold text-gray-900 dark:text-white">{{ t('dashboard.modelDistribution') }}</h3>
          </div>
          <div class="rounded-full border border-primary-100 bg-primary-50/80 px-3 py-1 text-xs font-semibold text-primary-700 dark:border-primary-900/30 dark:bg-primary-950/20 dark:text-primary-300">
            {{ models.length }} models
          </div>
        </div>
        <div class="flex items-center gap-6">
          <div class="h-52 w-52 flex-shrink-0">
            <Doughnut v-if="modelData" :data="modelData" :options="doughnutOptions" />
            <div v-else class="flex h-full items-center justify-center text-sm text-gray-500 dark:text-gray-400">{{ t('dashboard.noDataAvailable') }}</div>
          </div>
          <div class="max-h-56 min-w-0 flex-1 overflow-y-auto rounded-[22px] border border-gray-200/70 bg-slate-50/80 p-3 dark:border-dark-800 dark:bg-dark-900/70">
            <table class="w-full text-xs">
              <thead>
                <tr class="text-gray-500 dark:text-dark-400">
                  <th class="pb-2 text-left">{{ t('dashboard.model') }}</th>
                  <th class="pb-2 text-right">{{ t('dashboard.requests') }}</th>
                  <th class="pb-2 text-right">{{ t('dashboard.tokens') }}</th>
                  <th class="pb-2 text-right">{{ t('dashboard.actual') }}</th>
                  <th class="pb-2 text-right">{{ t('dashboard.standard') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="model in models" :key="model.model" class="border-t border-gray-100 dark:border-dark-800">
                  <td class="max-w-[100px] truncate py-2 font-medium text-gray-900 dark:text-white" :title="model.model">{{ model.model }}</td>
                  <td class="py-2 text-right text-gray-600 dark:text-gray-400">{{ formatNumber(model.requests) }}</td>
                  <td class="py-2 text-right text-gray-600 dark:text-gray-400">{{ formatTokens(model.total_tokens) }}</td>
                  <td class="py-2 text-right text-primary-600 dark:text-primary-400">${{ formatCost(model.actual_cost) }}</td>
                  <td class="py-2 text-right text-gray-400 dark:text-gray-500">${{ formatCost(model.cost) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <TokenUsageTrend :trend-data="trend" :loading="loading" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Select from '@/components/common/Select.vue'
import { Doughnut } from 'vue-chartjs'
import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'
import type { TrendDataPoint, ModelStat } from '@/types'
import { formatCostFixed as formatCost, formatNumberLocaleString as formatNumber, formatTokensK as formatTokens } from '@/utils/format'
import { DISTRIBUTION_CHART_COLORS } from '@/utils/chartPalette'
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, ArcElement, Title, Tooltip, Legend, Filler } from 'chart.js'
ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, ArcElement, Title, Tooltip, Legend, Filler)

const props = defineProps<{ loading: boolean, startDate: string, endDate: string, granularity: string, trend: TrendDataPoint[], models: ModelStat[] }>()
defineEmits(['update:startDate', 'update:endDate', 'update:granularity', 'dateRangeChange', 'granularityChange', 'refresh'])
const { t } = useI18n()

const modelData = computed(() => !props.models?.length ? null : {
  labels: props.models.map((m: ModelStat) => m.model),
  datasets: [{
    data: props.models.map((m: ModelStat) => m.total_tokens),
    backgroundColor: DISTRIBUTION_CHART_COLORS.slice(0, props.models.length)
  }]
})

const doughnutOptions = {
  responsive: true,
  maintainAspectRatio: false,
  cutout: '66%',
  plugins: {
    legend: { display: false },
    tooltip: {
      callbacks: {
        label: (context: any) => `${context.label}: ${formatTokens(context.parsed)} tokens`
      }
    }
  }
}
</script>
