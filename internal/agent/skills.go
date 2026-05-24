package agent

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/softbreezee/claw-os/internal/config"
	"gopkg.in/yaml.v3"
)

// Skill represents a discovered skill.
type Skill struct {
	Name        string            // directory name
	Layer       string            // "agent", "user", "managed", "bundled", "extra"
	Content     string            // contents of SKILL.md (with {baseDir} replaced)
	BaseDir     string            // absolute path to the skill directory
	Description string            // from frontmatter
	Metadata    *SkillMetadata    // parsed OpenClaw metadata
	Gated       bool              // true if gating requirements not met
	GateReason  string            // reason gating failed
}

// SkillFrontmatter represents the YAML frontmatter of a SKILL.md file.
type SkillFrontmatter struct {
	Name        string       `yaml:"name"`
	Description string       `yaml:"description"`
	Homepage    string       `yaml:"homepage"`
	Metadata    yaml.Node    `yaml:"metadata"`
}

// SkillMetadata represents the skill metadata block.
// Prefers the "pawnix" key; falls back to "openclaw" since OpenClaw
// skills authored by upstream are commonly redistributed.
type SkillMetadata struct {
	Pawnix   *OpenClawMeta `json:"pawnix"`
	OpenClaw *OpenClawMeta `json:"openclaw"`
}

// Meta returns the effective metadata, preferring pawnix over openclaw.
func (m *SkillMetadata) Meta() *OpenClawMeta {
	if m.Pawnix != nil {
		return m.Pawnix
	}
	return m.OpenClaw
}

// OpenClawMeta holds OpenClaw-specific metadata.
type OpenClawMeta struct {
	Emoji      string           `json:"emoji"`
	Homepage   string           `json:"homepage"`
	Always     bool             `json:"always"`
	OS         []string         `json:"os"`
	Requires   *SkillRequires   `json:"requires"`
	PrimaryEnv string           `json:"primaryEnv"`
	Install    json.RawMessage  `json:"install"`
}

// SkillRequires holds gating requirements.
type SkillRequires struct {
	Bins    []string `json:"bins"`
	AnyBins []string `json:"anyBins"`
	Env     []string `json:"env"`
	Config  []string `json:"config"`
}

// SkillsLoader discovers and merges skills from multiple layers with OpenClaw compatibility.
type SkillsLoader struct {
	homeDir   string
	agentDir  string
	agentID   string // for ~/.pawnix/agents/<id>/agent/skills/ scoped skills
	teamDir   string
	skillsCfg config.SkillsConfig
	globalCfg config.SkillsCfg
}

// NewSkillsLoader creates a new skills loader.
func NewSkillsLoader(homeDir, agentDir, teamDir string, skillsCfg config.SkillsConfig) *SkillsLoader {
	return &SkillsLoader{
		homeDir:   homeDir,
		agentDir:  agentDir,
		teamDir:   teamDir,
		skillsCfg: skillsCfg,
	}
}

// NewSkillsLoaderWithGlobal creates a skills loader with global SkillsCfg for env injection and entries.
//
// agentID may be empty for callers that don't have it on hand; passing
// it enables a second per-agent skills source rooted at
// ~/.pawnix/agents/<agentID>/agent/skills/. That path is what the
// Web UI's "move skill to agent" flow targets, so it MUST be
// discoverable by the runtime — otherwise the agent silently loses
// access to skills the user thought were installed for it.
func NewSkillsLoaderWithGlobal(homeDir, agentDir, teamDir, agentID string, skillsCfg config.SkillsConfig, globalCfg config.SkillsCfg) *SkillsLoader {
	sl := NewSkillsLoader(homeDir, agentDir, teamDir, skillsCfg)
	sl.agentID = agentID
	sl.globalCfg = globalCfg
	return sl
}

// agentScopedSkillsDir returns the canonical per-agent skills location
// under ~/.pawnix/agents/<id>/agent/skills/, or "" when no agentID
// has been wired in (legacy callers).
func (sl *SkillsLoader) agentScopedSkillsDir() string {
	if sl.agentID == "" {
		return ""
	}
	return filepath.Join(sl.homeDir, "agents", sl.agentID, "agent", "skills")
}

// LoadSkills discovers skills from all layers and returns them merged.
// Precedence: agent workspace > user installed > managed > extra dirs.
func (sl *SkillsLoader) LoadSkills() []Skill {
	skills := make(map[string]Skill)

	disabled := make(map[string]bool, len(sl.skillsCfg.Disabled))
	for _, name := range sl.skillsCfg.Disabled {
		disabled[name] = true
	}
	// Also check global entries for enabled: false
	for name, entry := range sl.globalCfg.Entries {
		if !entry.Enabled {
			disabled[name] = true
		}
	}

	// Layer 5 (lowest): extra dirs from config
	for _, dir := range sl.globalCfg.Load.ExtraDirs {
		dir = expandPath(dir)
		for name, skill := range discoverSkillsEnhanced(dir, "extra") {
			if !disabled[name] {
				skills[name] = skill
			}
		}
	}

	// Layer 4: builtin skills shipped with the binary (the project's
	// skills/ directory). Loaded BEFORE user-installed so any user
	// skill of the same name overrides — that's the same precedence
	// the Web UI's Skills page documents and what users expect.
	if dir := builtinSkillsDir(); dir != "" {
		for name, skill := range discoverSkillsEnhanced(dir, "builtin") {
			if !disabled[name] {
				skills[name] = skill
			}
		}
	}

	// Layer 3: managed skills (~/.pawnix/managed-skills/)
	managedDir := pawnixManagedDir()
	for name, skill := range discoverSkillsEnhanced(managedDir, "managed") {
		if !disabled[name] {
			skills[name] = skill
		}
	}

	// Layer 2: user installed (~/.pawnix/skills/)
	userDir := filepath.Join(sl.homeDir, "skills")
	for name, skill := range discoverSkillsEnhanced(userDir, "user") {
		if !disabled[name] {
			skills[name] = skill
		}
	}

	// Layer 1.5: team skills
	if sl.teamDir != "" {
		teamSkillsDir := filepath.Join(sl.teamDir, "skills")
		for name, skill := range discoverSkillsEnhanced(teamSkillsDir, "team") {
			if !disabled[name] {
				skills[name] = skill
			}
		}
	}

	// Layer 1.25: per-agent scoped skills under ~/.pawnix/agents/<id>/agent/skills/.
	// This is where the Web UI's "move to agent" places skills, and
	// where the Skills page lists them from. Without this layer the
	// runtime silently ignores anything the user installed there —
	// they show up in the UI but the agent can't see them.
	if scoped := sl.agentScopedSkillsDir(); scoped != "" {
		for name, skill := range discoverSkillsEnhanced(scoped, "agent") {
			if !disabled[name] {
				skills[name] = skill
			}
		}
	}

	// Layer 1 (highest): agent workspace skills (i.e. the agent's own
	// working directory). Wins over the per-agent scoped layer above
	// because workspace edits should always trump installed defaults.
	agentSkillsDir := filepath.Join(sl.agentDir, "skills")
	for name, skill := range discoverSkillsEnhanced(agentSkillsDir, "agent") {
		if !disabled[name] {
			skills[name] = skill
		}
	}

	// Apply gating, agent filter, and env injection
	result := make([]Skill, 0, len(skills))
	for _, s := range skills {
		if s.Gated {
			slog.Debug("skill gated", "name", s.Name, "reason", s.GateReason)
			continue
		}
		// Filter by agent assignment: if the global config entry has non-empty
		// Agents, only load when this agentID is in the list.
		if sl.agentID != "" {
			if entry, ok := sl.globalCfg.Entries[s.Name]; ok && len(entry.Agents) > 0 {
				allowed := false
				for _, a := range entry.Agents {
					if a == sl.agentID {
						allowed = true
						break
					}
				}
				if !allowed {
					slog.Debug("skill filtered by agent", "skill", s.Name, "agent", sl.agentID, "assigned_to", entry.Agents)
					continue
				}
			}
		}
		result = append(result, s)
	}
	return result
}

// BuildSkillsSummary returns an XML summary of all skills for the system prompt.
func (sl *SkillsLoader) BuildSkillsSummary(skills []Skill) string {
	if len(skills) == 0 {
		return ""
	}

	alwaysLoad := make(map[string]bool, len(sl.skillsCfg.AlwaysLoad)+len(sl.globalCfg.AlwaysLoad))
	for _, name := range sl.skillsCfg.AlwaysLoad {
		alwaysLoad[name] = true
	}
	for _, name := range sl.globalCfg.AlwaysLoad {
		alwaysLoad[name] = true
	}

	var sb strings.Builder
	sb.WriteString("<skills>\n")

	for _, skill := range skills {
		if alwaysLoad[skill.Name] || (skill.Metadata != nil && skill.Metadata.Meta() != nil && skill.Metadata.Meta().Always) {
			fmt.Fprintf(&sb, "<skill name=%q layer=%q>\n%s\n</skill>\n", skill.Name, skill.Layer, skill.Content)
		} else {
			summary := skill.Description
			if summary == "" {
				summary = firstLine(skill.Content)
			}
			fmt.Fprintf(&sb, "<skill name=%q layer=%q summary=%q />\n", skill.Name, skill.Layer, summary)
		}
	}

	sb.WriteString("</skills>")
	return sb.String()
}

// SkillEnvVars returns environment variables for a specific skill from global config.
func (sl *SkillsLoader) SkillEnvVars(skillName string) map[string]string {
	entry, ok := sl.globalCfg.Entries[skillName]
	if !ok {
		return nil
	}
	env := make(map[string]string, len(entry.Env)+1)
	for k, v := range entry.Env {
		env[k] = v
	}
	// If apiKey is set and the skill has a primaryEnv, inject it
	if entry.APIKey != "" {
		// Find the skill to get primaryEnv
		// This is a convenience — the skill's primaryEnv tells us which env var the apiKey maps to
		for _, dir := range sl.allSkillDirs() {
			skillDir := filepath.Join(dir, skillName)
			fm := parseFrontmatter(filepath.Join(skillDir, "SKILL.md"))
			if fm != nil && fm.Metadata.Kind == yaml.MappingNode {
				meta := parseMetadata(&fm.Metadata)
				if meta != nil && meta.Meta() != nil && meta.Meta().PrimaryEnv != "" {
					env[meta.Meta().PrimaryEnv] = entry.APIKey
					break
				}
			}
		}
	}
	return env
}

// AllSkillDirs returns all skill directories in precedence order.
func (sl *SkillsLoader) AllSkillDirs() []string {
	return sl.allSkillDirs()
}

func (sl *SkillsLoader) allSkillDirs() []string {
	var dirs []string
	dirs = append(dirs, filepath.Join(sl.agentDir, "skills"))
	if scoped := sl.agentScopedSkillsDir(); scoped != "" {
		dirs = append(dirs, scoped)
	}
	if sl.teamDir != "" {
		dirs = append(dirs, filepath.Join(sl.teamDir, "skills"))
	}
	dirs = append(dirs, filepath.Join(sl.homeDir, "skills"))
	dirs = append(dirs, pawnixManagedDir())
	if b := builtinSkillsDir(); b != "" {
		dirs = append(dirs, b)
	}
	dirs = append(dirs, sl.globalCfg.Load.ExtraDirs...)
	return dirs
}

// discoverSkillsEnhanced scans a directory for skills. A "skill" can be:
//   * a subdirectory containing SKILL.md (the canonical layout, with
//     supporting files alongside), OR
//   * a top-level *.md file (single-file skill — convention used for
//     short, self-contained skills like tradingagents-ashare.md).
//
// Both forms are picked up because the Web UI's Skills page already
// shows them via scanSkillsDir; previously only the directory form
// reached the agent runtime, so agents couldn't see — let alone load
// — single-file skills the user clearly intended for them.
//
// In both cases we parse YAML frontmatter (when present), apply gating,
// and substitute {baseDir} in the content body.
func discoverSkillsEnhanced(dir string, layer string) map[string]Skill {
	result := make(map[string]Skill)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return result
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Canonical: <dir>/<name>/SKILL.md
			skillDir := filepath.Join(dir, entry.Name())
			skillFile := filepath.Join(skillDir, "SKILL.md")
			data, err := os.ReadFile(skillFile)
			if err != nil {
				continue
			}
			absDir, _ := filepath.Abs(skillDir)
			if sk, ok := buildSkillFromBytes(entry.Name(), layer, data, absDir); ok {
				result[sk.Name] = sk
			}
			continue
		}

		// Single-file: <dir>/<name>.md (skip README.md and dotfiles)
		name := entry.Name()
		lower := strings.ToLower(name)
		if !strings.HasSuffix(lower, ".md") || lower == "readme.md" || strings.HasPrefix(name, ".") {
			continue
		}
		skillPath := filepath.Join(dir, name)
		data, err := os.ReadFile(skillPath)
		if err != nil {
			continue
		}
		// For single-file skills, {baseDir} resolves to the parent dir
		// so any sibling assets (in the same skills/ folder) remain
		// addressable. The skill itself is the file, not a subdir.
		absParent, _ := filepath.Abs(dir)
		skillName := strings.TrimSuffix(name, filepath.Ext(name))
		if sk, ok := buildSkillFromBytes(skillName, layer, data, absParent); ok {
			result[sk.Name] = sk
		}
	}

	return result
}

// buildSkillFromBytes turns raw SKILL content into a Skill record.
// Shared between the directory and single-file discovery branches so
// frontmatter parsing / gating / {baseDir} substitution stay identical.
func buildSkillFromBytes(name, layer string, data []byte, baseDir string) (Skill, bool) {
	content := strings.TrimSpace(string(data))

	fm := parseFrontmatterFromBytes(data)
	var meta *SkillMetadata
	var desc string
	if fm != nil {
		desc = fm.Description
		if fm.Metadata.Kind == yaml.MappingNode {
			meta = parseMetadata(&fm.Metadata)
		}
	}

	content = strings.ReplaceAll(content, "{baseDir}", baseDir)

	gated, gateReason := checkGating(meta)

	return Skill{
		Name:        name,
		Layer:       layer,
		Content:     content,
		BaseDir:     baseDir,
		Description: desc,
		Metadata:    meta,
		Gated:       gated,
		GateReason:  gateReason,
	}, true
}

// parseFrontmatter reads and parses YAML frontmatter from a SKILL.md file path.
func parseFrontmatter(path string) *SkillFrontmatter {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return parseFrontmatterFromBytes(data)
}

// parseFrontmatterFromBytes parses YAML frontmatter from raw bytes.
func parseFrontmatterFromBytes(data []byte) *SkillFrontmatter {
	content := string(data)

	if !strings.HasPrefix(strings.TrimSpace(content), "---") {
		return nil
	}

	// Find opening and closing ---
	trimmed := strings.TrimSpace(content)
	rest := trimmed[3:] // skip first ---
	endIdx := strings.Index(rest, "\n---")
	if endIdx < 0 {
		return nil
	}

	fmStr := rest[:endIdx]

	var fm SkillFrontmatter
	if err := yaml.Unmarshal([]byte(fmStr), &fm); err != nil {
		return nil
	}
	return &fm
}

// parseMetadata converts the yaml.Node metadata into our SkillMetadata struct.
func parseMetadata(node *yaml.Node) *SkillMetadata {
	if node == nil || node.Kind == 0 {
		return nil
	}
	// Marshal back to YAML then decode as JSON-like structure
	yamlBytes, err := yaml.Marshal(node)
	if err != nil {
		return nil
	}

	// Unmarshal YAML into a generic map, then marshal to JSON, then unmarshal to struct
	var raw interface{}
	if err := yaml.Unmarshal(yamlBytes, &raw); err != nil {
		return nil
	}

	jsonBytes, err := json.Marshal(convertYAMLToJSON(raw))
	if err != nil {
		return nil
	}

	var meta SkillMetadata
	if err := json.Unmarshal(jsonBytes, &meta); err != nil {
		return nil
	}
	return &meta
}

// convertYAMLToJSON converts YAML map[string]interface{} (which uses map[interface{}]interface{})
// to JSON-compatible map[string]interface{}.
func convertYAMLToJSON(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{}, len(val))
		for k, v := range val {
			result[k] = convertYAMLToJSON(v)
		}
		return result
	case map[interface{}]interface{}:
		result := make(map[string]interface{}, len(val))
		for k, v := range val {
			result[fmt.Sprint(k)] = convertYAMLToJSON(v)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, v := range val {
			result[i] = convertYAMLToJSON(v)
		}
		return result
	default:
		return v
	}
}

// checkGating validates whether a skill's requirements are met.
// Returns (gated, reason). gated=true means the skill should be skipped.
func checkGating(meta *SkillMetadata) (bool, string) {
	if meta == nil || meta.Meta() == nil {
		return false, ""
	}
	oc := meta.Meta()

	if oc.Always {
		return false, ""
	}

	// Check OS requirement
	if len(oc.OS) > 0 {
		currentOS := runtime.GOOS
		found := false
		for _, os := range oc.OS {
			if os == currentOS {
				found = true
				break
			}
		}
		if !found {
			return true, fmt.Sprintf("requires OS %v, current is %s", oc.OS, currentOS)
		}
	}

	if oc.Requires == nil {
		return false, ""
	}

	// Check required binaries
	for _, bin := range oc.Requires.Bins {
		if _, err := exec.LookPath(bin); err != nil {
			return true, fmt.Sprintf("required binary %q not found on PATH", bin)
		}
	}

	// Check anyBins (at least one must exist)
	if len(oc.Requires.AnyBins) > 0 {
		found := false
		for _, bin := range oc.Requires.AnyBins {
			if _, err := exec.LookPath(bin); err == nil {
				found = true
				break
			}
		}
		if !found {
			return true, fmt.Sprintf("none of required binaries %v found on PATH", oc.Requires.AnyBins)
		}
	}

	// Check required env vars
	for _, envVar := range oc.Requires.Env {
		if os.Getenv(envVar) == "" {
			return true, fmt.Sprintf("required env var %q not set", envVar)
		}
	}

	return false, ""
}

// pawnixManagedDir returns the Pawnix managed skills directory (~/.pawnix/managed-skills/).
// This is separate from user-installed skills (~/.pawnix/skills/) to avoid layer collision.
func pawnixManagedDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".pawnix", "managed-skills")
}

// builtinSkillsDir locates the project-shipped skills/ directory by
// probing common positions relative to the running binary. Returns ""
// if we can't find it (binary moved standalone, etc).
//
// Mirrors the same probe that internal/setup uses for the Web UI's
// Skills page, so the runtime view and the UI view stay in sync.
//
// Probe order:
//  1. {exeDir}/skills/      – installed layout (binary + skills side-by-side)
//  2. {exeDir}/../skills/   – go run / dev (bin/pawnix + ../skills/)
//  3. ./skills/             – CWD fallback
func builtinSkillsDir() string {
	candidates := []string{}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "skills"),
			filepath.Join(exeDir, "..", "skills"),
		)
	}
	candidates = append(candidates, "skills")
	for _, p := range candidates {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			return p
		}
	}
	return ""
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

func firstLine(s string) string {
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		return s[:idx]
	}
	if len(s) > 120 {
		return s[:120] + "..."
	}
	return s
}

// FindSkillForPath returns the skill name if the given path is within a skill directory.
func FindSkillForPath(path string, skillDirs []string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return ""
	}
	for _, dir := range skillDirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			continue
		}
		if strings.HasPrefix(absPath, absDir+string(filepath.Separator)) {
			// Extract skill name (first component after the skills dir)
			rel, err := filepath.Rel(absDir, absPath)
			if err != nil {
				continue
			}
			parts := strings.SplitN(rel, string(filepath.Separator), 2)
			if len(parts) > 0 {
				return parts[0]
			}
		}
	}
	return ""
}
