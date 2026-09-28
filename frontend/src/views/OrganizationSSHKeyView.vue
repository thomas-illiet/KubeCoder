<script setup lang="ts">
import { onBeforeUnmount, watch } from 'vue'
import { useRoute } from 'vue-router'
import OrganizationSSHKeyPanel from '../components/ssh-keys/OrganizationSSHKeyPanel.vue'
import { useNotifications } from '../composables/useNotifications'
import { useOrganizationSSHKey } from '../composables/useOrganizationSSHKey'

const route = useRoute()
const sshKey = useOrganizationSSHKey()
const notifications = useNotifications()

function slug(): string {
  return typeof route.params.organizationSlug === 'string' ? route.params.organizationSlug : ''
}

async function load(): Promise<void> {
  if (!slug()) return
  try { await sshKey.load(slug()) }
  catch (error) { notifications.error('SSH key could not be loaded', error instanceof Error ? error.message : undefined) }
}

async function regenerate(): Promise<void> {
  try {
    await sshKey.regenerate(slug())
    notifications.success('SSH key generated', 'The new public key is ready to register with your Git provider.')
  } catch (error) {
    notifications.error('SSH key could not be generated', error instanceof Error ? error.message : undefined)
  }
}

function copyError(error: unknown): void {
  notifications.error('Public key could not be copied', error instanceof Error ? error.message : undefined)
}

watch(() => route.params.organizationSlug, () => { sshKey.reset(); void load() }, { immediate: true })
onBeforeUnmount(sshKey.reset)
</script>

<template>
  <OrganizationSSHKeyPanel :ssh-key="sshKey.key.value" :loading="sshKey.loading.value" :saving="sshKey.saving.value" @regenerate="regenerate" @copied="notifications.success('Public key copied')" @copy-error="copyError" />
</template>
