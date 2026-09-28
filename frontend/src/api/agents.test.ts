import { describe, expect, it, vi } from 'vitest'
import { createAgent, fetchPublicAgents } from './agents'

describe('agent API', () => {
  it('loads the expurgated organization catalog', async () => {
    const page = { items: [], total: 0, limit: 100, offset: 0 }
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify(page)))
    await expect(fetchPublicAgents('http://api.test', 'token', 'northstar labs', fetcher)).resolves.toEqual(page)
    expect(fetcher).toHaveBeenCalledWith('http://api.test/api/v1/organizations/northstar%20labs/agents?limit=100', { headers: { Authorization: 'Bearer token' } })
  })

  it('sends a complete strict agent input', async () => {
    const input = { name: 'Atlas', description: 'Agent', runtime_adapter: 'opencode', runtime_version: '1', image: 'image', provider: 'openai', model: 'gpt', system_prompt: 'Help', capabilities: ['tests'], cpu_millis: 1000, memory_mb: 2048, storage_mb: 4096, max_duration_seconds: 60, active: true }
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify({ id: 'id', ...input, created_at: 'now', updated_at: 'now' })))
    await createAgent('http://api.test', 'token', input, fetcher)
    expect(fetcher).toHaveBeenCalledWith('http://api.test/api/v1/admin/agents', { method: 'POST', body: JSON.stringify(input), headers: { Authorization: 'Bearer token', 'Content-Type': 'application/json' } })
  })
})
