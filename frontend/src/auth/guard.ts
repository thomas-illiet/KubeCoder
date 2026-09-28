import type { Router } from 'vue-router'

export interface AuthGuardApi {
  isAuthenticated: { readonly value: boolean }
  isAdmin: { readonly value: boolean }
}

export interface OrganizationGuardApi {
  state: { readonly current: { readonly slug: string } | null }
  hasMembership(slug: string): boolean
}

export function installAuthGuard(router: Router, authApi: AuthGuardApi, organizationApi?: OrganizationGuardApi): void {
  router.beforeEach((to) => {
    if (to.meta.public) return true

    if (!authApi.isAuthenticated.value) {
      return {
        path: '/login',
        query: { redirect: to.fullPath },
      }
    }

    if (to.meta.requiresAdmin && !authApi.isAdmin.value) return { path: '/forbidden' }
    if (to.path === '/organization' && organizationApi?.state.current) {
      return { path: `/organizations/${organizationApi.state.current.slug}` }
    }
    if (to.meta.requiresOrganization) {
      const slug = typeof to.params.organizationSlug === 'string' ? to.params.organizationSlug : ''
      if (!organizationApi?.hasMembership(slug)) return { path: '/forbidden' }
    }
    return true
  })
}
