package api

// internal_test.go — white-box tests for unexported functions in package api.
// Uses package api (not package api_test) to access unexported symbols.
// Covers: jsonDecodeBody, SessionStore helpers.

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

// TestJSONDecodeBodyValid: jsonDecodeBody correctly decodes valid JSON.
func TestJSONDecodeBodyValid(t *testing.T) {
	body := `{"login":"octocat"}`
	var out struct {
		Login string `json:"login"`
	}
	if err := jsonDecodeBody(strings.NewReader(body), &out); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if out.Login != "octocat" {
		t.Errorf("expected login=octocat, got %q", out.Login)
	}
}

// TestJSONDecodeBodyInvalid: jsonDecodeBody returns error for invalid JSON.
func TestJSONDecodeBodyInvalid(t *testing.T) {
	body := `{not valid json`
	var out struct {
		Login string `json:"login"`
	}
	if err := jsonDecodeBody(strings.NewReader(body), &out); err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

// TestJSONDecodeBodyEmpty: jsonDecodeBody returns error for empty reader.
func TestJSONDecodeBodyEmpty(t *testing.T) {
	var out struct{}
	if err := jsonDecodeBody(strings.NewReader(""), &out); err == nil {
		t.Fatal("expected error for empty body, got nil")
	}
}

// TestSessionStoreCreateAndValidateState: createState generates a state nonce
// and validateState accepts it once (then rejects a second use).
func TestSessionStoreCreateAndValidateState(t *testing.T) {
	ss := NewSessionStore()
	state, err := ss.createState()
	if err != nil {
		t.Fatalf("createState: %v", err)
	}
	if state == "" {
		t.Fatal("createState returned empty state")
	}

	// Valid: first use.
	if !ss.validateState(state) {
		t.Error("validateState should return true for freshly-created state")
	}

	// Consumed: second use must fail.
	if ss.validateState(state) {
		t.Error("validateState should return false for already-consumed state")
	}
}

// TestSessionStoreValidateUnknownState: validateState rejects an unknown nonce.
func TestSessionStoreValidateUnknownState(t *testing.T) {
	ss := NewSessionStore()
	if ss.validateState("totally-unknown-nonce") {
		t.Error("validateState should return false for unknown state")
	}
}

// TestSessionStoreCreateSession: createSession and GetUserID round-trip.
func TestSessionStoreCreateSession(t *testing.T) {
	ss := NewSessionStore()
	sid, err := ss.createSession("alice")
	if err != nil {
		t.Fatalf("createSession: %v", err)
	}
	if sid == "" {
		t.Fatal("createSession returned empty session ID")
	}

	// Build a real http.Cookie slice.
	cookies := []*http.Cookie{
		{Name: sessionCookieName, Value: sid},
	}
	userID, ok := ss.GetUserID(cookies)
	if !ok {
		t.Fatal("GetUserID returned ok=false for valid session")
	}
	if userID != "alice" {
		t.Errorf("expected userID=alice, got %q", userID)
	}
}

// TestSessionStoreGetUserIDUnknown: GetUserID returns ("", false) for unknown cookie.
func TestSessionStoreGetUserIDUnknown(t *testing.T) {
	ss := NewSessionStore()
	cookies := []*http.Cookie{
		{Name: sessionCookieName, Value: "no-such-session"},
	}
	userID, ok := ss.GetUserID(cookies)
	if ok || userID != "" {
		t.Errorf("expected empty, got userID=%q ok=%v", userID, ok)
	}
}

// TestSessionStoreGetUserIDNoCookie: GetUserID with no matching cookie name.
func TestSessionStoreGetUserIDNoCookie(t *testing.T) {
	ss := NewSessionStore()
	userID, ok := ss.GetUserID(nil)
	if ok || userID != "" {
		t.Errorf("expected empty, got userID=%q ok=%v", userID, ok)
	}
}

// TestIdentityFnFromSessions: identityFnFromSessions wraps a SessionStore.
func TestIdentityFnFromSessions(t *testing.T) {
	ss := NewSessionStore()
	sid, err := ss.createSession("bob")
	if err != nil {
		t.Fatalf("createSession: %v", err)
	}

	fn := identityFnFromSessions(ss)

	// Build a fake request with the session cookie.
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sid})

	userID, ok := fn(req)
	if !ok || userID != "bob" {
		t.Errorf("expected bob ok=true, got %q ok=%v", userID, ok)
	}
}

// TestIdentityFnFromSessionsNoSession: identityFnFromSessions returns false with no cookie.
func TestIdentityFnFromSessionsNoSession(t *testing.T) {
	ss := NewSessionStore()
	fn := identityFnFromSessions(ss)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	userID, ok := fn(req)
	if ok || userID != "" {
		t.Errorf("expected empty, got %q ok=%v", userID, ok)
	}
}

// TestSweepExpiredRemovesExpiredEntries: sweepExpired removes entries past their
// expiry from both states and sessions maps (CR-02).
// Uses direct map manipulation to inject already-expired entries, then calls
// sweepExpired() synchronously — no real-time waiting required.
func TestSweepExpiredRemovesExpiredEntries(t *testing.T) {
	ss := NewSessionStore()

	// Inject an already-expired state nonce directly into the map.
	ss.mu.Lock()
	ss.states["expired-state"] = time.Now().Add(-1 * time.Hour) // already expired
	ss.states["valid-state"] = time.Now().Add(15 * time.Minute)  // still valid
	// Inject an already-expired session.
	ss.sessions["expired-session"] = sessionEntry{
		userID:    "old-user",
		expiresAt: time.Now().Add(-1 * time.Hour), // already expired
	}
	ss.sessions["valid-session"] = sessionEntry{
		userID:    "active-user",
		expiresAt: time.Now().Add(24 * time.Hour), // still valid
	}
	ss.mu.Unlock()

	// Call sweepExpired directly (deterministic — no ticker dependency).
	ss.sweepExpired()

	// Expired entries must be gone; valid entries must remain.
	ss.mu.RLock()
	_, hasExpiredState := ss.states["expired-state"]
	_, hasValidState := ss.states["valid-state"]
	_, hasExpiredSession := ss.sessions["expired-session"]
	_, hasValidSession := ss.sessions["valid-session"]
	ss.mu.RUnlock()

	if hasExpiredState {
		t.Error("expired-state was NOT removed by sweepExpired")
	}
	if !hasValidState {
		t.Error("valid-state was incorrectly removed by sweepExpired")
	}
	if hasExpiredSession {
		t.Error("expired-session was NOT removed by sweepExpired")
	}
	if !hasValidSession {
		t.Error("valid-session was incorrectly removed by sweepExpired")
	}
}
