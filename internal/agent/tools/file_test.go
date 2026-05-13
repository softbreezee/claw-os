package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// callTool finds a registered tool by name and invokes it with the given args.
// Centralised so each test stays focused on the assertion, not on registry
// plumbing.
func callTool(t *testing.T, workspace, name string, args map[string]any) (string, error) {
	t.Helper()
	r := NewRegistry(workspace)
	fn := r.GetFunc(name)
	if fn == nil {
		t.Fatalf("tool %q not registered", name)
	}
	raw, err := json.Marshal(args)
	if err != nil {
		t.Fatalf("marshal args: %v", err)
	}
	return fn(context.Background(), json.RawMessage(raw))
}

// TestWriteFile_EmptyPath_Rejected reproduces the production crash mode
// observed with kimi-k2.6: the model emitted a write_file call whose
// `path` field was missing/empty, and the tool happily joined the empty
// string against the workspace root, then asked the OS to write to that
// directory and got back the famously opaque "is a directory" error.
//
// The new guard turns that into a self-explanatory message so the model
// can correct itself on the next turn.
func TestWriteFile_EmptyPath_Rejected(t *testing.T) {
	ws := t.TempDir()

	cases := []struct {
		name string
		path string
	}{
		{"empty string", ""},
		{"only whitespace", "   "},
		{"only tab", "\t"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := callTool(t, ws, "write_file", map[string]any{
				"path":    tc.path,
				"content": "x",
			})
			if err == nil {
				t.Fatalf("expected error for empty path, got output %q", out)
			}
			if !strings.Contains(err.Error(), "path is required") {
				t.Fatalf("error should mention path requirement, got: %v", err)
			}
		})
	}
}

// TestWriteFile_DirectoryTarget_Rejected covers the second degenerate
// case: a non-empty path that still resolves to an existing directory
// (e.g. ".", "./", or a real subdirectory name). Without the guard the
// OS error remains "is a directory"; with the guard the model is told
// exactly which path collided so it can pick a real filename.
func TestWriteFile_DirectoryTarget_Rejected(t *testing.T) {
	ws := t.TempDir()
	if err := os.MkdirAll(filepath.Join(ws, "subdir"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	cases := []struct {
		name string
		path string
	}{
		{"dot resolves to workspace root", "."},
		{"existing subdirectory", "subdir"},
		{"trailing slash subdirectory", "subdir/"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := callTool(t, ws, "write_file", map[string]any{
				"path":    tc.path,
				"content": "x",
			})
			if err == nil {
				t.Fatalf("expected error when target is a directory")
			}
			if !strings.Contains(err.Error(), "directory") {
				t.Fatalf("error should mention directory collision, got: %v", err)
			}
		})
	}
}

// TestWriteFile_HappyPath ensures the guards don't regress normal usage:
// a relative filename should land inside workspace, content should match,
// and parent directories should be created on demand.
func TestWriteFile_HappyPath(t *testing.T) {
	ws := t.TempDir()

	out, err := callTool(t, ws, "write_file", map[string]any{
		"path":    "nested/dir/report.html",
		"content": "<h1>hi</h1>",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := filepath.Join(ws, "nested/dir/report.html")
	if !strings.Contains(out, want) {
		t.Fatalf("output should reference resolved path %q, got: %s", want, out)
	}
	got, err := os.ReadFile(want)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(got) != "<h1>hi</h1>" {
		t.Fatalf("content mismatch: %q", string(got))
	}
}

// TestReadFile_EmptyPath_Rejected mirrors the write_file guard: an
// empty path used to silently resolve to the workspace root and then
// fail with "is a directory". The guard surfaces the real cause.
func TestReadFile_EmptyPath_Rejected(t *testing.T) {
	ws := t.TempDir()
	_, err := callTool(t, ws, "read_file", map[string]any{
		"path": "",
	})
	if err == nil {
		t.Fatal("expected error for empty path")
	}
	if !strings.Contains(err.Error(), "path is required") {
		t.Fatalf("error should mention path requirement, got: %v", err)
	}
}

// TestReadFile_DirectoryTarget_Rejected makes read_file point users
// (and the model) at list_dir when they accidentally aim at a folder.
func TestReadFile_DirectoryTarget_Rejected(t *testing.T) {
	ws := t.TempDir()
	if err := os.MkdirAll(filepath.Join(ws, "logs"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	_, err := callTool(t, ws, "read_file", map[string]any{
		"path": "logs",
	})
	if err == nil {
		t.Fatal("expected error when reading a directory")
	}
	if !strings.Contains(err.Error(), "list_dir") {
		t.Fatalf("error should suggest list_dir, got: %v", err)
	}
}
