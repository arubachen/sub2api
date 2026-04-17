import { describe, expect, it, beforeEach, afterEach, vi } from 'vitest'
import { DEFAULT_SITE_NAME } from '@/constants/branding'

describe('resolveDocumentTitle', () => {
  beforeEach(() => {
    vi.stubGlobal('localStorage', {
      getItem: vi.fn(() => null),
      setItem: vi.fn(),
      removeItem: vi.fn(),
      clear: vi.fn()
    })
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    vi.resetModules()
  })

  it('路由存在标题时，使用“路由标题 - 站点名”格式', async () => {
    const { resolveDocumentTitle } = await import('@/router/title')
    expect(resolveDocumentTitle('Usage Records', 'My Site')).toBe('Usage Records - My Site')
  })

  it('路由无标题时，回退到站点名', async () => {
    const { resolveDocumentTitle } = await import('@/router/title')
    expect(resolveDocumentTitle(undefined, 'My Site')).toBe('My Site')
  })

  it('站点名为空时，回退默认站点名', async () => {
    const { resolveDocumentTitle } = await import('@/router/title')
    expect(resolveDocumentTitle('Dashboard', '')).toBe(`Dashboard - ${DEFAULT_SITE_NAME}`)
    expect(resolveDocumentTitle(undefined, '   ')).toBe(DEFAULT_SITE_NAME)
  })

  it('站点名变更时仅影响后续路由标题计算', async () => {
    const { resolveDocumentTitle } = await import('@/router/title')
    const before = resolveDocumentTitle('Admin Dashboard', 'Alpha')
    const after = resolveDocumentTitle('Admin Dashboard', 'Beta')

    expect(before).toBe('Admin Dashboard - Alpha')
    expect(after).toBe('Admin Dashboard - Beta')
  })
})
