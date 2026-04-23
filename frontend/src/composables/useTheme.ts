import { ref } from 'vue'

export const THEME_EVENT = 'SUB2API_THEME_CHANGED'

const isDark = ref(
  typeof document !== 'undefined' && document.documentElement.classList.contains('dark'),
)

function syncThemeState() {
  if (typeof document === 'undefined') {
    return
  }
  isDark.value = document.documentElement.classList.contains('dark')
}

function applyTheme(nextIsDark: boolean) {
  if (typeof document === 'undefined') {
    return
  }
  isDark.value = nextIsDark
  document.documentElement.classList.toggle('dark', nextIsDark)
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem('theme', nextIsDark ? 'dark' : 'light')
  }
  window.dispatchEvent(
    new CustomEvent(THEME_EVENT, {
      detail: { dark: nextIsDark },
    }),
  )
}

function toggleTheme() {
  applyTheme(!isDark.value)
}

export function useTheme() {
  syncThemeState()
  return {
    isDark,
    applyTheme,
    toggleTheme,
    syncThemeState,
  }
}
