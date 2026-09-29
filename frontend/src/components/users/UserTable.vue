<script setup lang="ts">
import type { CurrentUser } from '../../api/users'
import DataTableEmptyRow from '../DataTableEmptyRow.vue'
import StatusChip from '../StatusChip.vue'

defineProps<{
  users: readonly CurrentUser[]
  loading: boolean
  currentUserId: string
  orderBy: string
  orderDirection: 'asc' | 'desc'
}>()

type SortKey = 'display_name' | 'email' | 'is_admin' | 'created_at' | 'updated_at'
const emit = defineEmits<{
  sort: [column: SortKey]
  changeRole: [user: CurrentUser]
  delete: [user: CurrentUser]
}>()

function initials(user: CurrentUser): string {
  const source = user.display_name || user.username || user.email
  return source.split(/\s+/).slice(0, 2).map((part) => part[0]?.toUpperCase()).join('') || 'US'
}

function formatDate(value: string): string {
  return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' }).format(new Date(value))
}
</script>

<template>
  <div class="user-table-shell" :aria-busy="loading">
    <v-progress-linear v-if="loading" class="user-table-progress" indeterminate color="primary" />
    <div class="table-scroll user-table-content" :class="{ 'user-table-content--loading': loading }">
      <table class="data-table">
        <thead>
          <tr>
            <th :aria-sort="orderBy === 'display_name' ? (orderDirection === 'asc' ? 'ascending' : 'descending') : 'none'"><button class="sort-header" type="button" @click="emit('sort', 'display_name')">USER<v-icon :icon="orderBy === 'display_name' ? (orderDirection === 'asc' ? 'mdi-arrow-up' : 'mdi-arrow-down') : 'mdi-unfold-more-horizontal'" size="16" /></button></th>
            <th class="table-cell--center" :aria-sort="orderBy === 'is_admin' ? (orderDirection === 'asc' ? 'ascending' : 'descending') : 'none'"><button class="sort-header" type="button" @click="emit('sort', 'is_admin')">ROLE<v-icon :icon="orderBy === 'is_admin' ? (orderDirection === 'asc' ? 'mdi-arrow-up' : 'mdi-arrow-down') : 'mdi-unfold-more-horizontal'" size="16" /></button></th>
            <th>PREFERRED ORGANIZATION</th>
            <th class="table-cell--center" :aria-sort="orderBy === 'created_at' ? (orderDirection === 'asc' ? 'ascending' : 'descending') : 'none'"><button class="sort-header" type="button" @click="emit('sort', 'created_at')">PROVISIONED<v-icon :icon="orderBy === 'created_at' ? (orderDirection === 'asc' ? 'mdi-arrow-up' : 'mdi-arrow-down') : 'mdi-unfold-more-horizontal'" size="16" /></button></th>
            <th class="table-cell--center" :aria-sort="orderBy === 'updated_at' ? (orderDirection === 'asc' ? 'ascending' : 'descending') : 'none'"><button class="sort-header" type="button" @click="emit('sort', 'updated_at')">LAST SYNCHRONIZED<v-icon :icon="orderBy === 'updated_at' ? (orderDirection === 'asc' ? 'mdi-arrow-up' : 'mdi-arrow-down') : 'mdi-unfold-more-horizontal'" size="16" /></button></th>
            <th class="table-cell--center"><span class="sr-only">Actions</span></th>
          </tr>
        </thead>
        <tbody>
          <DataTableEmptyRow
            v-if="!loading && users.length === 0"
            :colspan="6"
            title="No users found"
            description="Users appear here after their first successful OpenID Connect sign-in."
            icon="mdi-account-search-outline"
          />
          <tr v-for="user in users" :key="user.id">
            <td>
              <div class="user-identity">
                <div class="user-identity__avatar">{{ initials(user) }}</div>
                <div class="user-identity__copy">
                  <strong>{{ user.display_name || user.username }}</strong>
                  <span>{{ user.email || user.username }}</span>
                </div>
              </div>
            </td>
            <td class="table-cell--center">
              <StatusChip
                :label="user.is_admin ? 'Administrator' : 'User'"
                :color="user.is_admin ? 'warning' : 'info'"
                :icon="user.is_admin ? 'mdi-shield-account-outline' : 'mdi-account-outline'"
              />
            </td>
            <td>
              <div v-if="user.preferred_organization" class="organization-cell">
                <span>{{ user.preferred_organization.name }}</span>
                <small>{{ user.preferred_organization.slug }}</small>
              </div>
              <span v-else class="text-medium-emphasis">Not selected</span>
            </td>
            <td class="table-cell--center">{{ formatDate(user.created_at) }}</td>
            <td class="table-cell--center">{{ formatDate(user.updated_at) }}</td>
            <td class="table-cell--center">
              <v-menu>
                <template #activator="{ props: menuProps }"><v-btn v-bind="menuProps" icon="mdi-dots-vertical" size="small" variant="text" :aria-label="`Actions for ${user.display_name || user.username}`" /></template>
                <v-list density="compact">
                  <v-list-item
                    title="Update role"
                    prepend-icon="mdi-account-edit-outline"
                    :disabled="user.id === currentUserId"
                    @click="emit('changeRole', user)"
                  />
                  <v-list-item title="Delete user" prepend-icon="mdi-delete-outline" base-color="error" :disabled="user.id === currentUserId" @click="emit('delete', user)" />
                </v-list>
              </v-menu>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.user-table-shell { position: relative; overflow: hidden; }
.user-table-progress { position: absolute; z-index: 2; top: 0; right: 0; left: 0; }
.user-table-content { transition: opacity .2s ease, filter .2s ease; }
.user-table-content--loading { pointer-events: none; opacity: .42; filter: saturate(.7); }
.user-identity { display: flex; align-items: center; gap: 12px; min-width: 210px; }
.user-identity__avatar { flex: 0 0 auto; width: 36px; height: 36px; display: grid; place-items: center; border: 1px solid #3b4658; border-radius: 10px; color: #cbd4ff; background: #252c38; font-size: 11px; font-weight: 800; }
.user-identity__copy { display: flex; min-width: 0; flex-direction: column; }
.user-identity__copy strong { color: #eef1f7; font-size: 13px; }
.user-identity__copy span { overflow: hidden; max-width: 280px; color: #8692a5; font-size: 11px; text-overflow: ellipsis; white-space: nowrap; }
.organization-cell { display: flex; flex-direction: column; }
.organization-cell small { color: #748096; }
.sort-header { display: inline-flex; align-items: center; gap: .35rem; padding: 0; border: 0; background: transparent; color: inherit; font: inherit; letter-spacing: inherit; cursor: pointer; }
.sort-header:focus-visible { outline: 2px solid rgb(var(--v-theme-primary)); outline-offset: 3px; border-radius: 2px; }
.sr-only { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip: rect(0, 0, 0, 0); white-space: nowrap; border: 0; }
</style>
