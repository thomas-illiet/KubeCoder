import { inject } from 'vue'
import { organizationKey } from '../organizations/service'

// useOrganizations returns the installed organization context.
export function useOrganizations() {
  const organizations = inject(organizationKey)
  if (!organizations) throw new Error('Organization service is not installed.')
  return organizations
}
