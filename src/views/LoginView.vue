<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useNotifications } from '../composables/useNotifications'

const router = useRouter()
const { info } = useNotifications()
const email = ref('alex.martin@example.test')
const password = ref('')
const remember = ref(true)

function recoverPassword() {
  info('Password recovery', 'Password recovery is not connected in this visual mock. Use SSO to continue.')
}
</script>

<template>
  <v-card class="auth-card">
    <div class="auth-card__header">
      <div class="auth-icon"><v-icon icon="mdi-login-variant" size="24" /></div>
      <h1>Sign in</h1>
      <p>Access your KubeCoder workspace.</p>
    </div>
    <div class="auth-card__body">
      <v-text-field v-model="email" label="Email address" type="email" autocomplete="username" prepend-inner-icon="mdi-email-outline" />
      <v-text-field v-model="password" label="Password" type="password" autocomplete="current-password" prepend-inner-icon="mdi-lock-outline" />
      <div class="d-flex align-center justify-space-between mb-5">
        <v-checkbox v-model="remember" label="Keep me signed in" hide-details density="compact" />
        <v-btn variant="text" size="small" @click="recoverPassword">Forgot password?</v-btn>
      </div>
      <v-btn block color="primary" size="large" prepend-icon="mdi-login" @click="router.push('/organization')">Sign in</v-btn>
      <div class="auth-divider"><span>or</span></div>
      <v-btn block variant="outlined" size="large" prepend-icon="mdi-shield-account-outline" @click="router.push('/organization')">Continue with SSO</v-btn>
      <div class="security-note mt-5"><v-icon icon="mdi-information-outline" color="info" /><span>This is a visual mock. No credentials are submitted or stored.</span></div>
    </div>
  </v-card>
</template>
