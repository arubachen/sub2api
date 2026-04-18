import { describe, expect, it } from 'vitest'
import { AVATAR_MAX_UPLOAD_BYTES, validateAvatarFile } from '../avatar'

describe('avatar utils', () => {
  it('rejects unsupported image types', () => {
    const file = new File(['avatar'], 'avatar.gif', { type: 'image/gif' })
    expect(validateAvatarFile(file)).toBe('invalid-type')
  })

  it('rejects files above the upload limit', () => {
    const file = new File([new Uint8Array(AVATAR_MAX_UPLOAD_BYTES + 1)], 'avatar.png', { type: 'image/png' })
    expect(validateAvatarFile(file)).toBe('too-large')
  })

  it('accepts supported image types within the size budget', () => {
    const file = new File(['avatar'], 'avatar.png', { type: 'image/png' })
    expect(validateAvatarFile(file)).toBeNull()
  })
})
