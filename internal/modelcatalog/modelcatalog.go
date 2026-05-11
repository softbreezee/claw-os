// Package modelcatalog maintains a JSON-based registry of model capabilities
// (context window, pricing tiers, etc.) used by the compaction engine and
// cost tracker to make informed decisions.
//
// The registry is persisted as ~/.fastclaw/model-catalog.json and ships with
// sensible built-in defaults for common models. Users can edit it in the
// Settings UI; changes take effect after manual reload or process restart.
package modelcatalog

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fastclaw-ai/fastclaw/internal/config"
)

// ModelInfo holds key capabilities for a single model.
type ModelInfo struct {
	ContextWindow  int     `json:"contextWindow"`  // total context window in tokens
	SoftThreshold  float64 `json:"softThreshold"`  // fraction of contextWindow for soft compaction (default 0.7)
	HardThreshold  float64 `json:"hardThreshold"`  // fraction of contextWindow for hard compaction (default 0.92)
	Description    string  `json:"description"`    // human-readable label
}

// Catalog is the top-level structure persisted on disk.
type Catalog struct {
	Models map[string]ModelInfo `json:"models"`
}

// DefaultThreshold is the fallback token count used when the model is not
// found in the catalog. At 80K tokens it is very conservative.
const DefaultThreshold = 80_000

// HardThresholdRatio is the fraction of contextWindow above which compaction
// MUST succeed — if push comes to shove, the engine will summarise more
// aggressively rather than risk a 400 from the provider.
const HardThresholdRatio = 0.92

// SoftThresholdRatio is the fraction of contextWindow at which the engine
// prefers to compact to keep room for the current turn's input + output.
const SoftThresholdRatio = 0.70

// builtinDefaults ships with the binary so the catalog is immediately useful
// even before the user touches the Settings UI.
var builtinDefaults = map[string]ModelInfo{
	// DeepSeek family
	"deepseek-v4-pro":   {ContextWindow: 1_048_576, SoftThreshold: SoftThresholdRatio, HardThreshold: HardThresholdRatio, Description: "DeepSeek v4 Pro"},
	"deepseek-v4-flash": {ContextWindow: 131_072, SoftThreshold: SoftThresholdRatio, HardThreshold: HardThresholdRatio, Description: "DeepSeek v4 Flash"},
	"deepseek-chat":     {ContextWindow: 131_072, SoftThreshold: SoftThresholdRatio, HardThreshold: HardThresholdRatio, Description: "DeepSeek Chat (v3)"},
	"deepseek-reasoner": {ContextWindow: 131_072, SoftThreshold: SoftThresholdRatio, HardThreshold: HardThresholdRatio, Description: "DeepSeek R1 Reasoner"},

	// Anthropic
	"claude-3-5-sonnet":      {ContextWindow: 200_000, SoftThreshold: SoftThresholdRatio, HardThreshold: HardThresholdRatio, Description: "Claude 3.5 Sonnet"},
	"claude-3-5-haiku":       {ContextWindow: 200_000, SoftThreshold: SoftThresholdRatio, HardThreshold: HardThresholdRatio, Description: "Claude 3.5 Haiku"},
	"claude-sonnet-4":        {ContextWindow: 200_000, SoftThreshold: SoftThresholdRatio, HardThreshold: HardThresholdRatio, Description: "Claude Sonnet 4"},
	"claude-opus-4":          {ContextWindow: 200_000, SoftThreshold: SoftThresholdRatio, HardThreshold: HardThresholdRatio, Description: "Claude Opus 4"},
	"claude-3-opus":          {ContextWindow: 200_000, SoftThreshold: SoftThresholdRatio, HardThreshold: HardThresholdRatio, Description: "Claude 3 Opus"},

	// OpenAI
	"gpt-4o":         {ContextWindow: 128_000, SoftThreshold: SoftThresholdRatio, HardThreshold: HardThresholdRatio, Description: "GPT-4o"},
	"gpt-4o-mini":    {ContextWindow: 128_000, SoftThreshold: SoftThresholdRatio, HardThreshold: HardThresholdRatio, Description: "GPT-4o mini"},
	"gpt-4-turbo":    {ContextWindow: 128_000, SoftThreshold: SoftThresholdRatio, HardThreshold: HardThresholdRatio, Description: "GPT-4 Turbo"},
	"gpt-5":          {ContextWindow: 400_000, SoftThreshold: SoftThresholdRatio, HardThreshold: HardThresholdRatio, Description: "GPT-5"},
	"o1":             {ContextWindow: 200_000, SoftThreshold: SoftThresholdRatio, HardThreshold: HardThresholdRatio, Description: "o1"},
	"o1-mini":        {ContextWindow: 200_000, SoftThreshold: SoftThresholdRatio, HardThreshold: HardThresholdRatio, Description: "o1-mini"},
	"o3":             {ContextWindow: 200_000, SoftThreshold: SoftThresholdRatio, HardThreshold: HardThresholdRatio, Description: "o3"},
	"o3-mini":        {ContextWindow: 200_000, SoftThreshold: SoftThresholdRatio, HardThreshold: HardThresholdRatio, Description: "o3-mini"},
	"o4-mini":        {ContextWindow: 200_000, SoftThreshold: SoftThresholdRatio, HardThreshold: HardThresholdRatio, Description: "o4-mini"},

	// Google
	"gemini-2.0-flash":     {ContextWindow: 1_048_576, SoftThreshold: SoftThresholdRatio, HardThreshold: HardThresholdRatio, Description: "Gemini 2.0 Flash"},
	"gemini-2.0-pro":       {ContextWindow: 2_097_152, SoftThreshold: SoftThresholdRatio, HardThreshold: HardThresholdRatio, Description: "Gemini 2.0 Pro"},
	"gemini-2.5-flash":     {ContextWindow: 1_048_576, SoftThreshold: SoftThresholdRatio, HardThreshold: HardThresholdRatio, Description: "Gemini 2.5 Flash"},
	"gemini-2.5-pro":       {ContextWindow: 2_097_152, SoftThreshold: SoftThresholdRatio, HardThreshold: HardThresholdRatio, Description: "Gemini 2.5 Pro"},

	// Moonshot / Kimi
	"kimi-k2":    {ContextWindow: 262_144, SoftThreshold: SoftThresholdRatio, HardThreshold: HardThresholdRatio, Description: "Kimi K2"},
	"kimi-k2.5":  {ContextWindow: 262_144, SoftThreshold: SoftThresholdRatio, HardThreshold: HardThresholdRatio, Description: "Kimi K2.5"},

	// Qwen
	"qwen2.5-72b":  {ContextWindow: 131_072, SoftThreshold: SoftThresholdRatio, HardThreshold: HardThresholdRatio, Description: "Qwen 2.5 72B"},
	"qwen-max":     {ContextWindow: 131_072, SoftThreshold: SoftThresholdRatio, HardThreshold: HardThresholdRatio, Description: "Qwen Max"},
}

// ResolvedThreshold holds the model-specific compaction thresholds after lookup.
type ResolvedThreshold struct {
	Soft int // token count at which soft compaction triggers
	Hard int // token count at which hard compaction triggers
}

var (
	globalMu      sync.RWMutex
	globalCatalog *Catalog // nil until first lookup or reload
	globalPath    string
)

func initPath() string {
	globalMu.RLock()
	if globalPath != "" {
		path := globalPath
		globalMu.RUnlock()
		return path
	}
	globalMu.RUnlock()

	homeDir, err := config.HomeDir()
	if err != nil {
		return ""
	}
	p := filepath.Join(homeDir, "model-catalog.json")

	globalMu.Lock()
	globalPath = p
	globalMu.Unlock()
	return p
}

// loadCatalog returns the current Catalog, loading from disk or built-in defaults
// as needed. Caller must hold at least RLock.
func loadCatalog() *Catalog {
	if globalCatalog != nil {
		return globalCatalog
	}
	// Upgrade to write lock for lazy init
	globalMu.RUnlock()
	globalMu.Lock()
	defer func() { globalMu.Unlock(); globalMu.RLock() }()

	if globalCatalog != nil {
		return globalCatalog
	}

	path := globalPath
	cat := tryLoadDisk(path)
	if cat == nil {
		cat = &Catalog{Models: builtinDefaults}
	}
	globalCatalog = cat
	return cat
}

func tryLoadDisk(path string) *Catalog {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var cat Catalog
	if err := json.Unmarshal(data, &cat); err != nil {
		slog.Warn("modelcatalog: failed to parse, ignoring disk file", "path", path, "error", err)
		return nil
	}
	if cat.Models == nil {
		cat.Models = map[string]ModelInfo{}
	}
	slog.Info("modelcatalog: loaded from disk", "path", path, "models", len(cat.Models))
	return &cat
}

// Reload forces a re-read from disk. Called after the user saves an updated
// catalog via the Settings UI.
func Reload() error {
	path := initPath()

	globalMu.Lock()
	defer globalMu.Unlock()

	cat := tryLoadDisk(path)
	if cat == nil {
		// Disk file doesn't exist yet — fall back to built-in defaults for the
		// in-memory copy. The file will be created on the next Save() call.
		cat = &Catalog{Models: builtinDefaults}
	}
	globalCatalog = cat
	return nil
}

// Save persists the given catalog to disk and updates the in-memory copy.
func Save(cat *Catalog) error {
	path := initPath()
	if path == "" {
		return fmt.Errorf("modelcatalog: cannot determine home directory")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("modelcatalog: mkdir: %w", err)
	}

	data, err := json.MarshalIndent(cat, "", "  ")
	if err != nil {
		return fmt.Errorf("modelcatalog: marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("modelcatalog: write: %w", err)
	}

	globalMu.Lock()
	globalCatalog = cat
	globalMu.Unlock()

	slog.Info("modelcatalog: saved", "path", path, "models", len(cat.Models))
	return nil
}

// Get returns the current catalog for inspection (e.g. Settings UI GET).
func Get() *Catalog {
	initPath()
	globalMu.RLock()
	defer globalMu.RUnlock()
	return loadCatalog()
}

// LookupThreshold resolves compaction thresholds for a model ID, given its
// context window from the catalog. Returns (soft, hard) token counts.
// Falls back to DefaultThreshold if the model is not in the catalog.
func LookupThreshold(modelID string) ResolvedThreshold {
	cat := Get()

	// Strip provider prefix: "deepseek/deepseek-v4-pro" → "deepseek-v4-pro",
	// "openrouter/anthropic/claude-sonnet-4" → "claude-sonnet-4"
	clean := modelID
	if idx := strings.LastIndex(clean, "/"); idx >= 0 {
		clean = clean[idx+1:]
	}

	// Exact match first
	if info, ok := cat.Models[clean]; ok {
		return resolvedFrom(info)
	}

	// Prefix match — covers families like "deepseek-v4-pro-20250501"
	lower := strings.ToLower(clean)
	for key, info := range cat.Models {
		if strings.HasPrefix(lower, strings.ToLower(key)) {
			return resolvedFrom(info)
		}
	}

	// Also check from config.ModelEntry.contextWindow — if the user filled
	// it in the Models UI but didn't add an entry in the catalog, honour it.
	cfg, err := config.Load()
	if err == nil {
		for _, prov := range cfg.Providers {
			for _, m := range prov.Models {
				// Match on the ID the agent uses ("provider/model")
				if m.ID == clean && m.ContextWindow > 0 {
					return ResolvedThreshold{
						Soft: int(float64(m.ContextWindow) * SoftThresholdRatio),
						Hard: int(float64(m.ContextWindow) * HardThresholdRatio),
					}
				}
			}
		}
	}

	return ResolvedThreshold{Soft: DefaultThreshold, Hard: DefaultThreshold}
}

func resolvedFrom(info ModelInfo) ResolvedThreshold {
	soft := info.SoftThreshold
	if soft <= 0 {
		soft = SoftThresholdRatio
	}
	hard := info.HardThreshold
	if hard <= 0 {
		hard = HardThresholdRatio
	}
	return ResolvedThreshold{
		Soft: int(float64(info.ContextWindow) * soft),
		Hard: int(float64(info.ContextWindow) * hard),
	}
}

// MergeBuiltins adds any built-in models that are missing from the user's
// catalog, so they benefit from updates without losing manual edits.
func MergeBuiltins(cat *Catalog) {
	if cat.Models == nil {
		cat.Models = builtinDefaults
		return
	}
	for k, v := range builtinDefaults {
		if _, exists := cat.Models[k]; !exists {
			cat.Models[k] = v
		}
	}
}