<template>
  <div class="card">
    <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <h2 class="text-lg font-medium text-gray-900 dark:text-white">
        {{ t('profile.editProfile') }}
      </h2>
    </div>
    <div class="px-6 py-6">
      <form @submit.prevent="handleUpdateProfile" class="space-y-5">
        <div class="space-y-3">
          <label class="input-label">
            {{ t('profile.avatar') }}
          </label>
          <div class="flex flex-col gap-4 sm:flex-row sm:items-center">
            <UserAvatar
              :avatar-url="avatarPreviewUrl"
              :username="username"
              :email="authStore.user?.email"
              size-class="h-20 w-20"
              text-class="text-2xl font-bold"
              rounded-class="rounded-2xl"
              shadow-class="shadow-lg shadow-primary-500/20"
            />
            <div class="flex flex-wrap gap-2">
              <label class="btn btn-secondary cursor-pointer">
                <input
                  ref="fileInputRef"
                  type="file"
                  accept="image/jpeg,image/png,image/webp"
                  class="hidden"
                  @change="handleAvatarSelected"
                />
                {{ avatarValue ? t('profile.changeAvatar') : t('profile.uploadAvatar') }}
              </label>
              <button
                v-if="avatarValue"
                type="button"
                class="btn btn-secondary text-red-600 dark:text-red-400"
                @click="removeAvatar"
              >
                {{ t('profile.removeAvatar') }}
              </button>
            </div>
          </div>
          <p class="input-hint">
            {{ t('profile.avatarHint') }}
          </p>
          <p v-if="avatarError" class="input-error-text">
            {{ avatarError }}
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
          <button type="button" :disabled="loading || !isDirty" class="btn btn-secondary" @click="resetForm">
            {{ t('common.cancel') }}
          </button>
          <button type="submit" :disabled="loading || !isDirty" class="btn btn-primary">
            {{ loading ? t('profile.updating') : t('profile.updateProfile') }}
          </button>
        </div>
      </form>
    </div>
  </div>

  <AvatarCropDialog
    :show="showAvatarCropDialog"
    :image-url="avatarCropSource"
    @close="closeCropDialog"
    @apply="handleAvatarApplied"
  />
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { userAPI } from '@/api'
import UserAvatar from '@/components/common/UserAvatar.vue'
import AvatarCropDialog from '@/components/user/profile/AvatarCropDialog.vue'
import { loadImageDataUrl, validateAvatarFile } from '@/utils/avatar'

const props = defineProps<{
  initialUsername: string
  initialAvatarUrl?: string
}>()

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const username = ref(props.initialUsername)
const avatarValue = ref(props.initialAvatarUrl || '')
const avatarCropSource = ref('')
const avatarError = ref('')
const loading = ref(false)
const showAvatarCropDialog = ref(false)
const fileInputRef = ref<HTMLInputElement | null>(null)

const avatarPreviewUrl = computed(() => avatarValue.value || authStore.user?.avatar_url || '')
const normalizedInitialUsername = computed(() => props.initialUsername.trim())
const normalizedUsername = computed(() => username.value.trim())
const initialAvatarValue = computed(() => props.initialAvatarUrl || '')
const isDirty = computed(
  () =>
    normalizedUsername.value !== normalizedInitialUsername.value ||
    avatarValue.value !== initialAvatarValue.value,
)

watch(
  () => props.initialUsername,
  (val) => {
    username.value = val
  },
)

watch(
  () => props.initialAvatarUrl,
  (val) => {
    avatarValue.value = val || ''
  },
)

const translateAvatarError = (code: string) => {
  switch (code) {
    case 'invalid-type':
      return t('profile.avatarInvalidType')
    case 'too-large':
      return t('profile.avatarTooLarge')
    default:
      return t('profile.avatarReadFailed')
  }
}

const handleAvatarSelected = async (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  avatarError.value = ''

  if (!file) return

  const validationError = validateAvatarFile(file)
  if (validationError) {
    avatarError.value = translateAvatarError(validationError)
    input.value = ''
    return
  }

  try {
    avatarCropSource.value = await loadImageDataUrl(file)
    showAvatarCropDialog.value = true
  } catch {
    avatarError.value = t('profile.avatarReadFailed')
  } finally {
    input.value = ''
  }
}

const handleAvatarApplied = (value: string) => {
  avatarValue.value = value
  avatarCropSource.value = ''
  showAvatarCropDialog.value = false
  avatarError.value = ''
}

const closeCropDialog = () => {
  avatarCropSource.value = ''
  showAvatarCropDialog.value = false
}

const removeAvatar = () => {
  avatarValue.value = ''
  avatarError.value = ''
  closeCropDialog()
  if (fileInputRef.value) {
    fileInputRef.value.value = ''
  }
}

const resetForm = () => {
  username.value = props.initialUsername
  avatarValue.value = props.initialAvatarUrl || ''
  avatarError.value = ''
  closeCropDialog()
  if (fileInputRef.value) {
    fileInputRef.value.value = ''
  }
}

const handleUpdateProfile = async () => {
  if (!isDirty.value) {
    return
  }

  loading.value = true
  try {
    const updatedUser = await userAPI.updateProfile({
      username: normalizedUsername.value,
      avatar_url: avatarValue.value,
    })
    authStore.updateCurrentUser(updatedUser)
    username.value = updatedUser.username || ''
    avatarValue.value = updatedUser.avatar_url || ''
    avatarError.value = ''
    appStore.showSuccess(t('profile.updateSuccess'))
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('profile.updateFailed'))
  } finally {
    loading.value = false
  }
}
</script>
