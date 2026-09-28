<script setup lang="ts">
import { computed, ref } from 'vue'
import DataTableEmptyRow from '../components/DataTableEmptyRow.vue'
import FilterCard from '../components/FilterCard.vue'
import SectionCard from '../components/SectionCard.vue'
import StatusChip from '../components/StatusChip.vue'
import TablePaginationCard from '../components/TablePaginationCard.vue'
import { useNotifications } from '../composables/useNotifications'
import { usePagination } from '../composables/usePagination'

const query = ref('')
const status = ref('All statuses')
const scopeFilter = ref('All scopes')
const showMoreFilters = ref(false)
const editorOpen = ref(false)
const previewOpen = ref(false)
type Agent = { name: string; description: string; version: string; scope: string; status: string; statusColor: string; runtime: string; updated: string; icon: string }
const selected = ref<Agent | null>(null)
const publishDialog = ref(false)
const { success, warning } = useNotifications()

const agents: Agent[] = [
  { name: 'Atlas', description: 'General-purpose agent for application development', version: 'v3', scope: 'Global', status: 'Published', statusColor: 'success', runtime: 'OpenCode 1.2', updated: '12 min ago', icon: 'mdi-creation-outline' },
  { name: 'Sentinel', description: 'Security analysis and dependency reviews', version: 'v2', scope: 'Northstar Labs', status: 'Published', statusColor: 'success', runtime: 'OpenCode 1.2', updated: 'Yesterday', icon: 'mdi-shield-search-outline' },
  { name: 'Pathfinder', description: 'Codebase exploration and documentation', version: 'v4', scope: 'Global', status: 'Published', statusColor: 'success', runtime: 'OpenCode 1.3-rc', updated: '3 hours ago', icon: 'mdi-map-search-outline' },
  { name: 'Forge', description: 'Go service migration and modernization', version: 'v1', scope: 'Northstar Labs', status: 'Draft', statusColor: 'info', runtime: 'OpenCode 1.2', updated: '2 days ago', icon: 'mdi-hammer-wrench' },
  { name: 'Archivist', description: 'Documentation maintenance and changelogs', version: 'v2', scope: 'Global', status: 'Suspended', statusColor: 'error', runtime: 'OpenCode 1.1', updated: '8 days ago', icon: 'mdi-bookshelf' },
]

const filtered = computed(() => agents.filter((agent) => {
  const matchesQuery = `${agent.name} ${agent.description}`.toLowerCase().includes(query.value.toLowerCase())
  const matchesStatus = status.value === 'All statuses' || agent.status === status.value
  const matchesScope = scopeFilter.value === 'All scopes' || agent.scope === scopeFilter.value
  return matchesQuery && matchesStatus && matchesScope
}))
const { page, paginatedItems: paginatedAgents, itemsPerPage } = usePagination(filtered)

function openAgent(agent?: Agent) {
  selected.value = agent ?? null
  editorOpen.value = true
}

function saveDraft() {
  editorOpen.value = false
  success('Draft saved', `${selected.value?.name ?? 'New agent'} was saved as a draft.`)
}

function publishAgent() {
  publishDialog.value = false
  editorOpen.value = false
  success('Agent published', `${selected.value?.name ?? 'New agent'} is now available for repository bindings.`)
}

function duplicateAgent(agent: Agent) {
  success('Agent duplicated', `A draft copy of ${agent.name} was created.`)
}

function suspendAgent(agent: Agent) {
  warning('Agent suspended', `${agent.name} is no longer available for new bindings.`)
}
</script>

<template>
  <FilterCard title="Filters" subtitle="Refine the displayed definitions" class="mb-4">
    <v-text-field v-model="query" hide-details placeholder="Search agents…" prepend-inner-icon="mdi-magnify" />
    <IconSelect v-model="status" hide-details :items="['All statuses', 'Published', 'Review', 'Draft', 'Suspended']" min-width="180" />
    <IconSelect v-if="showMoreFilters" v-model="scopeFilter" hide-details :items="['All scopes', 'Global', 'Northstar Labs']" min-width="180" />
    <v-spacer />
    <v-btn variant="text" prepend-icon="mdi-filter-variant" @click="showMoreFilters = !showMoreFilters">{{ showMoreFilters ? 'Fewer filters' : 'More filters' }}</v-btn>
  </FilterCard>

  <SectionCard title="Agent definitions" subtitle="5 definitions · 3 published">
    <template #actions>
      <v-btn color="primary" prepend-icon="mdi-plus" @click="openAgent()">New agent</v-btn>
    </template>
    <div class="table-scroll">
      <table class="data-table">
        <thead><tr><th>AGENT</th><th class="table-cell--center">SCOPE</th><th class="table-cell--center">RUNTIME</th><th class="table-cell--center">VERSION</th><th class="table-cell--center">STATUS</th><th class="table-cell--center">UPDATED</th><th class="table-cell--center" aria-label="Actions"></th></tr></thead>
        <tbody>
          <DataTableEmptyRow v-if="filtered.length === 0" :colspan="7" />
          <tr v-for="agent in paginatedAgents" :key="agent.name">
            <td><div class="entity-cell"><div class="entity-icon"><v-icon :icon="agent.icon" size="19" /></div><div><div class="entity-name">{{ agent.name }}</div><div class="entity-meta">{{ agent.description }}</div></div></div></td>
            <td class="table-cell--center"><v-chip size="small" variant="outlined" color="default">{{ agent.scope }}</v-chip></td>
            <td class="table-cell--center"><span class="text-medium-emphasis">{{ agent.runtime }}</span></td>
            <td class="table-cell--center"><span class="code-text">{{ agent.version }}</span></td>
            <td class="table-cell--center"><StatusChip :label="agent.status" :color="agent.statusColor" :icon="agent.status === 'Published' ? 'mdi-check-circle-outline' : 'mdi-circle-medium'" /></td>
            <td class="table-cell--center"><span class="text-medium-emphasis">{{ agent.updated }}</span></td>
            <td class="table-cell--center"><v-menu><template #activator="{ props }"><v-btn v-bind="props" icon="mdi-dots-horizontal" variant="text" size="small" aria-label="Actions" /></template><v-list class="pa-2" width="210"><v-list-item title="Edit" prepend-icon="mdi-pencil-outline" @click="openAgent(agent)" /><v-list-item title="Public preview" prepend-icon="mdi-eye-outline" @click="selected = agent; previewOpen = true" /><v-list-item title="Duplicate" prepend-icon="mdi-content-copy" @click="duplicateAgent(agent)" /><v-divider class="my-2" /><v-list-item title="Suspend" prepend-icon="mdi-pause-circle-outline" class="text-warning" @click="suspendAgent(agent)" /></v-list></v-menu></td>
          </tr>
        </tbody>
      </table>
    </div>
  </SectionCard>
  <TablePaginationCard v-model="page" :total="filtered.length" :items-per-page="itemsPerPage" item-label="agents" />

  <v-navigation-drawer v-model="editorOpen" location="right" temporary width="600" class="detail-drawer">
    <div class="drawer-header d-flex align-center justify-space-between">
      <div><div class="text-h6 font-weight-bold">{{ selected ? `Edit ${selected.name}` : 'New agent' }}</div><div class="text-caption text-medium-emphasis">Private administration definition</div></div>
      <v-btn icon="mdi-close" variant="text" aria-label="Close" @click="editorOpen = false" />
    </div>
    <div class="form-section">
      <div class="form-section__title">Tools and MCP servers</div>
      <IconSelect
        label="Allowed MCP servers"
        :items="['Filesystem workspace · Global · STDIO', 'Git provider · Global · SSE', 'Northstar issue tracker · Organization · SSE', 'Kubernetes read-only · Organization · STDIO']"
        :model-value="[]"
        multiple
        chips
        closable-chips
        prepend-inner-icon="mdi-server-network-outline"
        hint="No server is selected by default. Only servers enabled in the organization can be linked."
        persistent-hint
      />
      <div class="security-note mt-3"><v-icon icon="mdi-link-lock" color="secondary" /><span>Sensitive values remain references to Secrets. No private endpoint, command, or secret will be exposed in the public agent profile.</span></div>
    </div>
    <div class="form-section">
      <div class="form-section__title">General information</div>
      <v-text-field label="Agent name" :model-value="selected?.name ?? ''" placeholder="E.g. Atlas" />
      <v-textarea label="Public description" :model-value="selected?.description ?? ''" rows="2" />
      <div class="form-row"><IconSelect label="Scope" :items="['Global', 'Northstar Labs']" model-value="Northstar Labs" /><IconSelect label="Adapter" :items="['OpenCode 1.2', 'OpenCode 1.3-rc']" model-value="OpenCode 1.2" /></div>
    </div>
    <div class="form-section">
      <div class="form-section__title">Engine and model</div>
      <v-text-field label="Image OCI" model-value="registry.internal/opencode@sha256:••••b8e1" prepend-inner-icon="mdi-cube-outline" />
      <div class="form-row"><IconSelect label="LLM provider" :items="['OpenAI', 'OpenAI-compatible endpoint']" model-value="OpenAI" /><v-text-field label="Model" model-value="gpt-5.2-codex" /></div>
      <v-text-field label="OpenAI API endpoint" model-value="https://api.openai.com/v1" prepend-inner-icon="mdi-api" hint="OpenAI or OpenAI-compatible endpoint" persistent-hint />
      <v-text-field label="OpenAI API key" type="password" placeholder="Enter a new key" prepend-inner-icon="mdi-key-variant" autocomplete="new-password" hint="Write-only: the value will never be displayed again" persistent-hint />
      <div class="security-note mb-4"><v-icon icon="mdi-shield-key-outline" color="secondary" /><span>The key will be stored as the agent’s private credential. Only its reference will be used in snapshots.</span></div>
      <v-textarea label="System prompt" rows="3" model-value="You are a software development agent…" />
    </div>
    <div class="form-section">
      <div class="form-section__title">Maximum resources</div>
      <div class="form-row"><v-text-field label="CPU" model-value="2 vCPU" /><v-text-field label="Memory" model-value="4 GiB" /></div>
      <div class="form-row"><v-text-field label="Storage" model-value="10 GiB" /><v-text-field label="Duration" model-value="60 min" /></div>
    </div>
    <div class="form-section">
      <div class="form-section__title">User-facing projection</div>
      <div class="security-note"><v-icon icon="mdi-shield-lock-outline" color="info" /><span>The prompt, credentials, internal provider, and network rules will remain absent from the public projection.</span></div>
      <v-btn block variant="outlined" class="mt-4" prepend-icon="mdi-eye-outline" @click="previewOpen = true">Preview public profile</v-btn>
    </div>
    <template #append>
      <div class="dialog-actions d-flex justify-end ga-2"><v-btn variant="text" @click="editorOpen = false">Cancel</v-btn><v-btn variant="outlined" @click="saveDraft">Save draft</v-btn><v-btn color="primary" @click="publishDialog = true">Validate & publish</v-btn></div>
    </template>
  </v-navigation-drawer>

  <v-dialog v-model="previewOpen" max-width="520">
    <v-card class="section-card">
      <div class="pa-6">
        <div class="d-flex align-start justify-space-between"><div class="entity-icon mb-4"><v-icon :icon="selected?.icon ?? 'mdi-robot-outline'" /></div><StatusChip label="Available" color="success" icon="mdi-check-circle-outline" /></div>
        <h2 class="text-h5 font-weight-bold mb-2">{{ selected?.name ?? 'New agent' }}</h2>
        <p class="text-body-2 text-medium-emphasis">{{ selected?.description ?? 'Agent description as it will appear to members.' }}</p>
        <v-divider class="my-5" />
        <div class="eyebrow mb-3">PUBLIC CAPABILITIES</div>
        <div class="d-flex flex-wrap ga-2"><v-chip size="small">Code editing</v-chip><v-chip size="small">Tests</v-chip><v-chip size="small">Documentation</v-chip></div>
        <div class="security-note mt-5"><v-icon icon="mdi-eye-off-outline" color="secondary" /><span>Sensitive settings are correctly hidden in this preview.</span></div>
      </div>
      <v-card-actions class="dialog-actions"><v-spacer /><v-btn @click="previewOpen = false">Close</v-btn></v-card-actions>
    </v-card>
  </v-dialog>

  <v-dialog v-model="publishDialog" max-width="460">
    <v-card class="section-card pa-2"><div class="pa-5"><div class="entity-icon mb-4"><v-icon icon="mdi-upload-outline" /></div><h3 class="mb-2">Publish this version?</h3><p class="text-body-2 text-medium-emphasis">It will become selectable for new repository bindings. Existing sessions will not be changed.</p></div><v-card-actions class="dialog-actions"><v-spacer /><v-btn @click="publishDialog = false">Cancel</v-btn><v-btn color="primary" @click="publishAgent">Publish</v-btn></v-card-actions></v-card>
  </v-dialog>
</template>
