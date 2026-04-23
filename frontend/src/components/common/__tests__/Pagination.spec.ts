import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../Pagination.vue')
const componentSource = readFileSync(componentPath, 'utf8')

describe('Pagination shell alignment', () => {
  it('keeps the footer row at the shared 64px shell rhythm', () => {
    expect(componentSource).toContain('min-h-16')
    expect(componentSource).toContain('items-center justify-between')
  })
})
