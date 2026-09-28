<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  modelValue: number
  total: number
  itemsPerPage?: number
  itemLabel?: string
  hint?: string
}>(), {
  itemsPerPage: 3,
  itemLabel: 'items',
  hint: '',
})

const emit = defineEmits<{ 'update:modelValue': [value: number] }>()

const pageCount = computed(() => Math.max(1, Math.ceil(props.total / props.itemsPerPage)))
const firstItem = computed(() => props.total === 0 ? 0 : (props.modelValue - 1) * props.itemsPerPage + 1)
const lastItem = computed(() => Math.min(props.modelValue * props.itemsPerPage, props.total))
</script>

<template>
  <v-card class="table-pagination-card" elevation="0" rounded="lg">
    <div class="table-pagination-card__summary" aria-live="polite">
      <div class="table-pagination-card__icon"><v-icon icon="mdi-table-arrow-right" size="18" /></div>
      <div>
        <div class="table-pagination-card__title">Pagination</div>
        <div class="table-pagination-card__range"><strong>{{ firstItem }}–{{ lastItem }}</strong><span>of {{ total }} {{ itemLabel }}</span></div>
      </div>
      <span v-if="hint" class="table-pagination-card__hint">{{ hint }}</span>
    </div>
    <v-pagination
      :model-value="modelValue"
      :length="pageCount"
      :total-visible="5"
      density="compact"
      size="small"
      aria-label="Table pagination"
      @update:model-value="emit('update:modelValue', $event)"
    />
  </v-card>
</template>
