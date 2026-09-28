import type { CurrentUser } from './users'

export interface Organization {
  id: string
  name: string
  slug: string
  created_at: string
  updated_at: string
}

export interface AdminOrganization extends Organization {
  member_count: number
}

export interface OrganizationMember extends CurrentUser {
  joined_at: string
}

export interface Page<T> {
  items: T[]
  total: number
  limit: number
  offset: number
}

export interface OrganizationInput {
  name: string
  slug: string
}

async function request<T>(url: string, token: string, fetcher: typeof fetch, init: RequestInit = {}): Promise<T> {
  const response = await fetcher(url, {
    ...init,
    headers: {
      Authorization: `Bearer ${token}`,
      ...(init.body ? { 'Content-Type': 'application/json' } : {}),
      ...init.headers,
    },
  })
  if (!response.ok) throw new Error(`Organization request failed (HTTP ${response.status}).`)
  if (response.status === 204) return undefined as T
  return response.json() as Promise<T>
}

// fetchOrganizations returns organizations accessible to the current user.
export function fetchOrganizations(apiBaseUrl: string, token: string, query = '', fetcher: typeof fetch = fetch): Promise<Page<Organization>> {
  const parameters = new URLSearchParams({ limit: '100' })
  if (query) parameters.set('query', query)
  return request(`${apiBaseUrl}/api/v1/organizations?${parameters}`, token, fetcher)
}

// fetchOrganization returns one organization through its membership-protected route.
export function fetchOrganization(apiBaseUrl: string, token: string, slug: string, fetcher: typeof fetch = fetch): Promise<Organization> {
  return request(`${apiBaseUrl}/api/v1/organizations/${encodeURIComponent(slug)}`, token, fetcher)
}

// preferOrganization persists the current user organization preference.
export function preferOrganization(apiBaseUrl: string, token: string, slug: string, fetcher: typeof fetch = fetch): Promise<Organization> {
  return request(`${apiBaseUrl}/api/v1/organizations/${encodeURIComponent(slug)}/preferred`, token, fetcher, { method: 'PUT' })
}

// fetchAdminOrganizations returns all organizations to a platform administrator.
export function fetchAdminOrganizations(apiBaseUrl: string, token: string, query = '', fetcher: typeof fetch = fetch): Promise<Page<AdminOrganization>> {
  const parameters = new URLSearchParams({ limit: '100' })
  if (query) parameters.set('query', query)
  return request(`${apiBaseUrl}/api/v1/admin/organizations?${parameters}`, token, fetcher)
}

// createOrganization creates a platform organization.
export function createOrganization(apiBaseUrl: string, token: string, input: OrganizationInput, fetcher: typeof fetch = fetch): Promise<Organization> {
  return request(`${apiBaseUrl}/api/v1/admin/organizations`, token, fetcher, { method: 'POST', body: JSON.stringify(input) })
}

// renameOrganization changes an organization display name.
export function renameOrganization(apiBaseUrl: string, token: string, id: string, name: string, fetcher: typeof fetch = fetch): Promise<Organization> {
  return request(`${apiBaseUrl}/api/v1/admin/organizations/${id}`, token, fetcher, { method: 'PATCH', body: JSON.stringify({ name }) })
}

// deleteOrganization permanently deletes an organization.
export function deleteOrganization(apiBaseUrl: string, token: string, id: string, fetcher: typeof fetch = fetch): Promise<void> {
  return request(`${apiBaseUrl}/api/v1/admin/organizations/${id}`, token, fetcher, { method: 'DELETE' })
}

// fetchOrganizationMembers returns users assigned to an organization.
export function fetchOrganizationMembers(apiBaseUrl: string, token: string, id: string, fetcher: typeof fetch = fetch): Promise<Page<OrganizationMember>> {
  return request(`${apiBaseUrl}/api/v1/admin/organizations/${id}/members?limit=100`, token, fetcher)
}

// fetchProvisionedUsers returns users that can be assigned to organizations.
export function fetchProvisionedUsers(apiBaseUrl: string, token: string, query: string, fetcher: typeof fetch = fetch): Promise<Page<CurrentUser>> {
  const parameters = new URLSearchParams({ limit: '20', query })
  return request(`${apiBaseUrl}/api/v1/admin/users?${parameters}`, token, fetcher)
}

// addOrganizationMember assigns a provisioned user to an organization.
export function addOrganizationMember(apiBaseUrl: string, token: string, organizationID: string, userID: string, fetcher: typeof fetch = fetch): Promise<void> {
  return request(`${apiBaseUrl}/api/v1/admin/organizations/${organizationID}/members`, token, fetcher, { method: 'POST', body: JSON.stringify({ user_id: userID }) })
}

// removeOrganizationMember revokes a user membership.
export function removeOrganizationMember(apiBaseUrl: string, token: string, organizationID: string, userID: string, fetcher: typeof fetch = fetch): Promise<void> {
  return request(`${apiBaseUrl}/api/v1/admin/organizations/${organizationID}/members/${userID}`, token, fetcher, { method: 'DELETE' })
}
