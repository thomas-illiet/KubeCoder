import { describe, expect, it, vi } from 'vitest'
import { fetchOrganizationSSHKey, regenerateOrganizationSSHKey } from './organizationSSHKeys'

const key = {
  public_key: 'ssh-ed25519 AAAA test',
  fingerprint: 'SHA256:test',
  algorithm: 'ssh-ed25519' as const,
  created_at: '2026-09-28T12:00:00Z',
  updated_at: '2026-09-28T12:00:00Z',
}

describe('organization SSH key API', () => {
  it('loads only public key metadata from the organization endpoint', async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify(key)))
    await expect(fetchOrganizationSSHKey('http://api.test', 'token', 'northstar labs', fetcher)).resolves.toEqual(key)
    expect(fetcher).toHaveBeenCalledWith('http://api.test/api/v1/organizations/northstar%20labs/ssh-key', {
      headers: { Authorization: 'Bearer token' },
    })
  })

  it('regenerates the organization key with a bodyless POST', async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify(key)))
    await regenerateOrganizationSSHKey('http://api.test', 'token', 'northstar-labs', fetcher)
    expect(fetcher).toHaveBeenCalledWith('http://api.test/api/v1/organizations/northstar-labs/ssh-key/regenerate', {
      method: 'POST',
      headers: { Authorization: 'Bearer token' },
    })
  })

  it('preserves the HTTP status needed for the empty state', async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(null, { status: 404 }))
    await expect(fetchOrganizationSSHKey('http://api.test', 'token', 'legacy', fetcher)).rejects.toMatchObject({ status: 404 })
  })
})
