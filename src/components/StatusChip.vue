<script setup lang="ts">
withDefaults(defineProps<{ label: string; color?: string; icon?: string }>(), {
  color: 'secondary',
  icon: 'mdi-circle-medium',
})
</script>

<template>
  <v-chip :color="color" :prepend-icon="icon" size="small" variant="tonal" class="font-weight-bold">
    {{ label }}
  </v-chip>
</template>
