<script setup lang="ts">
import { shallowRef, watch } from 'vue'
import type { Secret } from '../../api/secrets'
const props = defineProps<{ modelValue: boolean; secret: Secret | null; saving?: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean]; submit: [input: { value: string; expires_at: string | null }] }>()
const value = shallowRef(''); const expiresAt = shallowRef('')
watch(() => props.modelValue, (open) => { value.value = ''; if (open) expiresAt.value = props.secret?.expires_at?.slice(0, 10) ?? '' })
function close(){ value.value='';emit('update:modelValue',false) }
function submit(){ if(!value.value)return;emit('submit',{value:value.value,expires_at:expiresAt.value?new Date(`${expiresAt.value}T23:59:59`).toISOString():null});value.value='' }
</script>
<template><v-dialog :model-value="modelValue" max-width="540" @update:model-value="!$event && close()"><v-card class="app-dialog"><div class="dialog-header"><div><div class="text-h6">Replace value</div><code>{{ secret?.variable_name }}</code></div><v-btn icon="mdi-close" variant="text" @click="close" /></div><div class="pa-5"><v-alert type="warning" variant="tonal" class="mb-4">The previous encrypted value will be permanently overwritten.</v-alert><v-text-field v-model="value" label="New value" type="password" autocomplete="new-password" /><v-text-field v-model="expiresAt" label="Expiration date" type="date" /></div><v-card-actions class="dialog-actions"><v-spacer /><v-btn @click="close">Cancel</v-btn><v-btn color="primary" :disabled="!value" :loading="saving" @click="submit">Replace permanently</v-btn></v-card-actions></v-card></v-dialog></template>
