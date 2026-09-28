<script setup lang="ts">
import { computed, ref } from 'vue'
import DataTableEmptyRow from '../components/DataTableEmptyRow.vue'
import FilterCard from '../components/FilterCard.vue'
import SectionCard from '../components/SectionCard.vue'
import TablePaginationCard from '../components/TablePaginationCard.vue'
import { useNotifications } from '../composables/useNotifications'
import { usePagination } from '../composables/usePagination'

type Transport = 'STDIO' | 'SSE'
type OrganizationMcpServer = {
  name: string
  description: string
  source: 'Global' | 'Organization'
  transport: Transport
  enabled: boolean
}

const query = ref('')
const transport = ref('All transports')
const editorOpen = ref(false)
const editingServer = ref<OrganizationMcpServer | null>(null)
const editorTransport = ref<Transport>('STDIO')
const testing = ref(false)
const testSucceeded = ref(false)
const { success, info } = useNotifications()

const servers = ref<OrganizationMcpServer[]>([
  { name: 'Filesystem workspace', description: 'Files in the session workspace', source: 'Global', transport: 'STDIO', enabled: false },
  { name: 'Git provider', description: 'Repository issues and pull requests', source: 'Global', transport: 'SSE', enabled: false },
  { name: 'Northstar issue tracker', description: 'Northstar tickets and projects', source: 'Organization', transport: 'SSE', enabled: false },
  { name: 'Kubernetes read-only', description: 'Diagnostics for authorized workloads', source: 'Organization', transport: 'STDIO', enabled: false },
])

const filtered = computed(() => servers.value.filter((server) => {
  const matchesQuery = `${server.name} ${server.description}`.toLocaleLowerCase('en').includes(query.value.toLocaleLowerCase('en'))
  return matchesQuery && (transport.value === 'All transports' || server.transport === transport.value)
}))
const { page, paginatedItems: paginatedServers, itemsPerPage } = usePagination(filtered)
const enabledCount = computed(() => servers.value.filter((server) => server.enabled).length)

function openAdd() {
  editingServer.value = null
  editorTransport.value = 'STDIO'
  testSucceeded.value = false
  editorOpen.value = true
}

function openEdit(server: OrganizationMcpServer) {
  editingServer.value = server
  editorTransport.value = server.transport
  testSucceeded.value = false
  editorOpen.value = true
}

function testConnection() {
  testing.value = true
  testSucceeded.value = false
  window.setTimeout(() => {
    testing.value = false
    testSucceeded.value = true
  }, 900)
}

function saveServer() {
  const action = editingServer.value ? 'updated' : 'added'
  const name = editingServer.value?.name ?? 'MCP server'
  editorOpen.value = false
  success(`MCP server ${action}`, `${name} remains disabled until explicitly enabled.`)
}

function notifyActivation(server: OrganizationMcpServer, enabled: boolean | null) {
  info(`MCP server ${enabled === true ? 'enabled' : 'disabled'}`, `${server.name} was updated for Northstar Labs.`)
}
</script>

<template>
  <FilterCard title="Filters" subtitle="Servers are disabled by default in the organization" class="mb-4">
    <v-text-field v-model="query" hide-details placeholder="Search MCP servers…" prepend-inner-icon="mdi-magnify" />
    <IconSelect v-model="transport" hide-details :items="['All transports', 'STDIO', 'SSE']" min-width="190" />
    <v-spacer />
    <v-chip color="secondary" variant="tonal" prepend-icon="mdi-domain">Northstar Labs</v-chip>
  </FilterCard>

  <SectionCard title="MCP servers" :subtitle="`${enabledCount} enabled · ${servers.length - enabledCount} disabled`">
    <template #actions>
      <v-btn color="primary" prepend-icon="mdi-plus" @click="openAdd">Add server</v-btn>
    </template>
    <div class="table-scroll">
      <table class="data-table">
        <thead><tr><th>SERVER</th><th class="table-cell--center">SOURCE</th><th class="table-cell--center">TRANSPORT</th><th class="table-cell--center">ACTIVATION</th><th class="table-cell--center">ACTIONS</th></tr></thead>
        <tbody>
          <DataTableEmptyRow v-if="filtered.length === 0" :colspan="5" />
          <tr v-for="server in paginatedServers" :key="server.name">
            <td><div class="entity-cell"><div class="entity-icon"><v-icon icon="mdi-server-network-outline" size="19" /></div><div><div class="entity-name">{{ server.name }}</div><div class="entity-meta">{{ server.description }}</div></div></div></td>
            <td class="table-cell--center"><v-chip size="small" variant="outlined" :color="server.source === 'Global' ? 'default' : 'secondary'">{{ server.source }}</v-chip></td>
            <td class="table-cell--center"><v-chip size="small" :color="server.transport === 'STDIO' ? 'secondary' : 'info'" variant="tonal">{{ server.transport }}</v-chip></td>
            <td class="table-cell--center"><v-switch v-model="server.enabled" hide-details color="success" density="compact" :label="server.enabled ? 'Enabled' : 'Disabled'" @update:model-value="notifyActivation(server, $event)" /></td>
            <td class="table-cell--center">
              <v-btn v-if="server.source === 'Organization'" variant="text" size="small" prepend-icon="mdi-pencil-outline" @click="openEdit(server)">Edit</v-btn>
              <v-chip v-else size="small" variant="text" prepend-icon="mdi-lock-outline" color="default">Managed globally</v-chip>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <div class="security-note ma-4"><v-icon icon="mdi-shield-lock-outline" color="secondary" /><span>Adding a server does not enable it. Explicit activation is required before it can be linked to a new agent version.</span></div>
  </SectionCard>
  <TablePaginationCard v-model="page" :total="filtered.length" :items-per-page="itemsPerPage" item-label="servers" />

  <v-dialog v-model="editorOpen" max-width="720">
    <v-card class="app-dialog mcp-dialog">
      <div class="dialog-header">
        <div><div class="text-h6 font-weight-bold">{{ editingServer ? 'Edit MCP server' : 'Add MCP server' }}</div><div class="text-caption text-medium-emphasis">Definition owned by Northstar Labs</div></div>
        <v-btn icon="mdi-close" variant="text" aria-label="Close" @click="editorOpen = false" />
      </div>
      <div class="pa-5">
        <v-text-field label="Name" :model-value="editingServer?.name ?? ''" placeholder="E.g. Linear" />
        <div class="field-label">Transport type</div>
        <v-btn-toggle v-model="editorTransport" mandatory divided class="transport-toggle mb-5">
          <v-btn value="STDIO" prepend-icon="mdi-console-line">STDIO</v-btn>
          <v-btn value="SSE" prepend-icon="mdi-access-point-network">SSE</v-btn>
        </v-btn-toggle>
        <template v-if="editorTransport === 'STDIO'">
          <v-text-field label="Command" placeholder="npx" prepend-inner-icon="mdi-console" />
          <v-textarea label="Command arguments" rows="3" placeholder="-y&#10;@modelcontextprotocol/server-filesystem&#10;/workspace" />
          <v-textarea label="Environment variables" rows="3" placeholder="TOKEN=secret://organization/mcp-token" hint="One variable per line. Only secret:// references are accepted." persistent-hint />
        </template>
        <template v-else>
          <v-text-field label="Endpoint SSE" placeholder="https://mcp.example.com/sse" prepend-inner-icon="mdi-link-variant" />
          <v-textarea label="Headers" rows="3" placeholder="Authorization=secret://organization/mcp-token" hint="One entry per line. Sensitive values use a secret:// reference." persistent-hint />
        </template>
        <div class="security-note mt-1"><v-icon icon="mdi-docker" color="info" /><span>The test uses a dedicated ephemeral Docker container. Adding or editing keeps the server disabled until you explicitly enable it in the list.</span></div>
        <v-alert v-if="testSucceeded" type="success" variant="tonal" density="compact" class="mt-4" title="Test successful">MCP handshake validated, then the container was removed.</v-alert>
      </div>
      <v-card-actions class="dialog-actions">
        <v-spacer />
        <v-btn variant="text" @click="editorOpen = false">Cancel</v-btn>
        <v-btn variant="outlined" prepend-icon="mdi-flask-outline" :loading="testing" @click="testConnection">Test in a container</v-btn>
        <v-btn color="primary" @click="saveServer">{{ editingServer ? 'Save' : 'Add server' }}</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
