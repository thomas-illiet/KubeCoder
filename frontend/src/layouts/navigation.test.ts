import { describe, expect, it } from 'vitest'
import { canShowWorkspaceSwitch, documentationTarget, isSearchShortcut } from './navigation'

describe('workspace navigation', () => {
  it('hides administration from non-administrators', () => {
    expect(canShowWorkspaceSwitch('organization', false)).toBe(false)
    expect(canShowWorkspaceSwitch('organization', true)).toBe(true)
  })

  it('always lets users leave the administration layout', () => {
    expect(canShowWorkspaceSwitch('admin', true)).toBe(true)
  })

  it('selects the documentation route for the active workspace', () => {
    expect(documentationTarget('organization', 'northstar-labs')).toBe('/organizations/northstar-labs/documentation')
    expect(documentationTarget('admin')).toBe('/admin/documentation')
  })

  it('recognizes the search shortcut on macOS and other platforms', () => {
    expect(isSearchShortcut({ key: 'k', metaKey: true, ctrlKey: false })).toBe(true)
    expect(isSearchShortcut({ key: 'K', metaKey: false, ctrlKey: true })).toBe(true)
    expect(isSearchShortcut({ key: 'k', metaKey: false, ctrlKey: false })).toBe(false)
  })
})
