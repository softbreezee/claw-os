"use client";

import { useEffect, useState, useRef, useCallback } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  getStatus,
  getChatHistory,
  getChatSessions,
  sendChatStream,
  deleteChatSession,
  type AgentInfo,
  type ChatHistoryMessage,
  type ChatStreamEvent,
} from "@/lib/api";
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

interface ChatMessage {
  id: string;
  role: "user" | "agent" | "tool-group";
  content: string;
  timestamp: number;
  toolCalls?: ToolCall[];
}

interface ChatSession {
  id: string;
  preview: string;
}

function generateSessionId() {
  return `s-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
}

function relativeTime(ts: number): string {
  if (!ts) return "";
  const diff = Math.floor((Date.now() - ts) / 1000);
  if (diff < 60) return "just now";
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  return new Date(ts).toLocaleDateString([], { month: "short", day: "numeric" });
}

function buildChatMessages(history: ChatHistoryMessage[]): ChatMessage[] {
  const msgs: ChatMessage[] = [];
  let i = 0;
  while (i < history.length) {
    const h = history[i];
    if (h.role === "user") {
      msgs.push({ id: `h-${i}`, role: "user", content: h.content || "", timestamp: 0 });
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
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState("");
  const [sending, setSending] = useState(false);
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const abortRef = useRef<(() => void) | null>(null);

  useEffect(() => {
    getStatus()
      .then((status) => {
        if (status.agents?.length > 0) {
          setAgents(status.agents);
          setSelectedAgent(status.agents[0].id);
        }
      })
      .catch(() => {});
  }, []);

  const loadSessions = useCallback((agentId: string) => {
    getChatSessions(agentId)
      .then((list) => setSessions(list || []))
      .catch(() => setSessions([]));
  }, []);

  useEffect(() => {
    if (!selectedAgent) return;
    loadSessions(selectedAgent);
  }, [selectedAgent, loadSessions]);

  useEffect(() => {
    if (!selectedAgent || !sessionId) return;
    getChatHistory(selectedAgent, sessionId)
      .then((history) => {
        if (!history || history.length === 0) { setMessages([]); return; }
        setMessages(buildChatMessages(history));
      })
      .catch(() => setMessages([]));
  }, [selectedAgent, sessionId]);

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

  const handleSend = useCallback(async (text?: string) => {
    const msg = (text ?? input).trim();
    if (!msg || !selectedAgent || sending) return;

    setInput("");
    const userMsgId = `u-${Date.now()}`;
    setMessages((prev) => [...prev, { id: userMsgId, role: "user", content: msg, timestamp: Date.now() }]);
    setSending(true);

    let aborted = false;
    abortRef.current = () => { aborted = true; };

    let curGroupId = `tg-${Date.now()}`;
    let curCalls: ToolCall[] = [];
    let curContent = "";

    const startNewGroup = () => {
      curGroupId = `tg-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`;
      curCalls = [];
      curContent = "";
    };
    startNewGroup();

    try {
      await sendChatStream(selectedAgent, sessionId, msg, (evt: ChatStreamEvent) => {
        if (aborted) return;
        switch (evt.type) {
          case "content": {
            const content = evt.data?.content || "";
            if (curCalls.length > 0) startNewGroup();
            curContent = content;
            setMessages((prev) => [...prev, { id: `a-${Date.now()}`, role: "agent", content, timestamp: Date.now() }]);
            break;
          }
          case "tool_call": {
            curCalls.push({ id: evt.data?.id || "", name: evt.data?.name || "", arguments: evt.data?.arguments || "{}" });
            const groupId = curGroupId;
            const calls = [...curCalls];
            const content = curContent;
            setMessages((prev) => {
              const last = prev[prev.length - 1];
              if (content && last?.role === "agent" && last.content === content) {
                return [...prev.slice(0, -1), { id: groupId, role: "tool-group" as const, content, timestamp: Date.now(), toolCalls: calls }];
              }
              const idx = prev.findIndex((m) => m.id === groupId);
              if (idx >= 0) { const u = [...prev]; u[idx] = { ...u[idx], toolCalls: calls }; return u; }
              return [...prev, { id: groupId, role: "tool-group" as const, content, timestamp: Date.now(), toolCalls: calls }];
            });
            break;
          }
          case "tool_result": {
            const tc = curCalls.find((c) => c.id === (evt.data?.id || ""));
            if (tc) tc.result = evt.data?.result || "";
            const groupId = curGroupId;
            const calls = [...curCalls];
            setMessages((prev) => {
              const idx = prev.findIndex((m) => m.id === groupId);
              if (idx < 0) return prev;
              const u = [...prev]; u[idx] = { ...u[idx], toolCalls: calls }; return u;
            });
            break;
          }
        }
      });
      loadSessions(selectedAgent);
    } catch {
      if (!aborted) {
        setMessages((prev) => [
          ...prev,
          { id: `e-${Date.now()}`, role: "agent", content: "⚠️ Failed to get a response. Is the gateway running?", timestamp: Date.now() },
        ]);
      }
    } finally {
      setSending(false);
      abortRef.current = null;
      textareaRef.current?.focus();
    }
  }, [input, selectedAgent, sessionId, sending, loadSessions]);

  const handleStop = () => {
    abortRef.current?.();
    setSending(false);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) { e.preventDefault(); handleSend(); }
  };

  const handleCopy = (msg: ChatMessage) => {
    navigator.clipboard.writeText(msg.content);
    setCopiedId(msg.id);
    setTimeout(() => setCopiedId(null), 1500);
  };

  const handleNewChat = useCallback(() => {
    setSessionId(generateSessionId());
    setMessages([]);
  }, []);

  const handleDeleteSession = async (sid: string, e: React.MouseEvent) => {
    e.stopPropagation();
    if (!selectedAgent) return;
    try {
      await deleteChatSession(selectedAgent, sid);
      loadSessions(selectedAgent);
      if (sessionId === sid) handleNewChat();
    } catch { /* ignore */ }
  };

  const currentAgent = agents.find((a) => a.id === selectedAgent);

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
                onClick={() => setSessionId(s.id)}
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
        <div className="shrink-0 px-4 pb-5 pt-3 border-t border-border bg-card/20">
          <div className="mx-auto max-w-3xl">
            <div className="flex flex-col rounded-xl border border-border bg-card shadow-sm focus-within:border-primary/40 focus-within:shadow-md transition-all">
              <textarea
                ref={textareaRef}
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder={selectedAgent ? `Message ${selectedAgent}…` : "Select an agent first"}
                disabled={!selectedAgent}
                rows={1}
                className="w-full resize-none bg-transparent px-4 pt-3 pb-2 text-[14px] leading-relaxed placeholder:text-muted-foreground/40 outline-none disabled:opacity-40"
                style={{ maxHeight: 180, minHeight: 42 }}
              />
              <div className="flex items-center justify-between px-3 pb-2.5">
                <p className="text-[11px] text-muted-foreground/40 select-none">
                  {sending ? "Responding…" : "↵ Send  ·  ⇧↵ New line"}
                </p>
                <div className="flex items-center gap-2">
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
                      disabled={!input.trim() || !selectedAgent}
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

function MessageBubble({
  msg,
  copiedId,
  onCopy,
}: {
  msg: ChatMessage;
  copiedId: string | null;
  onCopy: (m: ChatMessage) => void;
}) {
  const isUser = msg.role === "user";

  return (
    <div className={`flex gap-2 py-0.5 ${isUser ? "justify-end" : "justify-start"}`}>
      {!isUser && (
        <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary/10 mt-0.5">
          <Bot className="h-3.5 w-3.5 text-primary" />
        </div>
      )}

      <div className={`group max-w-[78%] ${isUser ? "order-first" : ""}`}>
        <div
          className={`rounded-2xl px-4 py-2.5 ${
            isUser
              ? "bg-primary text-primary-foreground rounded-br-sm"
              : "bg-muted rounded-bl-sm"
          }`}
        >
          <div className="text-[14px] leading-relaxed prose prose-sm dark:prose-invert max-w-none prose-p:my-1 prose-pre:my-2 prose-ul:my-1 prose-ol:my-1 prose-code:text-[12px]">
            <ReactMarkdown remarkPlugins={[remarkGfm]}>{msg.content}</ReactMarkdown>
          </div>
        </div>

        <div className={`flex items-center gap-1.5 mt-1 ${isUser ? "justify-end" : "justify-start"}`}>
          {msg.timestamp > 0 && (
            <span className="text-[10px] text-muted-foreground/50 select-none">
              {relativeTime(msg.timestamp)}
            </span>
          )}
          {!isUser && (
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
              <ReactMarkdown remarkPlugins={[remarkGfm]}>{msg.content}</ReactMarkdown>
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
