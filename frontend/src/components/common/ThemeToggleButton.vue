<template>
  <button
    type="button"
    @click="cycleThemeMode"
    :class="[
      'inline-flex items-center justify-center rounded-full border border-gray-200/90 bg-white/90 p-2 text-slate-700 shadow-sm transition-all hover:border-gray-300 hover:bg-white hover:text-slate-950 dark:border-white/10 dark:bg-white/5 dark:text-slate-200 dark:hover:border-white/20 dark:hover:bg-white/10 dark:hover:text-white',
      buttonClass
    ]"
    :title="themeModeLabel"
    :aria-label="themeModeLabel"
  >
    <span class="theme-toggle-icon" v-html="themeIconSvg"></span>
    <span v-if="showLabel" class="ml-2 text-sm font-medium">{{ themeModeLabel }}</span>
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useTheme } from '@/composables/useTheme'
import lightThemeIcon from '@/assets/theme-toggle/light.svg?raw'
import darkThemeIcon from '@/assets/theme-toggle/dark.svg?raw'
import systemThemeIcon from '@/assets/theme-toggle/system.svg?raw'

withDefaults(defineProps<{
  showLabel?: boolean
  buttonClass?: string
}>(), {
  showLabel: false,
  buttonClass: ''
})

const { t } = useI18n()
const { themeMode, cycleThemeMode } = useTheme()

const themeModeLabel = computed(() => {
  if (themeMode.value === 'system') return t('home.themeModeSystem')
  return themeMode.value === 'dark' ? t('home.themeModeDark') : t('home.themeModeLight')
})

const themeIconSvg = computed(() => {
  const svg =
    themeMode.value === 'system'
      ? systemThemeIcon
      : themeMode.value === 'dark'
        ? darkThemeIcon
        : lightThemeIcon

  return svg
    .replace('<svg ', '<svg fill="currentColor" aria-hidden="true" ')
    .replace(/<path /g, '<path fill="currentColor" ')
})
</script>

<style scoped>
.theme-toggle-icon {
  display: inline-block;
  flex-shrink: 0;
  line-height: 0;
}

.theme-toggle-icon :deep(svg) {
  display: block;
  width: 1.25rem;
  height: 1.25rem;
  fill: currentColor;
}

.theme-toggle-icon :deep(path) {
  fill: currentColor;
}
</style>
