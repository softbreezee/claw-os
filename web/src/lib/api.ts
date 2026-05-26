export interface TaskInfo {
  id: string;
  agentId: string;
  chatKey: string;
  status: string;
  createdAt: string;
  duration?: number;
  error?: string;
}

export interface StatusResponse {
  configured: boolean;
  running: boolean;
  port: number;
  uptime: string;
  agents: AgentInfo[];
  channels: ChannelInfo[];
  provider: ProviderInfo;
  cronJobs?: number;
  plugins?: number;
}

export interface AgentInfo {
  id: string;
  model: string;
  workspace: string;
}

export interface ChannelInfo {
  type: string;
  botUsername: string;
  enabled?: boolean;
  status?: string;
}

// Detailed channel record returned by GET /api/channels — drives the
// Channels management UI. Tokens arrive masked ("ab12****wxyz") and
// can be re-submitted as-is to mean "keep the existing value".
export interface ChannelAccount {
  id: string;
  botToken: string;       // masked; submit unchanged to preserve
  agentId?: string;       // bound agent, derived from cfg.bindings
  myChatId?: string;      // user's "send to me" address on this bot
                          //   - telegram: numeric chat_id
                          //   - slack:    Uxxx user id (DM target)
                          //   - discord:  user id
                          //   - email:    email address (future)
                          //   - wechat:   openid (future)
  botUsername?: string;   // populated client-side after a successful test
}

export interface ChannelDetail {
  type: string;           // "telegram" | "discord" | "slack"
  enabled: boolean;
  botToken?: string;      // masked; legacy / fallback bot token
  appToken?: string;      // masked; slack-only
  accounts: ChannelAccount[];
  status?: string;        // "connected" | "disconnected"
}

export interface ProviderInfo {
  name: string;
  model: string;
  apiBase: string;
  apiKey: string;
}

export interface AgentDetail {
  id: string;
  model: string;
  embedModel?: string;
  workspace: string;
  maxTokens?: number;
  temperature?: number;
  maxToolIterations?: number;
  thinking?: string;
  soul?: string;
  skills?: string[];
  tools?: string[];
  /** All workspace files keyed by filename, e.g. "SOUL.md" → content. */
  files?: Record<string, string>;
}

export interface SkillInfo {
  name: string;
  description: string;
  location: string;
  type: string;
  builtin?: boolean;
  agents?: string[];
  owner?: string;
  kind?: string;
  tags?: string[];
}

export interface PluginInfo {
  id: string;
  type: string;
  version: string;
  status: string;
  enabled: boolean;
  config?: Record<string, unknown>;
}

export interface CronJobInfo {
  id: string;
  name: string;
  type: string;
  schedule: string;
  agentId: string;
  channel: string;       // "" / "web" / "telegram" / "slack" / "discord"
  accountId: string;     // bot account within the channel, empty for web/Inbox
  chatId: string;        // recipient address; empty for web/Inbox
  message: string;
  enabled: boolean;
  lastRun?: string;   // RFC3339; absent when never fired
  nextRun?: string;   // RFC3339; absent for malformed schedules
  createdAt?: string; // RFC3339; from store, used to display "added 2h ago"
}

export interface NotificationInfo {
  id: string;
  agentId: string;
  source: string;     // "cron" | "webhook" | "system" | "agent"
  sourceId: string;
  title: string;
  body: string;
  link: string;
  read: boolean;
  createdAt: string;  // RFC3339
}

export interface ModelCost {
  input: number;
  output: number;
  cacheRead: number;
  cacheWrite: number;
}

export interface ModelEntry {
  id: string;
  name: string;
  reasoning: boolean;
  input: string[];
  cost: ModelCost;
  contextWindow: number;
  maxTokens: number;
}

export interface ProviderData {
  apiKey: string;
  apiBase: string;
  apiType?: string;
  authType?: string;
  models?: ModelEntry[];
}

export interface ConfigResponse {
  providers: Record<string, ProviderData>;
  agents: {
    defaults: {
      model: string;
      maxTokens: number;
      temperature: number;
      maxToolIterations: number;
    };
    list: Array<{ id: string; model?: string }>;
  };
  channels: Record<string, { enabled: boolean; botToken?: string }>;
  storage: { type: string; dsn?: string };
  hooks: { enabled: boolean; token?: string; path?: string; port?: number };
  cronJobs?: Array<Record<string, unknown>>;
}

// Status
export async function getStatus(): Promise<StatusResponse> {
  const res = await fetch("/api/status");
  return res.json();
}

export async function getTasks(): Promise<TaskInfo[]> {
  const res = await fetch("/api/tasks");
  return res.json();
}

// Daemon restart — fires-and-forgets, caller should poll getStatus() afterward
export async function restartDaemon(): Promise<void> {
  await fetch("/api/daemon/restart", { method: "POST" }).catch(() => {});
}

// Poll /api/status until gateway is running again (up to timeoutMs)
export async function waitForGateway(timeoutMs = 15000): Promise<boolean> {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    await new Promise((r) => setTimeout(r, 800));
    try {
      const s = await getStatus();
      if (s.running) return true;
    } catch { /* still offline */ }
  }
  return false;
}

// Provider
export async function testProvider(config: { apiBase: string; apiKey: string; model: string; apiType?: string; authType?: string }) {
  const res = await fetch("/api/test-provider", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(config),
  });
  return res.json();
}

// Config
export async function saveConfig(config: Record<string, unknown>) {
  const res = await fetch("/api/save-config", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(config),
  });
  return res.json();
}

export async function getConfig(): Promise<ConfigResponse> {
  const res = await fetch("/api/config");
  return res.json();
}

export async function updateConfig(config: Record<string, unknown>) {
  const res = await fetch("/api/config", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(config),
  });
  return res.json();
}

// Chat
export interface ChatAttachment {
  type: "image"; // future: "document" | "audio"
  url: string;   // data: URL or absolute http(s)://; ready to <img src=…>
}

export interface ChatHistoryMessage {
  role: "user" | "assistant" | "tool";
  content?: string;
  attachments?: ChatAttachment[];
  toolCalls?: { id: string; name: string; arguments: string }[];
  name?: string;
  toolCallId?: string;
}

// fileURL builds the URL for the local-file proxy. Use kind="upload"
// for files in ~/.pawnix/uploads and kind="workspace" for files in
// the current agent's workspace directory.
export function fileURL(opts: { kind: "upload" | "workspace"; path: string; agentId?: string }): string {
  const params = new URLSearchParams({ kind: opts.kind, path: opts.path });
  if (opts.agentId) params.set("agentId", opts.agentId);
  return `/api/files?${params.toString()}`;
}

// openWorkspace asks the server to open the agent's workspace directory
// in the host OS file explorer (Finder, Explorer, xdg-open). Returns
// silently on success; throws on transport failure.
export async function openWorkspace(agentId: string): Promise<{ ok: boolean; path?: string; error?: string }> {
  const res = await fetch(`/api/workspace/open?agentId=${encodeURIComponent(agentId)}`, {
    method: "POST",
  });
  return res.json();
}

export async function getChatHistory(agentId: string, sessionId: string): Promise<ChatHistoryMessage[]> {
  const res = await fetch(`/api/chat/history?agentId=${encodeURIComponent(agentId)}&sessionId=${encodeURIComponent(sessionId)}`);
  return res.json();
}

export async function getChatSessions(agentId: string): Promise<{ id: string; preview: string }[]> {
  const res = await fetch(`/api/chat/sessions?agentId=${encodeURIComponent(agentId)}`);
  return res.json();
}

export interface ExternalSession {
  id: string;       // "discord:15040..."
  channel: string;  // "discord" | "telegram"
  chatId: string;   // channel-specific chat/group ID
  preview: string;  // first user message
}

export async function getExternalSessions(agentId: string): Promise<ExternalSession[]> {
  const res = await fetch(`/api/chat/external-sessions?agentId=${encodeURIComponent(agentId)}`);
  return res.json();
}

export async function getExternalHistory(agentId: string, channel: string, chatId: string): Promise<ChatHistoryMessage[]> {
  const res = await fetch(`/api/chat/external-history?agentId=${encodeURIComponent(agentId)}&channel=${encodeURIComponent(channel)}&chatId=${encodeURIComponent(chatId)}`);
  return res.json();
}

export async function sendChat(agentId: string, sessionId: string, message: string): Promise<{ response: string }> {
  const res = await fetch("/api/chat", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ agentId, sessionId, message }),
  });
  return res.json();
}

// ChatStreamEvent is a single SSE event from the chat task subsystem.
// task_* events mark lifecycle transitions; content / tool_call / tool_result
// stream agent output.
export interface ChatStreamEvent {
  type:
    | "content"
    | "tool_call"
    | "tool_result"
    | "done"
    | "task_pending"
    | "task_running"
    | "task_done"
    | "task_error"
    | "task_cancelled";
  data?: Record<string, string>;
  // Monotonic per-task sequence – used for resumable subscriptions.
  seq?: number;
}

// ─────────────────────────────────────────────────────────────────────────────
// Async chat task API (submit → subscribe → cancel).
// ─────────────────────────────────────────────────────────────────────────────

export type ChatTaskStatus = "pending" | "running" | "done" | "failed" | "cancelled";

export interface ChatTaskRecord {
  id: string;
  agentId: string;
  sessionKey: string;
  status: ChatTaskStatus;
  message: string;
  result?: string;
  error?: string;
  createdAt: string;
  startedAt?: string;
  doneAt?: string;
  updatedAt: string;
}

// submitChat enqueues a new chat task and returns its taskId.
// Does NOT wait for the agent to respond; subscribe via subscribeTaskEvents.
// model is optional: pass "" / undefined to let the agent's configured
// default win, or pass a specific model name to override it for this
// single message only (the agent's persistent config is unchanged).
// submitChat enqueues a new chat task. When `files` is non-empty the
// request is sent as multipart/form-data so binaries (images today,
// other types later) can ride along with the text. Otherwise we fall
// back to the JSON path to keep request payloads tiny for plain text.
//
// model is optional ("" / undefined → the agent's configured default).
export async function submitChat(
  agentId: string,
  sessionId: string,
  message: string,
  model?: string,
  files?: File[],
  deliverToChannel?: string,
  deliverToChatId?: string,
): Promise<{ taskId: string; status: ChatTaskStatus }> {
  let res: Response;

  if (files && files.length > 0) {
    const form = new FormData();
    form.append("agentId", agentId);
    form.append("sessionId", sessionId);
    form.append("message", message);
    if (model) form.append("model", model);
    if (deliverToChannel) form.append("deliverToChannel", deliverToChannel);
    if (deliverToChatId) form.append("deliverToChatId", deliverToChatId);
    for (const f of files) {
      // The backend reads from r.MultipartForm.File["files"], so the
      // field name must literally be "files" for every file.
      form.append("files", f, f.name);
    }
    res = await fetch("/api/chat/submit", {
      method: "POST",
      // Do NOT set Content-Type — the browser fills in the multipart
      // boundary parameter when given a FormData body.
      body: form,
    });
  } else {
    res = await fetch("/api/chat/submit", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        agentId,
        sessionId,
        message,
        ...(model ? { model } : {}),
        ...(deliverToChannel ? { deliverToChannel } : {}),
        ...(deliverToChatId ? { deliverToChatId } : {}),
      }),
    });
  }

  if (!res.ok) {
    let detail = "";
    try {
      const body = await res.json();
      if (body?.error) detail = `: ${body.error}`;
    } catch {
      /* not JSON; ignore */
    }
    throw new Error(`submit failed: ${res.status}${detail}`);
  }
  return res.json();
}

// subscribeTaskEvents opens an SSE connection for a task. Resolves when
// the bus closes the topic (terminal event received) or the abort signal
// fires. Safe to call repeatedly for the same taskId – each call gets
// its own subscription.
//
// PR3: pass `afterSeq` to resume from a known position. The backend
// replays buffered events with seq > afterSeq before subscribing to live
// updates, so a reconnecting client never misses events (as long as the
// gap is within the server-side ring buffer of ~256 events per task).
export async function subscribeTaskEvents(
  taskId: string,
  onEvent: (evt: ChatStreamEvent) => void,
  signal?: AbortSignal,
  afterSeq?: number,
): Promise<void> {
  const url = afterSeq && afterSeq > 0
    ? `/api/chat/tasks/${encodeURIComponent(taskId)}/events?after=${afterSeq}`
    : `/api/chat/tasks/${encodeURIComponent(taskId)}/events`;
  const res = await fetch(url, { signal });
  if (!res.ok || !res.body) {
    throw new Error(`subscribe failed: ${res.status}`);
  }

  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";

  // Propagate abort to the reader so the in-flight read() rejects promptly.
  const onAbort = () => { reader.cancel().catch(() => {}); };
  if (signal) {
    if (signal.aborted) {
      await reader.cancel().catch(() => {});
      throw new DOMException("Aborted", "AbortError");
    }
    signal.addEventListener("abort", onAbort);
  }

  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });

      const lines = buffer.split("\n");
      buffer = lines.pop() || "";

      for (const line of lines) {
        if (!line.startsWith("data: ")) continue;
        try {
          const evt = JSON.parse(line.slice(6)) as ChatStreamEvent;
          onEvent(evt);
        } catch { /* skip */ }
      }
    }
  } finally {
    if (signal) signal.removeEventListener("abort", onAbort);
  }
}

// cancelTask asks the backend to cancel an in-flight task. Idempotent –
// returns ok even for already-terminal tasks.
export async function cancelTask(taskId: string): Promise<{ ok: boolean }> {
  const res = await fetch(`/api/chat/tasks/${encodeURIComponent(taskId)}/cancel`, {
    method: "POST",
  });
  if (!res.ok) throw new Error(`cancel failed: ${res.status}`);
  return res.json();
}

export async function getTask(taskId: string): Promise<ChatTaskRecord> {
  const res = await fetch(`/api/chat/tasks/${encodeURIComponent(taskId)}`);
  if (!res.ok) throw new Error(`get task failed: ${res.status}`);
  return res.json();
}

export async function listTasks(filters?: {
  agentId?: string;
  sessionKey?: string;
  status?: ChatTaskStatus;
  limit?: number;
  offset?: number;
}): Promise<{ tasks: ChatTaskRecord[]; total: number }> {
  const params = new URLSearchParams();
  if (filters?.agentId) params.append("agentId", filters.agentId);
  if (filters?.sessionKey) params.append("sessionKey", filters.sessionKey);
  if (filters?.status) params.append("status", filters.status);
  if (filters?.limit) params.append("limit", String(filters.limit));
  if (filters?.offset) params.append("offset", String(filters.offset));
  const res = await fetch(`/api/chat/tasks?${params}`);
  if (!res.ok) throw new Error(`list tasks failed: ${res.status}`);
  return res.json();
}

// ─────────────────────────────────────────────────────────────────────────────

export async function deleteChatSession(agentId: string, sessionId: string): Promise<{ ok: boolean }> {
  const res = await fetch(`/api/chat/sessions?agentId=${encodeURIComponent(agentId)}&sessionId=${encodeURIComponent(sessionId)}`, {
    method: "DELETE",
  });
  return res.json();
}

// Agents
export async function getAgents(): Promise<AgentDetail[]> {
  const res = await fetch("/api/agents");
  return res.json();
}

export async function createAgent(agent: Partial<AgentDetail>) {
  const res = await fetch("/api/agents", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(agent),
  });
  return res.json();
}

export async function updateAgent(id: string, agent: Partial<AgentDetail>) {
  const res = await fetch(`/api/agents/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(agent),
  });
  return res.json();
}

export async function deleteAgent(id: string) {
  const res = await fetch(`/api/agents/${id}`, {
    method: "DELETE",
  });
  return res.json();
}

// Skills
export async function getSkills(): Promise<SkillInfo[]> {
  const res = await fetch("/api/skills");
  return res.json();
}

export interface SkillDetail {
  name: string;
  content: string;
  location: string;
}

// getSkill returns the full SKILL.md content for a single skill,
// used by the Skills page detail dialog. Skill name with slashes
// or other URL-unsafe chars is encoded.
export async function getSkill(name: string): Promise<SkillDetail> {
  const res = await fetch(`/api/skills/${encodeURIComponent(name)}`);
  if (!res.ok) throw new Error(`fetch skill failed: ${res.status}`);
  return res.json();
}

export async function deleteSkill(name: string) {
  const res = await fetch(`/api/skills/${name}`, {
    method: "DELETE",
  });
  return res.json();
}

// Move a skill between writable layers.
//   scope: "user"        – ~/.pawnix/skills/ (shared)
//   scope: "agent:<id>"  – ~/.pawnix/agents/<id>/agent/skills/
// Builtin skills cannot be moved (immutable; create a same-named override
// in user/ or an agent workspace to take precedence at runtime).
export async function updateSkillScope(name: string, agents: string[], tags?: string[]): Promise<{ ok: boolean; error?: string }> {
  const body: Record<string, unknown> = { agents };
  if (tags !== undefined) body.tags = tags;
  const res = await fetch(`/api/skills/${encodeURIComponent(name)}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  return res.json();
}

export async function moveSkill(name: string, scope: string): Promise<{ ok: boolean; location?: string; error?: string }> {
  const res = await fetch(`/api/skills/${encodeURIComponent(name)}/move`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ scope }),
  });
  return res.json();
}

// Plugins
export async function getPlugins(): Promise<PluginInfo[]> {
  const res = await fetch("/api/plugins");
  return res.json();
}

export async function updatePlugin(id: string, data: Partial<PluginInfo>) {
  const res = await fetch(`/api/plugins/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  return res.json();
}

// Channels
//
// NOTE: GET /api/channels now returns the *detailed* shape (with
// accounts + bindings) used by the Channels management page. The
// Status endpoint still returns the lightweight ChannelInfo shape
// for the Overview / sidebar pills. We expose two helpers so each
// caller picks the right shape; getChannels() preserves the old
// signature so existing callers (Overview pill, etc.) still compile,
// while getChannelsDetailed() returns the richer shape.
export async function getChannels(): Promise<ChannelInfo[]> {
  const res = await fetch("/api/channels");
  const data = (await res.json()) as ChannelDetail[] | null;
  if (!Array.isArray(data)) return [];
  // Project to the legacy shape so old callers keep working unchanged.
  return data.map((c) => ({
    type: c.type,
    botUsername: c.accounts?.[0]?.botUsername ?? "",
    enabled: c.enabled,
    status: c.status,
  }));
}

export async function getChannelsDetailed(): Promise<ChannelDetail[]> {
  const res = await fetch("/api/channels");
  const data = await res.json();
  return Array.isArray(data) ? data : [];
}

export interface ChannelUpdateBody {
  enabled: boolean;
  botToken?: string;
  appToken?: string;
  accounts: Array<{
    id: string;
    botToken: string;     // pass masked value to keep, or new value to update
    agentId?: string;
    myChatId?: string;    // user's "send to me" address (telegram chat_id, etc)
  }>;
}

export async function updateChannel(
  type: string,
  body: ChannelUpdateBody,
): Promise<{ ok: boolean; needsRestart?: boolean; error?: string }> {
  const res = await fetch(`/api/channels/${encodeURIComponent(type)}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  return res.json();
}

export async function deleteChannel(
  type: string,
): Promise<{ ok: boolean; needsRestart?: boolean; error?: string }> {
  const res = await fetch(`/api/channels/${encodeURIComponent(type)}`, {
    method: "DELETE",
  });
  return res.json();
}

export async function testChannel(
  type: string,
  botToken: string,
): Promise<{ ok: boolean; botUsername?: string; firstName?: string; error?: string }> {
  const res = await fetch("/api/channels/test", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ type, botToken }),
  });
  return res.json();
}

// Cron Jobs
export async function getCronJobs(): Promise<CronJobInfo[]> {
  const res = await fetch("/api/cron");
  return res.json();
}

export async function createCronJob(job: Partial<CronJobInfo>) {
  const res = await fetch("/api/cron", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(job),
  });
  return res.json();
}

export async function updateCronJob(id: string, job: Partial<CronJobInfo>) {
  const res = await fetch(`/api/cron/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(job),
  });
  return res.json();
}

export async function deleteCronJob(id: string) {
  const res = await fetch(`/api/cron/${id}`, {
    method: "DELETE",
  });
  return res.json();
}

// runCronJobNow nudges the scheduler to fire the job on its next poll
// cycle (within ~60s). Useful for smoke-testing a freshly created job
// without waiting for its real schedule.
export async function runCronJobNow(id: string) {
  const res = await fetch(`/api/cron/${id}/run`, {
    method: "POST",
  });
  return res.json();
}

// --- Notifications (the OS-level inbox) ---

export interface NotificationListParams {
  unreadOnly?: boolean;
  source?: string;
  agentId?: string;
  limit?: number;
  offset?: number;
}

export async function getNotifications(params: NotificationListParams = {}): Promise<NotificationInfo[]> {
  const q = new URLSearchParams();
  if (params.unreadOnly) q.set("unreadOnly", "true");
  if (params.source) q.set("source", params.source);
  if (params.agentId) q.set("agentId", params.agentId);
  if (params.limit) q.set("limit", String(params.limit));
  if (params.offset) q.set("offset", String(params.offset));
  const qs = q.toString();
  const res = await fetch(`/api/notifications${qs ? `?${qs}` : ""}`);
  if (!res.ok) return [];
  return res.json();
}

export async function getUnreadNotificationCount(): Promise<{ count: number }> {
  const res = await fetch("/api/notifications/unread-count");
  if (!res.ok) return { count: 0 };
  return res.json();
}

export async function markNotificationRead(id: string, read = true) {
  const res = await fetch(`/api/notifications/${id}/read`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ read }),
  });
  return res.json();
}

export async function markAllNotificationsRead() {
  const res = await fetch("/api/notifications/read-all", { method: "POST" });
  return res.json();
}

export async function deleteNotification(id: string) {
  const res = await fetch(`/api/notifications/${id}`, { method: "DELETE" });
  return res.json();
}

// Apps
export interface AppEntry {
  name: string;
  url: string;
  description?: string;
}

export async function getApps(): Promise<AppEntry[]> {
  const res = await fetch("/api/apps");
  return res.json();
}

export async function createApp(app: AppEntry) {
  const res = await fetch("/api/apps", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(app),
  });
  return res.json();
}

export async function updateApp(oldName: string, app: AppEntry) {
  const res = await fetch(`/api/apps/${encodeURIComponent(oldName)}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(app),
  });
  return res.json();
}

export async function deleteApp(name: string) {
  const res = await fetch(`/api/apps/${encodeURIComponent(name)}`, {
    method: "DELETE",
  });
  return res.json();
}

// Session Context Info
export interface SessionContextInfo {
  currentTokens: number;
  contextWindow: number;
  softThreshold: number;
  hardThreshold: number;
  messageCount: number;
  compactionCount: number;
  modelId: string;
}

export async function getSessionContextInfo(agentId: string, sessionId: string): Promise<SessionContextInfo | null> {
  try {
    const res = await fetch(
      `/api/chat/context-info?agentId=${encodeURIComponent(agentId)}&sessionId=${encodeURIComponent(sessionId)}`
    );
    if (!res.ok) return null;
    return res.json();
  } catch {
    return null;
  }
}

// System Prompt Preview — labelled breakdown of the agent's static system prompt.
export interface SystemPromptSectionInfo {
  name: string;
  content: string;
  tokens: number;
}

export interface SystemPromptInfo {
  sections: SystemPromptSectionInfo[];
  totalTokens: number;
  modelId: string;
}

export async function getSessionSystemPrompt(agentId: string): Promise<SystemPromptInfo | null> {
  try {
    const res = await fetch(`/api/chat/system-prompt?agentId=${encodeURIComponent(agentId)}`);
    if (!res.ok) return null;
    return res.json();
  } catch {
    return null;
  }
}

// Model Catalog
export interface ModelCatalogModelInfo {
  contextWindow: number;
  softThreshold: number;
  hardThreshold: number;
  description: string;
}

export interface ModelCatalog {
  models: Record<string, ModelCatalogModelInfo>;
}

export async function getModelCatalog(): Promise<ModelCatalog> {
  const res = await fetch("/api/model-catalog");
  return res.json();
}

export async function saveModelCatalog(catalog: ModelCatalog): Promise<{ ok: boolean; message?: string; error?: string }> {
  const res = await fetch("/api/model-catalog", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(catalog),
  });
  return res.json();
}

export async function reloadModelCatalog(): Promise<{ ok: boolean; message?: string; error?: string }> {
  const res = await fetch("/api/model-catalog/reload", {
    method: "POST",
  });
  return res.json();
}

// ── MCP Servers ──

export interface McpServerInfo {
  name: string;
  type: string;   // "stdio" | "http"
  url?: string;
  command?: string;
  args?: string[];
  status: string; // "connected" | "disconnected" | "unknown"
  toolCount?: number;
}

export async function getMcpServers(): Promise<McpServerInfo[]> {
  const res = await fetch("/api/mcp");
  return res.json();
}

export async function createMcpServer(data: Record<string, unknown>): Promise<{ ok: boolean; name?: string; error?: string }> {
  const res = await fetch("/api/mcp", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  return res.json();
}

export async function updateMcpServer(name: string, data: Record<string, unknown>): Promise<{ ok: boolean; error?: string }> {
  const res = await fetch(`/api/mcp/${encodeURIComponent(name)}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  return res.json();
}

export async function deleteMcpServer(name: string): Promise<{ ok: boolean; error?: string }> {
  const res = await fetch(`/api/mcp/${encodeURIComponent(name)}`, {
    method: "DELETE",
  });
  return res.json();
}

export async function rebuildEmbeddings(agentId: string): Promise<{ ok: boolean; facts: number; inserted: number; message?: string; error?: string }> {
  const res = await fetch(`/api/agents/${encodeURIComponent(agentId)}/rebuild-embeddings`, {
    method: "POST",
  });
  return res.json();
}
