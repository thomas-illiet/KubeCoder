<script setup lang="ts">
import { shallowRef, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { AdminOrganization, OrganizationMember } from '../../api/organizations'
import type { CurrentUser } from '../../api/users'
import OrganizationFormDialog from '../../components/organizations/OrganizationFormDialog.vue'
import OrganizationMembersPanel from '../../components/organizations/OrganizationMembersPanel.vue'
import AdminSecretsPanel from '../../components/secrets/AdminSecretsPanel.vue'
import SectionCard from '../../components/SectionCard.vue'
import TablePaginationCard from '../../components/TablePaginationCard.vue'
import { useNotifications } from '../../composables/useNotifications'
import { useOrganizations } from '../../composables/useOrganizations'

const route = useRoute()
const router = useRouter()
const service = useOrganizations()
const notifications = useNotifications()
const organization = shallowRef<AdminOrganization | null>(null)
const members = shallowRef<OrganizationMember[]>([])
const memberPage = shallowRef(1)
const memberTotal = shallowRef(0)
const organizationMemberCount = shallowRef(0)
const memberQuery = shallowRef<string | null>('')
const membersPerPage = 20
type MemberSortKey = 'display_name' | 'username' | 'email' | 'joined_at'
type SortDirection = 'asc' | 'desc'
const memberSortBy = shallowRef<MemberSortKey>('display_name')
const memberSortDirection = shallowRef<SortDirection>('asc')
const membersLoading = shallowRef(false)
const users = shallowRef<CurrentUser[]>([])
const loading = shallowRef(false)
const usersLoading = shallowRef(false)
const formOpen = shallowRef(false)
const deleteOpen = shallowRef(false)
const finalDeleteOpen = shallowRef(false)
const deleteConfirmation = shallowRef('')
const deleting = shallowRef(false)
let userSearchRevision = 0
let memberLoadRevision = 0

// formatCreatedAt formats the organization creation date in the user's locale.
function formatCreatedAt(value: string): string {
  return new Intl.DateTimeFormat(undefined, { dateStyle: 'long' }).format(new Date(value))
}

// loadDetail loads the organization and its membership candidates.
async function loadDetail(id: string): Promise<void> {
  loading.value = true
  try {
    const organizations = await service.listAdmin({ limit: 100 })
    organization.value = (organizations.items ?? []).find((item) => item.id === id) ?? null
    if (!organization.value) {
      await router.replace('/admin/organizations')
      return
    }
    await loadMembers(id)
    users.value = []
  } catch (error) {
    notifications.error('Organization details could not be loaded', error instanceof Error ? error.message : undefined)
  } finally {
    loading.value = false
  }
}

// loadMembers requests the current member page and preserves rows during refreshes.
async function loadMembers(id: string, revision = ++memberLoadRevision): Promise<void> {
  membersLoading.value = true
  try {
    const result = await service.listMembers(id, {
      limit: membersPerPage,
      offset: (memberPage.value - 1) * membersPerPage,
      query: memberQuery.value?.trim() ?? '',
      orderBy: memberSortBy.value,
      orderDirection: memberSortDirection.value,
    })
    if (revision !== memberLoadRevision) return
    if (result.items.length === 0 && result.total > 0 && memberPage.value > 1) {
      memberPage.value -= 1
      return
    }
    members.value = result.items ?? []
    memberTotal.value = result.total
    if (!memberQuery.value?.trim()) organizationMemberCount.value = result.total
  } catch (error) {
    if (revision === memberLoadRevision) notifications.error('Organization members could not be loaded', error instanceof Error ? error.message : undefined)
  } finally {
    if (revision === memberLoadRevision) membersLoading.value = false
  }
}

// changeMemberSort toggles the selected server-side order and resets pagination.
function changeMemberSort(column: MemberSortKey): void {
  memberSortDirection.value = memberSortBy.value === column && memberSortDirection.value === 'asc' ? 'desc' : 'asc'
  memberSortBy.value = column
  if (memberPage.value !== 1) memberPage.value = 1
  else if (organization.value) void loadMembers(organization.value.id)
}

// searchUsers loads one bounded page of provisioned users matching an explicit query.
async function searchUsers(query: string): Promise<void> {
  const revision = ++userSearchRevision
  if (query.length < 2) {
    users.value = []
    usersLoading.value = false
    return
  }
  usersLoading.value = true
  try {
    const page = await service.listUsers(query)
    if (revision === userSearchRevision) users.value = page.items ?? []
  } catch (error) {
    if (revision === userSearchRevision) {
      users.value = []
      notifications.error('Users could not be searched', error instanceof Error ? error.message : undefined)
    }
  } finally {
    if (revision === userSearchRevision) usersLoading.value = false
  }
}

// addMember grants membership to a provisioned user.
async function addMember(userID: string): Promise<void> {
  if (!organization.value) return
  try {
    await service.addMember(organization.value.id, userID)
    if (memberQuery.value?.trim()) organizationMemberCount.value += 1
    notifications.success('Member added')
    await loadMembers(organization.value.id)
  } catch (error) {
    notifications.error('Member could not be added', error instanceof Error ? error.message : undefined)
  }
}

// removeMember revokes a user membership.
async function removeMember(userID: string): Promise<void> {
  if (!organization.value) return
  try {
    await service.removeMember(organization.value.id, userID)
    if (memberQuery.value?.trim()) organizationMemberCount.value = Math.max(0, organizationMemberCount.value - 1)
    notifications.success('Member removed')
    await Promise.all([loadMembers(organization.value.id), service.refresh()])
  } catch (error) {
    notifications.error('Member could not be removed', error instanceof Error ? error.message : undefined)
  }
}

// renameOrganization updates the organization name and refreshes its detail.
async function renameOrganization(input: { name: string; slug: string }): Promise<void> {
  if (!organization.value) return
  try {
    const current = organization.value
    const renamed = await service.rename(current.id, input.name)
    organization.value = { ...current, ...renamed }
    formOpen.value = false
    notifications.success('Organization renamed')
  } catch (error) {
    notifications.error('Organization could not be renamed', error instanceof Error ? error.message : undefined)
  }
}

// openDeleteOrganization opens a clean organization deletion confirmation.
function openDeleteOrganization(): void {
  deleteConfirmation.value = ''
  deleteOpen.value = true
}

// closeDeleteOrganization closes the deletion dialog and clears its confirmation.
function closeDeleteOrganization(): void {
  deleteOpen.value = false
  deleteConfirmation.value = ''
}

// requestFinalDeletion advances to the final destructive confirmation.
function requestFinalDeletion(): void {
  if (!organization.value || deleteConfirmation.value !== organization.value.slug) return
  deleteOpen.value = false
  finalDeleteOpen.value = true
}

// cancelFinalDeletion closes the deletion flow and clears its written confirmation.
function cancelFinalDeletion(): void {
  finalDeleteOpen.value = false
  deleteConfirmation.value = ''
}

// deleteOrganization permanently deletes the organization after slug confirmation.
async function deleteOrganization(): Promise<void> {
  if (!organization.value || deleteConfirmation.value !== organization.value.slug || !finalDeleteOpen.value) return
  deleting.value = true
  try {
    await service.remove(organization.value.id)
    await service.refresh()
    notifications.success('Organization deleted')
    await router.replace('/admin/organizations')
  } catch (error) {
    notifications.error('Organization could not be deleted', error instanceof Error ? error.message : undefined)
  } finally {
    deleting.value = false
  }
}

watch(() => route.params.organizationId, (id) => {
  if (typeof id === 'string') {
    memberPage.value = 1
    memberSortBy.value = 'display_name'
    memberSortDirection.value = 'asc'
    memberQuery.value = ''
    void loadDetail(id)
  }
}, { immediate: true })

watch(memberPage, () => {
  if (organization.value) void loadMembers(organization.value.id)
})

watch(memberQuery, (_value, _previous, onCleanup) => {
  const revision = ++memberLoadRevision
  membersLoading.value = true
  const timeout = window.setTimeout(() => {
    if (!organization.value) return
    if (memberPage.value !== 1) memberPage.value = 1
    else void loadMembers(organization.value.id, revision)
  }, 300)
  onCleanup(() => window.clearTimeout(timeout))
})
</script>

<template>
  <v-btn variant="text" prepend-icon="mdi-arrow-left" class="mb-4" to="/admin/organizations">Back to organizations</v-btn>
  <SectionCard v-if="organization" :title="organization.name" :subtitle="organization.slug" icon="mdi-domain">
    <template #actions>
      <div class="organization-actions">
        <v-btn size="small" color="primary" variant="tonal" prepend-icon="mdi-pencil-outline" @click="formOpen = true">Rename</v-btn>
        <v-btn size="small" color="error" variant="tonal" prepend-icon="mdi-delete-outline" @click="openDeleteOrganization">Delete</v-btn>
      </div>
    </template>
    <div class="organization-meta">
      <div class="organization-meta__item">
        <v-icon icon="mdi-calendar-blank-outline" size="18" />
        <div><div class="organization-meta__label">Created</div><div class="organization-meta__value">{{ formatCreatedAt(organization.created_at) }}</div></div>
      </div>
      <div class="organization-meta__item">
        <v-icon icon="mdi-account-multiple-outline" size="18" />
        <div><div class="organization-meta__label">Members</div><div class="organization-meta__value">{{ organizationMemberCount }}</div></div>
      </div>
      <div class="organization-meta__item">
        <v-icon icon="mdi-source-repository" size="18" />
        <div><div class="organization-meta__label">Repositories</div><div class="organization-meta__value">{{ organization.repository_count }}</div></div>
      </div>
    </div>
  </SectionCard>
  <SectionCard class="mt-4" title="Members" subtitle="Provisioned users with access to this organization">
    <OrganizationMembersPanel v-model:query="memberQuery" :members="members" :users="users" :loading="loading || membersLoading" :searching="usersLoading" :sort-by="memberSortBy" :sort-direction="memberSortDirection" @search="searchUsers" @sort="changeMemberSort" @add="addMember" @remove="removeMember" />
  </SectionCard>
  <TablePaginationCard v-model="memberPage" :total="memberTotal" :items-per-page="membersPerPage" item-label="members" />
  <div v-if="organization" class="mt-8"><AdminSecretsPanel :organization-id="organization.id" /></div>

  <OrganizationFormDialog v-if="organization" v-model="formOpen" :organization="organization" @submit="renameOrganization" />
  <v-dialog :model-value="deleteOpen" max-width="620" @update:model-value="!$event && closeDeleteOrganization()">
    <v-card v-if="organization" class="section-card">
      <div class="pa-8">
        <div class="d-flex align-center ga-4 mb-6">
          <v-avatar size="48" color="error" variant="tonal"><v-icon icon="mdi-delete-alert-outline" /></v-avatar>
          <div><h3>Delete organization?</h3><div class="text-body-2 text-medium-emphasis">{{ organization.name }} · {{ organization.slug }}</div></div>
        </div>
        <div class="removal-notice" role="alert">
          <v-icon class="removal-notice__icon" icon="mdi-alert-outline" size="28" />
          <div><div class="removal-notice__title">The organization will be permanently deleted</div><p>Every membership will be removed immediately. This action cannot be undone.</p></div>
        </div>
        <div class="delete-confirmation">
          <p>Type <strong>{{ organization.slug }}</strong> to confirm.</p>
          <v-text-field v-model="deleteConfirmation" label="Organization slug" hide-details autofocus />
        </div>
      </div>
      <v-card-actions class="dialog-actions"><v-spacer /><v-btn rounded="lg" variant="text" prepend-icon="mdi-close" @click="closeDeleteOrganization">Cancel</v-btn><v-btn min-width="180" rounded="lg" color="error" variant="flat" prepend-icon="mdi-arrow-right" :disabled="deleteConfirmation !== organization.slug" @click="requestFinalDeletion">Continue</v-btn></v-card-actions>
    </v-card>
  </v-dialog>

  <v-dialog :model-value="finalDeleteOpen" max-width="620" persistent>
    <v-card v-if="organization" class="section-card">
      <div class="pa-8">
        <div class="d-flex align-center ga-4 mb-6">
          <v-avatar size="48" color="error" variant="tonal"><v-icon icon="mdi-delete-alert-outline" /></v-avatar>
          <div><h3>Are you absolutely sure?</h3><div class="text-body-2 text-medium-emphasis">Delete {{ organization.name }} permanently</div></div>
        </div>
        <div class="removal-notice" role="alert">
          <v-icon class="removal-notice__icon" icon="mdi-alert-octagon-outline" size="28" />
          <div><div class="removal-notice__title">This is your final confirmation</div><p>The organization, all of its memberships, and its configuration will be permanently removed. You will not be able to recover it.</p></div>
        </div>
      </div>
      <v-card-actions class="dialog-actions"><v-btn rounded="lg" color="success" variant="tonal" prepend-icon="mdi-shield-check-outline" :disabled="deleting" @click="cancelFinalDeletion">No, keep organization</v-btn><v-spacer /><v-btn min-width="210" rounded="lg" color="error" variant="flat" prepend-icon="mdi-delete-alert-outline" :loading="deleting" @click="deleteOrganization">Yes, delete organization</v-btn></v-card-actions>
    </v-card>
  </v-dialog>
</template>

<style scoped>
.organization-actions { display: flex; align-items: center; gap: 8px; }
.organization-meta { display: flex; gap: 32px; padding: 16px 20px; }
.organization-meta__item { display: flex; align-items: center; gap: 10px; color: rgb(var(--v-theme-on-surface)); }
.organization-meta__label { color: var(--kc-muted); font-size: 11px; line-height: 1.2; }
.organization-meta__value { margin-top: 3px; font-size: 13px; font-weight: 600; }
.removal-notice { display: flex; gap: 16px; align-items: flex-start; padding: 22px 24px; color: rgb(var(--v-theme-on-surface)); background: rgba(var(--v-theme-error), .11); border: 1px solid rgba(var(--v-theme-error), .32); border-radius: 14px; }
.removal-notice__icon { flex: 0 0 auto; margin-top: 1px; color: rgb(var(--v-theme-error)); }
.removal-notice__title { margin-bottom: 7px; color: #f3f5f9; font-size: 15px; font-weight: 700; }
.removal-notice p { max-width: 480px; margin: 0; color: #c0c7d4; font-size: 14px; line-height: 1.65; }
.delete-confirmation { margin-top: 24px; }
.delete-confirmation p { margin: 0 0 10px; color: #c0c7d4; font-size: 14px; }

@media (max-width: 600px) {
  .organization-actions { width: 100%; }
  .organization-actions .v-btn { flex: 1; }
  .organization-meta { flex-direction: column; gap: 14px; }
}
</style>
