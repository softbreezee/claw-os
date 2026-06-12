---
name: browser
description: |
  Anti-detect browser automation for sites with bot detection, login walls, or JS-rendered content — 小红书 / 抖音 / X(Twitter)/ 东方财富 / 雪球 / 同花顺 and similar. Trigger whenever a task needs to: open a real web page that web_fetch can't render, read a 博主/KOL 主页 or feed, search a platform for recent posts, fill forms, click through pages, or extract data from a site that blocks plain HTTP scraping. The browser runs in a separate Podman container (claw-browser); you drive it via the exec tool.
---

# Browser Automation (camoufox-cli in claw-browser container)

claw-os runs on the physical host; the browser runs in a long-lived
Docker container named **claw-browser**. You drive it through the
`exec` tool by prefixing every command with `podman exec claw-browser`.

It's built on **Camoufox** (anti-detect Firefox): `navigator.webdriver=false`,
randomized canvas/WebGL/audio fingerprints, real Firefox UA — passes bot
detection on sites that block Chromium automation. Use it whenever
`web_fetch` returns an empty shell, gets blocked, or the page needs JS to
render its content.

> **Read the full command reference** before anything beyond a basic
> open+snapshot: use the Read tool on
> `{Base directory}/references/commands.md` (replace `{Base directory}`
> with the path shown when this skill loads).

## CRITICAL: first call is slow — set a long timeout

The first `camoufox-cli open` cold-starts the headless Firefox daemon:
**2-3 minutes**. The exec tool's default timeout is 30s and its hard
ceiling is 300s. On your FIRST browser command in a session you MUST
pass `timeout: 280` or the daemon gets killed mid-startup and every
later call fails. Subsequent calls in the same session are sub-second
(default timeout is fine).

```jsonc
// FIRST call — long timeout, chain open + snapshot
{"command": "podman exec claw-browser camoufox-cli open https://example.com && podman exec claw-browser camoufox-cli snapshot -i", "timeout": 280}

// LATER calls — default timeout is plenty
{"command": "podman exec claw-browser camoufox-cli click @e5 && podman exec claw-browser camoufox-cli snapshot -i"}
```

## Core workflow

1. **Open**: `podman exec claw-browser camoufox-cli open <url>`
2. **Snapshot**: `podman exec claw-browser camoufox-cli snapshot -i`
   → returns interactive elements with refs like `[ref=e1]`
3. **Interact** via refs: `click @e1` / `fill @e1 "text"` / `select @e1 "opt"`
4. **Re-snapshot** after ANY navigation or DOM change — refs go stale
5. **Close** when the whole task is done: `camoufox-cli close`
   (keep it open if the user may have follow-ups)

```bash
podman exec claw-browser camoufox-cli open https://www.xiaohongshu.com/user/profile/SOMEID
podman exec claw-browser camoufox-cli snapshot -i
# - link "笔记标题一" [ref=e3]
# - link "笔记标题二" [ref=e4]
podman exec claw-browser camoufox-cli text @e3   # 读某条标题/正文
```

## Ref lifecycle (most common source of errors)

Refs (`@e1`, `@e2`, …) are temporary — assigned per snapshot, invalidated
by navigation / form submit / dynamic content / scroll-triggered loading.
**Always re-snapshot after the page changes** before using a ref. Using a
stale ref clicks the wrong element or errors out.

## JS-heavy sites (小红书 / 抖音 / SPA)

These render content via JavaScript. `text body` returns the JS bundle,
not the visible content. **Use `snapshot -i`** — the aria tree exposes the
rendered, accessible labels (笔记标题、作者、互动数). If the snapshot looks
unfinished, `wait` for an element then re-snapshot:

```bash
podman exec claw-browser camoufox-cli wait 3000
podman exec claw-browser camoufox-cli snapshot -i
```

## Login-walled content (小红书 关注流 / 私域)

Public 主页 / 搜索结果 usually work without login. If a page needs login,
DO NOT enter the user's password. Instead use a persistent profile the
user has logged into ONCE by hand:

```bash
# 用户手动在 headed 模式登录一次,fingerprint+cookie 存进 profile
podman exec claw-browser camoufox-cli --persistent /profiles/xhs open https://www.xiaohongshu.com
# (用户手动扫码/登录)
# 之后 agent 复用这个已登录 profile,无需密码
podman exec claw-browser camoufox-cli --persistent /profiles/xhs open <url> && \
podman exec claw-browser camoufox-cli snapshot -i
```

Never automate the login step itself. The human logs in; you reuse the
session. (See references/commands.md → Persistent Identity.)

## Common gotchas

- **"command not found: podman" / 容器连不上** — the claw-browser
  container isn't set up. Tell the user to run
  `deploy/browser/claw-browser.sh start`.
- **"Ref @eN not found"** — stale ref, re-snapshot.
- **Empty/garbage snapshot** — page not loaded, `wait 3000` then re-snapshot.
- **Too many elements** — scope: `snapshot -i -s "#main"`.
- **Cold-start timeout on first call** — you forgot `timeout: 280`.

## Cleanup

Close sessions when done so daemons don't leak:
```bash
podman exec claw-browser camoufox-cli close          # default session
podman exec claw-browser camoufox-cli close --all    # all sessions
```

The container itself stays up (managed by claw-browser.sh) — only the
browser session/daemon inside it is closed.
