export type LayoutMode = 'organization' | 'admin'

export function canShowWorkspaceSwitch(mode: LayoutMode, isAdmin: boolean): boolean {
  return mode === 'admin' || isAdmin
}
