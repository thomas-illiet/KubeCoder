<script setup lang="ts">
defineProps<{ message: string }>()

function reload() {
  window.location.reload()
}
</script>

<template>
  <v-app class="auth-layout">
    <v-main>
      <div class="auth-shell">
        <div class="auth-brand">
          <div class="brand__mark"><v-icon icon="mdi-kubernetes" size="26" /></div>
          <div class="brand__copy"><span class="brand__name">KubeCoder</span><span class="brand__tag">CONFIGURATION ERROR</span></div>
        </div>
        <v-card class="auth-card auth-card--compact">
          <div class="auth-card__header">
            <div class="auth-icon"><v-icon icon="mdi-alert-circle-outline" color="error" size="26" /></div>
            <h1>Unable to start</h1>
            <p>{{ message }}</p>
          </div>
          <div class="auth-card__body">
            <v-btn block color="primary" prepend-icon="mdi-refresh" @click="reload">Reload configuration</v-btn>
          </div>
        </v-card>
      </div>
    </v-main>
  </v-app>
</template>
