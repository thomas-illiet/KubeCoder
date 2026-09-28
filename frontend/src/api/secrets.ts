import type { Page } from './organizations'

export type SecretScope = 'PLATFORM' | 'ORGANIZATION'
export type SecretStatus = 'ACTIVE' | 'EXPIRING_SOON' | 'EXPIRED'
export type SecretTargetType = 'AGENT' | 'REPOSITORY'
export type SecretSortBy = 'variable_name' | 'scope' | 'binding_count' | 'value_replaced_at' | 'expires_at' | 'status'
export type SortOrder = 'asc' | 'desc'

export interface SecretBinding { id: string; target_type: SecretTargetType; target_id: string; target_name: string }
export interface SecretTarget { id: string; type: SecretTargetType; name: string }
export interface Secret {
  id: string
  scope: SecretScope
  organization_id?: string
  variable_name: string
  description: string
  expires_at: string | null
  value_replaced_at: string
  status: SecretStatus
  bindings: SecretBinding[]
  binding_count: number
  created_at: string
  updated_at: string
}
export interface SecretInput { scope: SecretScope; variable_name: string; description: string; value: string; expires_at?: string | null }
export interface SecretFilters { query?: string; scope?: SecretScope; status?: SecretStatus; sort_by?: SecretSortBy; sort_order?: SortOrder; limit?: number; offset?: number }
export interface SecretOwner { organizationID?: string; organizationSlug?: string }

async function request<T>(url: string, token: string, fetcher: typeof fetch, init: RequestInit = {}): Promise<T> {
  const response = await fetcher(url, { ...init, headers: { Authorization: `Bearer ${token}`, ...(init.body ? { 'Content-Type': 'application/json' } : {}), ...init.headers } })
  if (!response.ok) {
    const problem = await response.json().catch(() => null) as { detail?: string } | null
    throw new Error(problem?.detail ?? `Secret request failed (HTTP ${response.status}).`)
  }
  if (response.status === 204) return undefined as T
  return response.json() as Promise<T>
}

function parameters(filters: SecretFilters): string {
  const value = new URLSearchParams({ limit: String(filters.limit ?? 20), offset: String(filters.offset ?? 0) })
  if (filters.query) value.set('query', filters.query)
  if (filters.scope) value.set('scope', filters.scope)
  if (filters.status) value.set('status', filters.status)
  if (filters.sort_by) value.set('sort_by', filters.sort_by)
  if (filters.sort_order) value.set('sort_order', filters.sort_order)
  return value.toString()
}
function managedCollection(baseURL: string, owner: SecretOwner = {}): string {
  if (owner.organizationSlug) return `${baseURL}/api/v1/organizations/${encodeURIComponent(owner.organizationSlug)}/secrets`
  if (owner.organizationID) return `${baseURL}/api/v1/admin/organizations/${owner.organizationID}/secrets`
  return `${baseURL}/api/v1/admin/secrets`
}

export function fetchAdminSecrets(baseURL: string, token: string, filters: SecretFilters = {}, owner: SecretOwner = {}, fetcher: typeof fetch = fetch): Promise<Page<Secret>> { return request(`${managedCollection(baseURL, owner)}?${parameters(filters)}`, token, fetcher) }
export function fetchOrganizationSecrets(baseURL: string, token: string, slug: string, filters: SecretFilters = {}, fetcher: typeof fetch = fetch): Promise<Page<Secret>> { return request(`${baseURL}/api/v1/organizations/${encodeURIComponent(slug)}/secrets?${parameters(filters)}`, token, fetcher) }
export function createSecret(baseURL: string, token: string, input: SecretInput, owner: SecretOwner = {}, fetcher: typeof fetch = fetch): Promise<Secret> { return request(managedCollection(baseURL, owner), token, fetcher, { method: 'POST', body: JSON.stringify(input) }) }
export function replaceSecret(baseURL: string, token: string, id: string, input: { value: string; expires_at?: string | null }, owner: SecretOwner = {}, fetcher: typeof fetch = fetch): Promise<Secret> { return request(`${managedCollection(baseURL, owner)}/${id}/replace`, token, fetcher, { method: 'POST', body: JSON.stringify(input) }) }
export function fetchSecretTargets(baseURL: string, token: string, owner: SecretOwner = {}, fetcher: typeof fetch = fetch): Promise<SecretTarget[]> { return request(`${managedCollection(baseURL, owner)}/targets`, token, fetcher) }
export function addSecretBinding(baseURL: string, token: string, id: string, input: { target_type: SecretTargetType; target_id: string }, owner: SecretOwner = {}, fetcher: typeof fetch = fetch): Promise<SecretBinding> { return request(`${managedCollection(baseURL, owner)}/${id}/bindings`, token, fetcher, { method: 'POST', body: JSON.stringify(input) }) }
export function removeSecretBinding(baseURL: string, token: string, id: string, bindingID: string, owner: SecretOwner = {}, fetcher: typeof fetch = fetch): Promise<void> { return request(`${managedCollection(baseURL, owner)}/${id}/bindings/${bindingID}`, token, fetcher, { method: 'DELETE' }) }
