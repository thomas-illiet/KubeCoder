export interface SelectPresentation {
  icon: string
  color: string
}

type SelectItemObject = { title?: unknown; value?: unknown; icon?: unknown; color?: unknown }

const DEFAULT_PRESENTATION: SelectPresentation = {
  icon: 'mdi-circle-small',
  color: 'medium-emphasis',
}

const presentations: Record<string, SelectPresentation> = {
  // Filter summaries
  'all statuses': { icon: 'mdi-filter-variant', color: 'info' },
  'all scopes': { icon: 'mdi-filter-variant', color: 'info' },
  'all sources': { icon: 'mdi-filter-variant', color: 'info' },
  'all roles': { icon: 'mdi-filter-variant', color: 'info' },
  'all repositories': { icon: 'mdi-filter-variant', color: 'info' },
  'all providers': { icon: 'mdi-filter-variant', color: 'info' },
  'all transports': { icon: 'mdi-filter-variant', color: 'info' },

  // Lifecycle and health states
  enabled: { icon: 'mdi-check-circle-outline', color: 'success' },
  disabled: { icon: 'mdi-cancel', color: 'medium-emphasis' },
  active: { icon: 'mdi-check-circle-outline', color: 'success' },
  connected: { icon: 'mdi-link-variant', color: 'success' },
  ready: { icon: 'mdi-check-decagram-outline', color: 'success' },
  completed: { icon: 'mdi-check-circle-outline', color: 'success' },
  published: { icon: 'mdi-publish', color: 'success' },
  pending: { icon: 'mdi-clock-outline', color: 'warning' },
  'in progress': { icon: 'mdi-progress-clock', color: 'info' },
  review: { icon: 'mdi-eye-outline', color: 'info' },
  draft: { icon: 'mdi-file-edit-outline', color: 'warning' },
  'expiring soon': { icon: 'mdi-clock-alert-outline', color: 'warning' },
  expired: { icon: 'mdi-calendar-remove-outline', color: 'error' },
  interrupted: { icon: 'mdi-stop-circle-outline', color: 'error' },
  suspended: { icon: 'mdi-pause-circle-outline', color: 'error' },
  'action required': { icon: 'mdi-alert-circle-outline', color: 'warning' },

  // Scope and permissions
  global: { icon: 'mdi-earth', color: 'info' },
  organization: { icon: 'mdi-domain', color: 'secondary' },
  repository: { icon: 'mdi-source-repository', color: 'primary' },
  'northstar labs': { icon: 'mdi-domain', color: 'secondary' },
  owner: { icon: 'mdi-shield-crown-outline', color: 'warning' },
  admin: { icon: 'mdi-shield-account-outline', color: 'info' },
  member: { icon: 'mdi-account-outline', color: 'secondary' },

  // Providers and transports
  github: { icon: 'mdi-github', color: 'high-emphasis' },
  gitlab: { icon: 'mdi-gitlab', color: 'warning' },
  bitbucket: { icon: 'mdi-bitbucket', color: 'info' },
  'github.com': { icon: 'mdi-github', color: 'high-emphasis' },
  'github enterprise server': { icon: 'mdi-server-network', color: 'secondary' },
  'gitlab.com': { icon: 'mdi-gitlab', color: 'warning' },
  'self-managed': { icon: 'mdi-server-network', color: 'secondary' },
  'bitbucket cloud': { icon: 'mdi-bitbucket', color: 'info' },
  'bitbucket data center': { icon: 'mdi-server-network', color: 'secondary' },
  openai: { icon: 'mdi-creation-outline', color: 'success' },
  'openai-compatible endpoint': { icon: 'mdi-api', color: 'info' },
  stdio: { icon: 'mdi-console-line', color: 'secondary' },
  sse: { icon: 'mdi-access-point-network', color: 'info' },

  // Product entities
  secrets: { icon: 'mdi-key-variant', color: 'warning' },
  none: { icon: 'mdi-minus-circle-outline', color: 'medium-emphasis' },
  atlas: { icon: 'mdi-compass-rose', color: 'primary' },
  pathfinder: { icon: 'mdi-compass-outline', color: 'secondary' },
  sentinel: { icon: 'mdi-shield-search-outline', color: 'info' },
}

function isSelectItemObject(item: unknown): item is SelectItemObject {
  return typeof item === 'object' && item !== null
}

function itemText(item: unknown): string {
  if (isSelectItemObject(item)) {
    return String(item.title ?? item.value ?? '')
  }
  return String(item)
}

/** Returns the semantic icon and color used by every select option. */
export function selectPresentation(item: unknown): SelectPresentation {
  if (isSelectItemObject(item) && typeof item.icon === 'string') {
    return {
      icon: item.icon,
      color: typeof item.color === 'string' ? item.color : DEFAULT_PRESENTATION.color,
    }
  }

  const text = itemText(item).trim().toLowerCase()
  const exact = presentations[text]
  if (exact) return exact

  if (text.includes('repository') || ['identity-service', 'platform-k8s', 'checkout-api'].includes(text)) {
    return { icon: 'mdi-source-repository', color: 'primary' }
  }
  if (text.includes('git provider')) return { icon: 'mdi-source-branch', color: 'primary' }
  if (text.includes('issue tracker')) return { icon: 'mdi-ticket-outline', color: 'warning' }
  if (text.includes('kubernetes')) return { icon: 'mdi-kubernetes', color: 'info' }
  if (text.includes('filesystem')) return { icon: 'mdi-folder-outline', color: 'secondary' }
  if (text.includes('token') || text.includes('credential')) return { icon: 'mdi-key-variant', color: 'warning' }
  if (text.includes('opencode')) return { icon: 'mdi-application-cog-outline', color: 'primary' }
  if (/^[a-z]+\/[a-z_]+$/i.test(text)) return { icon: 'mdi-earth-clock', color: 'info' }

  return DEFAULT_PRESENTATION
}
