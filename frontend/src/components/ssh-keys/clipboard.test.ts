import { describe, expect, it, vi } from 'vitest'
import { copyPublicKey } from './clipboard'

describe('SSH public key clipboard', () => {
  it('copies the complete public key without exposing private material', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined)
    await copyPublicKey('ssh-ed25519 AAAA organization', { writeText })
    expect(writeText).toHaveBeenCalledWith('ssh-ed25519 AAAA organization')
  })
})
