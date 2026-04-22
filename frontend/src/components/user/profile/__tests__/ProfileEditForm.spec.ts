import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'

import ProfileEditForm from '../ProfileEditForm.vue'

const { updateProfile, showError, showSuccess } = vi.hoisted(() => ({
  updateProfile: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

const authStoreState = {
  user: {
    email: 'user@example.com',
    username: 'Alice',
  } as Record<string, unknown>,
}

const messages: Record<string, string> = {
  'profile.username': 'Username',
  'profile.enterUsername': 'Enter username',
  'profile.editProfile': 'Edit profile',
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
  useAuthStore: () => authStoreState,
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
    showError.mockReset()
    showSuccess.mockReset()
    authStoreState.user = {
      email: 'user@example.com',
      username: 'Alice',
    }
  })

  it('resets unsaved edits when cancel is clicked', async () => {
    const wrapper = mount(ProfileEditForm, {
      props: {
        initialUsername: 'Alice',
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
      email: 'user@example.com',
      username: '',
    })

    const wrapper = mount(ProfileEditForm, {
      props: {
        initialUsername: 'Alice',
      },
    })

    await wrapper.get('#username').setValue('   ')
    await wrapper.get('form').trigger('submit.prevent')

    expect(showError).not.toHaveBeenCalled()
    expect(updateProfile).toHaveBeenCalledWith({
      username: '',
    })
    expect(authStoreState.user).toEqual({
      email: 'user@example.com',
      username: '',
    })
    expect(showSuccess).toHaveBeenCalledWith('Profile updated')
  })
})
