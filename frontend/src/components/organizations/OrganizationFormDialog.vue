<script setup lang="ts">
import { shallowRef, watch } from 'vue'
import type { Organization } from '../../api/organizations'

const props = defineProps<{ modelValue: boolean; organization?: Organization | null }>()
const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  submit: [input: { name: string; slug: string }]
}>()

const name = shallowRef('')
const slug = shallowRef('')
const slugManuallyEdited = shallowRef(false)

// slugify converts an organization name into a valid URL slug.
function slugify(value: string): string {
  return value
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 63)
    .replace(/-+$/g, '')
}

watch(() => props.modelValue, (open) => {
  if (!open) return
  slugManuallyEdited.value = false
  name.value = props.organization?.name ?? ''
  slug.value = props.organization?.slug ?? ''
})

watch(name, (value) => {
  if (!props.organization && !slugManuallyEdited.value) slug.value = slugify(value)
})

// updateSlug records an explicit slug override made by the user.
function updateSlug(value: string): void {
  slugManuallyEdited.value = true
  slug.value = value
}

// submit emits the validated organization form values.
function submit(): void {
  if (!name.value.trim() || !slug.value.trim()) return
  emit('submit', { name: name.value.trim(), slug: slug.value.trim().toLowerCase() })
}
</script>

<template>
  <v-dialog :model-value="modelValue" max-width="600" @update:model-value="emit('update:modelValue', $event)">
    <v-card class="section-card">
      <div class="pa-8">
        <h3 class="mb-6">{{ organization ? 'Rename organization' : 'Create organization' }}</h3>
        <v-text-field v-model="name" label="Name" maxlength="120" />
        <v-text-field :model-value="slug" label="Slug" maxlength="63" :disabled="Boolean(organization)" hint="Lowercase letters, numbers, and hyphens. The slug is immutable." persistent-hint @update:model-value="updateSlug" />
      </div>
      <v-card-actions class="dialog-actions">
        <v-spacer />
        <v-btn rounded="lg" variant="text" prepend-icon="mdi-close" @click="emit('update:modelValue', false)">Cancel</v-btn>
        <v-btn min-width="110" rounded="lg" :color="organization ? 'primary' : 'success'" variant="flat" @click="submit">{{ organization ? 'Save' : 'Create' }}</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
