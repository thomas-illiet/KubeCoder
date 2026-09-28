<script setup lang="ts">
defineProps<{ title?: string; subtitle?: string }>()
</script>

<template>
  <v-card class="filter-card section-card">
    <div v-if="title || subtitle" class="filter-card__heading">
      <div v-if="title" class="section-card__title">{{ title }}</div>
      <div v-if="subtitle" class="section-card__subtitle">{{ subtitle }}</div>
    </div>
    <div class="toolbar"><slot /></div>
  </v-card>
</template>
