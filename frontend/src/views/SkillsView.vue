<script setup lang="ts">
import { computed, ref } from 'vue'
import DataTableEmptyRow from '../components/DataTableEmptyRow.vue'
import FilterCard from '../components/FilterCard.vue'
import SectionCard from '../components/SectionCard.vue'
import StatusChip from '../components/StatusChip.vue'
import TablePaginationCard from '../components/TablePaginationCard.vue'
import { useNotifications } from '../composables/useNotifications'
import { usePagination } from '../composables/usePagination'

const scope = ref('organization')
const query = ref('')
const statusFilter = ref('All statuses')
const editorOpen = ref(false)
const detailOpen = ref(false)
type AdminPolicy = 'Enabled' | 'Disabled'
type SkillScope = 'global' | 'organization' | 'repository'
type Skill = { name: string; description: string; version: string; owner: string; scope: SkillScope; local: AdminPolicy; effective: AdminPolicy; source: string; locked: boolean; digest: string }
const selected = ref<Skill | null>(null)
const { success, info } = useNotifications()

const skills: Skill[] = [
  { name: 'Go reviewer', description: 'Go conventions, concurrency, and testing', version: '2.4.1', owner: 'Global', scope: 'global', local: 'Enabled', effective: 'Enabled', source: 'Global', locked: false, digest: 'sha256:7af2…c109' },
  { name: 'Security baseline', description: 'OWASP checks and secret leak prevention', version: '1.8.0', owner: 'Global', scope: 'organization', local: 'Enabled', effective: 'Enabled', source: 'Organization', locked: true, digest: 'sha256:9be1…4d82' },
  { name: 'Kubernetes operator', description: 'Manifests, Pod security, and policies', version: '3.1.0', owner: 'Northstar Labs', scope: 'organization', local: 'Enabled', effective: 'Enabled', source: 'Organization', locked: false, digest: 'sha256:21cd…ef09' },
  { name: 'Legacy deploy', description: 'Legacy delivery procedures', version: '1.3.2', owner: 'Northstar Labs', scope: 'repository', local: 'Disabled', effective: 'Disabled', source: 'Repository', locked: false, digest: 'sha256:77ac…013f' },
  { name: 'Documentation', description: 'Internal documentation style and structure', version: '2.0.0', owner: 'Northstar Labs', scope: 'organization', local: 'Enabled', effective: 'Enabled', source: 'Organization', locked: false, digest: 'sha256:f8e2…88aa' },
]

const filtered = computed(() => skills.filter((item) => {
  const matchesSearch = `${item.name} ${item.description}`.toLowerCase().includes(query.value.trim().toLowerCase())
  const matchesScope = item.scope === scope.value
  const matchesStatus = statusFilter.value === 'All statuses' || item.effective === statusFilter.value
  return matchesSearch && matchesScope && matchesStatus
}))
const { page, paginatedItems: paginatedSkills, itemsPerPage } = usePagination(filtered)

function inspect(item: Skill) { selected.value = item; detailOpen.value = true }
function updatePolicy(skill: Skill, policy: string) {
  if (policy !== 'Enabled' && policy !== 'Disabled') return
  skill.local = policy
  skill.effective = policy
  info('Skill policy updated', `${skill.name} is now ${policy.toLowerCase()} by the administrator.`)
}
function saveSkill(publish: boolean) { editorOpen.value = false; success(publish ? 'Skill published' : 'Skill saved', publish ? 'The new skill is available to the organization.' : 'The skill draft was saved.') }
</script>

<template>
  <FilterCard title="Filters" subtitle="Choose a scope, then refine the catalog" class="mb-4">
    <div class="filter-card__tabs"><v-btn-toggle v-model="scope" mandatory class="scope-tabs">
      <v-btn value="global" prepend-icon="mdi-earth">Global</v-btn>
      <v-btn value="organization" prepend-icon="mdi-domain">Organization</v-btn>
      <v-btn value="repository" prepend-icon="mdi-source-repository">Repository</v-btn>
    </v-btn-toggle></div>
    <v-text-field v-model="query" hide-details placeholder="Search skills…" prepend-inner-icon="mdi-magnify" />
    <IconSelect v-model="statusFilter" hide-details :items="['All statuses', 'Enabled', 'Disabled']" min-width="170" />
    <v-spacer />
    <div class="text-caption text-medium-emphasis"><v-icon icon="mdi-information-outline" size="16" class="mr-1" />Admin policies are explicit; organization users configure inheritance</div>
  </FilterCard>

  <SectionCard title="Skill policies" subtitle="Local activation and effective status">
    <template #actions><v-btn color="primary" prepend-icon="mdi-plus" @click="editorOpen = true">New skill</v-btn></template>
    <div class="table-scroll">
      <table class="data-table">
        <thead><tr><th>SKILL</th><th class="table-cell--center">OWNER</th><th class="table-cell--center">VERSION</th><th class="table-cell--center">POLICY</th><th class="table-cell--center">EFFECTIVE STATUS</th><th class="table-cell--center">SOURCE</th><th class="table-cell--center" aria-label="Actions"></th></tr></thead>
        <tbody><DataTableEmptyRow v-if="filtered.length === 0" :colspan="7" /><tr v-for="skill in paginatedSkills" :key="skill.name" @click="inspect(skill)" style="cursor:pointer">
          <td><div class="entity-cell"><div class="entity-icon"><v-icon icon="mdi-puzzle-outline" size="19" /></div><div><div class="entity-name">{{ skill.name }}</div><div class="entity-meta">{{ skill.description }}</div></div></div></td>
          <td class="table-cell--center"><v-chip size="small" variant="outlined">{{ skill.owner }}</v-chip></td><td class="code-text table-cell--center">{{ skill.version }}</td>
          <td class="table-cell--center" @click.stop><v-btn-toggle :model-value="skill.local" mandatory density="compact" class="scope-tabs" @update:model-value="updatePolicy(skill, $event)"><v-btn value="Enabled" size="x-small" color="success">Enable</v-btn><v-btn value="Disabled" size="x-small" color="error">Disable</v-btn></v-btn-toggle></td>
          <td class="table-cell--center"><StatusChip :label="skill.effective" :color="skill.effective === 'Enabled' ? 'success' : 'error'" :icon="skill.effective === 'Enabled' ? 'mdi-check-circle-outline' : 'mdi-minus-circle-outline'" /></td>
          <td class="table-cell--center"><span class="text-medium-emphasis"><v-icon :icon="skill.source === 'Global' ? 'mdi-earth' : skill.source === 'Repository' ? 'mdi-source-repository' : 'mdi-domain'" size="15" class="mr-1" />{{ skill.source }}</span><v-icon v-if="skill.locked" icon="mdi-lock-outline" size="14" color="warning" class="ml-2" /></td>
          <td class="table-cell--center"><v-btn icon="mdi-chevron-right" variant="text" size="small" :aria-label="`Inspect ${skill.name}`" @click.stop="inspect(skill)" /></td>
        </tr></tbody>
      </table>
    </div>
  </SectionCard>
  <TablePaginationCard v-model="page" :total="filtered.length" :items-per-page="itemsPerPage" item-label="skills" hint="Locked policies cannot be re-enabled at a lower level" />

  <v-navigation-drawer v-model="detailOpen" location="right" temporary width="480" class="detail-drawer">
    <div class="drawer-header d-flex justify-space-between"><div><div class="text-h6 font-weight-bold">{{ selected?.name }}</div><div class="text-caption text-medium-emphasis">Version {{ selected?.version }}</div></div><v-btn icon="mdi-close" variant="text" aria-label="Close" @click="detailOpen = false" /></div>
    <div class="form-section"><div class="form-section__title">Effective status</div><div class="d-flex align-center justify-space-between pa-3 rounded-lg" style="background:rgba(49,212,157,.05); border:1px solid rgba(49,212,157,.12)"><div><StatusChip :label="selected?.effective ?? ''" :color="selected?.effective === 'Enabled' ? 'success' : 'error'" :icon="selected?.effective === 'Enabled' ? 'mdi-check-circle-outline' : 'mdi-minus-circle-outline'" /><div class="text-caption text-medium-emphasis mt-2">Explicit administrator policy</div></div><v-icon :icon="selected?.effective === 'Enabled' ? 'mdi-check-decagram-outline' : 'mdi-cancel'" size="28" :color="selected?.effective === 'Enabled' ? 'success' : 'error'" /></div></div>
    <div class="form-section"><div class="form-section__title">Policy behavior</div><div class="review-grid"><div class="review-item"><span>Administrator</span><strong>{{ selected?.local }}</strong></div><div class="review-item"><span>Organization users</span><strong>Inherit or override</strong></div></div><div class="security-note mt-4"><v-icon icon="mdi-information-outline" color="info" /><span>Administrator policies are always explicit. Inheritance is configured only from the organization workspace.</span></div></div>
    <div class="form-section"><div class="form-section__title">Integrity</div><div class="text-caption text-medium-emphasis mb-1">Content digest</div><code class="text-caption">{{ selected?.digest }}</code><div class="security-note mt-4"><v-icon icon="mdi-check-decagram-outline" color="success" /><span>Content scanned; no sensitive data detected.</span></div></div>
  </v-navigation-drawer>

  <v-dialog v-model="editorOpen" max-width="640">
    <v-card class="section-card"><div class="pa-6"><div class="d-flex justify-space-between mb-5"><div><h3>New organization skill</h3><div class="text-caption text-medium-emphasis">Content will be versioned and scanned before publication</div></div><v-btn icon="mdi-close" variant="text" aria-label="Close" @click="editorOpen = false" /></div><div class="form-row"><v-text-field label="Name" placeholder="E.g. Frontend review" /><v-text-field label="Version" model-value="1.0.0" /></div><v-textarea label="Description" rows="2" /><v-textarea label="Markdown instructions" rows="7" placeholder="# Instructions&#10;&#10;Describe the conventions to follow…" class="code-text" /><v-switch label="Allow repositories to disable this skill" color="primary" hide-details /></div><v-card-actions class="dialog-actions"><v-spacer /><v-btn @click="editorOpen = false">Cancel</v-btn><v-btn variant="outlined" @click="saveSkill(false)">Save</v-btn><v-btn color="primary" @click="saveSkill(true)">Scan & publish</v-btn></v-card-actions></v-card>
  </v-dialog>
</template>
