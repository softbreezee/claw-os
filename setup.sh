#!/bin/bash
# =============================================================================
# FastClaw 环境检查脚本
# =============================================================================
# 检查开发/运行所需依赖是否就绪，输出缺失项及安装指引。
# 本脚本不会自动安装任何东西。
#
# 用法：
#   chmod +x setup.sh
#   ./setup.sh
# =============================================================================

set -e

# ─── 颜色输出 ─────────────────────────────────────────────────────────────────
if [ -t 1 ]; then
    RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
    BLUE='\033[0;34m'; BOLD='\033[1m'; NC='\033[0m'
else
    RED=''; GREEN=''; YELLOW=''; BLUE=''; BOLD=''; NC=''
fi

ok()    { printf "  ${GREEN}✅${NC} %s\n" "$*"; }
warn()  { printf "  ${YELLOW}⚠️${NC}  %s\n" "$*"; }
fail()  { printf "  ${RED}❌${NC} %s\n" "$*"; }
title() { printf "\n${BLUE}${BOLD}═══ %s ═══${NC}\n" "$*"; }

has() { command -v "$1" >/dev/null 2>&1; }

version_gte() {
    printf '%s\n%s\n' "$2" "$1" | sort -V -C 2>/dev/null
}

detect_os() {
    case "$(uname -s)" in
        Darwin) OS="macos" ;;
        Linux)  OS="linux" ;;
        *)      OS="unknown" ;;
    esac
}

# ─── 检查项 ───────────────────────────────────────────────────────────────────

check_go() {
    if has go; then
        local v; v=$(go version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
        if [ -n "$v" ] && version_gte "$v" "1.25.0"; then
            ok "Go $v"
        else
            warn "Go ${v:-unknown} — 需要 >= 1.25.0"
        fi
    else
        fail "Go 未安装"
        echo "    安装: https://go.dev/dl/  或  brew install go"
    fi
}

check_node() {
    if has node; then
        local v; v=$(node --version 2>/dev/null | sed 's/v//')
        if [ -n "$v" ] && version_gte "$v" "18.0.0"; then
            ok "Node.js $v"
        else
            warn "Node.js ${v:-unknown} — 建议 >= 18"
        fi
    else
        warn "Node.js 未安装 (仅构建 Web UI 需要)"
        echo "    安装: brew install node  或  https://nodejs.org/"
    fi
}

check_pnpm() {
    if has pnpm; then
        ok "pnpm $(pnpm --version 2>/dev/null)"
    else
        warn "pnpm 未安装 (仅构建 Web UI 需要)"
        echo "    安装: npm install -g pnpm  或  brew install pnpm"
    fi
}

check_psql() {
    if has psql; then
        local v; v=$(psql --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+' | head -1)
        ok "PostgreSQL client $v"
    else
        warn "psql 未安装"
    fi
}

check_pg_running() {
    if pg_isready -q 2>/dev/null; then
        ok "PostgreSQL 服务运行中"
    else
        warn "PostgreSQL 服务未运行或未安装"
        echo "    macOS:  brew install postgresql@17 && brew services start postgresql@17"
        echo "    Linux:  sudo apt install postgresql && sudo systemctl start postgresql"
    fi
}

check_pgvector() {
    if pg_isready -q 2>/dev/null && has psql; then
        if psql -U "$(whoami)" -d postgres -tAc \
            "SELECT 1 FROM pg_available_extensions WHERE name='vector'" 2>/dev/null | grep -q 1; then
            ok "pgvector 扩展可用"
        else
            warn "pgvector 扩展未安装"
            echo "    macOS:  brew install pgvector"
            echo "    Linux:  sudo apt install postgresql-17-pgvector"
        fi
    fi
}

check_config() {
    if [ -f "$HOME/.fastclaw/fastclaw.json" ]; then
        ok "配置文件 ~/.fastclaw/fastclaw.json"
    else
        warn "配置文件不存在，首次启动 fastclaw 时会进入配置向导"
    fi
}

check_make() {
    if has make; then
        ok "make"
    else
        warn "make 未安装 (构建项目需要)"
    fi
}

# ─── 主入口 ───────────────────────────────────────────────────────────────────

main() {
    detect_os

    printf "\n${BOLD}🔍 FastClaw 环境检查${NC}\n"
    printf "──────────────────────────────────────\n"

    title "运行时依赖"
    check_go
    check_psql
    check_pg_running
    check_pgvector
    check_config

    title "构建依赖 (仅开发/本地构建需要)"
    check_node
    check_pnpm
    check_make

    # ── 快速启动指引 ──
    echo ""
    echo "──────────────────────────────────────"
    echo ""
    echo "${BOLD}🚀 快速开始${NC}"
    echo ""
    echo "  1. 确保 Go 1.25+ 和 PostgreSQL 已就绪"
    echo "  2. 首次运行自动进入配置向导:"
    echo ""
    echo "       go run ./cmd/fastclaw"
    echo ""
    echo "  或构建后运行:"
    echo ""
    if [ "$OS" = "macos" ]; then
        echo "       brew install go node pnpm postgresql@17 pgvector"
        echo "       brew services start postgresql@17"
        echo ""
    fi
    echo "       make build         # 编译项目"
    echo "       ./bin/fastclaw     # 启动服务"
    echo ""
}

main "$@"