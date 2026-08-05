#!/usr/bin/env bash
# mcp-smoke-test.sh — 手测 `pawnix mcp` 的 JSON-RPC 回路（读 + 写）
#
# 用法:
#   make dev-build            # 先构建出 bin/pawnix
#   ./scripts/mcp-smoke-test.sh
#
# 前提: ~/.pawnix/config 已配 storage.type=postgres + 有效 DSN（记忆后端依赖 pgvector）。
# 说明: MCP server 逐行读 stdin、每行回一条 JSON-RPC 到 stdout；stderr 只放日志。
#       stdin 关闭后进程自动退出，所以下面把所有请求一次性喂进去即可。
#       这里带 --source smoke-test，写入的记忆会带上 source:smoke-test 标签。

set -euo pipefail

BIN="${BIN:-./bin/pawnix}"

if [ ! -x "$BIN" ]; then
    echo "找不到可执行文件: ${BIN} (先跑 make dev-build 或 make build)" >&2
    exit 1
fi

echo "== 发送 initialize / tools/list / memory_stats / memory_write / memory_search =="

# id=4 写入一条带唯一 nonce 的记忆；id=5 立刻按 nonce 检索，验证写-读回路打通。
NONCE="smoke-$(date +%s)"

REQUESTS=$(printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' \
  '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"memory_stats","arguments":{}}}' \
  "{\"jsonrpc\":\"2.0\",\"id\":4,\"method\":\"tools/call\",\"params\":{\"name\":\"memory_write\",\"arguments\":{\"content\":\"冒烟测试记忆 $NONCE\",\"kind\":\"fact\"}}}" \
  "{\"jsonrpc\":\"2.0\",\"id\":5,\"method\":\"tools/call\",\"params\":{\"name\":\"memory_search\",\"arguments\":{\"query\":\"$NONCE\",\"limit\":3}}}")

if command -v jq >/dev/null 2>&1; then
    echo "$REQUESTS" | "$BIN" mcp --source smoke-test | jq .
else
    echo "$REQUESTS" | "$BIN" mcp --source smoke-test
fi

echo ""
echo "== 期望 =="
echo "id=1: protocolVersion 2024-11-05 + serverInfo.name=pawnix-memory"
echo "id=2: tools 列表含 memory_search / memory_stats / memory_write（按名字排序）"
echo "id=3: content.text 显示记忆库健康状况（总条数/已嵌入/近24h）"
echo "id=4: content.text 显示「已写入记忆 [fact] (id=..., source=smoke-test)」"
echo "id=5: content.text 命中刚写入的「冒烟测试记忆 $NONCE」→ 写-读回路打通"
echo ""
echo "注: 冒烟测试会真的往共享池写一条 fact（带 source:smoke-test 标签）。"
echo "    如需清理: DELETE FROM memories WHERE 'source:smoke-test' = ANY(tags);"
