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

const buildUsage = `usage: debateos build [flags]

Flags:
  --dir <path>      Speech directory (default: DEBATEOS_DIR or ~/.config/debateos).
  --profile <name>  Translator profile name (default: vanilla-arch).
  --out <path>      Output directory for resolved.json, arch-profile/, and ISO (default: ./out).
  --dry-run         Print the build plan (resolved.json path, epoch, argv) but make no Runner calls.
  --skip-iso        Run translate (profile emission) then stop before the docker ISO build.
                    Use this on hosts that cannot run mkarchiso (devtmpfs restriction).

Output artifacts (when not --dry-run):
  <out>/resolved.json          Canonical resolved speech (deterministic JSON).
  <out>/arch-profile/          Arch translator profile tree (when translate runs).
  <out>/private-injection.tar  Private pane secrets tar (local only, never in ISO).
  <out>/*.iso                  Bootable ISO (requires capable host and no --skip-iso).
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
	profileFlag := fs.String("profile", "vanilla-arch", "translator profile name")
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
		cwd, _ := os.Getwd()
		outDir = filepath.Join(cwd, outDir)
	}

	// ── Step 1: Resolve the speech ──────────────────────────────────────────
	rs, err := loader.ResolveDir(speechDir)
	if err != nil {
		fmt.Fprintf(stderr, "build: resolve: %v\n", err)
		return 1
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
	profileDir := filepath.Join(outDir, "arch-profile")
	opinionsDir := filepath.Join(speechDir, "opinions")

	// FROZEN argv contract (translators/arch/translate header, 02-CONTEXT.md):
	//   translate <resolved.json> --opinions <path> --profile <name> --out <dir>
	translateBin := "translators/arch/translate"
	translateArgs := []string{
		resolvedJSONPath,
		"--opinions", opinionsDir,
		"--profile", *profileFlag,
		"--out", profileDir,
	}

	// Docker argv (T-03-DKARG: variadic, no sh -c).
	// Volumes: speech dir at /speech, out dir at /out.
	// SOURCE_DATE_EPOCH passed as -e flag.
	dockerArgs := []string{
		"run",
		"-v", speechDir + ":/speech",
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

	// Emit private-injection.tar next to the output artifacts (T-03-LEAK).
	// No private-pane assets to inject in the base implementation; callers
	// with a pane.yaml may pass assets. For now emit an empty tar so the
	// first-boot unit can always find the artifact.
	// (The manifest still provides version + created fields.)
	if _, tarErr := WriteInjectionTar(outDir, nil); tarErr != nil {
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
