// Package build implements the `debateos build` subcommand.
//
// The build pipeline:
//  1. Load + resolve the speech directory (reusing cli/internal/loader).
//  2. Write canonical resolved.json to --out.
//  3. Derive SOURCE_DATE_EPOCH ONCE from the canonical resolved.json sha256
//     (mirrors translators/arch/manifest.py derive_source_date_epoch exactly).
//  4. Invoke translators/arch/translate via Runner with the FROZEN argv contract.
//  5. Invoke docker run via Runner with speech/out mounts + SOURCE_DATE_EPOCH env.
//
// Escape hatches:
//   - --dry-run: print the build plan (resolved.json path, epoch, translate/docker
//     argv) and make ZERO Runner calls.
//   - --skip-iso: run translate (step 4) then stop before docker (step 5).
//     This path works on hosts that cannot run mkarchiso (e.g. Proxmox devtmpfs).
//
// Security:
//   - T-03-DKARG: all subprocess calls use Runner.Run(name, args...) variadic —
//     never "sh -c" string interpolation.
//   - T-03-EPOCH: single epoch derivation from sha256(resolved.json); exported
//     to both translate and docker subprocess environments.
package build

import (
	"crypto/sha256"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	yaml "go.yaml.in/yaml/v3"

	"github.com/mikl0s/debateos/cli/config"
	"github.com/mikl0s/debateos/cli/internal/loader"
	"github.com/mikl0s/debateos/cli/runner"
	"github.com/mikl0s/debateos/resolver/resolve"
)

// Epoch derivation constants mirror manifest.py exactly (BLD-03 / T-03-EPOCH).
// _MIN and _MAX must remain in sync with:
//
//	translators/arch/manifest.py: _MIN_EPOCH = 1577836800, _MAX_EPOCH = 2208988800
const (
	epochMin = int64(1577836800) // 2020-01-01T00:00:00Z
	epochMax = int64(2208988800) // 2040-01-01T00:00:00Z
)

// dockerImage is the image reference used for the docker build channel.
// Per 03-RESEARCH A3 and 03-CONTEXT BLD-01.
const dockerImage = "ghcr.io/mikl0s/debateos:latest"

// foundationConfig holds the translator-specific build parameters for one
// foundation. The registry is the single source of truth for which translator
// binary, profile directory name, and default profile correspond to each
// foundation string (DEB-03 / T-04-10).
type foundationConfig struct {
	TranslateBin   string // e.g. "translators/arch/translate"
	ProfileDir     string // subdirectory name under outDir, e.g. "arch-profile"
	DefaultProfile string // default profile when --profile is empty, e.g. "vanilla-arch"
}

// foundationRegistry maps the speech Foundation field to build parameters.
// All values are compile-time constants — no runtime derivation from speech data
// (T-04-11: no injection surface).
//
// To add a new foundation: add a row here and add the translator under translators/.
// Future: "ubuntu": {"translators/debian/translate", "debian-profile", "ubuntu"},
var foundationRegistry = map[string]foundationConfig{
	"arch":   {"translators/arch/translate", "arch-profile", "vanilla-arch"},
	"debian": {"translators/debian/translate", "debian-profile", "debian"},
}

const buildUsage = `usage: debateos build [flags]

Flags:
  --dir <path>      Speech directory (default: DEBATEOS_DIR or ~/.config/debateos).
  --profile <name>  Translator profile name (default: foundation-specific — arch→vanilla-arch,
                    debian→debian). Explicit --profile overrides the foundation default.
  --out <path>      Output directory for resolved.json, <foundation>-profile/, and ISO (default: ./out).
  --dry-run         Print the build plan (resolved.json path, epoch, argv) but make no Runner calls.
  --skip-iso        Run translate (profile emission) then stop before the docker ISO build.
                    Use this on hosts that cannot run mkarchiso or lb build (devtmpfs restriction).

Output artifacts (when not --dry-run):
  <out>/resolved.json                Canonical resolved speech (deterministic JSON).
  <out>/<foundation>-profile/        Translator profile tree (when translate runs).
  <out>/private-injection.tar        Private pane secrets tar (local only, never in ISO).
  <out>/*.iso                        Bootable ISO (requires capable host and no --skip-iso).
`

// Run is the entry point for the "build" subcommand. It returns an exit code
// (0 = success, non-zero = failure) and never calls os.Exit.
//
// r is the Runner used for all external subprocess calls. In production main()
// passes runner.ExecRunner{}; tests pass *runner.FakeRunner.
func Run(args []string, stdout, stderr io.Writer, r runner.Runner) int {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	fs.SetOutput(stderr)

	dirFlag := fs.String("dir", "", "speech directory (overrides DEBATEOS_DIR)")
	profileFlag := fs.String("profile", "", "translator profile name (default: foundation-specific)")
	outFlag := fs.String("out", "out", "output directory")
	dryRunFlag := fs.Bool("dry-run", false, "print build plan, make no Runner calls")
	skipISOFlag := fs.Bool("skip-iso", false, "stop after profile emission, skip docker ISO build")

	if err := fs.Parse(args); err != nil {
		fmt.Fprint(stderr, buildUsage)
		return 1
	}

	// Resolve the speech directory.
	speechDir, err := resolveDir(*dirFlag)
	if err != nil {
		fmt.Fprintf(stderr, "build: %v\n", err)
		return 1
	}

	// Expand and clean the output dir.
	outDir := *outFlag
	if !filepath.IsAbs(outDir) {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(stderr, "build: getcwd: %v\n", err)
			return 1
		}
		outDir = filepath.Join(cwd, outDir)
	}

	// ── Step 1: Resolve the speech ──────────────────────────────────────────
	rs, err := loader.ResolveDir(speechDir)
	if err != nil {
		fmt.Fprintf(stderr, "build: resolve: %v\n", err)
		return 1
	}

	// ── Step 1b: Foundation registry lookup (DEB-03 / T-04-10) ─────────────
	// rs.Foundation comes from speech.yaml; the registry is a closed set of
	// compile-time constants (T-04-11: no injection surface from speech data).
	fc, ok := foundationRegistry[rs.Foundation]
	if !ok {
		fmt.Fprintf(stderr, "build: unknown foundation %q — no translator registered\n", rs.Foundation)
		return 1
	}

	// Resolve effective profile: explicit --profile overrides the foundation default.
	effectiveProfile := *profileFlag
	if effectiveProfile == "" {
		effectiveProfile = fc.DefaultProfile
	}

	// ── Step 2: Write canonical resolved.json ───────────────────────────────
	canonicalBytes, err := resolve.CanonicalJSON(rs)
	if err != nil {
		fmt.Fprintf(stderr, "build: canonical JSON: %v\n", err)
		return 1
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(stderr, "build: mkdir %s: %v\n", outDir, err)
		return 1
	}

	resolvedJSONPath := filepath.Join(outDir, "resolved.json")
	if err := os.WriteFile(resolvedJSONPath, canonicalBytes, 0644); err != nil {
		fmt.Fprintf(stderr, "build: write resolved.json: %v\n", err)
		return 1
	}

	// ── Step 3: Derive SOURCE_DATE_EPOCH once ───────────────────────────────
	// Single derivation point (T-03-EPOCH): same bytes → same epoch.
	// Mirror manifest.py derive_source_date_epoch exactly.
	epoch := DeriveEpoch(canonicalBytes)
	epochEnv := fmt.Sprintf("SOURCE_DATE_EPOCH=%d", epoch)

	// ── Build argv ──────────────────────────────────────────────────────────
	// translateBin and profileDir come from the foundationRegistry (no hardcoding).
	translateBin := fc.TranslateBin
	profileDir := filepath.Join(outDir, fc.ProfileDir)
	opinionsDir := filepath.Join(speechDir, "opinions")

	// FROZEN argv contract (translator entrypoints, CONTEXT.md):
	//   translate <resolved.json> --opinions <path> --profile <name> --out <dir>
	translateArgs := []string{
		resolvedJSONPath,
		"--opinions", opinionsDir,
		"--profile", effectiveProfile,
		"--out", profileDir,
	}

	// Docker argv (T-03-DKARG: variadic, no sh -c).
	// Volumes: speech dir at /speech (read-only — WR-07: container must not
	// modify the user's speech dir), out dir at /out (read-write for artifacts).
	// SOURCE_DATE_EPOCH passed as -e flag.
	dockerArgs := []string{
		"run",
		"-v", speechDir + ":/speech:ro",
		"-v", outDir + ":/out",
		"-e", epochEnv,
		dockerImage,
	}

	// ── Step 4: --dry-run gate ──────────────────────────────────────────────
	if *dryRunFlag {
		fmt.Fprintf(stdout, "build plan:\n")
		fmt.Fprintf(stdout, "  resolved.json: %s\n", resolvedJSONPath)
		fmt.Fprintf(stdout, "  %s: %d\n", "SOURCE_DATE_EPOCH", epoch)
		fmt.Fprintf(stdout, "  translate argv: %s %s\n", translateBin, joinArgs(translateArgs))
		fmt.Fprintf(stdout, "  docker argv: docker %s\n", joinArgs(dockerArgs))
		return 0
	}

	// ── Step 5: Invoke translate via Runner ─────────────────────────────────
	if err := r.Run(translateBin, translateArgs...); err != nil {
		fmt.Fprintf(stderr, "build: translate: %v\n", err)
		return 1
	}

	// ── Load private pane from config dir (CR-03 / PRIV-01) ────────────────
	// pane.yaml lives only in the XDG config dir (never the speech dir);
	// its key→value entries are injected as private file assets at
	// etc/debateos/<key> inside private-injection.tar.
	// Absence of pane.yaml is not an error — emit an empty (manifest-only) tar.
	paneAssets := loadPaneAssets(stderr)

	// Emit private-injection.tar next to the output artifacts (T-03-LEAK).
	// The tar carries private pane assets (if any); the first-boot unit
	// finds this artifact on mounted removable media and applies it.
	if _, tarErr := WriteInjectionTar(outDir, paneAssets); tarErr != nil {
		fmt.Fprintf(stderr, "build: injection tar: %v\n", tarErr)
		return 1
	}

	// ── Step 6: --skip-iso gate ─────────────────────────────────────────────
	if *skipISOFlag {
		fmt.Fprintf(stdout, "build: profile emitted to %s (--skip-iso: stopping before docker)\n", profileDir)
		return 0
	}

	// ── Step 7: Invoke docker via Runner ────────────────────────────────────
	if err := r.Run("docker", dockerArgs...); err != nil {
		fmt.Fprintf(stderr, "build: docker: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "build: ISO written to %s\n", outDir)
	return 0
}

// DeriveEpoch computes SOURCE_DATE_EPOCH from canonical resolved.json bytes.
//
// Algorithm mirrors translators/arch/manifest.py derive_source_date_epoch
// exactly (BLD-03 / T-03-EPOCH consistency assert):
//  1. SHA-256 hash of content_bytes.
//  2. First 4 bytes interpreted as big-endian uint32 (raw).
//  3. epochMin + (raw % (epochMax - epochMin)).
//
// The result is always in [epochMin, epochMax) = [2020-01-01, 2040-01-01).
func DeriveEpoch(contentBytes []byte) int64 {
	digest := sha256.Sum256(contentBytes)
	raw := binary.BigEndian.Uint32(digest[:4])
	return epochMin + (int64(raw) % (epochMax - epochMin))
}

// resolveDir returns dirFlag if non-empty, otherwise delegates to config.DebateOSDir().
func resolveDir(dirFlag string) (string, error) {
	if dirFlag != "" {
		return dirFlag, nil
	}
	return config.DebateOSDir()
}

// loadPaneAssets reads pane.yaml from the DebateOS config dir and converts
// each key→value entry into a PaneAsset stored at etc/debateos/<key> (0600).
//
// Absence of pane.yaml is not an error — returns nil.
// On any other error, a warning is printed to stderr and nil is returned so
// the build continues (degraded but not broken).
//
// Security (PRIV-01 / T-03-LEAK): pane.yaml never leaves the local machine;
// assets are stored in the injection tar only, not in the arch-profile tree.
func loadPaneAssets(stderr io.Writer) []PaneAsset {
	configDir, err := config.DebateOSDir()
	if err != nil {
		// Config dir unavailable — skip pane merge silently.
		return nil
	}

	paneYAMLPath := filepath.Join(configDir, "pane.yaml")
	data, err := os.ReadFile(paneYAMLPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		fmt.Fprintf(stderr, "build: warning: could not read pane.yaml (%v) — private pane will not be injected\n", err)
		return nil
	}

	var paneData map[string]string
	if err := yaml.Unmarshal(data, &paneData); err != nil {
		fmt.Fprintf(stderr, "build: warning: could not parse pane.yaml (%v) — private pane will not be injected\n", err)
		return nil
	}

	// Convert each pane entry to an asset stored at etc/debateos/<key>.
	// Use 0600 so first-boot tooling applies restrictive permissions.
	assets := make([]PaneAsset, 0, len(paneData))
	for key, value := range paneData {
		assets = append(assets, PaneAsset{
			Dst:     filepath.Join("etc", "debateos", key),
			Content: []byte(value),
			Mode:    0600,
		})
	}
	return assets
}

// joinArgs joins a string slice with spaces for display purposes only.
// This must NEVER be passed to a shell — Runner.Run takes variadic args.
func joinArgs(args []string) string {
	out := ""
	for i, a := range args {
		if i > 0 {
			out += " "
		}
		out += a
	}
	return out
}
