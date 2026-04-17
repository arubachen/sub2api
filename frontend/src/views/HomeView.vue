<template>
  <div v-if="homeContent" class="min-h-screen">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <div v-else v-html="homeContent"></div>
  </div>

  <div v-else class="juliu-home relative min-h-screen overflow-hidden bg-slate-950 text-white">
    <div class="pointer-events-none absolute inset-0">
      <div class="hero-grid absolute inset-0 opacity-40"></div>
      <div class="hero-noise absolute inset-0 opacity-20"></div>
      <div class="hero-orb hero-orb-cyan absolute left-1/2 top-24 h-[28rem] w-[28rem] -translate-x-1/2 rounded-full"></div>
      <div class="hero-orb hero-orb-emerald absolute bottom-0 right-[-8rem] h-[24rem] w-[24rem] rounded-full"></div>
      <div class="hero-beam absolute left-1/2 top-1/2 h-[22rem] w-[22rem] -translate-x-1/2 -translate-y-1/2 rounded-full"></div>
    </div>

    <header class="fixed inset-x-0 top-0 z-30 border-b border-white/10 bg-slate-950/70 backdrop-blur-xl">
      <nav class="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
        <router-link to="/home" class="flex items-center gap-3">
          <div class="flex h-11 w-11 items-center justify-center rounded-2xl bg-white/5 ring-1 ring-white/10">
            <img
              v-if="siteLogo"
              :src="siteLogo"
              alt="Logo"
              class="h-8 w-8 object-contain"
            />
            <JuliuFlowLogo v-else class="h-8 w-8" />
          </div>
          <div>
            <p class="text-sm font-semibold tracking-[0.28em] text-cyan-300/80">JULIU</p>
            <p class="text-base font-semibold text-white">{{ headerBrandName }}</p>
          </div>
        </router-link>

        <div class="hidden items-center gap-8 text-sm text-slate-300 md:flex">
          <a href="#features" class="transition-colors hover:text-white">{{ t('home.juliu.nav.features') }}</a>
          <a href="#flow" class="transition-colors hover:text-white">{{ t('home.juliu.nav.flow') }}</a>
          <a href="#providers" class="transition-colors hover:text-white">{{ t('home.juliu.nav.models') }}</a>
        </div>

        <div class="flex items-center gap-3">
          <ThemeToggleButton button-class="border-white/15 bg-white/5 text-slate-300 hover:border-white/30 hover:bg-white/10" />
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="hidden text-sm font-medium text-slate-300 transition-colors hover:text-white sm:inline-flex"
          >
            {{ t('home.docs') }}
          </a>
          <router-link
            :to="isAuthenticated ? dashboardPath : '/login'"
            class="inline-flex items-center rounded-full border border-cyan-400/30 bg-cyan-400/10 px-4 py-2 text-sm font-semibold text-cyan-100 transition-all hover:border-cyan-300 hover:bg-cyan-400/20"
          >
            {{ isAuthenticated ? t('home.dashboard') : t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

    <main class="relative z-10">
      <section class="flex min-h-screen items-center justify-center px-6 pb-20 pt-28 text-center">
        <div class="mx-auto flex max-w-5xl flex-col items-center">
          <div class="signal-shell mb-10 flex h-40 w-40 items-center justify-center rounded-full border border-white/10 bg-white/5 shadow-[0_0_120px_rgba(34,211,238,0.14)] backdrop-blur-md md:h-48 md:w-48">
            <img
              v-if="siteLogo"
              :src="siteLogo"
              alt="Logo"
              class="h-24 w-24 object-contain md:h-28 md:w-28"
            />
            <JuliuFlowLogo v-else class="h-24 w-24 md:h-28 md:w-28" />
          </div>

          <p class="mb-5 text-xs font-semibold uppercase tracking-[0.55em] text-cyan-300/75">
            {{ t('home.juliu.heroBadge') }}
          </p>
          <h1 class="max-w-4xl text-5xl font-semibold tracking-tight text-white md:text-7xl">
            {{ brandName }}
          </h1>
          <p class="mt-6 max-w-3xl text-lg leading-8 text-slate-300 md:text-2xl md:leading-10">
            {{ heroSubtitle }}
          </p>

          <div class="mt-10 flex flex-col items-center justify-center gap-4 sm:flex-row">
            <router-link
              :to="isAuthenticated ? dashboardPath : '/login'"
              class="inline-flex items-center justify-center rounded-full bg-cyan-500 px-8 py-4 text-base font-semibold text-slate-950 shadow-[0_0_30px_rgba(6,182,212,0.28)] transition-all hover:bg-cyan-300"
            >
              {{ isAuthenticated ? t('home.goToDashboard') : t('home.juliu.primaryCta') }}
              <Icon name="arrowRight" size="md" class="ml-2" :stroke-width="2" />
            </router-link>

            <a
              :href="secondaryActionHref"
              :target="secondaryActionTarget"
              :rel="secondaryActionTarget ? 'noopener noreferrer' : undefined"
              class="inline-flex items-center justify-center rounded-full border border-white/15 bg-white/5 px-8 py-4 text-base font-semibold text-white transition-all hover:border-white/30 hover:bg-white/10"
            >
              {{ secondaryActionLabel }}
            </a>
          </div>

          <div class="mt-10 flex flex-wrap items-center justify-center gap-3">
            <span
              v-for="tag in valueTags"
              :key="tag"
              class="rounded-full border border-white/12 bg-white/5 px-4 py-2 text-sm text-slate-200 backdrop-blur-sm"
            >
              {{ tag }}
            </span>
          </div>

          <a
            href="#features"
            class="mt-14 inline-flex items-center gap-2 text-sm font-medium text-slate-300 transition-colors hover:text-white"
          >
            {{ t('home.juliu.scrollCta') }}
            <Icon name="arrowDown" size="sm" />
          </a>
        </div>
      </section>

      <section id="features" class="border-t border-white/8 px-6 py-20">
        <div class="mx-auto max-w-6xl">
          <div class="mx-auto max-w-2xl text-center">
            <p class="text-sm font-semibold uppercase tracking-[0.4em] text-cyan-300/70">{{ t('home.juliu.capabilityEyebrow') }}</p>
            <h2 class="mt-4 text-3xl font-semibold text-white md:text-5xl">{{ t('home.juliu.capabilityTitle') }}</h2>
          </div>

          <div class="mt-16 grid gap-10 md:grid-cols-2 xl:grid-cols-4">
            <article
              v-for="feature in features"
              :key="feature.title"
              class="group border-t border-white/12 pt-6"
            >
              <div class="mb-5 inline-flex h-12 w-12 items-center justify-center rounded-2xl border border-white/10 bg-white/5 text-cyan-300 transition-all group-hover:border-cyan-300/40 group-hover:bg-cyan-400/10 group-hover:text-cyan-100">
                <Icon :name="feature.icon" size="lg" />
              </div>
              <h3 class="text-xl font-semibold text-white">{{ feature.title }}</h3>
              <p class="mt-3 text-sm leading-7 text-slate-300">{{ feature.desc }}</p>
            </article>
          </div>
        </div>
      </section>

      <section id="flow" class="border-t border-white/8 px-6 py-20">
        <div class="mx-auto grid max-w-6xl gap-14 lg:grid-cols-[0.9fr_1.1fr] lg:items-start">
          <div>
            <p class="text-sm font-semibold uppercase tracking-[0.4em] text-emerald-300/70">{{ t('home.juliu.flowEyebrow') }}</p>
            <h2 class="mt-4 text-3xl font-semibold text-white md:text-5xl">{{ t('home.juliu.flowTitle') }}</h2>
            <p class="mt-6 max-w-xl text-base leading-8 text-slate-300">
              {{ t('home.juliu.flowDescription') }}
            </p>
          </div>

          <div class="space-y-8">
            <article
              v-for="step in flowSteps"
              :key="step.index"
              class="rounded-3xl border border-white/10 bg-white/5 p-6 backdrop-blur-sm transition-all hover:border-cyan-300/30 hover:bg-white/[0.07]"
            >
              <div class="flex flex-col gap-4 sm:flex-row sm:items-start">
                <span class="text-sm font-semibold tracking-[0.35em] text-cyan-300">{{ step.index }}</span>
                <div>
                  <h3 class="text-xl font-semibold text-white">{{ step.title }}</h3>
                  <p class="mt-3 text-sm leading-7 text-slate-300">{{ step.desc }}</p>
                </div>
              </div>
            </article>
          </div>
        </div>
      </section>

      <section id="providers" class="border-t border-white/8 px-6 py-20">
        <div class="mx-auto max-w-6xl">
          <div class="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
            <div>
              <p class="text-sm font-semibold uppercase tracking-[0.4em] text-cyan-300/70">{{ t('home.juliu.modelsEyebrow') }}</p>
              <h2 class="mt-4 text-3xl font-semibold text-white md:text-5xl">{{ t('home.juliu.modelsTitle') }}</h2>
            </div>
            <p class="max-w-xl text-sm leading-7 text-slate-300">
              {{ t('home.juliu.modelsDescription') }}
            </p>
          </div>

          <div class="mt-12 flex flex-wrap gap-3">
            <span
              v-for="provider in providers"
              :key="provider"
              class="rounded-full border border-white/12 bg-white/5 px-5 py-3 text-sm text-slate-100 backdrop-blur-sm"
            >
              {{ provider }}
            </span>
          </div>
        </div>
      </section>

      <section class="px-6 pb-24 pt-8">
        <div class="mx-auto max-w-6xl overflow-hidden rounded-[2rem] border border-cyan-300/15 bg-[linear-gradient(135deg,rgba(8,47,73,0.95),rgba(15,23,42,0.95)_40%,rgba(6,95,70,0.92))] p-8 shadow-[0_0_80px_rgba(6,182,212,0.12)] md:p-12">
          <div class="flex flex-col gap-10 lg:flex-row lg:items-center lg:justify-between">
            <div class="max-w-2xl">
              <p class="text-sm font-semibold uppercase tracking-[0.4em] text-cyan-200/80">{{ t('home.juliu.finalEyebrow') }}</p>
              <h2 class="mt-4 text-3xl font-semibold text-white md:text-5xl">{{ t('home.juliu.finalTitle') }}</h2>
              <p class="mt-6 text-base leading-8 text-cyan-50/80">
                {{ t('home.juliu.finalDescription') }}
              </p>
            </div>

            <div class="flex flex-col gap-4 sm:flex-row lg:flex-col">
              <router-link
                :to="isAuthenticated ? dashboardPath : '/login'"
                class="inline-flex items-center justify-center rounded-full bg-white px-8 py-4 text-base font-semibold text-slate-950 transition-all hover:bg-cyan-100"
              >
                {{ isAuthenticated ? t('home.goToDashboard') : t('home.juliu.finalPrimaryCta') }}
              </router-link>
              <a
                :href="secondaryActionHref"
                :target="secondaryActionTarget"
                :rel="secondaryActionTarget ? 'noopener noreferrer' : undefined"
                class="inline-flex items-center justify-center rounded-full border border-white/20 px-8 py-4 text-base font-semibold text-white transition-all hover:bg-white/10"
              >
                {{ secondaryActionLabel }}
              </a>
            </div>
          </div>
        </div>
      </section>
    </main>

    <footer class="relative z-10 border-t border-white/8 px-6 py-8">
      <div class="mx-auto flex max-w-6xl flex-col gap-4 text-sm text-slate-400 sm:flex-row sm:items-center sm:justify-between">
        <p>&copy; {{ currentYear }} {{ brandName }}. {{ t('home.footer.allRightsReserved') }}</p>
        <div class="flex items-center gap-5">
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="transition-colors hover:text-white"
          >
            {{ t('home.docs') }}
          </a>
          <a
            :href="githubUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="transition-colors hover:text-white"
          >
            GitHub
          </a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import ThemeToggleButton from '@/components/common/ThemeToggleButton.vue'
import JuliuFlowLogo from '@/components/common/JuliuFlowLogo.vue'
import Icon from '@/components/icons/Icon.vue'
import { DEFAULT_SITE_NAME, DEFAULT_SITE_SUBTITLE } from '@/constants/branding'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || DEFAULT_SITE_NAME)
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || DEFAULT_SITE_SUBTITLE)
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const githubUrl = 'https://github.com/Wei-Shaw/sub2api'

const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const currentYear = computed(() => new Date().getFullYear())

const brandName = computed(() => (siteName.value === DEFAULT_SITE_NAME ? '聚流 Juliu' : siteName.value))
const headerBrandName = computed(() => {
  const trimmed = brandName.value.trim()
  const withoutRomanization = trimmed.replace(/\s+Juliu$/i, '').trim()
  return withoutRomanization || trimmed
})

const heroSubtitleFallbacks = [
  DEFAULT_SITE_SUBTITLE,
  '一站式大模型 API Token 汇聚平台。连接全球顶尖模型，聚沙成塔，汇流成海。',
  'A one-stop AI token hub connecting leading models worldwide. Gather access, unify routing, and scale with confidence.'
]

const heroSubtitle = computed(() => {
  const subtitle = siteSubtitle.value.trim()
  if (!subtitle || heroSubtitleFallbacks.includes(subtitle)) {
    return t('home.juliu.fallbackSubtitle')
  }
  return subtitle
})

const secondaryActionHref = computed(() => docUrl.value || '#features')
const secondaryActionTarget = computed(() => (docUrl.value ? '_blank' : undefined))
const secondaryActionLabel = computed(() => (docUrl.value ? t('home.docs') : t('home.juliu.secondaryCta')))

const valueTags = computed(() => [
  t('home.juliu.tags.unifiedAccess'),
  t('home.juliu.tags.stableRouting'),
  t('home.juliu.tags.usageBilling')
])

const features = computed<
  Array<{ icon: 'bolt' | 'shield' | 'cpu' | 'chart'; title: string; desc: string }>
>(() => [
  {
    icon: 'bolt',
    title: t('home.juliu.features.speed.title'),
    desc: t('home.juliu.features.speed.desc')
  },
  {
    icon: 'shield',
    title: t('home.juliu.features.stability.title'),
    desc: t('home.juliu.features.stability.desc')
  },
  {
    icon: 'cpu',
    title: t('home.juliu.features.aggregation.title'),
    desc: t('home.juliu.features.aggregation.desc')
  },
  {
    icon: 'chart',
    title: t('home.juliu.features.analytics.title'),
    desc: t('home.juliu.features.analytics.desc')
  }
])

const flowSteps = computed(() => [
  {
    index: '01',
    title: t('home.juliu.flowSteps.connect.title'),
    desc: t('home.juliu.flowSteps.connect.desc')
  },
  {
    index: '02',
    title: t('home.juliu.flowSteps.orchestrate.title'),
    desc: t('home.juliu.flowSteps.orchestrate.desc')
  },
  {
    index: '03',
    title: t('home.juliu.flowSteps.serve.title'),
    desc: t('home.juliu.flowSteps.serve.desc')
  }
])

const providers = computed(() => [
  'GPT-5.4',
  'Claude 4.7',
  'Gemini 2.5',
  t('home.juliu.providers.openaiCompatible'),
  'Anthropic',
  t('home.juliu.providers.multiAccountPool')
])

onMounted(() => {
  authStore.checkAuth()

  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
.hero-grid {
  background-image:
    linear-gradient(rgba(148, 163, 184, 0.12) 1px, transparent 1px),
    linear-gradient(90deg, rgba(148, 163, 184, 0.12) 1px, transparent 1px);
  background-size: 72px 72px;
  mask-image: radial-gradient(circle at center, black 35%, transparent 85%);
}

.hero-noise {
  background:
    radial-gradient(circle at top, rgba(34, 211, 238, 0.22), transparent 38%),
    radial-gradient(circle at 20% 80%, rgba(16, 185, 129, 0.15), transparent 28%),
    linear-gradient(180deg, rgba(15, 23, 42, 0.25), rgba(2, 6, 23, 0.95));
}

.hero-orb {
  filter: blur(90px);
  opacity: 0.9;
}

.hero-orb-cyan {
  background: radial-gradient(circle, rgba(34, 211, 238, 0.24), transparent 68%);
  animation: driftCentered 18s ease-in-out infinite;
}

.hero-orb-emerald {
  background: radial-gradient(circle, rgba(52, 211, 153, 0.18), transparent 72%);
  animation: driftFloat 22s ease-in-out infinite reverse;
}

.hero-beam {
  border: 1px solid rgba(103, 232, 249, 0.1);
  background: radial-gradient(circle, rgba(34, 211, 238, 0.12), transparent 62%);
  box-shadow: 0 0 140px rgba(34, 211, 238, 0.12);
  animation: pulseGlow 5s ease-in-out infinite;
}

.signal-shell {
  position: relative;
}

.signal-shell::before,
.signal-shell::after {
  content: '';
  position: absolute;
  inset: -22px;
  border-radius: 9999px;
  border: 1px solid rgba(103, 232, 249, 0.16);
}

.signal-shell::before {
  animation: spinSlow 16s linear infinite;
}

.signal-shell::after {
  inset: -38px;
  border-color: rgba(255, 255, 255, 0.06);
  animation: spinSlow 28s linear infinite reverse;
}

@keyframes driftCentered {
  0%,
  100% {
    transform: translate(-50%, 0) scale(1);
  }
  50% {
    transform: translate(-50%, -28px) scale(1.06);
  }
}

@keyframes driftFloat {
  0%,
  100% {
    transform: translate3d(0, 0, 0) scale(1);
  }
  50% {
    transform: translate3d(0, -28px, 0) scale(1.06);
  }
}

@keyframes pulseGlow {
  0%,
  100% {
    opacity: 0.55;
    transform: translate(-50%, -50%) scale(0.92);
  }
  50% {
    opacity: 1;
    transform: translate(-50%, -50%) scale(1.04);
  }
}

@keyframes spinSlow {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
