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

export interface UserPage {
  items: CurrentUser[]
  total: number
  limit: number
  offset: number
}

export interface AdminUserPageRequest {
  query?: string
  role?: 'all' | 'admin' | 'user'
  limit?: number
  offset?: number
  orderBy?: 'display_name' | 'username' | 'email' | 'is_admin' | 'created_at' | 'updated_at'
  orderDirection?: 'asc' | 'desc'
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

// fetchAdminUsers returns one bounded page of users provisioned by OIDC sign-in.
export async function fetchAdminUsers(apiBaseUrl: string, accessToken: string, options: AdminUserPageRequest = {}, fetcher: typeof fetch = fetch): Promise<UserPage> {
  const parameters = new URLSearchParams({
    limit: String(options.limit ?? 20),
    offset: String(options.offset ?? 0),
  })
  if (options.query) parameters.set('query', options.query)
  if (options.role && options.role !== 'all') parameters.set('role', options.role)
  if (options.orderBy) parameters.set('order_by', options.orderBy)
  if (options.orderDirection) parameters.set('order_direction', options.orderDirection)

  const response = await fetcher(`${apiBaseUrl}/api/v1/admin/users?${parameters}`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  })
  if (!response.ok) {
    const problem = await response.json().catch(() => null) as { detail?: string } | null
    throw new Error(problem?.detail ?? `Users could not be loaded (HTTP ${response.status}).`)
  }
  return response.json() as Promise<UserPage>
}

async function mutateAdminUser(apiBaseUrl: string, accessToken: string, userID: string, init: RequestInit, fetcher: typeof fetch): Promise<CurrentUser | void> {
  const response = await fetcher(`${apiBaseUrl}/api/v1/admin/users/${encodeURIComponent(userID)}`, {
    ...init,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      ...(init.body ? { 'Content-Type': 'application/json' } : {}),
    },
  })
  if (!response.ok) {
    const problem = await response.json().catch(() => null) as { detail?: string } | null
    throw new Error(problem?.detail ?? `The user could not be updated (HTTP ${response.status}).`)
  }
  if (response.status === 204) return
  return response.json() as Promise<CurrentUser>
}

export function updateAdminUserRole(apiBaseUrl: string, accessToken: string, userID: string, isAdmin: boolean, fetcher: typeof fetch = fetch): Promise<CurrentUser> {
  return mutateAdminUser(apiBaseUrl, accessToken, userID, { method: 'PATCH', body: JSON.stringify({ is_admin: isAdmin }) }, fetcher) as Promise<CurrentUser>
}

export function deleteAdminUser(apiBaseUrl: string, accessToken: string, userID: string, fetcher: typeof fetch = fetch): Promise<void> {
  return mutateAdminUser(apiBaseUrl, accessToken, userID, { method: 'DELETE' }, fetcher) as Promise<void>
}
