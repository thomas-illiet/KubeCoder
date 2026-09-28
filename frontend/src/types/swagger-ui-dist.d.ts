declare module 'swagger-ui-dist/swagger-ui-es-bundle.js' {
  import type { SwaggerRequest } from '../documentation/swagger'

  interface SwaggerUIInstance {
    destroy?: () => void
  }

  interface SwaggerUIConfiguration {
    domNode: HTMLElement
    url: string
    defaultModelsExpandDepth?: number
    deepLinking?: boolean
    displayRequestDuration?: boolean
    persistAuthorization?: boolean
    requestInterceptor?: (request: SwaggerRequest) => SwaggerRequest
    presets?: unknown[]
    layout?: string
  }

  interface SwaggerUIBundleFactory {
    (configuration: SwaggerUIConfiguration): SwaggerUIInstance
    presets: { apis: unknown }
  }

  const SwaggerUIBundle: SwaggerUIBundleFactory
  export default SwaggerUIBundle
}

declare module 'swagger-ui-dist/swagger-ui.css'
