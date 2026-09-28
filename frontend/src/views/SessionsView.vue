<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'
import DataTableEmptyRow from '../components/DataTableEmptyRow.vue'
import FilterCard from '../components/FilterCard.vue'
import NewSessionDialog, { type CreatedSession } from '../components/NewSessionDialog.vue'
import SectionCard from '../components/SectionCard.vue'
import StatusChip from '../components/StatusChip.vue'
import TablePaginationCard from '../components/TablePaginationCard.vue'
import { useNotifications } from '../composables/useNotifications'
import { usePagination } from '../composables/usePagination'

const sessionOpen = ref(false)
const route = useRoute()
const organizationSlug = computed(() => String(route.params.organizationSlug ?? ''))
const query = ref('')
const repositoryFilter = ref('All repositories')
const statusFilter = ref('All statuses')
const { success } = useNotifications()
const sessions = ref([
  { id: 'oauth-migration', title: 'Complete the OAuth migration', repo: 'identity-service', branch: 'feat/oidc', agent: 'Atlas', run: 'In progress', updated: '8 min ago' },
  { id: 'manifest-update', title: 'Update manifests', repo: 'platform-k8s', branch: 'main', agent: 'Pathfinder', run: 'Pending', updated: '36 min ago' },
  { id: 'payment-review', title: 'Review the payment module', repo: 'checkout-api', branch: 'main', agent: 'Sentinel', run: 'Completed', updated: 'Yesterday' },
  { id: 'api-documentation', title: 'Document the endpoints', repo: 'identity-service', branch: 'docs/api', agent: 'Archivist', run: 'Interrupted', updated: '3 days ago' },
])
const filteredSessions = computed(() => sessions.value.filter((session) => {
  const search = query.value.trim().toLowerCase()
  const matchesSearch = !search || `${session.title} ${session.repo} ${session.branch} ${session.agent}`.toLowerCase().includes(search)
  const matchesRepository = repositoryFilter.value === 'All repositories' || session.repo === repositoryFilter.value
  const matchesStatus = statusFilter.value === 'All statuses' || session.run === statusFilter.value
  return matchesSearch && matchesRepository && matchesStatus
}))
const { page, paginatedItems: paginatedSessions, itemsPerPage } = usePagination(filteredSessions)

function createSession(session: CreatedSession) {
  const title = session.firstMessage.length > 64 ? `${session.firstMessage.slice(0, 61)}…` : session.firstMessage
  sessions.value.unshift({ id: `session-${Date.now()}`, title, repo: session.repository, branch: session.branch, agent: session.profile, run: 'Pending', updated: 'Now' })
  success('Session created', `${session.profile} is queued for ${session.repository}.`)
}
</script>

<template>
  <FilterCard title="Filters" subtitle="Find a conversation by repository or status" class="mb-4"><v-text-field v-model="query" hide-details placeholder="Search sessions…" prepend-inner-icon="mdi-magnify" /><IconSelect v-model="repositoryFilter" hide-details :items="['All repositories', 'identity-service', 'platform-k8s', 'checkout-api']" /><IconSelect v-model="statusFilter" hide-details :items="['All statuses', 'In progress', 'Pending', 'Completed', 'Interrupted']" /><v-spacer /><v-btn color="primary" prepend-icon="mdi-plus" @click="sessionOpen = true">New session</v-btn></FilterCard>
  <SectionCard title="Sessions" subtitle="Conversation and run history"><div class="table-scroll"><table class="data-table"><thead><tr><th>SESSION</th><th>REPOSITORY</th><th class="table-cell--center">BRANCH</th><th class="table-cell--center">AGENT</th><th class="table-cell--center">STATUS</th><th class="table-cell--center">UPDATED</th><th class="table-cell--center" aria-label="Actions"></th></tr></thead><tbody><DataTableEmptyRow v-if="filteredSessions.length === 0" :colspan="7" /><tr v-for="session in paginatedSessions" :key="session.id"><td><div class="entity-cell"><div class="entity-icon"><v-icon icon="mdi-message-text-outline" /></div><div><div class="entity-name">{{ session.title }}</div><div class="entity-meta">Persistent conversation</div></div></div></td><td class="code-text">{{ session.repo }}</td><td class="table-cell--center"><v-chip size="small" variant="outlined">{{ session.branch }}</v-chip></td><td class="table-cell--center">{{ session.agent }}</td><td class="table-cell--center"><StatusChip :label="session.run" :color="session.run === 'Completed' ? 'success' : session.run === 'Interrupted' ? 'error' : session.run === 'Pending' ? 'warning' : 'info'" /></td><td class="text-medium-emphasis table-cell--center">{{ session.updated }}</td><td class="table-cell--center"><v-btn :to="`/organizations/${organizationSlug}/sessions/${session.id}`" icon="mdi-arrow-right" variant="text" aria-label="Open session" /></td></tr></tbody></table></div></SectionCard>
  <TablePaginationCard v-model="page" :total="filteredSessions.length" :items-per-page="itemsPerPage" item-label="sessions" hint="Persistent history" />

  <NewSessionDialog v-model="sessionOpen" @created="createSession" />
</template>
