<script setup lang="ts">
import type { Repository } from '../../api/repositories'
import DataTableEmptyRow from '../DataTableEmptyRow.vue'
import StatusChip from '../StatusChip.vue'

defineProps<{ items: readonly Repository[]; loading: boolean }>()
const emit = defineEmits<{ edit: [repository: Repository]; delete: [repository: Repository] }>()
</script>

<template>
  <div class="table-scroll" :aria-busy="loading"><v-progress-linear v-if="loading" indeterminate /><table class="data-table"><thead><tr><th>REPOSITORY</th><th>PROVIDER</th><th>BRANCH</th><th>AGENT</th><th>STATUS</th><th aria-label="Actions"></th></tr></thead><tbody>
    <DataTableEmptyRow v-if="!loading && items.length === 0" :colspan="6" title="No repositories found" description="Add a repository or adjust the current filters." icon="mdi-source-repository" />
    <tr v-for="repository in items" :key="repository.id"><td><div class="entity-cell"><div class="entity-icon"><v-icon icon="mdi-source-repository" /></div><div><div class="entity-name code-text">{{ repository.name }}</div><div class="entity-meta">{{ repository.clone_url }}</div></div></div></td><td class="text-capitalize">{{ repository.provider }}</td><td><v-chip size="small" variant="outlined">{{ repository.default_branch }}</v-chip></td><td>{{ repository.agent?.name ?? 'Not configured' }}</td><td><StatusChip :label="repository.agent ? 'Ready' : 'Action required'" :color="repository.agent ? 'success' : 'warning'" /></td><td class="text-right"><v-btn icon="mdi-pencil-outline" variant="text" size="small" aria-label="Edit repository" @click="emit('edit', repository)" /><v-btn icon="mdi-delete-outline" variant="text" color="error" size="small" aria-label="Delete repository" @click="emit('delete', repository)" /></td></tr>
  </tbody></table></div>
</template>
