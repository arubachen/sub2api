<template>
  <button
    type="button"
    @click="cycleThemeMode"
    :class="[
      'inline-flex items-center justify-center rounded-full border border-white/10 bg-white/5 p-2 text-slate-300 transition-all hover:border-white/20 hover:bg-white/10 hover:text-white dark:border-white/10 dark:bg-white/5 dark:text-slate-300 dark:hover:border-white/20 dark:hover:bg-white/10 dark:hover:text-white',
      buttonClass
    ]"
    :title="themeModeLabel"
    :aria-label="themeModeLabel"
  >
    <Icon v-if="themeMode === 'system'" name="cog" size="md" />
    <Icon v-else-if="isDark" name="moon" size="md" />
    <Icon v-else name="sun" size="md" />
    <span v-if="showLabel" class="ml-2 text-sm font-medium">{{ themeModeLabel }}</span>
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'
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
const { isDark, themeMode, cycleThemeMode } = useTheme()

const themeModeLabel = computed(() => {
  if (themeMode.value === 'system') return t('home.themeModeSystem')
  return themeMode.value === 'dark' ? t('home.themeModeDark') : t('home.themeModeLight')
})
</script>
