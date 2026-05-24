"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { getStatus, getTasks, type StatusResponse, type TaskInfo } from "@/lib/api";
import {
  Activity,
  Bot,
  Radio,
  Server,
  Brain,
  RefreshCw,
  MessageSquare,
  ArrowRight,
  Settings,
  Cpu,
  Zap,
  CheckCircle2,
  XCircle,
  Send,
} from "lucide-react";
import { Button } from "@/components/ui/button";

function StatCard({
  label,
  value,
  sub,
  icon: Icon,
  color,
}: {
  label: string;
  value: React.ReactNode;
  sub?: string;
  icon: React.ElementType;
  color: string;
}) {
  return (
    <div className="rounded-xl border border-border bg-card p-5 flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <span className="text-sm text-muted-foreground">{label}</span>
        <div className={`flex h-8 w-8 items-center justify-center rounded-lg ${color}`}>
          <Icon className="h-4 w-4" />
        </div>
      </div>
      <div>
        <div className="text-2xl font-semibold tracking-tight">{value}</div>
        {sub && <p className="text-xs text-muted-foreground mt-0.5">{sub}</p>}
      </div>
    </div>
  );
}

function QuickAction({
  href,
  icon: Icon,
  iconBg,
  label,
  desc,
}: {
  href: string;
  icon: React.ElementType;
  iconBg: string;
  label: string;
  desc: string;
}) {
  return (
    <Link href={href}>
      <div className="group flex items-center gap-4 rounded-xl border border-border bg-card p-4 transition-all hover:bg-muted/40 hover:border-border/80 hover:shadow-sm">
        <div className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-xl ${iconBg} transition-transform group-hover:scale-105`}>
          <Icon className="h-5 w-5" />
        </div>
        <div className="flex-1 min-w-0">
          <p className="text-sm font-medium leading-tight">{label}</p>
          <p className="text-xs text-muted-foreground mt-0.5">{desc}</p>
        </div>
        <ArrowRight className="h-4 w-4 text-muted-foreground/40 group-hover:text-muted-foreground group-hover:translate-x-0.5 transition-all shrink-0" />
      </div>
    </Link>
  );
}

export default function OverviewPage() {
  const [status, setStatus] = useState<StatusResponse | null>(null);
  const [tasks, setTasks] = useState<TaskInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);

  const fetchStatus = () => {
    setLoading(true);
    Promise.all([
      getStatus().then((s) => { setStatus(s); setLastUpdated(new Date()); }).catch(() => setStatus(null)),
      getTasks().then(setTasks).catch(() => setTasks([])),
    ]).finally(() => setLoading(false));
  };

  useEffect(() => {
    fetchStatus();
    const iv = setInterval(fetchStatus, 10000);
    return () => clearInterval(iv);
  }, []);

  if (loading && !status) {
    return (
      <div className="flex h-full min-h-screen items-center justify-center">
        <div className="h-7 w-7 animate-spin rounded-full border-2 border-muted border-t-primary" />
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6 max-w-5xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Dashboard</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            {lastUpdated
              ? `Updated ${lastUpdated.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", second: "2-digit" })}`
              : "Loading…"}
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={fetchStatus} disabled={loading}>
          <RefreshCw className={`h-4 w-4 mr-2 ${loading ? "animate-spin" : ""}`} />
          Refresh
        </Button>
      </div>

      {/* Stats */}
      <div className="grid gap-4 grid-cols-2 md:grid-cols-4">
        <StatCard
          label="Status"
          value={
            <Badge
              className={
                status?.running
                  ? "bg-emerald-500/15 text-emerald-600 dark:text-emerald-400 border-emerald-500/20 hover:bg-emerald-500/15"
                  : "bg-muted text-muted-foreground border-border"
              }
            >
              {status?.running ? (
                <><CheckCircle2 className="h-3 w-3 mr-1" /> Running</>
              ) : (
                <><XCircle className="h-3 w-3 mr-1" /> Stopped</>
              )}
            </Badge>
          }
          sub={status?.uptime ? `Uptime: ${status.uptime}` : "Not running"}
          icon={Activity}
          color="bg-emerald-500/10 text-emerald-500"
        />
        <StatCard
          label="Agents"
          value={status?.agents?.length ?? 0}
          sub="Loaded agents"
          icon={Bot}
          color="bg-violet-500/10 text-violet-500"
        />
        <StatCard
          label="Channels"
          value={status?.channels?.length ?? 0}
          sub="Connected"
          icon={Radio}
          color="bg-blue-500/10 text-blue-500"
        />
        <StatCard
          label="Port"
          value={<span className="font-mono">{status?.port ?? "—"}</span>}
          sub="Gateway port"
          icon={Server}
          color="bg-amber-500/10 text-amber-500"
        />
      </div>

      {/* Quick Actions */}
      <div>
        <h2 className="text-sm font-medium text-muted-foreground mb-3">Quick actions</h2>
        <div className="grid gap-3 md:grid-cols-3">
          <QuickAction href="/chat/" icon={MessageSquare} iconBg="bg-violet-500/10 text-violet-500" label="Chat" desc="Talk to your agents" />
          <QuickAction href="/agents/" icon={Bot} iconBg="bg-blue-500/10 text-blue-500" label="Agents" desc="Manage agent configs" />
          <QuickAction href="/settings/" icon={Settings} iconBg="bg-amber-500/10 text-amber-500" label="Settings" desc="Gateway configuration" />
        </div>
      </div>

      {/* Channel CTA — surfaces only when the user hasn't wired up a
          messaging bot yet. Replaces the missing onboard step that used
          to walk users through Telegram setup. Hidden the moment any
          channel exists so it doesn't nag long-time users. */}
      {status?.running && (!status?.channels || status.channels.length === 0) && (
        <Link href="/channels/" className="block">
          <div className="group rounded-xl border border-blue-500/30 bg-gradient-to-br from-blue-500/5 to-blue-500/10 p-5 transition-colors hover:from-blue-500/10 hover:to-blue-500/15">
            <div className="flex items-center gap-4">
              <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-blue-500 to-blue-600 transition-transform group-hover:scale-105">
                <Send className="h-6 w-6 text-white" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-base font-medium leading-tight">Connect a messaging bot</p>
                <p className="text-sm text-muted-foreground mt-0.5">
                  Hook up Telegram (or Discord / Slack) so your agents can reply in real chat — takes about a minute.
                </p>
              </div>
              <ArrowRight className="h-5 w-5 text-blue-500/60 group-hover:text-blue-500 group-hover:translate-x-0.5 transition-all shrink-0" />
            </div>
          </div>
        </Link>
      )}

      {/* Agents + Provider */}
      <div className="grid gap-4 md:grid-cols-2">
        {/* Agents table */}
        <div className="rounded-xl border border-border bg-card overflow-hidden">
          <div className="flex items-center gap-2 px-5 py-4 border-b border-border">
            <Cpu className="h-4 w-4 text-violet-500" />
            <h3 className="font-medium text-sm">Agents</h3>
            <Badge variant="secondary" className="ml-auto font-mono text-[10px]">
              {status?.agents?.length ?? 0}
            </Badge>
          </div>
          {status?.agents && status.agents.length > 0 ? (
            <div className="divide-y divide-border/50">
              {status.agents.map((agent) => {
                const agentTasks = tasks.filter((t) => t.agentId === agent.id);
                const running = agentTasks.some((t) => t.status === "running");
                const pending = agentTasks.filter((t) => t.status === "pending").length;
                return (
                <div key={agent.id} className="flex items-center justify-between px-5 py-3 hover:bg-muted/30 transition-colors">
                  <div className="flex items-center gap-3">
                    <div className="flex h-7 w-7 items-center justify-center rounded-full bg-primary/10 text-primary text-xs font-bold">
                      {agent.id[0]?.toUpperCase()}
                    </div>
                    <div>
                      <p className="text-sm font-medium leading-tight">{agent.id}</p>
                      <p className="text-[10px] text-muted-foreground font-mono">{agent.model}</p>
                    </div>
                  </div>
                  {running ? (
                    <Badge className="bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/20 text-[10px]">
                      Busy
                    </Badge>
                  ) : pending > 0 ? (
                    <Badge className="bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20 text-[10px]">
                      {pending} queued
                    </Badge>
                  ) : (
                    <Badge variant="outline" className="bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20 text-[10px]">
                      Idle
                    </Badge>
                  )}
                </div>
                );
              })}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-10 text-center px-5">
              <Bot className="h-8 w-8 text-muted-foreground/30 mb-2" />
              <p className="text-sm text-muted-foreground">No agents configured</p>
              <Link href="/agents/" className="text-xs text-primary mt-1 hover:underline">Add an agent →</Link>
            </div>
          )}
        </div>

        {/* Provider */}
        <div className="rounded-xl border border-border bg-card overflow-hidden">
          <div className="flex items-center gap-2 px-5 py-4 border-b border-border">
            <Brain className="h-4 w-4 text-amber-500" />
            <h3 className="font-medium text-sm">LLM Provider</h3>
          </div>
          {status?.provider ? (
            <div className="divide-y divide-border/50">
              {[
                { label: "Provider", value: status.provider.name || "default", mono: false },
                { label: "Model", value: status.provider.model, mono: true },
                { label: "API Base", value: status.provider.apiBase, mono: true },
                { label: "API Key", value: status.provider.apiKey, mono: true },
              ].map(({ label, value, mono }) => (
                <div key={label} className="flex items-center justify-between px-5 py-3">
                  <span className="text-sm text-muted-foreground">{label}</span>
                  <span className={`text-sm truncate max-w-[55%] text-right ${mono ? "font-mono text-[12px]" : ""}`}>
                    {value || "—"}
                  </span>
                </div>
              ))}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-10 text-center px-5">
              <Zap className="h-8 w-8 text-muted-foreground/30 mb-2" />
              <p className="text-sm text-muted-foreground">No provider configured</p>
              <Link href="/settings/" className="text-xs text-primary mt-1 hover:underline">Configure →</Link>
            </div>
          )}
        </div>
      </div>

      {/* Channels */}
      {status?.channels && status.channels.length > 0 && (
        <div className="rounded-xl border border-border bg-card overflow-hidden">
          <div className="flex items-center gap-2 px-5 py-4 border-b border-border">
            <Radio className="h-4 w-4 text-blue-500" />
            <h3 className="font-medium text-sm">Channels</h3>
            <Badge variant="secondary" className="ml-auto font-mono text-[10px]">{status.channels.length}</Badge>
          </div>
          <div className="divide-y divide-border/50">
            {status.channels.map((ch, i) => (
              <div key={i} className="flex items-center justify-between px-5 py-3 hover:bg-muted/30 transition-colors">
                <div className="flex items-center gap-3">
                  <div className="flex h-7 w-7 items-center justify-center rounded-full bg-blue-500/10">
                    <Radio className="h-3.5 w-3.5 text-blue-500" />
                  </div>
                  <div>
                    <p className="text-sm font-medium capitalize">{ch.type}</p>
                    {ch.botUsername && <p className="text-[10px] text-muted-foreground font-mono">@{ch.botUsername}</p>}
                  </div>
                </div>
                <Badge
                  variant="outline"
                  className={
                    ch.status === "connected" || ch.enabled !== false
                      ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20 text-[10px]"
                      : "bg-muted text-muted-foreground text-[10px]"
                  }
                >
                  {ch.status || (ch.enabled !== false ? "Active" : "Disabled")}
                </Badge>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Task Queue */}
      <div className="rounded-xl border border-border bg-card overflow-hidden">
        <div className="flex items-center gap-2 px-5 py-4 border-b border-border">
          <Activity className="h-4 w-4 text-orange-500" />
          <h3 className="font-medium text-sm">Task Queue</h3>
          <Badge variant="secondary" className="ml-auto font-mono text-[10px]">{tasks.length}</Badge>
        </div>
        {tasks.length > 0 ? (
          <div className="divide-y divide-border/50">
            {tasks.map((t) => (
              <div key={t.id} className="flex items-center justify-between px-5 py-3 hover:bg-muted/30 transition-colors">
                <div className="flex items-center gap-3 flex-1 min-w-0">
                  <span className={`flex h-1.5 w-1.5 rounded-full shrink-0 ${
                    t.status === "running" ? "bg-blue-500 animate-pulse" :
                    t.status === "pending" ? "bg-amber-500" :
                    t.status === "done" ? "bg-emerald-500" :
                    "bg-red-500"
                  }`} />
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <p className="text-sm font-medium">{t.agentId}</p>
                      <span className="text-[10px] text-muted-foreground capitalize">{t.status}</span>
                    </div>
                    <p className="text-[10px] text-muted-foreground/50 font-mono truncate" title={t.chatKey}>{t.chatKey}</p>
                  </div>
                </div>
                <div className="text-right">
                  {t.duration != null ? (
                    <span className="text-[11px] text-muted-foreground font-mono">{(t.duration / 1000).toFixed(1)}s</span>
                  ) : null}
                  {t.error ? (
                    <span className="text-[10px] text-red-500 line-clamp-1 max-w-[200px]" title={t.error}>{t.error.slice(0, 40)}</span>
                  ) : null}
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center py-8 text-center">
            <p className="text-sm text-muted-foreground">No tasks in queue</p>
          </div>
        )}
      </div>
    </div>
  );
}
