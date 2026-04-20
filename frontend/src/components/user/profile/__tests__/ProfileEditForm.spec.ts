import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'

import ProfileEditForm from '../ProfileEditForm.vue'

const { updateProfile, updateCurrentUser, showError, showSuccess } = vi.hoisted(() => ({
  updateProfile: vi.fn(),
  updateCurrentUser: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

const messages: Record<string, string> = {
  'profile.avatar': 'Avatar',
  'profile.changeAvatar': 'Change avatar',
  'profile.uploadAvatar': 'Upload avatar',
  'profile.removeAvatar': 'Remove avatar',
  'profile.avatarHint': 'Upload a square avatar.',
  'profile.username': 'Username',
  'profile.enterUsername': 'Enter username (optional)',
  'profile.editProfile': 'Edit Profile',
  'profile.updating': 'Saving...',
  'profile.updateProfile': 'Save profile',
  'profile.updateSuccess': 'Profile updated',
  'profile.updateFailed': 'Update failed',
  'common.optional': 'optional',
  'common.cancel': 'Cancel',
}

vi.mock('@/api', () => ({
  userAPI: {
    updateProfile,
  },
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    user: {
      email: 'user@example.com',
      username: 'Alice',
      avatar_url: '',
    },
    updateCurrentUser,
  }),
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
  }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => messages[key] ?? key,
    }),
  }
})

describe('ProfileEditForm', () => {
  beforeEach(() => {
    updateProfile.mockReset()
    updateCurrentUser.mockReset()
    showError.mockReset()
    showSuccess.mockReset()
  })

  it('resets unsaved edits when cancel is clicked', async () => {
    const wrapper = mount(ProfileEditForm, {
      props: {
        initialUsername: 'Alice',
        initialAvatarUrl: '',
      },
      global: {
        stubs: {
          UserAvatar: true,
          AvatarCropDialog: true,
        },
      },
    })

    const input = wrapper.get('#username')
    await input.setValue('Bob')
    expect((input.element as HTMLInputElement).value).toBe('Bob')

    await wrapper.get('button[type="button"]').trigger('click')
    await nextTick()

    expect((wrapper.get('#username').element as HTMLInputElement).value).toBe('Alice')
    expect(updateProfile).not.toHaveBeenCalled()
  })

  it('allows submitting an empty username to clear it', async () => {
    updateProfile.mockResolvedValue({
      username: '',
      email: 'user@example.com',
      avatar_url: '',
    })

    const wrapper = mount(ProfileEditForm, {
      props: {
        initialUsername: 'Alice',
        initialAvatarUrl: '',
      },
      global: {
        stubs: {
          UserAvatar: true,
          AvatarCropDialog: true,
        },
      },
    })

    await wrapper.get('#username').setValue('   ')
    await wrapper.get('form').trigger('submit.prevent')

    expect(showError).not.toHaveBeenCalled()
    expect(updateProfile).toHaveBeenCalledWith({
      username: '',
      avatar_url: '',
    })
    expect(updateCurrentUser).toHaveBeenCalled()
    expect(showSuccess).toHaveBeenCalledWith('Profile updated')
  })
})
