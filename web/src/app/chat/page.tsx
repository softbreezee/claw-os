"use client";

import React, { useEffect, useState, useRef, useCallback, useMemo } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  getStatus,
  getChatHistory,
  getChatSessions,
  getSkills,
  submitChat,
  subscribeTaskEvents,
  cancelTask,
  getTask,
  deleteChatSession,
  fileURL,
  openWorkspace,
  getSessionContextInfo,
  getSessionSystemPrompt,
  getExternalSessions,
  getExternalHistory,
  sendToChannel,
  type AgentInfo,
  type ExternalSession,
  type ChatAttachment,
  type ChatHistoryMessage,
  type ChatStreamEvent,
  type SkillInfo,
  type SessionContextInfo,
  type SystemPromptInfo,
} from "@/lib/api";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Bot,
  Send,
  Copy,
  Check,
  SquarePen,
  MessageSquare,
  Wrench,
  ChevronDown,
  ChevronRight,
  Trash2,
  Square,
  Zap,
  CheckCircle2,
  AlertCircle,
  Clock,
  Paperclip,
  X,
  FileText,
  FolderOpen,
  Sparkles,
  Eye,
  FileCode,
  Radio,
} from "lucide-react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

interface ToolCall {
  id: string;
  name: string;
  arguments: string;
  result?: string;
  error?: boolean;
}

// MessageAttachment is what the chat UI renders alongside a message.
// It's a structural superset of api.ChatAttachment plus a `name` for
// the live-upload case (object URLs don't carry a filename).
interface MessageAttachment {
  type: "image" | "file";
  url: string;
  name?: string;
}

interface ChatMessage {
  id: string;
  role: "user" | "agent" | "tool-group";
  content: string;
  timestamp: number;
  toolCalls?: ToolCall[];
  attachments?: MessageAttachment[];
}

interface ChatSession {
  id: string;
  preview: string;
}

// Per-(agent, session) runtime that survives tab switches.
// Without this, switching agents while one is streaming would carry the
// `sending=true` lock to the new tab and silently swallow user messages.
interface RuntimeState {
  sending: boolean;
  messages: ChatMessage[];
  abort?: () => void;
  // PR2: server-side task ID for the in-flight request. Set when
  // submitChat resolves; cleared on terminal events. Also used to
  // re-subscribe after a tab switch (see useEffect below).
  taskId?: string;
  // PR3: highest seq we've successfully processed for the current task.
  // Sent as ?after= when re-subscribing so we don't replay events the
  // UI has already absorbed.
  lastSeq?: number;
  // Streaming accumulators – kept here (not in component locals) so the
  // background SSE handler can keep writing into the right tab even when
  // the user is looking at a different one.
  curGroupId: string;
  curCalls: ToolCall[];
  curContent: string;
}

function makeRuntime(initial: ChatMessage[] = []): RuntimeState {
  return {
    sending: false,
    messages: initial,
    abort: undefined,
    curGroupId: "",
    curCalls: [],
    curContent: "",
  };
}

function rtKey(agentId: string, sessionId: string): string {
  return `${agentId}::${sessionId}`;
}

// PR3: persist the (agent, session) → taskId map across page reloads so a
// background task survives a refresh / browser restart. We deliberately
// only persist the taskId; the messages array is reloaded from server
// history on next mount, and the resume useEffect re-subscribes via the
// taskId. This keeps localStorage tiny (one entry, ~50 bytes per task).
const TASK_STORAGE_KEY = "pawnix.chat.tasks.v1";

function loadPersistedTaskIds(): Record<string, string> {
  if (typeof window === "undefined") return {};
  try {
    const raw = window.localStorage.getItem(TASK_STORAGE_KEY);
    if (!raw) return {};
    const parsed = JSON.parse(raw);
    return typeof parsed === "object" && parsed !== null ? parsed : {};
  } catch {
    return {};
  }
}

function persistTaskId(key: string, taskId: string | undefined) {
  if (typeof window === "undefined") return;
  try {
    const cur = loadPersistedTaskIds();
    if (taskId) {
      cur[key] = taskId;
    } else {
      delete cur[key];
    }
    window.localStorage.setItem(TASK_STORAGE_KEY, JSON.stringify(cur));
  } catch {
    // Storage quota / private mode – best-effort, drop silently.
  }
}

function generateSessionId() {
  return `s-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
}

function parseExternID(id: string): [string, string] {
  const idx = id.indexOf(":");
  if (idx < 0) return ["", ""];
  return [id.slice(0, idx), id.slice(idx + 1)];
}

function relativeTime(ts: number): string {
  if (!ts) return "";
  const diff = Math.floor((Date.now() - ts) / 1000);
  if (diff < 60) return "just now";
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  return new Date(ts).toLocaleDateString([], { month: "short", day: "numeric" });
}

// chatAttachmentsToMsg converts API-shape attachments into the slightly
// richer in-memory shape. The current backend only emits images, but
// the UI is type-permissive (`type: "image" | "file"`) so a future
// backend that adds documents won't require touching this code path.
function chatAttachmentsToMsg(atts?: ChatAttachment[]): MessageAttachment[] | undefined {
  if (!atts || atts.length === 0) return undefined;
  return atts.map((a) => ({ type: a.type === "image" ? "image" : "file", url: a.url }));
}

function CodeBlock({ children, className }: { children: string; className?: string }) {
  const [copied, setCopied] = useState(false);
  const match = /language-(\w+)/.exec(className || "");
  const handleCopy = () => {
    navigator.clipboard.writeText(children);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };
  return (
    <div className="relative group my-2">
      {match && (
        <div className="flex items-center justify-between px-4 py-1.5 bg-muted/50 rounded-t-md border border-border border-b-0 text-[11px] text-muted-foreground font-mono">
          <span>{match[1]}</span>
          <button
            onClick={handleCopy}
            className="flex items-center gap-1 text-muted-foreground/60 hover:text-foreground transition-colors"
          >
            {copied ? <Check className="h-3 w-3 text-emerald-500" /> : <Copy className="h-3 w-3" />}
            <span className="text-[10px]">{copied ? "Copied" : "Copy"}</span>
          </button>
        </div>
      )}
      <pre className={`${match ? "rounded-t-none" : "rounded-md"} bg-muted/30 border border-border p-3 overflow-x-auto text-[13px] leading-relaxed`}>
        <code className={className}>{children}</code>
      </pre>
    </div>
  );
}

function buildExternalMessages(history: ChatHistoryMessage[]): ChatMessage[] {
  return history
    .filter((h) => h.role !== "tool") // skip raw tool outputs in external view
    .map((h, i) => {
      const msg: ChatMessage = {
        id: `h-${i}`,
        role: h.role === "assistant" ? "agent" : "user",
        content: h.content || "",
        timestamp: 0,
      };
      if (h.role === "assistant" && h.toolCalls && h.toolCalls.length > 0) {
        msg.role = "tool-group";
        msg.toolCalls = h.toolCalls.map((tc) => ({ ...tc, result: undefined }));
      }
      return msg;
    });
}

function buildChatMessages(history: ChatHistoryMessage[]): ChatMessage[] {
  const msgs: ChatMessage[] = [];
  let i = 0;
  while (i < history.length) {
    const h = history[i];
    if (h.role === "user") {
      msgs.push({
        id: `h-${i}`,
        role: "user",
        content: h.content || "",
        timestamp: 0,
        attachments: chatAttachmentsToMsg(h.attachments),
      });
      i++;
    } else if (h.role === "assistant" && h.toolCalls && h.toolCalls.length > 0) {
      const calls: ToolCall[] = h.toolCalls.map((tc) => ({ ...tc, result: undefined }));
      i++;
      while (i < history.length && history[i].role === "tool") {
        const toolMsg = history[i];
        const call = calls.find((c) => c.id === toolMsg.toolCallId);
        if (call) call.result = toolMsg.content;
        i++;
      }
      msgs.push({ id: `h-tool-${i}`, role: "tool-group", content: h.content || "", timestamp: 0, toolCalls: calls });
      if (i < history.length && history[i].role === "assistant" && history[i].content) {
        msgs.push({ id: `h-${i}`, role: "agent", content: history[i].content || "", timestamp: 0 });
        i++;
      }
    } else if (h.role === "assistant") {
      msgs.push({ id: `h-${i}`, role: "agent", content: h.content || "", timestamp: 0 });
      i++;
    } else {
      i++;
    }
  }
  return msgs;
}

const STARTER_PROMPTS = [
  "What can you help me with?",
  "Summarize the files in my workspace",
  "Show me what skills you have available",
  "What tools do you have access to?",
];

export default function ChatPage() {
  const [agents, setAgents] = useState<AgentInfo[]>([]);
  const [selectedAgent, setSelectedAgent] = useState<string>("");
  const [sessionId, setSessionId] = useState<string>(() => generateSessionId());
  const [sessions, setSessions] = useState<ChatSession[]>([]);
  const [externSessions, setExternSessions] = useState<ExternalSession[]>([]);
  const [externChannel, setExternChannel] = useState<string | null>(null); // "discord:chatId" when viewing external
  // All per-(agent, session) runtime state lives in this map. The currently
  // visible tab is just a selector view over it. This is the structural fix
  // that allows multiple agents to stream concurrently without their states
  // bleeding into each other.
  //
  // On first mount we hydrate from localStorage so a background task that
  // was started in a previous session can be re-attached. Only the taskId
  // is persisted; messages are loaded from server history.
  const [runtimes, setRuntimes] = useState<Record<string, RuntimeState>>(() => {
    const persisted = loadPersistedTaskIds();
    const init: Record<string, RuntimeState> = {};
    for (const [k, taskId] of Object.entries(persisted)) {
      init[k] = { ...makeRuntime(), taskId };
    }
    return init;
  });
  const [input, setInput] = useState("");
  const [copiedId, setCopiedId] = useState<string | null>(null);

  // Pending file attachments for the next send. Cleared after a
  // successful submit (or never set if the user only sends text).
  // We keep object URLs alongside so previews don't have to re-encode
  // the bytes; remember to URL.revokeObjectURL on remove and unmount
  // to avoid leaking memory across long sessions.
  const [pendingFiles, setPendingFiles] = useState<{
    id: string;
    file: File;
    previewURL: string; // object URL for image thumbnails; "" for non-image
  }[]>([]);
  const [isDragging, setIsDragging] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  // Per-agent ad-hoc model override. null/"" means "use the agent's
  // configured default". Resets on agent switch so the user always sees
  // the agent's own default after navigating away and back.
  const [modelOverride, setModelOverride] = useState<string>("");
  // Skills available for the launcher dropdown. Refreshed on mount;
  // we don't poll because installing a skill is a deliberate user
  // action — the next page navigation will pick up the new state.
  const [skills, setSkills] = useState<SkillInfo[]>([]);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  // Context window usage info — polled every 10s and after each message round-trip.
  const [contextInfo, setContextInfo] = useState<SessionContextInfo | null>(null);

  // System prompt preview — fetched lazily when the user opens the modal so we
  // don't pay the (small) cost on every page load. Cleared on agent switch so
  // the cached preview never lies about which agent it represents.
  const [systemPromptInfo, setSystemPromptInfo] = useState<SystemPromptInfo | null>(null);
  const [showSystemPromptModal, setShowSystemPromptModal] = useState(false);
  const [systemPromptLoading, setSystemPromptLoading] = useState(false);

  // Mutable mirror of `runtimes` so the SSE callback (long-lived closure)
  // can read the latest state without stale-closure bugs and keep the
  // streaming accumulators (curCalls / curContent) outside React state.
  const runtimesRef = useRef<Record<string, RuntimeState>>({});
  runtimesRef.current = runtimes;

  const currentKey = rtKey(selectedAgent, sessionId);
  const current = runtimes[currentKey] ?? makeRuntime();
  const messages = current.messages;
  const sending = current.sending;

  // Helper: mutate the runtime for a specific key, then trigger a re-render
  // by rebuilding the top-level object reference. Also syncs the taskId to
  // localStorage when it changes so background tasks survive page reloads.
  const updateRuntime = useCallback((key: string, updater: (r: RuntimeState) => RuntimeState) => {
    setRuntimes((prev) => {
      const cur = prev[key] ?? makeRuntime();
      const next = updater(cur);
      if (cur.taskId !== next.taskId) {
        persistTaskId(key, next.taskId);
      }
      return { ...prev, [key]: next };
    });
  }, []);

  useEffect(() => {
    getStatus()
      .then((status) => {
        if (status.agents?.length > 0) {
          setAgents(status.agents);
          setSelectedAgent(status.agents[0].id);
        }
      })
      .catch(() => {});
    // Fetch skills once on mount for the launcher button. Cheap call,
    // doesn't grow with conversation length.
    getSkills().then(setSkills).catch(() => setSkills([]));
  }, []);

  // Group skills for the launcher dropdown: orchestrator-style first
  // (suite > protocol), then atomic. Within each group sort by name.
  const skillGroups = useMemo(() => {
    const suites: SkillInfo[] = [];
    const protocols: SkillInfo[] = [];
    const atoms: SkillInfo[] = [];
    for (const s of skills) {
      if (s.kind === "suite") suites.push(s);
      else if (s.kind === "protocol") protocols.push(s);
      else atoms.push(s);
    }
    const byName = (a: SkillInfo, b: SkillInfo) => a.name.localeCompare(b.name);
    suites.sort(byName);
    protocols.sort(byName);
    atoms.sort(byName);
    return { suites, protocols, atoms };
  }, [skills]);

  // pickSkill prepends a "use this skill" hint to the input box and
  // focuses the textarea. We deliberately do NOT auto-send — the user
  // typically wants to add target / context after picking the skill
  // (e.g. "use signal-agent-v8: analyse Nebius").
  const pickSkill = useCallback((skill: SkillInfo) => {
    const cn = (s: string) => s.replace(/`/g, "");
    const prefix = `Use the \`${cn(skill.name)}\` ${
      skill.kind === "suite" ? "orchestrator skill (load it first, then follow the routing table)"
      : skill.kind === "protocol" ? "protocol skill (load it first, then follow the steps)"
      : "skill"
    }: `;
    setInput((cur) => prefix + (cur.trim() ? cur : ""));
    // Focus + place cursor at end of the prepended hint so user can
    // immediately start typing the task.
    setTimeout(() => {
      const el = textareaRef.current;
      if (el) {
        el.focus();
        el.setSelectionRange(prefix.length, prefix.length + (input.trim() ? input.length : 0));
      }
    }, 0);
  }, [input]);

  // Reset model override whenever the user picks a different agent.
  // Each agent has its own configured default; the override is meant
  // to be a per-message tweak within one agent context.
  useEffect(() => {
    setModelOverride("");
    // Drop any cached system-prompt preview — it belonged to the old agent
    // and would be misleading if the user opened the modal again.
    setSystemPromptInfo(null);
  }, [selectedAgent]);

  // Fetch the system-prompt breakdown on demand. We refresh every time the
  // modal is opened so workspace edits to AGENTS.md / SOUL.md / skills are
  // reflected without the user needing to reload the page.
  const openSystemPromptModal = useCallback(async () => {
    if (!selectedAgent) return;
    setShowSystemPromptModal(true);
    setSystemPromptLoading(true);
    try {
      const info = await getSessionSystemPrompt(selectedAgent);
      setSystemPromptInfo(info);
    } finally {
      setSystemPromptLoading(false);
    }
  }, [selectedAgent]);

  const loadSessions = useCallback((agentId: string) => {
    getChatSessions(agentId)
      .then((list) => setSessions(list || []))
      .catch(() => setSessions([]));
    getExternalSessions(agentId)
      .then((list) => setExternSessions(list || []))
      .catch(() => setExternSessions([]));
  }, []);

  useEffect(() => {
    if (!selectedAgent) return;
    loadSessions(selectedAgent);
  }, [selectedAgent, loadSessions]);

  // When an external channel is selected, load its history and use its
  // channel/chatId as the session key (for message routing).
  useEffect(() => {
    if (!externChannel || !selectedAgent) return;
    const [ch, chatID] = parseExternID(externChannel);
    if (!ch || !chatID) return;
    // Use a mirror session ID ("ext_discord_...") so the web chat
    // doesn't pollute the real Discord session's history. The real
    // session is loaded via getExternalHistory for display only.
    const mirrorSid = `ext_${externChannel.replace(":", "_")}`;
    const displayKey = rtKey(selectedAgent, mirrorSid);
    setSessionId(mirrorSid);
    // Load both real Discord history and the mirror session's messages
    // so Web UI messages that haven't synced to Discord are still visible.
    const refresh = () => {
      getExternalHistory(selectedAgent, ch, chatID)
        .then((extMsgs) => {
          const msgs = (!extMsgs || extMsgs.length === 0) ? [] : buildExternalMessages(extMsgs);
          getChatHistory(selectedAgent, mirrorSid)
            .then((mirrorMsgs) => {
              if (mirrorMsgs && mirrorMsgs.length > 0) {
                const extra = buildExternalMessages(mirrorMsgs);
                for (const m of extra) {
                  if (!msgs.find((x) => x.role === m.role && x.content === m.content)) {
                    msgs.push(m);
                  }
                }
              }
            })
            .catch(() => {})
            .finally(() => updateRuntime(displayKey, (s) => ({ ...s, messages: msgs })));
        })
        .catch(() => updateRuntime(displayKey, (s) => ({ ...s, messages: [] })));
    };
    refresh();
    const iv = setInterval(refresh, 30000);
    return () => clearInterval(iv);
  }, [externChannel, selectedAgent]);

  // Poll context info every 10s; also refresh immediately when (agent, session) changes
  // or when a send completes (sending flips false→true→false).
  useEffect(() => {
    if (!selectedAgent || !sessionId) return;
    let alive = true;
    const refresh = () => {
      getSessionContextInfo(selectedAgent, sessionId).then((info) => {
        if (alive) setContextInfo(info);
      });
    };
    refresh();
    const timer = setInterval(refresh, 10_000);
    return () => { alive = false; clearInterval(timer); };
  }, [selectedAgent, sessionId, sending]); // re-trigger when sending flips so we get fresh stats after each turn

  // Load history when (agent, session) changes – but only if there is no
  // live runtime already (i.e. don't clobber an in-flight stream when the
  // user navigates back to a tab that is still working).
  useEffect(() => {
    if (!selectedAgent || !sessionId) return;
    // External sessions (discord:...) loaded by external effect;
    // ext_discord_... sessions now handled by updated handleChatHistory API.
    if (sessionId.includes(":")) return;
    const key = rtKey(selectedAgent, sessionId);
    const existing = runtimesRef.current[key];
    if (existing && (existing.sending || existing.messages.length > 0)) return;

    getChatHistory(selectedAgent, sessionId)
      .then((history) => {
        const msgs = !history || history.length === 0 ? [] : buildChatMessages(history);
        updateRuntime(key, (r) => ({ ...r, messages: msgs }));
      })
      .catch(() => updateRuntime(key, (r) => ({ ...r, messages: [] })));
  }, [selectedAgent, sessionId, updateRuntime]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  useEffect(() => {
    const el = textareaRef.current;
    if (el) {
      el.style.height = "auto";
      el.style.height = Math.min(el.scrollHeight, 180) + "px";
    }
  }, [input]);

  // ───────────────────────── Attachment helpers ───────────────────────────
  //
  // Browsers expose three independent ways to send a file: file picker,
  // paste, and drag-and-drop. We funnel all three through addFiles() so
  // the de-dup, size guard, and preview-URL bookkeeping live in one place.

  // Upper bound mirrors the backend's maxAttachmentBytes (25 MiB). We
  // reject early in the UI so a typo on a 1 GB file doesn't waste a
  // round-trip.
  const MAX_ATTACHMENT_MB = 25;

  const addFiles = useCallback((files: File[]) => {
    if (files.length === 0) return;
    setPendingFiles((prev) => {
      const next = [...prev];
      for (const f of files) {
        if (f.size > MAX_ATTACHMENT_MB * 1024 * 1024) {
          // eslint-disable-next-line no-alert
          alert(`"${f.name}" exceeds ${MAX_ATTACHMENT_MB} MB and was skipped.`);
          continue;
        }
        // De-dup by (name, size, lastModified) — pasting the same image
        // twice or dragging the same file twice should noop.
        const dup = next.some(
          (p) => p.file.name === f.name && p.file.size === f.size && p.file.lastModified === f.lastModified
        );
        if (dup) continue;
        const isImage = f.type.startsWith("image/");
        next.push({
          id: `att-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`,
          file: f,
          previewURL: isImage ? URL.createObjectURL(f) : "",
        });
      }
      return next;
    });
  }, []);

  const removeAttachment = useCallback((id: string) => {
    setPendingFiles((prev) => {
      const target = prev.find((p) => p.id === id);
      if (target?.previewURL) URL.revokeObjectURL(target.previewURL);
      return prev.filter((p) => p.id !== id);
    });
  }, []);

  const clearAttachments = useCallback(() => {
    setPendingFiles((prev) => {
      for (const p of prev) {
        if (p.previewURL) URL.revokeObjectURL(p.previewURL);
      }
      return [];
    });
  }, []);

  // Revoke preview URLs on unmount to keep long-lived sessions tidy.
  useEffect(() => {
    return () => {
      for (const p of pendingFiles) {
        if (p.previewURL) URL.revokeObjectURL(p.previewURL);
      }
    };
    // We deliberately don't list pendingFiles in deps — the cleanup
    // above runs on unmount only. Per-item revoke happens in
    // removeAttachment / addFiles dedup flows.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handlePaste = useCallback(
    (e: React.ClipboardEvent<HTMLTextAreaElement>) => {
      const items = e.clipboardData?.items;
      if (!items) return;
      const files: File[] = [];
      for (let i = 0; i < items.length; i++) {
        const it = items[i];
        if (it.kind === "file") {
          const f = it.getAsFile();
          if (f) files.push(f);
        }
      }
      if (files.length > 0) {
        e.preventDefault();
        addFiles(files);
      }
    },
    [addFiles],
  );

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      setIsDragging(false);
      const dropped = Array.from(e.dataTransfer?.files ?? []);
      if (dropped.length > 0) addFiles(dropped);
    },
    [addFiles],
  );

  const handleDragOver = useCallback((e: React.DragEvent) => {
    if (e.dataTransfer?.types?.includes("Files")) {
      e.preventDefault();
      setIsDragging(true);
    }
  }, []);

  const handleDragLeave = useCallback(() => setIsDragging(false), []);

  const handleFileInputChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const files = Array.from(e.target.files ?? []);
      if (files.length > 0) addFiles(files);
      // Reset so selecting the same file twice in a row still triggers onChange.
      e.target.value = "";
    },
    [addFiles],
  );

  // streamTaskEvents wires a task subscription into the per-tab runtime.
  // Used by both handleSend (fresh tasks) and the resume effect (tabs the
  // user navigated away from while a task was running). Encapsulates all
  // the SSE event → runtime mutation logic so both call sites stay sane.
  const streamTaskEvents = useCallback(
    async (key: string, taskId: string, ctrl: AbortController, afterSeq?: number) => {
      const startNewGroup = () => {
        const r = runtimesRef.current[key];
        if (!r) return;
        r.curGroupId = `tg-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`;
        r.curCalls = [];
        r.curContent = "";
      };

      try {
        await subscribeTaskEvents(taskId, (evt: ChatStreamEvent) => {
          const r = runtimesRef.current[key];
          if (!r) return;
          // PR3: track the highest seq we've absorbed; written back to
          // runtime so a future reconnect can resume from this point.
          if (evt.seq != null && (!r.lastSeq || evt.seq > r.lastSeq)) {
            r.lastSeq = evt.seq;
          }
          switch (evt.type) {
            case "content": {
              const content = evt.data?.content || "";
              if (r.curCalls.length > 0) startNewGroup();
              r.curContent = content;
              updateRuntime(key, (cur) => ({
                ...cur,
                messages: [...cur.messages, { id: `a-${Date.now()}`, role: "agent", content, timestamp: Date.now() }],
              }));
              break;
            }
            case "tool_call": {
              r.curCalls.push({ id: evt.data?.id || "", name: evt.data?.name || "", arguments: evt.data?.arguments || "{}" });
              const groupId = r.curGroupId;
              const calls = [...r.curCalls];
              const content = r.curContent;
              updateRuntime(key, (cur) => {
                const last = cur.messages[cur.messages.length - 1];
                if (content && last?.role === "agent" && last.content === content) {
                  return { ...cur, messages: [...cur.messages.slice(0, -1), { id: groupId, role: "tool-group" as const, content, timestamp: Date.now(), toolCalls: calls }] };
                }
                const idx = cur.messages.findIndex((m) => m.id === groupId);
                if (idx >= 0) {
                  const u = [...cur.messages]; u[idx] = { ...u[idx], toolCalls: calls };
                  return { ...cur, messages: u };
                }
                return { ...cur, messages: [...cur.messages, { id: groupId, role: "tool-group" as const, content, timestamp: Date.now(), toolCalls: calls }] };
              });
              break;
            }
            case "tool_result": {
              const tc = r.curCalls.find((c) => c.id === (evt.data?.id || ""));
              if (tc) tc.result = evt.data?.result || "";
              const groupId = r.curGroupId;
              const calls = [...r.curCalls];
              updateRuntime(key, (cur) => {
                const idx = cur.messages.findIndex((m) => m.id === groupId);
                if (idx < 0) return cur;
                const u = [...cur.messages]; u[idx] = { ...u[idx], toolCalls: calls };
                return { ...cur, messages: u };
              });
              break;
            }
            // task_done / task_error / task_cancelled are handled in
            // the finally block below via runtime cleanup. We don't
            // need to surface them as messages.
          }
        }, ctrl.signal, afterSeq);
      } finally {
        // Cleanup runtime: clear sending flag, drop abort func, mark
        // any never-completed tool calls as cancelled so spinners stop.
        // Also clear lastSeq – the task is terminal, no resume possible.
        updateRuntime(key, (cur) => ({
          ...cur,
          sending: false,
          abort: undefined,
          taskId: undefined,
          lastSeq: undefined,
          messages: cur.messages.map((m) => {
            if (m.role !== "tool-group" || !m.toolCalls) return m;
            const hasUnfinished = m.toolCalls.some((tc) => tc.result == null);
            if (!hasUnfinished) return m;
            return {
              ...m,
              toolCalls: m.toolCalls.map((tc) =>
                tc.result == null ? { ...tc, result: "[cancelled]", error: true } : tc
              ),
            };
          }),
        }));
      }
    },
    [updateRuntime],
  );

  const handleSend = useCallback(async (text?: string) => {
    const msg = (text ?? input).trim();
    // Empty text is OK as long as at least one attachment is queued —
    // "[user uploaded a screenshot]" is a perfectly valid prompt for a
    // multimodal model.
    const hasAttachments = pendingFiles.length > 0;
    if ((!msg && !hasAttachments) || !selectedAgent) return;
    // Per-tab guard: if THIS tab is already streaming, ignore. Other tabs
    // are unaffected because each (agent, session) has its own runtime.
    const key = rtKey(selectedAgent, sessionId);
    if (runtimesRef.current[key]?.sending) return;

    setInput("");
    const agentForReq = selectedAgent;
    const sessionForReq = sessionId;
    const userMsgId = `u-${Date.now()}`;
    const initialGroupId = `tg-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`;

    // Capture attachments at submit time (state may move on while the
    // request is in flight). Snapshot the File objects so the upload
    // payload is stable even if the user clears the queue mid-flight.
    const filesToSend = pendingFiles.map((p) => p.file);

    // Build the attachment list shown on the user's bubble. Object URLs
    // are cheap to render immediately (no upload round-trip needed).
    // We hand ownership of these URLs to the message — they outlive
    // pendingFiles, so the cleanup that runs on remove/unmount must
    // not touch them. We rebuild fresh URLs here so the previews persist
    // even after the pending queue is cleared below.
    const msgAttachments: MessageAttachment[] = pendingFiles.map((p) => ({
      type: p.file.type.startsWith("image/") ? "image" : "file",
      url: URL.createObjectURL(p.file),
      name: p.file.name,
    }));

    // ctrl.abort() unsubscribes the SSE stream. The actual task cancel
    // (kill subprocess on the server) is sent via cancelTask in handleStop.
    const ctrl = new AbortController();

    updateRuntime(key, (r) => ({
      ...r,
      sending: true,
      abort: () => ctrl.abort(),
      messages: [
        ...r.messages,
        {
          id: userMsgId,
          role: "user",
          content: msg,
          timestamp: Date.now(),
          attachments: msgAttachments.length > 0 ? msgAttachments : undefined,
        },
      ],
      curGroupId: initialGroupId,
      curCalls: [],
      curContent: "",
    }));

    // Clear the pending queue right after we've snapshotted it. Doing
    // it before the await keeps the UI responsive (preview chips
    // disappear immediately, no stale state if the user starts typing
    // a new message).
    clearAttachments();

    try {
      const [deliverCh, deliverChat] = externChannel ? parseExternID(externChannel) : ["", ""];
      // Forward user message to Discord so the real session stays in sync
      if (deliverCh && deliverChat) {
        sendToChannel(agentForReq, deliverCh, deliverChat, msg).catch(() => {});
      }
      const { taskId } = await submitChat(
        agentForReq,
        sessionForReq,
        msg,
        modelOverride || undefined,
        filesToSend,
        deliverCh || undefined,
        deliverChat || undefined,
      );
      // Persist taskId BEFORE subscribing so handleStop knows what to
      // cancel and the resume effect can find this in-flight task.
      updateRuntime(key, (cur) => ({ ...cur, taskId }));
      await streamTaskEvents(key, taskId, ctrl);
      loadSessions(agentForReq);
    } catch (err) {
      const aborted = (err as Error)?.name === "AbortError" || ctrl.signal.aborted;
      if (!aborted) {
        const errMsg = (err as Error)?.message || "Failed to get a response. Is the gateway running?";
        updateRuntime(key, (cur) => ({
          ...cur,
          sending: false,
          abort: undefined,
          taskId: undefined,
          messages: [
            ...cur.messages,
            { id: `e-${Date.now()}`, role: "agent", content: `⚠️ ${errMsg}`, timestamp: Date.now() },
          ],
        }));
      }
    } finally {
      if (rtKey(selectedAgent, sessionId) === key) {
        textareaRef.current?.focus();
      }
    }
  }, [input, selectedAgent, sessionId, modelOverride, pendingFiles, clearAttachments, loadSessions, updateRuntime, streamTaskEvents]);

  const handleStop = useCallback(() => {
    const r = runtimesRef.current[currentKey];
    // Server-side cancel kills the subprocess + stops the LLM call.
    // Fire-and-forget; the backend marks the task cancelled regardless
    // of whether the response reaches us.
    if (r?.taskId) {
      cancelTask(r.taskId).catch(() => { /* idempotent on backend */ });
    }
    // Local abort unsubscribes the SSE stream → triggers streamTaskEvents'
    // finally block → cleans up the runtime.
    r?.abort?.();
  }, [currentKey]);

  // Resume effect: when the user navigates back to a tab whose task is
  // still running (or finished while away), re-subscribe so the UI keeps
  // updating. Without this, switching away mid-stream would orphan the
  // task and the user would never see the result.
  useEffect(() => {
    if (!selectedAgent || !sessionId) return;
    const key = rtKey(selectedAgent, sessionId);
    const r = runtimesRef.current[key];
    if (!r?.taskId || r.sending) return;
    // We have a known taskId for this tab but no active subscription.
    // Either it finished while we were away, or the previous SSE stream
    // dropped. Either way, getTask + (maybe) re-subscribe will catch us up.
    let cancelled = false;
    // Resume from the highest seq we already absorbed so we don't replay
    // events the UI has already rendered.
    const resumeFromSeq = r.lastSeq;
    (async () => {
      try {
        const t = await getTask(r.taskId!);
        if (cancelled) return;
        if (t.status === "running" || t.status === "pending") {
          const ctrl = new AbortController();
          updateRuntime(key, (cur) => ({ ...cur, sending: true, abort: () => ctrl.abort() }));
          await streamTaskEvents(key, t.id, ctrl, resumeFromSeq);
        } else {
          // Already terminal – just clear residual state.
          updateRuntime(key, (cur) => ({ ...cur, taskId: undefined, sending: false, lastSeq: undefined }));
        }
      } catch {
        // Task disappeared (e.g. server restart). Drop the dangling ref.
        updateRuntime(key, (cur) => ({ ...cur, taskId: undefined, sending: false, lastSeq: undefined }));
      }
    })();
    return () => { cancelled = true; };
  }, [selectedAgent, sessionId, updateRuntime, streamTaskEvents]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) { e.preventDefault(); handleSend(); }
  };

  const handleCopy = (msg: ChatMessage) => {
    navigator.clipboard.writeText(msg.content);
    setCopiedId(msg.id);
    setTimeout(() => setCopiedId(null), 1500);
  };

  const handleNewChat = useCallback(() => {
    // Switching to a fresh session – the new (agent, session) tuple gets
    // its own runtime entry on first access, so we don't need to clear
    // anything here. The OLD tab's runtime (and any in-flight stream) is
    // preserved so background tasks keep running.
    setSessionId(generateSessionId());
    setExternChannel(null);
  }, []);

  const handleDeleteSession = async (sid: string, e: React.MouseEvent) => {
    e.stopPropagation();
    if (!selectedAgent) return;
    try {
      await deleteChatSession(selectedAgent, sid);
      loadSessions(selectedAgent);
      // Drop the runtime for the deleted session, aborting any in-flight stream.
      const dropKey = rtKey(selectedAgent, sid);
      const dropped = runtimesRef.current[dropKey];
      dropped?.abort?.();
      setRuntimes((prev) => {
        const next = { ...prev };
        delete next[dropKey];
        return next;
      });
      if (sessionId === sid) handleNewChat();
    } catch { /* ignore */ }
  };

  const currentAgent = agents.find((a) => a.id === selectedAgent);

  // Models known to this gateway: union of all agents' configured models.
  // Cheap to compute, gives the user a quick switcher between models they
  // have already configured for some agent. Free-form input could come
  // later; for now this covers the "I have flash AND pro configured, let
  // me pick per message" case.
  const availableModels = useMemo(() => {
    const set = new Set<string>();
    for (const a of agents) {
      if (a.model) set.add(a.model);
    }
    return Array.from(set).sort();
  }, [agents]);

  // What's actually going to be used on the next send. Exposed in the UI
  // as a label so a user who picked a non-default model in a previous turn
  // doesn't accidentally send with the wrong model after navigating away.
  const effectiveModel = modelOverride || currentAgent?.model || "";

  return (
    <div className="flex h-[calc(100vh-3rem)] md:h-screen bg-background">
      {/* ── Left panel: Agents + Sessions ── */}
      <aside className="hidden w-60 flex-col border-r border-border bg-card/20 lg:flex shrink-0">
        {/* Agent list */}
        <div className="border-b border-border">
          <p className="px-4 pt-4 pb-2 text-[10px] font-semibold uppercase tracking-widest text-muted-foreground/60">
            Agents
          </p>
          <div className="px-2 pb-2 space-y-0.5">
            {agents.map((agent) => (
              <button
                key={agent.id}
                onClick={() => { setSelectedAgent(agent.id); handleNewChat(); }}
                className={`flex w-full items-center gap-2.5 rounded-lg px-3 py-2 text-sm transition-colors ${
                  selectedAgent === agent.id
                    ? "bg-primary/10 text-primary font-medium"
                    : "text-muted-foreground hover:bg-muted/50 hover:text-foreground"
                }`}
              >
                <div className={`flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-[10px] font-bold ${
                  selectedAgent === agent.id ? "bg-primary text-primary-foreground" : "bg-muted text-muted-foreground"
                }`}>
                  {agent.id[0]?.toUpperCase()}
                </div>
                <div className="flex-1 min-w-0 text-left">
                  <p className="truncate text-xs font-medium leading-tight">{agent.id}</p>
                  <p className="truncate text-[10px] opacity-60 leading-tight font-mono">{agent.model.split("/").pop()}</p>
                </div>
              </button>
            ))}
          </div>
        </div>

        {/* Session history */}
        <div className="flex flex-1 flex-col min-h-0">
          <div className="flex items-center justify-between px-4 pt-4 pb-2">
            <p className="text-[10px] font-semibold uppercase tracking-widest text-muted-foreground/60">
              History
            </p>
            <button
              onClick={handleNewChat}
              className="rounded-md p-1 text-muted-foreground hover:text-foreground hover:bg-muted/50 transition-colors"
              title="New chat"
            >
              <SquarePen className="h-3.5 w-3.5" />
            </button>
          </div>

          <div className="flex-1 overflow-y-auto px-2 pb-2 space-y-0.5">
            {sessions.length === 0 && (
              <p className="px-3 py-2 text-xs text-muted-foreground/50 italic">No conversations yet</p>
            )}
            {sessions.map((s) => (
              <div
                key={s.id}
                className={`group flex w-full items-center gap-1.5 rounded-lg px-3 py-2 text-sm transition-colors cursor-pointer ${
                  sessionId === s.id
                    ? "bg-primary/10 text-primary"
                    : "text-muted-foreground hover:bg-muted/50 hover:text-foreground"
                }`}
                onClick={() => { setSessionId(s.id); setExternChannel(null); }}
              >
                <MessageSquare className="h-3.5 w-3.5 shrink-0 opacity-60" />
                <span className="flex-1 truncate text-xs">{s.preview}</span>
                <button
                  onClick={(e) => handleDeleteSession(s.id, e)}
                  className="shrink-0 rounded p-0.5 opacity-0 group-hover:opacity-100 hover:bg-destructive/10 hover:text-destructive transition-all"
                  title="Delete"
                >
                  <Trash2 className="h-3 w-3" />
                </button>
              </div>
            ))}
          </div>

          {/* External sessions (Discord, Telegram) */}
          {externSessions.length > 0 && (
            <div className="flex flex-col min-h-0 border-t border-border pt-1">
              <div className="px-4 pt-3 pb-2">
                <p className="text-[10px] font-semibold uppercase tracking-widest text-muted-foreground/60">
                  Channels
                </p>
              </div>
              <div className="overflow-y-auto px-2 pb-2 space-y-0.5 max-h-48">
                {externSessions.map((es) => (
                  <div
                    key={es.id}
                    className={`group flex w-full items-center gap-1.5 rounded-lg px-3 py-2 text-sm transition-colors cursor-pointer ${
                      externChannel === es.id
                        ? "bg-primary/10 text-primary"
                        : "text-muted-foreground hover:bg-muted/50 hover:text-foreground"
                    }`}
                    onClick={() => { setExternChannel(es.id); setSessionId(""); }}
                  >
                    <Radio className="h-3.5 w-3.5 shrink-0 opacity-60" />
                    <div className="flex-1 min-w-0">
                      <p className="truncate text-xs font-medium">{es.channel}</p>
                      <p className="truncate text-[10px] text-muted-foreground/60">{es.preview}</p>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </aside>

      {/* ── Main chat area ── */}
      <div className="flex flex-1 flex-col min-w-0">
        {/* Chat header */}
        <header className="flex h-12 items-center justify-between border-b border-border px-4 shrink-0 bg-card/30 backdrop-blur-sm">
          <div className="flex items-center gap-2.5 min-w-0">
            <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary/10">
              <Bot className="h-4 w-4 text-primary" />
            </div>
            <div className="min-w-0">
              <p className="text-sm font-semibold leading-tight truncate">
                {selectedAgent || "No agent selected"}
              </p>
              {currentAgent && (
                <p className="text-[10px] text-muted-foreground font-mono leading-tight truncate">
                  {currentAgent.model}
                </p>
              )}
            </div>
          </div>

          <div className="flex items-center gap-2">
            {/* Mobile agent selector */}
            {agents.length > 1 && (
              <select
                value={selectedAgent}
                onChange={(e) => { setSelectedAgent(e.target.value); handleNewChat(); }}
                className="rounded-lg border border-border bg-card px-2 py-1 text-xs lg:hidden"
              >
                {agents.map((a) => <option key={a.id} value={a.id}>{a.id}</option>)}
              </select>
            )}
            {/* Context window usage ring */}
            {contextInfo && contextInfo.contextWindow > 0 && (
              <ContextRing info={contextInfo} />
            )}

            {/* View System Prompt — opens a modal with the full, labelled
                breakdown of what the agent prepends to every LLM call.
                Useful for understanding the baseline token cost shown by
                the ContextRing on a fresh session. */}
            <button
              onClick={openSystemPromptModal}
              disabled={!selectedAgent}
              className="flex h-8 w-8 items-center justify-center rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted/50 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
              title="View system prompt"
            >
              <Eye className="h-4 w-4" />
            </button>

            {/* Open the agent's workspace directory in the host file
                explorer. Shows briefly to confirm the request was
                received — actual GUI launch happens server-side. */}
            <button
              onClick={async () => {
                if (!selectedAgent) return;
                try {
                  const r = await openWorkspace(selectedAgent);
                  if (!r.ok) {
                    // eslint-disable-next-line no-alert
                    alert(`Failed to open: ${r.error || "unknown error"}`);
                  }
                } catch {
                  // eslint-disable-next-line no-alert
                  alert("Failed to open workspace");
                }
              }}
              disabled={!selectedAgent}
              className="flex h-8 w-8 items-center justify-center rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted/50 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
              title={
                currentAgent?.workspace
                  ? `Open workspace folder\n${currentAgent.workspace}`
                  : "Open workspace folder"
              }
            >
              <FolderOpen className="h-4 w-4" />
            </button>

            <button
              onClick={handleNewChat}
              className="flex h-8 w-8 items-center justify-center rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted/50 transition-colors"
              title="New Chat"
            >
              <SquarePen className="h-4 w-4" />
            </button>
          </div>
        </header>

        {/* Messages */}
        <div className="flex-1 overflow-y-auto min-h-0">
          <div className="mx-auto max-w-3xl px-4 py-6 space-y-1">
            {messages.length === 0 && !sending && (
              <EmptyState agentName={selectedAgent} onPrompt={(p) => handleSend(p)} />
            )}

            {messages.map((msg) =>
              msg.role === "tool-group" ? (
                <ToolCallGroup key={msg.id} msg={msg} />
              ) : (
                <MessageBubble
                  key={msg.id}
                  msg={msg}
                  copiedId={copiedId}
                  onCopy={handleCopy}
                  agentId={selectedAgent}
                />
              )
            )}

            {sending && (
              <div className="flex justify-start pt-1">
                <div className="flex items-center gap-2 rounded-2xl rounded-bl-md bg-muted px-4 py-3">
                  <div className="flex gap-1">
                    {[0, 150, 300].map((delay) => (
                      <span
                        key={delay}
                        className="inline-block h-1.5 w-1.5 rounded-full bg-muted-foreground/50 animate-bounce"
                        style={{ animationDelay: `${delay}ms` }}
                      />
                    ))}
                  </div>
                  <span className="text-xs text-muted-foreground/60">Thinking…</span>
                </div>
              </div>
            )}

            <div ref={messagesEndRef} />
          </div>
        </div>

        {/* Input */}
        <div
          className="shrink-0 px-4 pb-5 pt-3 border-t border-border bg-card/20"
          onDragOver={handleDragOver}
          onDragLeave={handleDragLeave}
          onDrop={handleDrop}
        >
          <div className="mx-auto max-w-3xl">
            <div
              className={`flex flex-col rounded-xl border bg-card shadow-sm focus-within:border-primary/40 focus-within:shadow-md transition-all ${
                isDragging ? "border-primary border-dashed bg-primary/5" : "border-border"
              }`}
            >
              {/* Hidden file input — triggered by the paperclip button.
                  Accept any file but the agent loop currently only inlines
                  images; non-images get a textual breadcrumb so the model
                  knows they were sent. */}
              <input
                ref={fileInputRef}
                type="file"
                multiple
                accept="image/*"
                onChange={handleFileInputChange}
                className="hidden"
              />

              {/* Attachment preview strip — only rendered when files are
                  queued. Sits above the textarea so it doesn't fight for
                  vertical space when empty. */}
              {pendingFiles.length > 0 && (
                <div className="flex flex-wrap gap-2 px-3 pt-3">
                  {pendingFiles.map((p) => {
                    const isImage = p.file.type.startsWith("image/");
                    return (
                      <div
                        key={p.id}
                        className="group relative flex items-center gap-2 rounded-lg border border-border bg-muted/40 pl-2 pr-7 py-1.5"
                      >
                        {isImage && p.previewURL ? (
                          // eslint-disable-next-line @next/next/no-img-element
                          <img
                            src={p.previewURL}
                            alt={p.file.name}
                            className="h-8 w-8 rounded object-cover"
                          />
                        ) : (
                          <div className="flex h-8 w-8 items-center justify-center rounded bg-background/50">
                            <FileText className="h-4 w-4 text-muted-foreground" />
                          </div>
                        )}
                        <div className="min-w-0 max-w-[140px]">
                          <p className="truncate text-[11px] font-medium leading-tight">
                            {p.file.name}
                          </p>
                          <p className="text-[10px] text-muted-foreground/60 leading-tight">
                            {(p.file.size / 1024).toFixed(0)} KB
                          </p>
                        </div>
                        <button
                          type="button"
                          onClick={() => removeAttachment(p.id)}
                          className="absolute right-1 top-1 rounded p-0.5 text-muted-foreground/60 hover:bg-destructive/10 hover:text-destructive transition-colors"
                          title="Remove"
                        >
                          <X className="h-3 w-3" />
                        </button>
                      </div>
                    );
                  })}
                </div>
              )}

              <textarea
                ref={textareaRef}
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={handleKeyDown}
                onPaste={handlePaste}
                placeholder={selectedAgent ? `Message ${selectedAgent}…` : "Select an agent first"}
                disabled={!selectedAgent}
                rows={1}
                className="w-full resize-none bg-transparent px-4 pt-3 pb-2 text-[14px] leading-relaxed placeholder:text-muted-foreground/40 outline-none disabled:opacity-40"
                style={{ maxHeight: 180, minHeight: 42 }}
              />
              <div className="flex items-center justify-between px-3 pb-2.5 gap-2">
                {/*
                  Per-message model picker. Disabled while sending so the
                  user can't change models mid-stream (the in-flight task
                  is already bound to whatever was picked at submit time).
                  Defaults to the agent's configured model and resets on
                  agent switch (see useEffect on selectedAgent).
                */}
                <div className="flex items-center gap-2 min-w-0">
                  {availableModels.length > 0 && (
                    <select
                      value={effectiveModel}
                      onChange={(e) => {
                        const next = e.target.value;
                        setModelOverride(next === currentAgent?.model ? "" : next);
                      }}
                      disabled={sending || !selectedAgent}
                      title={
                        modelOverride
                          ? `Override active for this message (default: ${currentAgent?.model})`
                          : "Pick a model for this message"
                      }
                      className={`max-w-[180px] truncate rounded-md border bg-card px-1.5 py-0.5 text-[11px] font-mono outline-none transition-colors ${
                        modelOverride
                          ? "border-primary/40 text-primary"
                          : "border-border text-muted-foreground/70 hover:text-foreground"
                      } disabled:opacity-40`}
                    >
                      {availableModels.map((m) => (
                        <option key={m} value={m}>{m}</option>
                      ))}
                    </select>
                  )}
                  <p className="text-[11px] text-muted-foreground/40 select-none truncate">
                    {sending ? "Responding…" : "↵ Send  ·  ⇧↵ New line"}
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  {/* Skill launcher — picks any installed skill and
                      prepends a "Use the X skill: " hint to the input.
                      Especially useful for orchestrator (suite/protocol)
                      skills where the agent benefits from being told
                      explicitly which entry point to follow. */}
                  <DropdownMenu>
                    <DropdownMenuTrigger
                      disabled={sending || !selectedAgent || skills.length === 0}
                      title="Launch with a skill"
                      className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground/70 hover:text-primary hover:bg-primary/10 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
                    >
                      <Sparkles className="h-3.5 w-3.5" />
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" className="w-72 max-h-[60vh] overflow-y-auto">
                      {skillGroups.suites.length > 0 && (
                        <>
                          <div className="px-2 py-1.5 text-[10px] uppercase tracking-wider text-muted-foreground/70 font-semibold select-none">
                            编排器 (Orchestrators)
                          </div>
                          {skillGroups.suites.map((s) => (
                            <DropdownMenuItem
                              key={s.name}
                              onClick={() => pickSkill(s)}
                              className="flex flex-col items-start gap-0.5 py-2"
                            >
                              <span className="text-sm font-mono">{s.name}</span>
                              {s.description && (
                                <span className="text-[10px] text-muted-foreground line-clamp-2">
                                  {s.description}
                                </span>
                              )}
                            </DropdownMenuItem>
                          ))}
                          {(skillGroups.protocols.length > 0 || skillGroups.atoms.length > 0) && <DropdownMenuSeparator />}
                        </>
                      )}
                      {skillGroups.protocols.length > 0 && (
                        <>
                          <div className="px-2 py-1.5 text-[10px] uppercase tracking-wider text-muted-foreground/70 font-semibold select-none">
                            协议 (Protocols)
                          </div>
                          {skillGroups.protocols.map((s) => (
                            <DropdownMenuItem
                              key={s.name}
                              onClick={() => pickSkill(s)}
                              className="flex flex-col items-start gap-0.5 py-2"
                            >
                              <span className="text-sm font-mono">{s.name}</span>
                              {s.description && (
                                <span className="text-[10px] text-muted-foreground line-clamp-2">
                                  {s.description}
                                </span>
                              )}
                            </DropdownMenuItem>
                          ))}
                          {skillGroups.atoms.length > 0 && <DropdownMenuSeparator />}
                        </>
                      )}
                      {skillGroups.atoms.length > 0 && (
                        <>
                          <div className="px-2 py-1.5 text-[10px] uppercase tracking-wider text-muted-foreground/70 font-semibold select-none">
                            原子 Skill ({skillGroups.atoms.length})
                          </div>
                          {skillGroups.atoms.map((s) => (
                            <DropdownMenuItem
                              key={s.name}
                              onClick={() => pickSkill(s)}
                              className="flex flex-col items-start gap-0.5 py-1.5"
                            >
                              <span className="text-xs font-mono">{s.name}</span>
                              {s.description && (
                                <span className="text-[10px] text-muted-foreground line-clamp-1">
                                  {s.description}
                                </span>
                              )}
                            </DropdownMenuItem>
                          ))}
                        </>
                      )}
                    </DropdownMenuContent>
                  </DropdownMenu>

                  {/* Paperclip — opens the file picker. Hidden during
                      sending to avoid the user mutating the queue while
                      a request is mid-flight (the in-flight task already
                      owns its own file snapshot, so changes wouldn't
                      affect the current turn anyway). */}
                  <button
                    type="button"
                    onClick={() => fileInputRef.current?.click()}
                    disabled={sending || !selectedAgent}
                    title="Attach images (also: paste / drag-and-drop)"
                    className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground/70 hover:text-foreground hover:bg-muted/60 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
                  >
                    <Paperclip className="h-3.5 w-3.5" />
                  </button>

                  {sending ? (
                    <Button
                      onClick={handleStop}
                      size="sm"
                      variant="outline"
                      className="h-7 gap-1.5 text-xs text-destructive border-destructive/40 hover:bg-destructive/10 hover:border-destructive"
                    >
                      <Square className="h-3 w-3 fill-current" />
                      Stop
                    </Button>
                  ) : (
                    <Button
                      onClick={() => handleSend()}
                      // Allow sending with attachments only (no text) —
                      // a screenshot with no caption is a real use case.
                      disabled={(!input.trim() && pendingFiles.length === 0) || !selectedAgent}
                      size="sm"
                      className="h-7 gap-1.5 text-xs"
                    >
                      <Send className="h-3 w-3" />
                      Send
                    </Button>
                  )}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* System Prompt Preview Modal — full-screen overlay with the labelled,
          token-counted breakdown of what gets prepended to every LLM call.
          Sections are ordered by token weight so the dominant contributor
          (usually AGENTS.md or skills) is immediately visible. */}
      {showSystemPromptModal && (
        <SystemPromptModal
          info={systemPromptInfo}
          loading={systemPromptLoading}
          agentId={selectedAgent}
          onClose={() => setShowSystemPromptModal(false)}
          onRefresh={openSystemPromptModal}
        />
      )}
    </div>
  );
}

// ── Context Ring ─────────────────────────────────────────────────────────────
// A small SVG donut chart that shows how much of the model's context window
// is in use. Color-coded: green (<70%), yellow (70-92%), red (>92%).
// Hover shows a tooltip with detailed stats.

function ContextRing({ info }: { info: SessionContextInfo }) {
  const [showTip, setShowTip] = useState(false);

  const ratio = info.contextWindow > 0 ? info.currentTokens / info.contextWindow : 0;
  const pct = Math.min(ratio * 100, 100);

  // Color thresholds match soft/hard compaction boundaries
  const softRatio = info.contextWindow > 0 ? info.softThreshold / info.contextWindow : 0.7;
  const hardRatio = info.contextWindow > 0 ? info.hardThreshold / info.contextWindow : 0.92;

  const color =
    ratio >= hardRatio ? "#ef4444" :   // red — above hard threshold
    ratio >= softRatio ? "#f59e0b" :   // amber — above soft threshold
    "#22c55e";                          // green — healthy

  // SVG donut: r=8, circumference = 2π×8 ≈ 50.27
  const R = 8;
  const C = 2 * Math.PI * R;
  const dash = (pct / 100) * C;

  const fmtK = (n: number) => n >= 1000 ? `${(n / 1000).toFixed(1)}k` : String(n);

  return (
    <div className="relative flex items-center" onMouseEnter={() => setShowTip(true)} onMouseLeave={() => setShowTip(false)}>
      {/* Ring icon */}
      <div className="flex h-8 w-8 items-center justify-center rounded-lg hover:bg-muted/50 transition-colors cursor-default">
        <svg width="20" height="20" viewBox="0 0 20 20">
          {/* Track */}
          <circle
            cx="10" cy="10" r={R}
            fill="none"
            stroke="currentColor"
            strokeWidth="2.5"
            className="text-muted-foreground/20"
          />
          {/* Progress arc */}
          <circle
            cx="10" cy="10" r={R}
            fill="none"
            stroke={color}
            strokeWidth="2.5"
            strokeLinecap="round"
            strokeDasharray={`${dash} ${C}`}
            strokeDashoffset={C / 4} // start at top (12 o'clock)
            style={{ transition: "stroke-dasharray 0.4s ease" }}
          />
        </svg>
      </div>

      {/* Tooltip */}
      {showTip && (
        <div className="absolute right-0 top-full mt-2 z-50 w-64 rounded-xl border border-border bg-popover shadow-lg p-3 text-xs space-y-2">
          {/* Header */}
          <div className="flex items-center justify-between">
            <span className="font-semibold text-foreground">Context Usage</span>
            <span className="font-mono font-bold" style={{ color }}>{pct.toFixed(1)}%</span>
          </div>

          {/* Progress bar */}
          <div className="h-1.5 rounded-full bg-muted overflow-hidden">
            <div
              className="h-full rounded-full transition-all"
              style={{ width: `${pct}%`, backgroundColor: color }}
            />
          </div>

          {/* Stats grid */}
          <div className="grid grid-cols-2 gap-x-3 gap-y-1 text-muted-foreground">
            <span>Used</span>
            <span className="font-mono text-foreground text-right">{fmtK(info.currentTokens)} tk</span>
            <span>Context window</span>
            <span className="font-mono text-foreground text-right">{fmtK(info.contextWindow)} tk</span>
            <span>Soft threshold</span>
            <span className="font-mono text-right">{fmtK(info.softThreshold)} tk</span>
            <span>Hard threshold</span>
            <span className="font-mono text-right">{fmtK(info.hardThreshold)} tk</span>
            <span>Messages</span>
            <span className="font-mono text-foreground text-right">{info.messageCount}</span>
            <span>Compactions</span>
            <span className="font-mono text-foreground text-right">{info.compactionCount}</span>
          </div>

          {/* Model */}
          <div className="border-t border-border pt-2 font-mono text-[10px] text-muted-foreground/60 truncate">
            {info.modelId}
          </div>
        </div>
      )}
    </div>
  );
}

// ── System Prompt Modal ──────────────────────────────────────────────────────
// Full-screen overlay that surfaces the labelled, token-counted breakdown
// of the agent's system prompt — exactly what gets sent to the LLM as the
// `system` message every turn. Sections are sorted by token weight so the
// dominant contributor is at the top, with a per-section copy button and
// expand/collapse for the rendered content.

function SystemPromptModal({
  info,
  loading,
  agentId,
  onClose,
  onRefresh,
}: {
  info: SystemPromptInfo | null;
  loading: boolean;
  agentId: string;
  onClose: () => void;
  onRefresh: () => void;
}) {
  // Per-section expand state. Default: only the dominant section is expanded
  // so the modal opens to a digestible overview rather than a wall of text.
  const [expanded, setExpanded] = useState<Record<string, boolean>>({});
  const [copiedKey, setCopiedKey] = useState<string | null>(null);

  // Esc to dismiss — standard modal affordance.
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  // Sort sections by token count desc so the dominant contributor is on top.
  // We keep the original order as a tiebreaker for sections of equal size.
  const sortedSections = useMemo(() => {
    if (!info) return [];
    return info.sections
      .map((s, idx) => ({ ...s, _origIdx: idx }))
      .sort((a, b) => b.tokens - a.tokens || a._origIdx - b._origIdx);
  }, [info]);

  // Auto-expand the heaviest section the first time we get data for an agent.
  useEffect(() => {
    if (!info || sortedSections.length === 0) return;
    setExpanded((prev) => {
      // Only seed if nothing is currently expanded; otherwise respect user choice.
      const anyOpen = Object.values(prev).some(Boolean);
      if (anyOpen) return prev;
      const top = sortedSections[0];
      return { [top.name]: true };
    });
  }, [info, sortedSections]);

  const fmtK = (n: number) => (n >= 1000 ? `${(n / 1000).toFixed(1)}k` : String(n));

  const copySection = async (key: string, text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedKey(key);
      setTimeout(() => setCopiedKey(null), 1500);
    } catch {
      /* clipboard API blocked — non-fatal */
    }
  };

  const copyAll = async () => {
    if (!info) return;
    const joined = info.sections.map((s) => s.content).join("\n\n---\n\n");
    await copySection("__all__", joined);
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4"
      onClick={onClose}
    >
      <div
        className="relative flex h-[85vh] w-full max-w-4xl flex-col rounded-2xl border border-border bg-card shadow-2xl overflow-hidden"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between border-b border-border px-5 py-3 shrink-0">
          <div className="flex items-center gap-2.5 min-w-0">
            <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-primary/10">
              <FileCode className="h-4 w-4 text-primary" />
            </div>
            <div className="min-w-0">
              <p className="text-sm font-semibold leading-tight">System Prompt</p>
              <p className="text-[11px] text-muted-foreground font-mono leading-tight truncate">
                {agentId}
                {info?.modelId ? `  ·  ${info.modelId}` : ""}
              </p>
            </div>
          </div>

          <div className="flex items-center gap-1.5">
            {info && (
              <span className="rounded-md bg-muted px-2 py-1 text-[11px] font-mono text-muted-foreground">
                {fmtK(info.totalTokens)} tk total
              </span>
            )}
            <button
              onClick={copyAll}
              disabled={!info}
              className="flex h-8 items-center gap-1.5 rounded-lg px-2.5 text-xs text-muted-foreground hover:text-foreground hover:bg-muted/50 transition-colors disabled:opacity-40"
              title="Copy entire system prompt"
            >
              {copiedKey === "__all__" ? (
                <Check className="h-3.5 w-3.5 text-emerald-500" />
              ) : (
                <Copy className="h-3.5 w-3.5" />
              )}
              <span>Copy all</span>
            </button>
            <button
              onClick={onRefresh}
              disabled={loading}
              className="flex h-8 items-center gap-1.5 rounded-lg px-2.5 text-xs text-muted-foreground hover:text-foreground hover:bg-muted/50 transition-colors disabled:opacity-40"
              title="Refresh — re-reads workspace files"
            >
              Refresh
            </button>
            <button
              onClick={onClose}
              className="flex h-8 w-8 items-center justify-center rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted/50 transition-colors"
              title="Close (Esc)"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
        </div>

        {/* Body */}
        <div className="flex-1 overflow-y-auto">
          {loading && !info ? (
            <div className="flex h-full items-center justify-center">
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <div className="h-4 w-4 rounded-full border-2 border-primary border-t-transparent animate-spin" />
                Loading system prompt…
              </div>
            </div>
          ) : !info ? (
            <div className="flex h-full items-center justify-center">
              <p className="text-sm text-muted-foreground">No system prompt available.</p>
            </div>
          ) : (
            <div className="space-y-2 p-4">
              {sortedSections.map((s) => {
                const pct = info.totalTokens > 0 ? (s.tokens / info.totalTokens) * 100 : 0;
                const isOpen = !!expanded[s.name];
                return (
                  <div
                    key={s.name}
                    className="rounded-xl border border-border/60 bg-background/50 overflow-hidden"
                  >
                    {/* Section header — click to expand */}
                    <button
                      onClick={() => setExpanded((p) => ({ ...p, [s.name]: !p[s.name] }))}
                      className="flex w-full items-center gap-3 px-4 py-2.5 hover:bg-muted/30 transition-colors text-left"
                    >
                      {isOpen ? (
                        <ChevronDown className="h-4 w-4 text-muted-foreground/60 shrink-0" />
                      ) : (
                        <ChevronRight className="h-4 w-4 text-muted-foreground/60 shrink-0" />
                      )}
                      <span className="text-sm font-semibold text-foreground shrink-0">
                        {s.name}
                      </span>
                      <div className="flex-1 min-w-0 flex items-center gap-2">
                        {/* Inline mini progress bar showing share of total */}
                        <div className="flex-1 h-1.5 rounded-full bg-muted overflow-hidden max-w-[200px]">
                          <div
                            className="h-full rounded-full bg-primary/60 transition-all"
                            style={{ width: `${pct}%` }}
                          />
                        </div>
                        <span className="text-[11px] text-muted-foreground font-mono shrink-0">
                          {pct.toFixed(1)}%
                        </span>
                      </div>
                      <span className="text-xs font-mono text-muted-foreground shrink-0 tabular-nums">
                        {fmtK(s.tokens)} tk
                      </span>
                    </button>

                    {/* Section body */}
                    {isOpen && (
                      <div className="border-t border-border/60 bg-muted/10">
                        <div className="flex items-center justify-end px-3 py-1.5 border-b border-border/40">
                          <button
                            onClick={() => copySection(s.name, s.content)}
                            className="flex items-center gap-1.5 rounded-md px-2 py-0.5 text-[11px] text-muted-foreground hover:text-foreground hover:bg-muted/60 transition-colors"
                            title="Copy section"
                          >
                            {copiedKey === s.name ? (
                              <Check className="h-3 w-3 text-emerald-500" />
                            ) : (
                              <Copy className="h-3 w-3" />
                            )}
                            <span>{copiedKey === s.name ? "Copied" : "Copy"}</span>
                          </button>
                        </div>
                        <pre className="text-[12px] font-mono px-4 py-3 overflow-x-auto whitespace-pre-wrap break-words leading-relaxed text-foreground/90 max-h-[40vh] overflow-y-auto">
                          {s.content || <span className="text-muted-foreground italic">(empty)</span>}
                        </pre>
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          )}
        </div>

        {/* Footer hint */}
        <div className="border-t border-border px-5 py-2 text-[11px] text-muted-foreground/70 shrink-0">
          This is exactly what gets sent to the LLM as the <code className="font-mono">system</code> message every turn.
          Edit AGENTS.md / SOUL.md / skills in the workspace and click Refresh to re-render.
        </div>
      </div>
    </div>
  );
}

// ── Empty state ──────────────────────────────────────────────────────────────

function EmptyState({ agentName, onPrompt }: { agentName: string; onPrompt: (p: string) => void }) {
  return (
    <div className="flex flex-col items-center justify-center py-20 text-center select-none">
      <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-primary/10 mb-5 shadow-inner">
        <Zap className="h-8 w-8 text-primary" />
      </div>
      <h2 className="text-xl font-semibold mb-1">
        {agentName ? `Chat with ${agentName}` : "Start a conversation"}
      </h2>
      <p className="text-sm text-muted-foreground mb-8 max-w-xs">
        Ask anything, run tools, or browse your workspace.
      </p>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 w-full max-w-md">
        {STARTER_PROMPTS.map((p) => (
          <button
            key={p}
            onClick={() => onPrompt(p)}
            className="flex items-center gap-2 rounded-xl border border-border bg-card/50 px-4 py-3 text-sm text-left text-muted-foreground hover:bg-muted/50 hover:text-foreground hover:border-border/80 transition-colors group"
          >
            <MessageSquare className="h-4 w-4 shrink-0 opacity-50 group-hover:opacity-80 transition-opacity" />
            <span>{p}</span>
          </button>
        ))}
      </div>
    </div>
  );
}

// ── Message bubble ───────────────────────────────────────────────────────────

// rewriteMarkdownImageSrc resolves relative image paths in agent
// markdown output to a /api/files URL so the workspace-resident image
// the agent just generated actually loads in the browser. Absolute
// http(s):// and data: URLs are left untouched.
//
// We accept any relative form ("foo.png", "./foo.png", "subdir/foo.png")
// — they're all rooted at the agent workspace anyway, and the backend
// runs filepath.Clean before serving, so trailing dots can't escape.
function rewriteMarkdownImageSrc(src: string | undefined, agentId: string): string {
  if (!src) return "";
  if (/^(https?:|data:|blob:|\/api\/)/i.test(src)) return src;
  // Strip any leading "./" — looks ugly in the URL otherwise.
  const cleaned = src.replace(/^\.\//, "");
  return fileURL({ kind: "workspace", path: cleaned, agentId });
}

function MessageBubble({
  msg,
  copiedId,
  onCopy,
  agentId,
}: {
  msg: ChatMessage;
  copiedId: string | null;
  onCopy: (m: ChatMessage) => void;
  agentId: string;
}) {
  const isUser = msg.role === "user";
  const attachments = msg.attachments ?? [];
  const hasAttachments = attachments.length > 0;
  const [lightboxURL, setLightboxURL] = useState<string | null>(null);

  return (
    <div className={`flex gap-2 py-0.5 ${isUser ? "justify-end" : "justify-start"}`}>
      {!isUser && (
        <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary/10 mt-0.5">
          <Bot className="h-3.5 w-3.5 text-primary" />
        </div>
      )}

      <div className={`group max-w-[78%] ${isUser ? "order-first" : ""}`}>
        {/* Attachments strip — sits above the bubble for user messages
            so the layout reads as "what I sent: image + text". When
            there's no text we still render the strip alone. */}
        {hasAttachments && (
          <div className={`mb-1.5 flex flex-wrap gap-1.5 ${isUser ? "justify-end" : "justify-start"}`}>
            {attachments.map((att, idx) => (
              <AttachmentTile
                key={idx}
                att={att}
                onPreview={(url) => setLightboxURL(url)}
              />
            ))}
          </div>
        )}

        {/* Bubble itself only renders when there's text content; an
            image-only message looks cleaner without an empty box. */}
        {msg.content && (
          <div
            className={`rounded-2xl px-4 py-2.5 ${
              isUser
                ? "bg-primary text-primary-foreground rounded-br-sm"
                : "bg-muted rounded-bl-sm"
            }`}
          >
            {/*
              User bubbles always need `prose-invert` (light text on the dark
              primary background). The agent bubble uses `bg-muted`, which is
              light in the light theme and dark in the dark theme — so it only
              wants invert under `dark:`.
            */}
            <div
              className={`text-[14px] leading-relaxed prose prose-sm max-w-none prose-p:my-1 prose-pre:my-2 prose-ul:my-1 prose-ol:my-1 prose-code:text-[12px] ${
                isUser ? "prose-invert" : "dark:prose-invert"
              }`}
            >
              <ReactMarkdown
                remarkPlugins={[remarkGfm]}
                components={{
                  pre: ({ children }) => {
                    // Extract code content from the pre > code structure
                    const child = React.isValidElement(children)
                      ? children as React.ReactElement<{ children?: string; className?: string }>
                      : null;
                    const code = child?.props?.children || "";
                    const className = child?.props?.className || "";
                    if (typeof code === "string" && code) {
                      return <CodeBlock className={className}>{code}</CodeBlock>;
                    }
                    return <pre>{children}</pre>;
                  },
                  // Resolve `![](file.png)` from the agent — the file
                  // typically lives in the agent's workspace, served
                  // through /api/files. Click to open lightbox.
                  img: ({ src, alt }) => {
                    const resolved = rewriteMarkdownImageSrc(typeof src === "string" ? src : undefined, agentId);
                    if (!resolved) return null;
                    return (
                      // eslint-disable-next-line @next/next/no-img-element
                      <img
                        src={resolved}
                        alt={alt || ""}
                        className="max-w-full max-h-80 rounded-md border border-border/50 cursor-zoom-in my-1"
                        onClick={() => setLightboxURL(resolved)}
                      />
                    );
                  },
                }}
              >
                {msg.content}
              </ReactMarkdown>
            </div>
          </div>
        )}

        <div className={`flex items-center gap-1.5 mt-1 ${isUser ? "justify-end" : "justify-start"}`}>
          {msg.timestamp > 0 && (
            <span className="text-[10px] text-muted-foreground/50 select-none">
              {relativeTime(msg.timestamp)}
            </span>
          )}
          {!isUser && msg.content && (
            <button
              onClick={() => onCopy(msg)}
              className="opacity-0 group-hover:opacity-100 p-0.5 rounded hover:bg-muted/80 text-muted-foreground/50 hover:text-muted-foreground transition-all"
              title="Copy"
            >
              {copiedId === msg.id ? (
                <Check className="h-3 w-3 text-emerald-500" />
              ) : (
                <Copy className="h-3 w-3" />
              )}
            </button>
          )}
        </div>
      </div>

      {/* Lightbox — full-screen preview when an attachment is clicked.
          Click anywhere to dismiss; Esc handled implicitly via the
          backdrop button being focusable. */}
      {lightboxURL && (
        <div
          onClick={() => setLightboxURL(null)}
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 p-6 cursor-zoom-out"
        >
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src={lightboxURL}
            alt=""
            className="max-h-full max-w-full rounded-lg shadow-2xl"
            onClick={(e) => e.stopPropagation()}
          />
        </div>
      )}
    </div>
  );
}

// AttachmentTile renders a single attachment chip. Images get a
// thumbnail; non-images get a labelled file icon. Clicking an image
// pops the lightbox via the onPreview callback.
function AttachmentTile({
  att,
  onPreview,
}: {
  att: MessageAttachment;
  onPreview: (url: string) => void;
}) {
  if (att.type === "image") {
    return (
      <button
        type="button"
        onClick={() => onPreview(att.url)}
        className="block overflow-hidden rounded-lg border border-border/50 hover:border-primary/40 transition-colors cursor-zoom-in"
        title={att.name || "View image"}
      >
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img
          src={att.url}
          alt={att.name || ""}
          className="block max-h-56 max-w-[280px] object-contain bg-background/40"
        />
      </button>
    );
  }
  return (
    <div className="flex items-center gap-2 rounded-lg border border-border bg-muted/30 px-3 py-2">
      <FileText className="h-4 w-4 text-muted-foreground" />
      <span className="text-xs text-muted-foreground">{att.name || "file"}</span>
    </div>
  );
}

// ── Tool call group ──────────────────────────────────────────────────────────

function ToolCallGroup({ msg }: { msg: ChatMessage }) {
  const [open, setOpen] = useState(false);
  const [expanded, setExpanded] = useState<Record<string, boolean>>({});
  const tools = msg.toolCalls || [];
  const doneCount = tools.filter((tc) => tc.result != null).length;
  const allDone = doneCount === tools.length && tools.length > 0;

  const toggle = (id: string) => setExpanded((p) => ({ ...p, [id]: !p[id] }));

  return (
    <div className="flex gap-2 py-0.5 justify-start">
      <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-amber-500/10 mt-0.5">
        <Wrench className="h-3.5 w-3.5 text-amber-500" />
      </div>

      <div className="flex-1 max-w-[78%] space-y-2">
        {msg.content && (
          <div className="rounded-2xl rounded-bl-sm bg-muted px-4 py-2.5">
            <div className="text-[14px] leading-relaxed prose prose-sm dark:prose-invert max-w-none prose-p:my-1">
              <ReactMarkdown remarkPlugins={[remarkGfm]} components={{
                pre: ({ children }) => {
                  const child = React.isValidElement(children)
                    ? children as React.ReactElement<{ children?: string; className?: string }>
                    : null;
                  const code = child?.props?.children || "";
                  const className = child?.props?.className || "";
                  if (typeof code === "string" && code) {
                    return <CodeBlock className={className}>{code}</CodeBlock>;
                  }
                  return <pre>{children}</pre>;
                },
              }}>{msg.content}</ReactMarkdown>
            </div>
          </div>
        )}

        <div className="rounded-xl border border-border/60 bg-card overflow-hidden shadow-sm">
          {/* Header */}
          <button
            onClick={() => setOpen(!open)}
            className="flex w-full items-center gap-2.5 px-3.5 py-2.5 hover:bg-muted/30 transition-colors"
          >
            {allDone ? (
              <CheckCircle2 className="h-4 w-4 text-emerald-500 shrink-0" />
            ) : doneCount === tools.length && tools.length === 0 ? (
              <AlertCircle className="h-4 w-4 text-muted-foreground shrink-0" />
            ) : (
              <div className="h-4 w-4 rounded-full border-2 border-amber-500 border-t-transparent animate-spin shrink-0" />
            )}

            <div className="flex-1 min-w-0 text-left">
              <span className="text-xs font-medium text-foreground">
                {allDone
                  ? `Used ${tools.length} tool${tools.length !== 1 ? "s" : ""}`
                  : `Running tools… (${doneCount}/${tools.length})`}
              </span>
              <span className="ml-2 text-[10px] text-muted-foreground/60 font-mono">
                {tools.map((t) => t.name).join(" · ")}
              </span>
            </div>

            {open ? (
              <ChevronDown className="h-3.5 w-3.5 text-muted-foreground/60 shrink-0" />
            ) : (
              <ChevronRight className="h-3.5 w-3.5 text-muted-foreground/60 shrink-0" />
            )}
          </button>

          {/* Tool list */}
          {open && (
            <div className="border-t border-border/60 divide-y divide-border/40">
              {tools.map((tc) => (
                <div key={tc.id}>
                  <button
                    onClick={() => toggle(tc.id)}
                    className="flex w-full items-center gap-2.5 px-4 py-2 hover:bg-muted/20 transition-colors"
                  >
                    {tc.result === undefined ? (
                      <Clock className="h-3.5 w-3.5 text-amber-500/70 shrink-0 animate-pulse" />
                    ) : tc.error ? (
                      <AlertCircle className="h-3.5 w-3.5 text-destructive shrink-0" />
                    ) : (
                      <CheckCircle2 className="h-3.5 w-3.5 text-emerald-500 shrink-0" />
                    )}
                    <span className="text-xs font-mono font-medium text-foreground shrink-0">{tc.name}</span>
                    <span className="flex-1 truncate text-left text-[11px] text-muted-foreground/50 font-mono">
                      {(() => {
                        try {
                          const args = JSON.parse(tc.arguments);
                          return Object.entries(args).map(([k, v]) => `${k}=${JSON.stringify(v)}`).join(", ");
                        } catch {
                          return tc.arguments;
                        }
                      })()}
                    </span>
                    {expanded[tc.id] ? (
                      <ChevronDown className="h-3 w-3 text-muted-foreground/40 shrink-0" />
                    ) : (
                      <ChevronRight className="h-3 w-3 text-muted-foreground/40 shrink-0" />
                    )}
                  </button>

                  {expanded[tc.id] && (
                    <div className="px-4 pb-3 pt-1 space-y-2.5 bg-muted/10">
                      <div>
                        <p className="text-[10px] font-semibold uppercase tracking-widest text-muted-foreground/50 mb-1.5">Input</p>
                        <pre className="text-[11px] font-mono bg-muted/60 rounded-lg p-3 overflow-x-auto whitespace-pre-wrap break-all max-h-48 leading-relaxed">
                          {(() => {
                            try { return JSON.stringify(JSON.parse(tc.arguments), null, 2); }
                            catch { return tc.arguments; }
                          })()}
                        </pre>
                      </div>
                      {tc.result != null ? (
                        <div>
                          <p className="text-[10px] font-semibold uppercase tracking-widest text-muted-foreground/50 mb-1.5">Output</p>
                          <pre className={`text-[11px] font-mono rounded-lg p-3 overflow-x-auto whitespace-pre-wrap break-all max-h-64 leading-relaxed ${
                            tc.error ? "bg-destructive/5 text-destructive" : "bg-muted/60"
                          }`}>
                            {tc.result.length > 3000 ? tc.result.slice(0, 3000) + "\n… (truncated)" : tc.result}
                          </pre>
                        </div>
                      ) : (
                        <div className="flex items-center gap-2 py-2 text-xs text-muted-foreground/50 italic">
                          <div className="h-3 w-3 rounded-full border border-amber-500/60 border-t-transparent animate-spin" />
                          Executing…
                        </div>
                      )}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
