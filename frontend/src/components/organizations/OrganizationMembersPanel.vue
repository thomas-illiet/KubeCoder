<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'
import type { OrganizationMember } from '../../api/organizations'
import type { CurrentUser } from '../../api/users'
import DataTableEmptyRow from '../DataTableEmptyRow.vue'

type MemberSortKey = 'display_name' | 'username' | 'email' | 'joined_at'
type SortDirection = 'asc' | 'desc'
const props = defineProps<{
  members: OrganizationMember[]
  users: CurrentUser[]
  sortBy: MemberSortKey
  sortDirection: SortDirection
  loading?: boolean
  searching?: boolean
}>()
const emit = defineEmits<{
  add: [userID: string]
  remove: [userID: string]
  search: [query: string]
  sort: [column: MemberSortKey]
}>()
const memberQuery = defineModel<string | null>('query', { default: '' })
const selectedUser = shallowRef<CurrentUser | null>(null)
const addMemberQuery = shallowRef('')
const addMemberOpen = shallowRef(false)
const removeTarget = shallowRef<CurrentUser | null>(null)
const availableUsers = computed(() => props.users.filter((user) => !props.members.some((member) => member.id === user.id)))
const noDataText = computed(() => addMemberQuery.value.trim().length < 2 ? 'Type at least 2 characters to search.' : 'No provisioned users found.')
const hasMemberQuery = computed(() => Boolean(memberQuery.value?.trim()))

// formatJoinedAt formats the date on which organization access was granted.
function formatJoinedAt(value: string): string {
  return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' }).format(new Date(value))
}

function ariaSort(column: MemberSortKey): 'ascending' | 'descending' | 'none' {
  if (props.sortBy !== column) return 'none'
  return props.sortDirection === 'asc' ? 'ascending' : 'descending'
}

function sortIcon(column: MemberSortKey): string {
  if (props.sortBy !== column) return 'mdi-unfold-more-horizontal'
  return props.sortDirection === 'asc' ? 'mdi-arrow-up' : 'mdi-arrow-down'
}

watch(addMemberQuery, (value, _previous, onCleanup) => {
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
  addMemberQuery.value = ''
  addMemberOpen.value = true
}

// closeAddMember closes the dialog and clears its transient search state.
function closeAddMember(): void {
  addMemberOpen.value = false
  selectedUser.value = null
  addMemberQuery.value = ''
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
    <div class="member-toolbar pa-4">
      <v-text-field
        v-model="memberQuery"
        class="member-search"
        hide-details
        clearable
        placeholder="Search members…"
        prepend-inner-icon="mdi-magnify"
        aria-label="Search organization members"
      />
      <v-spacer />
      <v-btn color="primary" variant="flat" rounded="lg" prepend-icon="mdi-account-plus-outline" @click="openAddMember">Add member</v-btn>
    </div>
    <div class="member-table-shell" :aria-busy="loading">
      <v-progress-linear v-if="loading" class="member-table-progress" indeterminate color="primary" />
      <div class="table-scroll member-table-content" :class="{ 'member-table-content--loading': loading }">
      <table class="data-table">
        <thead><tr>
          <th :aria-sort="ariaSort('display_name')"><button class="sort-header" type="button" @click="emit('sort', 'display_name')">MEMBER<v-icon :icon="sortIcon('display_name')" size="16" /></button></th>
          <th :aria-sort="ariaSort('username')"><button class="sort-header" type="button" @click="emit('sort', 'username')">USERNAME<v-icon :icon="sortIcon('username')" size="16" /></button></th>
          <th :aria-sort="ariaSort('email')"><button class="sort-header" type="button" @click="emit('sort', 'email')">EMAIL<v-icon :icon="sortIcon('email')" size="16" /></button></th>
          <th :aria-sort="ariaSort('joined_at')"><button class="sort-header" type="button" @click="emit('sort', 'joined_at')">ADDED<v-icon :icon="sortIcon('joined_at')" size="16" /></button></th>
          <th aria-label="Actions"></th>
        </tr></thead>
        <tbody>
          <tr v-for="member in members" :key="member.id">
            <td>{{ member.display_name }}</td><td>{{ member.username }}</td><td>{{ member.email }}</td><td>{{ formatJoinedAt(member.joined_at) }}</td>
            <td class="table-cell--center"><v-btn size="small" color="error" variant="tonal" prepend-icon="mdi-delete-outline" :aria-label="`Remove ${member.display_name}`" @click="requestRemove(member)">Delete</v-btn></td>
          </tr>
          <DataTableEmptyRow
            v-if="!loading && members.length === 0"
            :colspan="5"
            :title="hasMemberQuery ? 'No members found' : 'No members assigned'"
            :description="hasMemberQuery ? 'Adjust your search to display organization members.' : 'Add a provisioned user to grant access to this organization.'"
            icon="mdi-account-off-outline"
          />
        </tbody>
      </table>
      </div>
    </div>

    <v-dialog :model-value="addMemberOpen" max-width="680" @update:model-value="!$event && closeAddMember()">
      <v-card class="section-card">
        <div class="pa-8">
          <h3 class="mb-2">Add an organization member</h3>
          <p class="text-body-2 text-medium-emphasis mb-6">Search for a user who has already signed in to the platform.</p>
          <v-autocomplete
            v-model="selectedUser"
            v-model:search="addMemberQuery"
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
.member-toolbar { display: flex; align-items: center; gap: 16px; }
.member-search { max-width: 420px; }
.member-table-shell { position: relative; overflow: hidden; }
.member-table-progress { position: absolute; z-index: 2; top: 0; right: 0; left: 0; }
.member-table-content { transition: opacity .2s ease, filter .2s ease; }
.member-table-content--loading { pointer-events: none; opacity: .42; filter: saturate(.7); }
.sort-header { display: inline-flex; align-items: center; gap: .35rem; padding: 0; border: 0; background: transparent; color: inherit; font: inherit; letter-spacing: inherit; cursor: pointer; }
.sort-header:focus-visible { outline: 2px solid rgb(var(--v-theme-primary)); outline-offset: 3px; border-radius: 2px; }
.removal-notice { display: flex; gap: 16px; align-items: flex-start; padding: 22px 24px; color: rgb(var(--v-theme-on-surface)); background: rgba(var(--v-theme-error), .11); border: 1px solid rgba(var(--v-theme-error), .32); border-radius: 14px; }
.removal-notice__icon { flex: 0 0 auto; margin-top: 1px; color: rgb(var(--v-theme-error)); }
.removal-notice__title { margin-bottom: 7px; color: #f3f5f9; font-size: 15px; font-weight: 700; }
.removal-notice p { max-width: 480px; margin: 0; color: #c0c7d4; font-size: 14px; line-height: 1.65; }

@media (max-width: 700px) {
  .member-toolbar { align-items: stretch; flex-direction: column; }
  .member-search { max-width: none; }
}
</style>
