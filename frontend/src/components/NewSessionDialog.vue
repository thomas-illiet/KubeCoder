<script setup lang="ts">
import { computed, ref, watch } from 'vue'

type RepositoryId = 'identity-service' | 'platform-k8s' | 'checkout-api'
type ProfileName = 'Atlas' | 'Pathfinder' | 'Sentinel'

export type CreatedSession = {
  repository: RepositoryId
  branch: string
  profile: ProfileName
  firstMessage: string
}

type RepositoryOption = {
  title: string
  value: RepositoryId
  icon: string
  color: string
  defaultBranch: string
  defaultProfile: ProfileName
}

type ProfileOption = {
  title: ProfileName
  value: ProfileName
  icon: string
  color: string
  description: string
  runtime: string
  resources: { cpu: string; memory: string; storage: string; timeout: string }
  skill: string
}

const open = defineModel<boolean>({ default: false })
const emit = defineEmits<{ created: [session: CreatedSession] }>()

const repositories: RepositoryOption[] = [
  { title: 'identity-service', value: 'identity-service', icon: 'mdi-source-repository', color: 'primary', defaultBranch: 'main', defaultProfile: 'Atlas' },
  { title: 'platform-k8s', value: 'platform-k8s', icon: 'mdi-kubernetes', color: 'info', defaultBranch: 'trunk', defaultProfile: 'Pathfinder' },
  { title: 'checkout-api', value: 'checkout-api', icon: 'mdi-source-repository', color: 'secondary', defaultBranch: 'main', defaultProfile: 'Sentinel' },
]

const profiles: ProfileOption[] = [
  { title: 'Atlas', value: 'Atlas', icon: 'mdi-compass-rose', color: 'primary', description: 'General implementation and repository maintenance', runtime: 'OpenCode 1.2', resources: { cpu: '2 vCPU', memory: '4 GiB', storage: '10 GiB', timeout: '45 min' }, skill: 'Repository implementation' },
  { title: 'Pathfinder', value: 'Pathfinder', icon: 'mdi-compass-outline', color: 'secondary', description: 'Infrastructure, manifests, and delivery workflows', runtime: 'OpenCode 1.2', resources: { cpu: '2 vCPU', memory: '4 GiB', storage: '12 GiB', timeout: '45 min' }, skill: 'Kubernetes operations' },
  { title: 'Sentinel', value: 'Sentinel', icon: 'mdi-shield-search-outline', color: 'info', description: 'Focused code review and security analysis', runtime: 'OpenCode 1.2', resources: { cpu: '1 vCPU', memory: '3 GiB', storage: '8 GiB', timeout: '30 min' }, skill: 'Security review' },
]

const step = ref(1)
const repositoryId = ref<RepositoryId>('identity-service')
const profileName = ref<ProfileName>('Atlas')
const revision = ref('main')
const firstMessage = ref('')

const repository = computed(() => repositories.find((item) => item.value === repositoryId.value) ?? repositories[0])
const profile = computed(() => profiles.find((item) => item.value === profileName.value) ?? profiles[0])
const effectiveSkills = computed(() => [
  { name: profile.value.skill, source: 'Profile', icon: profile.value.icon },
  { name: 'Organization conventions', source: 'Organization', icon: 'mdi-domain' },
  { name: repositoryId.value === 'platform-k8s' ? 'Manifest validation' : 'Repository conventions', source: 'Repository', icon: repositoryId.value === 'platform-k8s' ? 'mdi-kubernetes' : 'mdi-source-repository' },
])
const canContinue = computed(() => {
  if (step.value === 1) return Boolean(repositoryId.value && revision.value.trim() && firstMessage.value.trim())
  if (step.value === 2) return Boolean(profileName.value)
  return true
})

watch(open, (isOpen) => {
  if (!isOpen) return
  step.value = 1
  repositoryId.value = 'identity-service'
  profileName.value = 'Atlas'
  revision.value = 'main'
  firstMessage.value = ''
})

watch(repositoryId, () => {
  revision.value = repository.value.defaultBranch
  profileName.value = repository.value.defaultProfile
})

function close() {
  open.value = false
}

function continueWizard() {
  if (!canContinue.value) return
  if (step.value < 4) {
    step.value += 1
    return
  }

  emit('created', {
    repository: repositoryId.value,
    branch: revision.value.trim(),
    profile: profileName.value,
    firstMessage: firstMessage.value.trim(),
  })
  close()
}
</script>

<template>
  <v-dialog v-model="open" max-width="920" persistent scrollable>
    <v-card class="app-dialog repository-dialog session-create-dialog">
      <div class="dialog-header">
        <div>
          <div class="text-h6 font-weight-bold">New session</div>
          <div class="text-caption text-medium-emphasis">Prepare the workspace and verify the effective run configuration</div>
        </div>
        <v-btn icon="mdi-close" variant="text" aria-label="Close new session dialog" @click="close" />
      </div>

      <v-stepper v-model="step" class="repository-stepper session-stepper" alt-labels>
        <v-stepper-header>
          <v-stepper-item :complete="step > 1" title="Context" :value="1" />
          <v-divider />
          <v-stepper-item :complete="step > 2" title="Profile" :value="2" />
          <v-divider />
          <v-stepper-item :complete="step > 3" title="Configuration" :value="3" />
          <v-divider />
          <v-stepper-item title="Review" :value="4" />
        </v-stepper-header>

        <v-stepper-window>
          <v-stepper-window-item :value="1">
            <div class="repository-step-content session-step-content">
              <div class="step-heading">
                <div><h3>Define the work context</h3><p>Select the repository and starting revision, then describe the first task.</p></div>
                <v-chip size="small" variant="tonal" color="success" prepend-icon="mdi-check-decagram-outline">Repository ready</v-chip>
              </div>
              <div class="form-row">
                <IconSelect v-model="repositoryId" label="Repository" :items="repositories" />
                <v-text-field v-model="revision" label="Branch or revision" placeholder="main" prepend-inner-icon="mdi-source-branch" />
              </div>
              <v-textarea v-model="firstMessage" label="First message" rows="4" counter="1000" maxlength="1000" placeholder="Describe what you want the agent to do in this repository…" />
              <div class="security-note"><v-icon icon="mdi-lock-outline" color="secondary" /><span>The session receives an isolated persistent workspace. Credentials and private endpoints are resolved by the backend and are never exposed here.</span></div>
            </div>
          </v-stepper-window-item>

          <v-stepper-window-item :value="2">
            <div class="repository-step-content session-step-content">
              <div class="step-heading"><div><h3>Choose a published profile</h3><p>Only profiles available for {{ repository.title }} are shown.</p></div><v-chip size="small" variant="outlined" prepend-icon="mdi-source-repository">{{ repository.title }}</v-chip></div>
              <div class="session-profile-grid">
                <v-card
                  v-for="option in profiles"
                  :key="option.value"
                  :class="['session-profile-option', { 'session-profile-option--active': profileName === option.value }]"
                  role="button"
                  tabindex="0"
                  :aria-pressed="profileName === option.value"
                  @click="profileName = option.value"
                  @keydown.enter="profileName = option.value"
                  @keydown.space.prevent="profileName = option.value"
                >
                  <div class="provider-option__icon"><v-icon :icon="option.icon" :color="option.color" size="25" /></div>
                  <div class="provider-option__copy"><strong>{{ option.title }}</strong><span>{{ option.description }}</span><small>{{ option.runtime }}</small></div>
                  <v-icon :icon="profileName === option.value ? 'mdi-check-circle' : 'mdi-circle-outline'" :color="profileName === option.value ? 'success' : 'medium-emphasis'" />
                </v-card>
              </div>
              <div class="security-note mt-5"><v-icon icon="mdi-eye-off-outline" color="info" /><span>Model, provider, system prompt, image, tools, and network rules remain managed by administrators.</span></div>
            </div>
          </v-stepper-window-item>

          <v-stepper-window-item :value="3">
            <div class="repository-step-content session-step-content">
              <div class="step-heading"><div><h3>Effective configuration</h3><p>This read-only snapshot will be validated again before the run starts.</p></div><v-chip size="small" color="info" variant="tonal" prepend-icon="mdi-camera-outline">Snapshot preview</v-chip></div>
              <div class="session-configuration-grid">
                <section class="session-config-panel">
                  <div class="session-config-panel__title"><v-icon icon="mdi-puzzle-outline" size="19" color="primary" />Effective skills</div>
                  <div v-for="skill in effectiveSkills" :key="skill.name" class="session-skill-row">
                    <div><v-icon :icon="skill.icon" size="17" /><span>{{ skill.name }}</span></div>
                    <v-chip size="x-small" variant="outlined">{{ skill.source }}</v-chip>
                  </div>
                </section>
                <section class="session-config-panel">
                  <div class="session-config-panel__title"><v-icon icon="mdi-gauge" size="19" color="secondary" />Run resources</div>
                  <div class="session-resource-grid">
                    <div><span>CPU</span><strong>{{ profile.resources.cpu }}</strong></div>
                    <div><span>Memory</span><strong>{{ profile.resources.memory }}</strong></div>
                    <div><span>Workspace</span><strong>{{ profile.resources.storage }}</strong></div>
                    <div><span>Timeout</span><strong>{{ profile.resources.timeout }}</strong></div>
                  </div>
                </section>
              </div>
              <div class="session-runtime-line"><span>Runtime</span><strong><v-icon icon="mdi-application-cog-outline" size="18" class="mr-2" />{{ profile.runtime }}</strong></div>
            </div>
          </v-stepper-window-item>

          <v-stepper-window-item :value="4">
            <div class="repository-step-content session-step-content">
              <div class="step-heading"><div><h3>Review before starting</h3><p>Confirm the context and the checks used to create the first queued run.</p></div><v-chip color="success" variant="tonal" prepend-icon="mdi-check-circle-outline">Ready to start</v-chip></div>
              <div class="review-grid session-review-grid">
                <div class="review-item"><span>Repository</span><strong><v-icon icon="mdi-source-repository" size="18" class="mr-2" />{{ repository.title }}</strong></div>
                <div class="review-item"><span>Branch or revision</span><strong class="code-text">{{ revision }}</strong></div>
                <div class="review-item"><span>Profile</span><strong><v-icon :icon="profile.icon" size="18" class="mr-2" />{{ profile.title }}</strong></div>
                <div class="review-item"><span>Effective skills</span><strong>{{ effectiveSkills.length }} enabled</strong></div>
              </div>
              <div class="session-message-preview"><span>First message</span><p>{{ firstMessage }}</p></div>
              <div class="session-check-list">
                <div><v-icon icon="mdi-check-circle" color="success" size="18" /><span>Repository access and published profile available</span></div>
                <div><v-icon icon="mdi-check-circle" color="success" size="18" /><span>Required secret bindings resolved; no collisions detected</span></div>
                <div><v-icon icon="mdi-check-circle" color="success" size="18" /><span>Resources are within organization limits</span></div>
              </div>
            </div>
          </v-stepper-window-item>
        </v-stepper-window>
      </v-stepper>

      <v-card-actions class="dialog-actions">
        <v-btn variant="text" @click="close">Cancel</v-btn>
        <v-spacer />
        <v-btn v-if="step > 1" variant="outlined" prepend-icon="mdi-arrow-left" @click="step -= 1">Back</v-btn>
        <v-btn color="primary" :disabled="!canContinue" :append-icon="step < 4 ? 'mdi-arrow-right' : 'mdi-play-outline'" @click="continueWizard">{{ step < 4 ? 'Continue' : 'Create & start' }}</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
