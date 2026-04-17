<template>
  <button
    type="button"
    @click="toggleTheme"
    :class="[
      'inline-flex items-center justify-center rounded-full border border-white/10 bg-white/5 p-2 text-slate-300 transition-all hover:border-white/20 hover:bg-white/10 hover:text-white dark:border-white/10 dark:bg-white/5 dark:text-slate-300 dark:hover:border-white/20 dark:hover:bg-white/10 dark:hover:text-white',
      buttonClass
    ]"
    :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
    :aria-label="isDark ? t('home.switchToLight') : t('home.switchToDark')"
  >
    <Icon v-if="isDark" name="sun" size="md" />
    <Icon v-else name="moon" size="md" />
    <span v-if="showLabel" class="ml-2 text-sm font-medium">{{ isDark ? t('home.switchToLight') : t('home.switchToDark') }}</span>
  </button>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useTheme } from '@/composables/useTheme'
import Icon from '@/components/icons/Icon.vue'

withDefaults(defineProps<{
  showLabel?: boolean
  buttonClass?: string
}>(), {
  showLabel: false,
  buttonClass: ''
})

const { t } = useI18n()
const { isDark, toggleTheme } = useTheme()
</script>
