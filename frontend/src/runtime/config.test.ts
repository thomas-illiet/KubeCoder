import { describe, expect, it, vi } from 'vitest'
import { loadRuntimeConfig, parseRuntimeConfig } from './config'

const validConfig = {
  apiBaseUrl: 'http://localhost:30081/',
  oidc: {
    authority: 'http://localhost:30080/realms/kubecoder',
    clientId: 'kubecoder-web',
    redirectUri: 'http://localhost:5173/auth/callback',
    postLogoutRedirectUri: 'http://localhost:5173/logout/callback',
  },
}

describe('runtime configuration', () => {
  it('loads and validates config.json without caching it', async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify(validConfig), { status: 200 }))
    const config = await loadRuntimeConfig(fetcher)
    expect(fetcher).toHaveBeenCalledWith('/config.json', { cache: 'no-store' })
    expect(config.apiBaseUrl).toBe('http://localhost:30081')
  })

  it('rejects missing required fields', () => {
    expect(() => parseRuntimeConfig({ apiBaseUrl: 'http://localhost:30081', oidc: {} })).toThrow('oidc.authority')
    expect(() => parseRuntimeConfig({ ...validConfig, unexpected: true })).toThrow('Unknown runtime configuration field')
    expect(() => parseRuntimeConfig({ ...validConfig, apiBaseUrl: '/api' })).toThrow('absolute URL')
  })

  it('reports missing and invalid configuration documents', async () => {
    await expect(loadRuntimeConfig(vi.fn<typeof fetch>().mockResolvedValue(new Response('', { status: 404 })))).rejects.toThrow('HTTP 404')
    await expect(loadRuntimeConfig(vi.fn<typeof fetch>().mockResolvedValue(new Response('{', { status: 200 })))).rejects.toThrow('valid JSON')
  })
})
