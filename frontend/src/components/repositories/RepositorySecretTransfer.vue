<script setup lang="ts">
import { computed, shallowRef } from 'vue'
import type { Secret } from '../../api/secrets'

const props = defineProps<{ secrets: readonly Secret[]; modelValue: readonly string[] }>()
const emit = defineEmits<{ 'update:modelValue': [value: string[]] }>()

const query = shallowRef('')
const selectedIDs = computed(() => new Set(props.modelValue))
const normalizedQuery = computed(() => query.value.trim().toLocaleLowerCase())
const visibleSecrets = computed(() => props.secrets.filter((secret) => {
  if (!normalizedQuery.value) return true
  return `${secret.variable_name} ${secret.description ?? ''} ${secret.scope}`.toLocaleLowerCase().includes(normalizedQuery.value)
}))
const availableSecrets = computed(() => visibleSecrets.value.filter((secret) => !selectedIDs.value.has(secret.id)))
const usedSecrets = computed(() => visibleSecrets.value.filter((secret) => selectedIDs.value.has(secret.id)))

function add(secretID: string): void {
  if (selectedIDs.value.has(secretID)) return
  emit('update:modelValue', [...props.modelValue, secretID])
}

function remove(secretID: string): void {
  emit('update:modelValue', props.modelValue.filter((id) => id !== secretID))
}

function addVisible(): void {
  emit('update:modelValue', [...new Set([...props.modelValue, ...availableSecrets.value.map((secret) => secret.id)])])
}

function removeVisible(): void {
  const visibleUsed = new Set(usedSecrets.value.map((secret) => secret.id))
  emit('update:modelValue', props.modelValue.filter((id) => !visibleUsed.has(id)))
}
</script>

<template>
  <div class="secret-transfer">
    <div class="secret-transfer__filters">
      <v-text-field v-model="query" label="Search secrets" prepend-inner-icon="mdi-magnify" clearable hide-details />
    </div>
    <div class="secret-transfer__grid">
      <v-card class="secret-transfer__card" variant="flat">
        <div class="secret-transfer__header"><div><strong>Available</strong><span>{{ availableSecrets.length }} secrets</span></div><v-btn icon="mdi-chevron-double-right" size="small" variant="text" :disabled="availableSecrets.length === 0" aria-label="Use all visible secrets" @click="addVisible" /></div>
        <div class="secret-transfer__list">
          <button v-for="secret in availableSecrets" :key="secret.id" type="button" class="secret-transfer__item" @click="add(secret.id)">
            <span><strong>{{ secret.variable_name }}</strong><small>{{ secret.scope }}</small></span><v-icon icon="mdi-chevron-right" size="20" />
          </button>
          <div v-if="availableSecrets.length === 0" class="secret-transfer__empty">No available secrets match this filter.</div>
        </div>
      </v-card>
      <v-icon class="secret-transfer__direction" icon="mdi-swap-horizontal" size="24" />
      <v-card class="secret-transfer__card secret-transfer__card--used" variant="flat">
        <div class="secret-transfer__header"><div><strong>Used by repository</strong><span>{{ usedSecrets.length }} shown · {{ modelValue.length }} total</span></div><v-btn icon="mdi-chevron-double-left" size="small" variant="text" :disabled="usedSecrets.length === 0" aria-label="Remove all visible secrets" @click="removeVisible" /></div>
        <div class="secret-transfer__list">
          <button v-for="secret in usedSecrets" :key="secret.id" type="button" class="secret-transfer__item" @click="remove(secret.id)">
            <v-icon icon="mdi-chevron-left" size="20" /><span><strong>{{ secret.variable_name }}</strong><small>{{ secret.scope }}</small></span>
          </button>
          <div v-if="usedSecrets.length === 0" class="secret-transfer__empty">No used secrets match this filter.</div>
        </div>
      </v-card>
    </div>
  </div>
</template>
