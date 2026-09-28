<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()
</script>

<template>
  <v-card class="auth-card auth-card--compact">
    <div class="auth-card__header">
      <div class="auth-icon auth-icon--success"><v-icon icon="mdi-check" size="26" /></div>
      <h1>Signed out</h1>
      <p>Your KubeCoder session has ended.</p>
    </div>
    <div class="auth-card__body">
      <v-btn block color="primary" size="large" prepend-icon="mdi-login" @click="router.push('/login')">Sign in again</v-btn>
      <v-btn block variant="text" class="mt-2" prepend-icon="mdi-arrow-left" @click="router.push('/organization')">Return to workspace</v-btn>
    </div>
  </v-card>
</template>
