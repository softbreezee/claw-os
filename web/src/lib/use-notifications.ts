"use client";

import { useEffect, useState } from "react";
import { getUnreadNotificationCount, type NotificationInfo } from "@/lib/api";

// Polling interval for the unread badge. 5s is a comfortable trade-off:
// fast enough that a fresh cron-job notification feels "live", slow
// enough that it doesn't dominate the request log on a quiet system.
// When we add SSE (next milestone) this falls back to a heartbeat.
const POLL_MS = 5_000;

let lastUnreadCount = 0;
const subscribers = new Set<(n: number) => void>();
let pollHandle: ReturnType<typeof setInterval> | null = null;
let inflight = false;
let toastEnabled = false;
let lastSeenIds = new Set<string>();
let onNewNotificationCallback: ((notif: NotificationInfo) => void) | null = null;

// useUnreadNotifications returns the live unread count and ensures
// exactly one polling loop runs across the whole UI regardless of how
// many sidebar/inbox/topbar components subscribe.
export function useUnreadNotifications() {
  const [count, setCount] = useState(lastUnreadCount);

  useEffect(() => {
    subscribers.add(setCount);
    setCount(lastUnreadCount);
    ensurePolling();
    return () => {
      subscribers.delete(setCount);
      if (subscribers.size === 0 && pollHandle) {
        clearInterval(pollHandle);
        pollHandle = null;
      }
    };
  }, []);

  return count;
}

function ensurePolling() {
  if (pollHandle) return;
  refresh(); // immediate first tick so the badge appears without the 5s wait
  pollHandle = setInterval(refresh, POLL_MS);
}

async function refresh() {
  if (inflight) return;
  inflight = true;
  try {
    const r = await getUnreadNotificationCount();
    if (r.count !== lastUnreadCount) {
      lastUnreadCount = r.count;
      subscribers.forEach((fn) => fn(r.count));
    }
    // Toast detection: when the count goes UP (new notifications) and
    // the user has enabled browser notifications, fetch the newest
    // unread items and toast the ones we haven't seen.
    if (toastEnabled && r.count > 0 && onNewNotificationCallback) {
      await checkForNewToasts();
    }
  } catch {
    // Silently swallow — sidebar shouldn't error on transient network blips.
  } finally {
    inflight = false;
  }
}

async function checkForNewToasts() {
  try {
    const res = await fetch("/api/notifications?unreadOnly=true&limit=10");
    if (!res.ok) return;
    const list: NotificationInfo[] = await res.json();
    for (const n of list) {
      if (!lastSeenIds.has(n.id)) {
        lastSeenIds.add(n.id);
        onNewNotificationCallback?.(n);
      }
    }
    // Cap memory: keep the last 200 IDs we've seen.
    if (lastSeenIds.size > 200) {
      const arr = Array.from(lastSeenIds);
      lastSeenIds = new Set(arr.slice(arr.length - 200));
    }
  } catch {
    // ignore
  }
}

// enableBrowserNotifications wires the polling loop to fire a callback
// for every freshly-arrived unread notification. The callback is what
// actually calls `new Notification(...)` — the hook stays UI-free so
// the same plumbing can be repurposed for in-app toast components later.
//
// Idempotent: calling twice replaces the callback. Pass null to disable.
export function setBrowserNotificationCallback(cb: ((n: NotificationInfo) => void) | null) {
  onNewNotificationCallback = cb;
  toastEnabled = cb !== null;
  if (toastEnabled) {
    // Seed lastSeenIds with current unread set so we don't blast toasts
    // for everything in the inbox the first time the user enables it.
    fetch("/api/notifications?unreadOnly=true&limit=200")
      .then((r) => r.json())
      .then((list: NotificationInfo[]) => {
        list.forEach((n) => lastSeenIds.add(n.id));
      })
      .catch(() => {});
  }
}

// notifyUnreadChanged is exported so pages that mutate notifications
// (mark read, delete) can force an immediate refresh instead of waiting
// for the next 5s tick. Avoids the awkward "I just clicked Mark Read
// and the badge still shows for 4 seconds" UX.
export function notifyUnreadChanged() {
  refresh();
}
