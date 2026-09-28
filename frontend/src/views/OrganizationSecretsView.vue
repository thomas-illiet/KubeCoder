<script setup lang="ts">
import { computed, ref } from 'vue'
import DataTableEmptyRow from '../components/DataTableEmptyRow.vue'
import FilterCard from '../components/FilterCard.vue'
import SectionCard from '../components/SectionCard.vue'
import StatusChip from '../components/StatusChip.vue'
import TablePaginationCard from '../components/TablePaginationCard.vue'
import { useNotifications } from '../composables/useNotifications'
import { usePagination } from '../composables/usePagination'

type Secret = {
  name: string
  description: string
  scope: 'Organization' | 'Repository'
  target: string
  bindings: number
  rotated: string
  expires: string
  status: 'Active' | 'Expiring soon' | 'Expired'
}

const query = ref('')
const scope = ref('All scopes')
const status = ref('All statuses')
const createOpen = ref(false)
const newScope = ref<'Organization' | 'Repository'>('Organization')
const rotateOpen = ref(false)
const detailsOpen = ref(false)
const selected = ref<Secret | null>(null)
const { success } = useNotifications()

const secrets: Secret[] = [
  { name: 'MODEL_API_TOKEN', description: 'Access to the organization model provider', scope: 'Organization', target: 'MODEL_API_TOKEN', bindings: 3, rotated: 'Sep 12, 2026', expires: 'Dec 12, 2026', status: 'Active' },
  { name: 'S3_ARTIFACTS_KEY', description: 'Write access for run artifacts', scope: 'Organization', target: 'S3_ARTIFACTS_KEY', bindings: 1, rotated: 'Sep 4, 2026', expires: 'No expiration', status: 'Active' },
  { name: 'NPM_REGISTRY_TOKEN', description: 'Package registry for checkout-api', scope: 'Repository', target: 'NPM_TOKEN', bindings: 2, rotated: 'Aug 18, 2026', expires: 'Oct 3, 2026', status: 'Expiring soon' },
  { name: 'LEGACY_SIGNING_KEY', description: 'Legacy key retained for audit purposes', scope: 'Repository', target: 'SIGNING_KEY', bindings: 0, rotated: 'Jul 2, 2026', expires: 'Sep 20, 2026', status: 'Expired' },
]

const filtered = computed(() => secrets.filter((secret) => {
  const matchesQuery = `${secret.name} ${secret.description} ${secret.target}`.toLocaleLowerCase('en').includes(query.value.toLocaleLowerCase('en'))
  const matchesScope = scope.value === 'All scopes' || secret.scope === scope.value
  const matchesStatus = status.value === 'All statuses' || secret.status === status.value
  return matchesQuery && matchesScope && matchesStatus
}))
const { page, paginatedItems: paginatedSecrets, itemsPerPage } = usePagination(filtered)

function openDetails(secret: Secret) {
  selected.value = secret
  detailsOpen.value = true
}

function openRotation(secret: Secret) {
  selected.value = secret
  rotateOpen.value = true
}

function createSecret() { createOpen.value = false; success('Secret created', 'The protected value is available to authorized organization runs.') }
function rotateSecret() { rotateOpen.value = false; success('Secret rotated', `${selected.value?.name ?? 'The secret'} now uses a new version.`) }
</script>

<template>
  <div class="security-note mb-4"><v-icon icon="mdi-shield-lock-outline" size="20" color="secondary" /><span><strong class="text-high-emphasis">Write-only values.</strong> A value is visible only while it is being entered. It is never returned by the API or displayed again on this page. Global platform secrets are never visible from an organization.</span></div>

  <FilterCard title="Filters" subtitle="Search Northstar Labs secrets" class="mb-4">
    <v-text-field v-model="query" hide-details placeholder="Search secrets…" prepend-inner-icon="mdi-magnify" />
    <IconSelect v-model="scope" hide-details :items="['All scopes', 'Organization', 'Repository']" min-width="190" />
    <IconSelect v-model="status" hide-details :items="['All statuses', 'Active', 'Expiring soon', 'Expired']" min-width="180" />
  </FilterCard>

  <SectionCard title="Organization secrets" subtitle="Metadata, expiration dates, and explicit bindings">
    <template #actions><v-btn color="primary" prepend-icon="mdi-plus" @click="createOpen = true">New secret</v-btn></template>
    <div class="table-scroll">
      <table class="data-table">
        <thead><tr><th>SECRET</th><th class="table-cell--center">SCOPE</th><th>TARGET VARIABLE</th><th class="table-cell--center">BINDINGS</th><th class="table-cell--center">LAST ROTATION</th><th class="table-cell--center">EXPIRATION</th><th class="table-cell--center">STATUS</th><th class="table-cell--center">ACTIONS</th></tr></thead>
        <tbody>
          <DataTableEmptyRow v-if="filtered.length === 0" :colspan="8" />
          <tr v-for="secret in paginatedSecrets" :key="secret.name">
            <td><div class="entity-cell"><div class="entity-icon"><v-icon icon="mdi-key-variant" size="19" /></div><div><div class="entity-name code-text">{{ secret.name }}</div><div class="entity-meta">{{ secret.description }}</div></div></div></td>
            <td class="table-cell--center"><v-chip size="small" variant="outlined">{{ secret.scope }}</v-chip></td>
            <td><span class="code-text text-medium-emphasis">{{ secret.target }}</span></td>
            <td class="table-cell--center">{{ secret.bindings }} binding{{ secret.bindings > 1 ? 's' : '' }}</td>
            <td class="text-medium-emphasis table-cell--center">{{ secret.rotated }}</td>
            <td class="text-medium-emphasis table-cell--center">{{ secret.expires }}</td>
            <td class="table-cell--center"><StatusChip :label="secret.status" :color="secret.status === 'Active' ? 'success' : secret.status === 'Expired' ? 'error' : 'warning'" :icon="secret.status === 'Active' ? 'mdi-check-circle-outline' : secret.status === 'Expired' ? 'mdi-close-circle-outline' : 'mdi-clock-alert-outline'" /></td>
            <td class="table-cell--center"><div class="d-flex ga-1"><v-btn variant="text" size="small" prepend-icon="mdi-information-outline" @click="openDetails(secret)">Details</v-btn><v-btn variant="text" size="small" prepend-icon="mdi-refresh" @click="openRotation(secret)">Rotate</v-btn></div></td>
          </tr>
        </tbody>
      </table>
    </div>
  </SectionCard>
  <TablePaginationCard v-model="page" :total="filtered.length" :items-per-page="itemsPerPage" item-label="secrets" hint="OWNER and ADMIN only" />

  <v-dialog v-model="createOpen" max-width="620">
    <v-card class="app-dialog">
      <div class="dialog-header"><div><div class="text-h6 font-weight-bold">Create secret</div><div class="text-caption text-medium-emphasis">The value will be protected immediately after confirmation</div></div><v-btn icon="mdi-close" variant="text" aria-label="Close" @click="createOpen = false" /></div>
      <div class="pa-5">
        <div class="form-row"><v-text-field label="Logical name" placeholder="EXAMPLE_TOKEN" hint="Uppercase letters, digits, and underscores" persistent-hint /><v-text-field label="Target variable" placeholder="EXAMPLE_TOKEN" /></div>
        <v-textarea label="Description" rows="2" />
        <div class="form-row"><IconSelect v-model="newScope" label="Scope" :items="['Organization', 'Repository']" /><IconSelect label="Repository" :items="['identity-service', 'platform-k8s', 'checkout-api']" :disabled="newScope !== 'Repository'" hint="Required only for Repository scope" persistent-hint /></div>
        <v-text-field label="Expiration date" type="date" hint="Optional; an expired secret blocks any new run that attempts to resolve it" persistent-hint />
        <v-text-field label="Secret value" type="password" placeholder="Enter the value" prepend-inner-icon="mdi-key-variant" autocomplete="new-password" />
        <div class="security-note"><v-icon icon="mdi-eye-off-outline" color="secondary" /><span>After creation, only the name, scope, non-reversible fingerprint, and metadata will remain visible.</span></div>
      </div>
      <v-card-actions class="dialog-actions"><v-spacer /><v-btn variant="text" @click="createOpen = false">Cancel</v-btn><v-btn color="primary" @click="createSecret">Create secret</v-btn></v-card-actions>
    </v-card>
  </v-dialog>

  <v-dialog v-model="rotateOpen" max-width="540">
    <v-card class="app-dialog">
      <div class="dialog-header"><div><div class="text-h6 font-weight-bold">Rotate secret</div><div class="text-caption text-medium-emphasis code-text">{{ selected?.name }}</div></div><v-btn icon="mdi-close" variant="text" aria-label="Close" @click="rotateOpen = false" /></div>
      <div class="pa-5"><v-text-field label="New value" type="password" autocomplete="new-password" prepend-inner-icon="mdi-key-variant" /><v-text-field label="New expiration date" type="date" /><div class="security-note"><v-icon icon="mdi-history" color="info" /><span>Runs that have already started keep their current version. Future bindings will use the new version.</span></div></div>
      <v-card-actions class="dialog-actions"><v-spacer /><v-btn variant="text" @click="rotateOpen = false">Cancel</v-btn><v-btn color="primary" @click="rotateSecret">Create version</v-btn></v-card-actions>
    </v-card>
  </v-dialog>

  <v-navigation-drawer v-model="detailsOpen" location="right" temporary width="480" class="detail-drawer">
    <div class="drawer-header d-flex align-center justify-space-between"><div><div class="text-h6 font-weight-bold">{{ selected?.name }}</div><div class="text-caption text-medium-emphasis">Secret metadata</div></div><v-btn icon="mdi-close" variant="text" aria-label="Close" @click="detailsOpen = false" /></div>
    <div class="form-section"><div class="form-section__title">Status and scope</div><div class="d-flex justify-space-between align-center mb-4"><StatusChip :label="selected?.status ?? ''" :color="selected?.status === 'Active' ? 'success' : selected?.status === 'Expired' ? 'error' : 'warning'" /><v-chip size="small" variant="outlined">{{ selected?.scope }}</v-chip></div><div class="d-flex justify-space-between py-2"><span class="text-medium-emphasis">Variable</span><code>{{ selected?.target }}</code></div><div class="d-flex justify-space-between py-2"><span class="text-medium-emphasis">Expiration</span><span>{{ selected?.expires }}</span></div><div class="d-flex justify-space-between py-2"><span class="text-medium-emphasis">Bindings</span><strong>{{ selected?.bindings }}</strong></div></div>
    <div class="form-section"><div class="security-note"><v-icon icon="mdi-eye-off-outline" color="secondary" /><span>The value and internal vault reference are never exposed in this panel.</span></div></div>
  </v-navigation-drawer>
</template>
