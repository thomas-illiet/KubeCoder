<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useDisplay } from 'vuetify'
import { useRoute, useRouter } from 'vue-router'
import { useNotifications } from '../composables/useNotifications'
import { useAuth } from '../composables/useAuth'

type LayoutMode = 'organization' | 'admin'
type NavigationItem = { title: string; icon: string; to: string }
type NavigationGroup = { title: string; items: NavigationItem[] }

const props = defineProps<{ mode: LayoutMode }>()
const { smAndDown: mobile } = useDisplay()
const route = useRoute()
const router = useRouter()
const { success } = useNotifications()
const { displayName, username, initials } = useAuth()
const drawer = ref(!mobile.value)
const rail = ref(false)
const commandOpen = ref(false)
const commandQuery = ref('')
const vibesActive = ref(false)
const organizationOpen = ref(false)
const organizationQuery = ref('')
const organization = ref('Northstar Labs')
let vibesTimer: ReturnType<typeof setTimeout> | undefined

watch(mobile, (isMobile) => {
  drawer.value = !isMobile
})

const vibePods = [
  { icon: 'mdi-cube-outline', x: '8%', y: '18%', delay: '-.4s', duration: '4.8s' },
  { icon: 'mdi-hexagon-multiple-outline', x: '20%', y: '72%', delay: '-1.8s', duration: '5.6s' },
  { icon: 'mdi-cube-outline', x: '38%', y: '12%', delay: '-2.6s', duration: '5.2s' },
  { icon: 'mdi-robot-happy-outline', x: '60%', y: '78%', delay: '-1.2s', duration: '5.9s' },
  { icon: 'mdi-hexagon-multiple-outline', x: '76%', y: '16%', delay: '-3.1s', duration: '5.4s' },
  { icon: 'mdi-cube-outline', x: '90%', y: '66%', delay: '-2.2s', duration: '4.9s' },
]

const organizations = [
  { name: 'Northstar Labs', slug: 'northstar-labs', role: 'Owner', initials: 'NL' },
  { name: 'Orbit Systems', slug: 'orbit-systems', role: 'Admin', initials: 'OS' },
  { name: 'Acme Engineering', slug: 'acme-engineering', role: 'Member', initials: 'AE' },
  { name: 'Lumen Platform', slug: 'lumen-platform', role: 'Admin', initials: 'LP' },
  { name: 'Nova Research', slug: 'nova-research', role: 'Member', initials: 'NR' },
]

const organizationGroups: NavigationGroup[] = [
  {
    title: 'General',
    items: [
      { title: 'Overview', icon: 'mdi-view-dashboard-outline', to: '/organization' },
    ],
  },
  {
    title: 'Workspace',
    items: [
      { title: 'Repositories', icon: 'mdi-source-repository', to: '/organization/repositories' },
      { title: 'Sessions', icon: 'mdi-message-text-outline', to: '/organization/sessions' },
    ],
  },
  {
    title: 'Configuration',
    items: [
      { title: 'Skills', icon: 'mdi-puzzle-outline', to: '/organization/skills' },
      { title: 'MCP servers', icon: 'mdi-server-network-outline', to: '/organization/mcp' },
      { title: 'Secrets', icon: 'mdi-key-variant', to: '/organization/secrets' },
    ],
  },
  {
    title: 'Organization',
    items: [
      { title: 'Settings', icon: 'mdi-tune-variant', to: '/organization/settings' },
    ],
  },
]

const adminGroups: NavigationGroup[] = [
  {
    title: 'General',
    items: [
      { title: 'Overview', icon: 'mdi-shield-crown-outline', to: '/admin' },
    ],
  },
  {
    title: 'Agent platform',
    items: [
      { title: 'Agents', icon: 'mdi-robot-outline', to: '/admin/agents' },
      { title: 'Skills', icon: 'mdi-puzzle-outline', to: '/admin/skills' },
      { title: 'MCP servers', icon: 'mdi-server-network-outline', to: '/admin/mcp' },
    ],
  },
  {
    title: 'Security & access',
    items: [
      { title: 'Secrets', icon: 'mdi-key-variant', to: '/admin/secrets' },
      { title: 'Members', icon: 'mdi-account-group-outline', to: '/admin/members' },
    ],
  },
  {
    title: 'Infrastructure',
    items: [
      { title: 'Images & adapters', icon: 'mdi-cube-outline', to: '/admin/runtimes' },
    ],
  },
]

const navigationGroups = computed(() => props.mode === 'admin' ? adminGroups : organizationGroups)
const navigationItems = computed(() => navigationGroups.value.flatMap((group) => group.items))
const filteredNavigationItems = computed(() => {
  const term = commandQuery.value.trim().toLocaleLowerCase('en')
  if (!term) return navigationItems.value
  return navigationItems.value.filter((item) => item.title.toLocaleLowerCase('en').includes(term))
})
const vibesCommandReady = computed(() => commandQuery.value.trim().toLocaleLowerCase('en') === 'kubectl get vibes')
const layoutLabel = computed(() => props.mode === 'admin' ? 'ADMINISTRATION' : 'ORGANIZATION')
const switchTarget = computed(() => props.mode === 'admin' ? '/organization' : '/admin')
const switchTitle = computed(() => props.mode === 'admin' ? 'Back to organization' : 'Administration')
const switchSubtitle = computed(() => props.mode === 'admin' ? 'Return to the developer workspace' : 'Open the dedicated console')
const switchIcon = computed(() => props.mode === 'admin' ? 'mdi-arrow-left' : 'mdi-shield-crown-outline')
const profileTarget = computed(() => props.mode === 'admin' ? '/admin/profile' : '/organization/profile')
const title = computed(() => route.meta.title as string)
const subtitle = computed(() => route.meta.subtitle as string)
const currentOrganization = computed(() => organizations.find((item) => item.name === organization.value) ?? organizations[0])
const organizationResults = computed(() => {
  const term = organizationQuery.value.trim().toLocaleLowerCase('en')
  if (!term) return organizations.slice(0, 3)
  return organizations.filter((item) => `${item.name} ${item.slug}`.toLocaleLowerCase('en').includes(term)).slice(0, 5)
})

function selectOrganization(name: string) {
  organization.value = name
  organizationOpen.value = false
  organizationQuery.value = ''
}

function toggleDrawer() {
  if (mobile.value) drawer.value = !drawer.value
  else rail.value = !rail.value
}

function openNavigationItem(item: NavigationItem) {
  router.push(item.to)
  commandOpen.value = false
  commandQuery.value = ''
}

function stopVibes(reconciled = false) {
  if (!vibesActive.value) return
  vibesActive.value = false
  if (vibesTimer) clearTimeout(vibesTimer)
  vibesTimer = undefined
  if (reconciled) success('Cluster vibes successfully reconciled ✨', 'All pods are still healthy. Probably.')
}

function triggerVibes() {
  if (!vibesCommandReady.value) return
  commandOpen.value = false
  commandQuery.value = ''
  vibesActive.value = true
  if (vibesTimer) clearTimeout(vibesTimer)
  vibesTimer = setTimeout(() => stopVibes(true), 8000)
}

function handleEscape(event: KeyboardEvent) {
  if (event.key === 'Escape' && vibesActive.value) stopVibes()
}

onMounted(() => window.addEventListener('keydown', handleEscape))
onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleEscape)
  if (vibesTimer) clearTimeout(vibesTimer)
})

function resetCommandSearch() {
  commandQuery.value = ''
}
</script>

<template>
  <v-app :class="['workspace-layout', `workspace-layout--${mode}`, { 'workspace-layout--vibing': vibesActive }]">
    <v-navigation-drawer v-model="drawer" :rail="rail && !mobile" :temporary="mobile" mobile-breakpoint="md" width="270" class="app-drawer">
      <div class="brand" :class="{ 'brand--rail': rail && !mobile }">
        <div class="brand__mark"><v-icon :icon="mode === 'admin' ? 'mdi-shield-crown-outline' : 'mdi-kubernetes'" size="26" /></div>
        <div v-if="!rail || mobile" class="brand__copy"><span class="brand__name">KubeCoder</span><span class="brand__tag">{{ mode === 'admin' ? 'ADMIN CONSOLE' : 'ORGANIZATION WORKSPACE' }}</span></div>
      </div>

      <v-list nav density="comfortable" class="nav-list px-3">
        <template v-for="(group, groupIndex) in navigationGroups" :key="group.title">
          <v-list-subheader v-if="!rail || mobile" class="nav-group-title">{{ group.title }}</v-list-subheader>
          <v-divider v-else-if="groupIndex > 0" class="nav-group-divider" />
          <v-list-item v-for="item in group.items" :key="item.to" :prepend-icon="item.icon" :title="item.title" :to="item.to" :exact="item.to === '/admin' || item.to === '/organization'" rounded="lg" />
        </template>
      </v-list>

      <template #append>
        <div class="px-3 pb-4">
          <v-list nav density="compact" class="layout-switch pa-0 mb-3">
            <v-list-item :prepend-icon="switchIcon" :title="switchTitle" :subtitle="switchSubtitle" :to="switchTarget" rounded="lg" />
          </v-list>
        </div>
      </template>
    </v-navigation-drawer>

    <v-app-bar flat height="74" class="app-bar px-2 px-sm-4">
      <v-btn icon="mdi-menu" variant="text" aria-label="Collapse navigation" @click="toggleDrawer" />
      <v-btn v-if="mode === 'organization'" variant="text" class="org-switcher ml-1 ml-sm-3" @click="organizationOpen = true"><div class="org-avatar org-avatar--small">{{ currentOrganization.initials }}</div><span class="d-none d-sm-inline ml-2">{{ organization }}</span><v-icon icon="mdi-chevron-down" size="18" class="ml-1" /></v-btn>
      <v-spacer />
      <v-btn class="search-trigger d-none d-md-flex" variant="outlined" color="default" @click="commandOpen = true"><v-icon icon="mdi-magnify" size="20" class="mr-2" />Search<span class="shortcut ml-8">⌘ K</span></v-btn>
      <v-btn icon="mdi-magnify" variant="text" class="d-md-none" aria-label="Search" @click="commandOpen = true" />
      <v-menu><template #activator="{ props: menuProps }"><v-btn v-bind="menuProps" icon class="ml-1 user-avatar" aria-label="Profile menu">{{ initials }}</v-btn></template><v-list width="230" class="pa-2"><v-list-item :title="displayName" :subtitle="username || 'OpenID Connect user'" /><v-divider class="my-2" /><v-list-item title="Profile" prepend-icon="mdi-account-circle-outline" @click="router.push(profileTarget)" /><v-list-item title="Sign out" prepend-icon="mdi-logout" @click="router.push('/logout')" /></v-list></v-menu>
    </v-app-bar>

    <v-main>
      <div class="page-shell">
        <div class="breadcrumb mb-3"><span>{{ mode === 'admin' ? 'KubeCoder' : organization }}</span><v-icon icon="mdi-chevron-right" size="15" /><span>{{ route.meta.section }}</span><v-icon icon="mdi-chevron-right" size="15" /><strong>{{ title }}</strong></div>
        <div class="page-title-row"><div><h1>{{ title }}</h1><p>{{ subtitle }}</p></div></div>
        <router-view />
      </div>
    </v-main>

    <v-dialog v-model="commandOpen" max-width="640" @after-leave="resetCommandSearch">
      <v-card class="command-dialog">
        <v-text-field v-model="commandQuery" autofocus hide-details :placeholder="`Search ${layoutLabel.toLocaleLowerCase('en')}…`" prepend-inner-icon="mdi-magnify" variant="solo" flat @keydown.enter.prevent="triggerVibes" />
        <v-divider />
        <div class="pa-3">
          <template v-if="vibesCommandReady">
            <div class="eyebrow px-2 mb-2">HIDDEN COMMAND DISCOVERED</div>
            <v-list bg-color="transparent"><v-list-item prepend-icon="mdi-party-popper" title="kubectl get vibes" subtitle="Reconcile the cluster's current vibe" append-icon="mdi-keyboard-return" rounded="lg" class="vibes-command" @click="triggerVibes" /></v-list>
          </template>
          <template v-else>
            <div class="eyebrow px-2 mb-2">QUICK ACCESS · {{ layoutLabel }}</div>
            <v-list v-if="filteredNavigationItems.length" bg-color="transparent"><v-list-item v-for="item in filteredNavigationItems" :key="item.to" :prepend-icon="item.icon" :title="item.title" append-icon="mdi-arrow-right" @click="openNavigationItem(item)" /></v-list>
            <div v-else class="command-empty"><v-icon icon="mdi-console-line" size="20" /><span>Command not found</span></div>
          </template>
        </div>
      </v-card>
    </v-dialog>

    <v-dialog v-if="mode === 'organization'" v-model="organizationOpen" max-width="640">
      <v-card class="app-dialog organization-dialog">
        <div class="dialog-header"><div><div class="text-h6 font-weight-bold">Switch organization</div><div class="text-caption text-medium-emphasis">Search across 2,437 accessible organizations</div></div><v-btn icon="mdi-close" variant="text" aria-label="Close" @click="organizationOpen = false" /></div>
        <div class="pa-4"><v-text-field v-model="organizationQuery" autofocus hide-details placeholder="Organization name or identifier…" prepend-inner-icon="mdi-magnify" clearable /><div class="d-flex align-center justify-space-between mt-4 mb-2 px-2"><div class="eyebrow">{{ organizationQuery ? 'RESULTS' : 'RECENT ORGANIZATIONS' }}</div><div class="text-caption text-medium-emphasis">5 results maximum</div></div><v-list bg-color="transparent"><v-list-item v-for="item in organizationResults" :key="item.slug" :title="item.name" :subtitle="`${item.slug} · ${item.role}`" rounded="lg" @click="selectOrganization(item.name)"><template #prepend><div class="org-avatar mr-3">{{ item.initials }}</div></template><template #append><v-icon v-if="organization === item.name" icon="mdi-check-circle" color="success" /></template></v-list-item></v-list><div v-if="organizationQuery && organizationResults.length === 0" class="empty-state py-8"><div class="empty-state__icon"><v-icon icon="mdi-domain-off" /></div><h3>No organizations found</h3><p>Try a different name or identifier.</p></div><div class="security-note mt-3"><v-icon icon="mdi-server-network-outline" color="info" /><span>Search will run server-side with pagination; the full list will never be loaded in the browser.</span></div></div>
        <v-card-actions class="dialog-actions"><v-btn variant="text" prepend-icon="mdi-open-in-new" @click="router.push('/organization'); organizationOpen = false">Open organization workspace</v-btn><v-spacer /><v-btn variant="text" prepend-icon="mdi-cog-outline" @click="router.push('/organization/settings'); organizationOpen = false">Available settings</v-btn></v-card-actions>
      </v-card>
    </v-dialog>

    <Transition name="vibes">
      <div v-if="vibesActive" class="vibes-overlay" role="dialog" aria-modal="true" aria-label="Cluster party mode" @click.self="stopVibes()">
        <div class="vibes-aurora vibes-aurora--one" />
        <div class="vibes-aurora vibes-aurora--two" />
        <div v-for="(pod, index) in vibePods" :key="index" class="vibes-pod" :style="{ '--pod-x': pod.x, '--pod-y': pod.y, '--pod-delay': pod.delay, '--pod-duration': pod.duration }">
          <v-icon :icon="pod.icon" size="22" />
        </div>
        <div class="vibes-terminal" @click.stop>
          <div class="vibes-terminal__bar">
            <div class="vibes-terminal__lights"><span /><span /><span /></div>
            <div class="vibes-terminal__title"><v-icon icon="mdi-kubernetes" size="17" /> cluster-party</div>
            <div class="vibes-terminal__live"><span /> LIVE</div>
          </div>
          <div class="vibes-terminal__body">
            <div class="vibes-terminal__command"><span>alex@kubecoder</span>:<strong>~</strong>$ kubectl get vibes</div>
            <div class="vibes-terminal__table" aria-label="Cluster vibe status">
              <div class="vibes-terminal__row vibes-terminal__row--header"><span>NAME</span><span>READY</span><span>STATUS</span></div>
              <div class="vibes-terminal__row"><span>atlas</span><span>1/1</span><strong>Vibing</strong></div>
              <div class="vibes-terminal__row"><span>opencode</span><span>1/1</span><strong>Dancing</strong></div>
              <div class="vibes-terminal__row"><span>coffee</span><span>3/3</span><strong>Overprovisioned</strong></div>
            </div>
            <div class="vibes-terminal__message"><v-icon icon="mdi-auto-fix" size="16" /> Reconciliation in progress<span class="vibes-terminal__dots">...</span></div>
          </div>
          <div class="vibes-terminal__footer"><span>Press Esc to restore professional mode</span><v-btn size="small" variant="text" prepend-icon="mdi-keyboard-esc" @click="stopVibes()">Escape</v-btn></div>
        </div>
      </div>
    </Transition>
  </v-app>
</template>
