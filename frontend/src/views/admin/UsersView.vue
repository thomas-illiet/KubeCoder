<script setup lang="ts">
import { shallowRef } from 'vue'
import type { CurrentUser } from '../../api/users'
import FilterCard from '../../components/FilterCard.vue'
import SectionCard from '../../components/SectionCard.vue'
import TablePaginationCard from '../../components/TablePaginationCard.vue'
import UserActionDialog from '../../components/users/UserActionDialog.vue'
import UserTable from '../../components/users/UserTable.vue'
import { useAuth } from '../../composables/useAuth'
import { useAdminUsers } from '../../composables/useAdminUsers'

const auth = useAuth()
const { users, query, role, page, total, loading, mutating, orderBy, orderDirection, itemsPerPage, reload, changeSort, updateRole, remove } = useAdminUsers()
const selectedUser = shallowRef<CurrentUser | null>(null)
const action = shallowRef<'role' | 'delete'>('role')
const actionOpen = shallowRef(false)

const roleOptions = [
  { title: 'All roles', value: 'all' },
  { title: 'Administrators', value: 'admin' },
  { title: 'Users', value: 'user' },
]

function openAction(user: CurrentUser, nextAction: 'role' | 'delete'): void {
  selectedUser.value = user
  action.value = nextAction
  actionOpen.value = true
}

async function confirmAction(isAdmin?: boolean): Promise<void> {
  if (!selectedUser.value) return
  const succeeded = action.value === 'delete'
    ? await remove(selectedUser.value)
    : await updateRole(selectedUser.value, isAdmin ?? selectedUser.value.is_admin)
  if (succeeded) actionOpen.value = false
}
</script>

<template>
  <FilterCard title="Users" subtitle="Search identities provisioned through OpenID Connect" class="mb-4">
    <v-text-field v-model="query" hide-details clearable placeholder="Search by name, username, or email…" prepend-inner-icon="mdi-magnify" />
    <v-select v-model="role" :items="roleOptions" hide-details label="Role" prepend-inner-icon="mdi-shield-account-outline" max-width="220" />
    <v-spacer />
    <v-btn variant="outlined" prepend-icon="mdi-refresh" :loading="loading" @click="reload()">Refresh</v-btn>
  </FilterCard>

  <SectionCard title="Platform users" :subtitle="`${total} provisioned ${total === 1 ? 'user' : 'users'}`">
    <UserTable
      :users="users"
      :loading="loading"
      :current-user-id="auth.state.currentUser?.id ?? ''"
      :order-by="orderBy"
      :order-direction="orderDirection"
      @sort="changeSort"
      @change-role="openAction($event, 'role')"
      @delete="openAction($event, 'delete')"
    />
  </SectionCard>

  <TablePaginationCard v-model="page" :total="total" :items-per-page="itemsPerPage" item-label="users" hint="Accounts are synchronized at sign-in" />

  <UserActionDialog v-model="actionOpen" :user="selectedUser" :action="action" :loading="mutating" @confirm="confirmAction" />
</template>
