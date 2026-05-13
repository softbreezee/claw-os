// Package upload is a tiny content-addressed file store for chat
// attachments uploaded via the Web UI.
//
// Design choices:
//
//   * Content-addressed (sha256-named) so the same image uploaded twice
//     deduplicates automatically. Cheaper than chasing per-task GC.
//
//   * Files live on the local filesystem under
//     ~/.pawnix/uploads/<sha>.<ext>. The path is what gets stored
//     in session messages (cheap to persist), and we only base64-inline
//     the bytes when actually building an LLM request.
//
//   * No public HTTP endpoint serves these files. The reason is most
//     hosted LLM APIs (OpenAI, Anthropic) need to fetch attachment
//     URLs from their own backends, and a localhost gateway URL won't
//     work for them. Inlining as data: URLs sidesteps the whole
//     network-reachability problem.
//
//   * Best-effort retention: nothing here cleans old files yet (the
//     larger Stage 3 plan covers that). For now `du -sh ~/.pawnix/uploads`
//     is the user's monitoring tool.
package upload

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

// Store is the upload directory. Construct with NewStore; safe for
// concurrent use (filesystem operations are atomic at the rename level).
type Store struct {
	dir string
}

// NewStore opens (and creates if missing) an upload directory rooted
// at the given path. Typically the caller passes
// filepath.Join(homeDir, ".pawnix", "uploads").
func NewStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("upload: mkdir %q: %w", dir, err)
	}
	return &Store{dir: dir}, nil
}

// Dir returns the underlying upload directory. Useful for diagnostics.
func (s *Store) Dir() string { return s.dir }

// Save streams the reader to disk under a content-addressed name and
// returns the absolute path. The mimeType drives the file extension
// (best-effort; falls back to the originalName's extension, then ".bin").
//
// Implementation notes:
//   - We hash on the way in so we don't have to rewind. This means a
//     temp file is used and renamed into place — atomic on POSIX.
//   - If a file with the same hash already exists, the temp file is
//     deleted and the existing path is returned (dedup).
func (s *Store) Save(r io.Reader, mimeType, originalName string) (string, error) {
	tmp, err := os.CreateTemp(s.dir, "upload-*.tmp")
	if err != nil {
		return "", fmt.Errorf("upload: create temp: %w", err)
	}
	tmpPath := tmp.Name()
	// We may rename or remove tmp depending on outcome; defer the
	// remove and let success paths neutralise it by clearing tmpPath.
	defer func() {
		if tmpPath != "" {
			_ = os.Remove(tmpPath)
		}
	}()

	hasher := sha256.New()
	mw := io.MultiWriter(tmp, hasher)
	if _, err := io.Copy(mw, r); err != nil {
		tmp.Close()
		return "", fmt.Errorf("upload: write: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("upload: close: %w", err)
	}

	sum := hex.EncodeToString(hasher.Sum(nil))
	ext := pickExtension(mimeType, originalName)
	finalPath := filepath.Join(s.dir, sum+ext)

	// Dedup: if the target exists, drop the temp file and reuse.
	if _, err := os.Stat(finalPath); err == nil {
		return finalPath, nil
	}

	if err := os.Rename(tmpPath, finalPath); err != nil {
		return "", fmt.Errorf("upload: rename: %w", err)
	}
	tmpPath = "" // suppress deferred remove
	return finalPath, nil
}

// pickExtension returns the most appropriate file extension (including
// the leading dot) for the given content. Order of preference:
//   1. Canonical extension for the MIME type (".jpg" for image/jpeg).
//   2. Extension from the original filename, if any.
//   3. ".bin" as a last-resort placeholder.
func pickExtension(mimeType, originalName string) string {
	if mimeType != "" {
		// mime.ExtensionsByType returns sorted-but-arbitrary aliases;
		// prefer well-known ones for image types.
		if pref := preferredImageExt(mimeType); pref != "" {
			return pref
		}
		exts, _ := mime.ExtensionsByType(mimeType)
		if len(exts) > 0 {
			return exts[0]
		}
	}
	if originalName != "" {
		if ext := filepath.Ext(originalName); ext != "" {
			return ext
		}
	}
	return ".bin"
}

// preferredImageExt picks the conventional extension for common image
// MIME types. Avoids surprises like image/jpeg → ".jfif" that mime.ExtensionsByType
// has historically returned on some platforms.
func preferredImageExt(mimeType string) string {
	switch strings.ToLower(mimeType) {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/heic":
		return ".heic"
	}
	return ""
}
