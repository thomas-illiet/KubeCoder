export interface OrganizationSSHKey {
  public_key: string
  fingerprint: string
  algorithm: 'ssh-ed25519'
  created_at: string
  updated_at: string
}

export class SSHKeyRequestError extends Error {
  constructor(public readonly status: number) {
    super(`Organization SSH key request failed (HTTP ${status}).`)
  }
}

async function request(apiBaseUrl: string, token: string, slug: string, fetcher: typeof fetch, init?: RequestInit): Promise<OrganizationSSHKey> {
  const response = await fetcher(`${apiBaseUrl}/api/v1/organizations/${encodeURIComponent(slug)}/ssh-key${init ? '/regenerate' : ''}`, {
    ...init,
    headers: { Authorization: `Bearer ${token}` },
  })
  if (!response.ok) throw new SSHKeyRequestError(response.status)
  return response.json() as Promise<OrganizationSSHKey>
}

export function fetchOrganizationSSHKey(apiBaseUrl: string, token: string, slug: string, fetcher: typeof fetch = fetch): Promise<OrganizationSSHKey> {
  return request(apiBaseUrl, token, slug, fetcher)
}

export function regenerateOrganizationSSHKey(apiBaseUrl: string, token: string, slug: string, fetcher: typeof fetch = fetch): Promise<OrganizationSSHKey> {
  return request(apiBaseUrl, token, slug, fetcher, { method: 'POST' })
}
