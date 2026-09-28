import { describe, expect, it, vi } from 'vitest'
import { createSwaggerOptions, type SwaggerRequest } from './swagger'

describe('Swagger configuration', () => {
  it('uses the configured API and injects the current access token', () => {
    const getAccessToken = vi.fn(() => 'current-token')
    const options = createSwaggerOptions('http://api.test', getAccessToken)
    const request: SwaggerRequest = { headers: { Accept: 'application/json' } }

    expect(options.url).toBe('http://api.test/openapi.yaml')
    expect(options.defaultModelsExpandDepth).toBe(-1)
    expect(options.persistAuthorization).toBe(false)
    expect(options.requestInterceptor(request)).toEqual({
      headers: { Accept: 'application/json', Authorization: 'Bearer current-token' },
    })
    expect(getAccessToken).toHaveBeenCalledOnce()
  })

  it('does not add an authorization header without a token', () => {
    const options = createSwaggerOptions('http://api.test', () => '')
    expect(options.requestInterceptor({})).toEqual({})
  })
})
