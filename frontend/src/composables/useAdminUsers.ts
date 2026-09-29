import { onMounted, readonly, shallowRef, watch } from 'vue'
import { deleteAdminUser, fetchAdminUsers, updateAdminUserRole, type AdminUserPageRequest, type CurrentUser } from '../api/users'
import { useAuth } from './useAuth'
import { useNotifications } from './useNotifications'
import { useRuntimeConfig } from '../runtime/config'

const ITEMS_PER_PAGE = 20

export function useAdminUsers() {
  const auth = useAuth()
  const config = useRuntimeConfig()
  const notifications = useNotifications()
  const users = shallowRef<CurrentUser[]>([])
  const query = shallowRef('')
  const role = shallowRef<'all' | 'admin' | 'user'>('all')
  const page = shallowRef(1)
  const total = shallowRef(0)
  const loading = shallowRef(false)
  const mutating = shallowRef(false)
  const orderBy = shallowRef<NonNullable<AdminUserPageRequest['orderBy']>>('display_name')
  const orderDirection = shallowRef<NonNullable<AdminUserPageRequest['orderDirection']>>('asc')
  let loadRevision = 0

  async function load(revision = ++loadRevision): Promise<void> {
    const token = auth.accessToken.value
    if (!token) return
    loading.value = true
    try {
      const result = await fetchAdminUsers(config.apiBaseUrl, token, {
        query: query.value.trim(),
        role: role.value,
        limit: ITEMS_PER_PAGE,
        offset: (page.value - 1) * ITEMS_PER_PAGE,
        orderBy: orderBy.value,
        orderDirection: orderDirection.value,
      })
      if (revision !== loadRevision) return
      users.value = result.items ?? []
      total.value = result.total
    } catch (error) {
      if (revision === loadRevision) {
        users.value = []
        notifications.error('Users could not be loaded', error instanceof Error ? error.message : undefined)
      }
    } finally {
      if (revision === loadRevision) loading.value = false
    }
  }

  function resetPageAndLoad(): void {
    if (page.value !== 1) page.value = 1
    else void load()
  }

  function changeSort(column: NonNullable<AdminUserPageRequest['orderBy']>): void {
    orderDirection.value = orderBy.value === column && orderDirection.value === 'asc' ? 'desc' : 'asc'
    orderBy.value = column
    resetPageAndLoad()
  }

  async function updateRole(user: CurrentUser, isAdmin: boolean): Promise<boolean> {
    const token = auth.accessToken.value
    if (!token) return false
    mutating.value = true
    try {
      await updateAdminUserRole(config.apiBaseUrl, token, user.id, isAdmin)
      notifications.success(isAdmin ? 'Administrator role granted' : 'Administrator role removed')
      await load()
      return true
    } catch (error) {
      notifications.error('User role could not be changed', error instanceof Error ? error.message : undefined)
      return false
    } finally {
      mutating.value = false
    }
  }

  async function remove(user: CurrentUser): Promise<boolean> {
    const token = auth.accessToken.value
    if (!token) return false
    mutating.value = true
    try {
      await deleteAdminUser(config.apiBaseUrl, token, user.id)
      notifications.success('User deleted')
      if (users.value.length === 1 && page.value > 1) page.value -= 1
      else await load()
      return true
    } catch (error) {
      notifications.error('User could not be deleted', error instanceof Error ? error.message : undefined)
      return false
    } finally {
      mutating.value = false
    }
  }

  watch(query, (_value, _previous, onCleanup) => {
    const revision = ++loadRevision
    loading.value = true
    if (page.value !== 1) {
      page.value = 1
      return
    }
    const timeout = window.setTimeout(() => void load(revision), 300)
    onCleanup(() => window.clearTimeout(timeout))
  })

  watch(page, () => void load())
  watch(role, resetPageAndLoad)
  onMounted(() => void load())

  return {
    users: readonly(users),
    query,
    role,
    page,
    total: readonly(total),
    loading: readonly(loading),
    mutating: readonly(mutating),
    orderBy: readonly(orderBy),
    orderDirection: readonly(orderDirection),
    itemsPerPage: ITEMS_PER_PAGE,
    reload: load,
    changeSort,
    updateRole,
    remove,
  }
}
