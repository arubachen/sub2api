import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../AppSidebar.vue')
const componentSource = readFileSync(componentPath, 'utf8')
const stylePath = resolve(dirname(fileURLToPath(import.meta.url)), '../../../style.css')
const styleSource = readFileSync(stylePath, 'utf8')

describe('AppSidebar custom SVG styles', () => {
  it('does not override uploaded SVG fill or stroke colors', () => {
    expect(componentSource).toContain('.sidebar-svg-icon {')
    expect(componentSource).toContain('color: currentColor;')
    expect(componentSource).toContain('display: block;')
    expect(componentSource).not.toContain('stroke: currentColor;')
    expect(componentSource).not.toContain('fill: none;')
  })
})

describe('AppSidebar header styles', () => {
  it('does not clip the version badge dropdown', () => {
    const sidebarHeaderBlockMatch = styleSource.match(/\.sidebar-header\s*\{[\s\S]*?\n {2}\}/)
    const sidebarBrandBlockMatch = componentSource.match(/\.sidebar-brand\s*\{[\s\S]*?\n\}/)

    expect(sidebarHeaderBlockMatch).not.toBeNull()
    expect(sidebarBrandBlockMatch).not.toBeNull()
    expect(sidebarHeaderBlockMatch?.[0]).not.toContain('@apply overflow-hidden;')
    expect(sidebarBrandBlockMatch?.[0]).not.toContain('overflow: hidden;')
  })
})

describe('AppSidebar shell polish', () => {
  it('uses the Juliu fallback logo asset outside the home page shell', () => {
    expect(componentSource).toContain("siteLogo || '/logo-light.svg'")
  })

  it('does not render a duplicate sidebar theme toggle anymore', () => {
    expect(componentSource).not.toContain('ThemeToggleButton')
  })

  it('uses the provided developer-mode tv svg for the ops icon with shell-colored contrast', () => {
    expect(componentSource).toContain("developer-mode-tv-outline-rounded.svg?raw")
    expect(componentSource).toContain("path: '/admin/ops'")
    expect(componentSource).toContain('iconSvg: opsSidebarIcon')
    expect(componentSource).toContain("iconSvgClass: 'sidebar-shell-icon'")
    expect(componentSource).toContain(".sidebar-shell-icon :deep([fill]:not([fill='none']))")
  })
})
