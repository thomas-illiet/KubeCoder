<script setup lang="ts">
import { onBeforeUnmount, shallowRef, watch } from 'vue'
import { useRoute } from 'vue-router'
import type { GitProvider, Repository, RepositoryInput, RepositoryPageRequest } from '../api/repositories'
import FilterCard from '../components/FilterCard.vue'
import RepositoryDeleteDialog from '../components/repositories/RepositoryDeleteDialog.vue'
import RepositoryFormDialog from '../components/repositories/RepositoryFormDialog.vue'
import RepositoryTable from '../components/repositories/RepositoryTable.vue'
import SectionCard from '../components/SectionCard.vue'
import TablePaginationCard from '../components/TablePaginationCard.vue'
import { useNotifications } from '../composables/useNotifications'
import { useRepositoryCatalog } from '../composables/useRepositoryCatalog'

const route = useRoute()
const catalog = useRepositoryCatalog()
const notifications = useNotifications()
const query = shallowRef('')
const provider = shallowRef<'all' | GitProvider>('all')
const configured = shallowRef<'all' | 'configured' | 'unconfigured'>('all')
const agentID = shallowRef<string | null>(null)
const page = shallowRef(1)
const itemsPerPage = 20
const editorOpen = shallowRef(false)
const selected = shallowRef<Repository | null>(null)
const deleteTarget = shallowRef<Repository | null>(null)
const saving = shallowRef(false)
const deleting = shallowRef(false)
let loadTimer: number | undefined

function slug(): string { return typeof route.params.organizationSlug === 'string' ? route.params.organizationSlug : '' }
function request(): RepositoryPageRequest {
  return { query: query.value.trim(), provider: provider.value === 'all' ? undefined : provider.value, configured: configured.value === 'all' ? undefined : configured.value === 'configured', agentID: agentID.value ?? undefined, limit: itemsPerPage, offset: (page.value - 1) * itemsPerPage }
}
async function load(): Promise<void> {
  if (!slug()) return
  try { await catalog.load(slug(), request()) }
  catch (error) { notifications.error('Repositories could not be loaded', error instanceof Error ? error.message : undefined) }
}
async function initialize(): Promise<void> {
  catalog.reset()
  if (!slug()) return
  try { await Promise.all([catalog.loadAgents(slug()), catalog.load(slug(), request())]) }
  catch (error) { notifications.error('Repository workspace could not be loaded', error instanceof Error ? error.message : undefined) }
}
function openCreate(): void { selected.value = null; editorOpen.value = true }
function openEdit(repository: Repository): void { selected.value = repository; editorOpen.value = true }
async function save(input: RepositoryInput): Promise<void> {
  saving.value = true
  try { await catalog.save(slug(), input, selected.value?.id); editorOpen.value = false; notifications.success(selected.value ? 'Repository updated' : 'Repository added'); await load() }
  catch (error) { notifications.error('Repository could not be saved', error instanceof Error ? error.message : undefined) }
  finally { saving.value = false }
}
async function remove(): Promise<void> {
  if (!deleteTarget.value) return
  deleting.value = true
  try { await catalog.remove(slug(), deleteTarget.value.id); deleteTarget.value = null; notifications.success('Repository deleted'); if (catalog.items.value.length === 1 && page.value > 1) page.value -= 1; else await load() }
  catch (error) { notifications.error('Repository could not be deleted', error instanceof Error ? error.message : undefined) }
  finally { deleting.value = false }
}

watch(() => route.params.organizationSlug, () => { page.value = 1; query.value = ''; provider.value = 'all'; configured.value = 'all'; agentID.value = null; void initialize() }, { immediate: true })
watch([query, provider, configured, agentID], () => { window.clearTimeout(loadTimer); if (page.value !== 1) { page.value = 1; return }; loadTimer = window.setTimeout(() => void load(), 300) })
watch(page, () => void load())
onBeforeUnmount(() => { window.clearTimeout(loadTimer); catalog.reset() })
</script>

<template>
  <FilterCard title="Filters" subtitle="Search repositories in the current organization" class="mb-4"><v-text-field v-model="query" hide-details placeholder="Search repositories…" prepend-inner-icon="mdi-magnify" /><IconSelect v-model="provider" hide-details :items="[{ title: 'All providers', value: 'all' }, { title: 'GitHub', value: 'github' }, { title: 'GitLab', value: 'gitlab' }, { title: 'Bitbucket', value: 'bitbucket' }]" /><IconSelect v-model="configured" hide-details :items="[{ title: 'All statuses', value: 'all' }, { title: 'Configured', value: 'configured' }, { title: 'Not configured', value: 'unconfigured' }]" /><IconSelect v-model="agentID" hide-details :items="[{ title: 'All agents', value: null }, ...catalog.agents.value.map((agent) => ({ title: agent.name, value: agent.id }))]" /></FilterCard>
  <SectionCard title="Repositories" :subtitle="`${catalog.total.value} repositories in this organization`"><template #actions><v-btn color="primary" prepend-icon="mdi-plus" @click="openCreate">Add repository</v-btn></template><RepositoryTable :items="catalog.items.value" :loading="catalog.loading.value" @edit="openEdit" @delete="deleteTarget = $event" /></SectionCard>
  <TablePaginationCard v-model="page" :total="catalog.total.value" :items-per-page="itemsPerPage" item-label="repositories" hint="Credentials remain backend-managed" />
  <RepositoryFormDialog v-model="editorOpen" :repository="selected" :agents="catalog.agents.value" :saving="saving" @submit="save" />
  <RepositoryDeleteDialog :repository="deleteTarget" :deleting="deleting" @close="deleteTarget = null" @confirm="remove" />
</template>
