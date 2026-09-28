import { readonly, shallowRef } from 'vue'
import { fetchPublicAgents, type PublicAgent } from '../api/agents'
import { createRepository, deleteRepository, fetchRepositories, updateRepository, type Repository, type RepositoryInput, type RepositoryPageRequest } from '../api/repositories'
import { useAuth } from './useAuth'
import { useRuntimeConfig } from '../runtime/config'

export function useRepositoryCatalog() {
  const auth = useAuth()
  const config = useRuntimeConfig()
  const items = shallowRef<Repository[]>([])
  const agents = shallowRef<PublicAgent[]>([])
  const total = shallowRef(0)
  const loading = shallowRef(false)
  let revision = 0

  function credentials(): [string, string] {
    if (!auth.accessToken.value) throw new Error('The authenticated session does not contain an access token.')
    return [config.apiBaseUrl, auth.accessToken.value]
  }

  async function load(slug: string, options: RepositoryPageRequest = {}): Promise<void> {
    const current = ++revision
    loading.value = true
    try {
      const [baseURL, token] = credentials()
      const result = await fetchRepositories(baseURL, token, slug, options)
      if (current !== revision) return
      items.value = result.items ?? []
      total.value = result.total
    } finally {
      if (current === revision) loading.value = false
    }
  }

  async function loadAgents(slug: string): Promise<void> {
    const [baseURL, token] = credentials()
    const result = await fetchPublicAgents(baseURL, token, slug)
    agents.value = result.items ?? []
  }

  async function save(slug: string, input: RepositoryInput, id?: string): Promise<Repository> {
    const [baseURL, token] = credentials()
    return id ? updateRepository(baseURL, token, slug, id, input) : createRepository(baseURL, token, slug, input)
  }

  async function remove(slug: string, id: string): Promise<void> {
    const [baseURL, token] = credentials()
    await deleteRepository(baseURL, token, slug, id)
  }

  function reset(): void {
    revision += 1
    items.value = []
    agents.value = []
    total.value = 0
    loading.value = false
  }

  return { items: readonly(items), agents: readonly(agents), total: readonly(total), loading: readonly(loading), load, loadAgents, save, remove, reset }
}
