"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Inbox as InboxIcon,
  Trash2,
  Check,
  CheckCheck,
  Bell,
  BellOff,
  Clock,
  Globe,
  Bot,
} from "lucide-react";
import {
  getNotifications,
  markNotificationRead,
  markAllNotificationsRead,
  deleteNotification,
  type NotificationInfo,
} from "@/lib/api";
import {
  notifyUnreadChanged,
  setBrowserNotificationCallback,
} from "@/lib/use-notifications";

// Inbox is the OS-level feed of agent-driven events: cron job results,
// webhook deliveries, finished long-running tasks. Today the only
// producer is the cron scheduler; the same surface absorbs the rest
// as new sources come online.

type FilterMode = "all" | "unread";

const sourceIcon = (source: string) => {
  switch (source) {
    case "cron":
      return Clock;
    case "webhook":
      return Globe;
    case "agent":
      return Bot;
    default:
      return InboxIcon;
  }
};

const sourceColor = (source: string) => {
  switch (source) {
    case "cron":
      return "text-violet-500";
    case "webhook":
      return "text-blue-500";
    case "agent":
      return "text-emerald-500";
    default:
      return "text-muted-foreground";
  }
};

function formatRelative(iso: string) {
  try {
    const d = new Date(iso);
    const diffMs = Date.now() - d.getTime();
    const sec = Math.floor(diffMs / 1000);
    if (sec < 60) return "just now";
    if (sec < 3600) return `${Math.floor(sec / 60)}m ago`;
    if (sec < 86400) return `${Math.floor(sec / 3600)}h ago`;
    if (sec < 7 * 86400) return `${Math.floor(sec / 86400)}d ago`;
    return d.toLocaleDateString();
  } catch {
    return iso;
  }
}

export default function InboxPage() {
  const [items, setItems] = useState<NotificationInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<FilterMode>("all");
  const [selected, setSelected] = useState<NotificationInfo | null>(null);
  const [browserNotifGranted, setBrowserNotifGranted] = useState(false);

  const fetchData = () => {
    setLoading(true);
    getNotifications({ unreadOnly: filter === "unread", limit: 100 })
      .then((list) => setItems(list))
      .catch(() => setItems([]))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    fetchData();
    // Light polling: refresh the list every 5s while the inbox is open
    // so newly fired cron jobs surface without manual refresh. Stops
    // automatically when the user navigates away.
    const iv = setInterval(fetchData, 5000);
    return () => clearInterval(iv);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [filter]);

  // Detect existing browser notification permission so the toggle button
  // shows the right state on mount.
  useEffect(() => {
    if (typeof Notification !== "undefined") {
      setBrowserNotifGranted(Notification.permission === "granted");
    }
  }, []);

  // Wire the global polling callback to fire native browser notifications
  // for newly-arrived items when the user has granted permission. Cleans
  // up when leaving the page so other pages don't double-fire.
  useEffect(() => {
    if (!browserNotifGranted) return;
    setBrowserNotificationCallback((n) => {
      try {
        const notif = new Notification(n.title || "Pawnix", {
          body: n.body || "",
          tag: n.id,
          icon: "/icon.svg",
        });
        notif.onclick = () => {
          window.focus();
          notif.close();
        };
      } catch {
        // ignore — permission may have been revoked
      }
    });
    return () => setBrowserNotificationCallback(null);
  }, [browserNotifGranted]);

  const handleEnableBrowserNotifs = async () => {
    if (typeof Notification === "undefined") return;
    if (Notification.permission === "granted") {
      setBrowserNotifGranted(true);
      return;
    }
    if (Notification.permission === "denied") {
      alert(
        "Browser notifications are blocked. Re-enable them in your browser's site settings to receive Pawnix toasts."
      );
      return;
    }
    const result = await Notification.requestPermission();
    setBrowserNotifGranted(result === "granted");
  };

  const handleClick = async (n: NotificationInfo) => {
    setSelected(n);
    if (!n.read) {
      await markNotificationRead(n.id, true);
      setItems((prev) =>
        prev.map((x) => (x.id === n.id ? { ...x, read: true } : x))
      );
      notifyUnreadChanged();
    }
  };

  const handleMarkAllRead = async () => {
    await markAllNotificationsRead();
    setItems((prev) => prev.map((x) => ({ ...x, read: true })));
    notifyUnreadChanged();
  };

  const handleDelete = async (id: string) => {
    await deleteNotification(id);
    setItems((prev) => prev.filter((x) => x.id !== id));
    if (selected?.id === id) setSelected(null);
    notifyUnreadChanged();
  };

  const unreadCount = items.filter((x) => !x.read).length;

  return (
    <div className="flex h-screen flex-col">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-border px-6 py-4">
        <div>
          <h2 className="text-2xl font-semibold tracking-tight">Inbox</h2>
          <p className="text-sm text-muted-foreground mt-1">
            Cron jobs, webhooks, and agent-initiated alerts land here.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleEnableBrowserNotifs}
            className={browserNotifGranted ? "text-primary border-primary/30" : ""}
            title={
              browserNotifGranted
                ? "Browser notifications enabled — Pawnix will show OS-level toasts for new items"
                : "Click to allow browser notifications"
            }
          >
            {browserNotifGranted ? (
              <>
                <Bell className="h-3.5 w-3.5 mr-1.5" />
                Toasts on
              </>
            ) : (
              <>
                <BellOff className="h-3.5 w-3.5 mr-1.5" />
                Enable toasts
              </>
            )}
          </Button>
          {unreadCount > 0 && (
            <Button variant="outline" size="sm" onClick={handleMarkAllRead}>
              <CheckCheck className="h-3.5 w-3.5 mr-1.5" />
              Mark all read
            </Button>
          )}
        </div>
      </div>

      {/* Filter chips */}
      <div className="flex items-center gap-2 border-b border-border px-6 py-2">
        <button
          onClick={() => setFilter("all")}
          className={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
            filter === "all"
              ? "bg-primary/10 text-primary"
              : "text-muted-foreground hover:bg-muted"
          }`}
        >
          All
        </button>
        <button
          onClick={() => setFilter("unread")}
          className={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
            filter === "unread"
              ? "bg-primary/10 text-primary"
              : "text-muted-foreground hover:bg-muted"
          }`}
        >
          Unread{unreadCount > 0 && ` (${unreadCount})`}
        </button>
      </div>

      {/* Two-column layout: list on the left, detail on the right. */}
      <div className="flex flex-1 min-h-0">
        {/* List */}
        <div className="w-full md:w-1/2 lg:w-2/5 border-r border-border overflow-y-auto">
          {loading ? (
            <div className="space-y-2 p-4">
              {[1, 2, 3].map((i) => (
                <Skeleton key={i} className="h-16 w-full" />
              ))}
            </div>
          ) : items.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-center px-6">
              <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-primary/10 mb-4">
                <InboxIcon className="h-7 w-7 text-primary" />
              </div>
              <p className="text-sm text-muted-foreground">
                {filter === "unread"
                  ? "All caught up — no unread notifications."
                  : "No notifications yet. Cron jobs and webhooks will appear here when they fire."}
              </p>
              {filter === "all" && (
                <Link
                  href="/cron"
                  className="text-xs text-primary hover:underline mt-3"
                >
                  Configure cron jobs →
                </Link>
              )}
            </div>
          ) : (
            <ul className="divide-y divide-border">
              {items.map((n) => {
                const Icon = sourceIcon(n.source);
                const isSelected = selected?.id === n.id;
                return (
                  <li
                    key={n.id}
                    onClick={() => handleClick(n)}
                    className={`group flex items-start gap-3 px-4 py-3 cursor-pointer transition-colors ${
                      isSelected
                        ? "bg-muted"
                        : "hover:bg-muted/50"
                    } ${!n.read ? "border-l-2 border-primary" : "border-l-2 border-transparent"}`}
                  >
                    <Icon className={`h-4 w-4 mt-1 shrink-0 ${sourceColor(n.source)}`} />
                    <div className="flex-1 min-w-0">
                      <div className="flex items-start justify-between gap-2">
                        <p
                          className={`text-sm line-clamp-1 ${
                            n.read ? "text-foreground/80" : "font-semibold text-foreground"
                          }`}
                        >
                          {n.title}
                        </p>
                        <span className="text-[10px] text-muted-foreground whitespace-nowrap shrink-0">
                          {formatRelative(n.createdAt)}
                        </span>
                      </div>
                      <p className="text-xs text-muted-foreground line-clamp-1 mt-0.5">
                        {n.body || "(no content)"}
                      </p>
                      {n.agentId && (
                        <Badge
                          variant="outline"
                          className="mt-1.5 text-[10px] px-1.5 py-0 h-4"
                        >
                          {n.agentId}
                        </Badge>
                      )}
                    </div>
                  </li>
                );
              })}
            </ul>
          )}
        </div>

        {/* Detail pane */}
        <div className="hidden md:flex md:w-1/2 lg:w-3/5 flex-col overflow-y-auto">
          {selected ? (
            <div className="p-6">
              <div className="flex items-start justify-between gap-4 mb-4">
                <div className="flex-1 min-w-0">
                  <h3 className="text-lg font-semibold text-foreground">
                    {selected.title}
                  </h3>
                  <div className="flex items-center gap-2 mt-2 text-xs text-muted-foreground">
                    <Badge variant="outline" className="text-[10px]">
                      {selected.source}
                    </Badge>
                    {selected.agentId && (
                      <Badge variant="outline" className="text-[10px]">
                        {selected.agentId}
                      </Badge>
                    )}
                    <span>{new Date(selected.createdAt).toLocaleString()}</span>
                  </div>
                </div>
                <div className="flex items-center gap-1">
                  {selected.read ? (
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8"
                      title="Mark as unread"
                      onClick={async () => {
                        await markNotificationRead(selected.id, false);
                        setItems((prev) =>
                          prev.map((x) =>
                            x.id === selected.id ? { ...x, read: false } : x
                          )
                        );
                        setSelected({ ...selected, read: false });
                        notifyUnreadChanged();
                      }}
                    >
                      <Check className="h-4 w-4" />
                    </Button>
                  ) : null}
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 text-muted-foreground hover:text-destructive"
                    title="Delete"
                    onClick={() => handleDelete(selected.id)}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              </div>

              <div className="prose prose-sm dark:prose-invert max-w-none whitespace-pre-wrap text-sm leading-relaxed text-foreground/90">
                {selected.body || (
                  <span className="text-muted-foreground italic">(no content)</span>
                )}
              </div>

              {selected.link && (
                <div className="mt-6">
                  <Link
                    href={selected.link}
                    className="text-sm text-primary hover:underline"
                  >
                    Open source →
                  </Link>
                </div>
              )}
            </div>
          ) : (
            <div className="flex flex-1 items-center justify-center text-sm text-muted-foreground">
              Select a notification to read.
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
