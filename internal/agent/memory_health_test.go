package agent

import "testing"

// TestRecentCoveragePct pins the contract Heartbeat consumes:
//   - empty window returns -1 ("not applicable" rather than 0%)
//   - full coverage returns 100
//   - partial returns the integer percentage (truncated, not rounded)
func TestRecentCoveragePct(t *testing.T) {
	cases := []struct {
		name string
		in   MemoryHealthSnapshot
		want int
	}{
		{"empty window", MemoryHealthSnapshot{}, -1},
		{"all embedded", MemoryHealthSnapshot{RecentTotal: 5, RecentEmbedded: 5}, 100},
		{"none embedded", MemoryHealthSnapshot{RecentTotal: 5, RecentEmbedded: 0}, 0},
		{"partial truncates", MemoryHealthSnapshot{RecentTotal: 7, RecentEmbedded: 5}, 71},
		// Total-only rows must not crash RecentCoveragePct: the
		// 24h window is a strict subset of total, so RecentTotal
		// can be 0 even while Total > 0.
		{"old data only", MemoryHealthSnapshot{Total: 100, WithEmbedding: 100, RecentTotal: 0}, -1},
	}
	for _, c := range cases {
		got := c.in.RecentCoveragePct()
		if got != c.want {
			t.Errorf("%s: got %d, want %d", c.name, got, c.want)
		}
	}
}
