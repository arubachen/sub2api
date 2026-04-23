import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../ThemeToggleButton.vue')
const componentSource = readFileSync(componentPath, 'utf8')
const composablePath = resolve(dirname(fileURLToPath(import.meta.url)), '../../../composables/useTheme.ts')
const composableSource = readFileSync(composablePath, 'utf8')

describe('ThemeToggleButton theme-mode cycle', () => {
  it('restores the three-state theme icons and labels', () => {
    expect(componentSource).toContain("light.svg?raw")
    expect(componentSource).toContain("dark.svg?raw")
    expect(componentSource).toContain("system.svg?raw")
    expect(componentSource).toContain('cycleThemeMode')
    expect(componentSource).toContain("themeMode.value === 'system'")
  })

  it('keeps system theme mode support in the shared composable', () => {
    expect(composableSource).toContain("export type ThemeMode = 'light' | 'dark' | 'system'")
    expect(composableSource).toContain("window.matchMedia('(prefers-color-scheme: dark)')")
    expect(composableSource).toContain("localStorage.setItem(THEME_STORAGE_KEY, mode)")
  })
})
