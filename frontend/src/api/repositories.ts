import type { PublicAgent } from './agents'
import type { Page } from './organizations'

export type GitProvider = 'github' | 'gitlab' | 'bitbucket'

export interface Repository {
  id: string
  organization_id: string
  name: string
  provider: GitProvider
  clone_url: string
  default_branch: string
  include_submodules: boolean
  agent: PublicAgent | null
  created_at: string
  updated_at: string
}

export interface RepositoryInput {
  name: string
  provider: GitProvider
  clone_url: string
  default_branch: string
  include_submodules: boolean
  agent_id: string | null
}

export interface RepositoryPageRequest {
  query?: string
  provider?: GitProvider
  configured?: boolean
  agentID?: string
  limit?: number
  offset?: number
}

async function request<T>(url: string, token: string, fetcher: typeof fetch, init: RequestInit = {}): Promise<T> {
  const response = await fetcher(url, { ...init, headers: { Authorization: `Bearer ${token}`, ...(init.body ? { 'Content-Type': 'application/json' } : {}), ...init.headers } })
  if (!response.ok) {
    const problem = await response.json().catch(() => null) as { detail?: string } | null
    throw new Error(problem?.detail ?? `Repository request failed (HTTP ${response.status}).`)
  }
  if (response.status === 204) return undefined as T
  return response.json() as Promise<T>
}

function collectionURL(baseURL: string, slug: string): string {
  return `${baseURL}/api/v1/organizations/${encodeURIComponent(slug)}/repositories`
}

export function fetchRepositories(baseURL: string, token: string, slug: string, options: RepositoryPageRequest = {}, fetcher: typeof fetch = fetch): Promise<Page<Repository>> {
  const parameters = new URLSearchParams({ limit: String(options.limit ?? 20), offset: String(options.offset ?? 0) })
  if (options.query) parameters.set('query', options.query)
  if (options.provider) parameters.set('provider', options.provider)
  if (options.configured !== undefined) parameters.set('configured', String(options.configured))
  if (options.agentID) parameters.set('agent_id', options.agentID)
  return request(`${collectionURL(baseURL, slug)}?${parameters}`, token, fetcher)
}

export function createRepository(baseURL: string, token: string, slug: string, input: RepositoryInput, fetcher: typeof fetch = fetch): Promise<Repository> {
  return request(collectionURL(baseURL, slug), token, fetcher, { method: 'POST', body: JSON.stringify(input) })
}

export function updateRepository(baseURL: string, token: string, slug: string, id: string, input: RepositoryInput, fetcher: typeof fetch = fetch): Promise<Repository> {
  return request(`${collectionURL(baseURL, slug)}/${id}`, token, fetcher, { method: 'PATCH', body: JSON.stringify(input) })
}

export function deleteRepository(baseURL: string, token: string, slug: string, id: string, fetcher: typeof fetch = fetch): Promise<void> {
  return request(`${collectionURL(baseURL, slug)}/${id}`, token, fetcher, { method: 'DELETE' })
}
