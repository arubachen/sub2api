import { ref } from 'vue'

const THEME_STORAGE_KEY = 'theme'
export const THEME_EVENT = 'sub2api-theme-change'

export type ThemeMode = 'light' | 'dark' | 'system'

function normalizeThemeMode(value: string | null): ThemeMode {
  if (value === 'light' || value === 'dark' || value === 'system') {
    return value
  }
  return 'system'
}

function resolveThemeMode(mode: ThemeMode) {
  return mode === 'system'
    ? typeof window !== 'undefined' && window.matchMedia('(prefers-color-scheme: dark)').matches
    : mode === 'dark'
}

const themeMode = ref<ThemeMode>(
  typeof window === 'undefined'
    ? 'system'
    : normalizeThemeMode(localStorage.getItem(THEME_STORAGE_KEY)),
)

const isDark = ref(
  typeof document !== 'undefined' && document.documentElement.classList.contains('dark'),
)

let mediaQuery: MediaQueryList | null = null
let mediaQueryBound = false

function applyTheme(dark: boolean) {
  if (typeof document === 'undefined') return
  isDark.value = dark
  document.documentElement.classList.toggle('dark', dark)
}

function emitThemeChange() {
  if (typeof window === 'undefined') return
  window.dispatchEvent(new CustomEvent(THEME_EVENT, { detail: { dark: isDark.value, mode: themeMode.value } }))
}

function syncTheme(emit = false) {
  if (typeof window === 'undefined') return
  themeMode.value = normalizeThemeMode(localStorage.getItem(THEME_STORAGE_KEY))
  applyTheme(resolveThemeMode(themeMode.value))
  if (emit) emitThemeChange()
}

function setThemeMode(mode: ThemeMode) {
  if (typeof window === 'undefined') return
  themeMode.value = mode
  localStorage.setItem(THEME_STORAGE_KEY, mode)
  applyTheme(resolveThemeMode(mode))
  emitThemeChange()
}

function cycleThemeMode() {
  const nextMode: ThemeMode =
    themeMode.value === 'system'
      ? 'light'
      : themeMode.value === 'light'
        ? 'dark'
        : 'system'

  setThemeMode(nextMode)
}

function handleSystemThemeChange() {
  if (themeMode.value !== 'system') return
  applyTheme(resolveThemeMode('system'))
  emitThemeChange()
}

function ensureThemeSync() {
  if (typeof window === 'undefined') return

  syncTheme()

  if (mediaQueryBound) return

  mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
  if (typeof mediaQuery.addEventListener === 'function') {
    mediaQuery.addEventListener('change', handleSystemThemeChange)
  } else {
    mediaQuery.addListener(handleSystemThemeChange)
  }
  mediaQueryBound = true
}

export function useTheme() {
  ensureThemeSync()

  return {
    isDark,
    themeMode,
    setThemeMode,
    cycleThemeMode,
    syncTheme,
  }
}
