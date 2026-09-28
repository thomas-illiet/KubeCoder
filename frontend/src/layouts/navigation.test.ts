import { describe, expect, it } from 'vitest'
import { canShowWorkspaceSwitch } from './navigation'

describe('workspace navigation', () => {
  it('hides administration from non-administrators', () => {
    expect(canShowWorkspaceSwitch('organization', false)).toBe(false)
    expect(canShowWorkspaceSwitch('organization', true)).toBe(true)
  })

  it('always lets users leave the administration layout', () => {
    expect(canShowWorkspaceSwitch('admin', true)).toBe(true)
  })
})
