import { computed, reactive } from 'vue'
import { describe, expect, it, vi } from 'vitest'
import type { AuthApi } from '../auth/oidc'
import type { CurrentUser } from '../api/users'
import type { RuntimeConfig } from '../runtime/config'
import { createOrganizationService } from './service'

const config: RuntimeConfig = {
  apiBaseUrl: 'http://api.test',
  oidc: {
    authority: 'http://identity.test', clientId: 'web',
    redirectUri: 'http://web.test/auth/callback', postLogoutRedirectUri: 'http://web.test/logout/callback',
  },
}

const organization = {
  id: 'fd9c4fb2-8b6c-47fa-81fe-36fdd9813af7', name: 'Northstar Labs', slug: 'northstar-labs',
  created_at: '2026-09-28T12:00:00Z', updated_at: '2026-09-28T12:00:00Z',
}

// fakeAuth creates the minimum authenticated state required by the organization service.
function fakeAuth(preferred: { id: string; name: string; slug: string } | null = null): AuthApi {
  const state = reactive({
    initialized: true, loading: false, error: null,
    user: { access_token: 'token', expired: false },
    currentUser: {
      id: 'a182ceac-939a-4a03-8a55-dd47185567cf', subject: 'subject', username: 'user',
      display_name: 'User', email: 'user@example.test', is_admin: false,
      preferred_organization: preferred, created_at: '2026-09-28T12:00:00Z', updated_at: '2026-09-28T12:00:00Z',
    },
  })
  return {
    state,
    isAuthenticated: computed(() => true),
    isAdmin: computed(() => false),
    accessToken: computed(() => 'token'),
    updatePreferredOrganization: (value: CurrentUser['preferred_organization']) => { state.currentUser.preferred_organization = value },
  } as unknown as AuthApi
}

describe('organization context', () => {
  it('keeps an empty organization list as a valid authenticated state', async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify({ items: [], total: 0, limit: 100, offset: 0 })))
    const service = createOrganizationService(config, fakeAuth(), fetcher)
    await service.initialize()
    expect(service.hasOrganizations.value).toBe(false)
    expect(service.state.current).toBeNull()
  })

  it('tolerates a legacy null organization list without leaving loading active', async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify({ items: null, total: 0, limit: 100, offset: 0 })))
    const service = createOrganizationService(config, fakeAuth(), fetcher)
    await service.initialize()
    expect(service.state.items).toEqual([])
    expect(service.state.loading).toBe(false)
  })

  it('uses a valid server-side preference without rewriting it', async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify({ items: [organization], total: 1, limit: 100, offset: 0 })))
    const service = createOrganizationService(config, fakeAuth({ id: organization.id, name: organization.name, slug: organization.slug }), fetcher)
    await service.initialize()
    expect(service.state.current?.slug).toBe('northstar-labs')
    expect(fetcher).toHaveBeenCalledTimes(1)
  })

  it('persists the first organization when no valid preference exists', async () => {
    const fetcher = vi.fn<typeof fetch>()
      .mockResolvedValueOnce(new Response(JSON.stringify({ items: [organization], total: 1, limit: 100, offset: 0 })))
      .mockResolvedValueOnce(new Response(JSON.stringify(organization)))
    const auth = fakeAuth()
    const service = createOrganizationService(config, auth, fetcher)
    await service.initialize()
    expect(service.state.current?.id).toBe(organization.id)
    expect(auth.state.currentUser?.preferred_organization?.id).toBe(organization.id)
    expect(fetcher.mock.calls[1]?.[0]).toBe('http://api.test/api/v1/organizations/northstar-labs/preferred')
  })
})
