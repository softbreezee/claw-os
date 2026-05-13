package agent

import (
	"log/slog"
	"strings"

	"github.com/fastclaw-ai/fastclaw/internal/bus"
	"github.com/fastclaw-ai/fastclaw/internal/config"
	"github.com/fastclaw-ai/fastclaw/internal/provider"
)

// Manager loads and manages all agent instances.
type Manager struct {
	agents       map[string]*Agent
	defaultAgent *Agent
	registry     *provider.Registry // shared across all agents; used for routing
}

// NewManager creates agents from resolved configs.
//
// The provider Registry is the source of truth for routing: each
// agent picks its provider by looking up the prefix of its model
// (e.g. "openai/kimi-k2.6" -> registry["openai"]). This means a
// single fastclaw instance can mix multiple LLM backends without
// the awkward "every agent shares the same provider" hack the older
// signature forced.
func NewManager(resolved []config.ResolvedAgent, registry *provider.Registry, mb *bus.MessageBus) (*Manager, error) {
	m := &Manager{
		agents:   make(map[string]*Agent),
		registry: registry,
	}

	homeDir, err := config.HomeDir()
	if err != nil {
		return nil, err
	}

	for _, rc := range resolved {
		prov := registry.For(rc.Model)
		ag := NewAgent(rc, prov, mb, homeDir)
		// Hand the agent the shared registry so per-call model
		// overrides (chat-UI picker) can re-route to the matching
		// upstream instead of being pinned to the agent's bound
		// provider — see Agent.effectiveProvider.
		ag.SetProviderRegistry(registry)
		m.agents[rc.ID] = ag

		slog.Info("loaded agent",
			"id", rc.ID,
			"model", rc.Model,
			"provider", providerNameForModel(rc.Model, registry),
			"workspace", rc.Workspace,
		)
	}

	// If only one agent, make it the default
	if len(m.agents) == 1 {
		for _, ag := range m.agents {
			m.defaultAgent = ag
		}
	}

	return m, nil
}

// providerNameForModel returns the registered provider name that would
// be selected for the given model, or "default" if no prefix matches.
// Pure diagnostic helper — does not mutate state.
func providerNameForModel(model string, registry *provider.Registry) string {
	if idx := strings.Index(model, "/"); idx > 0 {
		name := model[:idx]
		if registry.Has(name) {
			return name
		}
	}
	return "(default)"
}

// AgentByID returns an agent by its ID.
func (m *Manager) AgentByID(id string) *Agent {
	return m.agents[id]
}

// DefaultAgent returns the default agent (set when only one agent exists).
func (m *Manager) DefaultAgent() *Agent {
	return m.defaultAgent
}

// All returns all loaded agents.
func (m *Manager) All() []*Agent {
	result := make([]*Agent, 0, len(m.agents))
	for _, ag := range m.agents {
		result = append(result, ag)
	}
	return result
}

// Names returns all agent IDs.
func (m *Manager) Names() []string {
	names := make([]string, 0, len(m.agents))
	for name := range m.agents {
		names = append(names, name)
	}
	return names
}

// UpdateProvider replaces the LLM provider for all agents (hot-reload).
//
// Deprecated: prefer Rewire, which respects per-agent model→provider
// routing. UpdateProvider is kept for callers that genuinely want
// every agent to point at a single provider (none in tree).
func (m *Manager) UpdateProvider(prov provider.Provider) {
	for _, ag := range m.agents {
		ag.setProvider(prov)
	}
}

// Rewire re-resolves each agent's provider from the registry. Use
// this on hot-reload after the registry has been refreshed: an agent
// whose model changed from "deepseek/x" to "openai/y" will then start
// hitting the right upstream API.
//
// Agents whose model resolves to nil keep their existing provider —
// better to log "stale provider" than to crash on a nil pointer mid-
// conversation.
func (m *Manager) Rewire() {
	if m.registry == nil {
		return
	}
	for _, ag := range m.agents {
		prov := m.registry.For(ag.Model())
		if prov == nil {
			slog.Warn("rewire: no provider matched agent model, keeping previous",
				"agent", ag.Name(),
				"model", ag.Model(),
				"available", m.registry.Names(),
			)
			continue
		}
		ag.setProvider(prov)
	}
}

// Registry exposes the routing registry so callers (gateway, tests)
// can inspect available providers.
func (m *Manager) Registry() *provider.Registry {
	return m.registry
}
