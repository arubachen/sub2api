import { ref, onMounted, onBeforeUnmount } from 'vue'

const THEME_STORAGE_KEY = 'theme'
const THEME_EVENT = 'sub2api-theme-change'

export type ThemeMode = 'light' | 'dark' | 'system'

function normalizeThemeMode(value: string | null): ThemeMode {
  if (value === 'light' || value === 'dark' || value === 'system') {
    return value
  }
  return 'system'
}

function resolveThemeMode(mode: ThemeMode) {
  return mode === 'system'
    ? window.matchMedia('(prefers-color-scheme: dark)').matches
    : mode === 'dark'
}

export function useTheme() {
  const isDark = ref(document.documentElement.classList.contains('dark'))
  const themeMode = ref<ThemeMode>(
    typeof window === 'undefined' ? 'system' : normalizeThemeMode(localStorage.getItem(THEME_STORAGE_KEY))
  )

  const applyTheme = (dark: boolean) => {
    isDark.value = dark
    document.documentElement.classList.toggle('dark', dark)
  }

  const syncTheme = () => {
    if (typeof window === 'undefined') return
    themeMode.value = normalizeThemeMode(localStorage.getItem(THEME_STORAGE_KEY))
    applyTheme(resolveThemeMode(themeMode.value))
  }

  const setThemeMode = (mode: ThemeMode) => {
    themeMode.value = mode
    localStorage.setItem(THEME_STORAGE_KEY, mode)
    const dark = resolveThemeMode(mode)
    applyTheme(dark)
    window.dispatchEvent(new CustomEvent(THEME_EVENT, { detail: { dark, mode } }))
  }

  const cycleThemeMode = () => {
    const nextMode: ThemeMode =
      themeMode.value === 'system'
        ? 'light'
        : themeMode.value === 'light'
          ? 'dark'
          : 'system'
    setThemeMode(nextMode)
  }

  const handleThemeChange = (event?: Event) => {
    const customEvent = event as CustomEvent<{ dark?: boolean; mode?: ThemeMode }> | undefined
    if (customEvent?.detail?.mode) {
      themeMode.value = customEvent.detail.mode
    }
    if (typeof customEvent?.detail?.dark === 'boolean') {
      applyTheme(customEvent.detail.dark)
      return
    }
    if (themeMode.value === 'system') {
      syncTheme()
    }
  }

  let mediaQuery: MediaQueryList | null = null

  onMounted(() => {
    syncTheme()
    window.addEventListener(THEME_EVENT, handleThemeChange)

    mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    mediaQuery.addEventListener('change', handleThemeChange)
  })

  onBeforeUnmount(() => {
    window.removeEventListener(THEME_EVENT, handleThemeChange)
    mediaQuery?.removeEventListener('change', handleThemeChange)
  })

  return {
    isDark,
    themeMode,
    setThemeMode,
    cycleThemeMode,
    syncTheme
  }
}
