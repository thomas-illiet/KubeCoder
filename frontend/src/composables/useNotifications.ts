import { readonly, ref } from 'vue'

export type NotificationType = 'success' | 'error' | 'warning' | 'info'

export type AppNotification = {
  id: number
  type: NotificationType
  title: string
  message?: string
}

type NotificationInput = Omit<AppNotification, 'id'> & {
  duration?: number
}

const notifications = ref<AppNotification[]>([])
const timers = new Map<number, ReturnType<typeof setTimeout>>()
let nextId = 0

function dismiss(id: number) {
  const timer = timers.get(id)
  if (timer) clearTimeout(timer)
  timers.delete(id)
  notifications.value = notifications.value.filter((notification) => notification.id !== id)
}

function notify(input: NotificationInput) {
  const id = ++nextId
  notifications.value = [...notifications.value, { id, type: input.type, title: input.title, message: input.message }]

  if (input.duration !== 0) {
    timers.set(id, setTimeout(() => dismiss(id), input.duration ?? 5000))
  }

  return id
}

export function useNotifications() {
  return {
    notifications: readonly(notifications),
    dismiss,
    notify,
    success: (title: string, message?: string) => notify({ type: 'success', title, message }),
    error: (title: string, message?: string) => notify({ type: 'error', title, message, duration: 7000 }),
    warning: (title: string, message?: string) => notify({ type: 'warning', title, message, duration: 6000 }),
    info: (title: string, message?: string) => notify({ type: 'info', title, message }),
  }
}
