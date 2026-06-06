package bus

import "testing"

// TestOriginUserIsEmpty pins the backwards-compat contract: every
// existing producer (channel adapters, web POST, plugins) creates
// InboundMessage with Origin unset. If OriginUser ever becomes a
// non-empty string, the zero-value InboundMessage silently
// reclassifies as a runtime-injected message and Notification
// routing breaks across the gateway and agent loop.
func TestOriginUserIsEmpty(t *testing.T) {
	if OriginUser != "" {
		t.Fatalf(`OriginUser must be "" for zero-value compatibility, got %q`, OriginUser)
	}
	if OriginChannel != "" {
		t.Fatalf(`OriginChannel alias must equal OriginUser (""), got %q`, OriginChannel)
	}
	var m InboundMessage
	if m.Origin != OriginUser {
		t.Fatalf("zero-value InboundMessage.Origin should equal OriginUser, got %q", m.Origin)
	}
}

// TestOriginConstantsDistinct catches the typo class of bug where two
// constants accidentally collide; routing branches that depend on
// distinguishing origins would silently merge.
func TestOriginConstantsDistinct(t *testing.T) {
	cases := map[string]string{
		// OriginChannel is intentionally an alias of OriginUser, so we
		// don't include it in the distinctness check.
		"OriginUser":        OriginUser,
		"OriginCron":        OriginCron,
		"OriginWebhook":     OriginWebhook,
		"OriginInternal":    OriginInternal,
		"OriginHeartbeat":   OriginHeartbeat,
		"OriginSubAgent":    OriginSubAgent,
		"OriginGoalContext": OriginGoalContext,
		"OriginUserSteer":   OriginUserSteer,
	}
	seen := map[string]string{}
	for name, val := range cases {
		if prev, ok := seen[val]; ok {
			t.Errorf("%s and %s both equal %q — Origin constants must be distinct", prev, name, val)
		}
		seen[val] = name
	}
}

// TestIsRuntimeInjected pins the policy decision the gateway and
// agent loop both consult. Goal continuations / steer / heartbeat /
// subagent are NOT considered runtime-injected for routing purposes
// because they ARE part of the user's conversation thread; only
// genuinely out-of-band sources (cron / webhook / internal) trigger
// the Notification path.
func TestIsRuntimeInjected(t *testing.T) {
	runtime := []string{OriginCron, OriginWebhook, OriginInternal}
	for _, o := range runtime {
		if !IsRuntimeInjected(o) {
			t.Errorf("IsRuntimeInjected(%q) = false, want true", o)
		}
	}
	notRuntime := []string{
		OriginUser, OriginChannel, OriginHeartbeat,
		OriginSubAgent, OriginGoalContext, OriginUserSteer,
		"unknown-future-value",
	}
	for _, o := range notRuntime {
		if IsRuntimeInjected(o) {
			t.Errorf("IsRuntimeInjected(%q) = true, want false", o)
		}
	}
}
