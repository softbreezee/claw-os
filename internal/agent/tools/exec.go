package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fastclaw-ai/fastclaw/internal/sandbox"
)

type execArgs struct {
	Command string `json:"command"`
	Timeout int    `json:"timeout,omitempty"` // seconds, default 30
	Sandbox bool   `json:"sandbox,omitempty"` // force sandbox for this call
}

const (
	maxExecOutputBytes = 128 * 1024 // 128 KB cap on tool output sent to LLM
	maxExecTimeoutSec  = 300        // hard ceiling regardless of what the LLM requests
)

// dangerousPatterns are regex-like substrings whose presence blocks execution
// when sandbox is not enabled. The list is intentionally broad.
var dangerousPatterns = []string{
	// filesystem destruction
	"rm -rf /",
	"rm -rf ~",
	"rm -rf $home",
	"rm -rf ${home}",
	"mkfs",
	"dd if=",
	"> /dev/sd",
	"> /dev/nvme",
	// fork bombs
	":(){:|:&};:",
	":(){ :|:",
	// privilege escalation helpers
	"chmod 777 /etc",
	"chmod 777 /usr",
	"chown -r root /",
	"chown -r root /etc",
	// credential / shadow exfiltration
	"/etc/shadow",
	"/etc/passwd",
	"/etc/sudoers",
}

// SandboxConfig holds sandbox settings passed to the exec tool registration.
type SandboxConfig struct {
	Enabled   bool
	Image     string
	Pool      *sandbox.SandboxPool
	Workspace string
	AgentID   string
	Policy    *sandbox.Policy
}

// SkillEnvProvider returns environment variables for a skill by name.
type SkillEnvProvider func(skillName string) map[string]string

func registerExec(r *Registry, workspace string) {
	registerExecWithSandbox(r, nil, workspace)
}

func registerExecWithSandbox(r *Registry, sbCfg *SandboxConfig, workspace string) {
	registerExecFull(r, sbCfg, nil, nil, workspace)
}

// RegisterExecWithSkillEnv registers the exec tool with skill environment
// injection support. workspace is the agent's root directory and becomes
// the cwd of every `sh -c ...` we spawn — so a tool call like
// `wget https://x.zip` lands in the agent's own directory instead of
// the daemon's startup cwd. Sandbox mode already pins cwd via the
// container, so this only affects the no-sandbox path.
func RegisterExecWithSkillEnv(r *Registry, sbCfg *SandboxConfig, envProvider SkillEnvProvider, skillDirs []string, workspace string) {
	registerExecFull(r, sbCfg, envProvider, skillDirs, workspace)
}

func registerExecFull(r *Registry, sbCfg *SandboxConfig, envProvider SkillEnvProvider, skillDirs []string, workspace string) {
	r.Register("exec", "Execute a shell command and return stdout/stderr", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The shell command to execute",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Timeout in seconds (default 30)",
			},
			"sandbox": map[string]interface{}{
				"type":        "boolean",
				"description": "Force execution in sandbox container",
			},
		},
		"required": []string{"command"},
	}, makeExecToolFull(sbCfg, envProvider, skillDirs, workspace))
}

func makeExecTool(sbCfg *SandboxConfig) ToolFunc {
	return makeExecToolFull(sbCfg, nil, nil, "")
}

// workspace is set to the agent's root directory at construction. When
// non-empty AND we're not running through the sandbox, it becomes
// cmd.Dir for every `sh -c ...` so all relative paths produced by the
// LLM (downloads, cloned repos, scratch files, etc) anchor here instead
// of inheriting the daemon's startup cwd. Empty workspace falls back to
// process inherited cwd to preserve callers that didn't provide one
// (notably the makeExecTool helper used by older registrations).
func makeExecToolFull(sbCfg *SandboxConfig, envProvider SkillEnvProvider, skillDirs []string, workspace string) ToolFunc {
	return func(ctx context.Context, rawArgs json.RawMessage) (string, error) {
		var args execArgs
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return "", fmt.Errorf("parse args: %w", err)
		}

		if args.Command == "" {
			return "", fmt.Errorf("command is required")
		}

		// Block dangerous patterns when not running inside a sandbox.
		useSandbox := args.Sandbox || (sbCfg != nil && sbCfg.Enabled)
		if !useSandbox {
			lower := strings.ToLower(args.Command)
			for _, pat := range dangerousPatterns {
				if strings.Contains(lower, pat) {
					return "", fmt.Errorf("dangerous command blocked (enable sandbox to override): matched pattern %q", pat)
				}
			}
		}

		timeout := 30
		if args.Timeout > 0 {
			timeout = args.Timeout
		}
		if timeout > maxExecTimeoutSec {
			timeout = maxExecTimeoutSec
		}

		execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()

		if useSandbox && sbCfg != nil && sbCfg.Pool != nil {
			sb := sbCfg.Pool.Get(sbCfg.AgentID, sbCfg.Image, sbCfg.Workspace, sbCfg.Policy)
			out, err := sb.Exec(execCtx, args.Command, "/workspace")
			return truncateOutput(out), err
		}

		cmd := exec.CommandContext(execCtx, "sh", "-c", args.Command)

		// Anchor cwd to the agent's workspace so wget / git clone / mkdir
		// land where the LLM (and the file tools) expect them, not in
		// whatever directory the gateway daemon was started from. We
		// best-effort MkdirAll first because a freshly-configured
		// workspace dir might not exist on disk yet.
		if workspace != "" {
			_ = os.MkdirAll(workspace, 0o755)
			cmd.Dir = workspace
		}

		// Inject skill-specific env vars if the command references a skill directory
		if envProvider != nil && skillDirs != nil {
			skillEnv := resolveSkillEnv(args.Command, envProvider, skillDirs)
			if len(skillEnv) > 0 {
				cmd.Env = mergeEnv(os.Environ(), skillEnv)
			}
		}

		// Detach the child into its own process group (unix only) so
		// that cancellation can SIGKILL the whole tree, not just the
		// `sh -c …` parent. This matters when the LLM produces commands
		// like `nohup python -m http.server &` that fork grandchildren
		// holding the inherited stdout / stderr pipes — without group-
		// kill those grandchildren outlive the timeout and keep the
		// pipes open.
		setProcessGroup(cmd)

		// WaitDelay is the cleanest fix for the canonical "exec hangs
		// on backgrounded subprocess" bug:
		//
		//   nohup python3 -m http.server 8899 > /tmp/x.log 2>&1 &
		//
		// `sh -c …` exits immediately, but the spawned python inherits
		// the cmd.Stdout / cmd.Stderr pipe fds (the file redirect only
		// rebinds python's own stdout — the inherited fastclaw-side
		// pipe fd survives because Go reuses it). cmd.CombinedOutput
		// then waits on EOF that never comes and the agent loop hangs
		// indefinitely. WaitDelay tells exec.Cmd to force-close those
		// pipes 2s after the context is done (or after Wait completes
		// with pipes still open), so we always return — at worst with
		// a truncated tail. Combined with the process-group kill above
		// the grandchild also gets reaped instead of being orphaned.
		cmd.WaitDelay = 2 * time.Second

		output, err := cmd.CombinedOutput()
		result := truncateOutput(string(output))
		if err != nil {
			// When ctx fired the timeout (or the user pressed Stop),
			// surface a clear, model-friendly message instead of the
			// terse "signal: killed" so the LLM understands *why* it
			// got no useful output and can adapt (e.g. avoid spawning
			// long-running background processes from exec next time).
			if execCtx.Err() == context.DeadlineExceeded {
				suffix := fmt.Sprintf("\n[exec timed out after %ds — for long-running services (http.server, daemons), run them outside this tool or use disown/setsid + redirect ALL fds; e.g. `setsid sh -c 'python3 -m http.server 8899 </dev/null >/tmp/x.log 2>&1 &' </dev/null >/dev/null 2>&1`]", timeout)
				return result + suffix, nil
			}
			return fmt.Sprintf("%s\nError: %s", result, err.Error()), err
		}
		return result, nil
	}
}

// resolveSkillEnv checks if the command path references a skill directory
// and returns the skill's configured env vars.
func resolveSkillEnv(command string, envProvider SkillEnvProvider, skillDirs []string) map[string]string {
	// Check if any skill directory appears in the command
	for _, dir := range skillDirs {
		if strings.Contains(command, dir) {
			// Extract skill name from the path after the skill dir
			rest := command[strings.Index(command, dir)+len(dir):]
			if len(rest) > 0 && rest[0] == '/' {
				rest = rest[1:]
			}
			parts := strings.SplitN(rest, "/", 2)
			if len(parts) > 0 && parts[0] != "" {
				if env := envProvider(parts[0]); env != nil {
					return env
				}
			}
		}
	}
	return nil
}

// truncateOutput caps tool output at maxExecOutputBytes to protect context window.
func truncateOutput(s string) string {
	if len(s) <= maxExecOutputBytes {
		return s
	}
	return s[:maxExecOutputBytes] + fmt.Sprintf("\n\n[Output truncated: %d bytes omitted]", len(s)-maxExecOutputBytes)
}

// mergeEnv merges base env with additional vars. Additional vars override base.
func mergeEnv(base []string, additional map[string]string) []string {
	env := make([]string, 0, len(base)+len(additional))
	overridden := make(map[string]bool, len(additional))

	for _, e := range base {
		key := e
		if idx := strings.IndexByte(e, '='); idx >= 0 {
			key = e[:idx]
		}
		if _, ok := additional[key]; ok {
			overridden[key] = true
			continue // skip, will be added from additional
		}
		env = append(env, e)
	}

	for k, v := range additional {
		env = append(env, k+"="+v)
	}

	return env
}
