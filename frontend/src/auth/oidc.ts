import { UserManager, WebStorageStateStore, type User, type UserManagerSettings } from 'oidc-client-ts'
import { computed, reactive, readonly, type App, type ComputedRef, type DeepReadonly, type InjectionKey } from 'vue'
import { fetchCurrentUser, type CurrentUser } from '../api/users'
import type { RuntimeConfig } from '../runtime/config'
import { authenticationError, normalizeReturnTo } from './navigation'

interface AuthState {
  initialized: boolean
  loading: boolean
  user: User | null
  currentUser: CurrentUser | null
  error: string | null
}

interface RedirectState {
  returnTo?: string
}

export interface AuthApi {
  state: DeepReadonly<AuthState>
  isAuthenticated: ComputedRef<boolean>
  isAdmin: ComputedRef<boolean>
  accessToken: ComputedRef<string>
  updatePreferredOrganization(organization: CurrentUser['preferred_organization']): void
  initialize(): Promise<void>
  login(returnTo?: string): Promise<void>
  completeLogin(): Promise<string>
  logout(): Promise<void>
  completeLogout(): Promise<void>
}

export const authKey: InjectionKey<AuthApi> = Symbol('kubecoder-auth')

interface UserManagerPort {
  events: Pick<UserManager['events'], 'addUserLoaded' | 'addUserUnloaded' | 'addAccessTokenExpired' | 'addSilentRenewError'>
  getUser(): Promise<User | null>
  removeUser(): Promise<void>
  signinRedirect(args: { state: RedirectState }): Promise<void>
  signinRedirectCallback(): Promise<User>
  signoutRedirect(): Promise<void>
  signoutRedirectCallback(): Promise<unknown>
}

interface AuthDependencies {
  manager?: UserManagerPort
  fetcher?: typeof fetch
}

export function createAuth(config: RuntimeConfig, dependencies: AuthDependencies = {}): AuthApi {
  const manager: UserManagerPort = dependencies.manager ?? new UserManager({
    authority: config.oidc.authority,
    client_id: config.oidc.clientId,
    redirect_uri: config.oidc.redirectUri,
    post_logout_redirect_uri: config.oidc.postLogoutRedirectUri,
    response_type: 'code', scope: 'openid profile email', automaticSilentRenew: true, monitorSession: true,
    userStore: new WebStorageStateStore({ store: window.sessionStorage }),
  } satisfies UserManagerSettings)
  const fetcher = dependencies.fetcher ?? fetch
  const state = reactive<AuthState>({ initialized: false, loading: false, user: null, currentUser: null, error: null })
  let synchronizedToken = ''
  let synchronization: Promise<void> | null = null

  async function clearLocalSession(message?: string): Promise<void> {
    synchronizedToken = ''
    state.user = null
    state.currentUser = null
    if (message) state.error = message
    await manager.removeUser()
  }

  async function synchronizeUser(user: User): Promise<void> {
    if (!user.access_token) {
      await clearLocalSession('The OpenID Connect session does not contain an access token.')
      throw new Error(state.error ?? 'Access token missing.')
    }
    if (state.currentUser && synchronizedToken === user.access_token) {
      state.user = user
      return
    }
    if (synchronization) return synchronization
    synchronization = (async () => {
      try {
        const currentUser = await fetchCurrentUser(config.apiBaseUrl, user.access_token, fetcher)
        state.user = user
        state.currentUser = currentUser
        synchronizedToken = user.access_token
        state.error = null
      } catch (error) {
        await clearLocalSession(authenticationError(error))
        throw error
      } finally {
        synchronization = null
      }
    })()
    return synchronization
  }

  manager.events.addUserLoaded((user) => {
    void synchronizeUser(user).catch(() => undefined)
  })
  manager.events.addUserUnloaded(() => {
    synchronizedToken = ''
    state.user = null
    state.currentUser = null
  })
  manager.events.addAccessTokenExpired(() => {
    void clearLocalSession('Your session has expired. Please sign in again.').catch(() => undefined)
  })
  manager.events.addSilentRenewError((error) => {
    void clearLocalSession(authenticationError(error)).catch(() => undefined)
  })

  return {
    state: readonly(state),
    isAuthenticated: computed(() => Boolean(state.user && state.currentUser && !state.user.expired)),
    isAdmin: computed(() => state.currentUser?.is_admin === true),
    accessToken: computed(() => state.user?.access_token ?? ''),

    updatePreferredOrganization(organization: CurrentUser['preferred_organization']): void {
      if (state.currentUser) state.currentUser.preferred_organization = organization
    },

    async initialize(): Promise<void> {
      state.loading = true
      state.error = null
      try {
        const user = await manager.getUser()
        if (user && !user.expired) await synchronizeUser(user)
        else if (user?.expired) await clearLocalSession()
      } catch (error) {
        if (!state.error) await clearLocalSession(authenticationError(error))
      } finally {
        state.loading = false
        state.initialized = true
      }
    },

    async login(returnTo = '/organization'): Promise<void> {
      state.loading = true
      state.error = null
      try {
        await manager.signinRedirect({ state: { returnTo } satisfies RedirectState })
      } catch (error) {
        state.error = authenticationError(error)
        state.loading = false
        throw error
      }
    },

    async completeLogin(): Promise<string> {
      state.loading = true
      state.error = null
      try {
        const user = await manager.signinRedirectCallback()
        await synchronizeUser(user)
        const redirectState = user.state as RedirectState | undefined
        return normalizeReturnTo(redirectState?.returnTo)
      } finally {
        state.loading = false
      }
    },

    async logout(): Promise<void> {
      state.loading = true
      state.error = null
      try {
        await manager.signoutRedirect()
      } catch (error) {
        state.error = authenticationError(error)
        state.loading = false
        throw error
      }
    },

    async completeLogout(): Promise<void> {
      state.loading = true
      try {
        await manager.signoutRedirectCallback()
        await clearLocalSession()
      } finally {
        state.loading = false
      }
    },
  }
}

export function installAuth(app: App, auth: AuthApi): void {
  app.provide(authKey, auth)
}
