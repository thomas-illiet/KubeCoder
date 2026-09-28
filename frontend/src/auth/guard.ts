import type { Router } from 'vue-router'

export interface AuthGuardApi {
  isAuthenticated: { readonly value: boolean }
  isAdmin: { readonly value: boolean }
}

export function installAuthGuard(router: Router, authApi: AuthGuardApi): void {
  router.beforeEach((to) => {
    if (to.meta.public) return true

    if (!authApi.isAuthenticated.value) {
      return {
        path: '/login',
        query: { redirect: to.fullPath },
      }
    }

    if (to.meta.requiresAdmin && !authApi.isAdmin.value) return { path: '/forbidden' }
    return true
  })
}
