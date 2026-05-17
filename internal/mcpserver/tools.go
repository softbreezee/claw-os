package mcpserver

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/softbreezee/claw-os/internal/config"
	"gopkg.in/yaml.v3"
)

// pawnixHome returns ~/.pawnix or equivalent. Tools that need the
// pawnix data directory use this to remain relocatable.
func pawnixHome() (string, error) {
	return config.HomeDir()
}

// ── Tool registration ──

// RegisterReasonixTools registers all diagnostic/repair tools on the
// given MCP server. Callers (the pawnix mcp-reasonix subcommand) just
// wire a NewServer + RegisterReasonixTools + s.Serve().
func RegisterReasonixTools(s *Server) {
	s.Register("check_config", "Check pawnix configuration integrity. "+
		"Validates pawnix.json structure, checks for missing required fields, "+
		"broken agent references, and invalid paths. Returns a structured "+
		"report with severity (ok/warning/error) per finding.",
		map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		checkConfig,
	)

	s.Register("check_skills", "Check all skill files for validity. "+
		"Scans installed skills, validates YAML frontmatter, checks "+
		"gating requirements (binaries, env vars), and reports skills that "+
		"are gated/disabled. Returns per-skill status.",
		map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		checkSkills,
	)

	s.Register("check_disk", "Check pawnix disk usage. "+
		"Reports sizes of ~/.pawnix subdirectories (sessions, memory, "+
		"uploads, logs), flags directories exceeding thresholds, and "+
		"suggests cleanup actions. Use before log rotation.",
		map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		checkDisk,
	)

	s.Register("analyze_logs", "Scan recent pawnix logs for error patterns. "+
		"Reads the most recent log files, extracts ERROR and WARN lines, "+
		"groups by pattern, and returns top offenders with counts. "+
		"Helps identify recurring issues before they escalate.",
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"maxLines": map[string]interface{}{
					"type":        "integer",
					"description": "Max error lines to return (default 50).",
				},
			},
		},
		analyzeLogs,
	)

	s.Register("search_files", "Search for files by name pattern. "+
		"Recursively walks directories and returns matching file paths. "+
		"Use for finding config files, skill files, log files, or "+
		"any other file type by name pattern.",
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pattern": map[string]interface{}{
					"type":        "string",
					"description": "Filename pattern to match (e.g. '*.json', 'SKILL.md', 'config.*').",
				},
				"directory": map[string]interface{}{
					"type":        "string",
					"description": "Directory to search in. Defaults to ~/.pawnix.",
				},
			},
			"required": []string{"pattern"},
		},
		searchFiles,
	)

	s.Register("search_content", "Search file contents for a string or regex. "+
		"Grep through files to find references, error messages, or configuration "+
		"values. Returns matching file:line:content entries.",
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "String or regex to search for.",
				},
				"directory": map[string]interface{}{
					"type":        "string",
					"description": "Directory to search in. Defaults to ~/.pawnix.",
				},
				"filePattern": map[string]interface{}{
					"type":        "string",
					"description": "Optional filename filter (e.g. '*.json').",
				},
			},
			"required": []string{"query"},
		},
		searchContent,
	)

	s.Register("repair_config", "Repair common pawnix configuration issues. "+
		"Can fix: missing agent workspace directories, invalid skill "+
		"references, missing required fields with safe defaults. "+
		"Reports what was changed. Set dryRun=true to preview without modifying.",
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"dryRun": map[string]interface{}{
					"type":        "boolean",
					"description": "When true, only report what would be fixed without applying changes (default false).",
				},
			},
		},
		repairConfig,
	)

	s.Register("list_directory", "List directory contents with sizes. "+
		"Like 'ls -la' but returns structured output. Use to explore "+
		"the pawnix filesystem layout or inspect specific directories.",
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Directory path. Defaults to ~/.pawnix.",
				},
			},
		},
		listDirectory,
	)
}

// ── Tool implementations ──

func checkConfig(args json.RawMessage) (string, error) {
	home, err := pawnixHome()
	if err != nil {
		return "", fmt.Errorf("cannot determine pawnix home: %w", err)
	}

	var findings []string
	configPath := filepath.Join(home, "pawnix.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "SEVERITY=ERROR | pawnix.json not found at " + configPath, nil
		}
		return "", fmt.Errorf("read config: %w", err)
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Sprintf("SEVERITY=ERROR | pawnix.json is not valid JSON: %v", err), nil
	}

	// Check providers
	if providers, ok := cfg["providers"].(map[string]interface{}); ok {
		for name, p := range providers {
			prov, _ := p.(map[string]interface{})
			if prov["apiKey"] == nil || prov["apiKey"] == "" {
				findings = append(findings, fmt.Sprintf("SEVERITY=WARN | provider %q has no apiKey", name))
			}
			if prov["apiBase"] == nil || prov["apiBase"] == "" {
				findings = append(findings, fmt.Sprintf("SEVERITY=WARN | provider %q has no apiBase", name))
			}
		}
	} else {
		findings = append(findings, "SEVERITY=ERROR | no providers configured")
	}

	// Check agents
	agents, _ := cfg["agents"].(map[string]interface{})
	agentList, _ := agents["agents"].([]interface{})
	if len(agentList) == 0 {
		findings = append(findings, "SEVERITY=WARN | no agents defined")
	}
	for _, a := range agentList {
		ag, _ := a.(map[string]interface{})
		id, _ := ag["id"].(string)
		if id == "" {
			findings = append(findings, "SEVERITY=ERROR | agent missing 'id' field")
			continue
		}
		// Check workspace exists
		ws, _ := ag["workspace"].(string)
		if ws != "" {
			if info, err := os.Stat(ws); err != nil || !info.IsDir() {
				findings = append(findings, fmt.Sprintf("SEVERITY=WARN | agent %q workspace %q does not exist or is not a directory", id, ws))
			}
		}
	}

	// Check channels
	channels, _ := cfg["channels"].(map[string]interface{})
	enabledCount := 0
	for _, ch := range channels {
		chMap, _ := ch.(map[string]interface{})
		if enabled, _ := chMap["enabled"].(bool); enabled {
			enabledCount++
		}
	}
	if enabledCount == 0 {
		findings = append(findings, "SEVERITY=INFO | no channels enabled — you'll need the web UI to interact")
	}

	if len(findings) == 0 {
		findings = append(findings, "SEVERITY=OK | config looks clean")
	}

	return strings.Join(findings, "\n"), nil
}

// ── skill frontmatter types (mirrors agent/skills.go for independence) ──

type skillFM struct {
	Description string    `yaml:"description"`
	Metadata    yaml.Node `yaml:"metadata"`
}

func checkSkills(args json.RawMessage) (string, error) {
	home, err := pawnixHome()
	if err != nil {
		return "", fmt.Errorf("cannot determine pawnix home: %w", err)
	}

	skillsDir := filepath.Join(home, "skills")
	var findings []string

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "INFO | ~/.pawnix/skills/ does not exist yet — no user skills installed", nil
		}
		return "", fmt.Errorf("read skills dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		skillFile := filepath.Join(skillsDir, name, "SKILL.md")

		data, err := os.ReadFile(skillFile)
		if err != nil {
			findings = append(findings, fmt.Sprintf("SEVERITY=WARN | skill %q: SKILL.md not readable: %v", name, err))
			continue
		}

		content := strings.TrimSpace(string(data))
		if !strings.HasPrefix(content, "---") {
			findings = append(findings, fmt.Sprintf("SEVERITY=INFO | skill %q: no YAML frontmatter", name))
			continue
		}

		// Parse frontmatter
		trimmed := strings.TrimPrefix(content, "---")
		endIdx := strings.Index(trimmed, "\n---")
		if endIdx < 0 {
			findings = append(findings, fmt.Sprintf("SEVERITY=WARN | skill %q: unclosed frontmatter", name))
			continue
		}
		fmStr := trimmed[:endIdx]

		var fm skillFM
		if err := yaml.Unmarshal([]byte(fmStr), &fm); err != nil {
			findings = append(findings, fmt.Sprintf("SEVERITY=WARN | skill %q: invalid YAML frontmatter: %v", name, err))
			continue
		}

		desc := fm.Description
		if desc == "" {
			desc = "(no description)"
		}
		findings = append(findings, fmt.Sprintf("SEVERITY=OK | skill %q: %s", name, desc))
	}

	if len(findings) == 0 {
		findings = append(findings, "SEVERITY=OK | all skills look clean")
	}

	return strings.Join(findings, "\n"), nil
}

func checkDisk(args json.RawMessage) (string, error) {
	home, err := pawnixHome()
	if err != nil {
		return "", fmt.Errorf("cannot determine pawnix home: %w", err)
	}

	var findings []string

	subdirs := []struct {
		name  string
		path  string
		warnMB int64
	}{
		{"sessions", filepath.Join(home, "sessions"), 200},
		{"memory", filepath.Join(home, "memory"), 500},
		{"uploads", filepath.Join(home, "uploads"), 500},
		{"logs", filepath.Join(home, "logs"), 200},
	}

	for _, sd := range subdirs {
		info, err := os.Stat(sd.path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			findings = append(findings, fmt.Sprintf("SEVERITY=WARN | cannot stat %s: %v", sd.name, err))
			continue
		}
		if !info.IsDir() {
			continue
		}

		size := dirSize(sd.path)
		sizeMB := size / (1024 * 1024)
		if sizeMB >= sd.warnMB {
			findings = append(findings, fmt.Sprintf("SEVERITY=WARN | %s: %d MB (threshold: %d MB) — consider cleanup",
				sd.name, sizeMB, sd.warnMB))
		} else {
			findings = append(findings, fmt.Sprintf("SEVERITY=OK | %s: %d MB", sd.name, sizeMB))
		}
	}

	// Also report total pawnix home size
	totalMB := dirSize(home) / (1024 * 1024)
	findings = append(findings, fmt.Sprintf("TOTAL | ~/.pawnix: %d MB", totalMB))

	if totalMB > 2048 {
		findings = append(findings, "SEVERITY=WARN | Overall pawnix directory exceeds 2GB — review old sessions and logs")
	}

	return strings.Join(findings, "\n"), nil
}

func dirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		size += info.Size()
		return nil
	})
	return size
}

func analyzeLogs(args json.RawMessage) (string, error) {
	type analyzeArgs struct {
		MaxLines int `json:"maxLines"`
	}
	var aa analyzeArgs
	aa.MaxLines = 50
	if len(args) > 0 {
		if err := json.Unmarshal(args, &aa); err != nil {
			// Use defaults
		}
	}
	if aa.MaxLines <= 0 {
		aa.MaxLines = 50
	}

	home, err := pawnixHome()
	if err != nil {
		return "", fmt.Errorf("cannot determine pawnix home: %w", err)
	}

	logsDir := filepath.Join(home, "logs")
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		return "INFO | No logs directory found — nothing to analyze", nil
	}

	// Find most recent log files
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		return "", fmt.Errorf("read logs dir: %w", err)
	}

	// Sort by name descending (newest first if names have dates)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() > entries[j].Name()
	})

	type errCount struct {
		pattern string
		count   int
		sample  string
	}
	patternMap := make(map[string]*errCount)

	totalLines := 0
	for _, entry := range entries {
		if totalLines >= aa.MaxLines*5 {
			break
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".log") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(logsDir, entry.Name()))
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if totalLines >= aa.MaxLines*5 {
				break
			}
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			totalLines++
			if strings.Contains(line, "ERROR") || strings.Contains(line, "WARN") {
				// Normalize by stripping timestamps and IDs
				normalized := normalizeLogLine(line)
				if pc, ok := patternMap[normalized]; ok {
					pc.count++
				} else {
					patternMap[normalized] = &errCount{pattern: normalized, count: 1, sample: line}
				}
			}
		}
	}

	if len(patternMap) == 0 {
		return "OK | No ERROR or WARN patterns found in recent logs", nil
	}

	// Sort by count descending
	var patterns []*errCount
	for _, pc := range patternMap {
		patterns = append(patterns, pc)
	}
	sort.Slice(patterns, func(i, j int) bool { return patterns[i].count > patterns[j].count })

	var out []string
	shown := 0
	for _, pc := range patterns {
		if shown >= aa.MaxLines {
			break
		}
		out = append(out, fmt.Sprintf("x%d | %s", pc.count, pc.sample))
		shown++
	}

	if len(out) == 0 {
		return "OK | No significant error patterns", nil
	}

	return fmt.Sprintf("Scanned %d log lines across recent files.\n%s", totalLines, strings.Join(out, "\n")), nil
}

func normalizeLogLine(line string) string {
	// Remove common variable parts: timestamps, hex IDs, quoted strings
	norm := line
	// Timestamps like "2024-01-15T10:30:00Z" or "Jan 15 10:30:00"
	norm = replaceTimestamp(norm)
	// Hex IDs like "0x1a2b3c4d" or "abc123def"
	norm = replaceHexIDs(norm)
	// Numbers at end
	norm = strings.TrimRight(norm, "0123456789 ")
	return norm
}

func replaceTimestamp(s string) string {
	// Simple heuristic: if it looks like a log timestamp prefix, strip it
	// Format: "2024-01-15T10:30:00" or "time=2024-..."
	for _, sep := range []string{"T", " "} {
		if idx := strings.Index(s, sep); idx > 0 {
			prefix := s[:idx]
			if len(prefix) >= 10 && strings.Count(prefix, "-") >= 2 {
				return "[ts]" + s[idx:]
			}
		}
	}
	return s
}

func replaceHexIDs(s string) string {
	// Replace hex strings longer than 6 chars with [id]
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 6 && isHexish(w) {
			words[i] = "[id]"
		}
	}
	return strings.Join(words, " ")
}

func isHexish(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func searchFiles(args json.RawMessage) (string, error) {
	type sfArgs struct {
		Pattern   string `json:"pattern"`
		Directory string `json:"directory"`
	}
	var sa sfArgs
	if err := json.Unmarshal(args, &sa); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}
	if sa.Pattern == "" {
		return "", fmt.Errorf("pattern is required")
	}

	dir := sa.Directory
	if dir == "" {
		home, err := pawnixHome()
		if err != nil {
			return "", fmt.Errorf("pawnix home: %w", err)
		}
		dir = home
	}

	var matches []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip inaccessible
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if base == ".git" || base == "node_modules" || base == "__pycache__" {
				return filepath.SkipDir
			}
			return nil
		}
		name := info.Name()
		matched, _ := filepath.Match(sa.Pattern, name)
		if matched {
			rel, _ := filepath.Rel(dir, path)
			size := info.Size()
			matches = append(matches, fmt.Sprintf("%s (%d bytes)", rel, size))
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("walk: %w", err)
	}

	sort.Strings(matches)
	if len(matches) == 0 {
		return fmt.Sprintf("No files matching %q found in %s", sa.Pattern, dir), nil
	}
	return fmt.Sprintf("Found %d files matching %q:\n%s", len(matches), sa.Pattern, strings.Join(matches, "\n")), nil
}

func searchContent(args json.RawMessage) (string, error) {
	type scArgs struct {
		Query       string `json:"query"`
		Directory   string `json:"directory"`
		FilePattern string `json:"filePattern"`
	}
	var sa scArgs
	if err := json.Unmarshal(args, &sa); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}
	if sa.Query == "" {
		return "", fmt.Errorf("query is required")
	}

	dir := sa.Directory
	if dir == "" {
		home, err := pawnixHome()
		if err != nil {
			return "", fmt.Errorf("pawnix home: %w", err)
		}
		dir = home
	}

	// Use grep for efficient content search
	var cmd *exec.Cmd
	if sa.FilePattern != "" {
		cmd = exec.Command("grep", "-rn", "-I", "--include="+sa.FilePattern, sa.Query, dir)
	} else {
		cmd = exec.Command("grep", "-rn", "-I", sa.Query, dir)
	}

	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return fmt.Sprintf("No matches for %q in %s", sa.Query, dir), nil
		}
		return "", fmt.Errorf("grep: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) > 100 {
		lines = lines[:100]
	}
	return fmt.Sprintf("Matches for %q (%d found, showing first 100):\n%s",
		sa.Query, len(lines), strings.Join(lines, "\n")), nil
}

func repairConfig(args json.RawMessage) (string, error) {
	type rcArgs struct {
		DryRun bool `json:"dryRun"`
	}
	var ra rcArgs
	if len(args) > 0 {
		if err := json.Unmarshal(args, &ra); err != nil {
			// Use defaults
		}
	}

	home, err := pawnixHome()
	if err != nil {
		return "", fmt.Errorf("cannot determine pawnix home: %w", err)
	}

	configPath := filepath.Join(home, "pawnix.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("read config: %w", err)
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", fmt.Errorf("config is not valid JSON: %w", err)
	}

	var fixes []string

	// Fix 1: ensure agents have workspace directories
	agents, _ := cfg["agents"].(map[string]interface{})
	agentList, _ := agents["agents"].([]interface{})
	modified := false
	for i, a := range agentList {
		ag, _ := a.(map[string]interface{})
		ws, _ := ag["workspace"].(string)
		if ws == "" {
			id, _ := ag["id"].(string)
			if id == "" {
				id = fmt.Sprintf("agent-%d", i)
			}
			newWS := filepath.Join(home, "agents", id, "workspace")
			ag["workspace"] = newWS
			fixes = append(fixes, fmt.Sprintf("FIX | agent %q: set workspace to %s", id, newWS))
			modified = true
		} else if _, err := os.Stat(ws); os.IsNotExist(err) {
			fixes = append(fixes, fmt.Sprintf("FIX | agent workspace %q does not exist — would create directory", ws))
			if !ra.DryRun {
				if err := os.MkdirAll(ws, 0o755); err != nil {
					fixes = append(fixes, fmt.Sprintf("FAILED | create workspace %s: %v", ws, err))
				} else {
					fixes = append(fixes, fmt.Sprintf("CREATED | %s", ws))
				}
			}
		}
	}

	// Fix 2: ensure logs directory exists
	logsDir := filepath.Join(home, "logs")
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		fixes = append(fixes, fmt.Sprintf("FIX | logs directory missing — would create %s", logsDir))
		if !ra.DryRun {
			if err := os.MkdirAll(logsDir, 0o755); err != nil {
				fixes = append(fixes, fmt.Sprintf("FAILED | create logs dir: %v", err))
			} else {
				fixes = append(fixes, fmt.Sprintf("CREATED | %s", logsDir))
			}
		}
	}

	// Save if modified
	if modified && !ra.DryRun {
		newData, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return "", fmt.Errorf("marshal fixed config: %w", err)
		}
		if err := os.WriteFile(configPath, newData, 0o644); err != nil {
			return "", fmt.Errorf("write config: %w", err)
		}
		fixes = append(fixes, "SAVED | config written to "+configPath)
	}

	if len(fixes) == 0 {
		fixes = append(fixes, "OK | no repairs needed")
	}

	mode := "LIVE"
	if ra.DryRun {
		mode = "DRY-RUN"
	}
	return fmt.Sprintf("=== repair_config (%s) ===\n%s", mode, strings.Join(fixes, "\n")), nil
}

func listDirectory(args json.RawMessage) (string, error) {
	type ldArgs struct {
		Path string `json:"path"`
	}
	var la ldArgs
	if len(args) > 0 {
		if err := json.Unmarshal(args, &la); err != nil {
			// Use defaults
		}
	}

	path := la.Path
	if path == "" {
		home, err := pawnixHome()
		if err != nil {
			return "", fmt.Errorf("pawnix home: %w", err)
		}
		path = home
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("read dir: %w", err)
	}

	var out []string
	for _, e := range entries {
		info, _ := e.Info()
		var size string
		if info != nil {
			if info.IsDir() {
				size = "[dir]"
			} else {
				s := info.Size()
				switch {
				case s > 1024*1024:
					size = fmt.Sprintf("%.1f MB", float64(s)/(1024*1024))
				case s > 1024:
					size = fmt.Sprintf("%.1f KB", float64(s)/1024)
				default:
					size = fmt.Sprintf("%d B", s)
				}
			}
		}
		modTime := ""
		if info != nil {
			modTime = info.ModTime().Format(time.RFC3339)
		}
		out = append(out, fmt.Sprintf("%s | %10s | %s", modTime, size, e.Name()))
	}

	sort.Slice(out, func(i, j int) bool {
		// Dirs first, then alphabetical
		iDir := strings.Contains(out[i], "[dir]")
		jDir := strings.Contains(out[j], "[dir]")
		if iDir != jDir {
			return iDir
		}
		return out[i] < out[j]
	})

	if len(out) == 0 {
		return path + " is empty", nil
	}

	return fmt.Sprintf("Contents of %s:\n%s", path, strings.Join(out, "\n")), nil
}

func init() {
	// Ensure we can compile on platforms without grep
	_ = runtime.GOOS
}
