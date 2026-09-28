import { computed } from 'vue'
import { auth } from '../auth/oidc'

export function useAuth() {
  const displayName = computed(() => auth.profile.value?.name ?? auth.profile.value?.preferred_username ?? 'KubeCoder user')
  const username = computed(() => auth.profile.value?.preferred_username ?? '')
  const email = computed(() => auth.profile.value?.email ?? '')
  const initials = computed(() => {
    const parts = displayName.value.trim().split(/\s+/).filter(Boolean)
    return parts.slice(0, 2).map((part) => part[0]?.toUpperCase()).join('') || 'KC'
  })

  return {
    ...auth,
    displayName,
    username,
    email,
    initials,
  }
}

