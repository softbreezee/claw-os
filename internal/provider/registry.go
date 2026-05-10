package provider

import (
	"strings"
	"sync"
)

// Registry holds one Provider instance per configured provider name
// (matching the keys under config.Providers in fastclaw.json).
//
// Why this exists:
//   Earlier versions of fastclaw built a single global Provider from
//   the "default" entry. That meant every agent — regardless of its
//   model field — sent its requests to the same upstream API. As soon
//   as a user configured two providers (e.g. deepseek + openai/moonshot)
//   and pointed different agents at different models, every request
//   went to whichever provider happened to win the fallback dance,
//   and the rejected ones got a confusing "model not supported" 400.
//
//   Registry fixes this by keying off the "provider/" prefix on the
//   model string. A model named "openai/kimi-k2.6" is routed to the
//   Provider built from cfg.Providers["openai"]; "deepseek/v4-flash"
//   goes to cfg.Providers["deepseek"]; an unprefixed model falls back
//   to the default.
//
// Thread safety: callers hold the registry by *Registry pointer and
// can swap providers via Replace() during hot-reload. Lookups take
// a read lock and return a stable Provider reference; the agent then
// uses that reference for the duration of a single Chat call.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider // name -> Provider
	def       Provider            // fallback when prefix doesn't match
	defName   string              // for diagnostics
}

// NewRegistry constructs an empty registry. Use Set to register
// providers and SetDefault to pick the fallback.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Set registers (or replaces) a provider under the given name.
// Empty name is silently dropped to keep callers terse.
func (r *Registry) Set(name string, p Provider) {
	if name == "" || p == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[name] = p
}

// SetDefault marks one of the registered providers as the fallback
// for unprefixed model names. The name must already have been Set;
// otherwise the default is left unchanged.
func (r *Registry) SetDefault(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if p, ok := r.providers[name]; ok {
		r.def = p
		r.defName = name
	}
}

// For returns the Provider that should serve a request for the given
// model name. It strips the "provider/" prefix off the model when
// looking up the registry; the caller is responsible for stripping
// it again before sending the request upstream (Provider impls
// already call StripProviderPrefix in their buildRequest paths).
//
// Resolution order:
//  1. If model is "name/...": look up providers[name].
//  2. Otherwise (or if name not found): return the configured default.
//  3. If no default has been set: return any registered provider.
//  4. If the registry is empty: return nil.
func (r *Registry) For(model string) Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if idx := strings.Index(model, "/"); idx > 0 {
		name := model[:idx]
		if p, ok := r.providers[name]; ok {
			return p
		}
	}
	if r.def != nil {
		return r.def
	}
	// Last-resort: return any registered provider so a misconfigured
	// model name still gets *some* response (likely an upstream 4xx,
	// which is more debuggable than a nil-pointer panic).
	for _, p := range r.providers {
		return p
	}
	return nil
}

// Default returns the fallback provider, or nil if none was set.
func (r *Registry) Default() Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.def
}

// Names returns all registered provider names. Useful for diagnostics
// and for the "is this provider configured?" check used by the UI
// validator.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.providers))
	for n := range r.providers {
		names = append(names, n)
	}
	return names
}

// Has reports whether a provider with the given name is registered.
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.providers[name]
	return ok
}

// Replace atomically swaps the contents of the registry with the
// providers and default in `other`. Used during hot-reload so callers
// holding a *Registry pointer keep seeing updated routing without
// having to plumb new instances through every agent.
func (r *Registry) Replace(other *Registry) {
	other.mu.RLock()
	newProviders := make(map[string]Provider, len(other.providers))
	for k, v := range other.providers {
		newProviders[k] = v
	}
	newDef := other.def
	newDefName := other.defName
	other.mu.RUnlock()

	r.mu.Lock()
	r.providers = newProviders
	r.def = newDef
	r.defName = newDefName
	r.mu.Unlock()
}
