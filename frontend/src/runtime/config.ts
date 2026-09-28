export interface RuntimeConfig {
  apiBaseUrl: string
  oidc: {
    authority: string
    clientId: string
    redirectUri: string
    postLogoutRedirectUri: string
  }
}

function assertKeys(value: Record<string, unknown>, expected: string[], path: string): void {
  const unexpected = Object.keys(value).filter((key) => !expected.includes(key))
  if (unexpected.length > 0) throw new Error(`Unknown runtime configuration field ${path}${unexpected[0]}.`)
}

function requiredString(value: unknown, path: string): string {
  if (typeof value !== 'string' || value.trim() === '') {
    throw new Error(`Runtime configuration field ${path} must be a non-empty string.`)
  }
  return value
}

function requiredUrl(value: unknown, path: string): string {
  const url = requiredString(value, path)
  try {
    return new URL(url).toString().replace(/\/$/, '')
  } catch {
    throw new Error(`Runtime configuration field ${path} must be an absolute URL.`)
  }
}

export function parseRuntimeConfig(value: unknown): RuntimeConfig {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    throw new Error('Runtime configuration must be a JSON object.')
  }
  const root = value as Record<string, unknown>
  assertKeys(root, ['apiBaseUrl', 'oidc'], '')
  if (!root.oidc || typeof root.oidc !== 'object' || Array.isArray(root.oidc)) {
    throw new Error('Runtime configuration field oidc must be an object.')
  }
  const oidc = root.oidc as Record<string, unknown>
  assertKeys(oidc, ['authority', 'clientId', 'redirectUri', 'postLogoutRedirectUri'], 'oidc.')
  return {
    apiBaseUrl: requiredUrl(root.apiBaseUrl, 'apiBaseUrl'),
    oidc: {
      authority: requiredUrl(oidc.authority, 'oidc.authority'),
      clientId: requiredString(oidc.clientId, 'oidc.clientId'),
      redirectUri: requiredUrl(oidc.redirectUri, 'oidc.redirectUri'),
      postLogoutRedirectUri: requiredUrl(oidc.postLogoutRedirectUri, 'oidc.postLogoutRedirectUri'),
    },
  }
}

export async function loadRuntimeConfig(fetcher: typeof fetch = fetch): Promise<RuntimeConfig> {
  let response: Response
  try {
    response = await fetcher('/config.json', { cache: 'no-store' })
  } catch (error) {
    throw new Error('Runtime configuration could not be loaded.', { cause: error })
  }
  if (!response.ok) {
    throw new Error(`Runtime configuration could not be loaded (HTTP ${response.status}).`)
  }
  let payload: unknown
  try {
    payload = await response.json()
  } catch (error) {
    throw new Error('Runtime configuration is not valid JSON.', { cause: error })
  }
  return parseRuntimeConfig(payload)
}
