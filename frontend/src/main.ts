import { createApp, watch } from 'vue'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import 'vuetify/styles'
import '@mdi/font/css/materialdesignicons.css'
import './styles.css'
import App from './App.vue'
import IconSelect from './components/IconSelect.vue'
import StartupErrorView from './views/StartupErrorView.vue'
import { createAuth, installAuth } from './auth/oidc'
import { installAuthGuard } from './auth/guard'
import { router } from './router'
import { loadRuntimeConfig } from './runtime/config'

const vuetify = createVuetify({
  components,
  directives,
  theme: {
    defaultTheme: 'kubecoderDark',
    themes: {
      kubecoderDark: {
        dark: true,
        colors: {
          background: '#0a0d12',
          surface: '#171c24',
          'surface-bright': '#252c38',
          primary: '#8193ff',
          secondary: '#55c8bc',
          success: '#31d49d',
          warning: '#e9ad5a',
          error: '#fb7185',
          info: '#60a5fa',
        },
      },
    },
  },
  defaults: {
    VBtn: { rounded: 'lg', textTransform: 'none' },
    VCard: { rounded: 'lg', elevation: 0 },
    VTextField: { variant: 'outlined', density: 'comfortable', color: 'primary' },
    VSelect: { variant: 'outlined', density: 'comfortable', color: 'primary' },
    VTextarea: { variant: 'outlined', color: 'primary' },
  },
})

async function bootstrap() {
  const config = await loadRuntimeConfig()
  const auth = createAuth(config)
  await auth.initialize()
  installAuthGuard(router, auth)
  watch(auth.isAuthenticated, (authenticated, wasAuthenticated) => {
    if (wasAuthenticated && !authenticated) void router.replace('/login')
  })
  const app = createApp(App).component('IconSelect', IconSelect).use(router).use(vuetify)
  installAuth(app, auth)
  app.mount('#app')
}

void bootstrap().catch((error: unknown) => {
  const message = error instanceof Error ? error.message : 'The runtime configuration could not be loaded.'
  createApp(StartupErrorView, { message }).use(vuetify).mount('#app')
})
