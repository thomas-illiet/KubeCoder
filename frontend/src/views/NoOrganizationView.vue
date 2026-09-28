<script setup lang="ts">
import { useAuth } from '../composables/useAuth'

const { isAdmin } = useAuth()
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
        <v-btn v-if="isAdmin" color="primary" prepend-icon="mdi-shield-crown-outline" to="/admin/organizations">
          Go to administration
        </v-btn>
        <v-btn class="ml-2" variant="text" prepend-icon="mdi-logout" to="/logout">Sign out</v-btn>
      </v-card>
    </v-main>
  </v-app>
</template>
