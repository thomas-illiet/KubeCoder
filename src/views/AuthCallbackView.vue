<script setup lang="ts">
import { onMounted, shallowRef } from 'vue'
import { useRouter } from 'vue-router'
import { useAuth } from '../composables/useAuth'

const router = useRouter()
const { completeLogin } = useAuth()
const error = shallowRef('')

onMounted(async () => {
  try {
    const returnTo = await completeLogin()
    await router.replace(returnTo)
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : 'The OpenID Connect callback could not be completed.'
  }
})
</script>

<template>
  <v-card class="auth-card auth-card--compact">
    <div class="auth-card__header">
      <div class="auth-icon"><v-progress-circular v-if="!error" indeterminate color="primary" size="26" /><v-icon v-else icon="mdi-alert-circle-outline" color="error" size="26" /></div>
      <h1>{{ error ? 'Sign-in failed' : 'Completing sign-in' }}</h1>
      <p>{{ error || 'Validating the OpenID Connect response…' }}</p>
    </div>
    <div v-if="error" class="auth-card__body">
      <v-btn block color="primary" to="/login">Back to sign in</v-btn>
    </div>
  </v-card>
</template>

