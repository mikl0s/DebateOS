// Command gen-index regenerates the static registry index (registry/index.json)
// and its browse HTML (registry/index.html) from a source directory of
// point/opinion YAML. It is the thin wrapper the CI workflow
// (.github/workflows/registry-index.yml) and scripts/generate-index.sh call;
// the heavy lifting lives in package registry (GenerateIndex/LoadCapabilities/
// EmitHTML), which is unit-tested.
//
// Usage:
//
//	gen-index [source-dir] [out-dir]
//
// Defaults: source-dir = examples/omarchy, out-dir = registry.
// generatedAt is read from GENERATED_AT (the script supplies the source's last
// git commit date) so output is deterministic — re-running on unchanged source
// produces a byte-identical index.json and the CI commit step is a no-op.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mikl0s/debateos/registry"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "gen-index:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	sourceDir := "examples/omarchy"
	outDir := "registry"
	if len(args) >= 1 && args[0] != "" {
		sourceDir = args[0]
	}
	if len(args) >= 2 && args[1] != "" {
		outDir = args[1]
	}

	// Deterministic timestamp: the source's last-changed date, supplied by the
	// caller. Empty is allowed (renders as "") so the tool never embeds a
	// wall-clock value that would churn the committed index on every run.
	generatedAt := os.Getenv("GENERATED_AT")

	archCaps := filepath.Join("translators", "arch", "capabilities.json")
	debianCaps := filepath.Join("translators", "debian", "capabilities.json")
	caps, err := registry.LoadCapabilities(archCaps, debianCaps)
	if err != nil {
		return fmt.Errorf("load capabilities: %w", err)
	}

	idx, err := registry.GenerateIndex(sourceDir, caps, generatedAt)
	if err != nil {
		return fmt.Errorf("generate index from %s: %w", sourceDir, err)
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create out dir %s: %w", outDir, err)
	}

	// Write index.json (indented, deterministic — encoding/json sorts map keys
	// and the index entries are already sorted by ID in GenerateIndex).
	jsonPath := filepath.Join(outDir, "index.json")
	b, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal index: %w", err)
	}
	b = append(b, '\n')
	if err := os.WriteFile(jsonPath, b, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", jsonPath, err)
	}

	// Write the static browse HTML alongside it.
	htmlPath := filepath.Join(outDir, "index.html")
	hf, err := os.Create(htmlPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", htmlPath, err)
	}
	defer hf.Close()
	if err := registry.EmitHTML(idx, hf); err != nil {
		return fmt.Errorf("emit html: %w", err)
	}

	fmt.Printf("gen-index: wrote %s (%d points) and %s\n", jsonPath, len(idx.Points), htmlPath)
	return nil
}
