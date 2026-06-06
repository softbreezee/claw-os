package goal

// FoldUsage applies one model call's token counts to a goal, in place.
// Returns the delta added to TokensUsed and whether the goal just
// crossed its budget on this fold.
//
// Non-Active goals are skipped (Paused / BudgetLimited / Complete
// don't get billed). Cached input is excluded — the caller is
// expected to pass uncached input tokens (provider adapters already
// strip cache hits before we see the count).
//
// Mutates g rather than returning a copy because callers persist the
// same record they're folding into.
func FoldUsage(g *Goal, inputTokens, outputTokens int64) (delta int64, exhausted bool) {
	if g == nil || g.Status != StatusActive {
		return 0, false
	}
	if inputTokens < 0 {
		inputTokens = 0
	}
	if outputTokens < 0 {
		outputTokens = 0
	}
	delta = inputTokens + outputTokens
	if delta == 0 {
		return 0, false
	}
	g.TokensUsed += delta
	if g.TokenBudget != nil && g.TokensUsed >= *g.TokenBudget {
		g.Status = StatusBudgetLimited
		exhausted = true
	}
	return delta, exhausted
}
