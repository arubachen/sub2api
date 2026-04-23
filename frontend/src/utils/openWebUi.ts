const trimTrailingSlash = (value: string) => value.replace(/\/+$/, '')

export const normalizePortalBaseRoot = (value: string) =>
  trimTrailingSlash(value.trim()).replace(/\/v1\/?$/, '')

export const ensureV1BaseUrl = (value: string) => {
  const trimmed = trimTrailingSlash(value)
  return trimmed.endsWith('/v1') ? trimmed : `${trimmed}/v1`
}

export const buildOpenWebUiUrl = (apiBaseUrl: string, apiKey: string) => {
  const baseRoot = normalizePortalBaseRoot(apiBaseUrl)
  const apiRoot = ensureV1BaseUrl(baseRoot)
  const hash = new URLSearchParams({
    juliu_openai_url: apiRoot,
    juliu_openai_key: apiKey,
  })

  return `${baseRoot}/open-webui/#${hash.toString()}`
}
