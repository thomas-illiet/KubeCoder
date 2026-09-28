export interface SwaggerRequest {
  headers?: Record<string, string>
}

export interface SwaggerOptions {
  url: string
  defaultModelsExpandDepth: -1
  persistAuthorization: false
  requestInterceptor(request: SwaggerRequest): SwaggerRequest
}

export function createSwaggerOptions(apiBaseUrl: string, getAccessToken: () => string): SwaggerOptions {
  return {
    url: `${apiBaseUrl}/openapi.yaml`,
    defaultModelsExpandDepth: -1,
    persistAuthorization: false,
    requestInterceptor(request) {
      const token = getAccessToken()
      if (token) request.headers = { ...request.headers, Authorization: `Bearer ${token}` }
      return request
    },
  }
}
