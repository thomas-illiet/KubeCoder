<script setup lang="ts">
import { computed, ref } from 'vue'
import DataTableEmptyRow from '../components/DataTableEmptyRow.vue'
import FilterCard from '../components/FilterCard.vue'
import SectionCard from '../components/SectionCard.vue'
import StatusChip from '../components/StatusChip.vue'
import TablePaginationCard from '../components/TablePaginationCard.vue'
import { usePagination } from '../composables/usePagination'

const query = ref('')
const source = ref('All sources')

const skills = [
  { name: 'Repository review', version: '2.4.0', scope: 'Global', local: 'Inherit', effective: 'Enabled', source: 'Global' },
  { name: 'Northstar conventions', version: '1.3.0', scope: 'Organization', local: 'Enabled', effective: 'Enabled', source: 'Organization' },
  { name: 'Kubernetes operator', version: '3.1.0', scope: 'Organization', local: 'Enabled', effective: 'Enabled', source: 'Organization' },
  { name: 'Legacy deploy', version: '1.3.2', scope: 'Global', local: 'Disabled', effective: 'Disabled', source: 'Organization' },
]

const filteredSkills = computed(() => skills.filter((skill) => {
  const matchesSource = source.value === 'All sources' || skill.source === source.value
  const matchesQuery = skill.name.toLowerCase().includes(query.value.trim().toLowerCase())
  return matchesSource && matchesQuery
}))
const { page, paginatedItems: paginatedSkills, itemsPerPage } = usePagination(filteredSkills)
</script>

<template>
  <FilterCard title="Filters" subtitle="Filter skills by their effective configuration source" class="mb-4"><v-text-field v-model="query" hide-details placeholder="Search skills…" prepend-inner-icon="mdi-magnify" /><IconSelect v-model="source" hide-details label="Source" :items="['All sources', 'Global', 'Organization', 'Repository']" /></FilterCard>
  <SectionCard title="Skills" subtitle="Read-only in the user-facing projection"><div class="table-scroll"><table class="data-table"><thead><tr><th>SKILL</th><th class="table-cell--center">VERSION</th><th class="table-cell--center">SCOPE</th><th class="table-cell--center">LOCAL POLICY</th><th class="table-cell--center">EFFECTIVE STATUS</th><th class="table-cell--center">SOURCE</th></tr></thead><tbody><DataTableEmptyRow v-if="filteredSkills.length === 0" :colspan="6" /><tr v-for="skill in paginatedSkills" :key="skill.name"><td><div class="entity-cell"><div class="entity-icon"><v-icon icon="mdi-puzzle-outline" /></div><div class="entity-name">{{ skill.name }}</div></div></td><td class="code-text table-cell--center">{{ skill.version }}</td><td class="table-cell--center">{{ skill.scope }}</td><td class="table-cell--center"><v-chip size="small" variant="outlined">{{ skill.local }}</v-chip></td><td class="table-cell--center"><StatusChip :label="skill.effective" :color="skill.effective === 'Enabled' ? 'success' : 'default'" /></td><td class="text-medium-emphasis table-cell--center">{{ skill.source }}</td></tr></tbody></table></div></SectionCard>
  <TablePaginationCard v-model="page" :total="filteredSkills.length" :items-per-page="itemsPerPage" item-label="skills" hint="Explicit effective configuration" />
</template>
