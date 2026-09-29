<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'
import type { Secret } from '../../api/secrets'

const props = defineProps<{ secret: Secret | null; deleting?: boolean }>()
const emit = defineEmits<{ close: []; confirm: [] }>()
const confirmation = shallowRef('')
const finalConfirmation = shallowRef(false)
const valid = computed(() => Boolean(props.secret && confirmation.value === props.secret.variable_name))

watch(() => props.secret, () => {
  confirmation.value = ''
  finalConfirmation.value = false
})

function close(): void {
  finalConfirmation.value = false
  emit('close')
}
</script>

<template>
  <v-dialog :model-value="Boolean(secret) && !finalConfirmation" max-width="620" @update:model-value="!$event && close()">
    <v-card v-if="secret" class="section-card">
      <div class="pa-8">
        <div class="dialog-heading">
          <v-avatar size="48" color="error" variant="tonal"><v-icon icon="mdi-delete-alert-outline" /></v-avatar>
          <div><h3>Delete secret?</h3><div class="text-body-2 text-medium-emphasis code-text">{{ secret.variable_name }}</div></div>
        </div>
        <div class="removal-notice" role="alert">
          <v-icon class="removal-notice__icon" icon="mdi-alert-outline" size="28" />
          <div>
            <div class="removal-notice__title">Applications may immediately lose access</div>
            <p>The protected value and every binding ({{ secret.binding_count }} currently) will be permanently deleted. Agents and repositories using this secret may stop working. This action cannot be undone.</p>
          </div>
        </div>
        <div class="delete-confirmation">
          <p>Type <strong>{{ secret.variable_name }}</strong> to confirm.</p>
          <v-text-field v-model="confirmation" label="Variable name" hide-details autofocus />
        </div>
      </div>
      <v-card-actions class="dialog-actions"><v-spacer /><v-btn rounded="lg" variant="text" prepend-icon="mdi-close" @click="close">Cancel</v-btn><v-btn min-width="180" rounded="lg" color="error" variant="flat" prepend-icon="mdi-arrow-right" :disabled="!valid" @click="finalConfirmation = true">Continue</v-btn></v-card-actions>
    </v-card>
  </v-dialog>

  <v-dialog :model-value="Boolean(secret) && finalConfirmation" max-width="620" persistent>
    <v-card v-if="secret" class="section-card">
      <div class="pa-8">
        <div class="dialog-heading">
          <v-avatar size="48" color="error" variant="tonal"><v-icon icon="mdi-delete-alert-outline" /></v-avatar>
          <div><h3>Are you absolutely sure?</h3><div class="text-body-2 text-medium-emphasis">Delete <span class="code-text">{{ secret.variable_name }}</span> permanently</div></div>
        </div>
        <div class="removal-notice" role="alert">
          <v-icon class="removal-notice__icon" icon="mdi-alert-octagon-outline" size="28" />
          <div><div class="removal-notice__title">This is your final confirmation</div><p>The secret value and every binding will be removed immediately. The value cannot be viewed or recovered after deletion.</p></div>
        </div>
      </div>
      <v-card-actions class="dialog-actions"><v-btn rounded="lg" color="success" variant="tonal" prepend-icon="mdi-shield-check-outline" :disabled="deleting" @click="close">No, keep secret</v-btn><v-spacer /><v-btn min-width="190" rounded="lg" color="error" variant="flat" prepend-icon="mdi-delete-alert-outline" :loading="deleting" @click="emit('confirm')">Yes, delete secret</v-btn></v-card-actions>
    </v-card>
  </v-dialog>
</template>

<style scoped>
.dialog-heading { display: flex; align-items: center; gap: 16px; margin-bottom: 24px; }
.removal-notice { display: flex; gap: 16px; align-items: flex-start; padding: 22px 24px; color: rgb(var(--v-theme-on-surface)); background: rgba(var(--v-theme-error), .11); border: 1px solid rgba(var(--v-theme-error), .32); border-radius: 14px; }
.removal-notice__icon { flex: 0 0 auto; margin-top: 1px; color: rgb(var(--v-theme-error)); }
.removal-notice__title { margin-bottom: 7px; color: #f3f5f9; font-size: 15px; font-weight: 700; }
.removal-notice p { max-width: 480px; margin: 0; color: #c0c7d4; font-size: 14px; line-height: 1.65; }
.delete-confirmation { margin-top: 24px; }
.delete-confirmation p { margin: 0 0 10px; color: #c0c7d4; font-size: 14px; }
</style>
