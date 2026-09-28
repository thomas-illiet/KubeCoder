<script setup lang="ts">
import { ref } from 'vue'
import SectionCard from '../components/SectionCard.vue'
import StatusChip from '../components/StatusChip.vue'
import { useAuth } from '../composables/useAuth'
import { useNotifications } from '../composables/useNotifications'

const { displayName: oidcDisplayName, email, username, initials, subject } = useAuth()
const displayName = ref(oidcDisplayName.value)
const jobTitle = ref('Platform administrator')
const timezone = ref('Europe/Paris')
const { success } = useNotifications()
</script>

<template>
  <div class="profile-grid">
    <SectionCard title="Account" subtitle="Application profile synchronized from your authenticated identity">
      <div class="profile-summary">
        <div class="profile-avatar">{{ initials }}</div>
        <div><h2>{{ oidcDisplayName }}</h2><p>{{ email || username }}</p><StatusChip label="Active account" color="success" icon="mdi-check-circle-outline" /></div>
      </div>
      <v-divider />
      <div class="pa-5">
        <div class="profile-detail"><span>Authentication</span><strong>OpenID Connect</strong></div>
        <div class="profile-detail"><span>Account ID</span><code>{{ subject }}</code></div>
        <div class="profile-detail"><span>Username</span><strong>{{ username }}</strong></div>
      </div>
    </SectionCard>

    <SectionCard title="Personal information" subtitle="Basic preferences used across workspaces">
      <div class="pa-5">
        <v-text-field v-model="displayName" label="Display name" />
        <v-text-field label="Email address" :model-value="email" readonly hint="Managed by your identity provider" persistent-hint />
        <v-text-field v-model="jobTitle" label="Job title" />
        <IconSelect v-model="timezone" label="Timezone" :items="['Europe/Paris', 'Europe/London', 'America/New_York', 'Asia/Tokyo']" />
      </div>
      <v-card-actions class="dialog-actions"><v-spacer /><v-btn color="primary" @click="success('Profile updated', 'Your personal information was saved.')">Save profile</v-btn></v-card-actions>
    </SectionCard>

    <SectionCard title="Session security" subtitle="Current browser session">
      <div class="pa-5"><div class="security-note"><v-icon icon="mdi-shield-check-outline" color="success" /><span>Your identity and access are managed through standard OpenID Connect claims.</span></div><v-btn class="mt-4" color="error" variant="outlined" prepend-icon="mdi-logout" to="/logout">Sign out</v-btn></div>
    </SectionCard>
  </div>
</template>
