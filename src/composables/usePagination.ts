import { computed, ref, toValue, watch, type MaybeRefOrGetter } from 'vue'

export function usePagination<T>(items: MaybeRefOrGetter<readonly T[]>, itemsPerPage = 3) {
  const page = ref(1)
  const total = computed(() => toValue(items).length)
  const pageCount = computed(() => Math.max(1, Math.ceil(total.value / itemsPerPage)))
  const paginatedItems = computed(() => {
    const start = (page.value - 1) * itemsPerPage
    return toValue(items).slice(start, start + itemsPerPage)
  })

  watch(() => toValue(items), () => {
    page.value = 1
  })

  watch(pageCount, (count) => {
    if (page.value > count) page.value = count
  })

  return { page, pageCount, paginatedItems, total, itemsPerPage }
}
