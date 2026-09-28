<script setup lang="ts">
import { shallowRef, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { Organization, OrganizationMember } from '../../api/organizations'
import type { CurrentUser } from '../../api/users'
import OrganizationFormDialog from '../../components/organizations/OrganizationFormDialog.vue'
import OrganizationMembersPanel from '../../components/organizations/OrganizationMembersPanel.vue'
import SectionCard from '../../components/SectionCard.vue'
import { useNotifications } from '../../composables/useNotifications'
import { useOrganizations } from '../../composables/useOrganizations'

const route = useRoute()
const router = useRouter()
const service = useOrganizations()
const notifications = useNotifications()
const organization = shallowRef<Organization | null>(null)
const members = shallowRef<OrganizationMember[]>([])
const users = shallowRef<CurrentUser[]>([])
const loading = shallowRef(false)
const usersLoading = shallowRef(false)
const formOpen = shallowRef(false)
const deleteOpen = shallowRef(false)
const deleteConfirmation = shallowRef('')
let userSearchRevision = 0

// formatCreatedAt formats the organization creation date in the user's locale.
function formatCreatedAt(value: string): string {
  return new Intl.DateTimeFormat(undefined, { dateStyle: 'long' }).format(new Date(value))
}

// loadDetail loads the organization and its membership candidates.
async function loadDetail(id: string): Promise<void> {
  loading.value = true
  try {
    const [organizations, memberPage] = await Promise.all([
      service.listAdmin(), service.listMembers(id),
    ])
    organization.value = (organizations.items ?? []).find((item) => item.id === id) ?? null
    if (!organization.value) {
      await router.replace('/admin/organizations')
      return
    }
    members.value = memberPage.items ?? []
    users.value = []
  } catch (error) {
    notifications.error('Organization details could not be loaded', error instanceof Error ? error.message : undefined)
  } finally {
    loading.value = false
  }
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
    notifications.success('Member added')
    await loadDetail(organization.value.id)
  } catch (error) {
    notifications.error('Member could not be added', error instanceof Error ? error.message : undefined)
  }
}

// removeMember revokes a user membership.
async function removeMember(userID: string): Promise<void> {
  if (!organization.value) return
  try {
    await service.removeMember(organization.value.id, userID)
    notifications.success('Member removed')
    await Promise.all([loadDetail(organization.value.id), service.refresh()])
  } catch (error) {
    notifications.error('Member could not be removed', error instanceof Error ? error.message : undefined)
  }
}

// renameOrganization updates the organization name and refreshes its detail.
async function renameOrganization(input: { name: string; slug: string }): Promise<void> {
  if (!organization.value) return
  try {
    organization.value = await service.rename(organization.value.id, input.name)
    formOpen.value = false
    notifications.success('Organization renamed')
  } catch (error) {
    notifications.error('Organization could not be renamed', error instanceof Error ? error.message : undefined)
  }
}

// deleteOrganization permanently deletes the organization after slug confirmation.
async function deleteOrganization(): Promise<void> {
  if (!organization.value || deleteConfirmation.value !== organization.value.slug) return
  try {
    await service.remove(organization.value.id)
    await service.refresh()
    notifications.success('Organization deleted')
    await router.replace('/admin/organizations')
  } catch (error) {
    notifications.error('Organization could not be deleted', error instanceof Error ? error.message : undefined)
  }
}

watch(() => route.params.organizationId, (id) => {
  if (typeof id === 'string') void loadDetail(id)
}, { immediate: true })
</script>

<template>
  <v-btn variant="text" prepend-icon="mdi-arrow-left" class="mb-4" to="/admin/organizations">Back to organizations</v-btn>
  <SectionCard v-if="organization" :title="organization.name" :subtitle="organization.slug" icon="mdi-domain">
    <template #actions>
      <div class="organization-actions">
        <v-btn size="small" color="primary" variant="tonal" prepend-icon="mdi-pencil-outline" @click="formOpen = true">Rename</v-btn>
        <v-btn size="small" color="error" variant="tonal" prepend-icon="mdi-delete-outline" @click="deleteOpen = true">Delete</v-btn>
      </div>
    </template>
    <div class="organization-meta">
      <div class="organization-meta__item">
        <v-icon icon="mdi-calendar-blank-outline" size="18" />
        <div><div class="organization-meta__label">Created</div><div class="organization-meta__value">{{ formatCreatedAt(organization.created_at) }}</div></div>
      </div>
      <div class="organization-meta__item">
        <v-icon icon="mdi-account-multiple-outline" size="18" />
        <div><div class="organization-meta__label">Members</div><div class="organization-meta__value">{{ members.length }}</div></div>
      </div>
    </div>
  </SectionCard>
  <SectionCard class="mt-4" title="Members" subtitle="Provisioned users with access to this organization">
    <OrganizationMembersPanel :members="members" :users="users" :loading="loading" :searching="usersLoading" @search="searchUsers" @add="addMember" @remove="removeMember" />
  </SectionCard>

  <OrganizationFormDialog v-if="organization" v-model="formOpen" :organization="organization" @submit="renameOrganization" />
  <v-dialog v-model="deleteOpen" max-width="600">
    <v-card v-if="organization" class="section-card">
      <div class="pa-8"><h3 class="mb-2">Delete organization permanently</h3>
        <p class="text-body-2 text-medium-emphasis mb-6">Type <strong>{{ organization.slug }}</strong> to confirm. Every membership will be removed.</p>
        <v-text-field v-model="deleteConfirmation" label="Organization slug" hide-details />
      </div>
      <v-card-actions class="dialog-actions"><v-spacer /><v-btn rounded="lg" variant="text" prepend-icon="mdi-close" @click="deleteOpen = false">Cancel</v-btn><v-btn min-width="180" rounded="lg" color="error" variant="flat" :disabled="deleteConfirmation !== organization.slug" @click="deleteOrganization">Delete permanently</v-btn></v-card-actions>
    </v-card>
  </v-dialog>
</template>

<style scoped>
.organization-actions { display: flex; align-items: center; gap: 8px; }
.organization-meta { display: flex; gap: 32px; padding: 16px 20px; }
.organization-meta__item { display: flex; align-items: center; gap: 10px; color: rgb(var(--v-theme-on-surface)); }
.organization-meta__label { color: var(--kc-muted); font-size: 11px; line-height: 1.2; }
.organization-meta__value { margin-top: 3px; font-size: 13px; font-weight: 600; }

@media (max-width: 600px) {
  .organization-actions { width: 100%; }
  .organization-actions .v-btn { flex: 1; }
  .organization-meta { flex-direction: column; gap: 14px; }
}
</style>
