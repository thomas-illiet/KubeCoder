import type { Page } from './organizations'

export interface PublicAgent {
  id: string
  name: string
  description: string
  capabilities: readonly string[]
  active: boolean
}

export interface Agent extends PublicAgent {
  runtime_adapter: string
  runtime_version: string
  image: string
  provider: string
  model: string
  system_prompt: string
  cpu_millis: number
  memory_mb: number
  storage_mb: number
  max_duration_seconds: number
  created_at: string
  updated_at: string
}

export type AgentInput = Omit<Agent, 'id' | 'created_at' | 'updated_at' | 'capabilities'> & { capabilities: string[] }

async function request<T>(url: string, token: string, fetcher: typeof fetch, init: RequestInit = {}): Promise<T> {
  const response = await fetcher(url, { ...init, headers: { Authorization: `Bearer ${token}`, ...(init.body ? { 'Content-Type': 'application/json' } : {}), ...init.headers } })
  if (!response.ok) {
    const problem = await response.json().catch(() => null) as { detail?: string } | null
    throw new Error(problem?.detail ?? `Agent request failed (HTTP ${response.status}).`)
  }
  if (response.status === 204) return undefined as T
  return response.json() as Promise<T>
}

export function fetchAgents(baseURL: string, token: string, options: { query?: string; active?: boolean; limit?: number; offset?: number } = {}, fetcher: typeof fetch = fetch): Promise<Page<Agent>> {
  const parameters = new URLSearchParams({ limit: String(options.limit ?? 20), offset: String(options.offset ?? 0) })
  if (options.query) parameters.set('query', options.query)
  if (options.active !== undefined) parameters.set('active', String(options.active))
  return request(`${baseURL}/api/v1/admin/agents?${parameters}`, token, fetcher)
}

export function fetchPublicAgents(baseURL: string, token: string, slug: string, fetcher: typeof fetch = fetch): Promise<Page<PublicAgent>> {
  return request(`${baseURL}/api/v1/organizations/${encodeURIComponent(slug)}/agents?limit=100`, token, fetcher)
}

export function createAgent(baseURL: string, token: string, input: AgentInput, fetcher: typeof fetch = fetch): Promise<Agent> {
  return request(`${baseURL}/api/v1/admin/agents`, token, fetcher, { method: 'POST', body: JSON.stringify(input) })
}

export function updateAgent(baseURL: string, token: string, id: string, input: AgentInput, fetcher: typeof fetch = fetch): Promise<Agent> {
  return request(`${baseURL}/api/v1/admin/agents/${id}`, token, fetcher, { method: 'PATCH', body: JSON.stringify(input) })
}

export function deleteAgent(baseURL: string, token: string, id: string, fetcher: typeof fetch = fetch): Promise<void> {
  return request(`${baseURL}/api/v1/admin/agents/${id}`, token, fetcher, { method: 'DELETE' })
}
