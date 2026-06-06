"use client";

import { useState, useCallback, useEffect } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Textarea } from "@/components/ui/textarea";
import { testProvider, saveConfig, waitForGateway } from "@/lib/api";
import { Logo } from "@/components/logo";

// Provider presets shown in the onboard wizard. Keep the model lists
// in sync with internal/modelcatalog/modelcatalog.go — those entries
// are the ones whose context windows / compaction thresholds Pawnix
// actually knows about. Anything outside this list still works, but
// won't get an accurate token budget out of the box.
//
// `apiType` decides the wire protocol used by the provider runtime
// (see internal/provider/*.go). `authType` is informational today;
// the runtime picks the auth header from `apiType`. We preserve it
// in case future drivers branch on it.
//
// For aggregators like OpenRouter the model id is already
// vendor-prefixed (e.g. "anthropic/claude-sonnet-4"); the backend's
// resolveModelID() helper is aware of this and won't double-prefix.
const PROVIDERS: Record<
  string,
  { label: string; apiBase: string; apiType: string; authType: string; models: string[] }
> = {
  deepseek: {
    label: "DeepSeek",
    apiBase: "https://api.deepseek.com/v1",
    apiType: "openai-chat",
    authType: "bearer-token",
    models: ["deepseek-chat", "deepseek-reasoner", "deepseek-v4-flash", "deepseek-v4-pro"],
  },
  openai: {
    label: "OpenAI",
    apiBase: "https://api.openai.com/v1",
    apiType: "openai-chat",
    authType: "bearer-token",
    models: ["gpt-5", "gpt-4o", "gpt-4o-mini", "o3", "o3-mini", "o4-mini"],
  },
  anthropic: {
    label: "Anthropic",
    apiBase: "https://api.anthropic.com",
    apiType: "anthropic-messages",
    authType: "api-key",
    models: ["claude-sonnet-4", "claude-opus-4", "claude-3-5-sonnet", "claude-3-5-haiku"],
  },
  openrouter: {
    label: "OpenRouter",
    apiBase: "https://openrouter.ai/api/v1",
    apiType: "openai-chat",
    authType: "bearer-token",
    // OpenRouter ids are already vendor/model-shaped — keep them as-is.
    models: [
      "anthropic/claude-sonnet-4",
      "openai/gpt-5",
      "google/gemini-2.5-pro",
      "deepseek/deepseek-chat",
    ],
  },
  google: {
    label: "Google Gemini",
    apiBase: "https://generativelanguage.googleapis.com/v1beta/openai",
    apiType: "openai-chat",
    authType: "bearer-token",
    models: ["gemini-2.5-pro", "gemini-2.5-flash", "gemini-2.0-pro", "gemini-2.0-flash"],
  },
  moonshot: {
    label: "Moonshot (Kimi)",
    apiBase: "https://api.moonshot.cn/v1",
    apiType: "openai-chat",
    authType: "bearer-token",
    models: ["kimi-k2.5", "kimi-k2"],
  },
  qwen: {
    label: "Qwen (DashScope)",
    apiBase: "https://dashscope.aliyuncs.com/compatible-mode/v1",
    apiType: "openai-chat",
    authType: "bearer-token",
    models: ["qwen-max", "qwen2.5-72b"],
  },
  ollama: {
    label: "Ollama (Local)",
    apiBase: "http://localhost:11434/v1",
    apiType: "openai-chat",
    authType: "bearer-token",
    models: ["llama3", "mistral", "codellama"],
  },
  custom: {
    label: "Custom",
    apiBase: "",
    apiType: "openai-chat",
    authType: "bearer-token",
    models: [],
  },
};

const API_TYPE_OPTIONS = [
  { value: "openai-chat", label: "OpenAI Completions" },
  { value: "anthropic-messages", label: "Anthropic Messages" },
];

const AUTH_TYPE_OPTIONS = [
  { value: "api-key", label: "API Key" },
  { value: "bearer-token", label: "Bearer Token" },
];

// Storage backend options exposed in onboard. SQLite is intentionally
// omitted to keep the wizard simple — power users can edit pawnix.json
// directly. See docs/onboarding-issues.md (Phase 1 / storage entrypoint).
const STORAGE_OPTIONS = [
  { value: "file", label: "File (default, no DB)" },
  { value: "postgres", label: "PostgreSQL" },
];

// Fields whose change should invalidate a prior "Connected" badge.
// Editing any of these means the previous handshake no longer reflects
// what we'd actually save.
const TEST_INVALIDATING_KEYS = new Set([
  "apiBase",
  "apiKey",
  "model",
  "apiType",
  "authType",
  "provider",
]);

interface OnboardConfig {
  provider: string;
  providerName: string;
  apiBase: string;
  apiKey: string;
  apiType: string;
  authType: string;
  model: string;
  telegramEnabled: boolean;
  telegramToken: string;
  port: number;
  agentName: string;
  personality: string;
  // Storage backend (added in Phase 1 of onboarding fix-up). Maps to
  // saveConfigRequest.StorageType / StorageDSN on the backend.
  storageType: string; // "file" | "postgres"
  storageDsn: string;  // only used when storageType === "postgres"
}

const STEP_LABELS = [
  "Welcome",
  "LLM Provider",
  "Gateway",
  "Launch",
];

function ConfettiEffect() {
  const colors = [
    "#8b5cf6",
    "#06b6d4",
    "#10b981",
    "#f59e0b",
    "#ef4444",
    "#ec4899",
    "#6366f1",
  ];
  const pieces = Array.from({ length: 50 }, (_, i) => ({
    id: i,
    left: Math.random() * 100,
    delay: Math.random() * 2,
    color: colors[i % colors.length],
    size: 6 + Math.random() * 8,
    rotation: Math.random() * 360,
  }));

  return (
    <div className="pointer-events-none fixed inset-0 z-50 overflow-hidden">
      {pieces.map((p) => (
        <div
          key={p.id}
          className="confetti-piece"
          style={{
            left: `${p.left}%`,
            animationDelay: `${p.delay}s`,
            backgroundColor: p.color,
            width: `${p.size}px`,
            height: `${p.size}px`,
            borderRadius: p.id % 3 === 0 ? "50%" : "2px",
            transform: `rotate(${p.rotation}deg)`,
          }}
        />
      ))}
    </div>
  );
}

export default function OnboardPage() {
  const [step, setStep] = useState(0);
  const [config, setConfig] = useState<OnboardConfig>(() => {
    // Default to DeepSeek — it's the cheapest mainstream provider with
    // a sensible free tier, and was the headline gap before this fix.
    const initial = PROVIDERS.deepseek;
    return {
      provider: "deepseek",
      providerName: "deepseek",
      apiBase: initial.apiBase,
      apiKey: "",
      apiType: initial.apiType,
      authType: initial.authType,
      model: initial.models[0] || "",
      telegramEnabled: false,
      telegramToken: "",
      port: 18953,
      agentName: "Pawnix",
      personality:
        "You are a helpful, friendly AI assistant. You respond concisely and accurately.",
      storageType: "file",
      storageDsn: "",
    };
  });
  const [testStatus, setTestStatus] = useState<
    "idle" | "testing" | "success" | "error"
  >("idle");
  const [testError, setTestError] = useState("");
  const [showConfetti, setShowConfetti] = useState(false);
  const [launched, setLaunched] = useState(false);
  const [launchError, setLaunchError] = useState("");
  const [launching, setLaunching] = useState(false);
  const [jsonExpanded, setJsonExpanded] = useState(false);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  // updateConfig also invalidates the test-connection state when the
  // user edits a field that affects the handshake. Without this, the
  // green "Connected" badge would linger after editing the API Key,
  // letting users proceed with an untested config (Bug-2).
  const updateConfig = useCallback(
    (updates: Partial<OnboardConfig>) => {
      setConfig((prev) => ({ ...prev, ...updates }));
      const invalidates = Object.keys(updates).some((k) =>
        TEST_INVALIDATING_KEYS.has(k),
      );
      if (invalidates) {
        setTestStatus("idle");
        setTestError("");
      }
    },
    [],
  );

  const handleProviderChange = useCallback(
    (provider: string | null) => {
      if (!provider) return;
      const preset = PROVIDERS[provider];
      updateConfig({
        provider,
        providerName: provider === "custom" ? "" : provider,
        apiBase: preset.apiBase,
        apiType: preset.apiType,
        authType: preset.authType,
        model: preset.models[0] || "",
      });
    },
    [updateConfig],
  );

  const handleTestConnection = useCallback(async () => {
    setTestStatus("testing");
    setTestError("");
    try {
      const result = await testProvider({
        apiBase: config.apiBase,
        apiKey: config.apiKey,
        model: config.model,
        apiType: config.apiType,
        authType: config.authType,
      });
      if (result.ok) {
        setTestStatus("success");
      } else {
        const urlInfo = result.url ? `\nRequest URL: ${result.url}` : "";
        setTestStatus("error");
        setTestError((result.error || "Connection failed") + urlInfo);
      }
    } catch {
      setTestStatus("error");
      setTestError("Could not reach the server. Is Pawnix running?");
    }
  }, [config.apiBase, config.apiKey, config.model, config.apiType, config.authType]);

  // handleLaunch sends the wizard payload to /api/save-config and
  // honours the backend's {ok, error} contract. Previously a failure
  // path swallowed the error and still threw confetti, which made
  // users think onboarding succeeded while pawnix.json was either
  // missing or invalid (Bug-7).
  //
  // After save succeeds we poll /api/status until `running === true`
  // before redirecting (Bug-8 round 2). Without this the user gets a
  // 3-second timer that may or may not coincide with the daemon
  // actually finishing startup; on a cold boot with a slow PG migrate
  // the redirect lands on a 404 or a half-loaded shell. The poll
  // window is bounded so a genuinely-broken daemon doesn't trap the
  // user on the launch screen forever — they see a clear error and
  // a "go to chat anyway" escape hatch instead.
  const handleLaunch = useCallback(async () => {
    setLaunching(true);
    setLaunchError("");
    try {
      const result = (await saveConfig(
        config as unknown as Record<string, unknown>,
      )) as { ok?: boolean; error?: string };
      if (!result || !result.ok) {
        setLaunchError(result?.error || "Failed to save configuration");
        setLaunching(false);
        return;
      }
      setShowConfetti(true);
      setLaunched(true);
      setTimeout(() => setShowConfetti(false), 4000);

      // Wait until the gateway daemon has finished restarting and
      // the agent registry is populated. Status polling is cheap and
      // resolves the race between save-config returning OK and the
      // browser redirect landing on a not-yet-mounted /chat route.
      const ready = await waitForGateway(20_000);
      const host = window.location.hostname || "localhost";
      const port = config.port || window.location.port;
      const proto = window.location.protocol || "http:";
      const target = `${proto}//${host}:${port}/chat/`;
      if (!ready) {
        setLaunchError(
          "Saved your config, but the daemon is taking longer than expected to start. " +
          "If /chat doesn't load, check the logs at ~/.pawnix/daemon.log.",
        );
      }
      // Redirect either way — even on timeout the daemon is usually
      // a few more seconds from being ready and the chat shell will
      // refresh itself once it loads. Better than trapping the user
      // on a stale onboarding screen.
      window.location.href = target;
    } catch (err) {
      setLaunchError(
        err instanceof Error
          ? err.message
          : "Could not reach the setup server. Is it still running?",
      );
      setLaunching(false);
    }
  }, [config]);

  const canProceed = useCallback(() => {
    switch (step) {
      case 0:
        return true;
      case 1:
        // Require a green test before allowing the user to advance —
        // the wizard previously let unverified configs through and the
        // first agent call would fail at runtime (Bug-1).
        return (
          config.apiBase.length > 0 &&
          config.model.length > 0 &&
          testStatus === "success"
        );
      case 2: {
        if (!(config.agentName.length > 0 && config.port > 0)) return false;
        // Postgres requires a DSN before we'll let the wizard advance.
        if (config.storageType === "postgres" && !config.storageDsn.trim()) {
          return false;
        }
        // If the Telegram channel is enabled, a token is mandatory —
        // saving with an empty token would silently disable Telegram
        // on the backend (Bug-10 follow-up).
        if (config.telegramEnabled && !config.telegramToken.trim()) {
          return false;
        }
        return true;
      }
      case 3:
        return true;
      default:
        return false;
    }
  }, [step, config, testStatus]);

  if (!mounted) return null;

  return (
    <div className="relative flex min-h-screen flex-col items-center justify-center bg-background px-4 py-12">
      {showConfetti && <ConfettiEffect />}

      <div className="pointer-events-none absolute inset-0 overflow-hidden">
        <div className="absolute -top-[40%] left-1/2 h-[800px] w-[800px] -translate-x-1/2 rounded-full bg-primary/5 blur-3xl" />
      </div>

      <div className="relative mb-10 flex items-center gap-2">
        {STEP_LABELS.map((label, i) => (
          <div key={label} className="flex items-center gap-2">
            <button
              onClick={() => i < step && setStep(i)}
              className={`flex h-9 w-9 items-center justify-center rounded-full text-sm font-medium transition-all duration-300 ${
                i === step
                  ? "bg-primary text-primary-foreground shadow-lg shadow-primary/25 scale-110"
                  : i < step
                    ? "bg-primary/20 text-primary hover:bg-primary/30 cursor-pointer"
                    : "bg-muted text-muted-foreground"
              }`}
              disabled={i > step}
            >
              {i < step ? (
                <svg
                  className="h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  strokeWidth={2.5}
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="M5 13l4 4L19 7"
                  />
                </svg>
              ) : (
                i + 1
              )}
            </button>
            {i < STEP_LABELS.length - 1 && (
              <div
                className={`hidden h-px w-8 sm:block ${
                  i < step ? "bg-primary/40" : "bg-border"
                }`}
              />
            )}
          </div>
        ))}
      </div>

      <p className="relative mb-6 text-sm font-medium tracking-wide text-muted-foreground uppercase">
        {STEP_LABELS[step]}
      </p>

      <div className="relative w-full max-w-lg animate-fade-in-up" key={step}>
        {step === 0 && (
          <Card className="backdrop-blur-sm">
            <CardHeader className="space-y-6 pb-4 text-center">
              <div className="mx-auto flex h-16 w-16 items-center justify-center">
                <Logo className="h-16 w-16 rounded-2xl" />
              </div>
              <div>
                <CardTitle className="text-3xl font-bold">
                  <span className="animate-gradient-text bg-gradient-to-r from-violet-400 via-cyan-400 to-violet-400 bg-clip-text text-transparent">
                    Pawnix
                  </span>
                </CardTitle>
              </div>
            </CardHeader>
            <CardContent className="space-y-6 text-center">
              <p className="text-sm leading-relaxed text-muted-foreground">
                Set up your AI agent in a few simple steps. Configure your LLM
                provider, connect messaging channels, and launch your agent.
              </p>
              <Button
                onClick={() => setStep(1)}
                className="w-full"
                size="lg"
              >
                Get Started
                <svg
                  className="ml-2 h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  strokeWidth={2}
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="M13 7l5 5m0 0l-5 5m5-5H6"
                  />
                </svg>
              </Button>
            </CardContent>
          </Card>
        )}

        {step === 1 && (
          <Card className="backdrop-blur-sm">
            <CardHeader>
              <CardTitle className="text-xl">LLM Provider</CardTitle>
              <CardDescription>
                Choose your AI model provider and configure the connection.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-5">
              <div className="space-y-2">
                <Label>Provider</Label>
                <Select
                  value={config.provider}
                  onValueChange={handleProviderChange}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {Object.entries(PROVIDERS).map(([key, p]) => (
                      <SelectItem key={key} value={key}>
                        {p.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {config.provider === "custom" && (
                <div className="space-y-2">
                  <Label>Provider Name</Label>
                  <Input
                    value={config.providerName}
                    onChange={(e) => updateConfig({ providerName: e.target.value })}
                    placeholder="e.g. my-provider"
                    className="font-mono text-sm"
                  />
                </div>
              )}

              <div className="space-y-2">
                <Label>API Base URL</Label>
                <Input
                  value={config.apiBase}
                  onChange={(e) => updateConfig({ apiBase: e.target.value })}
                  placeholder="https://api.openai.com/v1"
                  className="font-mono text-sm"
                />
              </div>

              <div className="space-y-2">
                <Label>API Key</Label>
                <Input
                  type="password"
                  value={config.apiKey}
                  onChange={(e) => updateConfig({ apiKey: e.target.value })}
                  placeholder={
                    config.provider === "ollama"
                      ? "Not required for Ollama"
                      : "sk-..."
                  }
                  className="font-mono text-sm"
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>API Type</Label>
                  <Select
                    value={config.apiType}
                    onValueChange={(v) => v && updateConfig({ apiType: v })}
                  >
                    <SelectTrigger className="w-full text-sm">
                      <SelectValue>
                        {API_TYPE_OPTIONS.find((o) => o.value === config.apiType)?.label}
                      </SelectValue>
                    </SelectTrigger>
                    <SelectContent>
                      {API_TYPE_OPTIONS.map((opt) => (
                        <SelectItem key={opt.value} value={opt.value}>
                          {opt.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label>Auth Type</Label>
                  <Select
                    value={config.authType}
                    onValueChange={(v) => v && updateConfig({ authType: v })}
                  >
                    <SelectTrigger className="w-full text-sm">
                      <SelectValue>
                        {AUTH_TYPE_OPTIONS.find((o) => o.value === config.authType)?.label}
                      </SelectValue>
                    </SelectTrigger>
                    <SelectContent>
                      {AUTH_TYPE_OPTIONS.map((opt) => (
                        <SelectItem key={opt.value} value={opt.value}>
                          {opt.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <div className="space-y-2">
                <Label>Model</Label>
                <Input
                  value={config.model}
                  onChange={(e) => updateConfig({ model: e.target.value })}
                  placeholder="Enter model name"
                  className="font-mono text-sm"
                />
                {PROVIDERS[config.provider]?.models.length > 0 && (
                  <div className="flex flex-wrap gap-1.5">
                    {PROVIDERS[config.provider].models.map((m) => (
                      <button
                        key={m}
                        type="button"
                        onClick={() => updateConfig({ model: m })}
                        className={`rounded-md border px-2 py-0.5 text-xs font-mono transition-colors ${
                          config.model === m
                            ? "border-primary bg-primary/10 text-primary"
                            : "border-border text-muted-foreground hover:border-primary/50 hover:text-foreground"
                        }`}
                      >
                        {m}
                      </button>
                    ))}
                  </div>
                )}
              </div>

              <Separator />

              <div className="space-y-2">
                <div className="flex items-center gap-3">
                  <Button
                    variant="outline"
                    onClick={handleTestConnection}
                    disabled={testStatus === "testing"}
                  >
                    {testStatus === "testing" ? (
                      <>
                        <svg
                          className="mr-2 h-4 w-4 animate-spin"
                          fill="none"
                          viewBox="0 0 24 24"
                        >
                          <circle
                            className="opacity-25"
                            cx="12"
                            cy="12"
                            r="10"
                            stroke="currentColor"
                            strokeWidth="4"
                          />
                          <path
                            className="opacity-75"
                            fill="currentColor"
                            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
                          />
                        </svg>
                        Testing...
                      </>
                    ) : (
                      "Test Connection"
                    )}
                  </Button>
                  {testStatus === "success" && (
                    <Badge
                      variant="outline"
                      className="bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20"
                    >
                      Connected
                    </Badge>
                  )}
                  {testStatus === "idle" && (
                    <span className="text-xs text-muted-foreground">
                      Required before continuing
                    </span>
                  )}
                </div>
                {testStatus === "error" && (
                  <p className="text-sm text-destructive whitespace-pre-wrap break-all">
                    {testError || "Connection failed"}
                  </p>
                )}
              </div>

              <NavigationButtons
                step={step}
                setStep={setStep}
                canProceed={canProceed()}
              />
            </CardContent>
          </Card>
        )}

        {step === 2 && (
          <Card className="backdrop-blur-sm">
            <CardHeader>
              <CardTitle className="text-xl">Gateway Settings</CardTitle>
              <CardDescription>
                Configure your agent identity, server port, and storage backend.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-5">
              <div className="space-y-2">
                <Label>Agent Name</Label>
                <Input
                  value={config.agentName}
                  onChange={(e) =>
                    updateConfig({ agentName: e.target.value })
                  }
                  placeholder="My AI Agent"
                />
              </div>

              <div className="space-y-2">
                <Label>Port</Label>
                <Input
                  type="number"
                  value={config.port}
                  onChange={(e) =>
                    updateConfig({ port: parseInt(e.target.value) || 18953 })
                  }
                  className="font-mono"
                />
              </div>

              <Separator />

              <div className="space-y-2">
                <Label>Storage Backend</Label>
                <Select
                  value={config.storageType}
                  onValueChange={(v) => v && updateConfig({ storageType: v })}
                >
                  <SelectTrigger className="w-full text-sm">
                    <SelectValue>
                      {STORAGE_OPTIONS.find((o) => o.value === config.storageType)?.label}
                    </SelectValue>
                  </SelectTrigger>
                  <SelectContent>
                    {STORAGE_OPTIONS.map((opt) => (
                      <SelectItem key={opt.value} value={opt.value}>
                        {opt.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <p className="text-xs text-muted-foreground">
                  {config.storageType === "postgres"
                    ? "Sessions, memory, and agent state will live in PostgreSQL. Filesystem stays as cold storage."
                    : "Sessions, memory, and agent state are kept under ~/.pawnix as JSON files."}
                </p>
              </div>

              {config.storageType === "postgres" && (
                <div className="space-y-2">
                  <Label>PostgreSQL DSN</Label>
                  <Input
                    value={config.storageDsn}
                    onChange={(e) => updateConfig({ storageDsn: e.target.value })}
                    placeholder="postgres://user:pass@localhost:5432/pawnix?sslmode=disable"
                    className="font-mono text-sm"
                  />
                  <p className="text-xs text-muted-foreground">
                    Pawnix will auto-migrate the schema on first launch.
                  </p>
                </div>
              )}

              <Separator />

              {/*
                Telegram channel — onboard-time wiring of a Telegram bot.
                The OnboardConfig already carries the fields and the
                backend handler (saveConfigRequest.TelegramEnabled /
                TelegramToken) acts on them; we just never rendered a UI
                (Bug-10). Keeping it optional and behind a checkbox so
                Web-only users aren't forced through the bot-token flow.
              */}
              <div className="space-y-2">
                <Label className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    className="h-4 w-4 cursor-pointer"
                    checked={config.telegramEnabled}
                    onChange={(e) =>
                      updateConfig({ telegramEnabled: e.target.checked })
                    }
                  />
                  Telegram bot{" "}
                  <span className="text-xs text-muted-foreground">
                    (optional)
                  </span>
                </Label>
                {config.telegramEnabled && (
                  <>
                    <Input
                      type="password"
                      value={config.telegramToken}
                      onChange={(e) =>
                        updateConfig({ telegramToken: e.target.value })
                      }
                      placeholder="Bot token from @BotFather"
                      className="font-mono text-sm"
                    />
                    <p className="text-xs text-muted-foreground">
                      Pawnix will validate the token with Telegram&apos;s{" "}
                      <code>getMe</code> API on save. You can change or
                      remove this later under Channels.
                    </p>
                  </>
                )}
              </div>

              <Separator />

              <div className="space-y-2">
                <Label>
                  Personality{" "}
                  <span className="text-xs text-muted-foreground">(SOUL.md)</span>
                </Label>
                <Textarea
                  value={config.personality}
                  onChange={(e) =>
                    updateConfig({ personality: e.target.value })
                  }
                  rows={5}
                  placeholder="Describe your agent's personality, tone, and behavior..."
                  className="text-sm resize-none"
                />
                <p className="text-xs text-muted-foreground">
                  This defines how your agent communicates and behaves.
                </p>
              </div>

              <NavigationButtons
                step={step}
                setStep={setStep}
                canProceed={canProceed()}
              />
            </CardContent>
          </Card>
        )}

        {step === 3 && (
          <Card className="backdrop-blur-sm animate-pulse-glow">
            <CardHeader>
              <CardTitle className="text-xl">
                {launched ? "You're All Set!" : "Review & Launch"}
              </CardTitle>
              <CardDescription>
                {launched
                  ? "Pawnix is now configured and ready to go."
                  : "Review your configuration before launching."}
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-5">
              {launched ? (
                <div className="space-y-4 text-center py-4">
                  <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-emerald-500/20">
                    <svg
                      className="h-8 w-8 text-emerald-500"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      strokeWidth={2}
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        d="M5 13l4 4L19 7"
                      />
                    </svg>
                  </div>
                  <p className="text-lg font-medium">
                    Configuration saved successfully
                  </p>
                  <p className="text-sm text-muted-foreground">
                    Redirecting to dashboard...
                  </p>
                </div>
              ) : (
                <>
                  <div className="space-y-3">
                    <SummaryRow
                      label="Provider"
                      value={PROVIDERS[config.provider]?.label || config.provider}
                    />
                    <SummaryRow label="Model" value={config.model} mono />
                    <SummaryRow label="API Base" value={config.apiBase} mono />
                    <SummaryRow
                      label="API Key"
                      value={config.apiKey ? "********" : "Not set"}
                    />
                    <Separator />
                    <SummaryRow label="Agent Name" value={config.agentName} />
                    <SummaryRow
                      label="Port"
                      value={String(config.port)}
                      mono
                    />
                    <SummaryRow
                      label="Storage"
                      value={
                        config.storageType === "postgres"
                          ? "PostgreSQL"
                          : "File (~/.pawnix)"
                      }
                    />
                    {config.storageType === "postgres" && (
                      <SummaryRow
                        label="DSN"
                        value={config.storageDsn || "Not set"}
                        mono
                      />
                    )}
                  </div>

                  <button
                    onClick={() => setJsonExpanded(!jsonExpanded)}
                    className="flex w-full items-center justify-between rounded-lg border border-border bg-muted/30 px-4 py-2.5 text-sm text-muted-foreground transition-colors hover:bg-muted/50"
                  >
                    <span>JSON Preview</span>
                    <svg
                      className={`h-4 w-4 transition-transform ${jsonExpanded ? "rotate-180" : ""}`}
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      strokeWidth={2}
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        d="M19 9l-7 7-7-7"
                      />
                    </svg>
                  </button>
                  {jsonExpanded && (
                    <pre className="max-h-64 overflow-auto rounded-lg border border-border bg-background p-4 font-mono text-xs text-muted-foreground">
                      {JSON.stringify(config, null, 2)}
                    </pre>
                  )}

                  {launchError && (
                    <div className="rounded-lg border border-destructive/40 bg-destructive/10 px-4 py-3 text-sm text-destructive whitespace-pre-wrap break-all">
                      {launchError}
                    </div>
                  )}

                  <Button
                    onClick={handleLaunch}
                    disabled={launching}
                    className="w-full bg-gradient-to-r from-violet-600 to-cyan-600 text-white hover:from-violet-700 hover:to-cyan-700 transition-all disabled:opacity-60"
                    size="lg"
                  >
                    {launching ? "Launching..." : "Launch Pawnix"}
                    <svg
                      className="ml-2 h-4 w-4"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      strokeWidth={2}
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        d="M15.59 14.37a6 6 0 01-5.84 7.38v-4.8m5.84-2.58a14.98 14.98 0 006.16-12.12A14.98 14.98 0 009.631 8.41m5.96 5.96a14.926 14.926 0 01-5.841 2.58m-.119-8.54a6 6 0 00-7.381 5.84h4.8m2.58-5.84a14.927 14.927 0 00-2.58 5.84m2.699 2.7c-.103.021-.207.041-.311.06a15.09 15.09 0 01-2.448-2.448 14.9 14.9 0 01.06-.312m-2.24 2.39a4.493 4.493 0 00-1.757 4.306 4.493 4.493 0 004.306-1.758M16.5 9a1.5 1.5 0 11-3 0 1.5 1.5 0 013 0z"
                      />
                    </svg>
                  </Button>

                  <NavigationButtons
                    step={step}
                    setStep={setStep}
                    canProceed={canProceed()}
                    hidNext
                  />
                </>
              )}
            </CardContent>
          </Card>
        )}
      </div>

      <p className="relative mt-8 text-xs text-muted-foreground/50">
        Pawnix Agent Framework
      </p>
    </div>
  );
}

function SummaryRow({
  label,
  value,
  mono,
  capitalize,
}: {
  label: string;
  value: string;
  mono?: boolean;
  capitalize?: boolean;
}) {
  return (
    <div className="flex items-center justify-between gap-4">
      <span className="text-sm text-muted-foreground shrink-0">{label}</span>
      <span
        className={`text-sm text-right break-all ${mono ? "font-mono" : ""} ${capitalize ? "capitalize" : ""}`}
      >
        {value}
      </span>
    </div>
  );
}

function NavigationButtons({
  step,
  setStep,
  canProceed,
  hidNext,
}: {
  step: number;
  setStep: (s: number) => void;
  canProceed: boolean;
  hidNext?: boolean;
}) {
  return (
    <div className="flex items-center justify-between pt-2">
      <Button
        variant="ghost"
        onClick={() => setStep(step - 1)}
        disabled={step === 0}
      >
        <svg
          className="mr-1 h-4 w-4"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={2}
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M11 17l-5-5m0 0l5-5m-5 5h12"
          />
        </svg>
        Back
      </Button>
      {!hidNext && (
        <Button
          onClick={() => setStep(step + 1)}
          disabled={!canProceed}
        >
          Next
          <svg
            className="ml-1 h-4 w-4"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            strokeWidth={2}
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M13 7l5 5m0 0l-5 5m5-5H6"
            />
          </svg>
        </Button>
      )}
    </div>
  );
}
