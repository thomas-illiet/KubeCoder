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
const scopeFilter = ref('All scopes')
const statusFilter = ref('All statuses')
const createOpen = ref(false)
const rotateOpen = ref(false)
const historyOpen = ref(false)
const selectedName = ref('')
const { success, warning, info } = useNotifications()

const secrets = [
  { name: 'MODEL_API_TOKEN', description: 'Model provider access', scope: 'Organization', bindings: 3, rotated: 'Sep 12, 2026', expires: 'Dec 12, 2026', status: 'Active' },
  { name: 'S3_ARTIFACTS_KEY', description: 'Write access for run artifacts', scope: 'Organization', bindings: 1, rotated: 'Sep 4, 2026', expires: 'No expiration', status: 'Active' },
  { name: 'NPM_REGISTRY_TOKEN', description: 'Internal package registry', scope: 'Repository', bindings: 2, rotated: 'Aug 18, 2026', expires: 'Oct 3, 2026', status: 'Expiring soon' },
  { name: 'SIGNING_KEY', description: 'Release image signing', scope: 'Repository', bindings: 1, rotated: 'Jul 2, 2026', expires: 'Sep 20, 2026', status: 'Expired' },
]

const filteredSecrets = computed(() => secrets.filter((item) => {
  const matchesSearch = `${item.name} ${item.description}`.toLowerCase().includes(query.value.trim().toLowerCase())
  const matchesScope = scopeFilter.value === 'All scopes' || item.scope === scopeFilter.value
  const matchesStatus = statusFilter.value === 'All statuses' || item.status === statusFilter.value
  return matchesSearch && matchesScope && matchesStatus
}))
const { page: secretsPage, paginatedItems: paginatedSecrets, itemsPerPage: secretsPerPage } = usePagination(filteredSecrets)

function rotate(name: string) { selectedName.value = name; rotateOpen.value = true }
function createSecret() { createOpen.value = false; success('Secret created', 'The protected value is now available to authorized runs.') }
function rotateSecret() { rotateOpen.value = false; success('Secret rotated', `${selectedName.value} now uses a new version.`) }
function revokeSecret(name: string) { warning('Secret revoked', `${name} can no longer be resolved by new runs.`) }
function viewUsage(name: string, bindings: number) { info('Secret usage', `${name} is referenced by ${bindings} active binding${bindings === 1 ? '' : 's'}.`) }
</script>

<template>
  <div class="security-note mb-4"><v-icon icon="mdi-shield-lock-outline" size="20" color="secondary" /><span><strong class="text-high-emphasis">Write-only values.</strong> A secret value is visible only while it is being entered and is never returned by the interface.</span></div>
  <FilterCard title="Filters" subtitle="Search by name, scope, or status" class="secrets-filter-card mb-4">
    <v-text-field v-model="query" hide-details placeholder="Search secrets…" prepend-inner-icon="mdi-magnify" />
    <IconSelect v-model="scopeFilter" hide-details :items="['All scopes', 'Organization', 'Repository']" min-width="180" />
    <IconSelect v-model="statusFilter" hide-details :items="['All statuses', 'Active', 'Expiring soon', 'Expired']" min-width="170" />
    <v-spacer />
    <v-btn variant="text" prepend-icon="mdi-history" @click="historyOpen = true">Rotation history</v-btn>
  </FilterCard>

  <SectionCard title="Secrets" subtitle="Metadata and protected references">
    <template #actions><v-btn color="primary" prepend-icon="mdi-plus" @click="createOpen = true">New secret</v-btn></template>
    <div class="table-scroll">
      <table class="data-table">
        <thead><tr><th>SECRET</th><th class="table-cell--center">SCOPE</th><th class="table-cell--center">BINDINGS</th><th class="table-cell--center">LAST ROTATION</th><th class="table-cell--center">EXPIRATION</th><th class="table-cell--center">STATUS</th><th class="table-cell--center" aria-label="Actions"></th></tr></thead>
        <tbody><DataTableEmptyRow v-if="filteredSecrets.length === 0" :colspan="7" /><tr v-for="secret in paginatedSecrets" :key="secret.name">
          <td><div class="entity-cell"><div class="entity-icon"><v-icon icon="mdi-key-variant" size="19" /></div><div><div class="entity-name code-text">{{ secret.name }}</div><div class="entity-meta">{{ secret.description }}</div></div></div></td>
          <td class="table-cell--center"><v-chip size="small" variant="outlined">{{ secret.scope }}</v-chip></td>
          <td class="table-cell--center">{{ secret.bindings }} usage{{ secret.bindings > 1 ? 's' : '' }}</td>
          <td class="text-medium-emphasis table-cell--center">{{ secret.rotated }}</td>
          <td class="text-medium-emphasis table-cell--center">{{ secret.expires }}</td>
          <td class="table-cell--center"><StatusChip :label="secret.status" :color="secret.status === 'Active' ? 'success' : secret.status === 'Expired' ? 'error' : 'warning'" :icon="secret.status === 'Active' ? 'mdi-check-circle-outline' : secret.status === 'Expired' ? 'mdi-close-circle-outline' : 'mdi-clock-alert-outline'" /></td>
          <td class="table-cell--center"><v-menu><template #activator="{ props }"><v-btn v-bind="props" icon="mdi-dots-horizontal" variant="text" size="small" :aria-label="`Actions for ${secret.name}`" /></template><v-list class="pa-2" width="210"><v-list-item title="Rotate" prepend-icon="mdi-refresh" @click="rotate(secret.name)" /><v-list-item title="View usage" prepend-icon="mdi-link-variant" @click="viewUsage(secret.name, secret.bindings)" /><v-divider class="my-2" /><v-list-item title="Revoke" prepend-icon="mdi-cancel" class="text-error" @click="revokeSecret(secret.name)" /></v-list></v-menu></td>
        </tr></tbody>
      </table>
    </div>
  </SectionCard>
  <TablePaginationCard v-model="secretsPage" :total="filteredSecrets.length" :items-per-page="secretsPerPage" item-label="secrets" />

  <v-dialog v-model="createOpen" max-width="560">
    <v-card class="section-card">
      <div class="pa-6"><div class="d-flex align-center justify-space-between mb-5"><div><h3>Create secret</h3><div class="text-caption text-medium-emphasis">The value will be protected after confirmation</div></div><v-btn icon="mdi-close" variant="text" aria-label="Close" @click="createOpen = false" /></div>
        <v-text-field label="Logical name" placeholder="EXAMPLE_TOKEN" />
        <v-textarea label="Description" rows="2" />
        <div class="form-row"><IconSelect label="Scope" :items="['Organization', 'Repository']" model-value="Organization" /><v-text-field label="Target variable" placeholder="EXAMPLE_TOKEN" /></div>
        <v-text-field label="Expiration date" type="date" hint="Optional; an expired secret cannot be resolved for a new run" persistent-hint />
        <v-text-field label="Secret value" type="password" append-inner-icon="mdi-eye-off-outline" autocomplete="new-password" />
        <div class="security-note"><v-icon icon="mdi-information-outline" color="info" /><span>After creation, only the name, scope, fingerprint, and rotation date will remain visible.</span></div>
      </div>
      <v-card-actions class="dialog-actions"><v-spacer /><v-btn @click="createOpen = false">Cancel</v-btn><v-btn color="primary" @click="createSecret">Create secret</v-btn></v-card-actions>
    </v-card>
  </v-dialog>

  <v-dialog v-model="rotateOpen" max-width="500">
    <v-card class="section-card"><div class="pa-6"><h3 class="mb-1">Rotate {{ selectedName }}</h3><p class="text-body-2 text-medium-emphasis mb-5">Runs that have already started will keep their current version.</p><v-text-field label="New value" type="password" autocomplete="new-password" /><v-text-field label="Confirm value" type="password" autocomplete="new-password" /></div><v-card-actions class="dialog-actions"><v-spacer /><v-btn @click="rotateOpen = false">Cancel</v-btn><v-btn color="primary" @click="rotateSecret">Create version</v-btn></v-card-actions></v-card>
  </v-dialog>

  <v-dialog v-model="historyOpen" max-width="620">
    <v-card class="section-card">
      <div class="dialog-header"><div><h3>Rotation history</h3><div class="text-caption text-medium-emphasis">Most recent version change for each secret</div></div><v-btn icon="mdi-close" variant="text" aria-label="Close" @click="historyOpen = false" /></div>
      <v-list bg-color="transparent" class="pa-3">
        <v-list-item v-for="secret in secrets" :key="secret.name" :title="secret.name" :subtitle="`Rotated ${secret.rotated} · ${secret.bindings} binding${secret.bindings === 1 ? '' : 's'}`" prepend-icon="mdi-history" />
      </v-list>
      <v-card-actions class="dialog-actions"><v-spacer /><v-btn @click="historyOpen = false">Close</v-btn></v-card-actions>
    </v-card>
  </v-dialog>
</template>
