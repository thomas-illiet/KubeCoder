<script setup lang="ts">
import { onMounted, shallowRef, watch } from 'vue'
import { useRouter } from 'vue-router'
import type { AdminOrganization } from '../../api/organizations'
import OrganizationFormDialog from '../../components/organizations/OrganizationFormDialog.vue'
import DataTableEmptyRow from '../../components/DataTableEmptyRow.vue'
import FilterCard from '../../components/FilterCard.vue'
import SectionCard from '../../components/SectionCard.vue'
import TablePaginationCard from '../../components/TablePaginationCard.vue'
import { useNotifications } from '../../composables/useNotifications'
import { useOrganizations } from '../../composables/useOrganizations'

const router = useRouter()
const service = useOrganizations()
const notifications = useNotifications()
const items = shallowRef<AdminOrganization[]>([])
const query = shallowRef('')
const loading = shallowRef(false)
const formOpen = shallowRef(false)
const page = shallowRef(1)
const total = shallowRef(0)
const repositoryTotal = shallowRef(0)
const itemsPerPage = 20
type SortKey = 'name' | 'created_at' | 'member_count' | 'repository_count'
type SortDirection = 'asc' | 'desc'
const sortBy = shallowRef<SortKey>('name')
const sortDirection = shallowRef<SortDirection>('asc')
let loadRevision = 0

// loadOrganizations requests the current bounded page from the server.
async function loadOrganizations(revision = ++loadRevision): Promise<void> {
  loading.value = true
  try {
    const result = await service.listAdmin({
      query: query.value.trim(),
      limit: itemsPerPage,
      offset: (page.value - 1) * itemsPerPage,
      orderBy: sortBy.value,
      orderDirection: sortDirection.value,
    })
    if (revision !== loadRevision) return
    items.value = result.items ?? []
    total.value = result.total
    repositoryTotal.value = result.repository_total
  } catch (error) {
    if (revision === loadRevision) notifications.error('Organizations could not be loaded', error instanceof Error ? error.message : undefined)
  } finally {
    if (revision === loadRevision) loading.value = false
  }
}

// openCreate resets and opens the create form.
function openCreate(): void {
  formOpen.value = true
}

// saveOrganization creates an organization.
async function saveOrganization(input: { name: string; slug: string }): Promise<void> {
  try {
    await service.create(input)
    formOpen.value = false
    notifications.success('Organization created')
    await loadOrganizations()
  } catch (error) {
    notifications.error('Organization could not be saved', error instanceof Error ? error.message : undefined)
  }
}

// formatCreatedAt formats an API timestamp in the user's locale.
function formatCreatedAt(value: string): string {
  return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' }).format(new Date(value))
}

// changeSort toggles a server-side column order and returns to the first page.
function changeSort(column: SortKey): void {
  sortDirection.value = sortBy.value === column && sortDirection.value === 'asc' ? 'desc' : 'asc'
  sortBy.value = column
  if (page.value !== 1) page.value = 1
  else void loadOrganizations()
}

function ariaSort(column: SortKey): 'ascending' | 'descending' | 'none' {
  if (sortBy.value !== column) return 'none'
  return sortDirection.value === 'asc' ? 'ascending' : 'descending'
}

watch(query, (_value, _previous, onCleanup) => {
  const revision = ++loadRevision
  loading.value = true
  if (page.value !== 1) {
    page.value = 1
    return
  }
  const timeout = window.setTimeout(() => void loadOrganizations(revision), 300)
  onCleanup(() => window.clearTimeout(timeout))
})

watch(page, () => {
  void loadOrganizations()
})

onMounted(() => void loadOrganizations())
</script>

<template>
  <FilterCard title="Organizations" subtitle="Search and administer platform tenants" class="mb-4">
    <v-text-field v-model="query" hide-details placeholder="Search organizations…" prepend-inner-icon="mdi-magnify" />
    <v-spacer />
    <v-btn color="success" variant="flat" prepend-icon="mdi-domain-plus" @click="openCreate">Create organization</v-btn>
  </FilterCard>
  <SectionCard title="Platform organizations" :subtitle="`${total} organizations · ${repositoryTotal} repositories`">
    <div class="organization-table-shell" :aria-busy="loading">
      <v-progress-linear v-if="loading" class="organization-table-progress" indeterminate color="primary" />
      <div class="table-scroll organization-table-content" :class="{ 'organization-table-content--loading': loading }">
      <table class="data-table">
        <thead>
          <tr>
            <th :aria-sort="ariaSort('name')"><button class="sort-header" type="button" @click="changeSort('name')">ORGANIZATION<v-icon :icon="sortBy === 'name' ? (sortDirection === 'asc' ? 'mdi-arrow-up' : 'mdi-arrow-down') : 'mdi-unfold-more-horizontal'" size="16" /></button></th>
            <th :aria-sort="ariaSort('created_at')"><button class="sort-header" type="button" @click="changeSort('created_at')">CREATED<v-icon :icon="sortBy === 'created_at' ? (sortDirection === 'asc' ? 'mdi-arrow-up' : 'mdi-arrow-down') : 'mdi-unfold-more-horizontal'" size="16" /></button></th>
            <th :aria-sort="ariaSort('member_count')"><button class="sort-header" type="button" @click="changeSort('member_count')">MEMBERS<v-icon :icon="sortBy === 'member_count' ? (sortDirection === 'asc' ? 'mdi-arrow-up' : 'mdi-arrow-down') : 'mdi-unfold-more-horizontal'" size="16" /></button></th>
            <th :aria-sort="ariaSort('repository_count')"><button class="sort-header" type="button" @click="changeSort('repository_count')">REPOSITORIES<v-icon :icon="sortBy === 'repository_count' ? (sortDirection === 'asc' ? 'mdi-arrow-up' : 'mdi-arrow-down') : 'mdi-unfold-more-horizontal'" size="16" /></button></th>
            <th><span class="sr-only">Open</span></th>
          </tr>
        </thead>
        <tbody>
          <DataTableEmptyRow
            v-if="!loading && items.length === 0"
            :colspan="5"
            title="No organizations found"
            description="Create an organization or adjust your search to display results."
            icon="mdi-domain-off"
          />
          <tr v-for="organization in items" :key="organization.id" class="organization-row" tabindex="0" @click="router.push(`/admin/organizations/${organization.id}`)" @keydown.enter="router.push(`/admin/organizations/${organization.id}`)">
            <td><div class="d-flex align-center py-3"><div class="org-avatar mr-3">{{ organization.name.slice(0, 2).toUpperCase() }}</div><div><div class="font-weight-medium">{{ organization.name }}</div><div class="text-caption text-medium-emphasis">{{ organization.slug }}</div></div></div></td>
            <td>{{ formatCreatedAt(organization.created_at) }}</td>
            <td>{{ organization.member_count }}</td>
            <td>{{ organization.repository_count }}</td>
            <td class="text-right"><v-icon icon="mdi-chevron-right" aria-hidden="true" /></td>
          </tr>
        </tbody>
      </table>
      </div>
    </div>
  </SectionCard>
  <TablePaginationCard v-model="page" :total="total" :items-per-page="itemsPerPage" item-label="organizations" />

  <OrganizationFormDialog v-model="formOpen" @submit="saveOrganization" />
</template>

<style scoped>
.organization-row { cursor: pointer; }
.organization-row:focus-visible { outline: 2px solid rgb(var(--v-theme-primary)); outline-offset: -2px; }
.organization-table-shell { position: relative; overflow: hidden; }
.organization-table-progress { position: absolute; z-index: 2; top: 0; right: 0; left: 0; }
.organization-table-content { transition: opacity .2s ease, filter .2s ease; }
.organization-table-content--loading { pointer-events: none; opacity: .42; filter: saturate(.7); }
.sort-header { display: inline-flex; align-items: center; gap: .35rem; padding: 0; border: 0; background: transparent; color: inherit; font: inherit; letter-spacing: inherit; cursor: pointer; }
.sort-header:focus-visible { outline: 2px solid rgb(var(--v-theme-primary)); outline-offset: 3px; border-radius: 2px; }
.sr-only { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip: rect(0, 0, 0, 0); white-space: nowrap; border: 0; }
</style>
