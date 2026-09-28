import 'vue-router'

declare module 'vue-router' {
  interface RouteMeta {
    public?: boolean
    requiresAdmin?: boolean
    section?: string
    title?: string
    subtitle?: string
  }
}

export {}
