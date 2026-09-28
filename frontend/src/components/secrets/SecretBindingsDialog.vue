<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'
import type { Secret, SecretTarget, SecretTargetType } from '../../api/secrets'
const props = defineProps<{ modelValue: boolean; secret: Secret | null; targets: readonly SecretTarget[]; saving?: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean]; add: [input: { target_type: SecretTargetType; target_id: string }]; remove: [bindingID: string] }>()
const targetType = shallowRef<SecretTargetType>('AGENT'); const targetID = shallowRef<string | null>(null)
const allowedTypes = computed<SecretTargetType[]>(() => props.secret?.scope === 'PLATFORM' ? ['AGENT'] : ['AGENT', 'REPOSITORY'])
const available = computed(() => props.targets.filter(item => item.type === targetType.value && !props.secret?.bindings.some(binding => binding.target_id === item.id)))
watch(() => props.modelValue, open => { if(open){targetType.value=allowedTypes.value[0] ?? 'AGENT';targetID.value=null} })
watch(targetType,()=>{targetID.value=null})
</script>
<template><v-dialog :model-value="modelValue" max-width="620" @update:model-value="$emit('update:modelValue',$event)"><v-card class="app-dialog"><div class="dialog-header"><div><div class="text-h6">Bindings</div><code>{{ secret?.variable_name }}</code></div><v-btn icon="mdi-close" variant="text" @click="$emit('update:modelValue',false)" /></div><div class="pa-5"><div class="form-row"><IconSelect v-model="targetType" label="Target type" :items="allowedTypes" /><IconSelect v-model="targetID" label="Target" :items="available.map(item=>({title:item.name,value:item.id}))" /></div><div class="d-flex justify-end mb-4"><v-btn color="primary" prepend-icon="mdi-link-plus" :disabled="!targetID" :loading="saving" @click="targetID && $emit('add',{target_type:targetType,target_id:targetID})">Add binding</v-btn></div><v-list v-if="secret?.bindings.length" bg-color="transparent"><v-list-item v-for="binding in secret.bindings" :key="binding.id" :title="binding.target_name" :subtitle="binding.target_type"><template #append><v-btn icon="mdi-link-off" variant="text" color="error" @click="$emit('remove',binding.id)" /></template></v-list-item></v-list><div v-else class="empty-state py-6"><p>No explicit bindings.</p></div></div></v-card></v-dialog></template>
