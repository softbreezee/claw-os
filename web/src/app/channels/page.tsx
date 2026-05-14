"use client";

import { useEffect, useMemo, useState } from "react";
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
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Radio,
  MessageCircle,
  Hash,
  Send,
  Plus,
  Trash2,
  CheckCircle2,
  XCircle,
  Loader2,
  RotateCcw,
  ExternalLink,
} from "lucide-react";
import {
  getChannelsDetailed,
  updateChannel,
  deleteChannel,
  testChannel,
  restartDaemon,
  waitForGateway,
  getAgents,
  type ChannelDetail,
  type ChannelAccount,
  type AgentDetail,
} from "@/lib/api";

const CHANNEL_META: Record<
  string,
  {
    icon: React.ElementType;
    gradient: string;
    label: string;
    helpUrl?: string;
    // Per-channel UI strings for the "My chat ID" field. Each channel
    // calls its recipient address something different — telegram says
    // "chat ID", slack says "user ID", etc. UI labels accordingly so
    // users see the term they actually need to look up.
    myIdLabel: string;
    myIdPlaceholder: string;
    myIdHelp?: string;     // short hint shown under the input
    myIdHelpUrl?: string;  // optional link to "how do I find this"
  }
> = {
  telegram: {
    icon: Send,
    gradient: "from-blue-500 to-blue-600",
    label: "Telegram",
    helpUrl: "https://core.telegram.org/bots/tutorial#obtain-your-bot-token",
    myIdLabel: "My chat ID",
    myIdPlaceholder: "8175643861",
    myIdHelp:
      "Numeric chat ID this bot will send notifications to. Tip: start a chat with @userinfobot — it will reply with your numeric ID.",
    myIdHelpUrl: "https://t.me/userinfobot",
  },
  discord: {
    icon: Hash,
    gradient: "from-indigo-500 to-indigo-600",
    label: "Discord",
    helpUrl: "https://discord.com/developers/docs/intro",
    myIdLabel: "My user ID",
    myIdPlaceholder: "123456789012345678",
    myIdHelp:
      "Your Discord user ID (developer mode → right-click your name → Copy User ID). The bot will DM you here.",
  },
  slack: {
    icon: MessageCircle,
    gradient: "from-emerald-500 to-emerald-600",
    label: "Slack",
    helpUrl: "https://api.slack.com/apps",
    myIdLabel: "My user ID",
    myIdPlaceholder: "U0XXXXXXX",
    myIdHelp:
      "Your Slack user ID (Profile → ⋯ → Copy member ID). The bot will DM you here.",
  },
};

const ALL_CHANNEL_TYPES = ["telegram", "discord", "slack"] as const;

// Accounts in the page state carry a transient `_dirty` flag so we know
// whether a token field holds the masked placeholder (untouched) or a
// real new value the user typed in. The server uses the same masked
// value as a sentinel for "keep existing", so we never have to leak
// the real token back to the client.
type AccountDraft = ChannelAccount & {
  _new?: boolean; // newly added in this session (id is editable)
  _tokenDirty?: boolean; // user typed a new token
  _testStatus?: "idle" | "ok" | "fail";
  _testMessage?: string;
};

type ChannelDraft = {
  type: string;
  enabled: boolean;
  appToken?: string;
  appTokenDirty?: boolean;
  accounts: AccountDraft[];
  status?: string;
  // Original (server) snapshot used for "Discard changes" reset.
  _original: ChannelDetail;
};

function toDraft(c: ChannelDetail): ChannelDraft {
  return {
    type: c.type,
    enabled: c.enabled,
    appToken: c.appToken ?? "",
    appTokenDirty: false,
    accounts: (c.accounts ?? []).map((a) => ({ ...a })),
    status: c.status,
    _original: c,
  };
}

export default function ChannelsPage() {
  const [drafts, setDrafts] = useState<ChannelDraft[]>([]);
  const [agents, setAgents] = useState<AgentDetail[]>([]);
  const [loading, setLoading] = useState(true);
  const [savingType, setSavingType] = useState<string | null>(null);
  const [deletingType, setDeletingType] = useState<string | null>(null);
  const [restartBanner, setRestartBanner] = useState<string | null>(null);
  const [addOpen, setAddOpen] = useState(false);
  const [addType, setAddType] = useState<string>("telegram");
  const [addToken, setAddToken] = useState("");
  const [addAgentId, setAddAgentId] = useState<string>("");
  const [addError, setAddError] = useState<string | null>(null);
  const [addSubmitting, setAddSubmitting] = useState(false);

  const fetchAll = async () => {
    setLoading(true);
    try {
      const [chs, ags] = await Promise.all([
        getChannelsDetailed().catch(() => [] as ChannelDetail[]),
        getAgents().catch(() => [] as AgentDetail[]),
      ]);
      setDrafts(chs.map(toDraft));
      setAgents(ags);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAll();
  }, []);

  const existingTypes = useMemo(() => new Set(drafts.map((d) => d.type)), [drafts]);
  const availableToAdd = ALL_CHANNEL_TYPES.filter((t) => !existingTypes.has(t));

  const updateDraft = (type: string, mut: (d: ChannelDraft) => ChannelDraft) => {
    setDrafts((all) => all.map((d) => (d.type === type ? mut(d) : d)));
  };

  const saveAndRestart = async (type: string) => {
    const draft = drafts.find((d) => d.type === type);
    if (!draft) return;
    setSavingType(type);
    try {
      const res = await updateChannel(type, {
        enabled: draft.enabled,
        appToken: draft.appTokenDirty ? draft.appToken : undefined,
        accounts: draft.accounts.map((a) => ({
          id: a.id.trim(),
          // Always pass `botToken` straight through — when untouched it
          // contains the masked placeholder ("ab12****wxyz") and the
          // server treats that as "keep the on-disk secret"; when the
          // user typed something new it's the raw token.
          botToken: a.botToken,
          agentId: a.agentId || undefined,
          myChatId: (a.myChatId ?? "").trim() || undefined,
        })),
      });
      if (!res.ok) {
        setRestartBanner(`Save failed: ${res.error ?? "unknown error"}`);
        return;
      }
      // Trigger restart so the new bot accounts actually connect.
      // We persist the banner across the restart so the user gets a
      // clear "applying changes…" cue.
      setRestartBanner("Restarting gateway to apply changes…");
      await restartDaemon();
      const back = await waitForGateway(20_000);
      if (back) {
        setRestartBanner("Gateway is back online. Channels reloaded.");
        await fetchAll();
      } else {
        setRestartBanner("Saved, but gateway didn't come back within 20s. Check logs.");
      }
      // Auto-clear banner after a short delay to keep the page tidy.
      setTimeout(() => setRestartBanner(null), 4000);
    } finally {
      setSavingType(null);
    }
  };

  const removeChannel = async (type: string) => {
    setDeletingType(null);
    setSavingType(type);
    try {
      const res = await deleteChannel(type);
      if (!res.ok) {
        setRestartBanner(`Delete failed: ${res.error ?? "unknown error"}`);
        return;
      }
      setRestartBanner("Restarting gateway to apply changes…");
      await restartDaemon();
      await waitForGateway(20_000);
      await fetchAll();
      setRestartBanner("Channel removed.");
      setTimeout(() => setRestartBanner(null), 4000);
    } finally {
      setSavingType(null);
    }
  };

  const submitAdd = async () => {
    setAddError(null);
    if (!addToken.trim()) {
      setAddError("Bot token is required.");
      return;
    }
    setAddSubmitting(true);
    try {
      // Use the human-friendly default account id "main" — the user can
      // rename it later in the channel card if they wire up a second bot.
      const res = await updateChannel(addType, {
        enabled: true,
        accounts: [
          {
            id: "main",
            botToken: addToken.trim(),
            agentId: addAgentId || undefined,
          },
        ],
      });
      if (!res.ok) {
        setAddError(res.error ?? "Failed to save channel");
        return;
      }
      setAddOpen(false);
      setAddToken("");
      setAddAgentId("");
      setRestartBanner("Restarting gateway to apply changes…");
      await restartDaemon();
      await waitForGateway(20_000);
      await fetchAll();
      setRestartBanner("Channel added and gateway reloaded.");
      setTimeout(() => setRestartBanner(null), 4000);
    } finally {
      setAddSubmitting(false);
    }
  };

  return (
    <div className="p-6 space-y-6 max-w-5xl mx-auto">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h2 className="text-2xl font-semibold tracking-tight">Channels</h2>
          <p className="text-sm text-muted-foreground mt-1">
            Connect bots from Telegram, Discord and Slack. Each bot can be bound to a different agent.
          </p>
        </div>

        <DropdownMenu>
          {/* base-ui's DropdownMenuTrigger renders its own <button>, so
              we apply button-like classes directly instead of wrapping
              another <Button> with `asChild` (which the primitive doesn't
              support). The class set mirrors the default Button variant. */}
          <DropdownMenuTrigger
            disabled={availableToAdd.length === 0 || loading}
            className="inline-flex items-center justify-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow-xs transition-colors hover:bg-primary/90 disabled:pointer-events-none disabled:opacity-50"
          >
            <Plus className="h-4 w-4" />
            Add channel
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            {availableToAdd.length === 0 ? (
              <DropdownMenuItem disabled>All channel types added</DropdownMenuItem>
            ) : (
              availableToAdd.map((t) => {
                const Icon = CHANNEL_META[t].icon;
                return (
                  <DropdownMenuItem
                    key={t}
                    onClick={() => {
                      setAddType(t);
                      setAddToken("");
                      setAddAgentId(agents[0]?.id ?? "");
                      setAddError(null);
                      setAddOpen(true);
                    }}
                  >
                    <Icon className="h-4 w-4 mr-2" />
                    {CHANNEL_META[t].label}
                  </DropdownMenuItem>
                );
              })
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      {restartBanner && (
        <div className="rounded-lg border border-blue-500/30 bg-blue-500/5 px-4 py-3 text-sm text-blue-700 dark:text-blue-300 flex items-center gap-2">
          <Loader2 className="h-4 w-4 animate-spin" />
          {restartBanner}
        </div>
      )}

      {loading ? (
        <div className="grid gap-4">
          {[1, 2].map((i) => (
            <Skeleton key={i} className="h-48" />
          ))}
        </div>
      ) : drafts.length === 0 ? (
        <EmptyState
          onConnect={(t) => {
            setAddType(t);
            setAddToken("");
            setAddAgentId(agents[0]?.id ?? "");
            setAddError(null);
            setAddOpen(true);
          }}
        />
      ) : (
        <div className="space-y-4">
          {drafts.map((d) => (
            <ChannelCard
              key={d.type}
              draft={d}
              agents={agents}
              busy={savingType === d.type}
              onChange={(mut) => updateDraft(d.type, mut)}
              onSave={() => saveAndRestart(d.type)}
              onDelete={() => setDeletingType(d.type)}
              onReset={() =>
                updateDraft(d.type, () => toDraft(d._original))
              }
            />
          ))}
        </div>
      )}

      {/* Add-channel quick dialog */}
      <Dialog open={addOpen} onOpenChange={setAddOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              Connect {CHANNEL_META[addType]?.label ?? addType}
            </DialogTitle>
            <DialogDescription>
              Paste the bot token to register a new bot. You can add more bots later.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="space-y-2">
              <Label>Channel type</Label>
              <Select value={addType} onValueChange={(v) => v && setAddType(v)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {availableToAdd.map((t) => (
                    <SelectItem key={t} value={t}>
                      {CHANNEL_META[t].label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>
                Bot token{" "}
                {CHANNEL_META[addType]?.helpUrl && (
                  <a
                    href={CHANNEL_META[addType]!.helpUrl}
                    target="_blank"
                    rel="noreferrer"
                    className="text-xs text-primary hover:underline ml-1 inline-flex items-center gap-1"
                  >
                    where do I find this?
                    <ExternalLink className="h-3 w-3" />
                  </a>
                )}
              </Label>
              <Input
                type="password"
                value={addToken}
                onChange={(e) => setAddToken(e.target.value)}
                placeholder={
                  addType === "telegram"
                    ? "123456789:ABCdefGHIjklMNOpqrsTUVwxyz"
                    : "Paste bot token…"
                }
                className="font-mono"
                autoFocus
              />
            </div>
            <div className="space-y-2">
              <Label>Default agent (optional)</Label>
              <Select
                value={addAgentId || "__none__"}
                onValueChange={(v) => setAddAgentId(v && v !== "__none__" ? v : "")}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Pick an agent…" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="__none__">— None (bind later) —</SelectItem>
                  {agents.map((a) => (
                    <SelectItem key={a.id} value={a.id}>
                      {a.id}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <p className="text-xs text-muted-foreground">
                Messages sent to this bot will be routed to the chosen agent.
              </p>
            </div>
            {addError && (
              <p className="text-sm text-destructive">{addError}</p>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setAddOpen(false)} disabled={addSubmitting}>
              Cancel
            </Button>
            <Button onClick={submitAdd} disabled={addSubmitting}>
              {addSubmitting ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Saving…
                </>
              ) : (
                "Save & Restart"
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete confirmation */}
      <AlertDialog open={!!deletingType} onOpenChange={(o) => !o && setDeletingType(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              Remove {CHANNEL_META[deletingType ?? ""]?.label ?? deletingType}?
            </AlertDialogTitle>
            <AlertDialogDescription>
              This deletes every bot under this channel and all their agent bindings. The gateway will be restarted to apply the change.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={() => deletingType && removeChannel(deletingType)}>
              Remove
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

function EmptyState({ onConnect }: { onConnect: (t: string) => void }) {
  return (
    <div className="rounded-lg border border-border bg-card">
      <div className="flex flex-col items-center justify-center py-16 px-6 text-center">
        <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-blue-500/10 mb-4">
          <Radio className="h-7 w-7 text-blue-500" />
        </div>
        <p className="text-base font-medium mb-1">No channels yet</p>
        <p className="text-sm text-muted-foreground mb-6 max-w-md">
          Connect a messaging bot so users can talk to your agents from Telegram, Discord or Slack.
        </p>
        <div className="flex gap-2">
          {ALL_CHANNEL_TYPES.map((t) => {
            const Icon = CHANNEL_META[t].icon;
            return (
              <Button key={t} variant="outline" onClick={() => onConnect(t)}>
                <Icon className="h-4 w-4 mr-2" />
                {CHANNEL_META[t].label}
              </Button>
            );
          })}
        </div>
      </div>
    </div>
  );
}

function ChannelCard({
  draft,
  agents,
  busy,
  onChange,
  onSave,
  onDelete,
  onReset,
}: {
  draft: ChannelDraft;
  agents: AgentDetail[];
  busy: boolean;
  onChange: (mut: (d: ChannelDraft) => ChannelDraft) => void;
  onSave: () => void;
  onDelete: () => void;
  onReset: () => void;
}) {
  const meta = CHANNEL_META[draft.type] ?? {
    icon: Radio,
    gradient: "from-zinc-500 to-zinc-600",
    label: draft.type,
  };
  const Icon = meta.icon;

  // Dirty when any token was edited, enabled flag flipped, account
  // added/removed/renamed, or any agent binding changed. Shallow compare
  // against the original snapshot.
  const original = draft._original;
  const dirty =
    draft.enabled !== original.enabled ||
    draft.appTokenDirty ||
    draft.accounts.length !== (original.accounts?.length ?? 0) ||
    draft.accounts.some((a, i) => {
      const o = original.accounts?.[i];
      if (!o) return true;
      return (
        a._new ||
        a._tokenDirty ||
        a.id !== o.id ||
        (a.agentId ?? "") !== (o.agentId ?? "") ||
        (a.myChatId ?? "") !== (o.myChatId ?? "")
      );
    });

  const addAccount = () => {
    onChange((d) => {
      const usedIds = new Set(d.accounts.map((a) => a.id));
      let i = d.accounts.length + 1;
      let id = `bot${i}`;
      while (usedIds.has(id)) {
        i++;
        id = `bot${i}`;
      }
      return {
        ...d,
        accounts: [
          ...d.accounts,
          {
            id,
            botToken: "",
            agentId: agents[0]?.id ?? "",
            _new: true,
            _tokenDirty: true,
          },
        ],
      };
    });
  };

  const updateAccount = (idx: number, mut: (a: AccountDraft) => AccountDraft) => {
    onChange((d) => ({
      ...d,
      accounts: d.accounts.map((a, i) => (i === idx ? mut(a) : a)),
    }));
  };

  const removeAccount = (idx: number) => {
    onChange((d) => ({ ...d, accounts: d.accounts.filter((_, i) => i !== idx) }));
  };

  const runTest = async (idx: number) => {
    const a = draft.accounts[idx];
    if (!a._tokenDirty || !a.botToken.trim()) {
      updateAccount(idx, (acc) => ({
        ...acc,
        _testStatus: "fail",
        _testMessage: "Type the real token first to run a connection test.",
      }));
      return;
    }
    updateAccount(idx, (acc) => ({ ...acc, _testStatus: "idle", _testMessage: "Testing…" }));
    const res = await testChannel(draft.type, a.botToken.trim());
    if (res.ok) {
      updateAccount(idx, (acc) => ({
        ...acc,
        _testStatus: "ok",
        _testMessage: res.botUsername ? `Connected as @${res.botUsername}` : "Connected",
        botUsername: res.botUsername,
      }));
    } else {
      updateAccount(idx, (acc) => ({
        ...acc,
        _testStatus: "fail",
        _testMessage: res.error ?? "Test failed",
      }));
    }
  };

  return (
    <div className="rounded-xl border border-border bg-card overflow-hidden">
      {/* Header */}
      <div className="flex items-center gap-4 px-5 py-4 border-b border-border">
        <div
          className={`flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br ${meta.gradient}`}
        >
          <Icon className="h-5 w-5 text-white" />
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <p className="text-base font-medium">{meta.label}</p>
            <Badge
              variant="outline"
              className={
                draft.enabled
                  ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20"
                  : "bg-muted text-muted-foreground border-border"
              }
            >
              <span
                className={`mr-1.5 inline-block h-1.5 w-1.5 rounded-full ${
                  draft.enabled ? "bg-emerald-500" : "bg-muted-foreground"
                }`}
              />
              {draft.enabled ? "Enabled" : "Disabled"}
            </Badge>
          </div>
          <p className="text-xs text-muted-foreground mt-0.5">
            {draft.accounts.length} bot{draft.accounts.length === 1 ? "" : "s"} configured
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Switch
            checked={draft.enabled}
            onCheckedChange={(v) => onChange((d) => ({ ...d, enabled: !!v }))}
          />
          <Button
            variant="ghost"
            size="sm"
            className="text-destructive hover:text-destructive"
            onClick={onDelete}
            disabled={busy}
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* Slack-only: app token */}
      {draft.type === "slack" && (
        <div className="px-5 py-3 border-b border-border bg-muted/20">
          <div className="space-y-2 max-w-md">
            <Label className="text-xs">App token (xapp-…)</Label>
            <Input
              type="password"
              value={draft.appToken ?? ""}
              onChange={(e) =>
                onChange((d) => ({
                  ...d,
                  appToken: e.target.value,
                  appTokenDirty: true,
                }))
              }
              className="font-mono text-xs"
              placeholder={draft.appToken ? "" : "xapp-…"}
            />
            <p className="text-xs text-muted-foreground">
              Required for Socket Mode. Shared across all bots in this workspace.
            </p>
          </div>
        </div>
      )}

      {/* Accounts list */}
      <div className="divide-y divide-border/50">
        {draft.accounts.length === 0 && (
          <div className="px-5 py-6 text-sm text-muted-foreground text-center">
            No bots yet. Click <span className="font-medium">Add bot</span> below to register one.
          </div>
        )}
        {draft.accounts.map((a, i) => (
          <AccountRow
            key={i}
            account={a}
            agents={agents}
            channelType={draft.type}
            onChange={(mut) => updateAccount(i, mut)}
            onTest={() => runTest(i)}
            onRemove={() => removeAccount(i)}
          />
        ))}
      </div>

      {/* Footer */}
      <div className="flex items-center justify-between gap-3 px-5 py-3 border-t border-border bg-muted/10">
        <Button variant="outline" size="sm" onClick={addAccount} disabled={busy}>
          <Plus className="h-4 w-4 mr-2" />
          Add bot
        </Button>
        <div className="flex items-center gap-2">
          {dirty && (
            <Button variant="ghost" size="sm" onClick={onReset} disabled={busy}>
              <RotateCcw className="h-3.5 w-3.5 mr-1.5" />
              Discard
            </Button>
          )}
          <Button size="sm" onClick={onSave} disabled={!dirty || busy}>
            {busy ? (
              <>
                <Loader2 className="h-3.5 w-3.5 mr-1.5 animate-spin" />
                Saving…
              </>
            ) : (
              "Save & Restart"
            )}
          </Button>
        </div>
      </div>
    </div>
  );
}

function AccountRow({
  account,
  agents,
  channelType,
  onChange,
  onTest,
  onRemove,
}: {
  account: AccountDraft;
  agents: AgentDetail[];
  channelType: string;
  onChange: (mut: (a: AccountDraft) => AccountDraft) => void;
  onTest: () => void;
  onRemove: () => void;
}) {
  const meta = CHANNEL_META[channelType];
  return (
    <div className="px-5 py-4 grid gap-3 md:grid-cols-12 items-start">
      {/* Account ID */}
      <div className="md:col-span-3 space-y-1.5">
        <Label className="text-xs">Bot ID</Label>
        <Input
          value={account.id}
          onChange={(e) => onChange((a) => ({ ...a, id: e.target.value }))}
          // Existing accounts shouldn't be renamed casually — bindings reference the id.
          // We only allow editing for newly-added rows.
          disabled={!account._new}
          className="font-mono text-sm"
          placeholder="main"
        />
      </div>

      {/* Bot Token */}
      <div className="md:col-span-4 space-y-1.5">
        <Label className="text-xs">Bot token</Label>
        <Input
          type="password"
          value={account.botToken}
          onChange={(e) =>
            onChange((a) => ({
              ...a,
              botToken: e.target.value,
              _tokenDirty: true,
              _testStatus: undefined,
              _testMessage: undefined,
            }))
          }
          className="font-mono text-xs"
          placeholder={account._new ? "Paste token…" : ""}
        />
        {account._testMessage && (
          <p
            className={`text-[11px] flex items-center gap-1 ${
              account._testStatus === "ok"
                ? "text-emerald-600 dark:text-emerald-400"
                : account._testStatus === "fail"
                ? "text-destructive"
                : "text-muted-foreground"
            }`}
          >
            {account._testStatus === "ok" && <CheckCircle2 className="h-3 w-3" />}
            {account._testStatus === "fail" && <XCircle className="h-3 w-3" />}
            {account._testMessage}
          </p>
        )}
      </div>

      {/* Agent picker */}
      <div className="md:col-span-3 space-y-1.5">
        <Label className="text-xs">Bound agent</Label>
        <Select
          value={account.agentId || "__none__"}
          onValueChange={(v) =>
            onChange((a) => ({ ...a, agentId: v && v !== "__none__" ? v : "" }))
          }
        >
          <SelectTrigger className="text-sm">
            <SelectValue placeholder="None" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="__none__">— None —</SelectItem>
            {agents.map((ag) => (
              <SelectItem key={ag.id} value={ag.id}>
                {ag.id}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* My chat ID — the user's "send to me" address.
          Per-channel label/placeholder so users see "chat ID" for
          telegram, "user ID" for slack/discord, etc. Without this
          configured, agents can't push notifications to this channel
          (notify tool returns a friendly error). */}
      <div className="md:col-span-12 space-y-1.5 -mt-1">
        <div className="flex items-center gap-2">
          <Label className="text-xs">{meta?.myIdLabel ?? "My address"}</Label>
          {meta?.myIdHelpUrl && (
            <a
              href={meta.myIdHelpUrl}
              target="_blank"
              rel="noreferrer"
              className="text-[10px] text-primary hover:underline inline-flex items-center gap-0.5"
            >
              how to find?
              <ExternalLink className="h-2.5 w-2.5" />
            </a>
          )}
        </div>
        <Input
          value={account.myChatId ?? ""}
          onChange={(e) =>
            onChange((a) => ({ ...a, myChatId: e.target.value }))
          }
          className="font-mono text-xs max-w-md"
          placeholder={meta?.myIdPlaceholder ?? ""}
        />
        {meta?.myIdHelp && (
          <p className="text-[11px] text-muted-foreground max-w-2xl">
            {meta.myIdHelp}
          </p>
        )}
      </div>

      {/* Actions */}
      <div className="md:col-span-12 flex items-end justify-end gap-1 pt-1 -mt-2">
        {channelType === "telegram" && (
          <Button
            variant="ghost"
            size="sm"
            onClick={onTest}
            disabled={!account._tokenDirty}
            title={account._tokenDirty ? "Verify the token via Telegram getMe" : "Edit the token to enable test"}
          >
            Test
          </Button>
        )}
        <Button
          variant="ghost"
          size="sm"
          className="text-destructive hover:text-destructive"
          onClick={onRemove}
        >
          <Trash2 className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}
