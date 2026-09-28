<script setup lang="ts">
import AdminSecretsPanel from '../components/secrets/AdminSecretsPanel.vue'
</script>

<template>
  <div class="security-note mb-4"><v-icon icon="mdi-shield-lock-outline" /><span><strong>Write-only values.</strong> Secret values are accepted only during creation or replacement and are never returned.</span></div>
  <AdminSecretsPanel />
</template>
