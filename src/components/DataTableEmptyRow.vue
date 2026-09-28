<script setup lang="ts">
withDefaults(defineProps<{
  colspan: number
  title?: string
  description?: string
  icon?: string
}>(), {
  title: 'No items found',
  description: 'Adjust your filters or search to display results.',
  icon: 'mdi-database-search-outline',
})
</script>

<template>
  <tr class="data-table-empty-row">
    <td :colspan="colspan">
      <div class="data-table-empty-state">
        <div class="data-table-empty-state__icon"><v-icon :icon="icon" size="24" /></div>
        <div class="data-table-empty-state__title">{{ title }}</div>
        <div class="data-table-empty-state__description">{{ description }}</div>
      </div>
    </td>
  </tr>
</template>
