import { describe, expect, it } from 'vitest'
import { canContinueRepositoryStep } from './repositoryForm'

const emptyForm = { clone_url: '', name: '', default_branch: 'main' }

describe('repository form step validation', () => {
  it('allows provider selection to continue before a clone URL is entered', () => {
    expect(canContinueRepositoryStep(1, emptyForm)).toBe(true)
  })

  it('requires a valid HTTPS or SSH clone URL on the connection step', () => {
    expect(canContinueRepositoryStep(2, emptyForm)).toBe(false)
    expect(canContinueRepositoryStep(2, { ...emptyForm, clone_url: 'https://github.com/acme/app.git' })).toBe(true)
    expect(canContinueRepositoryStep(2, { ...emptyForm, clone_url: 'ssh://git@github.com/acme/app.git' })).toBe(true)
  })
})
