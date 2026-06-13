// Package pane manages the private pane: age X25519 identity generation,
// pane.yaml 0600 read/write, and age-encrypted backup/restore via the Runner
// interface.
//
// Key-management design (D16 / PRIV-01):
//   - age X25519 identities, local-only, no escrow, no central service.
//   - identity.age stored 0600 in the DebateOS config dir; generated on first use.
//   - Losing identity.age means losing the ability to decrypt backups — by design.
//   - Only pane.yaml.age (ciphertext) is ever committed/pushed; plaintext never staged.
package pane

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"filippo.io/age"
)

// identityFileName is the base name of the age identity file stored in the
// config dir alongside pane.yaml.
const identityFileName = "identity.age"

// LoadOrCreateIdentity returns the age X25519 identity stored at
// dir/identity.age, creating it (0600) if it does not yet exist.
//
// Security: the file is written with os.OpenFile(O_CREATE|O_WRONLY|O_TRUNC, 0600)
// so the private key is never exposed with a wider mode (T-03-PERM).
func LoadOrCreateIdentity(dir string) (*age.X25519Identity, error) {
	path := fmt.Sprintf("%s/%s", dir, identityFileName)

	// Try to read existing identity first.
	data, err := os.ReadFile(path)
	if err == nil {
		// File exists — parse and return. Strip trailing whitespace/newline that
		// fmt.Fprintln appended at write time.
		id, err := age.ParseX25519Identity(strings.TrimSpace(string(data)))
		if err != nil {
			return nil, fmt.Errorf("parse identity %s: %w", path, err)
		}
		return id, nil
	}
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read identity %s: %w", path, err)
	}

	// Generate a new X25519 identity.
	id, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, fmt.Errorf("generate identity: %w", err)
	}

	// Write 0600 — never wider (T-03-PERM / V4 file perms).
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return nil, fmt.Errorf("create identity %s: %w", path, err)
	}
	defer f.Close()

	// id.String() returns the "AGE-SECRET-KEY-1..." bech32 private key.
	if _, err := fmt.Fprintln(f, id.String()); err != nil {
		return nil, fmt.Errorf("write identity %s: %w", path, err)
	}
	return id, nil
}

// EncryptFile age-encrypts src to dst using the public key of identity.
// dst is created 0600 (T-03-PERM). The age format is streaming so large
// files do not need to be fully buffered.
//
// Security: only filippo.io/age is used — no hand-rolled crypto (V6).
func EncryptFile(id *age.X25519Identity, src, dst string) error {
	plaintext, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read %s: %w", src, err)
	}

	// Create dst 0600 before writing (T-03-PERM).
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	defer out.Close()

	w, err := age.Encrypt(out, id.Recipient())
	if err != nil {
		return fmt.Errorf("age.Encrypt: %w", err)
	}

	if _, err := w.Write(plaintext); err != nil {
		return fmt.Errorf("write encrypted: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("finalise encrypted: %w", err)
	}
	return nil
}

// DecryptFile age-decrypts src to dst using identity.
// dst is written 0600 (T-03-PERM).
func DecryptFile(id *age.X25519Identity, src, dst string) error {
	ciphertext, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read %s: %w", src, err)
	}

	r, err := age.Decrypt(bytes.NewReader(ciphertext), id)
	if err != nil {
		return fmt.Errorf("age.Decrypt: %w", err)
	}

	plaintext, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read decrypted: %w", err)
	}

	// Write 0600 (T-03-PERM).
	f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	defer f.Close()

	if _, err := f.Write(plaintext); err != nil {
		return fmt.Errorf("write %s: %w", dst, err)
	}
	return nil
}
