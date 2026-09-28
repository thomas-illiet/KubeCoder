import { computed, inject } from 'vue'
import { authKey } from '../auth/oidc'

export function useAuth() {
  const auth = inject(authKey)
  if (!auth) throw new Error('Authentication service is not installed.')
  const displayName = computed(() => auth.state.currentUser?.display_name || auth.state.currentUser?.username || 'KubeCoder user')
  const username = computed(() => auth.state.currentUser?.username ?? '')
  const email = computed(() => auth.state.currentUser?.email ?? '')
  const subject = computed(() => auth.state.currentUser?.subject ?? '')
  const initials = computed(() => {
    const parts = displayName.value.trim().split(/\s+/).filter(Boolean)
    return parts.slice(0, 2).map((part) => part[0]?.toUpperCase()).join('') || 'KC'
  })

  return {
    ...auth,
    displayName,
    username,
    email,
    subject,
    initials,
  }
}
