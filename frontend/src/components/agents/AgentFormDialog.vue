<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import type { Agent, AgentInput } from '../../api/agents'

const props = defineProps<{ modelValue: boolean; agent?: Agent | null; saving?: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean]; submit: [input: AgentInput] }>()
const form = reactive({ name: '', description: '', runtime_adapter: 'opencode', runtime_version: '', image: '', provider: '', model: '', system_prompt: '', capabilities: '', cpu_millis: 1000, memory_mb: 2048, storage_mb: 10240, max_duration_seconds: 3600, active: true })
const valid = computed(() => Boolean(form.name.trim() && form.description.trim() && form.runtime_adapter.trim() && form.runtime_version.trim() && form.image.trim() && form.provider.trim() && form.model.trim() && form.system_prompt.trim() && form.cpu_millis > 0 && form.memory_mb > 0 && form.storage_mb > 0 && form.max_duration_seconds > 0))

watch(() => props.modelValue, (open) => {
  if (!open) return
  const agent = props.agent
  Object.assign(form, agent ? { name: agent.name, description: agent.description, runtime_adapter: agent.runtime_adapter, runtime_version: agent.runtime_version, image: agent.image, provider: agent.provider, model: agent.model, system_prompt: agent.system_prompt, capabilities: agent.capabilities.join(', '), cpu_millis: agent.cpu_millis, memory_mb: agent.memory_mb, storage_mb: agent.storage_mb, max_duration_seconds: agent.max_duration_seconds, active: agent.active } : { name: '', description: '', runtime_adapter: 'opencode', runtime_version: '', image: '', provider: '', model: '', system_prompt: '', capabilities: '', cpu_millis: 1000, memory_mb: 2048, storage_mb: 10240, max_duration_seconds: 3600, active: true })
})

function submit(): void {
  if (!valid.value) return
  emit('submit', { name: form.name.trim(), description: form.description.trim(), runtime_adapter: form.runtime_adapter.trim(), runtime_version: form.runtime_version.trim(), image: form.image.trim(), provider: form.provider.trim(), model: form.model.trim(), system_prompt: form.system_prompt.trim(), capabilities: form.capabilities.split(',').map((value) => value.trim()).filter(Boolean), cpu_millis: form.cpu_millis, memory_mb: form.memory_mb, storage_mb: form.storage_mb, max_duration_seconds: form.max_duration_seconds, active: form.active })
}
</script>

<template>
  <v-dialog :model-value="modelValue" max-width="900" persistent scrollable @update:model-value="emit('update:modelValue', $event)">
    <v-card class="app-dialog"><div class="dialog-header"><div><div class="text-h6 font-weight-bold">{{ agent ? `Edit ${agent.name}` : 'Create agent' }}</div><div class="text-caption text-medium-emphasis">Platform-managed runtime definition</div></div><v-btn icon="mdi-close" variant="text" aria-label="Close" @click="emit('update:modelValue', false)" /></div>
      <div class="pa-6"><div class="form-section"><div class="form-section__title">Public profile</div><div class="form-row"><v-text-field v-model="form.name" label="Name" maxlength="120" /><v-switch v-model="form.active" label="Available" color="primary" /></div><v-textarea v-model="form.description" label="Description" rows="2" maxlength="1000" /><v-text-field v-model="form.capabilities" label="Capabilities" hint="Comma-separated public capabilities" persistent-hint /></div>
        <div class="form-section"><div class="form-section__title">Runtime</div><div class="form-row"><v-text-field v-model="form.runtime_adapter" label="Adapter" /><v-text-field v-model="form.runtime_version" label="Runtime version" /></div><v-text-field v-model="form.image" label="OCI image" placeholder="registry.example/opencode@sha256:…" /><div class="form-row"><v-text-field v-model="form.provider" label="Model provider" /><v-text-field v-model="form.model" label="Model" /></div><v-textarea v-model="form.system_prompt" label="System prompt" rows="4" /></div>
        <div class="form-section"><div class="form-section__title">Maximum resources</div><div class="form-row"><v-text-field v-model.number="form.cpu_millis" type="number" label="CPU (millicores)" min="1" /><v-text-field v-model.number="form.memory_mb" type="number" label="Memory (MB)" min="1" /></div><div class="form-row"><v-text-field v-model.number="form.storage_mb" type="number" label="Storage (MB)" min="1" /><v-text-field v-model.number="form.max_duration_seconds" type="number" label="Duration (seconds)" min="1" /></div></div></div>
      <v-card-actions class="dialog-actions"><v-spacer /><v-btn variant="text" @click="emit('update:modelValue', false)">Cancel</v-btn><v-btn color="primary" :disabled="!valid" :loading="saving" @click="submit">{{ agent ? 'Save changes' : 'Create agent' }}</v-btn></v-card-actions></v-card>
  </v-dialog>
</template>
