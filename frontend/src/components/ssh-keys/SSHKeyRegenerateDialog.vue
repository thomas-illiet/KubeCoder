<script setup lang="ts">
defineProps<{ saving: boolean }>()
const open = defineModel<boolean>({ required: true })
const emit = defineEmits<{ confirm: [] }>()
</script>

<template>
  <v-dialog v-model="open" max-width="560">
    <v-card class="app-dialog">
      <div class="dialog-header">
        <div class="d-flex align-center ga-3">
          <div class="danger-icon"><v-icon icon="mdi-alert-octagon-outline" /></div>
          <div>
            <div class="text-h6 font-weight-bold">Regenerate SSH key?</div>
            <div class="text-caption text-error font-weight-bold">Danger zone · irreversible action</div>
          </div>
        </div>
        <v-btn icon="mdi-close" variant="text" aria-label="Close" :disabled="saving" @click="open = false" />
      </div>
      <div class="pa-5">
        <v-alert type="error" variant="tonal" icon="mdi-alert-octagon-outline">
          <div class="text-subtitle-1 font-weight-bold mb-3">The current SSH key will stop working immediately.</div>
          <p class="mb-2">Regenerating the key will:</p>
          <ul class="danger-list mb-0">
            <li>Invalidate the current public key on every Git provider;</li>
            <li>Interrupt Git clones for all affected repositories;</li>
            <li>Require the new public key to be registered manually.</li>
          </ul>
        </v-alert>
      </div>
      <v-card-actions class="dialog-actions">
        <v-spacer />
        <v-btn variant="text" :disabled="saving" @click="open = false">Cancel</v-btn>
        <v-btn color="error" variant="flat" prepend-icon="mdi-alert-outline" :loading="saving" @click="emit('confirm')">Regenerate and invalidate</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<style scoped>
.danger-icon {
  display: grid;
  width: 2.5rem;
  height: 2.5rem;
  place-items: center;
  border-radius: 50%;
  background: rgba(var(--v-theme-error), 0.14);
  color: rgb(var(--v-theme-error));
}

.danger-list {
  padding-left: 1.25rem;
}

.danger-list li + li {
  margin-top: 0.35rem;
}
</style>
