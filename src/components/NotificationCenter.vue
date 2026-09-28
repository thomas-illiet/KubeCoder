<script setup lang="ts">
import type { NotificationType } from '../composables/useNotifications'
import { useNotifications } from '../composables/useNotifications'

const { notifications, dismiss } = useNotifications()

const presentation: Record<NotificationType, { icon: string; label: string }> = {
  success: { icon: 'mdi-check-circle', label: 'Success' },
  error: { icon: 'mdi-alert-circle', label: 'Error' },
  warning: { icon: 'mdi-alert', label: 'Warning' },
  info: { icon: 'mdi-information', label: 'Information' },
}
</script>

<template>
  <div class="notification-region" aria-live="polite" aria-label="Notifications">
    <TransitionGroup name="notification" tag="div" class="notification-stack">
      <v-card
        v-for="notification in notifications"
        :key="notification.id"
        class="app-notification"
        :class="`app-notification--${notification.type}`"
        role="status"
        rounded="lg"
        elevation="12"
      >
        <div class="app-notification__accent" />
        <v-icon
          class="app-notification__icon"
          :icon="presentation[notification.type].icon"
          :color="notification.type"
          size="24"
        />
        <div class="app-notification__content">
          <div class="app-notification__eyebrow">{{ presentation[notification.type].label }}</div>
          <div class="app-notification__title">{{ notification.title }}</div>
          <div v-if="notification.message" class="app-notification__message">{{ notification.message }}</div>
        </div>
        <v-btn
          icon="mdi-close"
          variant="text"
          size="small"
          class="app-notification__close"
          :aria-label="`Dismiss ${notification.title}`"
          @click="dismiss(notification.id)"
        />
      </v-card>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.notification-region {
  position: fixed;
  right: 24px;
  bottom: 24px;
  z-index: 2600;
  width: min(380px, calc(100vw - 32px));
  pointer-events: none;
}

.notification-stack {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.app-notification {
  position: relative;
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  gap: 12px;
  align-items: start;
  min-height: 88px;
  padding: 16px 12px 16px 18px;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(23, 28, 36, 0.98);
  backdrop-filter: blur(16px);
  pointer-events: auto;
}

.app-notification__accent {
  position: absolute;
  inset: 0 auto 0 0;
  width: 4px;
  background: rgb(var(--v-theme-info));
}

.app-notification--success .app-notification__accent { background: rgb(var(--v-theme-success)); }
.app-notification--error .app-notification__accent { background: rgb(var(--v-theme-error)); }
.app-notification--warning .app-notification__accent { background: rgb(var(--v-theme-warning)); }

.app-notification__icon { margin-top: 2px; }
.app-notification__content { min-width: 0; }

.app-notification__eyebrow {
  margin-bottom: 2px;
  color: rgba(255, 255, 255, 0.56);
  font-size: 0.68rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  line-height: 1.25;
  text-transform: uppercase;
}

.app-notification__title {
  color: rgba(255, 255, 255, 0.96);
  font-size: 0.92rem;
  font-weight: 700;
  line-height: 1.35;
}

.app-notification__message {
  margin-top: 4px;
  color: rgba(255, 255, 255, 0.68);
  font-size: 0.8rem;
  line-height: 1.4;
}

.app-notification__close { margin: -6px -4px 0 0; }

.notification-enter-active,
.notification-leave-active,
.notification-move {
  transition: opacity 180ms ease, transform 180ms ease;
}

.notification-enter-from,
.notification-leave-to {
  opacity: 0;
  transform: translateX(24px);
}

@media (max-width: 600px) {
  .notification-region {
    right: 16px;
    bottom: 16px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .notification-enter-active,
  .notification-leave-active,
  .notification-move { transition: none; }
}
</style>
