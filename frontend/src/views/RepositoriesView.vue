<script setup lang="ts">
import { computed, ref } from 'vue'
import DataTableEmptyRow from '../components/DataTableEmptyRow.vue'
import FilterCard from '../components/FilterCard.vue'
import SectionCard from '../components/SectionCard.vue'
import StatusChip from '../components/StatusChip.vue'
import TablePaginationCard from '../components/TablePaginationCard.vue'
import { useNotifications } from '../composables/useNotifications'
import { usePagination } from '../composables/usePagination'

type Provider = 'GitHub' | 'GitLab' | 'Bitbucket'
type Repository = {
  name: string
  provider: Provider
  branch: string
  agent: string
  status: string
  checked: string
}

const providerOptions: Array<{ value: Provider; icon: string; color: string; description: string }> = [
  { value: 'GitHub', icon: 'mdi-github', color: '#aab7ff', description: 'Connect a GitHub repository' },
  { value: 'GitLab', icon: 'mdi-gitlab', color: '#e9ad5a', description: 'Connect a GitLab project' },
  { value: 'Bitbucket', icon: 'mdi-bitbucket', color: '#60a5fa', description: 'Connect a Bitbucket repository' },
]

const query = ref('')
const providerFilter = ref('All providers')
const statusFilter = ref('All statuses')
const { success } = useNotifications()
const repositories: Repository[] = [
  { name: 'identity-service', provider: 'GitHub', branch: 'main', agent: 'Atlas', status: 'Ready', checked: '12 min ago' },
  { name: 'platform-k8s', provider: 'GitLab', branch: 'trunk', agent: 'Pathfinder', status: 'Ready', checked: '1 hour ago' },
  { name: 'checkout-api', provider: 'GitHub', branch: 'main', agent: 'Sentinel', status: 'Ready', checked: 'Yesterday' },
  { name: 'legacy-portal', provider: 'Bitbucket', branch: 'develop', agent: 'Not configured', status: 'Action required', checked: '6 days ago' },
]
const filteredRepositories = computed(() => repositories.filter((repository) => {
  const search = query.value.trim().toLowerCase()
  const matchesSearch = !search || `${repository.name} ${repository.provider} ${repository.branch} ${repository.agent}`.toLowerCase().includes(search)
  const matchesProvider = providerFilter.value === 'All providers' || repository.provider === providerFilter.value
  const matchesStatus = statusFilter.value === 'All statuses' || repository.status === statusFilter.value
  return matchesSearch && matchesProvider && matchesStatus
}))
const { page, paginatedItems: paginatedRepositories, itemsPerPage } = usePagination(filteredRepositories)

const editorOpen = ref(false)
const editing = ref(false)
const step = ref(1)
const provider = ref<Provider>('GitHub')
const displayName = ref('')
const defaultBranch = ref('main')
const agent = ref('Atlas')
const includeSubmodules = ref(false)

const githubOwner = ref('northstar-labs')
const githubRepository = ref('')
const gitlabNamespace = ref('northstar-labs')
const gitlabProject = ref('')
const bitbucketOwner = ref('northstar-labs')
const bitbucketRepository = ref('')

const dialogTitle = computed(() => editing.value ? `Edit ${displayName.value}` : 'Add repository')
const providerTarget = computed(() => {
  if (provider.value === 'GitHub') return `${githubOwner.value}/${githubRepository.value || 'repository'}`
  if (provider.value === 'GitLab') return `${gitlabNamespace.value}/${gitlabProject.value || 'project'}`
  return `${bitbucketOwner.value}/${bitbucketRepository.value || 'repository'}`
})
const canContinue = computed(() => {
  if (step.value === 1) return Boolean(provider.value)
  if (step.value === 2) {
    if (provider.value === 'GitHub') return Boolean(githubOwner.value && githubRepository.value)
    if (provider.value === 'GitLab') return Boolean(gitlabNamespace.value && gitlabProject.value)
    return Boolean(bitbucketOwner.value && bitbucketRepository.value)
  }
  if (step.value === 3) return Boolean(displayName.value && defaultBranch.value && agent.value)
  return true
})

function selectProvider(value: Provider) {
  if (editing.value) return
  provider.value = value
}

function resetForm() {
  step.value = 1
  provider.value = 'GitHub'
  displayName.value = ''
  defaultBranch.value = 'main'
  agent.value = 'Atlas'
  includeSubmodules.value = false
  githubOwner.value = 'northstar-labs'
  githubRepository.value = ''
  gitlabNamespace.value = 'northstar-labs'
  gitlabProject.value = ''
  bitbucketOwner.value = 'northstar-labs'
  bitbucketRepository.value = ''
}

function openCreate() {
  editing.value = false
  resetForm()
  editorOpen.value = true
}

function openEdit(repository: Repository) {
  editing.value = true
  resetForm()
  provider.value = repository.provider
  displayName.value = repository.name
  defaultBranch.value = repository.branch
  agent.value = repository.agent === 'Not configured' ? 'None' : repository.agent

  if (repository.provider === 'GitHub') githubRepository.value = repository.name
  if (repository.provider === 'GitLab') gitlabProject.value = repository.name
  if (repository.provider === 'Bitbucket') bitbucketRepository.value = repository.name
  editorOpen.value = true
}

function useDetectedName() {
  if (provider.value === 'GitHub') displayName.value = githubRepository.value
  if (provider.value === 'GitLab') displayName.value = gitlabProject.value.split('/').at(-1) ?? ''
  if (provider.value === 'Bitbucket') displayName.value = bitbucketRepository.value
}

function closeEditor() {
  editorOpen.value = false
}

function continueEditor() {
  if (step.value < 4) step.value += 1
  else {
    const action = editing.value ? 'updated' : 'added'
    closeEditor()
    success(`Repository ${action}`, `${displayName.value} is ready to use.`)
  }
}
</script>

<template>
  <FilterCard title="Filters" subtitle="Search accessible repositories" class="mb-4">
    <v-text-field v-model="query" hide-details placeholder="Search repositories…" prepend-inner-icon="mdi-magnify" />
    <IconSelect v-model="providerFilter" hide-details :items="['All providers', 'GitHub', 'GitLab', 'Bitbucket']" />
    <IconSelect v-model="statusFilter" hide-details :items="['All statuses', 'Ready', 'Action required']" />
  </FilterCard>

  <SectionCard title="Repositories" subtitle="Mock repositories visible in Northstar Labs">
    <template #actions><v-btn color="primary" prepend-icon="mdi-plus" @click="openCreate">Add repository</v-btn></template>
    <div class="table-scroll">
      <table class="data-table">
        <thead><tr><th>REPOSITORY</th><th class="table-cell--center">PROVIDER</th><th class="table-cell--center">BRANCH</th><th class="table-cell--center">AGENT</th><th class="table-cell--center">LAST CHECK</th><th class="table-cell--center">STATUS</th><th class="table-cell--center" aria-label="Actions"></th></tr></thead>
        <tbody>
          <DataTableEmptyRow v-if="filteredRepositories.length === 0" :colspan="7" />
          <tr v-for="repository in paginatedRepositories" :key="repository.name">
            <td><div class="entity-cell"><div class="entity-icon"><v-icon icon="mdi-source-repository" /></div><div><div class="entity-name code-text">{{ repository.name }}</div><div class="entity-meta">Organization repository</div></div></div></td>
            <td class="table-cell--center">{{ repository.provider }}</td>
            <td class="table-cell--center"><v-chip size="small" variant="outlined">{{ repository.branch }}</v-chip></td>
            <td class="table-cell--center">{{ repository.agent }}</td>
            <td class="text-medium-emphasis table-cell--center">{{ repository.checked }}</td>
            <td class="table-cell--center"><StatusChip :label="repository.status" :color="repository.status === 'Ready' ? 'success' : 'warning'" :icon="repository.status === 'Ready' ? 'mdi-check-circle-outline' : 'mdi-alert-outline'" /></td>
            <td class="table-cell--center"><v-btn icon="mdi-pencil-outline" variant="text" size="small" aria-label="Edit repository" @click="openEdit(repository)" /></td>
          </tr>
        </tbody>
      </table>
    </div>
  </SectionCard>
  <TablePaginationCard v-model="page" :total="filteredRepositories.length" :items-per-page="itemsPerPage" item-label="repositories" hint="Sensitive configuration hidden" />

  <v-dialog v-model="editorOpen" max-width="920" persistent scrollable>
    <v-card class="app-dialog repository-dialog">
      <div class="dialog-header">
        <div><div class="text-h6 font-weight-bold">{{ dialogTitle }}</div><div class="text-caption text-medium-emphasis">Provider-aware repository configuration</div></div>
        <v-btn icon="mdi-close" variant="text" aria-label="Close" @click="closeEditor" />
      </div>

      <v-stepper v-model="step" class="repository-stepper" alt-labels>
        <v-stepper-header>
          <v-stepper-item :complete="step > 1" title="Provider" :value="1" />
          <v-divider />
          <v-stepper-item :complete="step > 2" title="Connection" :value="2" />
          <v-divider />
          <v-stepper-item :complete="step > 3" title="Repository" :value="3" />
          <v-divider />
          <v-stepper-item title="Review" :value="4" />
        </v-stepper-header>

        <v-stepper-window>
          <v-stepper-window-item :value="1">
            <div class="repository-step-content">
              <div class="step-heading"><div><h3>Choose a Git provider</h3><p>The next step adapts to the selected provider and deployment type.</p></div><v-chip v-if="editing" size="small" variant="tonal" color="info" prepend-icon="mdi-lock-outline">Provider locked</v-chip></div>
              <div class="provider-grid">
                <v-card
                  v-for="option in providerOptions"
                  :key="option.value"
                  :class="['provider-option', { 'provider-option--active': provider === option.value, 'provider-option--disabled': editing && provider !== option.value }]"
                  role="button"
                  :tabindex="editing && provider !== option.value ? -1 : 0"
                  :aria-pressed="provider === option.value"
                  @click="selectProvider(option.value)"
                  @keydown.enter="selectProvider(option.value)"
                  @keydown.space.prevent="selectProvider(option.value)"
                >
                  <div class="provider-option__icon"><v-icon :icon="option.icon" :color="option.color" size="29" /></div>
                  <div class="provider-option__copy"><strong>{{ option.value }}</strong><span>{{ option.description }}</span></div>
                  <v-icon :icon="provider === option.value ? 'mdi-check-circle' : 'mdi-circle-outline'" :color="provider === option.value ? 'success' : 'default'" />
                </v-card>
              </div>
              <div v-if="editing" class="security-note mt-5"><v-icon icon="mdi-information-outline" color="info" /><span>The provider cannot be changed after creation. Create another repository binding to migrate to a different provider.</span></div>
            </div>
          </v-stepper-window-item>

          <v-stepper-window-item :value="2">
            <div class="repository-step-content">
              <div class="step-heading"><div><h3>{{ provider }} connection</h3><p>Only the fields supported by this provider are displayed.</p></div><v-chip size="small" variant="outlined" :prepend-icon="providerOptions.find((item) => item.value === provider)?.icon">{{ provider }}</v-chip></div>

              <template v-if="provider === 'GitHub'">
                <div class="form-row"><v-text-field v-model="githubOwner" label="Owner or organization" placeholder="northstar-labs" /><v-text-field v-model="githubRepository" label="Repository name" placeholder="identity-service" /></div>
              </template>

              <template v-else-if="provider === 'GitLab'">
                <div class="form-row"><v-text-field v-model="gitlabNamespace" label="Namespace" placeholder="northstar-labs/platform" /><v-text-field v-model="gitlabProject" label="Project path" placeholder="platform-k8s" /></div>
              </template>

              <template v-else>
                <div class="form-row"><v-text-field v-model="bitbucketOwner" label="Workspace or project key" placeholder="northstar-labs" /><v-text-field v-model="bitbucketRepository" label="Repository slug" placeholder="legacy-portal" /></div>
              </template>

              <div class="security-note mt-2"><v-icon icon="mdi-server-network" color="secondary" /><span>The backend resolves the deployment type and instance endpoint configured for this organization.</span></div>

            </div>
          </v-stepper-window-item>

          <v-stepper-window-item :value="3">
            <div class="repository-step-content">
              <div class="step-heading"><div><h3>Repository defaults</h3><p>Configure the values used when a new coding session starts.</p></div><v-btn size="small" variant="text" prepend-icon="mdi-auto-fix" @click="useDetectedName">Use detected name</v-btn></div>
              <div class="form-row"><v-text-field v-model="displayName" label="Display name" placeholder="identity-service" /><v-text-field v-model="defaultBranch" label="Default branch" placeholder="main" /></div>
              <IconSelect v-model="agent" label="Default agent" :items="['None', 'Atlas', 'Pathfinder', 'Sentinel']" />
              <div class="repository-options">
                <v-switch v-model="includeSubmodules" color="primary" label="Initialize Git submodules" hide-details />
              </div>
            </div>
          </v-stepper-window-item>

          <v-stepper-window-item :value="4">
            <div class="repository-step-content">
              <div class="step-heading"><div><h3>Review configuration</h3><p>Confirm the mock configuration before saving the repository binding.</p></div><v-chip color="success" variant="tonal" prepend-icon="mdi-check-circle-outline">Ready to save</v-chip></div>
              <div class="review-grid">
                <div class="review-item"><span>Provider</span><strong><v-icon :icon="providerOptions.find((item) => item.value === provider)?.icon" size="18" class="mr-2" />{{ provider }}</strong></div>
                <div class="review-item"><span>Repository</span><strong class="code-text">{{ providerTarget }}</strong></div>
                <div class="review-item"><span>Display name</span><strong>{{ displayName }}</strong></div>
                <div class="review-item"><span>Default branch</span><strong class="code-text">{{ defaultBranch }}</strong></div>
                <div class="review-item"><span>Agent</span><strong>{{ agent }}</strong></div>
                <div class="review-item"><span>Git submodules</span><strong><v-icon :icon="includeSubmodules ? 'mdi-check-circle-outline' : 'mdi-minus-circle-outline'" :color="includeSubmodules ? 'success' : 'medium-emphasis'" size="18" class="mr-2" />{{ includeSubmodules ? 'Enabled' : 'Disabled' }}</strong></div>
              </div>
              <div class="security-note mt-5"><v-icon icon="mdi-test-tube" color="info" /><span>The production workflow will test provider access and detect the default branch before persisting this binding.</span></div>
            </div>
          </v-stepper-window-item>
        </v-stepper-window>
      </v-stepper>

      <v-card-actions class="dialog-actions">
        <v-btn variant="text" @click="closeEditor">Cancel</v-btn>
        <v-spacer />
        <v-btn v-if="step > 1" variant="outlined" prepend-icon="mdi-arrow-left" @click="step -= 1">Back</v-btn>
        <v-btn color="primary" :disabled="!canContinue" :append-icon="step < 4 ? 'mdi-arrow-right' : 'mdi-content-save-outline'" @click="continueEditor">{{ step < 4 ? 'Continue' : (editing ? 'Save changes' : 'Add repository') }}</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
