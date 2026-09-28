<script setup lang="ts">
import { onMounted, shallowRef, watch } from 'vue'
import { createAgent, deleteAgent, fetchAgents, updateAgent, type Agent, type AgentInput } from '../api/agents'
import AgentFormDialog from '../components/agents/AgentFormDialog.vue'
import DataTableEmptyRow from '../components/DataTableEmptyRow.vue'
import FilterCard from '../components/FilterCard.vue'
import SectionCard from '../components/SectionCard.vue'
import StatusChip from '../components/StatusChip.vue'
import TablePaginationCard from '../components/TablePaginationCard.vue'
import { useAuth } from '../composables/useAuth'
import { useNotifications } from '../composables/useNotifications'
import { useRuntimeConfig } from '../runtime/config'

const auth = useAuth()
const config = useRuntimeConfig()
const notifications = useNotifications()
const items = shallowRef<Agent[]>([])
const total = shallowRef(0)
const loading = shallowRef(false)
const saving = shallowRef(false)
const query = shallowRef('')
const status = shallowRef<'all' | 'active' | 'inactive'>('all')
const page = shallowRef(1)
const itemsPerPage = 20
const editorOpen = shallowRef(false)
const selected = shallowRef<Agent | null>(null)
const deleteTarget = shallowRef<Agent | null>(null)
const deleteConfirmation = shallowRef('')
let revision = 0

async function load(current = ++revision): Promise<void> {
  loading.value = true
  try {
    const result = await fetchAgents(config.apiBaseUrl, auth.accessToken.value, { query: query.value.trim(), active: status.value === 'all' ? undefined : status.value === 'active', limit: itemsPerPage, offset: (page.value - 1) * itemsPerPage })
    if (current !== revision) return
    items.value = result.items ?? []
    total.value = result.total
  } catch (error) {
    if (current === revision) notifications.error('Agents could not be loaded', error instanceof Error ? error.message : undefined)
  } finally { if (current === revision) loading.value = false }
}

function openCreate(): void { selected.value = null; editorOpen.value = true }
function openEdit(agent: Agent): void { selected.value = agent; editorOpen.value = true }
async function save(input: AgentInput): Promise<void> {
  saving.value = true
  try {
    if (selected.value) await updateAgent(config.apiBaseUrl, auth.accessToken.value, selected.value.id, input)
    else await createAgent(config.apiBaseUrl, auth.accessToken.value, input)
    editorOpen.value = false
    notifications.success(selected.value ? 'Agent updated' : 'Agent created')
    await load()
  } catch (error) { notifications.error('Agent could not be saved', error instanceof Error ? error.message : undefined) }
  finally { saving.value = false }
}
function requestDelete(agent: Agent): void { deleteTarget.value = agent; deleteConfirmation.value = '' }
async function confirmDelete(): Promise<void> {
  if (!deleteTarget.value || deleteConfirmation.value !== deleteTarget.value.name) return
  try { await deleteAgent(config.apiBaseUrl, auth.accessToken.value, deleteTarget.value.id); deleteTarget.value = null; notifications.success('Agent deleted'); await load() }
  catch (error) { notifications.error('Agent could not be deleted', error instanceof Error ? error.message : undefined) }
}

watch([query, status], (_value, _old, onCleanup) => { const current = ++revision; if (page.value !== 1) { page.value = 1; return }; const timeout = window.setTimeout(() => void load(current), 300); onCleanup(() => window.clearTimeout(timeout)) })
watch(page, () => void load())
onMounted(() => void load())
</script>

<template>
  <FilterCard title="Filters" subtitle="Search platform agent definitions" class="mb-4"><v-text-field v-model="query" hide-details placeholder="Search agents…" prepend-inner-icon="mdi-magnify" /><IconSelect v-model="status" hide-details :items="[{ title: 'All statuses', value: 'all' }, { title: 'Available', value: 'active' }, { title: 'Unavailable', value: 'inactive' }]" /></FilterCard>
  <SectionCard title="Agent definitions" :subtitle="`${total} definitions`"><template #actions><v-btn color="primary" prepend-icon="mdi-plus" @click="openCreate">Create agent</v-btn></template><div class="table-scroll" :aria-busy="loading"><v-progress-linear v-if="loading" indeterminate /><table class="data-table"><thead><tr><th>AGENT</th><th>RUNTIME</th><th>MODEL</th><th>RESOURCES</th><th>STATUS</th><th aria-label="Actions"></th></tr></thead><tbody><DataTableEmptyRow v-if="!loading && items.length === 0" :colspan="6" title="No agents found" description="Create an agent or adjust the filters." icon="mdi-robot-off-outline" /><tr v-for="agent in items" :key="agent.id"><td><div class="entity-cell"><div class="entity-icon"><v-icon icon="mdi-robot-outline" /></div><div><div class="entity-name">{{ agent.name }}</div><div class="entity-meta">{{ agent.description }}</div></div></div></td><td>{{ agent.runtime_adapter }} {{ agent.runtime_version }}</td><td>{{ agent.provider }} · {{ agent.model }}</td><td>{{ agent.cpu_millis }}m · {{ agent.memory_mb }} MB</td><td><StatusChip :label="agent.active ? 'Available' : 'Unavailable'" :color="agent.active ? 'success' : 'warning'" /></td><td class="text-right"><v-btn icon="mdi-pencil-outline" variant="text" size="small" aria-label="Edit agent" @click="openEdit(agent)" /><v-btn icon="mdi-delete-outline" variant="text" color="error" size="small" aria-label="Delete agent" @click="requestDelete(agent)" /></td></tr></tbody></table></div></SectionCard>
  <TablePaginationCard v-model="page" :total="total" :items-per-page="itemsPerPage" item-label="agents" />
  <AgentFormDialog v-model="editorOpen" :agent="selected" :saving="saving" @submit="save" />
  <v-dialog :model-value="Boolean(deleteTarget)" max-width="520" @update:model-value="!$event && (deleteTarget = null)"><v-card v-if="deleteTarget" class="section-card"><div class="pa-7"><h3 class="mb-3">Delete {{ deleteTarget.name }}?</h3><p class="text-body-2 text-medium-emphasis mb-5">This is permanent and will be rejected while a repository still uses this agent. Type <strong>{{ deleteTarget.name }}</strong> to confirm.</p><v-text-field v-model="deleteConfirmation" label="Agent name" autofocus /></div><v-card-actions class="dialog-actions"><v-spacer /><v-btn @click="deleteTarget = null">Cancel</v-btn><v-btn color="error" :disabled="deleteConfirmation !== deleteTarget.name" @click="confirmDelete">Delete permanently</v-btn></v-card-actions></v-card></v-dialog>
</template>
