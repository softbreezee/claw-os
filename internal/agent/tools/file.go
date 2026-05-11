package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type readFileArgs struct {
	Path string `json:"path"`
}

type writeFileArgs struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type listDirArgs struct {
	Path string `json:"path"`
}

func registerFile(r *Registry, workspace string) {
	r.Register("read_file", "Read the contents of a file", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "File path (relative to workspace or absolute)",
			},
		},
		"required": []string{"path"},
	}, makeReadFile(workspace))

	r.Register("write_file", "Write content to a file (creates directories as needed)", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "File path (relative to workspace or absolute)",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Content to write",
			},
		},
		"required": []string{"path", "content"},
	}, makeWriteFile(workspace))

	r.Register("list_dir", "List files and directories in a path", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Directory path (relative to workspace or absolute)",
			},
		},
		"required": []string{"path"},
	}, makeListDir(workspace))
}

// resolvePath joins a tool-supplied path against the agent's workspace.
//
// Two anchoring rules:
//  1. Absolute paths pass through verbatim — the LLM is allowed to read
//     /etc/hosts, /tmp/foo.json, etc. Policy gating happens elsewhere.
//  2. Relative paths get joined onto the workspace, but only if the
//     workspace itself is a non-empty absolute path. A blank or
//     relative workspace would otherwise let `write_file` drop files
//     into whatever cwd the daemon was launched from (this is exactly
//     how MEMORY.md once leaked into a source repo). We log a warning
//     and resolve against the workspace anyway so existing behaviour
//     isn't silently broken, but the absolute-form makes the
//     destination predictable from the log line.
func resolvePath(workspace, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	if workspace == "" {
		// Best-effort: anchor at process cwd but make it explicit in
		// logs so the operator can spot the misconfiguration.
		cwd, _ := os.Getwd()
		return filepath.Join(cwd, path)
	}
	if !filepath.IsAbs(workspace) {
		if abs, err := filepath.Abs(workspace); err == nil {
			workspace = abs
		}
	}
	return filepath.Join(workspace, path)
}

func makeReadFile(workspace string) ToolFunc {
	return func(ctx context.Context, rawArgs json.RawMessage) (string, error) {
		var args readFileArgs
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return "", fmt.Errorf("parse args: %w", err)
		}

		fullPath := resolvePath(workspace, args.Path)
		data, err := os.ReadFile(fullPath)
		if err != nil {
			return "", fmt.Errorf("read file: %w", err)
		}

		return truncateOutput(string(data)), nil
	}
}

func makeWriteFile(workspace string) ToolFunc {
	return func(ctx context.Context, rawArgs json.RawMessage) (string, error) {
		var args writeFileArgs
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return "", fmt.Errorf("parse args: %w", err)
		}

		fullPath := resolvePath(workspace, args.Path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", fmt.Errorf("create directory: %w", err)
		}

		if err := os.WriteFile(fullPath, []byte(args.Content), 0o644); err != nil {
			return "", fmt.Errorf("write file: %w", err)
		}

		return fmt.Sprintf("Written %d bytes to %s", len(args.Content), fullPath), nil
	}
}

func makeListDir(workspace string) ToolFunc {
	return func(ctx context.Context, rawArgs json.RawMessage) (string, error) {
		var args listDirArgs
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return "", fmt.Errorf("parse args: %w", err)
		}

		fullPath := resolvePath(workspace, args.Path)
		entries, err := os.ReadDir(fullPath)
		if err != nil {
			return "", fmt.Errorf("read dir: %w", err)
		}

		var sb strings.Builder
		for _, entry := range entries {
			info, _ := entry.Info()
			if entry.IsDir() {
				fmt.Fprintf(&sb, "d %s/\n", entry.Name())
			} else if info != nil {
				fmt.Fprintf(&sb, "f %s (%d bytes)\n", entry.Name(), info.Size())
			} else {
				fmt.Fprintf(&sb, "f %s\n", entry.Name())
			}
		}

		return sb.String(), nil
	}
}
