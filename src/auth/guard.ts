import type { Router } from 'vue-router'

export interface AuthGuardApi {
  isAuthenticated: { readonly value: boolean }
}

export function installAuthGuard(router: Router, authApi: AuthGuardApi): void {
  router.beforeEach((to) => {
    if (to.meta.public || authApi.isAuthenticated.value) return true

    return {
      path: '/login',
      query: { redirect: to.fullPath },
    }
  })
}

