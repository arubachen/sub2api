import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'
import { DEFAULT_SITE_NAME } from '@/constants/branding'

const i18nFixture = vi.hoisted(() => ({
  locale: 'zh',
  messages: {
    zh: {
      home: {
        docs: '文档',
        dashboard: '控制台',
        login: '登录',
        goToDashboard: '进入控制台',
        footer: {
          allRightsReserved: '保留所有权利。'
        },
        juliu: {
          nav: {
            features: '能力',
            flow: '流程',
            models: '模型'
          },
          heroBadge: '聚流',
          fallbackSubtitle: '一站式大模型 API Token 汇聚平台。连接全球顶尖模型，聚沙成塔，汇流成海。',
          primaryCta: '立即开始',
          secondaryCta: '浏览能力',
          scrollCta: '向下浏览',
          capabilityEyebrow: 'Capability',
          capabilityTitle: '把分散的模型入口，汇成一条稳定的生产链路',
          flowEyebrow: 'Flow',
          flowTitle: '从接入、聚合到分发，每一步都更清晰',
          flowDescription: '用统一入口接住 GPT、Claude、Gemini 等上游账号与额度，把路由、鉴权、计费、监控全部收拢到一个后台里。',
          modelsEyebrow: 'Models',
          modelsTitle: '为常用模型提供统一出口',
          modelsDescription: '适合把多账号池、订阅额度和不同上游协议整合进同一套 OpenAI 兼容入口，减少前端与业务层的切换成本。',
          finalEyebrow: 'Start now',
          finalTitle: '准备好把你的 AI 调用入口整合成聚流了吗？',
          finalDescription: '把模型接入、额度管理、用户发放和调用统计汇成一套可运营的系统，直接开始上线使用。',
          finalPrimaryCta: '免费加入',
          tags: {
            unifiedAccess: '统一入口',
            stableRouting: '稳定路由',
            usageBilling: '精细计费'
          },
          features: {
            speed: {
              title: '极速响应',
              desc: '为多上游账号池提供统一入口，把模型请求快速汇总到稳定的 OpenAI 兼容出口。'
            },
            stability: {
              title: '安全稳定',
              desc: '集中管理账号状态、可用额度与调用策略，让高频生产流量更可控。'
            },
            aggregation: {
              title: '模型汇聚',
              desc: '把 GPT、Claude、Gemini 等主流大模型统一到一套接入方式中。'
            },
            analytics: {
              title: '精细统计',
              desc: '把调用量、消耗与用户发放收拢到同一后台，让运营和计费一眼可见。'
            }
          },
          flowSteps: {
            connect: {
              title: '接入上游账号与额度',
              desc: '把订阅、Token、账号池与代理能力接入后台，统一管理不同来源的资源。'
            },
            orchestrate: {
              title: '聚合路由、鉴权与计费',
              desc: '在一个系统里处理鉴权、分组、配额、限流和成本统计，减少额外拼装。'
            },
            serve: {
              title: '对外输出统一 API',
              desc: '给前端、自动化脚本和业务系统提供一致的 OpenAI 风格接口，降低集成复杂度。'
            }
          },
          providers: {
            openaiCompatible: 'OpenAI-Compatible',
            multiAccountPool: '多账号池'
          }
        }
      }
    },
    en: {
      home: {
        docs: 'Docs',
        dashboard: 'Dashboard',
        login: 'Login',
        goToDashboard: 'Go to Dashboard',
        footer: {
          allRightsReserved: 'All rights reserved.'
        },
        juliu: {
          nav: {
            features: 'Capabilities',
            flow: 'Flow',
            models: 'Models'
          },
          heroBadge: '聚流',
          fallbackSubtitle: 'A one-stop AI token hub connecting leading models worldwide. Gather access, unify routing, and scale with confidence.',
          primaryCta: 'Start now',
          secondaryCta: 'Explore capabilities',
          scrollCta: 'Scroll down',
          capabilityEyebrow: 'Capability',
          capabilityTitle: 'Bring scattered model access into one stable production flow',
          flowEyebrow: 'Flow',
          flowTitle: 'From access to orchestration to delivery, every step stays clearer',
          flowDescription: 'Use one entrypoint for GPT, Claude, Gemini, and other upstream accounts, then keep routing, auth, billing, and monitoring in a single control plane.',
          modelsEyebrow: 'Models',
          modelsTitle: 'Offer one unified gateway for everyday models',
          modelsDescription: 'Ideal for combining multi-account pools, subscription quotas, and different upstream protocols behind a single OpenAI-compatible endpoint.',
          finalEyebrow: 'Start now',
          finalTitle: 'Ready to turn your AI access layer into 聚流?',
          finalDescription: 'Bring model onboarding, quota control, user provisioning, and usage analytics into one operational system you can ship with immediately.',
          finalPrimaryCta: 'Join free',
          tags: {
            unifiedAccess: 'Unified access',
            stableRouting: 'Stable routing',
            usageBilling: 'Granular billing'
          },
          features: {
            speed: {
              title: 'Fast response',
              desc: 'Give multi-account upstream pools one shared entry and move requests quickly through a stable OpenAI-compatible gateway.'
            },
            stability: {
              title: 'Safe and reliable',
              desc: 'Centralize account status, available quota, and call policies so high-frequency production traffic stays controllable.'
            },
            aggregation: {
              title: 'Model aggregation',
              desc: 'Unify GPT, Claude, Gemini, and other major model families behind the same integration surface.'
            },
            analytics: {
              title: 'Detailed analytics',
              desc: 'Bring usage, cost, and user distribution into one backend so operations and billing are visible at a glance.'
            }
          },
          flowSteps: {
            connect: {
              title: 'Connect upstream accounts and quota',
              desc: 'Bring subscriptions, tokens, account pools, and proxy capabilities into one backend and manage different sources together.'
            },
            orchestrate: {
              title: 'Unify routing, auth, and billing',
              desc: 'Handle auth, grouping, quotas, rate limits, and cost analytics in one system instead of stitching extra layers together.'
            },
            serve: {
              title: 'Expose one consistent API',
              desc: 'Provide frontend apps, automation scripts, and business systems with a single OpenAI-style interface that is easier to integrate.'
            }
          },
          providers: {
            openaiCompatible: 'OpenAI-Compatible',
            multiAccountPool: 'Multi-account Pool'
          }
        }
      }
    }
  }
}))

function getByPath(source: Record<string, any>, key: string) {
  return key.split('.').reduce<any>((acc, part) => acc?.[part], source)
}

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      locale: {
        get value() {
          return i18nFixture.locale
        }
      },
      t: (key: string) => getByPath(i18nFixture.messages[i18nFixture.locale], key) ?? key
    })
  }
})

const RouterLinkStub = {
  props: ['to'],
  template: '<a :href="typeof to === \'string\' ? to : \'#\'"><slot /></a>'
}

describe('HomeView', () => {
  beforeEach(() => {
    i18nFixture.locale = 'zh'
    setActivePinia(createPinia())
    vi.stubGlobal('localStorage', {
      getItem: vi.fn(() => null),
      setItem: vi.fn(),
      removeItem: vi.fn(),
      clear: vi.fn()
    })
    vi.stubGlobal('matchMedia', vi.fn().mockImplementation(() => ({
      matches: false,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn()
    })))
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('renders the Juliu landing page in Chinese when no custom home content exists', async () => {
    const { useAuthStore } = await import('@/stores/auth')
    const { useAppStore } = await import('@/stores/app')
    const authStore = useAuthStore()
    const appStore = useAppStore()

    authStore.checkAuth = vi.fn()
    appStore.fetchPublicSettings = vi.fn().mockResolvedValue(undefined)
    appStore.publicSettingsLoaded = true
    appStore.siteName = DEFAULT_SITE_NAME
    appStore.siteLogo = ''
    appStore.docUrl = 'https://docs.example.com'
    appStore.cachedPublicSettings = null

    const { default: HomeView } = await import('@/views/HomeView.vue')

    const wrapper = mount(HomeView, {
      global: {
        stubs: {
          ThemeToggleButton: true,
          Icon: true,
          'router-link': RouterLinkStub
        }
      }
    })

    expect(wrapper.text()).toContain('聚流')
    expect(wrapper.text()).not.toContain('聚流 Juliu')
    expect(wrapper.text()).not.toContain('Juliu Token Hub')
    expect(wrapper.text()).toContain('把分散的模型入口，汇成一条稳定的生产链路')
    expect(wrapper.text()).not.toContain('Sora')
    expect(wrapper.text()).not.toContain('Linux DO')
    expect(authStore.checkAuth).toHaveBeenCalled()
  })

  it('renders the localized English copy for the Juliu landing page', async () => {
    i18nFixture.locale = 'en'

    const { useAuthStore } = await import('@/stores/auth')
    const { useAppStore } = await import('@/stores/app')
    const authStore = useAuthStore()
    const appStore = useAppStore()

    authStore.checkAuth = vi.fn()
    appStore.fetchPublicSettings = vi.fn().mockResolvedValue(undefined)
    appStore.publicSettingsLoaded = true
    appStore.siteName = DEFAULT_SITE_NAME
    appStore.siteLogo = ''
    appStore.docUrl = ''
    appStore.cachedPublicSettings = null

    const { default: HomeView } = await import('@/views/HomeView.vue')

    const wrapper = mount(HomeView, {
      global: {
        stubs: {
          ThemeToggleButton: true,
          Icon: true,
          'router-link': RouterLinkStub
        }
      }
    })

    expect(wrapper.text()).toContain('聚流')
    expect(wrapper.text()).not.toContain('Juliu Token Hub')
    expect(wrapper.text()).toContain('Bring scattered model access into one stable production flow')
    expect(wrapper.text()).toContain('Capabilities')
    expect(wrapper.text()).toContain('Multi-account Pool')
    expect(wrapper.text()).not.toContain('Sora')
    expect(wrapper.text()).not.toContain('Linux DO')
  })

  it('preserves admin-defined custom home content', async () => {
    const { useAuthStore } = await import('@/stores/auth')
    const { useAppStore } = await import('@/stores/app')
    const authStore = useAuthStore()
    const appStore = useAppStore()

    authStore.checkAuth = vi.fn()
    appStore.fetchPublicSettings = vi.fn().mockResolvedValue(undefined)
    appStore.publicSettingsLoaded = true
    appStore.cachedPublicSettings = {
      registration_enabled: true,
      email_verify_enabled: false,
      registration_email_suffix_whitelist: [],
      promo_code_enabled: false,
      password_reset_enabled: true,
      invitation_code_enabled: false,
      turnstile_enabled: false,
      turnstile_site_key: '',
      site_name: '聚流',
      site_logo: '',
      site_subtitle: '',
      api_base_url: '',
      contact_info: '',
      doc_url: '',
      home_content: '<div id="custom-home">custom homepage</div>',
      hide_ccs_import_button: false,
      purchase_subscription_enabled: false,
      purchase_subscription_url: '',
      custom_menu_items: [],
      linuxdo_oauth_enabled: false,
      sora_client_enabled: false,
      backend_mode_enabled: false,
      version: ''
    }

    const { default: HomeView } = await import('@/views/HomeView.vue')

    const wrapper = mount(HomeView, {
      global: {
        stubs: {
          ThemeToggleButton: true,
          Icon: true,
          'router-link': RouterLinkStub
        }
      }
    })

    expect(wrapper.find('#custom-home').exists()).toBe(true)
    expect(wrapper.html()).toContain('custom homepage')
  })
})
