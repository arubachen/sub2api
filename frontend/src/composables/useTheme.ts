import { ref, onMounted, onBeforeUnmount } from 'vue'

const THEME_STORAGE_KEY = 'theme'
const THEME_EVENT = 'sub2api-theme-change'

function getPreferredDarkMode() {
  const savedTheme = localStorage.getItem(THEME_STORAGE_KEY)
  return savedTheme === 'dark' || (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
}

export function useTheme() {
  const isDark = ref(document.documentElement.classList.contains('dark'))

  const applyTheme = (dark: boolean) => {
    isDark.value = dark
    document.documentElement.classList.toggle('dark', dark)
  }

  const syncTheme = () => {
    applyTheme(getPreferredDarkMode())
  }

  const setTheme = (dark: boolean) => {
    localStorage.setItem(THEME_STORAGE_KEY, dark ? 'dark' : 'light')
    applyTheme(dark)
    window.dispatchEvent(new CustomEvent(THEME_EVENT, { detail: { dark } }))
  }

  const toggleTheme = () => {
    setTheme(!isDark.value)
  }

  const handleThemeChange = (event?: Event) => {
    const customEvent = event as CustomEvent<{ dark?: boolean }> | undefined
    if (typeof customEvent?.detail?.dark === 'boolean') {
      applyTheme(customEvent.detail.dark)
      return
    }
    syncTheme()
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
    setTheme,
    toggleTheme,
    syncTheme
  }
}
