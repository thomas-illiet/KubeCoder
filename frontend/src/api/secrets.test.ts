import { describe, expect, it, vi } from 'vitest'
import { createSecret, fetchOrganizationSecrets, replaceSecret } from './secrets'

describe('secrets API', () => {
  it('lists the read-only organization projection with encoded slug', async () => {
    const fetcher=vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify({items:[],total:0,limit:20,offset:0}),{status:200}))
    await fetchOrganizationSecrets('http://api','token','north/star',{status:'ACTIVE'},fetcher)
    expect(fetcher).toHaveBeenCalledWith('http://api/api/v1/organizations/north%2Fstar/secrets?limit=20&offset=0&status=ACTIVE',expect.objectContaining({headers:expect.objectContaining({Authorization:'Bearer token'})}))
  })
  it('sends values only in write requests', async () => {
    const metadata={id:'1',scope:'PLATFORM',variable_name:'TOKEN'}
    const fetcher=vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify(metadata),{status:201}))
    await createSecret('http://api','token',{scope:'PLATFORM',variable_name:'TOKEN',description:'test',value:'protected'},{},fetcher)
    const init=fetcher.mock.calls[0]?.[1]
    expect(JSON.parse(String(init?.body))).toMatchObject({value:'protected'})
    expect(JSON.stringify(metadata)).not.toContain('protected')
  })
  it('uses the destructive replacement endpoint', async () => {
    const fetcher=vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify({id:'1'}),{status:200}))
    await replaceSecret('http://api','token','secret-id',{value:'new-value'},{organizationID:'org-id'},fetcher)
    expect(fetcher.mock.calls[0]?.[0]).toBe('http://api/api/v1/admin/organizations/org-id/secrets/secret-id/replace')
  })
  it('sends server sorting and pagination parameters', async () => {
    const fetcher=vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify({items:[],total:0,limit:20,offset:20}),{status:200}))
    await fetchOrganizationSecrets('http://api','token','northstar',{sort_by:'expires_at',sort_order:'desc',limit:20,offset:20},fetcher)
    expect(fetcher.mock.calls[0]?.[0]).toBe('http://api/api/v1/organizations/northstar/secrets?limit=20&offset=20&sort_by=expires_at&sort_order=desc')
  })
})
