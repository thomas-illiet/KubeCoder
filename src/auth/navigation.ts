export function normalizeReturnTo(value: unknown): string {
  return typeof value === 'string' && value.startsWith('/') && !value.startsWith('//')
    ? value
    : '/organization'
}

export function authenticationError(error: unknown): string {
  return error instanceof Error ? error.message : 'Unexpected OpenID Connect error'
}

