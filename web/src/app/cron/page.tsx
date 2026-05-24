"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";
import { Clock, Inbox, Pencil, Play, Plus, Send, Hash, MessageCircle, Trash2 } from "lucide-react";
import {
  getCronJobs,
  createCronJob,
  updateCronJob,
  deleteCronJob,
  runCronJobNow,
  getAgents,
  getChannelsDetailed,
  type CronJobInfo,
  type AgentDetail,
  type ChannelDetail,
} from "@/lib/api";

// DeliveryTarget describes one possible "where does the cron job
// reply land" choice. Inbox is always available; one entry per
// (channel, account) combo that has MyChatID configured.
//
// `value` is the wire form passed back when creating a job. It uses
// "channel:accountId" so the form state is a single string.
type DeliveryTarget = {
  value: string;            // "inbox" | "telegram:main" | ...
  label: string;            // "Inbox" | "Telegram (main → 8175643861)"
  channel: string;          // "" for inbox
  accountId: string;
  chatId: string;
  icon: React.ElementType;
};

const channelIcon = (channel: string): React.ElementType => {
  switch (channel) {
    case "telegram":
      return Send;
    case "discord":
      return Hash;
    case "slack":
      return MessageCircle;
    default:
      return Inbox;
  }
};

const channelColor = (channel: string): string => {
  switch (channel) {
    case "telegram":
      return "text-blue-500";
    case "discord":
      return "text-indigo-500";
    case "slack":
      return "text-emerald-500";
    default:
      return "text-muted-foreground";
  }
};

// channelDisplayName humanises the channel string for badges/options.
const channelDisplayName = (channel: string): string => {
  switch (channel) {
    case "telegram":
      return "Telegram";
    case "discord":
      return "Discord";
    case "slack":
      return "Slack";
    case "":
    case "web":
      return "Inbox";
    default:
      return channel;
  }
};

// buildDeliveryTargets walks the configured channels and returns the
// dropdown choices: Inbox (always) + one row per (channel, account)
// that has MyChatID set. Without MyChatID we still show the row but
// disable it with a hint, so the user understands why telegram isn't
// pickable yet.
function buildDeliveryTargets(channels: ChannelDetail[]): Array<DeliveryTarget & { disabled?: boolean; disabledReason?: string }> {
  const out: Array<DeliveryTarget & { disabled?: boolean; disabledReason?: string }> = [
    {
      value: "inbox",
      label: "Inbox (browser toast + dashboard)",
      channel: "",
      accountId: "",
      chatId: "",
      icon: Inbox,
    },
  ];
  for (const ch of channels) {
    if (!ch.enabled) continue;
    for (const acct of ch.accounts ?? []) {
      const hasChatId = (acct.myChatId ?? "").trim() !== "";
      out.push({
        value: `${ch.type}:${acct.id}`,
        label: hasChatId
          ? `${channelDisplayName(ch.type)} (${acct.id} → ${acct.myChatId})`
          : `${channelDisplayName(ch.type)} (${acct.id}) — set "My chat ID" in Channels`,
        channel: ch.type,
        accountId: acct.id,
        chatId: acct.myChatId ?? "",
        icon: channelIcon(ch.type),
        disabled: !hasChatId,
        disabledReason: hasChatId ? undefined : "Configure My chat ID in Channels first",
      });
    }
  }
  return out;
}

export default function CronPage() {
  const [jobs, setJobs] = useState<CronJobInfo[]>([]);
  const [agents, setAgents] = useState<AgentDetail[]>([]);
  const [channels, setChannels] = useState<ChannelDetail[]>([]);
  const [loading, setLoading] = useState(true);
  const [createOpen, setCreateOpen] = useState(false);
  const [deleteId, setDeleteId] = useState<string | null>(null);
  const [editJob, setEditJob] = useState<CronJobInfo | null>(null);
  const [saving, setSaving] = useState(false);

  const [newName, setNewName] = useState("");
  const [newSchedule, setNewSchedule] = useState("");
  const [newType, setNewType] = useState("cron");
  const [newAgentId, setNewAgentId] = useState("");
  const [newMessage, setNewMessage] = useState("");
  // newDelivery is the selected dropdown value ("inbox" or "channel:account").
  // Looked up at submit time against the current channels list to populate
  // the channel/accountId/chatId fields on the cron job record.
  const [newDelivery, setNewDelivery] = useState("inbox");

  const fetchData = () => {
    setLoading(true);
    Promise.all([getCronJobs(), getAgents(), getChannelsDetailed()])
      .then(([j, a, c]) => {
        setJobs(j);
        setAgents(a);
        setChannels(c);
      })
      .catch(() => {
        setJobs([]);
        setAgents([]);
        setChannels([]);
      })
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    fetchData();
  }, []);

  const deliveryTargets = buildDeliveryTargets(channels);

  const handleCreate = async () => {
    if (!newName.trim() || !newSchedule.trim()) return;
    setSaving(true);
    const target = deliveryTargets.find((t) => t.value === newDelivery)
      ?? deliveryTargets[0]; // fall back to Inbox
    await createCronJob({
      name: newName.trim(),
      type: newType,
      schedule: newSchedule.trim(),
      agentId: newAgentId,
      message: newMessage,
      channel: target.channel,
      accountId: target.accountId,
      chatId: target.chatId,
      enabled: true,
    });
    setCreateOpen(false);
    setNewName("");
    setNewSchedule("");
    setNewType("cron");
    setNewAgentId("");
    setNewMessage("");
    setNewDelivery("inbox");
    setSaving(false);
    fetchData();
  };

  const handleToggle = async (job: CronJobInfo) => {
    await updateCronJob(job.id, { enabled: !job.enabled });
    fetchData();
  };

  const handleDelete = async () => {
    if (!deleteId) return;
    await deleteCronJob(deleteId);
    setDeleteId(null);
    fetchData();
  };

  const handleRunNow = async (job: CronJobInfo) => {
    await runCronJobNow(job.id);
    // Bump the next-run column visually right away — the actual fire
    // happens within ~60s when the scheduler's poll loop ticks.
    fetchData();
  };

  const handleEdit = (job: CronJobInfo) => {
    setEditJob({ ...job });
  };

  const handleSaveEdit = async () => {
    if (!editJob) return;
    setSaving(true);
    await updateCronJob(editJob.id, {
      agentId: editJob.agentId,
      channel: editJob.channel,
      chatId: editJob.chatId,
      name: editJob.name,
      message: editJob.message,
    });
    setEditJob(null);
    setSaving(false);
    fetchData();
  };

  // formatRunTime renders the RFC3339 timestamps from the API as
  // human-readable wall clock with a relative hint. Bare "—" when
  // we have nothing yet (job created but never fired).
  const formatRunTime = (iso?: string) => {
    if (!iso) return "—";
    try {
      const d = new Date(iso);
      const now = Date.now();
      const diffMs = d.getTime() - now;
      const absMin = Math.abs(Math.round(diffMs / 60000));
      let rel: string;
      if (absMin < 1) {
        rel = "just now";
      } else if (absMin < 60) {
        rel = diffMs >= 0 ? `in ${absMin}m` : `${absMin}m ago`;
      } else if (absMin < 60 * 24) {
        const h = Math.round(absMin / 60);
        rel = diffMs >= 0 ? `in ${h}h` : `${h}h ago`;
      } else {
        const d2 = Math.round(absMin / (60 * 24));
        rel = diffMs >= 0 ? `in ${d2}d` : `${d2}d ago`;
      }
      return `${d.toLocaleString()} (${rel})`;
    } catch {
      return iso;
    }
  };

  const typeColor = (type: string) => {
    const colors: Record<string, string> = {
      cron: "bg-violet-500/10 text-violet-600 dark:text-violet-400 border-violet-500/20",
      interval: "bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/20",
      exact: "bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20",
    };
    return colors[type] || "bg-muted text-muted-foreground border-border";
  };

  return (
    <div className="p-6 space-y-6 max-w-5xl mx-auto">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-semibold tracking-tight">Cron Jobs</h2>
          <p className="text-sm text-muted-foreground mt-1">
            Schedule automated agent tasks
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="h-4 w-4 mr-2" />
          New Job
        </Button>
      </div>

      <div className="rounded-lg border border-border bg-card">
        {loading ? (
          <div className="p-6 space-y-3">
            {[1, 2].map((i) => (
              <Skeleton key={i} className="h-14 w-full" />
            ))}
          </div>
        ) : jobs.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-primary/10 mb-4">
              <Clock className="h-7 w-7 text-primary" />
            </div>
            <p className="text-sm text-muted-foreground">No cron jobs configured</p>
            <Button
              onClick={() => setCreateOpen(true)}
              variant="outline"
              className="mt-4"
            >
              Create your first job
            </Button>
          </div>
        ) : (
          <div className="overflow-x-auto -mx-6 px-6">
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent">
                <TableHead>Name</TableHead>
                <TableHead>Schedule</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Agent</TableHead>
                <TableHead>Delivery</TableHead>
                <TableHead>Last Run</TableHead>
                <TableHead>Next Run</TableHead>
                <TableHead className="sticky right-[120px] w-[80px] bg-card z-20 border-l border-border">Enabled</TableHead>
                <TableHead className="sticky right-0 w-[120px] bg-card z-20 text-right border-l border-border">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {jobs.map((job) => {
                const ChannelIcon = channelIcon(job.channel);
                // Delivery cell renders one of:
                //   - "Inbox" (when channel is empty / 'web')
                //   - "Telegram • main → 8175643861" (real IM)
                // The chatId line uses font-mono since it's a literal
                // identifier the user might want to verify at a glance.
                return (
                <TableRow key={job.id} className="hover:bg-muted/50 transition-colors">
                  <TableCell>
                    <div className="flex flex-col">
                      <span className="font-medium">{job.name}</span>
                      {job.message && (
                        <span className="text-xs text-muted-foreground line-clamp-1 max-w-[260px]">
                          {job.message}
                        </span>
                      )}
                    </div>
                  </TableCell>
                  <TableCell>
                    <code className="rounded bg-muted px-2 py-1 text-xs font-mono">
                      {job.schedule}
                    </code>
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline" className={typeColor(job.type)}>
                      {job.type}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <span className="text-sm text-muted-foreground">{job.agentId || "-"}</span>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-start gap-2">
                      <ChannelIcon className={`h-3.5 w-3.5 mt-0.5 shrink-0 ${channelColor(job.channel)}`} />
                      <div className="flex flex-col">
                        <span className="text-sm">{channelDisplayName(job.channel)}</span>
                        {job.channel && job.channel !== "web" && (
                          <span className="text-[10px] text-muted-foreground/50 leading-tight">+ Inbox</span>
                        )}
                        {job.chatId && (
                          <span className="text-[10px] text-muted-foreground font-mono leading-tight">
                            {job.accountId ? `${job.accountId} → ` : ""}{job.chatId}
                          </span>
                        )}
                      </div>
                    </div>
                  </TableCell>
                  <TableCell>
                    <span className="text-xs text-muted-foreground">
                      {formatRunTime(job.lastRun)}
                    </span>
                  </TableCell>
                  <TableCell>
                    <span className="text-xs text-muted-foreground">
                      {formatRunTime(job.nextRun)}
                    </span>
                  </TableCell>
                  <TableCell className="sticky right-[120px] w-[80px] bg-card z-10 border-l border-border">
                    <Switch
                      checked={job.enabled}
                      onCheckedChange={() => handleToggle(job)}
                    />
                  </TableCell>
                  <TableCell className="sticky right-0 w-[120px] bg-card z-10 text-right border-l border-border">
                    <div className="flex items-center justify-end gap-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-muted-foreground hover:text-primary"
                        title="Edit"
                        onClick={() => handleEdit(job)}
                      >
                        <Pencil className="h-3.5 w-3.5" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-muted-foreground hover:text-primary"
                        title="Run now (fires on next scheduler poll, within 60s)"
                        onClick={() => handleRunNow(job)}
                      >
                        <Play className="h-3.5 w-3.5" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-muted-foreground hover:text-destructive"
                        title="Delete job"
                        onClick={() => setDeleteId(job.id)}
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
                );
              })}
            </TableBody>
          </Table>
          </div>
        )}
      </div>

      {/* Create Dialog */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Cron Job</DialogTitle>
            <DialogDescription>
              Schedule an automated agent task
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="space-y-2">
              <Label>Job Name</Label>
              <Input
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                placeholder="daily-report"
              />
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Type</Label>
                <Select value={newType} onValueChange={(v) => v && setNewType(v)}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="cron">Cron Expression</SelectItem>
                    <SelectItem value="interval">Interval</SelectItem>
                    <SelectItem value="exact">Exact Time</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>Schedule</Label>
                <Input
                  value={newSchedule}
                  onChange={(e) => setNewSchedule(e.target.value)}
                  placeholder={newType === "cron" ? "*/5 * * * *" : newType === "interval" ? "5m" : "14:30"}
                  className="font-mono"
                />
              </div>
            </div>
            <div className="space-y-2">
              <Label>Agent</Label>
              <Select value={newAgentId} onValueChange={(v) => v && setNewAgentId(v)}>
                <SelectTrigger>
                  <SelectValue placeholder="Select agent" />
                </SelectTrigger>
                <SelectContent>
                  {agents.map((a) => (
                    <SelectItem key={a.id} value={a.id}>
                      {a.id}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>Delivery</Label>
              <Select value={newDelivery} onValueChange={(v) => v && setNewDelivery(v)}>
                <SelectTrigger>
                  <SelectValue placeholder="Where to deliver" />
                </SelectTrigger>
                <SelectContent>
                  {deliveryTargets.map((t) => {
                    const TargetIcon = t.icon;
                    return (
                      <SelectItem
                        key={t.value}
                        value={t.value}
                        disabled={t.disabled}
                      >
                        <span className="inline-flex items-center gap-2">
                          <TargetIcon className={`h-3.5 w-3.5 ${channelColor(t.channel)}`} />
                          {t.label}
                        </span>
                      </SelectItem>
                    );
                  })}
                </SelectContent>
              </Select>
              <p className="text-xs text-muted-foreground">
                Where the agent&apos;s reply will land when the schedule fires.
                {deliveryTargets.length === 1 && (
                  <> Configure a channel + My chat ID under <span className="font-medium">Channels</span> to push to Telegram / Slack / Discord.</>
                )}
              </p>
            </div>
            <div className="space-y-2">
              <Label>Message</Label>
              <Textarea
                value={newMessage}
                onChange={(e) => setNewMessage(e.target.value)}
                placeholder="Generate a daily status report..."
                rows={3}
                className="resize-none"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleCreate}
              disabled={!newName.trim() || !newSchedule.trim() || saving}
            >
              {saving ? "Creating..." : "Create Job"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation */}
      {/* Edit Dialog */}
      <Dialog open={!!editJob} onOpenChange={(o) => { if (!o) setEditJob(null); }}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>Edit {editJob?.name}</DialogTitle>
          </DialogHeader>
          <div className="space-y-3 py-2">
            <div className="space-y-1.5">
              <Label>Agent</Label>
              <Select value={editJob?.agentId || ""} onValueChange={(v) => v && setEditJob((j) => j ? { ...j, agentId: v } : j)}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  {agents.map((a) => <SelectItem key={a.id} value={a.id}>{a.id}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <Label>Channel</Label>
              <Select value={editJob?.channel || ""} onValueChange={(v) => v && setEditJob((j) => j ? { ...j, channel: v } : j)}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="discord">Discord</SelectItem>
                  <SelectItem value="telegram">Telegram</SelectItem>
                  <SelectItem value="web">Web / Inbox</SelectItem>
                </SelectContent>
              </Select>
              {editJob?.channel && editJob.channel !== "web" && (
                <p className="text-[11px] text-muted-foreground/50">Results always arrive in Inbox too</p>
              )}
            </div>
            <div className="space-y-1.5">
              <Label>Name</Label>
              <Input value={editJob?.name || ""} onChange={(e) => setEditJob((j) => j ? { ...j, name: e.target.value } : j)} />
            </div>
            <div className="space-y-1.5">
              <Label>Chat ID</Label>
              <Input value={editJob?.chatId || ""} onChange={(e) => setEditJob((j) => j ? { ...j, chatId: e.target.value } : j)} />
            </div>
            <div className="space-y-1.5">
              <Label>Message（触发指令）</Label>
              <Textarea
                value={editJob?.message || ""}
                onChange={(e) => setEditJob((j) => j ? { ...j, message: e.target.value } : j)}
                rows={6}
                className="resize-none font-mono text-sm"
              />
              <p className="text-xs text-muted-foreground">
                定时触发时，这条指令会作为 prompt 直接发给 Agent。建议写清楚具体任务步骤。
              </p>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditJob(null)}>Cancel</Button>
            <Button onClick={handleSaveEdit} disabled={saving}>Save</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog open={!!deleteId} onOpenChange={() => setDeleteId(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Cron Job</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete this job? This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
