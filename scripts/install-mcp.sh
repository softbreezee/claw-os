#!/usr/bin/env bash
# install-mcp.sh — 一键把 `pawnix mcp`（跨工具共享记忆后端）接进你本机的 AI 工具。
#
# 覆盖 4 个宿主：
#   claude    → Claude Code   （claude CLI，--source claude-code）
#   chatgpt   → Codex CLI      （~/.codex/config.toml，--source codex）
#   codewiz   → codex 类 CLI    （~/.codewiz/config.toml，--source codewiz）
#   hermes    → 你自己的 agent  （打印接入指引，--source hermes）
#
# 用法:
#   ./scripts/install-mcp.sh                 # 自动探测并接入
#   PAWNIX_BIN=/abs/path/pawnix ./scripts/install-mcp.sh
#   CODEX_CONFIG=~/.codex/config.toml CODEWIZ_CONFIG=~/.codewiz/config.toml ./scripts/install-mcp.sh
#
# 特性：幂等（重复跑安全，已存在的配置会跳过而非重复写入）、只读探测、不碰密码。
# 前提：先构建出 bin/pawnix（make build / make dev-build），且 ~/.pawnix/config 已配好
#       storage.type=postgres + 有效 DSN（记忆后端依赖 pgvector）。

set -euo pipefail

# ── 颜色（仅在 TTY 上着色）─────────────────────────────────────────────
if [ -t 1 ]; then
  BOLD=$'\033[1m'; DIM=$'\033[2m'; GREEN=$'\033[32m'; YELLOW=$'\033[33m'
  BLUE=$'\033[34m'; RED=$'\033[31m'; RESET=$'\033[0m'
else
  BOLD=""; DIM=""; GREEN=""; YELLOW=""; BLUE=""; RED=""; RESET=""
fi
say()  { printf '%s\n' "$*"; }
ok()   { printf '  %s✓%s %s\n' "$GREEN" "$RESET" "$*"; }
skip() { printf '  %s•%s %s\n' "$DIM" "$RESET" "$*"; }
warn() { printf '  %s!%s %s\n' "$YELLOW" "$RESET" "$*"; }
head() { printf '\n%s%s%s\n' "$BOLD" "$*" "$RESET"; }

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# ── 1. 解析 pawnix 可执行文件（一律用绝对路径）───────────────────────────
# 顺序：PAWNIX_BIN 环境变量 → PATH 上的 pawnix → 仓库内 bin/pawnix。
# 用绝对路径的原因：GUI/后台起的工具不继承你的 shell PATH，写相对路径会找不到。
abspath() {
  local dir base
  dir="$(cd "$(dirname "$1")" && pwd)" || return 1
  base="$(basename "$1")"
  printf '%s/%s\n' "$dir" "$base"
}

resolve_pawnix() {
  if [ -n "${PAWNIX_BIN:-}" ]; then
    printf '%s\n' "$PAWNIX_BIN"; return
  fi
  if command -v pawnix >/dev/null 2>&1; then
    command -v pawnix; return
  fi
  printf '%s\n' "$REPO_ROOT/bin/pawnix"
}

RAW_PAWNIX="$(resolve_pawnix)"
if [ ! -e "$RAW_PAWNIX" ]; then
  say "${RED}找不到 pawnix 可执行文件：${RESET}$RAW_PAWNIX"
  say "先构建：${BOLD}make build${RESET}（或 make dev-build），或设 PAWNIX_BIN 指向它。"
  exit 1
fi
PAWNIX="$(abspath "$RAW_PAWNIX")"
if [ ! -x "$PAWNIX" ]; then
  say "${RED}pawnix 不可执行：${RESET}$PAWNIX"
  say "试试：${BOLD}chmod +x \"$PAWNIX\"${RESET}"
  exit 1
fi

# 冒烟验证：cobra 的 --help 正常退出即说明二进制能跑、mcp 子命令存在。
if ! "$PAWNIX" mcp --help >/dev/null 2>&1; then
  say "${RED}\`$PAWNIX mcp --help\` 执行失败${RESET} —— 二进制可能损坏或架构不匹配（需与本机一致）。"
  exit 1
fi

head "Pawnix Memory MCP 接入"
say "使用二进制：${BLUE}$PAWNIX${RESET}"
say "所有工具将连到同一个共享记忆池（Postgres），靠 ${BOLD}source:<tool>${RESET} 标签区分来源。"

# ── 通用：把 [mcp_servers.memory] 幂等地并进一个 Codex 风格的 TOML ────────
ensure_codex_toml() {
  local cfg="$1" tag="$2" label="$3"
  mkdir -p "$(dirname "$cfg")"
  if [ -f "$cfg" ] && grep -qE '^\[mcp_servers\.memory\]' "$cfg" 2>/dev/null; then
    skip "$label：$cfg 里已存在 [mcp_servers.memory]，跳过（如需改动请手动编辑）"
    return 0
  fi
  # 追加前确保文件以换行结尾，避免和已有末行黏在一起。
  if [ -f "$cfg" ] && [ -s "$cfg" ] && [ -n "$(tail -c1 "$cfg" 2>/dev/null)" ]; then
    printf '\n' >> "$cfg"
  fi
  cat >> "$cfg" <<EOF
[mcp_servers.memory]
command = "$PAWNIX"
args = ["mcp", "--source", "$tag"]
EOF
  ok "$label：已写入 $cfg"
}

# ── 2. Claude Code ───────────────────────────────────────────────────
head "1) Claude Code"
if command -v claude >/dev/null 2>&1; then
  # 先删后加 → 幂等；-s user 装到用户级，所有项目都能用。
  claude mcp remove memory -s user >/dev/null 2>&1 || true
  if claude mcp add memory -s user -- "$PAWNIX" mcp --source claude-code >/dev/null 2>&1; then
    ok "已通过 \`claude mcp add\` 注册（user 级，source=claude-code）"
  else
    warn "\`claude mcp add\` 失败，手动加到 ~/.claude.json 的 mcpServers："
    say  "    \"memory\": { \"command\": \"$PAWNIX\", \"args\": [\"mcp\", \"--source\", \"claude-code\"] }"
  fi
else
  warn "未找到 claude CLI。装了 Claude Code 后再跑一次，或手动加到 ~/.claude.json 的 mcpServers："
  say  "    \"memory\": { \"command\": \"$PAWNIX\", \"args\": [\"mcp\", \"--source\", \"claude-code\"] }"
fi

# ── 3. ChatGPT = Codex CLI ───────────────────────────────────────────
head "2) ChatGPT（Codex CLI）"
CODEX_CONFIG="${CODEX_CONFIG:-$HOME/.codex/config.toml}"
ensure_codex_toml "$CODEX_CONFIG" "codex" "Codex CLI"
say "  ${DIM}注：ChatGPT 桌面/网页版连接器只支持远程 HTTP/SSE，接不了 stdio；这里对接的是 Codex CLI。${RESET}"

# ── 4. codewiz ───────────────────────────────────────────────────────
head "3) codewiz"
CODEWIZ_CONFIG="${CODEWIZ_CONFIG:-$HOME/.codewiz/config.toml}"
if command -v codewiz >/dev/null 2>&1 || [ -d "$(dirname "$CODEWIZ_CONFIG")" ]; then
  ensure_codex_toml "$CODEWIZ_CONFIG" "codewiz" "codewiz"
  say "  ${DIM}若 codewiz 用 JSON 而非 TOML 配置，改用与 Claude Code 相同的 mcpServers 结构、--source codewiz。${RESET}"
else
  warn "未探测到 codewiz（无 codewiz 命令、也没有 $(dirname "$CODEWIZ_CONFIG") 目录）。如已安装，手动加到它的 config.toml："
  say  "    [mcp_servers.memory]"
  say  "    command = \"$PAWNIX\""
  say  "    args = [\"mcp\", \"--source\", \"codewiz\"]"
  say  "  ${DIM}或设 CODEWIZ_CONFIG=/你的/config.toml 后重跑本脚本自动写入。${RESET}"
fi

# ── 5. hermes（你自己的 agent）──────────────────────────────────────
head "4) hermes（你自己的 agent）"
say "hermes 的配置路径因实现而异，脚本不擅自改动。把下面这段加到它的 MCP 配置即可："
say "    ${BOLD}JSON（mcpServers 结构）：${RESET}"
say "    \"memory\": { \"type\": \"stdio\", \"command\": \"$PAWNIX\", \"args\": [\"mcp\", \"--source\", \"hermes\"] }"
say "    ${BOLD}或直接起子进程：${RESET}"
say "    $PAWNIX mcp --source hermes"

# ── 收尾 ─────────────────────────────────────────────────────────────
head "完成"
say "接下来："
say "  1. ${BOLD}重启${RESET}上面这些工具，让它们重新 spawn \`pawnix mcp\` 子进程。"
say "  2. 冒烟自测读写回路：${BOLD}$REPO_ROOT/scripts/mcp-smoke-test.sh${RESET}"
say "  3. 打开使用看板：${BLUE}http://localhost:18953/memory/${RESET}"
say "     （若看板 404，说明前端还没 embed 新页面：在仓库根跑 ${BOLD}make build${RESET} 后重启 daemon）"
say ""
say "${DIM}提示：读积极、写保守——每轮开对话先 memory_search，只有跨会话的重要信息才 memory_write。${RESET}"
