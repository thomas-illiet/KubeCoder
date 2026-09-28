<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import type { SecretInput, SecretScope } from '../../api/secrets'

const props = defineProps<{ modelValue: boolean; organizationMode?: boolean; saving?: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean]; submit: [input: SecretInput] }>()
const form = reactive({ scope: 'PLATFORM' as SecretScope, variable_name: '', description: '', value: '', expires_at: '' })
const valid = computed(() => /^[A-Z_][A-Z0-9_]*$/.test(form.variable_name) && form.value.length > 0)
watch(() => props.modelValue, (open) => { if (open) Object.assign(form, { scope: props.organizationMode ? 'ORGANIZATION' : 'PLATFORM', variable_name: '', description: '', value: '', expires_at: '' }); else form.value = '' })
function close() { form.value = ''; emit('update:modelValue', false) }
function submit() { if (!valid.value) return; emit('submit', { scope: form.scope, variable_name: form.variable_name, description: form.description, value: form.value, expires_at: form.expires_at ? new Date(`${form.expires_at}T23:59:59`).toISOString() : null }); form.value = '' }
function normalizeVariableName(value: string) { form.variable_name = value.toUpperCase().replace(/\s+/g, '_') }
</script>
<template><v-dialog :model-value="modelValue" max-width="640" @update:model-value="!$event && close()"><v-card class="app-dialog"><div class="dialog-header"><div><div class="text-h6 font-weight-bold">Create secret</div><div class="text-caption text-medium-emphasis">The value is write-only and cannot be recovered</div></div><v-btn icon="mdi-close" variant="text" @click="close" /></div><div class="pa-5"><v-text-field :model-value="form.variable_name" label="Variable name" placeholder="MODEL_API_TOKEN" @update:model-value="normalizeVariableName" /><v-textarea v-model="form.description" label="Description" rows="2" /><v-text-field v-model="form.expires_at" label="Expiration date" type="date" /><v-text-field v-model="form.value" label="Secret value" type="password" autocomplete="new-password" /><div class="security-note"><v-icon icon="mdi-eye-off-outline" /><span>The value is encrypted after submission and never returned by the API.</span></div></div><v-card-actions class="dialog-actions"><v-spacer /><v-btn @click="close">Cancel</v-btn><v-btn color="primary" :disabled="!valid" :loading="saving" @click="submit">Create secret</v-btn></v-card-actions></v-card></v-dialog></template>
