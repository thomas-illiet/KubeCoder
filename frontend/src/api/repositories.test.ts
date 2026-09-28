import { describe, expect, it, vi } from 'vitest'
import { createRepository, deleteRepository, fetchRepositories } from './repositories'

describe('repository API', () => {
  it('serializes server-side filters and pagination', async () => {
    const page = { items: [], total: 0, limit: 20, offset: 20 }
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify(page)))
    await expect(fetchRepositories('http://api.test', 'token', 'northstar', { query: 'api', provider: 'github', configured: true, agentID: 'agent-id', limit: 20, offset: 20 }, fetcher)).resolves.toEqual(page)
    expect(fetcher).toHaveBeenCalledWith('http://api.test/api/v1/organizations/northstar/repositories?limit=20&offset=20&query=api&provider=github&configured=true&agent_id=agent-id', { headers: { Authorization: 'Bearer token' } })
  })

  it('permanently deletes through the tenant-scoped endpoint', async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(null, { status: 204 }))
    await expect(deleteRepository('http://api.test', 'token', 'northstar', 'repository-id', fetcher)).resolves.toBeUndefined()
    expect(fetcher).toHaveBeenCalledWith('http://api.test/api/v1/organizations/northstar/repositories/repository-id', { method: 'DELETE', headers: { Authorization: 'Bearer token' } })
  })

  it('sends the selected secret policy', async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify({ id: 'repository-id' }), { status: 201 }))
    await createRepository('http://api.test', 'token', 'northstar', { name: 'app', provider: 'github', clone_url: 'https://github.com/acme/app', default_branch: 'main', include_submodules: false, agent_id: null, secret_mode: 'SELECTED', secret_ids: ['platform-secret', 'organization-secret'] }, fetcher)
    expect(JSON.parse(String(fetcher.mock.calls[0]?.[1]?.body))).toMatchObject({ secret_mode: 'SELECTED', secret_ids: ['platform-secret', 'organization-secret'] })
  })
})
