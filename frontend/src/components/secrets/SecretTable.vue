<script setup lang="ts">
import type { Secret, SecretSortBy, SortOrder } from '../../api/secrets'
import DataTableEmptyRow from '../DataTableEmptyRow.vue'

const props = defineProps<{ items: readonly Secret[]; loading: boolean; readonly?: boolean; platformReadonly?: boolean; sortBy: SecretSortBy; sortOrder: SortOrder }>()
const emit = defineEmits<{ details: [secret: Secret]; replace: [secret: Secret]; bindings: [secret: Secret]; delete: [secret: Secret]; sort: [sortBy: SecretSortBy, sortOrder: SortOrder] }>()

const formatDate = (value: string | null) => value ? new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' }).format(new Date(value)) : 'No expiration'

function toggleSort(key: SecretSortBy) {
  emit('sort', key, props.sortBy === key && props.sortOrder === 'asc' ? 'desc' : 'asc')
}

function sortIcon(key: SecretSortBy) {
  if (props.sortBy !== key) return 'mdi-swap-vertical'
  return props.sortOrder === 'asc' ? 'mdi-arrow-up' : 'mdi-arrow-down'
}

</script>

<template>
  <div class="table-scroll" :aria-busy="loading">
    <v-progress-linear v-if="loading" indeterminate />
    <table class="data-table">
      <thead>
        <tr>
          <th><button class="sort-button" type="button" @click="toggleSort('variable_name')">VARIABLE NAME<v-icon :icon="sortIcon('variable_name')" size="16" /></button></th>
          <th class="table-cell--center"><button class="sort-button sort-button--center" type="button" @click="toggleSort('scope')">SCOPE<v-icon :icon="sortIcon('scope')" size="16" /></button></th>
          <th class="table-cell--center"><button class="sort-button sort-button--center" type="button" @click="toggleSort('binding_count')">BINDINGS<v-icon :icon="sortIcon('binding_count')" size="16" /></button></th>
          <th class="table-cell--center"><button class="sort-button sort-button--center" type="button" @click="toggleSort('value_replaced_at')">VALUE UPDATED<v-icon :icon="sortIcon('value_replaced_at')" size="16" /></button></th>
          <th class="table-cell--center"><button class="sort-button sort-button--center" type="button" @click="toggleSort('expires_at')">EXPIRATION<v-icon :icon="sortIcon('expires_at')" size="16" /></button></th>
          <th class="table-cell--center">ACTION</th>
        </tr>
      </thead>
      <tbody>
        <DataTableEmptyRow v-if="!loading && items.length === 0" :colspan="6" title="No secrets found" description="No protected values match the current filters." icon="mdi-key-off-outline" />
        <tr v-for="secret in items" :key="secret.id">
          <td><div class="entity-cell"><div class="entity-icon"><v-icon icon="mdi-key-variant" /></div><div><div class="entity-name code-text">{{ secret.variable_name }}</div><div class="entity-meta">{{ secret.description }}</div></div></div></td>
          <td class="table-cell--center"><v-chip size="small" variant="outlined">{{ secret.scope }}</v-chip></td>
          <td class="table-cell--center">{{ secret.binding_count }}</td><td class="table-cell--center">{{ formatDate(secret.value_replaced_at) }}</td><td class="table-cell--center">{{ formatDate(secret.expires_at) }}</td>
          <td class="table-cell--center">
            <v-btn v-if="platformReadonly && secret.scope === 'PLATFORM'" size="small" variant="text" prepend-icon="mdi-lock-outline" disabled>Platform managed</v-btn>
            <v-menu v-else>
              <template #activator="{ props: activatorProps }">
                <v-btn v-bind="activatorProps" size="small" variant="outlined" append-icon="mdi-chevron-down">Actions</v-btn>
              </template>
              <v-list density="compact">
                <v-list-item v-if="readonly" prepend-icon="mdi-information-outline" title="Details" @click="$emit('details', secret)" />
                <template v-else>
                  <v-list-item prepend-icon="mdi-refresh" title="Replace value" @click="$emit('replace', secret)" />
                  <v-list-item prepend-icon="mdi-link-variant" title="Manage bindings" @click="$emit('bindings', secret)" />
                  <v-divider />
                  <v-list-item prepend-icon="mdi-delete-outline" title="Delete" base-color="error" @click="$emit('delete', secret)" />
                </template>
              </v-list>
            </v-menu>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.sort-button {
  align-items: center;
  background: none;
  border: 0;
  color: inherit;
  cursor: pointer;
  display: inline-flex;
  font: inherit;
  gap: 0.35rem;
  letter-spacing: inherit;
  padding: 0;
  text-align: left;
}

.sort-button--center {
  justify-content: center;
  width: 100%;
}

</style>
