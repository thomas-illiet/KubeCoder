<script setup lang="ts">
import { computed, ref } from 'vue'
import DataTableEmptyRow from '../components/DataTableEmptyRow.vue'
import FilterCard from '../components/FilterCard.vue'
import SectionCard from '../components/SectionCard.vue'
import TablePaginationCard from '../components/TablePaginationCard.vue'
import { useNotifications } from '../composables/useNotifications'
import { usePagination } from '../composables/usePagination'

type Transport = 'STDIO' | 'SSE'
type McpServer = {
  name: string
  description: string
  scope: string
  transport: Transport
  target: string
  updated: string
}

const query = ref('')
const scope = ref('All scopes')
const transportFilter = ref('All transports')
const createOpen = ref(false)
const editingServer = ref<McpServer | null>(null)
const serverTransport = ref<Transport>('STDIO')
const serverScope = ref('Global')
const testing = ref(false)
const testResult = ref<'idle' | 'success'>('idle')
const { success } = useNotifications()

const servers: McpServer[] = [
  { name: 'Filesystem workspace', description: 'Controlled read and write access in the run workspace', scope: 'Global', transport: 'STDIO', target: 'mcp-server-filesystem /workspace', updated: '18 min ago' },
  { name: 'Git provider', description: 'Issues, pull requests, and Git metadata', scope: 'Global', transport: 'SSE', target: 'https://mcp.internal/git/sse', updated: '2 hours ago' },
  { name: 'Northstar issue tracker', description: 'Tickets accessible to Northstar agents', scope: 'Northstar Labs', transport: 'SSE', target: 'https://tools.northstar.example/mcp/sse', updated: 'Yesterday' },
  { name: 'Kubernetes read-only', description: 'Non-destructive diagnostics for authorized workloads', scope: 'Northstar Labs', transport: 'STDIO', target: 'kubecoder-mcp-k8s --read-only', updated: '3 days ago' },
]

const filtered = computed(() => servers.filter((server) => {
  const matchesQuery = `${server.name} ${server.description} ${server.target}`.toLocaleLowerCase('en').includes(query.value.toLocaleLowerCase('en'))
  const matchesScope = scope.value === 'All scopes' || server.scope === scope.value
  const matchesTransport = transportFilter.value === 'All transports' || server.transport === transportFilter.value
  return matchesQuery && matchesScope && matchesTransport
}))
const { page, paginatedItems: paginatedServers, itemsPerPage } = usePagination(filtered)

function openCreate() {
  editingServer.value = null
  serverTransport.value = 'STDIO'
  serverScope.value = 'Global'
  testResult.value = 'idle'
  createOpen.value = true
}

function openEdit(server: McpServer) {
  editingServer.value = server
  serverTransport.value = server.transport
  serverScope.value = server.scope
  testResult.value = 'idle'
  createOpen.value = true
}

function testConnection() {
  testing.value = true
  testResult.value = 'idle'
  window.setTimeout(() => {
    testing.value = false
    testResult.value = 'success'
  }, 900)
}

function saveServer() {
  const action = editingServer.value ? 'updated' : 'added'
  const name = editingServer.value?.name ?? 'MCP server'
  createOpen.value = false
  success(`MCP server ${action}`, `${name} was saved with ${serverTransport.value} transport.`)
}
</script>

<template>
  <FilterCard title="Filters" subtitle="Filters and the table remain separate areas" class="mb-4">
    <v-text-field v-model="query" hide-details placeholder="Search MCP servers…" prepend-inner-icon="mdi-magnify" />
    <IconSelect v-model="scope" hide-details :items="['All scopes', 'Global', 'Northstar Labs']" min-width="180" />
    <IconSelect v-model="transportFilter" hide-details :items="['All transports', 'STDIO', 'SSE']" min-width="180" />
  </FilterCard>

  <SectionCard title="MCP servers" subtitle="4 definitions · 2 global · 2 organization">
    <template #actions>
      <v-btn color="primary" prepend-icon="mdi-plus" @click="openCreate">Add server</v-btn>
    </template>
    <div class="table-scroll">
      <table class="data-table">
        <thead><tr><th>SERVER</th><th class="table-cell--center">SCOPE</th><th class="table-cell--center">TRANSPORT</th><th>TARGET</th><th class="table-cell--center">UPDATED</th><th class="table-cell--center">ACTIONS</th></tr></thead>
        <tbody>
          <DataTableEmptyRow v-if="filtered.length === 0" :colspan="6" />
          <tr v-for="server in paginatedServers" :key="server.name">
            <td><div class="entity-cell"><div class="entity-icon"><v-icon icon="mdi-server-network-outline" size="19" /></div><div><div class="entity-name">{{ server.name }}</div><div class="entity-meta">{{ server.description }}</div></div></div></td>
            <td class="table-cell--center"><v-chip size="small" variant="outlined">{{ server.scope }}</v-chip></td>
            <td class="table-cell--center"><v-chip size="small" :color="server.transport === 'STDIO' ? 'secondary' : 'info'" variant="tonal">{{ server.transport }}</v-chip></td>
            <td><span class="code-text mcp-target">{{ server.target }}</span></td>
            <td class="table-cell--center"><span class="text-medium-emphasis">{{ server.updated }}</span></td>
            <td class="table-cell--center"><v-btn variant="text" size="small" prepend-icon="mdi-pencil-outline" @click="openEdit(server)">Edit</v-btn></td>
          </tr>
        </tbody>
      </table>
    </div>
  </SectionCard>
  <TablePaginationCard v-model="page" :total="filtered.length" :items-per-page="itemsPerPage" item-label="servers" />

  <v-dialog v-model="createOpen" max-width="720">
    <v-card class="app-dialog mcp-dialog">
      <div class="dialog-header">
        <div><div class="text-h6 font-weight-bold">{{ editingServer ? 'Edit MCP server' : 'Add MCP server' }}</div><div class="text-caption text-medium-emphasis">Private definition with no implicit activation</div></div>
        <v-btn icon="mdi-close" variant="text" aria-label="Close" @click="createOpen = false" />
      </div>
      <div class="pa-5">
        <div class="form-row">
          <v-text-field label="Name" :model-value="editingServer?.name ?? ''" placeholder="E.g. Git provider" />
          <IconSelect v-model="serverScope" label="Scope" :items="['Global', 'Northstar Labs']" />
        </div>

        <div class="field-label">Transport type</div>
        <v-btn-toggle v-model="serverTransport" mandatory divided class="transport-toggle mb-5">
          <v-btn value="STDIO" prepend-icon="mdi-console-line">STDIO</v-btn>
          <v-btn value="SSE" prepend-icon="mdi-access-point-network">SSE</v-btn>
        </v-btn-toggle>

        <template v-if="serverTransport === 'STDIO'">
          <v-text-field label="Command" :model-value="editingServer?.transport === 'STDIO' ? editingServer.target.split(' ')[0] : ''" placeholder="npx" prepend-inner-icon="mdi-console" />
          <v-textarea label="Command arguments" rows="3" placeholder="-y&#10;@modelcontextprotocol/server-filesystem&#10;/workspace" hint="One argument per line, passed in the displayed order." persistent-hint />
          <v-textarea label="Environment variables" rows="3" placeholder="GITHUB_TOKEN=secret://organization/github-token" hint="One variable per line. Only secret:// references are accepted." persistent-hint />
        </template>

        <template v-else>
          <v-text-field label="SSE endpoint" :model-value="editingServer?.transport === 'SSE' ? editingServer.target : ''" placeholder="https://mcp.example.com/sse" prepend-inner-icon="mdi-link-variant" hint="HTTPS is required outside local environments." persistent-hint />
          <v-textarea label="Headers" rows="3" placeholder="Authorization=secret://organization/mcp-token" hint="One entry per line. Sensitive values use a secret:// reference." persistent-hint />
          <IconSelect label="Credential binding" :items="['None', 'mcp-git-token', 'issue-tracker-token']" model-value="None" prepend-inner-icon="mdi-key-variant" />
        </template>

        <div class="security-note mt-1"><v-icon icon="mdi-docker" color="info" /><span>The test will run in a dedicated ephemeral Docker container with restricted networking, limited resources, and automatic cleanup. No MCP process will be launched by the browser.</span></div>
        <v-alert v-if="testResult === 'success'" type="success" variant="tonal" density="compact" class="mt-4" title="Test successful">The test container started, validated the MCP handshake, and was then removed.</v-alert>
      </div>
      <v-card-actions class="dialog-actions">
        <v-spacer />
        <v-btn variant="text" @click="createOpen = false">Cancel</v-btn>
        <v-btn variant="outlined" prepend-icon="mdi-flask-outline" :loading="testing" @click="testConnection">Test in a container</v-btn>
        <v-btn color="primary" @click="saveServer">{{ editingServer ? 'Save' : 'Add server' }}</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
