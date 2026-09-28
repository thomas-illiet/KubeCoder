export type LayoutMode = 'organization' | 'admin'
export type NavigationItem = { title: string; icon: string; to: string }

export function organizationSettingsNavigation(organizationBase: string): NavigationItem[] {
  return [
    { title: 'Git SSH key', icon: 'mdi-key-chain-variant', to: `${organizationBase}/ssh-key` },
    { title: 'Settings', icon: 'mdi-tune-variant', to: `${organizationBase}/settings` },
  ]
}

export function canShowWorkspaceSwitch(mode: LayoutMode, isAdmin: boolean): boolean {
  return mode === 'admin' || isAdmin
}

export function documentationTarget(mode: LayoutMode, organizationSlug = ''): string {
  return mode === 'admin' ? '/admin/documentation' : `/organizations/${organizationSlug}/documentation`
}

export function isSearchShortcut(event: Pick<KeyboardEvent, 'key' | 'metaKey' | 'ctrlKey'>): boolean {
  return (event.metaKey || event.ctrlKey) && event.key.toLocaleLowerCase('en') === 'k'
}
