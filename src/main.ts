import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import 'vuetify/styles'
import '@mdi/font/css/materialdesignicons.css'
import './styles.css'
import App from './App.vue'
import IconSelect from './components/IconSelect.vue'

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

const router = createRouter({
  history: createWebHistory(),
  scrollBehavior: () => ({ top: 0 }),
  routes: [
    { path: '/', redirect: '/organization' },
    {
      path: '/organization',
      component: () => import('./layouts/OrganizationLayout.vue'),
      children: [
        { path: '', component: () => import('./views/OrganizationOverviewView.vue'), meta: { section: 'Organization', title: 'Overview', subtitle: 'Your developer workspace' } },
        { path: 'repositories', component: () => import('./views/RepositoriesView.vue'), meta: { section: 'Organization', title: 'Repositories', subtitle: 'Accessible repositories and agent configuration' } },
        { path: 'sessions', component: () => import('./views/SessionsView.vue'), meta: { section: 'Organization', title: 'Sessions', subtitle: 'Recent conversations and runs' } },
        { path: 'sessions/:sessionId', component: () => import('./views/SessionDetailView.vue'), meta: { section: 'Sessions', title: 'OpenCode session', subtitle: 'Conversation, execution events, and repository changes' } },
        { path: 'skills', component: () => import('./views/EffectiveSkillsView.vue'), meta: { section: 'Organization', title: 'Skills', subtitle: 'Capabilities effectively applied to your repositories' } },
        { path: 'mcp', component: () => import('./views/EffectiveMcpView.vue'), meta: { section: 'Organization', title: 'MCP servers', subtitle: 'Catalog and activation of organization tool connectors' } },
        { path: 'secrets', component: () => import('./views/OrganizationSecretsView.vue'), meta: { section: 'Organization', title: 'Secrets', subtitle: 'Protected values, rotations, and organization bindings' } },
        { path: 'settings', component: () => import('./views/OrganizationSettingsView.vue'), meta: { section: 'Organization', title: 'Settings', subtitle: 'Settings available in this organization' } },
        { path: 'profile', component: () => import('./views/ProfileView.vue'), meta: { section: 'Account', title: 'Profile', subtitle: 'Personal information and account preferences' } },
      ],
    },
    {
      path: '/admin',
      component: () => import('./layouts/AdminLayout.vue'),
      children: [
        { path: '', component: () => import('./views/DashboardView.vue'), meta: { section: 'Administration', title: 'Overview', subtitle: 'Platform administration' } },
        { path: 'agents', component: () => import('./views/AgentsView.vue'), meta: { section: 'Administration', title: 'Agents', subtitle: 'Definitions, versions, and publishing' } },
        { path: 'secrets', component: () => import('./views/SecretsView.vue'), meta: { section: 'Administration', title: 'Secrets', subtitle: 'Protected values, rotations, and bindings' } },
        { path: 'skills', component: () => import('./views/SkillsView.vue'), meta: { section: 'Administration', title: 'Skills', subtitle: 'Catalog and activation policies' } },
        { path: 'mcp', component: () => import('./views/McpServersView.vue'), meta: { section: 'Administration', title: 'MCP servers', subtitle: 'Global catalog and organization tool connectors' } },
        { path: 'members', component: () => import('./views/MembersView.vue'), meta: { section: 'Administration', title: 'Members', subtitle: 'Organization access and roles' } },
        { path: 'runtimes', component: () => import('./views/RuntimesView.vue'), meta: { section: 'Administration', title: 'Images & adapters', subtitle: 'Approved runtime supply chain' } },
        { path: 'profile', component: () => import('./views/ProfileView.vue'), meta: { section: 'Account', title: 'Profile', subtitle: 'Personal information and account preferences' } },
      ],
    },
    {
      path: '/',
      component: () => import('./layouts/AuthLayout.vue'),
      children: [
        { path: 'login', component: () => import('./views/LoginView.vue') },
        { path: 'logout', component: () => import('./views/LogoutView.vue') },
      ],
    },
    { path: '/:pathMatch(.*)*', redirect: '/organization' },
  ],
})

createApp(App).component('IconSelect', IconSelect).use(router).use(vuetify).mount('#app')
