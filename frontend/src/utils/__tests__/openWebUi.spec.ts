import { describe, expect, it } from 'vitest'

import { buildOpenWebUiUrl } from '../openWebUi'

describe('buildOpenWebUiUrl', () => {
  it('builds a juliu bridge URL from a public /v1 endpoint', () => {
    expect(buildOpenWebUiUrl('https://hub.juliu.one/v1', 'sk-test')).toBe(
      'https://hub.juliu.one/open-webui/#juliu_openai_url=https%3A%2F%2Fhub.juliu.one%2Fv1&juliu_openai_key=sk-test',
    )
  })

  it('normalizes a base URL without duplicating /v1', () => {
    expect(buildOpenWebUiUrl('https://hub.juliu.one/', 'sk-test')).toBe(
      'https://hub.juliu.one/open-webui/#juliu_openai_url=https%3A%2F%2Fhub.juliu.one%2Fv1&juliu_openai_key=sk-test',
    )
  })
})
