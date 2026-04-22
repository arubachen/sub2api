<template>
  <button
    type="button"
    @click="toggleTheme"
    :class="[
      'inline-flex items-center justify-center rounded-full border border-gray-200/90 bg-white/90 p-2 text-slate-700 shadow-sm transition-all hover:border-gray-300 hover:bg-white hover:text-slate-950 dark:border-white/10 dark:bg-white/5 dark:text-slate-200 dark:hover:border-white/20 dark:hover:bg-white/10 dark:hover:text-white',
      buttonClass,
    ]"
    :title="toggleTitle"
    :aria-label="toggleTitle"
  >
    <Icon v-if="isDark" name="sun" size="md" class="text-amber-500" />
    <Icon v-else name="moon" size="md" />
    <span v-if="showLabel" class="ml-2 text-sm font-medium">
      {{ toggleLabel }}
    </span>
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
  buttonClass: '',
})

const { t } = useI18n()
const { isDark, toggleTheme } = useTheme()

const toggleLabel = computed(() => (isDark.value ? t('nav.lightMode') : t('nav.darkMode')))
const toggleTitle = computed(() =>
  isDark.value ? t('home.switchToLight') : t('home.switchToDark'),
)
</script>
