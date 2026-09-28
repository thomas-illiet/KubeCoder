<script setup lang="ts">
import { onMounted, shallowRef } from 'vue'
import { useRouter } from 'vue-router'
import { useAuth } from '../composables/useAuth'

const router = useRouter()
const { completeLogin } = useAuth()
const redirecting = shallowRef(false)

onMounted(async () => {
  try {
    const returnTo = await completeLogin()
    await router.replace(returnTo)
  } catch {
    redirecting.value = true
    await router.replace('/login')
  }
})
</script>

<template>
  <v-card class="auth-card auth-card--compact">
    <div class="auth-card__header">
      <div class="auth-icon"><v-progress-circular indeterminate color="primary" size="26" /></div>
      <h1>Completing sign-in</h1>
      <p>{{ redirecting ? 'Returning to sign in…' : 'Validating the OpenID Connect response…' }}</p>
    </div>
  </v-card>
</template>
