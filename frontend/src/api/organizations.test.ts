import { describe, expect, it, vi } from 'vitest'
import { fetchProvisionedUsers, preferOrganization } from './organizations'

// Test preferred organization requests are authenticated and bodyless.
describe('organization API', () => {
  it('persists a preference through the slug endpoint', async () => {
    const organization = { id: 'id', name: 'Name', slug: 'name', created_at: 'now', updated_at: 'now' }
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify(organization)))
    await expect(preferOrganization('http://api.test', 'token', 'name', fetcher)).resolves.toEqual(organization)
    expect(fetcher).toHaveBeenCalledWith('http://api.test/api/v1/organizations/name/preferred', {
      method: 'PUT',
      headers: { Authorization: 'Bearer token' },
    })
  })

  it('searches a bounded page of provisioned users', async () => {
    const page = { items: [], total: 0, limit: 20, offset: 0 }
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify(page)))
    await expect(fetchProvisionedUsers('http://api.test', 'token', 'alice martin', fetcher)).resolves.toEqual(page)
    expect(fetcher).toHaveBeenCalledWith('http://api.test/api/v1/admin/users?limit=20&query=alice+martin', {
      headers: { Authorization: 'Bearer token' },
    })
  })
})
