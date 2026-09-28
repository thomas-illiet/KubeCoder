export interface CurrentUser {
  id: string
  subject: string
  username: string
  display_name: string
  email: string
  is_admin: boolean
  preferred_organization: {
    id: string
    name: string
    slug: string
  } | null
  created_at: string
  updated_at: string
}

function isCurrentUser(value: unknown): value is CurrentUser {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return false
  const user = value as Record<string, unknown>
  return typeof user.id === 'string'
    && typeof user.subject === 'string'
    && typeof user.username === 'string'
    && typeof user.display_name === 'string'
    && typeof user.email === 'string'
    && typeof user.is_admin === 'boolean'
    && (user.preferred_organization === null || isOrganizationSummary(user.preferred_organization))
    && typeof user.created_at === 'string'
    && typeof user.updated_at === 'string'
}

function isOrganizationSummary(value: unknown): boolean {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return false
  const organization = value as Record<string, unknown>
  return typeof organization.id === 'string'
    && typeof organization.name === 'string'
    && typeof organization.slug === 'string'
}

export async function fetchCurrentUser(apiBaseUrl: string, accessToken: string, fetcher: typeof fetch = fetch): Promise<CurrentUser> {
  const response = await fetcher(`${apiBaseUrl}/api/v1/users/me`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  })
  if (!response.ok) {
    throw new Error(`The user profile could not be loaded (HTTP ${response.status}).`)
  }
  const payload: unknown = await response.json()
  if (!isCurrentUser(payload)) {
    throw new Error('The user profile returned by the API is invalid.')
  }
  return payload
}
