// Package build — inject.go
//
// WriteInjectionTar assembles private-injection.tar into the OUTPUT directory
// (next to the ISO, never under the arch-profile tree). Each private-pane
// file asset is stored at its target-filesystem-relative path (e.g.
// etc/ssh/authorized_keys, home/user/.zshrc) plus a root-level
// debateos-private.json manifest.
//
// Security:
//   - T-03-TRAV: dst paths are sanitized with the same rule as
//     translators/arch/profile.py _sanitize_dst — absolute paths and `..`
//     traversal are rejected before any archive entry is written.
//   - T-03-LEAK: tar is written to --out (outDir), never inside the profile tree
//     (Pitfall 3 guard); the caller owns placement.
package build

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PaneAsset describes a single private-pane file asset to include in the
// injection tar. Dst is the target-filesystem-relative path
// (e.g. "etc/ssh/authorized_keys"). Content is the raw bytes.
type PaneAsset struct {
	Dst     string // target-relative path (sanitized — no absolute, no ..)
	Content []byte // file content
	Mode    int64  // Unix permission bits (e.g. 0600, 0644)
}

// privateManifest is the root-level debateos-private.json manifest embedded
// in the injection tar. Version, creation timestamp, and per-file metadata
// allow the first-boot unit to validate the archive and apply correct modes.
type privateManifest struct {
	Version int                    `json:"version"`
	Created string                 `json:"created"`
	Files   []privateManifestEntry `json:"files"`
}

type privateManifestEntry struct {
	Path string `json:"path"`
	Mode int64  `json:"mode"`
}

// sanitizeDst mirrors translators/arch/profile.py _sanitize_dst (T-02-08):
// rejects empty, absolute, and ../ traversal paths; returns the normalized
// relative form.
func sanitizeDst(dst string) (string, error) {
	// Reject empty or "."
	stripped := strings.TrimSpace(dst)
	if stripped == "" || stripped == "." {
		return "", fmt.Errorf(
			"injection dst is empty or '.': a concrete relative path is required "+
				"(T-03-TRAV path traversal guard)",
		)
	}

	// Reject absolute paths (mirrors os.path.isabs check in profile.py).
	if filepath.IsAbs(dst) {
		return "", fmt.Errorf(
			"injection dst %q is an absolute path: all dst paths must be "+
				"relative to the target root (no absolute paths — T-03-TRAV)",
			dst,
		)
	}

	// Normalize and check containment against a sentinel root.
	// We join with a known-safe sentinel and confirm the result stays inside.
	const sentinel = "/debateos-injection-root"
	joined := filepath.Clean(filepath.Join(sentinel, dst))
	if joined != sentinel && !strings.HasPrefix(joined, sentinel+"/") {
		return "", fmt.Errorf(
			"injection dst %q traverses outside the target root "+
				"(resolved to %q): path components '..' that escape the root "+
				"are rejected (T-03-TRAV)",
			dst, joined,
		)
	}

	// Return target-relative form (strip the sentinel prefix).
	rel, err := filepath.Rel(sentinel, joined)
	if err != nil {
		return "", fmt.Errorf("injection dst %q: rel path error: %w", dst, err)
	}
	return rel, nil
}

// WriteInjectionTar writes private-injection.tar into outDir (NOT under any
// profile subdirectory). Each asset is stored at its target-relative dst path;
// a debateos-private.json manifest is added at the tar root.
//
// Returns the absolute path of the written tar file, or an error if any dst
// fails sanitization or an I/O error occurs.
//
// Security: T-03-LEAK — tar is written to outDir (next to ISO), never into
// the arch-profile/ tree (Pitfall 3). Callers must not move it inside the
// profile tree.
func WriteInjectionTar(outDir string, assets []PaneAsset) (string, error) {
	// Sanitize all dsts before writing anything.
	sanitized := make([]string, len(assets))
	for i, a := range assets {
		clean, err := sanitizeDst(a.Dst)
		if err != nil {
			return "", fmt.Errorf("asset[%d]: %w", i, err)
		}
		sanitized[i] = clean
	}

	// Build the manifest.
	entries := make([]privateManifestEntry, len(assets))
	for i, a := range assets {
		entries[i] = privateManifestEntry{Path: sanitized[i], Mode: a.Mode}
	}
	manifest := privateManifest{
		Version: 1,
		Created: time.Now().UTC().Format(time.RFC3339),
		Files:   entries,
	}
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return "", fmt.Errorf("marshal manifest: %w", err)
	}

	// Ensure outDir exists.
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", outDir, err)
	}

	tarPath := filepath.Join(outDir, "private-injection.tar")
	f, err := os.Create(tarPath)
	if err != nil {
		return "", fmt.Errorf("create %s: %w", tarPath, err)
	}
	defer f.Close()

	tw := tar.NewWriter(f)
	defer tw.Close()

	// Write debateos-private.json at the tar root.
	hdr := &tar.Header{
		Name:     "debateos-private.json",
		Mode:     0644,
		Size:     int64(len(manifestBytes)),
		Typeflag: tar.TypeReg,
		ModTime:  time.Unix(0, 0).UTC(), // deterministic
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return "", fmt.Errorf("tar header manifest: %w", err)
	}
	if _, err := tw.Write(manifestBytes); err != nil {
		return "", fmt.Errorf("tar write manifest: %w", err)
	}

	// Write each asset at its sanitized target-relative path.
	for i, a := range assets {
		hdr := &tar.Header{
			Name:     sanitized[i],
			Mode:     a.Mode,
			Size:     int64(len(a.Content)),
			Typeflag: tar.TypeReg,
			ModTime:  time.Unix(0, 0).UTC(), // deterministic
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return "", fmt.Errorf("tar header %q: %w", sanitized[i], err)
		}
		if _, err := tw.Write(a.Content); err != nil {
			return "", fmt.Errorf("tar write %q: %w", sanitized[i], err)
		}
	}

	if err := tw.Flush(); err != nil {
		return "", fmt.Errorf("tar flush: %w", err)
	}

	return tarPath, nil
}
