"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  LayoutDashboard,
  MessageSquare,
  Bot,
  Sparkles,
  Puzzle,
  Cable,
  Radio,
  Clock,
  Settings,
  Brain,
  Menu,
  X,
  Sun,
  Moon,
  Circle,
  AppWindow,
  Inbox,
} from "lucide-react";
import { useState, useEffect } from "react";
import { useTheme } from "@/components/theme-provider";
import { Logo } from "@/components/logo";
import { useUnreadNotifications } from "@/lib/use-notifications";

const navGroups = [
  {
    label: "Core",
    items: [
      { href: "/overview/", label: "Overview", icon: LayoutDashboard },
      { href: "/chat/", label: "Chat", icon: MessageSquare },
      // Inbox carries an unread badge populated by useUnreadNotifications.
      // Placed in Core so users see it on every page without scrolling.
      { href: "/inbox/", label: "Inbox", icon: Inbox, badge: "inbox" as const },
    ],
  },
  {
    label: "Configure",
    items: [
      { href: "/agents/", label: "Agents", icon: Bot },
      { href: "/models/", label: "Models", icon: Brain },
      { href: "/skills/", label: "Skills", icon: Sparkles },
      { href: "/plugins/", label: "Plugins", icon: Puzzle },
      { href: "/mcp/", label: "MCP Servers", icon: Cable },
      { href: "/channels/", label: "Channels", icon: Radio },
      { href: "/cron/", label: "Cron Jobs", icon: Clock },
      { href: "/apps/", label: "Apps", icon: AppWindow },
    ],
  },
  {
    label: "System",
    items: [
      { href: "/settings/", label: "Settings", icon: Settings },
    ],
  },
];

function NavLinks({ onClick, pathname }: { onClick?: () => void; pathname: string }) {
  // Single subscription per sidebar instance keeps the polling cost
  // bounded regardless of how many nav items eventually want a badge.
  const unreadCount = useUnreadNotifications();

  return (
    <div className="space-y-4">
      {navGroups.map((group) => (
        <div key={group.label}>
          <p className="px-3 pb-1 text-[10px] font-semibold uppercase tracking-widest text-muted-foreground/50">
            {group.label}
          </p>
          <div className="space-y-0.5">
            {group.items.map((item) => {
              const isActive = pathname === item.href || pathname.startsWith(item.href);
              const badgeCount = item.badge === "inbox" ? unreadCount : 0;
              return (
                <Link
                  key={item.href}
                  href={item.href}
                  onClick={onClick}
                  className={`flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
                    isActive
                      ? "bg-primary/10 text-primary"
                      : "text-muted-foreground hover:bg-muted/50 hover:text-foreground"
                  }`}
                >
                  <item.icon className={`h-4 w-4 ${isActive ? "text-primary" : ""}`} />
                  <span className="flex-1">{item.label}</span>
                  {badgeCount > 0 && (
                    <span className="inline-flex h-5 min-w-[20px] items-center justify-center rounded-full bg-primary px-1.5 text-[10px] font-semibold text-primary-foreground">
                      {badgeCount > 99 ? "99+" : badgeCount}
                    </span>
                  )}
                  {badgeCount === 0 && isActive && (
                    <span className="h-1.5 w-1.5 rounded-full bg-primary" />
                  )}
                </Link>
              );
            })}
          </div>
        </div>
      ))}
    </div>
  );
}

export function SidebarLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const [mobileOpen, setMobileOpen] = useState(false);
  const { theme, toggleTheme } = useTheme();
  const [gatewayRunning, setGatewayRunning] = useState(false);
  const [version, setVersion] = useState("—");

  useEffect(() => {
    const check = () => {
      fetch("/api/status")
        .then((r) => r.json())
        .then((s) => {
          setGatewayRunning(s.running);
          if (s.version) setVersion(s.version);
        })
        .catch(() => setGatewayRunning(false));
    };
    check();
    const iv = setInterval(check, 15000);
    return () => clearInterval(iv);
  }, []);

  // Close mobile menu on route change
  useEffect(() => {
    setMobileOpen(false);
  }, [pathname]);

  return (
    <div className="flex min-h-screen bg-background">
      {/* ── Desktop sidebar ── */}
      <aside className="hidden w-56 flex-col border-r border-border bg-card/30 md:flex shrink-0">
        {/* Logo / status */}
        <div className="flex h-14 items-center gap-3 border-b border-border px-4">
          <Logo className="h-8 w-8 shrink-0 rounded-xl" />
          <div className="flex-1 min-w-0">
            <p className="text-sm font-semibold text-foreground leading-tight">Pawnix</p>
            <div className="flex items-center gap-1.5 mt-0.5">
              <span className={`h-1.5 w-1.5 rounded-full ${gatewayRunning ? "bg-emerald-500" : "bg-muted-foreground/40"}`} />
              <p className="text-[10px] text-muted-foreground leading-none">
                {gatewayRunning ? "Gateway running" : "Gateway stopped"}
              </p>
            </div>
          </div>
        </div>

        {/* Nav */}
        <nav className="flex-1 overflow-y-auto p-3 pt-4">
          <NavLinks pathname={pathname} />
        </nav>

        {/* Footer */}
        <div className="border-t border-border p-3">
          <div className="flex items-center justify-between">
            <span className="text-[10px] text-muted-foreground/40 font-mono">v{version}</span>
            <button
              onClick={toggleTheme}
              className="rounded-md p-1.5 text-muted-foreground hover:text-foreground hover:bg-muted/50 transition-colors"
              title={theme === "dark" ? "Switch to light mode" : "Switch to dark mode"}
            >
              {theme === "dark" ? <Sun className="h-3.5 w-3.5" /> : <Moon className="h-3.5 w-3.5" />}
            </button>
          </div>
        </div>
      </aside>

      {/* ── Mobile header ── */}
      <div className="fixed top-0 left-0 right-0 z-40 flex h-12 items-center justify-between border-b border-border bg-card/95 px-4 backdrop-blur-md md:hidden">
        <div className="flex items-center gap-2">
          <Logo className="h-7 w-7 rounded-lg" />
          <span className="text-sm font-semibold text-foreground">Pawnix</span>
          <span className={`h-1.5 w-1.5 rounded-full ${gatewayRunning ? "bg-emerald-500" : "bg-muted-foreground/40"}`} />
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={toggleTheme}
            className="rounded-md p-2 text-muted-foreground hover:text-foreground hover:bg-muted/50 transition-colors"
          >
            {theme === "dark" ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
          </button>
          <button
            onClick={() => setMobileOpen(!mobileOpen)}
            className="rounded-md p-2 text-muted-foreground hover:text-foreground hover:bg-muted/50 transition-colors"
          >
            {mobileOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
          </button>
        </div>
      </div>

      {/* ── Mobile drawer ── */}
      {mobileOpen && (
        <>
          <div
            className="fixed inset-0 z-30 bg-background/60 backdrop-blur-sm md:hidden"
            onClick={() => setMobileOpen(false)}
          />
          <div className="fixed top-12 left-0 bottom-0 z-40 w-64 flex flex-col border-r border-border bg-card shadow-xl md:hidden">
            <nav className="flex-1 overflow-y-auto p-3 pt-4">
              <NavLinks pathname={pathname} onClick={() => setMobileOpen(false)} />
            </nav>
            <div className="border-t border-border p-3">
              <div className="flex items-center gap-1.5">
                <Circle className={`h-1.5 w-1.5 fill-current ${gatewayRunning ? "text-emerald-500" : "text-muted-foreground/40"}`} />
                <span className="text-xs text-muted-foreground">
                  {gatewayRunning ? "Gateway running" : "Gateway stopped"}
                </span>
              </div>
            </div>
          </div>
        </>
      )}

      {/* ── Main content ── */}
      <main className="flex-1 pt-12 md:pt-0 overflow-y-auto min-w-0">{children}</main>
    </div>
  );
}
