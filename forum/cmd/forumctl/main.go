// forumctl is the single-binary CLI for the Forum service (FORM-05).
//
// Subcommands:
//
//	serve   — Start the HTTP server with the real GitHub OAuth provider.
//	         Configuration via environment variables (no secrets in code):
//	           FORUM_ADDR            — listen address (default: :8080)
//	           FORUM_DB              — SQLite database file (default: forum.db)
//	           GITHUB_CLIENT_ID      — GitHub OAuth app client ID (required for OAuth)
//	           GITHUB_CLIENT_SECRET  — GitHub OAuth app client secret (required for OAuth)
//	           GITHUB_REDIRECT_URL   — OAuth callback URL (default: http://localhost:8080/oauth/callback)
//
//	reindex — Rebuild the SQLite index from a static registry index.json file.
//	           FORUM_DB              — SQLite database file (default: forum.db)
//	           FORUM_INDEX           — path to registry index.json (default: registry/index.json)
//
// Security (D13, mandatory):
//   - Secrets are read from environment only — never from flags or config files.
//   - OAuth token is discarded after user-ID lookup (handled in forum/api/oauth.go).
//   - Reindex is idempotent; safe to run after total DB loss.
//
// Target platforms: linux/amd64 and linux/arm64 (Oracle A1 Ampere / FORM-05).
// Build for arm64: GOOS=linux GOARCH=arm64 go build -o forumctl-arm64 ./forum/cmd/forumctl
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/mikl0s/debateos/forum"
	"github.com/mikl0s/debateos/forum/api"
	"github.com/mikl0s/debateos/forum/store"
	"github.com/mikl0s/debateos/registry/index"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "serve":
		if err := runServe(); err != nil {
			log.Fatalf("forumctl serve: %v", err)
		}
	case "reindex":
		if err := runReindex(); err != nil {
			log.Fatalf("forumctl reindex: %v", err)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %q\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `forumctl — Forum service binary (FORM-05)

Usage:
  forumctl serve    Start the HTTP server (configure via environment)
  forumctl reindex  Rebuild the SQLite index from a registry index.json

Environment variables:
  FORUM_ADDR             Listen address       (default: :8080)
  FORUM_DB               SQLite database file (default: forum.db)
  GITHUB_CLIENT_ID       GitHub OAuth app client ID
  GITHUB_CLIENT_SECRET   GitHub OAuth app client secret
  GITHUB_REDIRECT_URL    OAuth callback URL   (default: http://localhost:8080/oauth/callback)
  FORUM_INDEX            Path to registry index.json (default: registry/index.json, reindex only)

Security notes:
  - All secrets must be supplied via environment, never command-line flags.
  - OAuth tokens are discarded after user-ID lookup; never persisted.
  - Reindex is idempotent — safe to run repeatedly from the registry index.
`)
}

// runServe starts the Forum HTTP server with the real GitHub OAuth provider.
func runServe() error {
	addr := envOrDefault("FORUM_ADDR", ":8080")
	dsn := envOrDefault("FORUM_DB", "forum.db")
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	redirectURL := envOrDefault("GITHUB_REDIRECT_URL", "http://localhost:8080/oauth/callback")

	// Open (or create) the SQLite store.
	db, err := store.Open(dsn)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	s := store.New(db)
	defer s.Close()

	// Build the OAuth provider and session store.
	var provider api.OAuthProvider
	if clientID != "" && clientSecret != "" {
		provider = api.NewRealGitHubOAuth(clientID, clientSecret, redirectURL)
		log.Printf("GitHub OAuth configured (client_id: %s, redirect: %s)", clientID, redirectURL)
	} else {
		// No OAuth credentials — serve in read-only mode (OAuth routes not mounted).
		// This allows the Forum to be started for read-only browsing without GitHub credentials.
		log.Println("WARNING: GITHUB_CLIENT_ID/GITHUB_CLIENT_SECRET not set; OAuth disabled (read-only mode)")
	}

	sessions := api.NewSessionStore()

	var router http.Handler
	if provider != nil {
		router = api.NewRouterWithOAuth(s, provider, sessions)
	} else {
		// Read-only mode: mount routes without OAuth (identityFn always returns false).
		router = api.NewRouter(s, func(r *http.Request) (string, bool) { return "", false })
	}

	log.Printf("forumctl serve listening on %s (db: %s)", addr, dsn)
	return http.ListenAndServe(addr, router)
}

// runReindex rebuilds the Forum's SQLite index from a registry index.json file.
func runReindex() error {
	dsn := envOrDefault("FORUM_DB", "forum.db")
	indexPath := envOrDefault("FORUM_INDEX", "registry/index.json")

	// Load the registry index.
	f, err := os.Open(indexPath)
	if err != nil {
		return fmt.Errorf("open index %q: %w", indexPath, err)
	}
	defer f.Close()

	var idx index.RegistryIndex
	if err := json.NewDecoder(f).Decode(&idx); err != nil {
		return fmt.Errorf("decode index %q: %w", indexPath, err)
	}

	log.Printf("forumctl reindex: loaded %d points from %q (schema v%d, generated %s)",
		len(idx.Points), indexPath, idx.Schema, idx.GeneratedAt)

	// Open (or create) the SQLite store.
	db, err := store.Open(dsn)
	if err != nil {
		return fmt.Errorf("open store %q: %w", dsn, err)
	}
	s := store.New(db)
	defer s.Close()

	ctx := context.Background()
	if err := forum.Reindex(ctx, s, &idx); err != nil {
		return fmt.Errorf("Reindex: %w", err)
	}

	log.Printf("forumctl reindex: complete — %d points upserted into %q", len(idx.Points), dsn)
	return nil
}

// envOrDefault returns the environment variable value or the fallback default.
func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
