// Package registry provides the static registry index generator for DebateOS.
//
// deps_guard.go anchors all Phase-5 Go dependencies so go mod tidy does not
// prune them before forum/ (05-03, 05-05) imports them. These are the ONLY
// place in Phase 5 that holds these deps as direct requires; later plans must
// NOT modify go.mod (single-owner invariant, 05-01 PLAN.md).
package registry

import (
	_ "github.com/go-chi/chi/v5"    // Forum HTTP router (05-03)
	_ "golang.org/x/oauth2"         // GitHub OAuth web flow (05-05)
	_ "modernc.org/sqlite"          // Pure-Go SQLite for Forum (05-03, 05-05)
)
