<template>
  <div class="card overflow-hidden">
    <div class="border-b border-gray-100 px-6 py-5 dark:border-dark-800">
      <p class="text-xs font-semibold uppercase tracking-[0.26em] text-gray-400 dark:text-dark-400">Workspace</p>
      <div class="mt-3 flex items-start justify-between gap-4">
        <div>
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('dashboard.quickActions') }}</h2>
          <p class="mt-1 text-sm leading-6 text-gray-500 dark:text-dark-400">{{ t('dashboard.welcomeMessage') }}</p>
        </div>
        <div class="rounded-full border border-primary-100 bg-primary-50 px-3 py-1 text-xs font-semibold text-primary-700 dark:border-primary-900/40 dark:bg-primary-950/30 dark:text-primary-300">
          {{ actions.length }} {{ t('common.available') }}
        </div>
      </div>
    </div>

    <div class="space-y-3 p-4">
      <button
        v-for="action in actions"
        :key="action.key"
        @click="router.push(action.to)"
        :class="[
          'group flex w-full items-center gap-4 rounded-[24px] border p-4 text-left transition-all duration-200 hover:-translate-y-0.5',
          action.wrapperClass
        ]"
      >
        <div :class="['flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-2xl ring-1 ring-inset transition-transform group-hover:scale-105', action.iconWrapClass]">
          <Icon :name="action.icon" size="lg" :class="action.iconClass" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ action.title }}</p>
          <p class="mt-1 text-xs leading-5 text-gray-500 dark:text-dark-400">{{ action.description }}</p>
        </div>
        <Icon
          name="chevronRight"
          size="md"
          :class="['transition-colors', action.arrowClass]"
        />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

const router = useRouter()
const { t } = useI18n()

const actions = computed(
  () => [
    {
      key: 'api-key',
      to: '/keys',
      icon: 'key',
      title: t('dashboard.createApiKey'),
      description: t('dashboard.generateNewKey'),
      wrapperClass:
        'border-primary-100 bg-primary-50/70 hover:border-primary-200 hover:bg-primary-50 dark:border-primary-900/40 dark:bg-primary-950/20 dark:hover:border-primary-900/60 dark:hover:bg-primary-950/30',
      iconWrapClass:
        'bg-white/90 text-primary-600 ring-primary-100 dark:bg-dark-900 dark:text-primary-400 dark:ring-primary-900/40',
      iconClass: 'text-primary-600 dark:text-primary-400',
      arrowClass: 'text-primary-500 dark:text-primary-400'
    },
    {
      key: 'usage',
      to: '/usage',
      icon: 'chart',
      title: t('dashboard.viewUsage'),
      description: t('dashboard.checkDetailedLogs'),
      wrapperClass:
        'border-gray-200/80 bg-white/90 hover:border-slate-300 hover:bg-slate-50 dark:border-dark-800 dark:bg-dark-950/70 dark:hover:border-dark-700 dark:hover:bg-dark-900',
      iconWrapClass: 'bg-slate-100 text-slate-700 ring-slate-200 dark:bg-dark-900 dark:text-slate-200 dark:ring-dark-800',
      iconClass: 'text-slate-700 dark:text-slate-200',
      arrowClass: 'text-gray-400 group-hover:text-slate-700 dark:text-dark-500 dark:group-hover:text-slate-200'
    },
    {
      key: 'redeem',
      to: '/redeem',
      icon: 'gift',
      title: t('dashboard.redeemCode'),
      description: t('dashboard.addBalanceWithCode'),
      wrapperClass:
        'border-gray-200/80 bg-white/90 hover:border-emerald-200 hover:bg-emerald-50/50 dark:border-dark-800 dark:bg-dark-950/70 dark:hover:border-emerald-900/30 dark:hover:bg-emerald-950/15',
      iconWrapClass: 'bg-emerald-50 text-emerald-600 ring-emerald-100 dark:bg-emerald-950/30 dark:text-emerald-400 dark:ring-emerald-900/30',
      iconClass: 'text-emerald-600 dark:text-emerald-400',
      arrowClass: 'text-gray-400 group-hover:text-emerald-600 dark:text-dark-500 dark:group-hover:text-emerald-400'
    }
  ] as const
)
</script>
