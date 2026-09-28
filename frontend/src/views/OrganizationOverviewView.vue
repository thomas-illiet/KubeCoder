<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import NewSessionDialog, { type CreatedSession } from '../components/NewSessionDialog.vue'
import SectionCard from '../components/SectionCard.vue'
import StatusChip from '../components/StatusChip.vue'
import { useNotifications } from '../composables/useNotifications'
import { useOrganizations } from '../composables/useOrganizations'

const router = useRouter()
const sessionOpen = ref(false)
const { success } = useNotifications()
const organizations = useOrganizations()
const organizationBase = computed(() => `/organizations/${organizations.state.current?.slug ?? ''}`)

function sessionCreated(session: CreatedSession) {
  success('Session created', `${session.profile} is queued for ${session.repository}.`)
}

const stats = [
  { label: 'Repositories', value: '4', caption: '3 ready for a session', icon: 'mdi-source-repository', color: '#7c86ff' },
  { label: 'Active sessions', value: '2', caption: '1 run in progress', icon: 'mdi-message-processing-outline', color: '#33d2d0' },
  { label: 'Agents', value: '3', caption: 'Published profiles', icon: 'mdi-robot-outline', color: '#5aa7ff' },
  { label: 'Action required', value: '1', caption: 'Repository needs configuration', icon: 'mdi-alert-circle-outline', color: '#f6b84c' },
]

const sessions = [
  { title: 'Complete the OAuth migration', repository: 'identity-service', agent: 'Atlas', status: 'In progress', time: '8 min ago' },
  { title: 'Update manifests', repository: 'platform-k8s', agent: 'Pathfinder', status: 'Pending', time: '36 min ago' },
  { title: 'Review the payment module', repository: 'checkout-api', agent: 'Sentinel', status: 'Completed', time: 'Yesterday' },
]
</script>

<template>
  <v-card class="workspace-hero mb-4">
    <div class="workspace-hero__content">
      <div><div class="eyebrow mb-2">ORGANIZATION WORKSPACE</div><h2>Resume work in {{ organizations.state.current?.name }}</h2><p>Choose a repository and a published agent, then start a session without accessing sensitive administration settings.</p></div>
      <v-btn color="primary" prepend-icon="mdi-plus" size="large" @click="sessionOpen = true">New session</v-btn>
    </div>
  </v-card>

  <div class="stat-grid mb-4">
    <v-card v-for="stat in stats" :key="stat.label" class="stat-card">
      <div class="stat-card__top"><span>{{ stat.label }}</span><div class="stat-card__icon" :style="{ color: stat.color, background: `${stat.color}15` }"><v-icon :icon="stat.icon" size="19" /></div></div>
      <div class="stat-card__value">{{ stat.value }}</div><div class="stat-card__caption">{{ stat.caption }}</div>
    </v-card>
  </div>

  <div class="split-grid">
    <SectionCard title="Recent sessions" subtitle="Resume a conversation or check its status">
      <template #actions><v-btn variant="text" color="primary" @click="router.push(`${organizationBase}/sessions`)">All sessions</v-btn></template>
      <div class="activity-list">
        <div v-for="item in sessions" :key="item.title" class="activity-item">
          <div class="activity-icon"><v-icon icon="mdi-message-text-outline" /></div>
          <div><div class="activity-title">{{ item.title }}</div><div class="activity-meta">{{ item.repository }} · {{ item.agent }}</div></div>
          <div class="d-flex align-center ga-3"><StatusChip :label="item.status" :color="item.status === 'Completed' ? 'success' : item.status === 'Pending' ? 'warning' : 'info'" /><span class="activity-time">{{ item.time }}</span></div>
        </div>
      </div>
    </SectionCard>
    <SectionCard title="Quick access" subtitle="Developer workflow">
      <v-list bg-color="transparent" class="py-2">
        <v-list-item title="Browse repositories" subtitle="Configure an authorized agent" prepend-icon="mdi-source-repository" append-icon="mdi-chevron-right" @click="router.push(`${organizationBase}/repositories`)" />
        <v-list-item title="Review skills" subtitle="Understand the effective configuration" prepend-icon="mdi-puzzle-outline" append-icon="mdi-chevron-right" @click="router.push(`${organizationBase}/skills`)" />
      </v-list>
    </SectionCard>
  </div>

  <NewSessionDialog v-model="sessionOpen" @created="sessionCreated" />
</template>
