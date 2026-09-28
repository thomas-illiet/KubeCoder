import { describe, expect, it, vi } from 'vitest'
import type { User } from 'oidc-client-ts'
import type { RuntimeConfig } from '../runtime/config'
import { createAuth } from './oidc'

const config: RuntimeConfig = {
  apiBaseUrl: 'http://api.test',
  oidc: {
    authority: 'http://identity.test', clientId: 'web',
    redirectUri: 'http://web.test/auth/callback', postLogoutRedirectUri: 'http://web.test/logout/callback',
  },
}

const oidcUser = { access_token: 'access-token', expired: false, state: {} } as User
const currentUser = {
  id: '2c47b281-568b-45a6-9006-2e9ba7237fbc', subject: 'subject', username: 'admin',
  display_name: 'Demo Admin', email: 'admin@kubecoder.local', is_admin: true,
  preferred_organization: null,
  created_at: '2026-09-28T12:00:00Z', updated_at: '2026-09-28T12:00:00Z',
}

function fakeManager(user: User | null = oidcUser) {
  return {
    events: {
      addUserLoaded: vi.fn(), addUserUnloaded: vi.fn(),
      addAccessTokenExpired: vi.fn(), addSilentRenewError: vi.fn(),
    },
    getUser: vi.fn().mockResolvedValue(user),
    removeUser: vi.fn().mockResolvedValue(undefined),
    signinRedirect: vi.fn().mockResolvedValue(undefined),
    signinRedirectCallback: vi.fn().mockResolvedValue(oidcUser),
    signoutRedirect: vi.fn().mockResolvedValue(undefined),
    signoutRedirectCallback: vi.fn().mockResolvedValue(undefined),
  }
}

describe('application authentication', () => {
  it('restores an OIDC session only after loading the backend profile', async () => {
    const manager = fakeManager()
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify(currentUser), { status: 200 }))
    const auth = createAuth(config, { manager, fetcher })
    await auth.initialize()
    expect(auth.isAuthenticated.value).toBe(true)
    expect(auth.isAdmin.value).toBe(true)
    expect(auth.state.currentUser).toEqual(currentUser)
  })

  it.each([
    ['HTTP failure', vi.fn<typeof fetch>().mockResolvedValue(new Response('', { status: 500 }))],
    ['network failure', vi.fn<typeof fetch>().mockRejectedValue(new Error('network unavailable'))],
  ])('clears the local session after a backend %s', async (_label, fetcher) => {
    const manager = fakeManager()
    const auth = createAuth(config, { manager, fetcher })
    await auth.initialize()
    expect(auth.isAuthenticated.value).toBe(false)
    expect(auth.state.currentUser).toBeNull()
    expect(auth.state.error).toBeTruthy()
    expect(manager.removeUser).toHaveBeenCalled()
  })

  it('loads the backend profile after the login callback', async () => {
    const manager = fakeManager(null)
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify(currentUser), { status: 200 }))
    const auth = createAuth(config, { manager, fetcher })
    await expect(auth.completeLogin()).resolves.toBe('/organization')
    expect(auth.isAuthenticated.value).toBe(true)
  })
})
