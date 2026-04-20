<template>
  <BaseDialog :show="show" :title="t('admin.risk.settings.title')" width="wide" @close="emit('close')">
    <div class="space-y-5">
      <div v-if="loadError" class="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-600 dark:border-red-900/50 dark:bg-red-950/20 dark:text-red-300">
        {{ loadError }}
      </div>

      <div v-if="validationErrors.length" class="rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-700 dark:border-amber-900/50 dark:bg-amber-950/20 dark:text-amber-200">
        <div class="font-semibold">{{ t('admin.risk.settings.validationTitle') }}</div>
        <ul class="mt-2 list-disc pl-5">
          <li v-for="msg in validationErrors" :key="msg">{{ msg }}</li>
        </ul>
      </div>

      <section class="rounded-2xl border border-gray-200 p-4 dark:border-dark-700">
        <div class="flex items-start justify-between gap-4">
          <div>
            <h4 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.risk.settings.ipIntelTitle') }}</h4>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.risk.settings.ipIntelHint') }}</p>
          </div>
          <label class="inline-flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
            <input v-model="draft.ip_intel_enabled" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500" />
            {{ t('common.enabled') }}
          </label>
        </div>

        <div class="mt-4 grid gap-4 md:grid-cols-2">
          <div>
            <label class="input-label">{{ t('admin.risk.settings.providerLabel') }}</label>
            <Select v-model="draft.ip_intel_provider" :options="providerOptions" />
          </div>
          <div class="flex items-end">
            <a :href="settings.ip_intel_docs_url || 'https://ipinfo.io/developers/lite-api'" target="_blank" rel="noreferrer" class="text-sm font-medium text-primary-600 hover:text-primary-500 dark:text-primary-400">
              {{ t('admin.risk.settings.providerDocs') }}
            </a>
          </div>
        </div>

        <div class="mt-4 grid gap-4 md:grid-cols-[1fr_auto] md:items-end">
          <div>
            <label class="input-label">{{ t('admin.risk.settings.tokenLabel') }}</label>
            <input v-model.trim="tokenInput" type="password" class="input" :placeholder="settings.ip_intel_token_configured ? t('admin.risk.settings.tokenConfiguredPlaceholder') : t('admin.risk.settings.tokenPlaceholder')" />
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
              {{ settings.ip_intel_token_configured ? t('admin.risk.settings.tokenConfiguredHint') : t('admin.risk.settings.tokenEmptyHint') }}
            </p>
          </div>
          <button v-if="settings.ip_intel_token_configured" type="button" class="btn btn-secondary h-10 px-3 text-sm" @click="clearExistingToken = !clearExistingToken">
            {{ clearExistingToken ? t('common.cancel') : t('admin.risk.settings.clearToken') }}
          </button>
        </div>
      </section>

      <section class="rounded-2xl border border-gray-200 p-4 dark:border-dark-700">
        <h4 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.risk.settings.thresholdTitle') }}</h4>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.risk.settings.thresholdHint') }}</p>
        <div class="mt-4 grid gap-4 md:grid-cols-3">
          <div>
            <label class="input-label">{{ t('admin.risk.settings.reviewThreshold') }}</label>
            <input v-model.number="draft.review_threshold" type="number" min="1" step="1" class="input" />
          </div>
          <div>
            <label class="input-label">{{ t('admin.risk.settings.throttleThreshold') }}</label>
            <input v-model.number="draft.throttle_threshold" type="number" min="1" step="1" class="input" />
          </div>
          <div>
            <label class="input-label">{{ t('admin.risk.settings.freezeThreshold') }}</label>
            <input v-model.number="draft.freeze_threshold" type="number" min="1" step="1" class="input" />
          </div>
        </div>
      </section>

      <section class="rounded-2xl border border-gray-200 p-4 dark:border-dark-700">
        <div class="flex items-start justify-between gap-4">
          <div>
            <h4 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.risk.settings.automationTitle') }}</h4>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.risk.settings.automationHint') }}</p>
          </div>
          <label class="inline-flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
            <input v-model="draft.auto_enabled" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500" />
            {{ t('common.enabled') }}
          </label>
        </div>

        <div class="mt-4 grid gap-4 md:grid-cols-2">
          <label class="rounded-2xl border border-gray-100 p-4 dark:border-dark-800">
            <div class="flex items-start gap-3">
              <input v-model="draft.auto_throttle" :disabled="!draft.auto_enabled" type="checkbox" class="mt-1 h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500 disabled:opacity-50" />
              <div>
                <p class="text-sm font-medium text-gray-900 dark:text-white">{{ t('admin.risk.settings.autoThrottleTitle') }}</p>
                <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.risk.settings.autoThrottleHint') }}</p>
              </div>
            </div>
            <div class="mt-4">
              <label class="input-label">{{ t('admin.risk.settings.autoThrottleCap') }}</label>
              <input v-model.number="draft.auto_throttle_concurrency_cap" :disabled="!draft.auto_enabled || !draft.auto_throttle" type="number" min="1" step="1" class="input disabled:opacity-60" />
            </div>
          </label>

          <label class="rounded-2xl border border-gray-100 p-4 dark:border-dark-800">
            <div class="flex items-start gap-3">
              <input v-model="draft.auto_freeze" :disabled="!draft.auto_enabled" type="checkbox" class="mt-1 h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500 disabled:opacity-50" />
              <div>
                <p class="text-sm font-medium text-gray-900 dark:text-white">{{ t('admin.risk.settings.autoFreezeTitle') }}</p>
                <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.risk.settings.autoFreezeHint') }}</p>
              </div>
            </div>
          </label>
        </div>
      </section>

      <div class="flex items-center justify-end gap-3 pt-2">
        <button type="button" class="btn btn-secondary px-4 py-2" @click="emit('close')">{{ t('common.cancel') }}</button>
        <button type="button" class="btn btn-primary px-4 py-2" :disabled="saving || validationErrors.length > 0" @click="save">
          {{ saving ? t('common.saving') : t('common.save') }}
        </button>
      </div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import { adminAPI } from '@/api/admin'
import type { UserRiskSettings, UpdateUserRiskSettingsRequest } from '@/api/admin/risk'
import { useAppStore } from '@/stores'

const props = defineProps<{ show: boolean }>()
const emit = defineEmits<{ (e: 'close'): void; (e: 'saved', value: UserRiskSettings): void }>()
const { t } = useI18n()
const appStore = useAppStore()

const settings = reactive<UserRiskSettings>({
  ip_intel_enabled: false,
  ip_intel_provider: 'ipinfo_lite',
  ip_intel_token_configured: false,
  ip_intel_docs_url: 'https://ipinfo.io/developers/lite-api',
  review_threshold: 50,
  throttle_threshold: 80,
  freeze_threshold: 120,
  auto_enabled: false,
  auto_throttle: false,
  auto_freeze: false,
  auto_throttle_concurrency_cap: 1
})

const draft = reactive({
  ip_intel_enabled: false,
  ip_intel_provider: 'ipinfo_lite',
  review_threshold: 50,
  throttle_threshold: 80,
  freeze_threshold: 120,
  auto_enabled: false,
  auto_throttle: false,
  auto_freeze: false,
  auto_throttle_concurrency_cap: 1
})

const tokenInput = ref('')
const clearExistingToken = ref(false)
const saving = ref(false)
const loadError = ref('')

const providerOptions = computed(() => [
  { value: 'ipinfo_lite', label: 'IPinfo Lite (free)' }
])

const validationErrors = computed(() => {
  const errors: string[] = []
  if (draft.review_threshold < 1 || draft.throttle_threshold < 1 || draft.freeze_threshold < 1) {
    errors.push(t('admin.risk.settings.validationPositive'))
  }
  if (!(draft.review_threshold < draft.throttle_threshold && draft.throttle_threshold < draft.freeze_threshold)) {
    errors.push(t('admin.risk.settings.validationAscending'))
  }
  if (draft.auto_throttle_concurrency_cap < 1) {
    errors.push(t('admin.risk.settings.validationThrottleCap'))
  }
  const willHaveToken = (!!tokenInput.value.trim()) || (settings.ip_intel_token_configured && !clearExistingToken.value)
  if (draft.ip_intel_enabled && !willHaveToken) {
    errors.push(t('admin.risk.settings.validationTokenRequired'))
  }
  return errors
})

watch(() => props.show, async (show) => {
  if (!show) return
  await load()
})

function resetDraft(value: UserRiskSettings) {
  settings.ip_intel_enabled = value.ip_intel_enabled
  settings.ip_intel_provider = value.ip_intel_provider
  settings.ip_intel_token_configured = value.ip_intel_token_configured
  settings.ip_intel_docs_url = value.ip_intel_docs_url
  settings.review_threshold = value.review_threshold
  settings.throttle_threshold = value.throttle_threshold
  settings.freeze_threshold = value.freeze_threshold
  settings.auto_enabled = value.auto_enabled
  settings.auto_throttle = value.auto_throttle
  settings.auto_freeze = value.auto_freeze
  settings.auto_throttle_concurrency_cap = value.auto_throttle_concurrency_cap

  draft.ip_intel_enabled = value.ip_intel_enabled
  draft.ip_intel_provider = value.ip_intel_provider
  draft.review_threshold = value.review_threshold
  draft.throttle_threshold = value.throttle_threshold
  draft.freeze_threshold = value.freeze_threshold
  draft.auto_enabled = value.auto_enabled
  draft.auto_throttle = value.auto_throttle
  draft.auto_freeze = value.auto_freeze
  draft.auto_throttle_concurrency_cap = value.auto_throttle_concurrency_cap
  tokenInput.value = ''
  clearExistingToken.value = false
}

async function load() {
  loadError.value = ''
  try {
    const data = await adminAPI.risk.getSettings()
    resetDraft(data)
  } catch (err: any) {
    console.error('[RiskSettingsDialog] Failed to load settings', err)
    loadError.value = err?.response?.data?.detail || err?.message || t('admin.risk.settings.loadFailed')
  }
}

async function save() {
  if (validationErrors.value.length) return
  saving.value = true
  try {
    const payload: UpdateUserRiskSettingsRequest = {
      ip_intel_enabled: draft.ip_intel_enabled,
      ip_intel_provider: draft.ip_intel_provider,
      review_threshold: draft.review_threshold,
      throttle_threshold: draft.throttle_threshold,
      freeze_threshold: draft.freeze_threshold,
      auto_enabled: draft.auto_enabled,
      auto_throttle: draft.auto_throttle,
      auto_freeze: draft.auto_freeze,
      auto_throttle_concurrency_cap: draft.auto_throttle_concurrency_cap
    }
    if (tokenInput.value.trim()) {
      payload.ip_intel_token = tokenInput.value.trim()
    }
    if (clearExistingToken.value) {
      payload.clear_ip_intel_token = true
    }
    const updated = await adminAPI.risk.updateSettings(payload)
    resetDraft(updated)
    appStore.showSuccess(t('admin.risk.settings.saveSuccess'))
    emit('saved', updated)
  } catch (err: any) {
    console.error('[RiskSettingsDialog] Failed to save settings', err)
    appStore.showError(err?.response?.data?.detail || err?.message || t('admin.risk.settings.saveFailed'))
  } finally {
    saving.value = false
  }
}
</script>
