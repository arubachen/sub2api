<template>
  <div :class="props.embedded ? 'space-y-4' : 'card'">
    <div
      v-if="!props.embedded"
      class="border-b border-gray-100 px-6 py-4 dark:border-dark-700"
    >
      <h2 class="text-lg font-medium text-gray-900 dark:text-white">
        {{ t('profile.editProfile') }}
      </h2>
    </div>
    <div :class="props.embedded ? '' : 'px-6 py-6'">
      <form @submit.prevent="handleUpdateProfile" class="space-y-4">
        <div v-if="props.embedded">
          <p class="text-sm font-semibold text-gray-900 dark:text-white">
            {{ t('profile.editProfile') }}
          </p>
        </div>
        <div>
          <label for="username" class="input-label">
            {{ t('profile.username') }}
            <span class="ml-1 text-xs font-normal text-gray-400 dark:text-gray-500">
              ({{ t('common.optional') }})
            </span>
          </label>
          <input
            id="username"
            v-model="username"
            type="text"
            class="input"
            :placeholder="t('profile.enterUsername')"
          />
        </div>

        <div class="flex justify-end gap-3 pt-4">
          <button
            type="button"
            :disabled="loading || !isDirty"
            class="btn btn-secondary"
            @click="resetForm"
          >
            {{ t('common.cancel') }}
          </button>
          <button type="submit" :disabled="loading || !isDirty" class="btn btn-primary">
            {{ loading ? t('profile.updating') : t('profile.updateProfile') }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { userAPI } from '@/api'

const props = withDefaults(defineProps<{
  initialUsername: string
  embedded?: boolean
}>(), {
  embedded: false,
})

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const username = ref(props.initialUsername)
const loading = ref(false)
const normalizedInitialUsername = computed(() => props.initialUsername.trim())
const normalizedUsername = computed(() => username.value.trim())
const isDirty = computed(() => normalizedUsername.value !== normalizedInitialUsername.value)

watch(() => props.initialUsername, (val) => {
  username.value = val
})

const resetForm = () => {
  username.value = props.initialUsername
}

const handleUpdateProfile = async () => {
  if (!isDirty.value) {
    return
  }

  loading.value = true
  try {
    const updatedUser = await userAPI.updateProfile({
      username: normalizedUsername.value
    })
    authStore.user = updatedUser
    username.value = updatedUser.username || ''
    appStore.showSuccess(t('profile.updateSuccess'))
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('profile.updateFailed'))
  } finally {
    loading.value = false
  }
}
</script>
