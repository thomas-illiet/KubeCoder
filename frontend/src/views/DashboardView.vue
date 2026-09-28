<script setup lang="ts">
import { useRouter } from 'vue-router'
import SectionCard from '../components/SectionCard.vue'
import StatusChip from '../components/StatusChip.vue'

const router = useRouter()

const stats = [
  { label: 'Published agents', value: '3', caption: '2 definitions not published', icon: 'mdi-robot-outline', color: '#7c86ff', glow: 'rgba(124,134,255,.16)' },
  { label: 'Secrets', value: '4', caption: '2 active · 2 need attention', icon: 'mdi-key-variant', color: '#33d2d0', glow: 'rgba(51,210,208,.13)' },
  { label: 'Members', value: '5', caption: '2 administrator roles', icon: 'mdi-account-group-outline', color: '#5aa7ff', glow: 'rgba(90,167,255,.14)' },
  { label: 'Alerts', value: '2', caption: 'Secret attention required', icon: 'mdi-alert-circle-outline', color: '#f6b84c', glow: 'rgba(246,184,76,.14)' },
]

const activity = [
  { icon: 'mdi-robot-outline', title: 'Agent “Atlas” published', meta: 'Alex Martin · version 3', time: '12 min ago' },
  { icon: 'mdi-key-variant', title: 'Secret “MODEL_API_TOKEN” rotated', meta: 'Samira Chen · organization', time: '1 hour ago' },
  { icon: 'mdi-puzzle-outline', title: 'Skill “Go reviewer” enabled', meta: 'Repository policy', time: '3 hours ago' },
  { icon: 'mdi-account-plus-outline', title: 'Invitation sent', meta: 'Jamie Rivera · Member role', time: 'Yesterday' },
]

const quickLinks = [
  { title: 'Create agent', subtitle: 'Configure a new definition', icon: 'mdi-robot-outline', to: '/admin/agents' },
  { title: 'Add secret', subtitle: 'Create a protected reference', icon: 'mdi-key-plus', to: '/admin/secrets' },
  { title: 'Invite member', subtitle: 'Grant workspace access', icon: 'mdi-account-plus-outline', to: '/admin/members' },
]
</script>

<template>
  <div class="stat-grid mb-4">
    <v-card v-for="stat in stats" :key="stat.label" class="stat-card" :style="{ '--stat-glow': stat.glow }">
      <div class="stat-card__top">
        <span>{{ stat.label }}</span>
        <div class="stat-card__icon" :style="{ color: stat.color, background: `${stat.color}15` }"><v-icon :icon="stat.icon" size="19" /></div>
      </div>
      <div class="stat-card__value">{{ stat.value }}</div>
      <div class="stat-card__caption">{{ stat.caption }}</div>
    </v-card>
  </div>

  <div class="split-grid mb-4">
    <SectionCard title="Configuration status" subtitle="Workspace readiness for future runs" icon="mdi-shield-check-outline">
      <div class="pa-5">
        <div class="progress-row">
          <div class="progress-row__labels"><span>Validated agent definitions</span><span>3 / 5</span></div>
          <v-progress-linear model-value="60" color="primary" bg-color="#252d42" height="7" />
        </div>
        <div class="progress-row">
          <div class="progress-row__labels"><span>Recently rotated secrets</span><span>2 / 4</span></div>
          <v-progress-linear model-value="50" color="secondary" bg-color="#252d42" height="7" />
        </div>
        <div class="progress-row">
          <div class="progress-row__labels"><span>Verified skills</span><span>4 / 5</span></div>
          <v-progress-linear model-value="80" color="info" bg-color="#252d42" height="7" />
        </div>
        <div class="security-note mt-5">
          <v-icon icon="mdi-information-outline" size="19" color="info" />
          <span>No secrets or runtime environment credentials are displayed in this dashboard.</span>
        </div>
      </div>
    </SectionCard>

    <SectionCard title="Quick actions" subtitle="Configure the workspace">
      <v-list bg-color="transparent" class="py-2">
        <v-list-item v-for="link in quickLinks" :key="link.title" :prepend-icon="link.icon" :title="link.title" :subtitle="link.subtitle" append-icon="mdi-chevron-right" min-height="62" @click="router.push(link.to)" />
      </v-list>
    </SectionCard>
  </div>

  <div class="split-grid">
    <SectionCard title="Recent activity" subtitle="Latest administration operations">
      <div class="activity-list">
        <div v-for="item in activity" :key="item.title" class="activity-item">
          <div class="activity-icon"><v-icon :icon="item.icon" size="17" /></div>
          <div><div class="activity-title">{{ item.title }}</div><div class="activity-meta">{{ item.meta }}</div></div>
          <div class="activity-time">{{ item.time }}</div>
        </div>
      </div>
    </SectionCard>

    <SectionCard title="Platform health" subtitle="Simulated components">
      <div class="pa-4">
        <div v-for="service in ['Control API', 'Reconciler', 'Artifact storage', 'OIDC provider']" :key="service" class="d-flex align-center justify-space-between py-2">
          <span class="text-body-2 text-medium-emphasis">{{ service }}</span>
          <StatusChip label="Operational" color="success" icon="mdi-check-circle-outline" />
        </div>
      </div>
    </SectionCard>
  </div>
</template>
