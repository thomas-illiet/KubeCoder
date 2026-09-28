<script setup lang="ts">
import { computed, reactive, shallowRef, watch } from 'vue'
import type { PublicAgent } from '../../api/agents'
import type { GitProvider, Repository, RepositoryInput } from '../../api/repositories'
import type { Secret } from '../../api/secrets'
import RepositorySecretTransfer from './RepositorySecretTransfer.vue'
import { canContinueRepositoryStep, isValidCloneURL } from './repositoryForm'

const props = defineProps<{ modelValue: boolean; repository?: Repository | null; agents: readonly PublicAgent[]; secrets: readonly Secret[]; saving?: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean]; submit: [input: RepositoryInput] }>()
const step = shallowRef(1)
const form = reactive<RepositoryInput>({ provider: 'github', clone_url: '', name: '', default_branch: 'main', agent_id: null, include_submodules: false, secret_mode: 'ALL', secret_ids: [] })
const providers: Array<{ value: GitProvider; title: string; icon: string }> = [{ value: 'github', title: 'GitHub', icon: 'mdi-github' }, { value: 'gitlab', title: 'GitLab', icon: 'mdi-gitlab' }, { value: 'bitbucket', title: 'Bitbucket', icon: 'mdi-bitbucket' }]
const selectedAgent = computed(() => props.agents.find((agent) => agent.id === form.agent_id))
const validURL = computed(() => isValidCloneURL(form.clone_url))
const canContinue = computed(() => canContinueRepositoryStep(step.value, form))

watch(() => props.modelValue, (open) => {
  if (!open) return
  step.value = 1
  const repository = props.repository
  Object.assign(form, repository ? { provider: repository.provider, clone_url: repository.clone_url, name: repository.name, default_branch: repository.default_branch, agent_id: repository.agent?.id ?? null, include_submodules: repository.include_submodules, secret_mode: repository.secret_mode, secret_ids: [...repository.secret_ids] } : { provider: 'github', clone_url: '', name: '', default_branch: 'main', agent_id: null, include_submodules: false, secret_mode: 'ALL', secret_ids: [] })
})

function next(): void {
  if (!canContinue.value) return
  if (step.value < 5) step.value += 1
  else emit('submit', { provider: form.provider, clone_url: form.clone_url.trim(), name: form.name.trim(), default_branch: form.default_branch.trim(), agent_id: form.agent_id, include_submodules: form.include_submodules, secret_mode: form.secret_mode, secret_ids: form.secret_mode === 'ALL' ? [] : [...form.secret_ids] })
}
</script>

<template>
  <v-dialog :model-value="modelValue" max-width="920" persistent scrollable @update:model-value="emit('update:modelValue', $event)"><v-card class="app-dialog repository-dialog"><div class="dialog-header"><div><div class="text-h6 font-weight-bold">{{ repository ? `Edit ${repository.name}` : 'Add repository' }}</div><div class="text-caption text-medium-emphasis">Organization repository configuration</div></div><v-btn icon="mdi-close" variant="text" aria-label="Close" @click="emit('update:modelValue', false)" /></div>
    <v-stepper v-model="step" class="repository-stepper" alt-labels><v-stepper-header><v-stepper-item :complete="step > 1" title="Provider" :value="1" /><v-divider /><v-stepper-item :complete="step > 2" title="Connection" :value="2" /><v-divider /><v-stepper-item :complete="step > 3" title="Repository" :value="3" /><v-divider /><v-stepper-item :complete="step > 4" title="Secrets" :value="4" /><v-divider /><v-stepper-item title="Review" :value="5" /></v-stepper-header><v-stepper-window>
      <v-stepper-window-item :value="1"><div class="repository-step-content"><div class="step-heading"><div><h3>Choose a Git provider</h3><p>The clone URL remains compatible with SaaS and self-hosted installations.</p></div></div><div class="provider-grid"><v-card v-for="provider in providers" :key="provider.value" :class="['provider-option', { 'provider-option--active': form.provider === provider.value }]" role="button" tabindex="0" @click="form.provider = provider.value" @keydown.enter="form.provider = provider.value"><div class="provider-option__icon"><v-icon :icon="provider.icon" size="29" /></div><div class="provider-option__copy"><strong>{{ provider.title }}</strong><span>{{ provider.value }}</span></div><v-icon :icon="form.provider === provider.value ? 'mdi-check-circle' : 'mdi-circle-outline'" :color="form.provider === provider.value ? 'success' : undefined" /></v-card></div></div></v-stepper-window-item>
      <v-stepper-window-item :value="2"><div class="repository-step-content"><div class="step-heading"><div><h3>Repository connection</h3><p>Only HTTPS and SSH clone URLs are accepted. Credentials are resolved exclusively by the backend.</p></div></div><v-text-field v-model="form.clone_url" label="Clone URL" placeholder="https://github.com/organization/repository.git" :error="Boolean(form.clone_url) && !validURL" hint="The URL is normalized before storage." persistent-hint /></div></v-stepper-window-item>
      <v-stepper-window-item :value="3"><div class="repository-step-content"><div class="step-heading"><div><h3>Repository defaults</h3><p>Choose the values offered when future sessions start.</p></div></div><div class="form-row"><v-text-field v-model="form.name" label="Display name" maxlength="120" /><v-text-field v-model="form.default_branch" label="Default branch" maxlength="255" /></div><IconSelect v-model="form.agent_id" label="Agent" :items="[{ title: 'Not configured', value: null }, ...agents.map((agent) => ({ title: agent.name, value: agent.id }))]" /><v-switch v-model="form.include_submodules" color="primary" label="Initialize Git submodules" /></div></v-stepper-window-item>
      <v-stepper-window-item :value="4"><div class="repository-step-content repository-step-content--secrets"><div class="step-heading"><div><h3>Repository secrets</h3><p>Use every applicable secret automatically, or select exactly which platform and organization secrets this repository can use.</p></div></div><v-radio-group v-model="form.secret_mode" inline hide-details class="mb-5"><v-radio label="All secrets" value="ALL" /><v-radio label="Selected secrets" value="SELECTED" /></v-radio-group><RepositorySecretTransfer v-if="form.secret_mode === 'SELECTED'" v-model="form.secret_ids" :secrets="secrets" /><div v-else class="security-note"><v-icon icon="mdi-check-circle-outline" color="success" /><span>All current and future platform and organization secrets will be available to this repository.</span></div></div></v-stepper-window-item>
      <v-stepper-window-item :value="5"><div class="repository-step-content"><div class="step-heading"><div><h3>Review configuration</h3><p>Confirm the values before persisting them.</p></div></div><div class="review-grid"><div class="review-item"><span>Provider</span><strong class="text-capitalize">{{ form.provider }}</strong></div><div class="review-item"><span>Clone URL</span><strong class="code-text">{{ form.clone_url }}</strong></div><div class="review-item"><span>Name</span><strong>{{ form.name }}</strong></div><div class="review-item"><span>Default branch</span><strong>{{ form.default_branch }}</strong></div><div class="review-item"><span>Agent</span><strong>{{ selectedAgent?.name ?? 'Not configured' }}</strong></div><div class="review-item"><span>Submodules</span><strong>{{ form.include_submodules ? 'Enabled' : 'Disabled' }}</strong></div><div class="review-item"><span>Secrets</span><strong>{{ form.secret_mode === 'ALL' ? 'All secrets' : `${form.secret_ids.length} selected` }}</strong></div></div></div></v-stepper-window-item>
    </v-stepper-window></v-stepper><v-card-actions class="dialog-actions"><v-btn variant="text" @click="emit('update:modelValue', false)">Cancel</v-btn><v-spacer /><v-btn v-if="step > 1" variant="outlined" prepend-icon="mdi-arrow-left" @click="step -= 1">Back</v-btn><v-btn color="primary" :disabled="!canContinue" :loading="saving" :append-icon="step < 5 ? 'mdi-arrow-right' : 'mdi-content-save-outline'" @click="next">{{ step < 5 ? 'Continue' : (repository ? 'Save changes' : 'Add repository') }}</v-btn></v-card-actions></v-card></v-dialog>
</template>
