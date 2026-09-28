import { describe, expect, it } from 'vitest'
import { createMemoryHistory, createRouter } from 'vue-router'
import { installAuthGuard } from './guard'

function testRouter(authenticated: boolean, admin = false) {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/login', component: {}, meta: { public: true } },
      { path: '/auth/callback', component: {}, meta: { public: true } },
      { path: '/organization', component: {} },
      { path: '/admin', component: {}, meta: { requiresAdmin: true } },
      { path: '/forbidden', component: {} },
    ],
  })
  installAuthGuard(router, { isAuthenticated: { value: authenticated }, isAdmin: { value: admin } })
  return router
}

describe('authentication guard', () => {
  it('allows public routes without a session', async () => {
    const router = testRouter(false)
    await router.push('/auth/callback')
    expect(router.currentRoute.value.fullPath).toBe('/auth/callback')
  })

  it('redirects anonymous users and preserves the requested destination', async () => {
    const router = testRouter(false)
    await router.push('/organization?tab=sessions')
    expect(router.currentRoute.value.fullPath).toBe('/login?redirect=/organization?tab=sessions')
  })

  it('allows protected routes with a session', async () => {
    const router = testRouter(true)
    await router.push('/organization')
    expect(router.currentRoute.value.path).toBe('/organization')
  })

  it('allows administrators to open administration routes', async () => {
    const router = testRouter(true, true)
    await router.push('/admin')
    expect(router.currentRoute.value.path).toBe('/admin')
  })

  it('redirects non-administrators to the forbidden page', async () => {
    const router = testRouter(true)
    await router.push('/admin')
    expect(router.currentRoute.value.path).toBe('/forbidden')
  })
})
