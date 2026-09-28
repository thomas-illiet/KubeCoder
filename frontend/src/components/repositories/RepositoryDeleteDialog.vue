<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'
import type { Repository } from '../../api/repositories'

const props = defineProps<{ repository: Repository | null; deleting?: boolean }>()
const emit = defineEmits<{ close: []; confirm: [] }>()
const confirmation = shallowRef('')
const valid = computed(() => Boolean(props.repository && confirmation.value === props.repository.name))
watch(() => props.repository, () => { confirmation.value = '' })
</script>

<template>
  <v-dialog :model-value="Boolean(repository)" max-width="540" @update:model-value="!$event && emit('close')"><v-card v-if="repository" class="section-card"><div class="pa-7"><h3 class="mb-3">Delete {{ repository.name }}?</h3><p class="text-body-2 text-medium-emphasis mb-5">The repository configuration and its agent binding will be permanently deleted. Type <strong>{{ repository.name }}</strong> to confirm.</p><v-text-field v-model="confirmation" label="Repository name" autofocus /></div><v-card-actions class="dialog-actions"><v-spacer /><v-btn @click="emit('close')">Cancel</v-btn><v-btn color="error" :disabled="!valid" :loading="deleting" @click="emit('confirm')">Delete permanently</v-btn></v-card-actions></v-card></v-dialog>
</template>
