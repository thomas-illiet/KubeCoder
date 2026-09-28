<script setup lang="ts">
import { computed, ref } from 'vue'
import DataTableEmptyRow from '../components/DataTableEmptyRow.vue'
import FilterCard from '../components/FilterCard.vue'
import SectionCard from '../components/SectionCard.vue'
import StatusChip from '../components/StatusChip.vue'
import TablePaginationCard from '../components/TablePaginationCard.vue'
import { useNotifications } from '../composables/useNotifications'
import { usePagination } from '../composables/usePagination'

const tab = ref('images')
const query = ref('')
const addOpen = ref(false)
const { success, info } = useNotifications()
function runChecks() { info('Supply-chain checks started', 'Images and adapters will be revalidated in the background.') }
function saveResource() { addOpen.value = false; success(tab.value === 'images' ? 'OCI image registered' : 'Adapter added', 'The resource was saved and queued for verification.') }
function showResourceDetails(name: string) { info('Resource details', `${name} is selected for inspection.`) }
const images = [
  { name: 'opencode-runtime', tag: '1.2.4', digest: 'sha256:4e8f…9a12', registry: 'registry.internal', scan: 'Verified', signed: true, used: '5 agents' },
  { name: 'opencode-runtime', tag: '1.3.0-rc.2', digest: 'sha256:b771…33cd', registry: 'registry.internal', scan: 'Warning', signed: true, used: '1 agent' },
  { name: 'workspace-base', tag: '2026.09', digest: 'sha256:02ac…7f18', registry: 'registry.internal', scan: 'Verified', signed: true, used: 'Global' },
]
const adapters = [
  { name: 'OpenCode Adapter', version: '1.2.0', protocol: 'RunEvent v1', resume: 'Native + reconstructed', status: 'Stable', tested: '4 hours ago' },
  { name: 'OpenCode Adapter', version: '1.3.0-rc.1', protocol: 'RunEvent v1', resume: 'Reconstructed', status: 'Candidate', tested: 'Yesterday' },
]
const filteredImages = computed(() => images.filter((image) => `${image.name} ${image.tag} ${image.digest} ${image.registry}`.toLowerCase().includes(query.value.trim().toLowerCase())))
const filteredAdapters = computed(() => adapters.filter((adapter) => `${adapter.name} ${adapter.version} ${adapter.protocol} ${adapter.status}`.toLowerCase().includes(query.value.trim().toLowerCase())))
const { page: imagesPage, paginatedItems: paginatedImages, itemsPerPage: imagesPerPage } = usePagination(filteredImages)
const { page: adaptersPage, paginatedItems: paginatedAdapters, itemsPerPage: adaptersPerPage } = usePagination(filteredAdapters)
</script>

<template>
  <FilterCard title="Filters" subtitle="Choose a type, then search for a resource" class="mb-4">
    <div class="filter-card__tabs"><v-btn-toggle v-model="tab" mandatory class="scope-tabs"><v-btn value="images" prepend-icon="mdi-cube-outline">OCI images</v-btn><v-btn value="adapters" prepend-icon="mdi-connection">Adapters</v-btn></v-btn-toggle></div>
    <v-text-field v-model="query" hide-details placeholder="Search…" prepend-inner-icon="mdi-magnify" />
    <v-spacer />
    <v-btn variant="text" prepend-icon="mdi-shield-sync-outline" @click="runChecks">Run checks again</v-btn>
  </FilterCard>
  <SectionCard title="Approved runtime supply chain" subtitle="Immutable images and versioned adapters">
    <template #actions><v-btn color="primary" prepend-icon="mdi-plus" @click="addOpen = true">{{ tab === 'images' ? 'Register image' : 'Add adapter' }}</v-btn></template>
    <div class="table-scroll">
      <table v-if="tab === 'images'" class="data-table"><thead><tr><th>IMAGE</th><th>REGISTRY</th><th>DIGEST</th><th class="table-cell--center">SIGNATURE</th><th class="table-cell--center">SCAN</th><th class="table-cell--center">USAGE</th><th class="table-cell--center" aria-label="Actions"></th></tr></thead><tbody><DataTableEmptyRow v-if="filteredImages.length === 0" :colspan="7" /><tr v-for="image in paginatedImages" :key="image.digest"><td><div class="entity-cell"><div class="entity-icon"><v-icon icon="mdi-cube-outline" /></div><div><div class="entity-name">{{ image.name }}</div><div class="entity-meta">Informational tag: {{ image.tag }}</div></div></div></td><td>{{ image.registry }}</td><td><code>{{ image.digest }}</code></td><td class="table-cell--center"><StatusChip :label="image.signed ? 'Signed' : 'Unsigned'" color="success" icon="mdi-check-decagram-outline" /></td><td class="table-cell--center"><StatusChip :label="image.scan" :color="image.scan === 'Verified' ? 'success' : 'warning'" :icon="image.scan === 'Verified' ? 'mdi-shield-check-outline' : 'mdi-alert-outline'" /></td><td class="table-cell--center">{{ image.used }}</td><td class="table-cell--center"><v-btn icon="mdi-information-outline" variant="text" size="small" :aria-label="`View details for ${image.name}:${image.tag}`" @click="showResourceDetails(`${image.name}:${image.tag}`)" /></td></tr></tbody></table>
      <table v-else class="data-table"><thead><tr><th>ADAPTER</th><th class="table-cell--center">VERSION</th><th class="table-cell--center">CONTRACT</th><th class="table-cell--center">RESUME</th><th class="table-cell--center">STATUS</th><th class="table-cell--center">LAST TEST</th><th class="table-cell--center" aria-label="Actions"></th></tr></thead><tbody><DataTableEmptyRow v-if="filteredAdapters.length === 0" :colspan="7" /><tr v-for="adapter in paginatedAdapters" :key="adapter.version"><td><div class="entity-cell"><div class="entity-icon"><v-icon icon="mdi-connection" /></div><div><div class="entity-name">{{ adapter.name }}</div><div class="entity-meta">Versioned runtime bridge</div></div></div></td><td class="code-text table-cell--center">{{ adapter.version }}</td><td class="table-cell--center">{{ adapter.protocol }}</td><td class="table-cell--center">{{ adapter.resume }}</td><td class="table-cell--center"><StatusChip :label="adapter.status" :color="adapter.status === 'Stable' ? 'success' : 'warning'" icon="mdi-circle-medium" /></td><td class="text-medium-emphasis table-cell--center">{{ adapter.tested }}</td><td class="table-cell--center"><v-btn icon="mdi-information-outline" variant="text" size="small" :aria-label="`View details for ${adapter.name} ${adapter.version}`" @click="showResourceDetails(`${adapter.name} ${adapter.version}`)" /></td></tr></tbody></table>
    </div>
  </SectionCard>
  <TablePaginationCard v-if="tab === 'images'" v-model="imagesPage" :total="filteredImages.length" :items-per-page="imagesPerPage" item-label="images" hint="Always referenced by digest in snapshots" />
  <TablePaginationCard v-else v-model="adaptersPage" :total="filteredAdapters.length" :items-per-page="adaptersPerPage" item-label="adapters" hint="Simulated supply chain" />

  <v-dialog v-model="addOpen" max-width="570"><v-card class="section-card"><div class="pa-6"><div class="d-flex justify-space-between mb-5"><div><h3>{{ tab === 'images' ? 'Register OCI image' : 'Add adapter' }}</h3><div class="text-caption text-medium-emphasis">The resource must pass checks before use</div></div><v-btn icon="mdi-close" variant="text" aria-label="Close" @click="addOpen = false" /></div><v-text-field :label="tab === 'images' ? 'Full reference' : 'Name'" :placeholder="tab === 'images' ? 'registry/image@sha256:…' : 'OpenCode Adapter'" /><div class="form-row"><v-text-field label="Version" placeholder="1.0.0" /><IconSelect label="Scope" :items="['Global', 'Northstar Labs']" model-value="Global" /></div><v-textarea label="Validation notes" rows="3" /></div><v-card-actions class="dialog-actions"><v-spacer/><v-btn @click="addOpen = false">Cancel</v-btn><v-btn color="primary" @click="saveResource">Save & verify</v-btn></v-card-actions></v-card></v-dialog>
</template>
