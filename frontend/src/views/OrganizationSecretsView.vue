<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import AdminSecretsPanel from '../components/secrets/AdminSecretsPanel.vue'

const route = useRoute()
const organizationSlug = computed(() => String(route.params.organizationSlug ?? ''))
</script>

<template>
  <div class="security-note mb-4">
    <v-icon icon="mdi-shield-lock-outline" />
    <span><strong>Protected values.</strong> Organization secrets can be managed here. Platform secrets are visible but remain locked and cannot be overridden or changed.</span>
  </div>
  <AdminSecretsPanel :organization-slug="organizationSlug" />
</template>
