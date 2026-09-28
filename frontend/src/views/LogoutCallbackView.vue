<script setup lang="ts">
import { onMounted, shallowRef } from 'vue'
import { useAuth } from '../composables/useAuth'

const { completeLogout } = useAuth()
const complete = shallowRef(false)
const error = shallowRef('')

onMounted(async () => {
  try {
    await completeLogout()
    complete.value = true
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : 'The OpenID Connect logout could not be completed.'
  }
})
</script>

<template>
  <v-card class="auth-card auth-card--compact">
    <div class="auth-card__header">
      <div class="auth-icon" :class="{ 'auth-icon--success': complete }">
        <v-progress-circular v-if="!complete && !error" indeterminate color="primary" size="26" />
        <v-icon v-else :icon="error ? 'mdi-alert-circle-outline' : 'mdi-check'" :color="error ? 'error' : undefined" size="26" />
      </div>
      <h1>{{ error ? 'Sign-out failed' : complete ? 'Signed out' : 'Signing out' }}</h1>
      <p>{{ error || (complete ? 'Your KubeCoder session has ended.' : 'Closing the OpenID Connect session…') }}</p>
    </div>
    <div v-if="complete || error" class="auth-card__body">
      <v-btn block color="primary" to="/login">Sign in again</v-btn>
    </div>
  </v-card>
</template>

