package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ─── Patch DSL Parser ───────────────────────────────────────────────────────
//
// apply_patch accepts a multi-file patch in a structured format inspired by
// OpenClaw's patch DSL. The two-phase design (parse + validate all ops, then
// apply only when every op is well-formed) keeps the filesystem consistent:
// a broken patch at line 50 won't leave half-written files behind.

// patchOp represents one file-level operation in a patch.
type patchOp struct {
	kind    string // "add", "update", "delete"
	path    string
	moveTo  string     // only for "update" with "Move to:"
	hunks   []hunkDesc // only for "update"
	content string     // only for "add" — raw lines joined with \n
}

// hunkDesc is a single search-and-replace block inside an update op.
type hunkDesc struct {
	contextLines string
	oldLines     string
	newLines     string
}

// patch holds the parsed result.
type patch struct {
	ops []patchOp
}

// ─── Parser ──────────────────────────────────────────────────────────────────

const (
	markerBegin  = "*** Begin Patch"
	markerEnd    = "*** End Patch"
	markerAdd    = "*** Add File:"
	markerUpdate = "*** Update File:"
	markerDelete = "*** Delete File:"
	markerMove   = "*** Move to:"
	markerEOF    = "*** End of File"
)

// parsePatchError wraps a patch parse/apply error with a line number.
type parsePatchError struct {
	line    int
	message string
}

func (e *parsePatchError) Error() string {
	return fmt.Sprintf("apply_patch line %d: %s", e.line, e.message)
}

// parsePatch parses a raw patch string into structured operations.
func parsePatch(input string) (*patch, error) {
	lines := strings.Split(input, "\n")
	p := &patch{}

	i := 0
	// Expect *** Begin Patch header
	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}
	if i >= len(lines) || !strings.HasPrefix(strings.TrimSpace(lines[i]), markerBegin) {
		return nil, &parsePatchError{line: i + 1, message: "patch must start with '*** Begin Patch'"}
	}
	i++

	// Parse file-level blocks
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			i++
			continue
		}
		if strings.HasPrefix(line, markerEnd) {
			i++
			// Skip trailing *** End Patch
			i++
			break
		}

		switch {
		case strings.HasPrefix(line, markerAdd):
			path := strings.TrimSpace(strings.TrimPrefix(line, markerAdd))
			if path == "" {
				return nil, &parsePatchError{line: i + 1, message: "Add File: missing path"}
			}
			i++
			var contentLines []string
			for i < len(lines) {
				cl := lines[i]
				if strings.HasPrefix(strings.TrimSpace(cl), "***") {
					break
				}
				if strings.HasPrefix(cl, "+") {
					contentLines = append(contentLines, cl[1:])
				} else {
					contentLines = append(contentLines, cl)
				}
				i++
			}
			// Trim trailing empty lines
			for len(contentLines) > 0 && contentLines[len(contentLines)-1] == "" {
				contentLines = contentLines[:len(contentLines)-1]
			}
			p.ops = append(p.ops, patchOp{
				kind:    "add",
				path:    path,
				content: strings.Join(contentLines, "\n"),
			})

		case strings.HasPrefix(line, markerUpdate):
			path := strings.TrimSpace(strings.TrimPrefix(line, markerUpdate))
			if path == "" {
				return nil, &parsePatchError{line: i + 1, message: "Update File: missing path"}
			}
			i++
			moveTo := ""
			// Check for optional *** Move to: directive
			if i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), markerMove) {
				moveTo = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[i]), markerMove))
				i++
			}
			// Parse hunks
			var hunks []hunkDesc
			for i < len(lines) {
				line = strings.TrimSpace(lines[i])
				if line == "" {
					i++
					continue
				}
				if strings.HasPrefix(line, markerAdd) || strings.HasPrefix(line, markerUpdate) ||
					strings.HasPrefix(line, markerDelete) || strings.HasPrefix(line, markerEnd) {
					break
				}
				// Hunk starts with @@
				if line == "@@" {
					hunk, nextLine, err := parseHunk(lines, i+1)
					if err != nil {
						return nil, err
					}
					hunks = append(hunks, hunk)
					i = nextLine
					// After a hunk, expect *** End of File
					if i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), markerEOF) {
						i++
						break // end of this Update File block
					}
					continue
				}
				// Also allow *** End of File without any hunks
				if strings.HasPrefix(line, markerEOF) {
					i++
					break
				}
				return nil, &parsePatchError{line: i + 1, message: fmt.Sprintf("unexpected line in Update File block: %q", line)}
			}
			p.ops = append(p.ops, patchOp{
				kind:   "update",
				path:   path,
				moveTo: moveTo,
				hunks:  hunks,
			})

		case strings.HasPrefix(line, markerDelete):
			path := strings.TrimSpace(strings.TrimPrefix(line, markerDelete))
			if path == "" {
				return nil, &parsePatchError{line: i + 1, message: "Delete File: missing path"}
			}
			p.ops = append(p.ops, patchOp{kind: "delete", path: path})
			i++

		default:
			return nil, &parsePatchError{line: i + 1, message: fmt.Sprintf("unexpected directive: %q", line)}
		}
	}

	return p, nil
}

// parseHunk reads a single hunk block starting at the line after @@.
// Returns the hunk and the index of the next unparsed line.
func parseHunk(lines []string, start int) (hunkDesc, int, error) {
	// States: "context", "old", "new". After seeing old, we expect new.
	// Lines without prefix are context.
	// Lines starting with - are old.
	// Lines starting with + are new.
	var contextLines, oldLines, newLines []string
	i := start

	for i < len(lines) {
		raw := lines[i]
		if raw == "" {
			// Empty line in a hunk is context
			contextLines = append(contextLines, "")
			i++
			continue
		}

		trimmed := strings.TrimLeft(raw, " \t")
		if strings.HasPrefix(trimmed, "***") || strings.HasPrefix(trimmed, "@@") {
			// End of hunk — next directive or next hunk
			break
		}

		// Check first significant character
		// Find the first non-whitespace character
		firstNonWS := -1
		for idx, ch := range raw {
			if ch != ' ' && ch != '\t' {
				firstNonWS = idx
				break
			}
		}
		// Context lines: no +/- prefix or just whitespace
		if firstNonWS == -1 {
			contextLines = append(contextLines, raw)
		} else if raw[firstNonWS] == '-' {
			oldLines = append(oldLines, raw[firstNonWS+1:])
		} else if raw[firstNonWS] == '+' {
			newLines = append(newLines, raw[firstNonWS+1:])
		} else {
			contextLines = append(contextLines, raw)
		}
		i++
	}

	// If nothing meaningful was parsed, error
	if len(oldLines) == 0 && len(newLines) == 0 && len(contextLines) == 0 {
		return hunkDesc{}, i, &parsePatchError{line: start, message: "empty hunk — expected context/old/new lines after @@"}
	}

	return hunkDesc{
		contextLines: strings.Join(contextLines, "\n"),
		oldLines:     strings.Join(oldLines, "\n"),
		newLines:     strings.Join(newLines, "\n"),
	}, i, nil
}

// ─── Applier ─────────────────────────────────────────────────────────────────

// applyPatchToFile searches for the old content anchored by context lines and
// replaces it with new content. Returns the new file content.
func applyPatchToFile(path string, content []byte, hunk hunkDesc) ([]byte, error) {
	if hunk.oldLines == "" && hunk.newLines == "" {
		// Pure context hunk — no-op
		return content, nil
	}
	if hunk.oldLines == "" {
		// Insert-only hunk (rare in practice)
		anchor := []byte(hunk.contextLines)
		pos := bytes.Index(content, anchor)
		if pos < 0 {
			return nil, fmt.Errorf("hunk anchor context not found in %s", path)
		}
		if hunk.newLines == "" {
			return content, nil
		}
		result := make([]byte, 0, len(content)+len(hunk.newLines)+1)
		result = append(result, content[:pos]...)
		result = append(result, []byte(hunk.newLines)...)
		result = append(result, '\n')
		result = append(result, content[pos:]...)
		return result, nil
	}

	// Build the search target: oldLines joined with \n
	search := []byte(hunk.oldLines)
	pos := bytes.Index(content, search)
	if pos < 0 {
		return nil, fmt.Errorf("hunk old content not found in %s — verify the file has not been modified concurrently", path)
	}

	if hunk.newLines == "" {
		// Delete: remove oldLines and the preceding newline if present
		end := pos + len(search)
		// Eat the following newline if present
		if end < len(content) && content[end] == '\n' {
			end++
		}
		result := make([]byte, 0, len(content)-len(search))
		result = append(result, content[:pos]...)
		result = append(result, content[end:]...)
		return result, nil
	}

	// Replace old with new
	result := make([]byte, 0, len(content)-len(search)+len(hunk.newLines))
	result = append(result, content[:pos]...)
	result = append(result, []byte(hunk.newLines)...)
	result = append(result, content[pos+len(search):]...)
	return result, nil
}

// ─── Tool Registration ──────────────────────────────────────────────────────

func registerApplyPatch(r *Registry) {
	desc := `Apply a multi-file patch with atomic semantics (all-or-nothing). Use this for complex edits spanning multiple files.

Patch format:
  *** Begin Patch
  *** Update File: path/to/file.go
  *** Move to: path/to/renamed.go
  @@
   context line
  -old line
  +new line
  *** End of File

  *** Add File: path/to/new.txt
  +line 1
  +line 2

  *** Delete File: path/to/delete.txt

  *** End Patch`

	r.Register("apply_patch", desc, map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"input": map[string]interface{}{
				"type":        "string",
				"description": "Full patch content in the structured patch format",
			},
		},
		"required": []string{"input"},
	}, makeApplyPatch())
}

type applyPatchArgs struct {
	Input string `json:"input"`
}

func makeApplyPatch() ToolFunc {
	return func(ctx context.Context, rawArgs json.RawMessage) (string, error) {
		var args applyPatchArgs
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return "", fmt.Errorf("parse args: %w", err)
		}

		if strings.TrimSpace(args.Input) == "" {
			return "", fmt.Errorf("apply_patch: input is required")
		}

		p, err := parsePatch(args.Input)
		if err != nil {
			return "", err
		}

		if len(p.ops) == 0 {
			return "", fmt.Errorf("apply_patch: no operations found in patch")
		}

		// Phase 1: compute all new file contents in memory.
		// We read each file and apply hunks locally; only after every
		// operation succeeds do we flush to disk.

		type fileWrite struct {
			path      string
			content   []byte
			isAdd     bool
			origMode  os.FileMode
			deleteOld []string // paths to remove after write (for "move to")
		}
		var writes []fileWrite
		var deletes []string

		// Read-back cache so multiple hunks to the same file compose correctly.
		fileCache := make(map[string][]byte)

		for _, op := range p.ops {
			switch op.kind {
			case "add":
				if op.content == "" {
					return "", fmt.Errorf("apply_patch: Add File %q has no content", op.path)
				}
				writes = append(writes, fileWrite{
					path:    op.path,
					content: []byte(op.content),
					isAdd:   true,
				})

			case "update":
				// Read file (from cache or disk)
				data, ok := fileCache[op.path]
				if !ok {
					var err error
					data, err = os.ReadFile(op.path)
					if err != nil {
						return "", fmt.Errorf("apply_patch: cannot read %s: %w", op.path, err)
					}
					if isBinary(data) {
						return "", fmt.Errorf("apply_patch: %s is binary; use write_file instead", op.path)
					}
					fileCache[op.path] = data
				}

				current := data
				for _, hunk := range op.hunks {
					newData, err := applyPatchToFile(op.path, current, hunk)
					if err != nil {
						return "", err
					}
					current = newData
				}
				fileCache[op.path] = current

				destPath := op.path
				if op.moveTo != "" {
					destPath = op.moveTo
				}

				var delOld []string
				if op.moveTo != "" {
					delOld = []string{op.path}
				}
				writes = append(writes, fileWrite{
					path:      destPath,
					content:   current,
					deleteOld: delOld,
				})

			case "delete":
				deletes = append(deletes, op.path)
			}
		}

		// Phase 2: apply all writes
		for _, w := range writes {
			dir := filepath.Dir(w.path)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return "", fmt.Errorf("apply_patch: create directory %s: %w", dir, err)
			}
			if err := os.WriteFile(w.path, w.content, 0o644); err != nil {
				return "", fmt.Errorf("apply_patch: write %s: %w", w.path, err)
			}
			for _, oldPath := range w.deleteOld {
				os.Remove(oldPath) // best-effort; don't fail the whole patch
			}
		}

		for _, path := range deletes {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return "", fmt.Errorf("apply_patch: delete %s: %w", path, err)
			}
		}

		// Build summary
		var parts []string
		for _, op := range p.ops {
			switch op.kind {
			case "add":
				parts = append(parts, fmt.Sprintf("added %s", op.path))
			case "update":
				dest := op.path
				if op.moveTo != "" {
					dest = fmt.Sprintf("%s → %s", op.path, op.moveTo)
				}
				parts = append(parts, fmt.Sprintf("updated %s", dest))
			case "delete":
				parts = append(parts, fmt.Sprintf("deleted %s", op.path))
			}
		}

		return fmt.Sprintf("Patch applied (%d ops): %s", len(p.ops), strings.Join(parts, "; ")), nil
	}
}
