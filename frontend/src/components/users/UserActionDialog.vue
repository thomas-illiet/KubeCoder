<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'
import type { CurrentUser } from '../../api/users'

const props = defineProps<{
  modelValue: boolean
  user: CurrentUser | null
  action: 'role' | 'delete'
  loading: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  confirm: [isAdmin?: boolean]
}>()

const title = computed(() => props.action === 'delete'
  ? 'Delete user'
  : 'Update role')
const description = computed(() => {
  const name = props.user?.display_name || props.user?.username || 'this user'
  if (props.action === 'delete') return 'This user account and all of its organization memberships will be deleted from KubeCoder.'
  return `Choose the platform access granted to ${name}.`
})
const selectedRole = shallowRef<'user' | 'admin'>('user')
const deleteConfirmation = shallowRef('')
const confirmationName = computed(() => props.user?.display_name || props.user?.username || '')
const confirmDisabled = computed(() => props.action === 'delete'
  ? deleteConfirmation.value !== confirmationName.value
  : selectedRole.value === (props.user?.is_admin ? 'admin' : 'user'))
const roleOptions = [
  { title: 'User', value: 'user', props: { subtitle: 'Organization workspace access only' } },
  { title: 'Administrator', value: 'admin', props: { subtitle: 'Full platform administration access' } },
]

watch(() => [props.modelValue, props.user] as const, ([open, user]) => {
  if (open && user) {
    selectedRole.value = user.is_admin ? 'admin' : 'user'
    deleteConfirmation.value = ''
  }
}, { immediate: true })
</script>

<template>
  <v-dialog :model-value="modelValue" max-width="520" @update:model-value="emit('update:modelValue', $event)">
    <v-card class="app-dialog">
      <div class="dialog-header">
        <div><div class="text-h6 font-weight-bold">{{ title }}</div><div class="text-caption text-medium-emphasis">{{ user?.email || user?.username }}</div></div>
        <v-btn icon="mdi-close" variant="text" aria-label="Close" :disabled="loading" @click="emit('update:modelValue', false)" />
      </div>
      <div class="pa-5">
        <div class="user-action-warning" :class="{ 'user-action-warning--delete': action === 'delete' }">
          <v-icon :icon="action === 'delete' ? 'mdi-alert-outline' : 'mdi-shield-account-outline'" />
          <p>{{ description }}</p>
        </div>
        <v-select
          v-if="action === 'role'"
          v-model="selectedRole"
          class="mt-4"
          :items="roleOptions"
          label="Role"
          prepend-inner-icon="mdi-shield-account-outline"
          hide-details
        />
        <v-text-field
          v-else
          v-model="deleteConfirmation"
          class="mt-4"
          :label="`Type ${confirmationName} to confirm`"
          :placeholder="confirmationName"
          autocomplete="off"
          autofocus
          hide-details
        />
      </div>
      <v-card-actions class="dialog-actions">
        <v-spacer />
        <v-btn variant="text" :disabled="loading" @click="emit('update:modelValue', false)">Cancel</v-btn>
        <v-btn
          :color="action === 'delete' ? 'error' : 'primary'"
          variant="flat"
          :loading="loading"
          :disabled="confirmDisabled"
          @click="emit('confirm', action === 'role' ? selectedRole === 'admin' : undefined)"
        >{{ action === 'delete' ? 'Delete user' : 'Update role' }}</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<style scoped>
.user-action-warning { display: flex; gap: 13px; align-items: flex-start; padding: 16px; border: 1px solid rgba(129, 147, 255, .3); border-radius: 12px; color: #cbd4ff; background: rgba(129, 147, 255, .08); }
.user-action-warning--delete { border-color: rgba(251, 113, 133, .35); color: #ffb3c0; background: rgba(251, 113, 133, .08); }
.user-action-warning p { margin: 0; color: #c7ceda; font-size: 13px; line-height: 1.55; }
</style>
