<script setup lang="ts">
defineProps<{ title: string; subtitle?: string; icon?: string }>()
</script>

<template>
  <v-card class="section-card">
    <div class="section-card__header">
      <div>
        <div class="section-card__title"><v-icon v-if="icon" :icon="icon" size="20" class="mr-2" />{{ title }}</div>
        <div v-if="subtitle" class="section-card__subtitle">{{ subtitle }}</div>
      </div>
      <slot name="actions" />
    </div>
    <slot />
  </v-card>
</template>
