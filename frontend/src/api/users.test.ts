import { describe, expect, it, vi } from 'vitest'
import { deleteAdminUser, fetchAdminUsers, fetchCurrentUser, updateAdminUserRole } from './users'

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

describe('user administration API', () => {
  it('requests a searchable paginated user page', async () => {
    const page = { items: [], total: 24, limit: 20, offset: 20 }
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify(page)))

    await expect(fetchAdminUsers('http://api.test', 'token', {
      query: 'alice martin',
      role: 'admin',
      limit: 20,
      offset: 20,
      orderBy: 'created_at',
      orderDirection: 'desc',
    }, fetcher)).resolves.toEqual(page)

    expect(fetcher).toHaveBeenCalledWith('http://api.test/api/v1/admin/users?limit=20&offset=20&query=alice+martin&role=admin&order_by=created_at&order_direction=desc', {
      headers: { Authorization: 'Bearer token' },
    })
  })

  it('surfaces the API problem detail', async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify({ detail: 'Administrator access is required.' }), { status: 403 }))

    await expect(fetchAdminUsers('http://api.test', 'token', {}, fetcher)).rejects.toThrow('Administrator access is required.')
  })

  it('updates roles and deletes users through the user endpoint', async () => {
    const fetcher = vi.fn<typeof fetch>()
      .mockResolvedValueOnce(new Response(JSON.stringify(user)))
      .mockResolvedValueOnce(new Response(null, { status: 204 }))

    await expect(updateAdminUserRole('http://api.test', 'token', user.id, false, fetcher)).resolves.toEqual(user)
    await expect(deleteAdminUser('http://api.test', 'token', user.id, fetcher)).resolves.toBeUndefined()

    expect(fetcher).toHaveBeenNthCalledWith(1, `http://api.test/api/v1/admin/users/${user.id}`, {
      method: 'PATCH',
      body: JSON.stringify({ is_admin: false }),
      headers: { Authorization: 'Bearer token', 'Content-Type': 'application/json' },
    })
    expect(fetcher).toHaveBeenNthCalledWith(2, `http://api.test/api/v1/admin/users/${user.id}`, {
      method: 'DELETE',
      headers: { Authorization: 'Bearer token' },
    })
  })
})
