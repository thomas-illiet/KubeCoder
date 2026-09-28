<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'
import type { OrganizationSSHKey } from '../../api/organizationSSHKeys'
import SectionCard from '../SectionCard.vue'
import { copyPublicKey } from './clipboard'
import SSHKeyRegenerateDialog from './SSHKeyRegenerateDialog.vue'

const props = defineProps<{ sshKey: OrganizationSSHKey | null; loading: boolean; saving: boolean }>()
const emit = defineEmits<{ regenerate: []; copied: []; copyError: [error: unknown] }>()
const dialogOpen = shallowRef(false)

const rotatedAt = computed(() => props.sshKey ? new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(props.sshKey.updated_at)) : '')
watch(() => props.sshKey?.fingerprint, (_current, previous) => {
  if (_current && _current !== previous) dialogOpen.value = false
})

async function copy(): Promise<void> {
  if (!props.sshKey) return
  try {
    await copyPublicKey(props.sshKey.public_key)
    emit('copied')
  } catch (error) {
    emit('copyError', error)
  }
}
</script>

<template>
  <div class="security-note mb-4">
    <v-icon icon="mdi-shield-key-outline" size="20" color="secondary" />
    <span><strong class="text-high-emphasis">Private key protected.</strong> It is encrypted in the database and is never returned by the API or displayed in KubeCoder.</span>
  </div>

  <SectionCard title="Git SSH key" subtitle="Organization identity for future SSH Git clones" icon="mdi-key-chain-variant">
    <template #actions>
      <v-btn color="primary" variant="outlined" prepend-icon="mdi-refresh" :disabled="loading" @click="dialogOpen = true">Regenerate</v-btn>
    </template>

    <div v-if="loading" class="d-flex justify-center pa-10"><v-progress-circular indeterminate color="primary" /></div>
    <div v-else-if="sshKey" class="pa-5">
      <v-alert class="mb-5" type="info" variant="tonal" icon="mdi-source-repository">
        <div class="font-weight-bold mb-1">What this key is for</div>
        <div>
          This SSH identity will allow KubeCoder workers to authenticate when cloning private Git repositories for this organization. It does not grant access by itself: you must first register the public key with your Git provider.
        </div>
      </v-alert>

      <div class="key-guide mb-6">
        <div class="key-guide__step">
          <div class="key-guide__number">1</div>
          <div>
            <div class="font-weight-bold">Copy the public key</div>
            <div class="text-body-2 text-medium-emphasis">The public part is safe to share and is the only key material displayed by KubeCoder.</div>
          </div>
        </div>
        <div class="key-guide__step">
          <div class="key-guide__number">2</div>
          <div>
            <div class="font-weight-bold">Register it with your Git provider</div>
            <div class="text-body-2 text-medium-emphasis">Add it as a read-only deploy key or machine identity for every repository that KubeCoder needs to clone.</div>
          </div>
        </div>
        <div class="key-guide__step">
          <div class="key-guide__number">3</div>
          <div>
            <div class="font-weight-bold">KubeCoder uses the private key</div>
            <div class="text-body-2 text-medium-emphasis">During a future clone, the backend will decrypt the matching private key for the worker. The private key is never returned by the API or shown in this page.</div>
          </div>
        </div>
      </div>

      <v-textarea class="public-key" label="Public key" :model-value="sshKey.public_key.trim()" rows="3" readonly hide-details />
      <div class="d-flex flex-wrap ga-3 mt-4">
        <v-chip prepend-icon="mdi-fingerprint" variant="outlined" class="code-text">{{ sshKey.fingerprint }}</v-chip>
        <v-chip prepend-icon="mdi-lock-outline" variant="outlined">{{ sshKey.algorithm }}</v-chip>
        <v-chip prepend-icon="mdi-clock-outline" variant="outlined">Generated {{ rotatedAt }}</v-chip>
      </div>
      <div class="d-flex justify-end mt-5"><v-btn color="primary" prepend-icon="mdi-content-copy" @click="copy">Copy public key</v-btn></div>

    </div>
    <div v-else class="empty-state pa-10">
      <div class="empty-state__icon"><v-icon icon="mdi-key-alert-outline" /></div>
      <h3>SSH key unavailable</h3>
      <p>The organization SSH key could not be loaded. New organizations receive their key automatically when they are created.</p>
    </div>
  </SectionCard>

  <SSHKeyRegenerateDialog v-model="dialogOpen" :saving="saving" @confirm="emit('regenerate')" />
</template>

<style scoped>
.key-guide {
  display: grid;
  gap: 1rem;
}

.key-guide__step {
  display: grid;
  grid-template-columns: 2rem minmax(0, 1fr);
  gap: 0.75rem;
  align-items: start;
}

.key-guide__number {
  display: grid;
  width: 2rem;
  height: 2rem;
  place-items: center;
  border: 1px solid rgb(var(--v-theme-primary));
  border-radius: 50%;
  color: rgb(var(--v-theme-primary));
  font-weight: 700;
}

.public-key :deep(textarea) {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.82rem;
  line-height: 1.5;
}
</style>
