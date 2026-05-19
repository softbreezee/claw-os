package agent

import (
	"context"
	"log/slog"
	"strings"

	"github.com/softbreezee/claw-os/internal/provider"
)

// embedProvider returns the provider that should serve embedding
// requests for the given embedModel string. It uses the provider
// registry's "prefix/" routing (same as chat model routing) so
// "qwen-embed/text-embedding-v3" hits the qwen-embed provider,
// not the agent's default chat provider.
func (a *Agent) embedProvider(embedModel string) provider.Provider {
	a.providerMu.RLock()
	reg := a.providerRegistry
	a.providerMu.RUnlock()
	if reg != nil {
		if p := reg.For(embedModel); p != nil {
			return p
		}
	}
	// Fallback: agent's own provider (works when the chat provider
	// also supports /embeddings, e.g. OpenAI).
	return a.getProvider()
}

// searchRelevantMemory turns the current user query into the top-K
// semantic hits from the agent's pg memory store, ready to be passed
// to ContextBuilder.BuildSystemPromptWithMemory.
//
// Returns nil (silently) in every "can't or shouldn't" branch:
//   - pg backend not wired (file-only storage)
//   - empty/whitespace user query (typing indicator, /slash command, etc)
//   - embedding model not configured
//   - embedding API call failed
//   - pg search returned nothing
//
// We never propagate an error: this is best-effort enrichment, the
// LLM call should not be blocked by a memory-search hiccup. All non-
// trivial failures are logged at debug so users running with -v can
// diagnose, but the default log level stays clean.
func (a *Agent) searchRelevantMemory(ctx context.Context, query string) []MemoryHit {
	pg := a.memory.PGStore()
	if pg == nil {
		slog.Info("memory search: skipped (no pg store)", "agent", a.name)
		return nil
	}
	q := strings.TrimSpace(query)
	if q == "" {
		return nil
	}
	embedModel := a.memoryCfg.EmbedModel
	if embedModel == "" {
		slog.Info("memory search: skipped (no embedModel configured)", "agent", a.name)
		return nil
	}

	// Route to the correct provider for the embed model. The embed
	// model string uses the same "provider/model" convention as chat
	// models (e.g. "qwen-embed/text-embedding-v3"), so registry.For()
	// picks the right upstream. Without this we'd hit the agent's
	// default chat provider (e.g. deepseek) which doesn't have an
	// /embeddings endpoint.
	prov := a.embedProvider(embedModel)
	if prov == nil {
		slog.Warn("memory search: embed provider not found", "agent", a.name, "embedModel", embedModel)
		return nil
	}

	slog.Info("memory search: embedding query", "agent", a.name, "embedModel", embedModel, "queryLen", len(q))
	emb, err := prov.Embed(ctx, q, embedModel)
	if err != nil {
		if err == provider.ErrEmbeddingNotSupported {
			slog.Info("memory search: embed not supported by provider", "agent", a.name)
			return nil
		}
		slog.Warn("memory search: embed failed", "agent", a.name, "error", err)
		return nil
	}

	topK := a.memoryCfg.SemanticTopK
	if topK <= 0 {
		topK = 5
	}
	hits, err := pg.SearchSemantic(ctx, a.memory.AgentID(), emb, topK)
	if err != nil {
		slog.Debug("memory search: pg query failed", "agent", a.name, "error", err)
		return nil
	}
	if len(hits) > 0 {
		slog.Info("memory search: hits",
			"agent", a.name, "count", len(hits), "top_k", topK)
	}
	return hits
}
