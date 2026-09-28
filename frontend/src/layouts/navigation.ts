export type LayoutMode = 'organization' | 'admin'

export function canShowWorkspaceSwitch(mode: LayoutMode, isAdmin: boolean): boolean {
  return mode === 'admin' || isAdmin
}

export function documentationTarget(mode: LayoutMode, organizationSlug = ''): string {
  return mode === 'admin' ? '/admin/documentation' : `/organizations/${organizationSlug}/documentation`
}

export function isSearchShortcut(event: Pick<KeyboardEvent, 'key' | 'metaKey' | 'ctrlKey'>): boolean {
  return (event.metaKey || event.ctrlKey) && event.key.toLocaleLowerCase('en') === 'k'
}
