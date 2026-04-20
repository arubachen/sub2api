<template>
  <div class="grid gap-4 lg:grid-cols-4">
    <article
      v-for="card in primaryCards"
      :key="card.key"
      class="group relative overflow-hidden rounded-[30px] border border-white/70 bg-white/88 p-5 shadow-[0_24px_60px_-36px_rgba(15,23,42,0.18)] backdrop-blur-xl transition-all duration-300 hover:-translate-y-0.5 hover:shadow-[0_28px_70px_-34px_rgba(8,47,73,0.22)] dark:border-dark-800 dark:bg-dark-950/82 dark:shadow-[0_24px_60px_-36px_rgba(2,6,23,0.6)] dark:hover:shadow-[0_30px_80px_-34px_rgba(6,182,212,0.18)]"
    >
      <div class="pointer-events-none absolute inset-0 opacity-0 transition-opacity duration-300 group-hover:opacity-100" :class="card.hoverGlow"></div>
      <div class="pointer-events-none absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-white/80 to-transparent dark:via-white/20"></div>

      <div class="relative flex items-start justify-between gap-4">
        <div class="min-w-0">
          <p class="text-xs font-semibold uppercase tracking-[0.22em] text-gray-400 dark:text-dark-400">{{ card.label }}</p>
          <p class="mt-4 text-[2rem] font-semibold tracking-tight text-gray-900 dark:text-white">{{ card.value }}</p>
          <p class="mt-2 text-xs leading-5 text-gray-500 dark:text-dark-400">{{ card.meta }}</p>
        </div>

        <div :class="['flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-2xl ring-1 ring-inset', card.iconWrap]">
          <Icon :name="card.icon" size="lg" :class="card.iconColor" :stroke-width="2" />
        </div>
      </div>
    </article>
  </div>

  <div class="mt-4 grid gap-4 lg:grid-cols-4">
    <article
      v-for="card in secondaryCards"
      :key="card.key"
      class="group relative overflow-hidden rounded-[28px] border border-white/70 bg-white/82 p-5 shadow-[0_20px_48px_-34px_rgba(15,23,42,0.16)] backdrop-blur-xl transition-all duration-300 hover:border-primary-100 hover:shadow-[0_24px_60px_-34px_rgba(8,47,73,0.2)] dark:border-dark-800 dark:bg-dark-950/78 dark:hover:border-primary-900/40 dark:hover:shadow-[0_24px_70px_-38px_rgba(6,182,212,0.16)]"
    >
      <div class="pointer-events-none absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-primary-200/80 to-transparent dark:via-primary-500/20"></div>
      <div class="relative flex items-start gap-4">
        <div :class="['mt-0.5 flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-2xl ring-1 ring-inset', card.iconWrap]">
          <Icon :name="card.icon" size="md" :class="card.iconColor" :stroke-width="2" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-xs font-semibold uppercase tracking-[0.22em] text-gray-400 dark:text-dark-400">{{ card.label }}</p>
          <div class="mt-2 flex items-baseline gap-2">
            <p class="text-[1.75rem] font-semibold tracking-tight text-gray-900 dark:text-white">{{ card.value }}</p>
            <span v-if="card.trailing" class="text-xs font-medium text-primary-600 dark:text-primary-400">{{ card.trailing }}</span>
          </div>
          <p class="mt-2 text-xs leading-5 text-gray-500 dark:text-dark-400">{{ card.meta }}</p>
        </div>
      </div>
    </article>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { UserDashboardStats as UserStatsType } from '@/api/usage'

const props = defineProps<{
  stats: UserStatsType
  balance: number
  isSimple: boolean
}>()

const { t } = useI18n()

const formatBalance = (b: number) =>
  new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  }).format(b)

const formatNumber = (n: number) => n.toLocaleString()
const formatCost = (c: number) => c.toFixed(4)
const formatTokens = (value: number) => {
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M`
  if (value >= 1000) return `${(value / 1000).toFixed(1)}K`
  return value.toString()
}
const formatDuration = (ms: number) => (ms >= 1000 ? `${(ms / 1000).toFixed(2)}s` : `${ms.toFixed(0)}ms`)


type DashboardIconName = 'creditCard' | 'key' | 'chart' | 'dollar' | 'cube' | 'database' | 'bolt' | 'clock'
type DashboardCard = {
  key: string
  icon: DashboardIconName
  label: string
  value: string
  meta: string
  iconWrap: string
  iconColor: string
  hoverGlow?: string
  trailing?: string
}

const primaryCards = computed<DashboardCard[]>(() => {
  const cards: Array<DashboardCard | null> = [
    !props.isSimple
      ? {
          key: 'balance',
          icon: 'creditCard',
          label: t('dashboard.balance'),
          value: `$${formatBalance(props.balance)}`,
          meta: t('common.available'),
          iconWrap: 'bg-blue-50 text-blue-700 ring-blue-100 dark:bg-blue-950/35 dark:text-blue-300 dark:ring-blue-900/40',
          iconColor: 'text-blue-700 dark:text-blue-300',
          hoverGlow: 'bg-[radial-gradient(circle_at_top_right,rgba(37,99,235,0.16),transparent_55%)] dark:bg-[radial-gradient(circle_at_top_right,rgba(96,165,250,0.18),transparent_55%)]'
        }
      : null,
    {
      key: 'api-keys',
      icon: 'key',
      label: t('dashboard.apiKeys'),
      value: String(props.stats?.total_api_keys || 0),
      meta: `${props.stats?.active_api_keys || 0} ${t('common.active')}`,
      iconWrap: 'bg-slate-100 text-slate-700 ring-slate-200 dark:bg-dark-900 dark:text-slate-200 dark:ring-dark-800',
      iconColor: 'text-slate-700 dark:text-slate-200',
      hoverGlow: 'bg-[radial-gradient(circle_at_top_right,rgba(148,163,184,0.12),transparent_55%)] dark:bg-[radial-gradient(circle_at_top_right,rgba(148,163,184,0.14),transparent_55%)]'
    },
    {
      key: 'today-requests',
      icon: 'chart',
      label: t('dashboard.todayRequests'),
      value: String(props.stats?.today_requests || 0),
      meta: `${t('common.total')}: ${formatNumber(props.stats?.total_requests || 0)}`,
      iconWrap: 'bg-amber-50 text-amber-700 ring-amber-100 dark:bg-amber-950/30 dark:text-amber-300 dark:ring-amber-900/30',
      iconColor: 'text-amber-700 dark:text-amber-300',
      hoverGlow: 'bg-[radial-gradient(circle_at_top_right,rgba(245,158,11,0.16),transparent_55%)] dark:bg-[radial-gradient(circle_at_top_right,rgba(251,191,36,0.18),transparent_55%)]'
    },
    {
      key: 'today-cost',
      icon: 'dollar',
      label: t('dashboard.todayCost'),
      value: `$${formatCost(props.stats?.today_actual_cost || 0)}`,
      meta: `${t('common.total')}: $${formatCost(props.stats?.total_actual_cost || 0)} • ${t('dashboard.standard')}: $${formatCost(props.stats?.today_cost || 0)}`,
      iconWrap: 'bg-violet-50 text-violet-700 ring-violet-100 dark:bg-violet-950/30 dark:text-violet-300 dark:ring-violet-900/30',
      iconColor: 'text-violet-700 dark:text-violet-300',
      hoverGlow: 'bg-[radial-gradient(circle_at_top_right,rgba(124,58,237,0.15),transparent_55%)] dark:bg-[radial-gradient(circle_at_top_right,rgba(167,139,250,0.18),transparent_55%)]'
    }
  ]

  return cards.filter((card): card is NonNullable<(typeof cards)[number]> => Boolean(card))
})

const secondaryCards = computed<DashboardCard[]>(() => {
  const cards: DashboardCard[] = [
    {
      key: 'today-tokens',
      icon: 'cube',
      label: t('dashboard.todayTokens'),
      value: formatTokens(props.stats?.today_tokens || 0),
      trailing: undefined,
      meta: `${t('dashboard.input')}: ${formatTokens(props.stats?.today_input_tokens || 0)} / ${t('dashboard.output')}: ${formatTokens(props.stats?.today_output_tokens || 0)}`,
      iconWrap: 'bg-slate-100 text-slate-700 ring-slate-200 dark:bg-dark-900 dark:text-slate-200 dark:ring-dark-800',
      iconColor: 'text-slate-700 dark:text-slate-200'
    },
    {
      key: 'total-tokens',
      icon: 'database',
      label: t('dashboard.totalTokens'),
      value: formatTokens(props.stats?.total_tokens || 0),
      trailing: undefined,
      meta: `${t('dashboard.input')}: ${formatTokens(props.stats?.total_input_tokens || 0)} / ${t('dashboard.output')}: ${formatTokens(props.stats?.total_output_tokens || 0)}`,
      iconWrap: 'bg-slate-100 text-slate-700 ring-slate-200 dark:bg-dark-900 dark:text-slate-200 dark:ring-dark-800',
      iconColor: 'text-slate-700 dark:text-slate-200'
    },
    {
      key: 'performance',
      icon: 'bolt',
      label: t('dashboard.performance'),
      value: formatTokens(props.stats?.rpm || 0),
      trailing: 'RPM',
      meta: `${formatTokens(props.stats?.tpm || 0)} TPM`,
      iconWrap: 'bg-blue-50 text-blue-700 ring-blue-100 dark:bg-blue-950/35 dark:text-blue-300 dark:ring-blue-900/40',
      iconColor: 'text-blue-700 dark:text-blue-300'
    },
    {
      key: 'avg-response',
      icon: 'clock',
      label: t('dashboard.avgResponse'),
      value: formatDuration(props.stats?.average_duration_ms || 0),
      trailing: undefined,
      meta: t('dashboard.averageTime'),
      iconWrap: 'bg-slate-100 text-slate-700 ring-slate-200 dark:bg-dark-900 dark:text-slate-200 dark:ring-dark-800',
      iconColor: 'text-slate-700 dark:text-slate-200'
    }
    ]

  return cards
})
</script>
