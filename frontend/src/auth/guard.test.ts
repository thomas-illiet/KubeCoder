import { describe, expect, it } from 'vitest'
import { createMemoryHistory, createRouter } from 'vue-router'
import { installAuthGuard } from './guard'

function testRouter(authenticated: boolean) {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/login', component: {}, meta: { public: true } },
      { path: '/auth/callback', component: {}, meta: { public: true } },
      { path: '/organization', component: {} },
    ],
  })
  installAuthGuard(router, { isAuthenticated: { value: authenticated } })
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
})

