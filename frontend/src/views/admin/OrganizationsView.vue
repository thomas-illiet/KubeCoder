<script setup lang="ts">
import { computed, onMounted, shallowRef } from 'vue'
import { useRouter } from 'vue-router'
import type { AdminOrganization } from '../../api/organizations'
import OrganizationFormDialog from '../../components/organizations/OrganizationFormDialog.vue'
import DataTableEmptyRow from '../../components/DataTableEmptyRow.vue'
import FilterCard from '../../components/FilterCard.vue'
import SectionCard from '../../components/SectionCard.vue'
import { useNotifications } from '../../composables/useNotifications'
import { useOrganizations } from '../../composables/useOrganizations'

const router = useRouter()
const service = useOrganizations()
const notifications = useNotifications()
const items = shallowRef<AdminOrganization[]>([])
const query = shallowRef('')
const loading = shallowRef(false)
const formOpen = shallowRef(false)
const visibleItems = computed(() => {
  const term = query.value.trim().toLowerCase()
  return term ? items.value.filter((item) => `${item.name} ${item.slug}`.toLowerCase().includes(term)) : items.value
})

// loadOrganizations refreshes the platform organization list.
async function loadOrganizations(): Promise<void> {
  loading.value = true
  try {
    items.value = (await service.listAdmin()).items ?? []
  } catch (error) {
    notifications.error('Organizations could not be loaded', error instanceof Error ? error.message : undefined)
  } finally {
    loading.value = false
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

onMounted(loadOrganizations)
</script>

<template>
  <FilterCard title="Organizations" subtitle="Search and administer platform tenants" class="mb-4">
    <v-text-field v-model="query" hide-details placeholder="Search organizations…" prepend-inner-icon="mdi-magnify" />
    <v-spacer />
    <v-btn color="success" variant="flat" prepend-icon="mdi-domain-plus" @click="openCreate">Create organization</v-btn>
  </FilterCard>
  <SectionCard title="Platform organizations" :subtitle="`${visibleItems.length} organizations`">
    <v-progress-linear v-if="loading" indeterminate />
    <div v-else class="table-scroll">
      <table class="data-table">
        <thead><tr><th>ORGANIZATION</th><th>CREATED</th><th>MEMBERS</th><th><span class="sr-only">Open</span></th></tr></thead>
        <tbody>
          <DataTableEmptyRow
            v-if="visibleItems.length === 0"
            :colspan="4"
            title="No organizations found"
            description="Create an organization or adjust your search to display results."
            icon="mdi-domain-off"
          />
          <tr v-for="organization in visibleItems" :key="organization.id" class="organization-row" tabindex="0" @click="router.push(`/admin/organizations/${organization.id}`)" @keydown.enter="router.push(`/admin/organizations/${organization.id}`)">
            <td><div class="d-flex align-center py-3"><div class="org-avatar mr-3">{{ organization.name.slice(0, 2).toUpperCase() }}</div><div><div class="font-weight-medium">{{ organization.name }}</div><div class="text-caption text-medium-emphasis">{{ organization.slug }}</div></div></div></td>
            <td>{{ formatCreatedAt(organization.created_at) }}</td>
            <td>{{ organization.member_count }}</td>
            <td class="text-right"><v-icon icon="mdi-chevron-right" aria-hidden="true" /></td>
          </tr>
        </tbody>
      </table>
    </div>
  </SectionCard>

  <OrganizationFormDialog v-model="formOpen" @submit="saveOrganization" />
</template>

<style scoped>
.organization-row { cursor: pointer; }
.organization-row:focus-visible { outline: 2px solid rgb(var(--v-theme-primary)); outline-offset: -2px; }
.sr-only { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip: rect(0, 0, 0, 0); white-space: nowrap; border: 0; }
</style>
