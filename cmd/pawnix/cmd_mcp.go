package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/softbreezee/claw-os/internal/config"
	"github.com/softbreezee/claw-os/internal/mcpserver"
	"github.com/softbreezee/claw-os/internal/provider"
	pgstore "github.com/softbreezee/claw-os/internal/store/pg"
)

// defaultMemoryEmbedModel matches the vector(1536) column shipped by
// db.Migrate; used when cfg.Memory.EmbedModel is unset.
const defaultMemoryEmbedModel = "openai/text-embedding-3-small"

// defaultMemoryPool is the shared agent_id all cooperating tools read
// from by default. One fixed pool = hermes / claude-code / codex see
// each other's memory; per-row source tags preserve provenance.
const defaultMemoryPool = "shared"

// mcpCmd wires `pawnix mcp` — a stdio MCP server that exposes the
// pgvector memory pool to any MCP client. This is the connection layer
// that lets hermes / claude-code / codex share one persistent memory.
func mcpCmd() *cobra.Command {
	var agentID string
	var source string
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Run the memory MCP server (stdio) for hermes / claude-code / codex",
		Long: `Exposes the PostgreSQL+pgvector memory pool over the Model Context
Protocol on stdin/stdout. Spawn this as a stdio MCP subprocess from any
MCP client (hermes has native support; claude-code and codex via their
mcpServers config). Shared state lives in Postgres, so multiple client
subprocesses collaborate naturally at the DB layer.

Tools: memory_search, memory_stats (read-only) and memory_write.
Set --source to tag every written memory with its origin tool, e.g.
  pawnix mcp --source claude-code`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCP(agentID, source)
		},
	}
	cmd.Flags().StringVar(&agentID, "agent", defaultMemoryPool,
		"memory pool id (default shared pool; override to read one agent's memory)")
	cmd.Flags().StringVar(&source, "source", "",
		"origin tag baked into every written memory (e.g. hermes, codex, claude-code)")
	return cmd
}

func runMCP(agentID, source string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if cfg.Storage.Type != "postgres" || cfg.Storage.DSN == "" {
		return fmt.Errorf("storage.type must be 'postgres' with a valid dsn (memory backend needs pgvector)")
	}

	ctx := context.Background()
	db, err := pgstore.Open(ctx, cfg.Storage.DSN)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer db.Close()

	memStore := pgstore.NewMemoryStore(db)

	// Observability: log every tool call to mcp_events so the dashboard
	// can show per-source / per-session usage. Telemetry is best-effort
	// and must NEVER break a tool call, so schema-provisioning failure
	// just disables logging (logger.store stays nil). runMCP calls Open
	// not Migrate, so the table self-provisions here.
	eventStore := pgstore.NewEventStore(db)
	if err := eventStore.EnsureSchema(ctx); err != nil {
		slog.Warn("mcp: telemetry disabled (schema provisioning failed)", "error", err)
		eventStore = nil
	}
	logger := &mcpLogger{
		store:   eventStore,
		connID:  uuid.NewString(),
		source:  source,
		agentID: agentID,
	}

	// Build the embedder for semantic search. Optional: if it can't be
	// constructed (no embed model, or provider not configured), reads
	// degrade to keyword+recency search rather than failing outright.
	embedder, embedModel := buildEmbedder(cfg)

	srv := mcpserver.NewServer("pawnix-memory", version)
	registerMemorySearch(srv, memStore, embedder, embedModel, agentID, logger)
	registerMemoryStats(srv, memStore, agentID, logger)
	registerMemoryWrite(srv, memStore, embedder, embedModel, agentID, source, logger)

	return srv.Serve()
}

// mcpLogger writes one mcp_events row per tool call. All fields are set
// once at spawn time except the per-call event. connID identifies this
// subprocess (≈ one client session); source/agentID come from flags.
type mcpLogger struct {
	store   *pgstore.EventStore
	connID  string
	source  string
	agentID string
}

// log records one tool call. Best-effort: a nil store (telemetry
// disabled) or a write error is swallowed after a warning on stderr —
// stdout carries the JSON-RPC stream, so slog (stderr) is safe here.
func (l *mcpLogger) log(tool, query, kind string, resultCount int, hit bool, callErr error, started time.Time) {
	if l == nil || l.store == nil {
		return
	}
	e := pgstore.MCPEvent{
		ConnectionID: l.connID,
		Source:       l.source,
		AgentID:      l.agentID,
		Tool:         tool,
		Query:        query,
		Kind:         kind,
		ResultCount:  resultCount,
		Hit:          hit,
		DurationMs:   int(time.Since(started).Milliseconds()),
	}
	if callErr != nil {
		e.Error = callErr.Error()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := l.store.Insert(ctx, e); err != nil {
		slog.Warn("mcp: telemetry insert failed", "tool", tool, "error", err)
	}
}

// buildEmbedder resolves the "provider/model" embed spec into a live
// provider. Returns (nil, "") when embeddings are unavailable so the
// caller can fall back to keyword search.
func buildEmbedder(cfg *config.Config) (provider.Provider, string) {
	embedModel := cfg.Memory.EmbedModel
	if embedModel == "" {
		embedModel = defaultMemoryEmbedModel
	}
	idx := strings.Index(embedModel, "/")
	if idx < 0 {
		return nil, ""
	}
	providerName := embedModel[:idx]
	provCfg, ok := cfg.Providers[providerName]
	if !ok {
		return nil, ""
	}
	// Embed() strips the provider prefix internally, so pass embedModel as-is.
	return provider.NewProvider(provCfg.APIKey, provCfg.APIBase, provCfg.APIType, provCfg.EmbedPath), embedModel
}

// ── Tool descriptions ─────────────────────────────────────────────
// The WHEN-to-read decision lives in the CALLING model, steered by
// these descriptions. Principle for reads: be aggressive — cheap to
// query, expensive to miss context.

const memorySearchDesc = `检索长期记忆库。这是一个跨会话、跨工具共享的持久记忆池（hermes、claude-code、codex 共用），存放用户偏好、历史决策、项目背景、之前确认过的事实。

【何时调用 —— 要积极】每轮对话开始时，只要问题可能涉及过往上下文，就先检索一次。宁可多查、不要漏查：检索成本很低，漏掉记忆会导致答非所问或重复劳动。

正例（应当检索）：
- 用户提到"上次""之前""我们说过的" → 立刻检索
- 出现某个项目名 / 人名 / 系统名 → 用它作关键词检索
- 用户问"我的偏好""我一般怎么做" → 检索
- 开始一个新任务前，先查相关背景

反例（不必检索）：
- 纯即时计算、翻译、格式转换等不依赖历史的任务
- 用户明确说"不用管以前的"

参数：query（检索词，自然语言或关键词）；limit（返回条数，默认 5）。`

const memoryStatsDesc = `查看记忆库健康状况：总条数、已生成向量嵌入的条数、最近 24 小时的写入量与嵌入覆盖率。用于诊断记忆系统是否正常（例如嵌入管线是否挂了、记忆池是否为空）。一般排查问题时才调用，正常对话无需调用。`

const memoryWriteDesc = `向长期记忆库写入一条记忆。这是跨会话、跨工具共享的持久池（hermes、claude-code、codex 共用），写进去的内容会长期影响未来所有会话，所以务必谨慎。

【何时调用 —— 要保守】只在信息满足"跨会话仍然重要"时才写。写之前先自问：下次新开一个会话，这条信息还值得被记住吗？拿不准就先问用户，不要擅自写。写多了会污染记忆池，比不写更糟。

应当写：
- 用户明确说"记住…""以后都这样""下次注意" → 写
- 稳定的用户偏好、长期项目背景、已确认的关键决策与事实
- 用户纠正过你、且这个纠正对以后仍然适用

不要写：
- 本轮任务的临时上下文、中间结果、可从文件/工具重新查到的信息
- 尚未确认、还在讨论中的想法
- 敏感个人信息（身份证件/银行卡/健康/家庭住址等），除非用户明确要求记住

参数：content（记忆正文，写清楚"是什么"和"为什么"，让未来的会话能独立看懂）；kind（fact=事实/user_note=用户偏好或叮嘱/report=阶段结论，默认 user_note）。来源标签由启动参数自动附加，无需手填。`

// ── Tool registration ─────────────────────────────────────────────

func registerMemorySearch(srv *mcpserver.Server, store *pgstore.MemoryStore, embedder provider.Provider, embedModel, agentID string, logger *mcpLogger) {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "检索词，自然语言或关键词均可",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "返回条数，默认 5",
			},
		},
		"required": []string{"query"},
	}

	srv.Register("memory_search", memorySearchDesc, schema, func(args json.RawMessage) (string, error) {
		var in struct {
			Query string `json:"query"`
			Limit int    `json:"limit"`
		}
		if err := json.Unmarshal(args, &in); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}
		if strings.TrimSpace(in.Query) == "" {
			return "", fmt.Errorf("query is required")
		}
		limit := in.Limit
		if limit <= 0 {
			limit = 5
		}

		started := time.Now()
		ctx := context.Background()
		var queryEmbedding []float32
		if embedder != nil {
			if emb, err := embedder.Embed(ctx, in.Query, embedModel); err == nil {
				queryEmbedding = emb
			}
		}

		// SearchSemantic transparently falls back to keyword/recency
		// search when queryEmbedding is empty.
		records, err := store.SearchSemantic(ctx, agentID, queryEmbedding, limit)
		logger.log("memory_search", in.Query, "", len(records), len(records) > 0, err, started)
		if err != nil {
			return "", err
		}
		return formatMemories(records), nil
	})
}

func registerMemoryStats(srv *mcpserver.Server, store *pgstore.MemoryStore, agentID string, logger *mcpLogger) {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}

	srv.Register("memory_stats", memoryStatsDesc, schema, func(args json.RawMessage) (string, error) {
		started := time.Now()
		ctx := context.Background()
		h, err := store.HealthStats(ctx, agentID)
		logger.log("memory_stats", "", "", int(h.Total), h.Total > 0, err, started)
		if err != nil {
			return "", err
		}
		var b strings.Builder
		fmt.Fprintf(&b, "记忆库健康状况（pool=%s）：\n", agentID)
		fmt.Fprintf(&b, "- 总条数: %d\n", h.Total)
		fmt.Fprintf(&b, "- 已嵌入: %d (%d%%)\n", h.WithEmbedding, h.CoveragePct())
		fmt.Fprintf(&b, "- 近24h写入: %d\n", h.RecentTotal)
		fmt.Fprintf(&b, "- 近24h已嵌入: %d (%d%%)\n", h.RecentEmbedded, h.RecentCoveragePct())
		return b.String(), nil
	})
}

func registerMemoryWrite(srv *mcpserver.Server, store *pgstore.MemoryStore, embedder provider.Provider, embedModel, agentID, source string, logger *mcpLogger) {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"content": map[string]interface{}{
				"type":        "string",
				"description": "记忆正文，写清楚是什么和为什么，让未来会话能独立看懂",
			},
			"kind": map[string]interface{}{
				"type":        "string",
				"description": "fact / user_note / report，默认 user_note",
				"enum":        []string{"fact", "user_note", "report"},
			},
		},
		"required": []string{"content"},
	}

	srv.Register("memory_write", memoryWriteDesc, schema, func(args json.RawMessage) (string, error) {
		var in struct {
			Content string `json:"content"`
			Kind    string `json:"kind"`
		}
		if err := json.Unmarshal(args, &in); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}
		content := strings.TrimSpace(in.Content)
		if content == "" {
			return "", fmt.Errorf("content is required")
		}
		kind := in.Kind
		if kind == "" {
			kind = "user_note"
		}

		// Provenance tag is baked in at spawn time via --source, not
		// self-reported by the model (a model can't reliably know which
		// tool it's running inside). Empty source → no tag.
		var tags []string
		if source != "" {
			tags = []string{"source:" + source}
		}

		started := time.Now()
		ctx := context.Background()
		var embedding []float32
		if embedder != nil {
			if emb, err := embedder.Embed(ctx, content, embedModel); err == nil {
				embedding = emb
			}
		}

		id, err := store.Insert(ctx, agentID, kind, content, embedding, tags)
		writeCount := 1
		if err != nil {
			writeCount = 0
		}
		logger.log("memory_write", "", kind, writeCount, false, err, started)
		if err != nil {
			return "", err
		}
		msg := fmt.Sprintf("已写入记忆 [%s] (id=%s", kind, id)
		if source != "" {
			msg += ", source=" + source
		}
		if embedding == nil {
			msg += ", 未生成向量嵌入(降级为关键词检索)"
		}
		msg += ")"
		return msg, nil
	})
}

func formatMemories(records []pgstore.MemoryRecord) string {
	if len(records) == 0 {
		return "（未找到相关记忆）"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "找到 %d 条相关记忆：\n\n", len(records))
	for i, r := range records {
		fmt.Fprintf(&b, "%d. [%s] %s\n", i+1, r.Kind, r.Content)
		meta := []string{}
		if len(r.Tags) > 0 {
			meta = append(meta, "标签: "+strings.Join(r.Tags, ", "))
		}
		if !r.CreatedAt.IsZero() {
			meta = append(meta, "时间: "+r.CreatedAt.Format("2006-01-02 15:04"))
		}
		if len(meta) > 0 {
			fmt.Fprintf(&b, "   (%s)\n", strings.Join(meta, " | "))
		}
	}
	return b.String()
}
