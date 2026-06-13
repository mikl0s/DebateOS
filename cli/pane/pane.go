package pane

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/mikl0s/debateos/cli/config"
	"github.com/mikl0s/debateos/cli/runner"
	"go.yaml.in/yaml/v3"
)

const paneFile = "pane.yaml"
const paneAgeFile = "pane.yaml.age"

const paneUsage = `usage: debateos pane <verb> [flags] [args]

Verbs:
  set <key> <value>  Write/update a private pane entry.
  get <key>          Print the value of a private pane entry.
  list               Print all private pane entry keys.
  backup             Age-encrypt pane.yaml and git commit/push the backup.
  restore            Git pull then age-decrypt pane.yaml.age back to pane.yaml.

Flags:
  --dir <path>  Override config directory (default: DEBATEOS_DIR or ~/.config/debateos).

Key-management note: the age X25519 identity is stored at identity.age (0600)
in the config dir. Losing identity.age means losing the ability to decrypt backups.
No escrow — local-only by design (PRIV-01/D16).
`

// paneData is the in-memory representation of pane.yaml.
type paneData map[string]string

// Run is the entry point for the "pane" subcommand. It returns an exit code
// (0 = success, non-zero = failure) and never calls os.Exit directly.
//
// r is the Runner used for git subprocess calls. In production main() passes
// runner.ExecRunner{}; tests pass *runner.FakeRunner.
func Run(args []string, stdout, stderr io.Writer, r runner.Runner) int {
	fs := flag.NewFlagSet("pane", flag.ContinueOnError)
	fs.SetOutput(stderr)
	dirFlag := fs.String("dir", "", "override config directory")

	// Parse only the flags; stop at first non-flag argument (verb).
	// We must parse flags that may appear before or after the verb.
	// Strategy: if args[0] looks like a verb, skip it for flag parsing.
	if err := fs.Parse(args); err != nil {
		fmt.Fprint(stderr, paneUsage)
		return 1
	}

	rest := fs.Args()
	if len(rest) == 0 {
		fmt.Fprint(stderr, paneUsage)
		return 1
	}

	verb := rest[0]
	verbArgs := rest[1:]

	dir, err := resolveDir(*dirFlag)
	if err != nil {
		fmt.Fprintf(stderr, "pane: %v\n", err)
		return 1
	}

	switch verb {
	case "set":
		return cmdSet(verbArgs, dir, stdout, stderr)
	case "get":
		return cmdGet(verbArgs, dir, stdout, stderr)
	case "list":
		return cmdList(verbArgs, dir, stdout, stderr)
	case "backup":
		return cmdBackup(verbArgs, dir, stdout, stderr, r)
	case "restore":
		return cmdRestore(verbArgs, dir, stdout, stderr, r)
	default:
		fmt.Fprintf(stderr, "pane: unknown verb %q\n\n", verb)
		fmt.Fprint(stderr, paneUsage)
		return 1
	}
}

// resolveDir returns dir if set, otherwise calls config.DebateOSDir().
func resolveDir(dirFlag string) (string, error) {
	if dirFlag != "" {
		return dirFlag, nil
	}
	return config.DebateOSDir()
}

// loadPane reads pane.yaml from dir. Returns empty map if file absent.
func loadPane(dir string) (paneData, error) {
	path := filepath.Join(dir, paneFile)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return paneData{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	// WR-03 / T-03-PERM: re-check permissions on existing file.
	// An external tool (editor, backup, rsync) may restore pane.yaml with
	// wider permissions (e.g. 0644).  Detect and reject before use so the
	// user is alerted rather than silently reading a potentially-exposed file.
	info, statErr := os.Stat(path)
	if statErr == nil && info.Mode().Perm() != 0600 {
		return nil, fmt.Errorf("%s has insecure permissions %04o (want 0600); "+
			"run: chmod 0600 %s", path, info.Mode().Perm(), path)
	}
	var m paneData
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if m == nil {
		m = paneData{}
	}
	return m, nil
}

// savePane writes pane.yaml to dir with mode 0600 (T-03-PERM) atomically.
//
// WR-04: the old O_TRUNC path truncated pane.yaml before writing; a process
// kill or disk-full mid-write would leave an unrecoverable corrupt/empty file.
// The temp+rename pattern is atomic on POSIX: the file is either fully written
// (rename succeeds) or the old content is preserved (rename never happens).
func savePane(dir string, m paneData) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	out, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal pane: %w", err)
	}

	// Write to a temp file in the same directory so the eventual rename is
	// within the same filesystem (atomic on POSIX).
	tmp, err := os.CreateTemp(dir, ".pane.yaml.tmp.*")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	// Clean up the temp file on any failure path.
	defer func() { _ = os.Remove(tmpName) }()

	// Set 0600 before writing — temp file must not be readable by others even
	// during the write window (T-03-PERM / Pitfall 4).
	// os.CreateTemp uses 0600 on Linux by default; the explicit Chmod is for
	// clarity and cross-platform safety.
	if err := os.Chmod(tmpName, 0600); err != nil {
		tmp.Close()
		return fmt.Errorf("chmod temp: %w", err)
	}
	if _, err := tmp.Write(out); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}

	// Atomic rename: replaces the destination only when the write is complete.
	return os.Rename(tmpName, filepath.Join(dir, paneFile))
}

// cmdSet handles: pane set <key> <value>
func cmdSet(args []string, dir string, stdout, stderr io.Writer) int {
	if len(args) != 2 {
		fmt.Fprintf(stderr, "pane set: requires <key> <value>\n")
		return 1
	}
	key, value := args[0], args[1]

	m, err := loadPane(dir)
	if err != nil {
		fmt.Fprintf(stderr, "pane set: %v\n", err)
		return 1
	}
	m[key] = value

	if err := savePane(dir, m); err != nil {
		fmt.Fprintf(stderr, "pane set: %v\n", err)
		return 1
	}
	return 0
}

// cmdGet handles: pane get <key>
func cmdGet(args []string, dir string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintf(stderr, "pane get: requires <key>\n")
		return 1
	}
	key := args[0]

	m, err := loadPane(dir)
	if err != nil {
		fmt.Fprintf(stderr, "pane get: %v\n", err)
		return 1
	}
	v, ok := m[key]
	if !ok {
		fmt.Fprintf(stderr, "pane get: key %q not found\n", key)
		return 1
	}
	fmt.Fprintln(stdout, v)
	return 0
}

// cmdList handles: pane list
func cmdList(args []string, dir string, stdout, stderr io.Writer) int {
	m, err := loadPane(dir)
	if err != nil {
		fmt.Fprintf(stderr, "pane list: %v\n", err)
		return 1
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Fprintln(stdout, k)
	}
	return 0
}

// cmdBackup handles: pane backup
//
// Security (T-03-PLAINTEXT): only pane.yaml.age (ciphertext) is staged/committed/pushed.
// plaintext pane.yaml is never added to git.
func cmdBackup(args []string, dir string, stdout, stderr io.Writer, r runner.Runner) int {
	// Ensure identity exists.
	id, err := LoadOrCreateIdentity(dir)
	if err != nil {
		fmt.Fprintf(stderr, "pane backup: identity: %v\n", err)
		return 1
	}

	src := filepath.Join(dir, paneFile)
	dst := filepath.Join(dir, paneAgeFile)

	if err := EncryptFile(id, src, dst); err != nil {
		fmt.Fprintf(stderr, "pane backup: encrypt: %v\n", err)
		return 1
	}

	// git add <dir>/pane.yaml.age
	// Security (T-03-GITARG): variadic args, never sh -c.
	if err := r.Run("git", "add", dst); err != nil {
		fmt.Fprintf(stderr, "pane backup: git add: %v\n", err)
		return 1
	}

	if err := r.Run("git", "commit", "-m", "backup: update pane.yaml.age"); err != nil {
		fmt.Fprintf(stderr, "pane backup: git commit: %v\n", err)
		return 1
	}

	if err := r.Run("git", "push"); err != nil {
		fmt.Fprintf(stderr, "pane backup: git push: %v\n", err)
		return 1
	}

	fmt.Fprintln(stdout, "backup: pane.yaml.age committed and pushed")
	return 0
}

// cmdRestore handles: pane restore
func cmdRestore(args []string, dir string, stdout, stderr io.Writer, r runner.Runner) int {
	// Pull latest from the user's private repo.
	if err := r.Run("git", "pull"); err != nil {
		fmt.Fprintf(stderr, "pane restore: git pull: %v\n", err)
		return 1
	}

	// Ensure identity exists.
	id, err := LoadOrCreateIdentity(dir)
	if err != nil {
		fmt.Fprintf(stderr, "pane restore: identity: %v\n", err)
		return 1
	}

	src := filepath.Join(dir, paneAgeFile)
	dst := filepath.Join(dir, paneFile)

	if err := DecryptFile(id, src, dst); err != nil {
		fmt.Fprintf(stderr, "pane restore: decrypt: %v\n", err)
		return 1
	}

	fmt.Fprintln(stdout, "restore: pane.yaml restored")
	return 0
}
