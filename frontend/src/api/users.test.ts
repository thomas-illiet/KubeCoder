import { describe, expect, it, vi } from 'vitest'
import { fetchCurrentUser } from './users'

const user = {
  id: '2c47b281-568b-45a6-9006-2e9ba7237fbc', subject: 'subject', username: 'admin',
  display_name: 'Demo Admin', email: 'admin@kubecoder.local', is_admin: true,
  preferred_organization: null,
  created_at: '2026-09-28T12:00:00Z', updated_at: '2026-09-28T12:00:00Z',
}

describe('current user API', () => {
  it('sends the OIDC access token and decodes the user', async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify(user), { status: 200 }))
    await expect(fetchCurrentUser('http://localhost:30081', 'secret-token', fetcher)).resolves.toEqual(user)
    expect(fetcher).toHaveBeenCalledWith('http://localhost:30081/api/v1/users/me', {
      headers: { Authorization: 'Bearer secret-token' },
    })
  })

  it('rejects HTTP errors and invalid responses', async () => {
    await expect(fetchCurrentUser('', 'token', vi.fn<typeof fetch>().mockResolvedValue(new Response('', { status: 500 })))).rejects.toThrow('HTTP 500')
    await expect(fetchCurrentUser('', 'token', vi.fn<typeof fetch>().mockResolvedValue(new Response('{}', { status: 200 })))).rejects.toThrow('invalid')
  })
})
