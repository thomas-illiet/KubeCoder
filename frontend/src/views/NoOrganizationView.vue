<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuth } from '../composables/useAuth'
import { useOrganizations } from '../composables/useOrganizations'

const { isAdmin } = useAuth()
const organizations = useOrganizations()
const router = useRouter()
const refreshing = ref(false)
const refreshError = ref<string | null>(null)

async function refreshOrganizations(): Promise<void> {
  refreshing.value = true
  refreshError.value = null
  try {
    await organizations.refresh()
    if (organizations.state.current) {
      await router.replace(`/organizations/${organizations.state.current.slug}`)
    }
  } catch (error) {
    refreshError.value = error instanceof Error ? error.message : 'Organizations could not be refreshed.'
  } finally {
    refreshing.value = false
  }
}
</script>

<template>
  <v-app>
    <v-main class="d-flex align-center justify-center pa-6">
      <v-card class="section-card pa-8 text-center" max-width="620">
        <div class="empty-state__icon mx-auto mb-4"><v-icon icon="mdi-domain-off" size="32" /></div>
        <h1 class="text-h4 font-weight-bold mb-3">No organization assigned</h1>
        <p class="text-body-1 text-medium-emphasis mb-6">
          Your account is active, but it is not a member of an organization yet.
          Contact a platform administrator to request access.
        </p>
        <v-alert v-if="refreshError" class="mb-4 text-left" type="error" variant="tonal">{{ refreshError }}</v-alert>
        <v-btn v-if="isAdmin" color="primary" prepend-icon="mdi-shield-crown-outline" to="/admin/organizations">
          Go to administration
        </v-btn>
        <v-btn
          :class="{ 'ml-2': isAdmin }"
          :loading="refreshing"
          prepend-icon="mdi-refresh"
          variant="outlined"
          @click="refreshOrganizations"
        >
          Refresh access
        </v-btn>
        <v-btn class="ml-2" variant="text" prepend-icon="mdi-logout" to="/logout">Sign out</v-btn>
      </v-card>
    </v-main>
  </v-app>
</template>
