<script setup lang="ts">
import { computed, onMounted, shallowRef } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuth } from '../composables/useAuth'
import { useNotifications } from '../composables/useNotifications'
import { normalizeReturnTo } from '../auth/navigation'

const route = useRoute()
const router = useRouter()
const { state, isAuthenticated, login } = useAuth()
const { info } = useNotifications()
const returnTo = computed(() => normalizeReturnTo(route.query.redirect))
const email = shallowRef('')
const password = shallowRef('')
const remember = shallowRef(true)

onMounted(async () => {
  if (isAuthenticated.value) await router.replace(returnTo.value)
})

function signInWithPassword() {
  password.value = ''
  info('Password authentication', 'The form is reserved for the future LDAP backend. No credentials were submitted or stored.')
}

function recoverPassword() {
  info('Password recovery', 'Password recovery will be available with the future LDAP backend.')
}

async function signInWithOidc() {
  await login(returnTo.value)
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
      <v-alert v-if="state.error" class="mb-4" type="error" variant="tonal" :text="state.error" />
      <v-form @submit.prevent="signInWithPassword">
        <v-text-field v-model="email" label="Email address" type="email" autocomplete="username" prepend-inner-icon="mdi-email-outline" />
        <v-text-field v-model="password" label="Password" type="password" autocomplete="current-password" prepend-inner-icon="mdi-lock-outline" />
        <div class="d-flex align-center justify-space-between mb-5">
          <v-checkbox v-model="remember" label="Keep me signed in" hide-details density="compact" />
          <v-btn variant="text" size="small" @click="recoverPassword">Forgot password?</v-btn>
        </div>
        <v-btn block color="primary" size="large" prepend-icon="mdi-login" type="submit">Sign in</v-btn>
      </v-form>
      <div class="auth-divider"><span>or</span></div>
      <v-btn block variant="outlined" size="large" prepend-icon="mdi-shield-account-outline" :loading="state.loading" @click="signInWithOidc">Continue with SSO</v-btn>
      <div class="security-note mt-5"><v-icon icon="mdi-lock-check-outline" color="success" /><span>Password authentication is reserved for the future LDAP backend. SSO uses OpenID Connect with PKCE.</span></div>
    </div>
  </v-card>
</template>
