"use client";

import { useEffect, useState } from "react";
import {
  getMemoryUsage,
  type UsageOverview,
} from "@/lib/api";
import { Card, CardContent } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Activity,
  BarChart3,
  Database,
  PenLine,
  RefreshCw,
  Search,
  Target,
  Users,
  type LucideIcon,
} from "lucide-react";

// Color accents per origin tool, used for the little status dot so the eye
// can group activity by source at a glance. Unknown sources fall back to a
// muted dot rather than crashing on a missing key.
const SOURCE_DOT: Record<string, string> = {
  "claude-code": "bg-emerald-500",
  codex: "bg-sky-500",
  hermes: "bg-violet-500",
  codewiz: "bg-amber-500",
};

function sourceDot(source: string): string {
  return SOURCE_DOT[source] ?? "bg-muted-foreground/40";
}

// Icon + short verb per memory tool for the activity feed.
const TOOL_META: Record<string, { icon: LucideIcon; label: string }> = {
  memory_search: { icon: Search, label: "检索" },
  memory_write: { icon: PenLine, label: "写入" },
  memory_stats: { icon: BarChart3, label: "健康检查" },
};

// timeAgo renders a compact relative time ("3m", "2h", "5d"); anything
// older than a week falls back to an absolute date.
function timeAgo(iso: string): string {
  if (!iso) return "—";
  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) return "—";
  const secs = Math.floor((Date.now() - then) / 1000);
  if (secs < 60) return `${Math.max(secs, 0)}s`;
  const mins = Math.floor(secs / 60);
  if (mins < 60) return `${mins}m`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days}d`;
  return new Date(iso).toLocaleDateString();
}

function pct(n: number, d: number): string {
  if (d <= 0) return "—";
  return `${Math.round((n / d) * 100)}%`;
}

function clip(s: string, n: number): string {
  return s.length > n ? s.slice(0, n) + "…" : s;
}

function SourceTag({ source }: { source: string }) {
  return (
    <span className="inline-flex items-center gap-1.5">
      <span className={`inline-block h-2 w-2 rounded-full ${sourceDot(source)}`} />
      <span className="truncate">{source}</span>
    </span>
  );
}

function StatCard({
  icon: Icon,
  label,
  value,
  hint,
}: {
  icon: LucideIcon;
  label: string;
  value: React.ReactNode;
  hint?: string;
}) {
  return (
    <Card>
      <CardContent>
        <div className="flex items-center gap-2 text-muted-foreground">
          <Icon className="h-4 w-4" />
          <span className="text-xs font-medium">{label}</span>
        </div>
        <div className="mt-2 text-2xl font-semibold tabular-nums">{value}</div>
        {hint ? <div className="mt-1 text-xs text-muted-foreground">{hint}</div> : null}
      </CardContent>
    </Card>
  );
}

export default function MemoryPage() {
  const [data, setData] = useState<UsageOverview | null>(null);
  const [loading, setLoading] = useState(true);

  function refresh() {
    setLoading(true);
    getMemoryUsage()
      .then(setData)
      .catch(() => setData({ available: false }))
      .finally(() => setLoading(false));
  }

  useEffect(() => {
    refresh();
  }, []);

  const sources = data?.sources ?? [];
  const sessions = data?.sessions ?? [];
  const recent = data?.recent ?? [];
  const hasEvents = (data?.totalEvents ?? 0) > 0;

  return (
    <div className="p-6 space-y-6 max-w-6xl mx-auto">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h2 className="text-2xl font-semibold tracking-tight">Memory</h2>
          <p className="text-sm text-muted-foreground mt-1">
            跨工具共享记忆的使用情况：哪个工具在读写、多少轮检索、检索了什么主题、有没有命中。
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={refresh} disabled={loading}>
          <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
          刷新
        </Button>
      </div>

      {loading ? (
        <div className="space-y-4">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {[0, 1, 2, 3].map((i) => (
              <Skeleton key={i} className="h-24 w-full" />
            ))}
          </div>
          <Skeleton className="h-48 w-full" />
          <Skeleton className="h-48 w-full" />
        </div>
      ) : !data?.available ? (
        <div className="rounded-lg border border-dashed p-12 text-center">
          <Database className="h-8 w-8 mx-auto text-muted-foreground" />
          <p className="mt-3 text-sm font-medium">记忆遥测需要 PostgreSQL</p>
          <p className="mt-1 text-sm text-muted-foreground max-w-md mx-auto">
            当前存储不是 postgres，或未配置 DSN。记忆池与调用日志都依赖 pgvector——
            在设置里把 storage 切到 postgres 后，这里就会显示各工具的使用情况。
          </p>
        </div>
      ) : !hasEvents ? (
        <div className="rounded-lg border border-dashed p-12 text-center">
          <Activity className="h-8 w-8 mx-auto text-muted-foreground" />
          <p className="mt-3 text-sm font-medium">还没有记忆调用记录</p>
          <p className="mt-1 text-sm text-muted-foreground max-w-md mx-auto">
            等 claude-code / codex / hermes / codewiz 通过 <code className="text-xs">pawnix mcp</code>{" "}
            检索或写入记忆后，这里会出现每个工具、每个会话的使用明细。
          </p>
        </div>
      ) : (
        <div className="space-y-8">
          {/* Totals */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <StatCard
              icon={Users}
              label="会话数"
              value={data.totalSessions ?? 0}
              hint="按连接（子进程）计"
            />
            <StatCard
              icon={Search}
              label="检索轮次"
              value={data.totalSearches ?? 0}
              hint="≈ 对话轮数"
            />
            <StatCard icon={PenLine} label="写入" value={data.totalWrites ?? 0} />
            <StatCard
              icon={Target}
              label="命中率"
              value={pct(data.totalHits ?? 0, data.totalSearches ?? 0)}
              hint={`${data.totalHits ?? 0} / ${data.totalSearches ?? 0} 命中`}
            />
          </div>

          {/* By source */}
          <section className="space-y-3">
            <div className="flex items-baseline gap-2">
              <h3 className="text-sm font-medium">按工具</h3>
              <span className="text-xs text-muted-foreground">每个来源工具的读写与命中</span>
            </div>
            <div className="rounded-lg border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>来源</TableHead>
                    <TableHead className="text-right">会话</TableHead>
                    <TableHead className="text-right">检索</TableHead>
                    <TableHead className="text-right">写入</TableHead>
                    <TableHead className="text-right">命中率</TableHead>
                    <TableHead className="text-right">最近活动</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {sources.map((s) => (
                    <TableRow key={s.source}>
                      <TableCell className="font-medium">
                        <SourceTag source={s.source} />
                      </TableCell>
                      <TableCell className="text-right tabular-nums">{s.sessions}</TableCell>
                      <TableCell className="text-right tabular-nums">{s.searches}</TableCell>
                      <TableCell className="text-right tabular-nums">{s.writes}</TableCell>
                      <TableCell className="text-right tabular-nums">
                        {pct(s.hits, s.searches)}
                        <span className="text-xs text-muted-foreground ml-1">
                          ({s.hits}/{s.searches})
                        </span>
                      </TableCell>
                      <TableCell className="text-right tabular-nums text-muted-foreground">
                        {timeAgo(s.lastActive)}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </section>

          {/* By session */}
          <section className="space-y-3">
            <div className="flex items-baseline gap-2">
              <h3 className="text-sm font-medium">按会话</h3>
              <span className="text-xs text-muted-foreground">
                轮次按 memory_search 次数估算（设计上每轮对话检索一次）
              </span>
            </div>
            <div className="rounded-lg border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>来源</TableHead>
                    <TableHead className="text-right">轮次</TableHead>
                    <TableHead className="text-right">写入</TableHead>
                    <TableHead className="text-right">命中</TableHead>
                    <TableHead>主题</TableHead>
                    <TableHead className="text-right">最近</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {sessions.map((s) => (
                    <TableRow key={s.connectionId}>
                      <TableCell className="font-medium align-top">
                        <SourceTag source={s.source} />
                      </TableCell>
                      <TableCell className="text-right tabular-nums align-top">{s.turns}</TableCell>
                      <TableCell className="text-right tabular-nums align-top">{s.writes}</TableCell>
                      <TableCell className="text-right tabular-nums align-top">{s.hits}</TableCell>
                      <TableCell>
                        <div className="flex flex-wrap gap-1">
                          {s.topics.length === 0 ? (
                            <span className="text-xs text-muted-foreground">—</span>
                          ) : (
                            <>
                              {s.topics.slice(0, 3).map((t, i) => (
                                <Badge key={i} variant="outline" title={t}>
                                  {clip(t, 24)}
                                </Badge>
                              ))}
                              {s.topics.length > 3 ? (
                                <span className="text-xs text-muted-foreground self-center">
                                  +{s.topics.length - 3}
                                </span>
                              ) : null}
                            </>
                          )}
                        </div>
                      </TableCell>
                      <TableCell className="text-right tabular-nums text-muted-foreground align-top">
                        {timeAgo(s.lastSeen)}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </section>

          {/* Recent activity */}
          <section className="space-y-3">
            <div className="flex items-baseline gap-2">
              <h3 className="text-sm font-medium">最近活动</h3>
              <span className="text-xs text-muted-foreground">最新的记忆调用（最多 50 条）</span>
            </div>
            <div className="rounded-lg border divide-y">
              {recent.map((e, i) => {
                const meta = TOOL_META[e.tool] ?? { icon: Activity, label: e.tool };
                const Icon = meta.icon;
                return (
                  <div key={i} className="flex items-center gap-3 px-4 py-2.5 text-sm">
                    <Icon className="h-4 w-4 shrink-0 text-muted-foreground" />
                    <span className="w-28 shrink-0 text-xs text-muted-foreground">
                      <SourceTag source={e.source} />
                    </span>
                    <span className="flex-1 truncate" title={e.query || e.kind || meta.label}>
                      {e.tool === "memory_search"
                        ? e.query || "（空检索）"
                        : e.tool === "memory_write"
                          ? `写入 ${e.kind || "记忆"}`
                          : meta.label}
                    </span>
                    {e.tool === "memory_search" ? (
                      <Badge variant={e.hit ? "secondary" : "outline"} className="shrink-0">
                        {e.hit ? `命中 ${e.resultCount}` : "未命中"}
                      </Badge>
                    ) : e.tool === "memory_write" ? (
                      <Badge variant="secondary" className="shrink-0">
                        已写入
                      </Badge>
                    ) : null}
                    <span className="w-10 shrink-0 text-right text-xs text-muted-foreground tabular-nums">
                      {timeAgo(e.createdAt)}
                    </span>
                  </div>
                );
              })}
            </div>
          </section>
        </div>
      )}
    </div>
  );
}
