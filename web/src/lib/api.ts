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

export interface ProviderInfo {
  name: string;
  model: string;
  apiBase: string;
  apiKey: string;
}

export interface AgentDetail {
  id: string;
  model: string;
  workspace: string;
  maxTokens?: number;
  temperature?: number;
  maxToolIterations?: number;
  thinking?: string;
  soul?: string;
  skills?: string[];
  tools?: string[];
}

export interface SkillInfo {
  name: string;
  description: string;
  location: string;
  type: string;
  // Set when the skill lives in an agent workspace; absent for shared skills.
  owner?: string;
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
  channel: string;
  chatId: string;
  message: string;
  enabled: boolean;
  lastRun?: string;
  nextRun?: string;
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
export interface ChatHistoryMessage {
  role: "user" | "assistant" | "tool";
  content?: string;
  toolCalls?: { id: string; name: string; arguments: string }[];
  name?: string;
  toolCallId?: string;
}

export async function getChatHistory(agentId: string, sessionId: string): Promise<ChatHistoryMessage[]> {
  const res = await fetch(`/api/chat/history?agentId=${encodeURIComponent(agentId)}&sessionId=${encodeURIComponent(sessionId)}`);
  return res.json();
}

export async function getChatSessions(agentId: string): Promise<{ id: string; preview: string }[]> {
  const res = await fetch(`/api/chat/sessions?agentId=${encodeURIComponent(agentId)}`);
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
export async function submitChat(
  agentId: string,
  sessionId: string,
  message: string,
): Promise<{ taskId: string; status: ChatTaskStatus }> {
  const res = await fetch("/api/chat/submit", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ agentId, sessionId, message }),
  });
  if (!res.ok) throw new Error(`submit failed: ${res.status}`);
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

export async function deleteSkill(name: string) {
  const res = await fetch(`/api/skills/${name}`, {
    method: "DELETE",
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
export async function getChannels(): Promise<ChannelInfo[]> {
  const res = await fetch("/api/channels");
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
