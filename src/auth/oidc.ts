import { UserManager, WebStorageStateStore, type User, type UserManagerSettings } from 'oidc-client-ts'
import { computed, reactive, readonly } from 'vue'
import { authenticationError, normalizeReturnTo } from './navigation'

export interface OidcProfile {
  sub: string
  name?: string
  preferred_username?: string
  email?: string
}

interface AuthState {
  initialized: boolean
  loading: boolean
  user: User | null
  error: string | null
}

interface RedirectState {
  returnTo?: string
}

const authority = import.meta.env.VITE_OIDC_AUTHORITY ?? 'http://localhost:30080/realms/kubecoder'
const clientId = import.meta.env.VITE_OIDC_CLIENT_ID ?? 'kubecoder-web'
const redirectUri = import.meta.env.VITE_OIDC_REDIRECT_URI ?? `${window.location.origin}/auth/callback`
const postLogoutRedirectUri = import.meta.env.VITE_OIDC_POST_LOGOUT_REDIRECT_URI ?? `${window.location.origin}/logout/callback`

export const oidcSettings: UserManagerSettings = {
  authority,
  client_id: clientId,
  redirect_uri: redirectUri,
  post_logout_redirect_uri: postLogoutRedirectUri,
  response_type: 'code',
  scope: 'openid profile email',
  automaticSilentRenew: true,
  monitorSession: true,
  userStore: new WebStorageStateStore({ store: window.sessionStorage }),
}

const manager = new UserManager(oidcSettings)
const state = reactive<AuthState>({
  initialized: false,
  loading: false,
  user: null,
  error: null,
})

manager.events.addUserLoaded((user) => {
  state.user = user
  state.error = null
})
manager.events.addUserUnloaded(() => {
  state.user = null
})
manager.events.addAccessTokenExpired(() => {
  state.user = null
})
manager.events.addSilentRenewError((error) => {
  state.error = error.message
})

export const auth = {
  state: readonly(state),
  isAuthenticated: computed(() => Boolean(state.user && !state.user.expired)),
  profile: computed(() => (state.user?.profile ?? null) as OidcProfile | null),

  async initialize(): Promise<void> {
    state.loading = true
    state.error = null
    try {
      const user = await manager.getUser()
      state.user = user && !user.expired ? user : null
      if (user?.expired) await manager.removeUser()
    } catch (error) {
      state.user = null
      state.error = authenticationError(error)
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
      state.user = user
      const redirectState = user.state as RedirectState | undefined
      return normalizeReturnTo(redirectState?.returnTo)
    } catch (error) {
      state.error = authenticationError(error)
      throw error
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
      await manager.removeUser()
      state.user = null
    } finally {
      state.loading = false
    }
  },
}
