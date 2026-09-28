import { describe, expect, it } from 'vitest'
import { authenticationError, normalizeReturnTo } from './navigation'

describe('OIDC navigation', () => {
  it('restores an internal destination after a callback', () => {
    expect(normalizeReturnTo('/organization/sessions?state=active')).toBe('/organization/sessions?state=active')
  })

  it('rejects external and malformed callback destinations', () => {
    expect(normalizeReturnTo('https://example.test')).toBe('/organization')
    expect(normalizeReturnTo('//example.test')).toBe('/organization')
    expect(normalizeReturnTo(undefined)).toBe('/organization')
  })

  it('exposes callback and logout errors without discarding their message', () => {
    expect(authenticationError(new Error('invalid state'))).toBe('invalid state')
    expect(authenticationError(null)).toBe('Unexpected OpenID Connect error')
  })
})

