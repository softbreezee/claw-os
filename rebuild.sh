#!/bin/bash
# 快速重构建 + 重启（开发用）
# 用法: ./rebuild.sh          → 仅 rebuild Go（前端没改）
#       ./rebuild.sh --web    → rebuild 前端 + Go

set -e
cd "$(dirname "${BASH_SOURCE[0]}")"

if [ "${1:-}" = "--web" ]; then
    echo "🔨 Building frontend..."
    cd web && pnpm build && cd ..
    rm -rf internal/setup/web
    cp -r web/out internal/setup/web
fi

echo "🔨 Building Go binary..."
go build -o bin/pawnix ./cmd/pawnix

echo "🔄 Restarting..."
./pawnix-manager.sh restart

echo "✅ Done"
