#!/usr/bin/env bash
# Memory pipeline end-to-end verification.
# Run AFTER chatting with the agent for >=6 turns (so AutoPersist has fired).
# See docs/memory-verification.md for the full flow.
#
# Modes (auto-detected, or override via env):
#   Local PG (default): reads DSN from ~/.pawnix/pawnix.json
#   Docker mode:        set PG_CONTAINER=<name> to use docker exec instead
#
# Env overrides:
#   PG_DSN        — explicit DSN, e.g. postgres://user:pass@localhost/db
#   PG_CONTAINER  — docker container name (activates docker mode)
#   PG_USER       — postgres user (docker mode only, default: postgres)
#   PG_DB         — postgres db   (docker mode only, default: pawnix)
#   LOG_FILE      — path to pawnix gateway log (default: /tmp/pawnix-verify.log)

set -uo pipefail

LOG_FILE="${LOG_FILE:-/tmp/pawnix-verify.log}"

# ---------------- Pretty printing --------------------------
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
DIM='\033[2m'
RESET='\033[0m'

ok()   { printf "${GREEN}✅${RESET} %s\n" "$1"; }
fail() { printf "${RED}❌${RESET} %s\n" "$1"; FAILED=1; }
warn() { printf "${YELLOW}⚠️${RESET}  %s\n" "$1"; }
hint() { printf "   ${DIM}↳ %s${RESET}\n" "$1"; }

FAILED=0

# ---------------- Detect mode & build psql_exec ------------
# Priority: PG_CONTAINER env > PG_DSN env > auto-detect from pawnix.json

USE_DOCKER=0
RESOLVED_DSN=""

if [[ -n "${PG_CONTAINER:-}" ]]; then
    # Explicit docker mode
    USE_DOCKER=1
    PG_USER="${PG_USER:-postgres}"
    PG_DB="${PG_DB:-pawnix}"
elif [[ -n "${PG_DSN:-}" ]]; then
    # Explicit DSN
    RESOLVED_DSN="$PG_DSN"
else
    # Auto-detect: parse DSN from ~/.pawnix/pawnix.json
    CONFIG_FILE="$HOME/.pawnix/pawnix.json"
    if [[ -f "$CONFIG_FILE" ]]; then
        # Try python3 first, then fallback to grep+sed
        if command -v python3 >/dev/null 2>&1; then
            RESOLVED_DSN=$(python3 -c "
import json, sys
try:
    d = json.load(open('$CONFIG_FILE'))
    print(d.get('storage', {}).get('dsn', ''))
except Exception:
    pass
" 2>/dev/null)
        fi
        # Fallback: crude grep for the dsn field
        if [[ -z "$RESOLVED_DSN" ]]; then
            RESOLVED_DSN=$(grep -o '"dsn"[[:space:]]*:[[:space:]]*"[^"]*"' "$CONFIG_FILE" \
                | sed 's/.*"dsn"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' | head -1)
        fi
    fi
fi

psql_exec() {
    if [[ $USE_DOCKER -eq 1 ]]; then
        docker exec "$PG_CONTAINER" psql -U "$PG_USER" -d "$PG_DB" -tA -c "$1" 2>/dev/null
    else
        psql "$RESOLVED_DSN" -tA -c "$1" 2>/dev/null
    fi
}

step() {
    printf "[%s] %-50s " "$1" "$2"
}

echo "============================================"
echo " Pawnix Memory Pipeline Verification"
echo "============================================"
if [[ $USE_DOCKER -eq 1 ]]; then
    echo " Mode: Docker  (container: $PG_CONTAINER)"
else
    # Mask password in displayed DSN
    DISPLAY_DSN=$(echo "$RESOLVED_DSN" | sed 's|://[^:]*:[^@]*@|://***:***@|')
    echo " Mode: Local PG (dsn: ${DISPLAY_DSN:-not found})"
fi
echo

# ---------------- [1/7] PG reachable -----------------------
step "1/7" "PG reachable"
if [[ $USE_DOCKER -eq 1 ]]; then
    if docker exec "$PG_CONTAINER" pg_isready -U "$PG_USER" -d "$PG_DB" >/dev/null 2>&1; then
        ok ""
    else
        fail ""
        hint "container '$PG_CONTAINER' not running. start it per docs/memory-verification.md Step 1"
        echo; echo "PIPELINE STATUS: BLOCKED ❌"; exit 1
    fi
else
    if [[ -z "$RESOLVED_DSN" ]]; then
        fail ""
        hint "could not find DSN. set PG_DSN=<dsn> or ensure ~/.pawnix/pawnix.json has storage.dsn"
        echo; echo "PIPELINE STATUS: BLOCKED ❌"; exit 1
    fi
    if psql "$RESOLVED_DSN" -tA -c "SELECT 1;" >/dev/null 2>&1; then
        ok ""
    else
        fail ""
        hint "cannot connect with DSN from pawnix.json. check PG is running and credentials are correct"
        hint "override: PG_DSN='postgres://user:pass@host/db' $0"
        echo; echo "PIPELINE STATUS: BLOCKED ❌"; exit 1
    fi
fi

# ---------------- [2/7] vector extension -------------------
step "2/7" "vector extension installed"
EXT=$(psql_exec "SELECT extname FROM pg_extension WHERE extname='vector';")
if [[ "$EXT" == "vector" ]]; then
    ok ""
else
    fail ""
    if [[ $USE_DOCKER -eq 1 ]]; then
        hint "run: docker exec $PG_CONTAINER psql -U $PG_USER -d $PG_DB -c 'CREATE EXTENSION vector;'"
    else
        hint "run: psql \"\$PG_DSN\" -c 'CREATE EXTENSION IF NOT EXISTS vector;'"
        hint "macOS: brew install pgvector  |  Linux: apt install postgresql-*-pgvector"
    fi
fi

# ---------------- [3/7] memories table ---------------------
step "3/7" "memories table exists"
TBL=$(psql_exec "SELECT tablename FROM pg_tables WHERE tablename='memories';")
if [[ "$TBL" == "memories" ]]; then
    ok ""
else
    fail ""
    hint "autoMigrate didn't run. check pawnix.json storage.autoMigrate=true and restart"
    echo; echo "PIPELINE STATUS: BLOCKED ❌"; exit 1
fi

# ---------------- [4/7] rows count -------------------------
step "4/7" "memories rows count"
COUNT=$(psql_exec "SELECT count(*) FROM memories;")
if [[ "${COUNT:-0}" -gt 0 ]]; then
    ok "($COUNT rows)"
else
    fail "(0 rows)"
    hint "AutoPersist hasn't fired. chat >= autoPersist.everyNTurns turns; check logs for 'auto-persist:'"
    echo; echo "PIPELINE STATUS: WRITE-PATH BROKEN ❌"; exit 1
fi

# ---------------- [5/7] embedding populated ----------------
step "5/7" "embedding column populated"
EMB_NN=$(psql_exec "SELECT count(*) FROM memories WHERE embedding IS NOT NULL;")
EMB_NULL=$(psql_exec "SELECT count(*) FROM memories WHERE embedding IS NULL;")

# NULL embeddings on non-memory kinds (e.g. ta_job JSON records) are expected
# and don't indicate a broken pipeline. Check if all NULLs are non-memory kinds.
NULL_FACTS=$(psql_exec "SELECT count(*) FROM memories WHERE embedding IS NULL AND kind IN ('fact','user_note');")

if [[ "${EMB_NN:-0}" == "${COUNT}" && "${EMB_NULL:-0}" == "0" ]]; then
    ok "($EMB_NN/$COUNT non-NULL)"
elif [[ "${EMB_NULL:-0}" -gt 0 && "${NULL_FACTS:-0}" == "0" ]]; then
    # All NULLs are non-fact kinds (e.g. ta_job) — this is expected behaviour
    warn "($EMB_NN/$COUNT non-NULL — ${EMB_NULL} are non-fact kinds, expected)"
    hint "NULL embeddings are all non-fact records (e.g. ta_job). embedding pipeline is healthy."
elif [[ "${EMB_NN:-0}" -gt 0 ]]; then
    warn "($EMB_NN/$COUNT non-NULL — partial, ${NULL_FACTS} fact/user_note rows missing embedding)"
    hint "some fact inserts had embed failures. check logs: grep 'embed failed' $LOG_FILE"
else
    fail "(0/$COUNT non-NULL)"
    hint "embedding pipeline broken. likely embedModel not routed to a real provider."
    hint "check: agent's embedModel config has 'provider/' prefix; provider is in registry"
    hint "logs: grep -E 'embed (failed|not supported)' $LOG_FILE"
fi

# ---------------- [6/7] semantic search works --------------
step "6/7" "semantic search returns results"
TOP=$(psql_exec "
    WITH probe AS (SELECT embedding FROM memories WHERE embedding IS NOT NULL LIMIT 1)
    SELECT (embedding <=> (SELECT embedding FROM probe))::numeric(10,3)
    FROM memories
    WHERE embedding IS NOT NULL
    ORDER BY embedding <=> (SELECT embedding FROM probe)
    LIMIT 1;")
if [[ -n "$TOP" ]]; then
    ok "(top hit: cosine $TOP)"
else
    fail ""
    hint "pgvector query failed. confirm vector extension + index 'memories_embedding_hnsw' exist"
fi

# ---------------- [7/7] read-path log evidence -------------
step "7/7" "log shows 'memory search: hits'"
if [[ ! -f "$LOG_FILE" ]]; then
    warn "(log file $LOG_FILE not found — skipped)"
    hint "re-run gateway with: ./bin/pawnix gateway 2>&1 | tee $LOG_FILE"
    hint "then open a NEW chat session and send a message to trigger memory search"
else
    HITS=$(grep -c "memory search: hits" "$LOG_FILE" 2>/dev/null || true)
    if [[ "${HITS:-0}" -gt 0 ]]; then
        ok "($HITS occurrences)"
        echo
        echo "Last 3 hits:"
        grep "memory search: hits" "$LOG_FILE" | tail -3 | sed 's/^/  /'
    else
        fail "(0 occurrences)"
        hint "read path not engaging. either:"
        hint "  (a) you haven't sent any messages in a NEW session since the writes"
        hint "  (b) memstore_adapter not wired — check gateway/memstore_adapter.go injection"
        hint "  (c) embedModel not configured on the agent"
    fi
fi

# ---------------- Summary ----------------------------------
echo
if [[ $FAILED -eq 0 ]]; then
    echo -e "PIPELINE STATUS: ${GREEN}HEALTHY${RESET} ✨"
    echo
    echo "Sample memories (most recent 3):"
    psql_exec "SELECT kind, left(content, 80) FROM memories ORDER BY created_at DESC LIMIT 3;" \
        | awk -F'|' '{ printf "  [%s] %s\n", $1, $2 }'
    exit 0
else
    echo -e "PIPELINE STATUS: ${RED}DEGRADED${RESET} ❌"
    echo "Fix the ❌ items above and re-run."
    exit 1
fi
