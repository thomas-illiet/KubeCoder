<script setup lang="ts">
import { computed, useAttrs } from 'vue'
import { selectPresentation } from '../shared/selectPresentation'

defineOptions({ inheritAttrs: false })

defineProps<{
  items: readonly unknown[]
}>()

const attrs = useAttrs()
const isMultiple = computed(() => attrs.multiple !== undefined && attrs.multiple !== false)
const usesChips = computed(() => attrs.chips !== undefined && attrs.chips !== false)
</script>

<template>
  <v-select v-bind="$attrs" :items="items" class="icon-select">
    <template #item="{ props, item }">
      <v-list-item v-bind="props" class="icon-select__item">
        <template #prepend>
          <span class="icon-select__badge">
            <v-icon :icon="selectPresentation(item.raw).icon" :color="selectPresentation(item.raw).color" size="18" />
          </span>
        </template>
      </v-list-item>
    </template>

    <template #selection="slotData">
      <slot name="selection" v-bind="slotData" :presentation="selectPresentation(slotData.item.raw)">
        <span class="icon-select__selection">
          <v-icon
            :icon="selectPresentation(slotData.item.raw).icon"
            :color="selectPresentation(slotData.item.raw).color"
            size="17"
          />
          <span>{{ slotData.item.title }}</span>
        </span>
      </slot>
    </template>

    <template v-if="isMultiple && usesChips" #chip="slotData">
      <slot name="selection" v-bind="slotData" :presentation="selectPresentation(slotData.item.raw)">
        <v-chip size="small" class="icon-select__chip">
          <v-icon
            :icon="selectPresentation(slotData.item.raw).icon"
            :color="selectPresentation(slotData.item.raw).color"
            size="15"
            start
          />
          {{ slotData.item.title }}
        </v-chip>
      </slot>
    </template>
  </v-select>
</template>
