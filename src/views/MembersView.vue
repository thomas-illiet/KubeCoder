<script setup lang="ts">
import { computed, ref } from 'vue'
import DataTableEmptyRow from '../components/DataTableEmptyRow.vue'
import FilterCard from '../components/FilterCard.vue'
import SectionCard from '../components/SectionCard.vue'
import StatusChip from '../components/StatusChip.vue'
import TablePaginationCard from '../components/TablePaginationCard.vue'
import { useNotifications } from '../composables/useNotifications'
import { usePagination } from '../composables/usePagination'

const inviteOpen = ref(false)
const roleDialog = ref(false)
const selectedMember = ref('')
const query = ref('')
const roleFilter = ref('All roles')
const { success, warning, info } = useNotifications()
const members = [
  { name: 'Alex Martin', email: 'alex.martin@example.test', initials: 'AM', role: 'OWNER', status: 'Active', last: 'Now', color: '#7c86ff' },
  { name: 'Samira Chen', email: 'samira.chen@example.test', initials: 'SC', role: 'ADMIN', status: 'Active', last: '1 hour ago', color: '#33d2d0' },
  { name: 'Morgan Lee', email: 'morgan.lee@example.test', initials: 'ML', role: 'MEMBER', status: 'Active', last: 'Yesterday', color: '#5aa7ff' },
  { name: 'Jamie Rivera', email: 'jamie.rivera@example.test', initials: 'JR', role: 'MEMBER', status: 'Invited', last: 'Invitation sent', color: '#f6b84c' },
  { name: 'Robin Dupont', email: 'robin.dupont@example.test', initials: 'RD', role: 'MEMBER', status: 'Suspended', last: '18 days ago', color: '#ff687d' },
]
const filteredMembers = computed(() => members.filter((member) => {
  const search = query.value.trim().toLowerCase()
  const matchesSearch = !search || `${member.name} ${member.email}`.toLowerCase().includes(search)
  const matchesRole = roleFilter.value === 'All roles' || member.role === roleFilter.value
  return matchesSearch && matchesRole
}))
const { page, paginatedItems: paginatedMembers, itemsPerPage } = usePagination(filteredMembers)
function changeRole(name: string) { selectedMember.value = name; roleDialog.value = true }
function sendInvitation() { inviteOpen.value = false; success('Invitation sent', 'The member will receive an email with a secure sign-in link.') }
function updateRole() { roleDialog.value = false; success('Role updated', `${selectedMember.value}'s permissions were updated.`) }
function resendInvitation(name: string) { success('Invitation resent', `A new invitation was sent to ${name}.`) }
function suspendAccess(name: string) { warning('Access suspended', `${name} can no longer access the organization.`) }
function viewActivity(name: string) { info('Member activity', `${name}'s recent organization activity is ready for review.`) }
</script>

<template>
  <div class="stat-grid mb-4">
    <v-card class="stat-card"><div class="stat-card__top"><span>Active members</span><div class="stat-card__icon"><v-icon icon="mdi-account-check-outline" /></div></div><div class="stat-card__value">3</div><div class="stat-card__caption">Out of 5 members</div></v-card>
    <v-card class="stat-card"><div class="stat-card__top"><span>Administrators</span><div class="stat-card__icon"><v-icon icon="mdi-shield-account-outline" /></div></div><div class="stat-card__value">2</div><div class="stat-card__caption">1 owner · 1 admin</div></v-card>
    <v-card class="stat-card"><div class="stat-card__top"><span>Invitations</span><div class="stat-card__icon"><v-icon icon="mdi-email-fast-outline" /></div></div><div class="stat-card__value">1</div><div class="stat-card__caption">Awaiting acceptance</div></v-card>
    <v-card class="stat-card"><div class="stat-card__top"><span>Suspended access</span><div class="stat-card__icon" style="color:#f6b84c;background:rgba(246,184,76,.1)"><v-icon icon="mdi-account-cancel-outline" /></div></div><div class="stat-card__value">1</div><div class="stat-card__caption">Access history retained</div></v-card>
  </div>
  <FilterCard title="Filters" subtitle="Search for a member or role" class="mb-4">
    <v-text-field v-model="query" hide-details placeholder="Search members…" prepend-inner-icon="mdi-magnify" />
    <IconSelect v-model="roleFilter" hide-details :items="['All roles', 'OWNER', 'ADMIN', 'MEMBER']" />
  </FilterCard>
  <SectionCard title="Organization members" subtitle="Local roles independent of the OIDC provider">
    <template #actions><v-btn color="primary" prepend-icon="mdi-account-plus-outline" @click="inviteOpen = true">Invite</v-btn></template>
    <div class="table-scroll"><table class="data-table"><thead><tr><th>MEMBER</th><th class="table-cell--center">ROLE</th><th class="table-cell--center">STATUS</th><th class="table-cell--center">LAST ACTIVITY</th><th>ACCESS</th><th class="table-cell--center" aria-label="Actions"></th></tr></thead><tbody>
      <DataTableEmptyRow v-if="filteredMembers.length === 0" :colspan="6" />
      <tr v-for="member in paginatedMembers" :key="member.email"><td><div class="entity-cell"><div class="org-avatar" :style="{color:member.color,background:`${member.color}15`}">{{ member.initials }}</div><div><div class="entity-name">{{ member.name }}</div><div class="entity-meta">{{ member.email }}</div></div></div></td><td class="table-cell--center"><v-chip size="small" :color="member.role === 'OWNER' ? 'primary' : member.role === 'ADMIN' ? 'secondary' : 'default'" variant="tonal">{{ member.role }}</v-chip></td><td class="table-cell--center"><StatusChip :label="member.status" :color="member.status === 'Active' ? 'success' : member.status === 'Invited' ? 'warning' : 'error'" :icon="member.status === 'Active' ? 'mdi-check-circle-outline' : 'mdi-circle-medium'" /></td><td class="text-medium-emphasis table-cell--center">{{ member.last }}</td><td><span class="text-body-2">All repositories</span></td><td class="table-cell--center"><v-menu><template #activator="{ props }"><v-btn v-bind="props" icon="mdi-dots-horizontal" variant="text" size="small" :aria-label="`Actions for ${member.name}`" /></template><v-list class="pa-2" width="210"><v-list-item title="Change role" prepend-icon="mdi-account-cog-outline" @click="changeRole(member.name)" /><v-list-item title="View activity" prepend-icon="mdi-history" @click="viewActivity(member.name)" /><v-list-item title="Resend invitation" prepend-icon="mdi-email-fast-outline" v-if="member.status === 'Invited'" @click="resendInvitation(member.name)" /><v-divider class="my-2" /><v-list-item title="Suspend access" prepend-icon="mdi-account-cancel-outline" class="text-error" @click="suspendAccess(member.name)" /></v-list></v-menu></td></tr>
    </tbody></table></div>
  </SectionCard>
  <TablePaginationCard v-model="page" :total="filteredMembers.length" :items-per-page="itemsPerPage" item-label="members" />

  <v-dialog v-model="inviteOpen" max-width="520"><v-card class="section-card"><div class="pa-6"><div class="d-flex justify-space-between mb-5"><div><h3>Invite member</h3><div class="text-caption text-medium-emphasis">The identity will be linked through OIDC upon acceptance</div></div><v-btn icon="mdi-close" variant="text" aria-label="Close" @click="inviteOpen = false" /></div><v-text-field label="Email address" placeholder="name@company.test" /><IconSelect label="Role" :items="['MEMBER', 'ADMIN']" model-value="MEMBER" /><div class="security-note"><v-icon icon="mdi-shield-account-outline" color="info" /><span>Application roles remain managed locally and do not depend on any IdP-specific claim.</span></div></div><v-card-actions class="dialog-actions"><v-spacer/><v-btn @click="inviteOpen = false">Cancel</v-btn><v-btn color="primary" @click="sendInvitation">Send invitation</v-btn></v-card-actions></v-card></v-dialog>
  <v-dialog v-model="roleDialog" max-width="440"><v-card class="section-card"><div class="pa-6"><h3 class="mb-2">Change role</h3><p class="text-body-2 text-medium-emphasis mb-5">{{ selectedMember }}</p><IconSelect label="New role" :items="['ADMIN', 'MEMBER']" model-value="MEMBER" /><v-checkbox label="I confirm the impact on permissions" color="primary" /></div><v-card-actions class="dialog-actions"><v-spacer/><v-btn @click="roleDialog = false">Cancel</v-btn><v-btn color="primary" @click="updateRole">Update</v-btn></v-card-actions></v-card></v-dialog>
</template>
