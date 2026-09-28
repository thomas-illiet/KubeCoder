<script setup lang="ts">
import SectionCard from '../components/SectionCard.vue'
import { useNotifications } from '../composables/useNotifications'

const { success } = useNotifications()
</script>

<template>
  <div class="two-col-grid">
    <SectionCard title="Organization context" subtitle="Information available to your role" icon="mdi-domain">
      <div class="pa-5"><v-text-field label="Name" model-value="Northstar Labs" readonly /><v-text-field label="Identifier" model-value="northstar-labs" readonly class="code-text" /><v-text-field label="Your role" model-value="Owner" readonly /><div class="security-note"><v-icon icon="mdi-information-outline" color="info" /><span>Sensitive settings and runtime policies remain in the Administration section.</span></div></div>
    </SectionCard>
    <SectionCard title="Personal preferences" subtitle="Apply only to your account" icon="mdi-tune-variant">
      <div class="pa-5"><IconSelect label="Default repository" :items="['None', 'identity-service', 'platform-k8s', 'checkout-api']" model-value="identity-service" /><IconSelect label="Preferred agent" :items="['None', 'Atlas', 'Pathfinder', 'Sentinel']" model-value="Atlas" /><v-switch label="Receive run completion notifications" color="primary" model-value hide-details /><v-switch label="Keep the last work context" color="primary" model-value hide-details class="mt-3" /></div>
      <v-card-actions class="dialog-actions"><v-spacer /><v-btn color="primary" @click="success('Preferences updated', 'Your organization workspace preferences were saved.')">Save preferences</v-btn></v-card-actions>
    </SectionCard>
  </div>
</template>
