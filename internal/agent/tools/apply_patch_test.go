package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ─── Parser tests ───────────────────────────────────────────────────────

func TestParsePatch_AddFile(t *testing.T) {
	input := `*** Begin Patch
*** Add File: new/config.json
+{"key": "value"}

*** End Patch`

	p, err := parsePatch(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(p.ops) != 1 {
		t.Fatalf("ops = %d, want 1", len(p.ops))
	}
	op := p.ops[0]
	if op.kind != "add" {
		t.Fatalf("kind = %s, want add", op.kind)
	}
	if op.path != "new/config.json" {
		t.Fatalf("path = %s, want new/config.json", op.path)
	}
	if op.content != `{"key": "value"}` {
		t.Fatalf("content = %q", op.content)
	}
}

func TestParsePatch_UpdateFile(t *testing.T) {
	input := `*** Begin Patch
*** Update File: src/main.go
@@
 package main
-old import
+new import

*** End of File

*** End Patch`

	p, err := parsePatch(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(p.ops) != 1 {
		t.Fatalf("ops = %d, want 1", len(p.ops))
	}
	op := p.ops[0]
	if op.kind != "update" {
		t.Fatalf("kind = %s, want update", op.kind)
	}
	if len(op.hunks) != 1 {
		t.Fatalf("hunks = %d, want 1", len(op.hunks))
	}
	if op.hunks[0].oldLines != "old import" {
		t.Fatalf("hunk oldLines = %q, want 'old import'", op.hunks[0].oldLines)
	}
	if op.hunks[0].newLines != "new import" {
		t.Fatalf("hunk newLines = %q, want 'new import'", op.hunks[0].newLines)
	}
}

func TestParsePatch_DeleteFile(t *testing.T) {
	input := `*** Begin Patch
*** Delete File: tmp/log.txt

*** End Patch`

	p, err := parsePatch(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(p.ops) != 1 {
		t.Fatalf("ops = %d, want 1", len(p.ops))
	}
	op := p.ops[0]
	if op.kind != "delete" {
		t.Fatalf("kind = %s, want delete", op.kind)
	}
	if op.path != "tmp/log.txt" {
		t.Fatalf("path = %s, want tmp/log.txt", op.path)
	}
}

func TestParsePatch_MoveTo(t *testing.T) {
	input := `*** Begin Patch
*** Update File: old/name.go
*** Move to: new/name.go
@@
 context
-old
+new

*** End of File

*** End Patch`

	p, err := parsePatch(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(p.ops) != 1 {
		t.Fatalf("ops = %d, want 1", len(p.ops))
	}
	op := p.ops[0]
	if op.moveTo != "new/name.go" {
		t.Fatalf("moveTo = %s, want new/name.go", op.moveTo)
	}
}

func TestParsePatch_MissingBegin(t *testing.T) {
	_, err := parsePatch("*** Update File: x.go")
	if err == nil {
		t.Fatal("expected error for missing Begin Patch")
	}
	if !strings.Contains(err.Error(), "Begin Patch") {
		t.Fatalf("error should mention Begin Patch, got: %v", err)
	}
}

func TestParsePatch_MultipleOps(t *testing.T) {
	input := `*** Begin Patch
*** Add File: a.txt
+hello

*** Update File: b.txt
@@
 context
-old
+new

*** End of File

*** Delete File: c.txt

*** End Patch`

	p, err := parsePatch(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(p.ops) != 3 {
		t.Fatalf("ops = %d, want 3", len(p.ops))
	}
	if p.ops[0].kind != "add" || p.ops[1].kind != "update" || p.ops[2].kind != "delete" {
		t.Fatalf("unexpected op order: %v", p.ops)
	}
}

// ─── Apply tool tests ───────────────────────────────────────────────────

func TestApplyPatch_AddFile(t *testing.T) {
	ws := t.TempDir()
	target := filepath.Join(ws, "new.txt")
	// apply_patch uses absolute or relative paths; we use absolute for test determinism
	input := "*** Begin Patch\n*** Add File: " + target + "\n+hello world\n\n*** End Patch"
	out, err := callTool(t, ws, "apply_patch", map[string]any{"input": input})
	if err != nil {
		t.Fatalf("apply_patch error: %v", err)
	}
	if !strings.Contains(out, "added") {
		t.Fatalf("output should mention added, got: %s", out)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(data) != "hello world" {
		t.Fatalf("content = %q, want 'hello world'", string(data))
	}
}

func TestApplyPatch_UpdateFile(t *testing.T) {
	ws := t.TempDir()
	target := filepath.Join(ws, "main.go")
	orig := "package main\n\nimport \"fmt\"\n\nfunc main() {}\n"
	if err := os.WriteFile(target, []byte(orig), 0o644); err != nil {
		t.Fatal(err)
	}

	input := "*** Begin Patch\n*** Update File: " + target + "\n@@\n import \"fmt\"\n-import \"fmt\"\n+import (\n+\t\"fmt\"\n+\t\"os\"\n+)\n\n*** End of File\n\n*** End Patch"
	out, err := callTool(t, ws, "apply_patch", map[string]any{"input": input})
	if err != nil {
		t.Fatalf("apply_patch error: %v", err)
	}
	if !strings.Contains(out, "updated") {
		t.Fatalf("output should mention updated, got: %s", out)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if !strings.Contains(string(data), "\"os\"") {
		t.Fatalf("content should contain 'os', got: %s", string(data))
	}
}

func TestApplyPatch_DeleteFile(t *testing.T) {
	ws := t.TempDir()
	target := filepath.Join(ws, "remove-me.txt")
	if err := os.WriteFile(target, []byte("bye"), 0o644); err != nil {
		t.Fatal(err)
	}

	input := "*** Begin Patch\n*** Delete File: " + target + "\n\n*** End Patch"
	out, err := callTool(t, ws, "apply_patch", map[string]any{"input": input})
	if err != nil {
		t.Fatalf("apply_patch error: %v", err)
	}
	if !strings.Contains(out, "deleted") {
		t.Fatalf("output should mention deleted, got: %s", out)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("file should be deleted, stat returned: %v", err)
	}
}

func TestApplyPatch_AtomicityOnHunkFailure(t *testing.T) {
	ws := t.TempDir()
	// Create a file with content that won't match the hunk
	target := filepath.Join(ws, "data.txt")
	if err := os.WriteFile(target, []byte("line 1\nline 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Patch with a hunk that won't find a match
	input := "*** Begin Patch\n*** Update File: " + target + "\n@@\n context\n-non-existent line\n+replacement\n\n*** End of File\n\n*** End Patch"
	_, err := callTool(t, ws, "apply_patch", map[string]any{"input": input})
	if err == nil {
		t.Fatal("expected error for non-matching hunk")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("error should mention not found, got: %v", err)
	}

	// File should be unchanged (atomicity)
	data, _ := os.ReadFile(target)
	if string(data) != "line 1\nline 2\n" {
		t.Fatalf("file was modified despite hunk failure: %s", string(data))
	}
}

func TestApplyPatch_EmptyInput(t *testing.T) {
	ws := t.TempDir()
	_, err := callTool(t, ws, "apply_patch", map[string]any{"input": ""})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
	if !strings.Contains(err.Error(), "input is required") {
		t.Fatalf("error should mention input is required, got: %v", err)
	}
}
