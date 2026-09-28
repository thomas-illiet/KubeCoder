<script setup lang="ts">
import { onMounted, shallowRef } from 'vue'
import { useAuth } from '../composables/useAuth'

const { logout } = useAuth()
const error = shallowRef('')

onMounted(async () => {
  try {
    await logout()
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : 'The OpenID Connect logout could not be started.'
  }
})
</script>

<template>
  <v-card class="auth-card auth-card--compact">
    <div class="auth-card__header">
      <div class="auth-icon"><v-progress-circular v-if="!error" indeterminate color="primary" size="26" /><v-icon v-else icon="mdi-alert-circle-outline" color="error" size="26" /></div>
      <h1>{{ error ? 'Sign-out failed' : 'Signing out' }}</h1>
      <p>{{ error || 'Redirecting to the OpenID Connect provider…' }}</p>
    </div>
    <div v-if="error" class="auth-card__body">
      <v-btn block color="primary" size="large" prepend-icon="mdi-login" to="/login">Back to sign in</v-btn>
    </div>
  </v-card>
</template>
