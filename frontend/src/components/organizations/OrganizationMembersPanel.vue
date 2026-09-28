<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'
import type { OrganizationMember } from '../../api/organizations'
import type { CurrentUser } from '../../api/users'
import DataTableEmptyRow from '../DataTableEmptyRow.vue'

const props = defineProps<{ members: OrganizationMember[]; users: CurrentUser[]; loading?: boolean; searching?: boolean }>()
const emit = defineEmits<{ add: [userID: string]; remove: [userID: string]; search: [query: string] }>()
const selectedUser = shallowRef<CurrentUser | null>(null)
const searchQuery = shallowRef('')
const addMemberOpen = shallowRef(false)
const removeTarget = shallowRef<CurrentUser | null>(null)
const availableUsers = computed(() => props.users.filter((user) => !props.members.some((member) => member.id === user.id)))
const noDataText = computed(() => searchQuery.value.trim().length < 2 ? 'Type at least 2 characters to search.' : 'No provisioned users found.')

// formatJoinedAt formats the date on which organization access was granted.
function formatJoinedAt(value: string): string {
  return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' }).format(new Date(value))
}

watch(searchQuery, (value, _previous, onCleanup) => {
  const query = value.trim()
  if (selectedUser.value && query === selectedUser.value.display_name) return
  if (query.length < 2) {
    emit('search', '')
    return
  }
  const timeout = window.setTimeout(() => emit('search', query), 300)
  onCleanup(() => window.clearTimeout(timeout))
})

// addSelected emits the selected provisioned user and clears the selection.
function addSelected(): void {
  if (!selectedUser.value) return
  emit('add', selectedUser.value.id)
  closeAddMember()
}

// openAddMember opens a clean member search dialog.
function openAddMember(): void {
  selectedUser.value = null
  searchQuery.value = ''
  addMemberOpen.value = true
}

// closeAddMember closes the dialog and clears its transient search state.
function closeAddMember(): void {
  addMemberOpen.value = false
  selectedUser.value = null
  searchQuery.value = ''
}

// requestRemove opens the membership removal confirmation dialog.
function requestRemove(member: CurrentUser): void {
  removeTarget.value = member
}

// confirmRemove emits the confirmed membership removal.
function confirmRemove(): void {
  if (!removeTarget.value) return
  emit('remove', removeTarget.value.id)
  removeTarget.value = null
}
</script>

<template>
  <div>
    <div class="d-flex justify-end pa-4">
      <v-btn color="primary" variant="flat" rounded="lg" prepend-icon="mdi-account-plus-outline" @click="openAddMember">Add member</v-btn>
    </div>
    <v-progress-linear v-if="loading" indeterminate />
    <div v-else class="table-scroll">
      <table class="data-table">
        <thead><tr><th>MEMBER</th><th>USERNAME</th><th>EMAIL</th><th>ADDED</th><th aria-label="Actions"></th></tr></thead>
        <tbody>
          <tr v-for="member in members" :key="member.id">
            <td>{{ member.display_name }}</td><td>{{ member.username }}</td><td>{{ member.email }}</td><td>{{ formatJoinedAt(member.joined_at) }}</td>
            <td class="table-cell--center"><v-btn size="small" color="error" variant="tonal" prepend-icon="mdi-delete-outline" :aria-label="`Remove ${member.display_name}`" @click="requestRemove(member)">Delete</v-btn></td>
          </tr>
          <DataTableEmptyRow
            v-if="members.length === 0"
            :colspan="5"
            title="No members assigned"
            description="Add a provisioned user to grant access to this organization."
            icon="mdi-account-off-outline"
          />
        </tbody>
      </table>
    </div>

    <v-dialog :model-value="addMemberOpen" max-width="680" @update:model-value="!$event && closeAddMember()">
      <v-card class="section-card">
        <div class="pa-8">
          <h3 class="mb-2">Add an organization member</h3>
          <p class="text-body-2 text-medium-emphasis mb-6">Search for a user who has already signed in to the platform.</p>
          <v-autocomplete
            v-model="selectedUser"
            v-model:search="searchQuery"
            :items="availableUsers"
            :loading="searching"
            :no-data-text="noDataText"
            item-title="display_name"
            item-value="id"
            return-object
            label="Search a provisioned user"
            prepend-inner-icon="mdi-account-search-outline"
            no-filter
            clearable
            autofocus
          >
            <template #item="{ props: itemProps, item }">
              <v-list-item v-bind="itemProps" :subtitle="`${item.raw.username} · ${item.raw.email}`" />
            </template>
          </v-autocomplete>
        </div>
        <v-card-actions class="dialog-actions">
          <v-spacer />
          <v-btn rounded="lg" variant="text" prepend-icon="mdi-close" @click="closeAddMember">Cancel</v-btn>
          <v-btn min-width="140" rounded="lg" color="primary" variant="flat" prepend-icon="mdi-account-plus-outline" :disabled="!selectedUser" @click="addSelected">Add member</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog :model-value="Boolean(removeTarget)" max-width="620" @update:model-value="!$event && (removeTarget = null)">
      <v-card v-if="removeTarget" class="section-card">
        <div class="pa-8">
          <div class="d-flex align-center ga-4 mb-6">
            <v-avatar size="48" color="error" variant="tonal"><v-icon icon="mdi-account-remove-outline" /></v-avatar>
            <div><h3>Remove organization member?</h3><div class="text-body-2 text-medium-emphasis">{{ removeTarget.display_name }} · {{ removeTarget.email }}</div></div>
          </div>
          <div class="removal-notice" role="alert">
            <v-icon class="removal-notice__icon" icon="mdi-alert-outline" size="28" />
            <div><div class="removal-notice__title">Organization access will be revoked</div><p>The user will immediately lose access to this organization. Their platform account and access to other organizations will not be deleted.</p></div>
          </div>
        </div>
        <v-card-actions class="dialog-actions">
          <v-spacer />
          <v-btn rounded="lg" variant="text" prepend-icon="mdi-close" @click="removeTarget = null">Cancel</v-btn>
          <v-btn min-width="150" rounded="lg" color="error" variant="flat" prepend-icon="mdi-account-remove-outline" @click="confirmRemove">Remove access</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<style scoped>
.removal-notice { display: flex; gap: 16px; align-items: flex-start; padding: 22px 24px; color: rgb(var(--v-theme-on-surface)); background: rgba(var(--v-theme-error), .11); border: 1px solid rgba(var(--v-theme-error), .32); border-radius: 14px; }
.removal-notice__icon { flex: 0 0 auto; margin-top: 1px; color: rgb(var(--v-theme-error)); }
.removal-notice__title { margin-bottom: 7px; color: #f3f5f9; font-size: 15px; font-weight: 700; }
.removal-notice p { max-width: 480px; margin: 0; color: #c0c7d4; font-size: 14px; line-height: 1.65; }
</style>
