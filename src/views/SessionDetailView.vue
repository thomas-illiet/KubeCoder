<script setup lang="ts">
import { computed, ref } from 'vue'
import StatusChip from '../components/StatusChip.vue'
import { useNotifications } from '../composables/useNotifications'

type ChatMessage = {
  id: number
  role: 'user' | 'assistant'
  author: string
  time: string
  content: string
  code?: string
}

type DiffLine = {
  type: 'meta' | 'context' | 'add' | 'remove'
  oldNo?: number
  newNo?: number
  text: string
}

type ChangedFile = {
  path: string
  status: 'M' | 'A'
  additions: number
  deletions: number
  lines: DiffLine[]
}

const activeTab = ref<'chat' | 'changes'>('chat')
const draft = ref('')
const { warning } = useNotifications()
const messages = ref<ChatMessage[]>([
  { id: 1, role: 'user', author: 'Alex Martin', time: '09:34', content: 'Finish the OAuth migration. Replace the legacy callback handler with the new OIDC flow and update the tests.' },
  { id: 2, role: 'assistant', author: 'Atlas · OpenCode', time: '09:35', content: 'I will inspect the authentication package and its existing tests before changing the callback flow.' },
  { id: 3, role: 'assistant', author: 'Atlas · OpenCode', time: '09:38', content: 'The callback now validates state through the session store, exchanges the authorization code, and keeps provider-specific details behind the standard OIDC adapter.', code: 'go test ./internal/auth/...\n✓ 18 tests passed · 0 failed' },
  { id: 4, role: 'user', author: 'Alex Martin', time: '09:40', content: 'Good. Please also make the missing-state error explicit and add coverage for it.' },
  { id: 5, role: 'assistant', author: 'Atlas · OpenCode', time: '09:42', content: 'Done. I added the explicit error and a regression test. Three files are changed and the focused test suite passes.' },
])

const files: ChangedFile[] = [
  {
    path: 'internal/auth/callback.go', status: 'M', additions: 12, deletions: 6,
    lines: [
      { type: 'meta', text: '@@ -41,13 +41,19 @@ func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {' },
      { type: 'context', oldNo: 41, newNo: 41, text: '    ctx := r.Context()' },
      { type: 'remove', oldNo: 42, text: '    state := r.URL.Query().Get("state")' },
      { type: 'remove', oldNo: 43, text: '    if state == "" { http.Error(w, "invalid request", 400); return }' },
      { type: 'add', newNo: 42, text: '    state := r.URL.Query().Get("state")' },
      { type: 'add', newNo: 43, text: '    if state == "" {' },
      { type: 'add', newNo: 44, text: '        respondError(w, ErrMissingOAuthState)' },
      { type: 'add', newNo: 45, text: '        return' },
      { type: 'add', newNo: 46, text: '    }' },
      { type: 'context', oldNo: 44, newNo: 47, text: '' },
      { type: 'remove', oldNo: 45, text: '    claims, err := h.provider.Exchange(ctx, code)' },
      { type: 'add', newNo: 48, text: '    claims, err := h.oidc.ExchangeCode(ctx, code, state)' },
      { type: 'context', oldNo: 46, newNo: 49, text: '    if err != nil {' },
      { type: 'context', oldNo: 47, newNo: 50, text: '        respondError(w, err)' },
    ],
  },
  {
    path: 'internal/auth/errors.go', status: 'A', additions: 7, deletions: 0,
    lines: [
      { type: 'meta', text: '@@ -0,0 +1,7 @@' },
      { type: 'add', newNo: 1, text: 'package auth' },
      { type: 'add', newNo: 2, text: '' },
      { type: 'add', newNo: 3, text: 'import "errors"' },
      { type: 'add', newNo: 4, text: '' },
      { type: 'add', newNo: 5, text: 'var ErrMissingOAuthState = errors.New(' },
      { type: 'add', newNo: 6, text: '    "missing OAuth state",' },
      { type: 'add', newNo: 7, text: ')' },
    ],
  },
  {
    path: 'internal/auth/callback_test.go', status: 'M', additions: 18, deletions: 2,
    lines: [
      { type: 'meta', text: '@@ -88,6 +88,22 @@ func TestCallback(t *testing.T) {' },
      { type: 'context', oldNo: 88, newNo: 88, text: '    t.Run("valid callback", testValidCallback)' },
      { type: 'add', newNo: 89, text: '' },
      { type: 'add', newNo: 90, text: '    t.Run("missing state", func(t *testing.T) {' },
      { type: 'add', newNo: 91, text: '        req := httptest.NewRequest(http.MethodGet, "/callback?code=test", nil)' },
      { type: 'add', newNo: 92, text: '        res := httptest.NewRecorder()' },
      { type: 'add', newNo: 93, text: '' },
      { type: 'add', newNo: 94, text: '        handler.Callback(res, req)' },
      { type: 'add', newNo: 95, text: '' },
      { type: 'add', newNo: 96, text: '        require.Equal(t, http.StatusBadRequest, res.Code)' },
      { type: 'add', newNo: 97, text: '        require.Contains(t, res.Body.String(), "missing OAuth state")' },
      { type: 'add', newNo: 98, text: '    })' },
    ],
  },
]

const selectedFilePath = ref(files[0].path)
const selectedFile = computed(() => files.find((file) => file.path === selectedFilePath.value) ?? files[0])
const totalAdditions = computed(() => files.reduce((total, file) => total + file.additions, 0))
const totalDeletions = computed(() => files.reduce((total, file) => total + file.deletions, 0))

function sendMessage() {
  const content = draft.value.trim()
  if (!content) return
  messages.value.push({ id: Date.now(), role: 'user', author: 'Alex Martin', time: 'Now', content })
  draft.value = ''
}

function stopRun() {
  warning('Run stopped', 'The active OpenCode execution was interrupted. Session history was preserved.')
}
</script>

<template>
  <v-card class="session-header mb-4">
    <div class="session-header__main">
      <v-btn to="/organization/sessions" icon="mdi-arrow-left" variant="text" aria-label="Back to sessions" />
      <div class="session-agent-icon"><v-icon icon="mdi-creation-outline" size="22" /></div>
      <div class="min-w-0"><div class="session-header__title">Complete the OAuth migration</div><div class="session-header__meta"><span class="code-text">identity-service</span><span>feat/oidc</span><span>Atlas · OpenCode 1.2</span></div></div>
      <v-spacer />
      <StatusChip label="In progress" color="success" icon="mdi-circle-slice-8" />
      <v-btn variant="outlined" prepend-icon="mdi-stop-circle-outline" @click="stopRun">Stop run</v-btn>
    </div>
    <div class="session-tabs">
      <v-tabs v-model="activeTab" color="primary" density="comfortable">
        <v-tab value="chat" prepend-icon="mdi-message-text-outline">Chat</v-tab>
        <v-tab value="changes" prepend-icon="mdi-source-branch"><span>Code changes</span><v-chip size="x-small" class="ml-2">{{ files.length }}</v-chip></v-tab>
      </v-tabs>
      <div class="session-connection"><span class="session-connection__dot" />Connected</div>
    </div>
  </v-card>

  <v-card v-if="activeTab === 'chat'" class="session-workspace">
    <div class="session-chat-grid">
      <section class="session-thread">
        <div class="session-messages">
          <template v-for="(message, index) in messages" :key="message.id">
            <div v-if="index === 2" class="tool-event">
              <div class="tool-event__icon"><v-icon icon="mdi-file-search-outline" size="18" /></div>
              <div><strong>Inspected authentication flow</strong><span>Read 8 files · searched 14 symbols</span></div>
              <StatusChip label="Completed" color="success" icon="mdi-check" />
            </div>
            <article class="chat-message" :class="`chat-message--${message.role}`">
              <div class="chat-message__avatar"><span v-if="message.role === 'user'">AM</span><v-icon v-else icon="mdi-creation-outline" size="19" /></div>
              <div class="chat-message__body">
                <div class="chat-message__meta"><strong>{{ message.author }}</strong><span>{{ message.time }}</span></div>
                <p>{{ message.content }}</p>
                <pre v-if="message.code" class="chat-code"><code>{{ message.code }}</code></pre>
              </div>
            </article>
            <div v-if="index === 2" class="tool-event">
              <div class="tool-event__icon"><v-icon icon="mdi-file-edit-outline" size="18" /></div>
              <div><strong>Applied repository patch</strong><span>3 files changed · +37 −8</span></div>
              <v-btn size="small" variant="text" append-icon="mdi-arrow-right" @click="activeTab = 'changes'">View changes</v-btn>
            </div>
          </template>
        </div>
        <div class="session-composer">
          <v-textarea v-model="draft" hide-details rows="2" auto-grow max-rows="5" placeholder="Ask OpenCode to continue, explain, or update the implementation…" @keydown.ctrl.enter.prevent="sendMessage" />
          <div class="session-composer__actions"><div class="text-caption text-medium-emphasis"><v-icon icon="mdi-information-outline" size="14" /> Ctrl + Enter to send</div><v-btn color="primary" prepend-icon="mdi-send" :disabled="!draft.trim()" @click="sendMessage">Send</v-btn></div>
        </div>
      </section>

      <aside class="session-context-panel">
        <div class="session-panel-section"><div class="session-panel-title">Run</div><div class="session-kv"><span>Duration</span><strong>08:42</strong></div><div class="session-kv"><span>Messages</span><strong>{{ messages.length }}</strong></div><div class="session-kv"><span>Context</span><strong>42%</strong></div><v-progress-linear model-value="42" height="5" color="primary" class="mt-2" /></div>
        <div class="session-panel-section"><div class="session-panel-title">Repository</div><div class="session-kv"><span>Branch</span><code>feat/oidc</code></div><div class="session-kv"><span>Base</span><code>main</code></div><div class="session-kv"><span>Changes</span><strong class="diff-summary"><span>+{{ totalAdditions }}</span> <em>−{{ totalDeletions }}</em></strong></div></div>
        <div class="session-panel-section"><div class="session-panel-title">Effective skills</div><div class="d-flex flex-wrap ga-2"><v-chip size="small" variant="outlined">Repository review</v-chip><v-chip size="small" variant="outlined">Go reviewer</v-chip><v-chip size="small" variant="outlined">Security baseline</v-chip></div></div>
        <div class="session-panel-section"><div class="session-panel-title">Execution</div><div class="session-kv"><span>Runtime</span><strong>OpenCode 1.2</strong></div><div class="session-kv"><span>Workspace</span><StatusChip label="Healthy" color="success" icon="mdi-check-circle-outline" /></div></div>
      </aside>
    </div>
  </v-card>

  <v-card v-else class="session-workspace code-workspace">
    <aside class="changed-files-panel">
      <div class="changed-files-header"><div><strong>Changed files</strong><span>{{ files.length }} files</span></div><div class="diff-summary"><span>+{{ totalAdditions }}</span><em>−{{ totalDeletions }}</em></div></div>
      <v-list bg-color="transparent" class="pa-2">
        <v-list-item v-for="file in files" :key="file.path" :active="selectedFilePath === file.path" rounded="lg" class="changed-file" @click="selectedFilePath = file.path">
          <template #prepend><span class="file-status" :class="`file-status--${file.status.toLowerCase()}`">{{ file.status }}</span></template>
          <v-list-item-title class="code-text">{{ file.path.split('/').pop() }}</v-list-item-title>
          <v-list-item-subtitle class="code-text">{{ file.path }}</v-list-item-subtitle>
          <template #append><span class="changed-file__stats"><b>+{{ file.additions }}</b><i>−{{ file.deletions }}</i></span></template>
        </v-list-item>
      </v-list>
    </aside>
    <section class="diff-panel">
      <div class="diff-panel__header"><div><v-icon icon="mdi-file-code-outline" size="18" /><span class="code-text">{{ selectedFile.path }}</span></div><div class="diff-summary"><span>+{{ selectedFile.additions }}</span><em>−{{ selectedFile.deletions }}</em></div></div>
      <div class="diff-viewer">
        <div v-for="(line, index) in selectedFile.lines" :key="index" class="diff-line" :class="`diff-line--${line.type}`">
          <span class="diff-line__number">{{ line.oldNo ?? '' }}</span><span class="diff-line__number">{{ line.newNo ?? '' }}</span><code>{{ line.type === 'add' ? '+' : line.type === 'remove' ? '-' : ' ' }}{{ line.text }}</code>
        </div>
      </div>
    </section>
  </v-card>
</template>
