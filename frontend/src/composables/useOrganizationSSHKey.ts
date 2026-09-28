import { readonly, shallowRef } from 'vue'
import { fetchOrganizationSSHKey, regenerateOrganizationSSHKey, SSHKeyRequestError, type OrganizationSSHKey } from '../api/organizationSSHKeys'
import { useRuntimeConfig } from '../runtime/config'
import { useAuth } from './useAuth'

export function useOrganizationSSHKey() {
  const auth = useAuth()
  const config = useRuntimeConfig()
  const key = shallowRef<OrganizationSSHKey | null>(null)
  const loading = shallowRef(false)
  const saving = shallowRef(false)
  let revision = 0

  function credentials(): [string, string] {
    if (!auth.accessToken.value) throw new Error('The authenticated session does not contain an access token.')
    return [config.apiBaseUrl, auth.accessToken.value]
  }

  async function load(slug: string): Promise<void> {
    const current = ++revision
    loading.value = true
    try {
      const [baseURL, token] = credentials()
      const result = await fetchOrganizationSSHKey(baseURL, token, slug)
      if (current === revision) key.value = result
    } catch (error) {
      if (error instanceof SSHKeyRequestError && error.status === 404) {
        if (current === revision) key.value = null
      }
      else throw error
    } finally {
      if (current === revision) loading.value = false
    }
  }

  async function regenerate(slug: string): Promise<OrganizationSSHKey> {
    saving.value = true
    try {
      const [baseURL, token] = credentials()
      const result = await regenerateOrganizationSSHKey(baseURL, token, slug)
      key.value = result
      return result
    } finally {
      saving.value = false
    }
  }

  function reset(): void {
    revision += 1
    key.value = null
    loading.value = false
    saving.value = false
  }

  return { key: readonly(key), loading: readonly(loading), saving: readonly(saving), load, regenerate, reset }
}
