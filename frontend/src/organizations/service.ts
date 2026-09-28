import { computed, reactive, readonly, type App, type ComputedRef, type DeepReadonly, type InjectionKey } from 'vue'
import {
  addOrganizationMember,
  createOrganization,
  deleteOrganization,
  fetchAdminOrganizations,
  fetchOrganizationMembers,
  fetchOrganizations,
  fetchProvisionedUsers,
  preferOrganization,
  removeOrganizationMember,
  renameOrganization,
  type AdminOrganization,
  type AdminOrganizationPage,
  type Organization,
  type OrganizationInput,
  type OrganizationMember,
  type Page,
  type PageRequest,
  type MemberPageRequest,
} from '../api/organizations'
import type { CurrentUser } from '../api/users'
import type { AuthApi } from '../auth/oidc'
import type { RuntimeConfig } from '../runtime/config'

interface OrganizationState {
  initialized: boolean
  loading: boolean
  items: Organization[]
  current: Organization | null
  error: string | null
}

export interface OrganizationApi {
  state: DeepReadonly<OrganizationState>
  hasOrganizations: ComputedRef<boolean>
  initialize(): Promise<void>
  refresh(): Promise<void>
  select(slug: string): Promise<Organization>
  hasMembership(slug: string): boolean
  reset(): void
  listAdmin(options?: PageRequest): Promise<AdminOrganizationPage>
  create(input: OrganizationInput): Promise<Organization>
  rename(id: string, name: string): Promise<Organization>
  remove(id: string): Promise<void>
  listMembers(id: string, options?: MemberPageRequest): Promise<Page<OrganizationMember>>
  listUsers(query: string): Promise<Page<CurrentUser>>
  addMember(organizationID: string, userID: string): Promise<void>
  removeMember(organizationID: string, userID: string): Promise<void>
}

export const organizationKey: InjectionKey<OrganizationApi> = Symbol('kubecoder-organizations')

// createOrganizationService creates the application organization context.
export function createOrganizationService(config: RuntimeConfig, auth: AuthApi, fetcher: typeof fetch = fetch): OrganizationApi {
  const state = reactive<OrganizationState>({ initialized: false, loading: false, items: [], current: null, error: null })

  function credentials(): [string, string] {
    const token = auth.accessToken.value
    if (!token) throw new Error('The authenticated session does not contain an access token.')
    return [config.apiBaseUrl, token]
  }

  async function refresh(): Promise<void> {
    const [baseURL, token] = credentials()
    const result = await fetchOrganizations(baseURL, token, '', fetcher)
    state.items = result.items ?? []
    const current = state.current ? state.items.find((item) => item.id === state.current?.id) : undefined
    if (current) {
      state.current = current
      return
    }

    const preferredID = auth.state.currentUser?.preferred_organization?.id
    const preferred = preferredID ? state.items.find((item) => item.id === preferredID) : undefined
    if (preferred) {
      state.current = preferred
      return
    }

    state.current = null
    if (state.items.length > 0) {
      await select([...state.items].sort((a, b) => a.name.localeCompare(b.name))[0].slug)
    }
  }

  async function select(slug: string): Promise<Organization> {
    const existing = state.items.find((item) => item.slug === slug)
    if (!existing) throw new Error('You do not have access to this organization.')
    const [baseURL, token] = credentials()
    const selected = await preferOrganization(baseURL, token, slug, fetcher)
    state.current = selected
    auth.updatePreferredOrganization({ id: selected.id, name: selected.name, slug: selected.slug })
    return selected
  }

  async function initialize(): Promise<void> {
    state.loading = true
    state.error = null
    try {
      if (!auth.isAuthenticated.value) {
        state.initialized = true
        return
      }
      await refresh()
    } catch (error) {
      state.error = error instanceof Error ? error.message : 'Organizations could not be loaded.'
      throw error
    } finally {
      state.loading = false
      state.initialized = true
    }
  }

  function reset(): void {
    state.initialized = false
    state.loading = false
    state.items = []
    state.current = null
    state.error = null
  }

  return {
    state: readonly(state),
    hasOrganizations: computed(() => state.items.length > 0),
    initialize,
    refresh,
    select,
    hasMembership: (slug: string) => state.items.some((item) => item.slug === slug),
    reset,
    listAdmin: (options = {}) => { const [url, token] = credentials(); return fetchAdminOrganizations(url, token, options, fetcher) },
    create: (input) => { const [url, token] = credentials(); return createOrganization(url, token, input, fetcher) },
    rename: (id, name) => { const [url, token] = credentials(); return renameOrganization(url, token, id, name, fetcher) },
    remove: (id) => { const [url, token] = credentials(); return deleteOrganization(url, token, id, fetcher) },
    listMembers: (id, options = {}) => { const [url, token] = credentials(); return fetchOrganizationMembers(url, token, id, options, fetcher) },
    listUsers: (query) => { const [url, token] = credentials(); return fetchProvisionedUsers(url, token, query, fetcher) },
    addMember: async (organizationID, userID) => {
      const [url, token] = credentials()
      await addOrganizationMember(url, token, organizationID, userID, fetcher)
      if (userID === auth.state.currentUser?.id) await refresh()
    },
    removeMember: async (organizationID, userID) => {
      const [url, token] = credentials()
      await removeOrganizationMember(url, token, organizationID, userID, fetcher)
      if (userID === auth.state.currentUser?.id) await refresh()
    },
  }
}

// installOrganizations provides organization state to the Vue component tree.
export function installOrganizations(app: App, organizations: OrganizationApi): void {
  app.provide(organizationKey, organizations)
}
