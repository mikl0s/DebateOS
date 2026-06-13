// Package api — OAuth flow for FORM-03 GitHub identity.
//
// Security invariants (D13, mandatory):
//   - GitHub OAuth ONLY. No native accounts, no passwords.
//   - OAuth state parameter validated via crypto/rand (T-05-13 CSRF).
//   - Access token used for user-ID lookup then DISCARDED — never stored (T-05-14).
//   - Session cookie is httpOnly + SameSite=Lax; stores an opaque session ID only.
//   - No secrets persist in SQLite (only the session ID → userID map, in-memory).
package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// OAuthProvider abstracts the GitHub OAuth web flow so tests can inject a fake
// provider with no network calls. Production uses RealGitHubOAuth.
type OAuthProvider interface {
	// AuthCodeURL returns the URL the browser should be redirected to so the
	// user can authorise the application. state is the CSRF-prevention nonce.
	AuthCodeURL(state string) string

	// Exchange converts an authorisation code into an access token.
	// The returned token string is passed to GetUserID and then DISCARDED.
	Exchange(ctx context.Context, code string) (token string, err error)

	// GetUserID fetches the GitHub user login (e.g. "octocat") using the token.
	// The token is NOT stored after this call returns.
	GetUserID(ctx context.Context, token string) (userID string, err error)
}

// RealGitHubOAuth wraps golang.org/x/oauth2 with the GitHub endpoint.
// It is the production implementation of OAuthProvider.
//
// Register at https://github.com/settings/applications/new with:
//   - Authorization callback URL: https://<your-host>/oauth/callback
//
// Supply clientID/clientSecret from GITHUB_CLIENT_ID and GITHUB_CLIENT_SECRET.
type RealGitHubOAuth struct {
	cfg *oauth2.Config
}

// NewRealGitHubOAuth creates a RealGitHubOAuth for the given client credentials.
// scope "read:user" is the minimum required to obtain the GitHub user login.
func NewRealGitHubOAuth(clientID, clientSecret, redirectURL string) *RealGitHubOAuth {
	return &RealGitHubOAuth{
		cfg: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"read:user"},
			Endpoint:     github.Endpoint,
		},
	}
}

// AuthCodeURL returns the GitHub authorization URL containing the CSRF state.
func (r *RealGitHubOAuth) AuthCodeURL(state string) string {
	return r.cfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

// Exchange converts the authorisation code into a GitHub access token.
// The returned string is the raw access token — call GetUserID then discard it.
func (r *RealGitHubOAuth) Exchange(ctx context.Context, code string) (string, error) {
	tok, err := r.cfg.Exchange(ctx, code)
	if err != nil {
		return "", err
	}
	return tok.AccessToken, nil
}

// GetUserID fetches the authenticated user's GitHub login via the /user API.
// The token is used for one request and is NOT retained.
func (r *RealGitHubOAuth) GetUserID(ctx context.Context, token string) (string, error) {
	// Use the token source to construct an authenticated client.
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(ctx, ts)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Parse only the login field — we need nothing else.
	var profile struct {
		Login string `json:"login"`
	}
	if err := jsonDecodeBody(resp.Body, &profile); err != nil {
		return "", err
	}
	return profile.Login, nil
}

// --------------------------------------------------------------------------
// Session store
// --------------------------------------------------------------------------

const (
	sessionCookieName = "forum_session"
	stateCookieName   = "oauth_state"
	sessionTTL        = 24 * time.Hour
	stateTTL          = 15 * time.Minute
)

// sessionEntry holds the GitHub user ID associated with a session token.
type sessionEntry struct {
	userID    string
	expiresAt time.Time
}

// SessionStore is an in-memory map of opaque session IDs → GitHub user IDs.
// It is the only session persistence in the Forum; no secrets are stored.
// Thread-safe. In a multi-instance deployment, replace with a Redis-backed store.
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]sessionEntry
	states   map[string]time.Time // state → expiry (short-lived CSRF nonces)
}

// NewSessionStore creates an empty SessionStore.
func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]sessionEntry),
		states:   make(map[string]time.Time),
	}
}

// createState generates a cryptographically random state nonce, stores it, and
// returns the hex-encoded string.
func (ss *SessionStore) createState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := hex.EncodeToString(b)
	ss.mu.Lock()
	ss.states[state] = time.Now().Add(stateTTL)
	ss.mu.Unlock()
	return state, nil
}

// validateState checks whether the state nonce is known and not expired.
// Consumed on first use (deleted after validation — one-time token).
func (ss *SessionStore) validateState(state string) bool {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	exp, ok := ss.states[state]
	if !ok {
		return false
	}
	delete(ss.states, state) // consume
	return time.Now().Before(exp)
}

// createSession generates an opaque session ID, maps it to userID, and returns
// the session ID. The access token is NOT stored here.
func (ss *SessionStore) createSession(userID string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	sid := hex.EncodeToString(b)
	ss.mu.Lock()
	ss.sessions[sid] = sessionEntry{userID: userID, expiresAt: time.Now().Add(sessionTTL)}
	ss.mu.Unlock()
	return sid, nil
}

// GetUserID looks up the userID for the session cookie contained in cookies.
// Returns ("", false) if no valid session is found.
func (ss *SessionStore) GetUserID(cookies []*http.Cookie) (string, bool) {
	for _, c := range cookies {
		if c.Name == sessionCookieName {
			ss.mu.RLock()
			entry, ok := ss.sessions[c.Value]
			ss.mu.RUnlock()
			if ok && time.Now().Before(entry.expiresAt) {
				return entry.userID, true
			}
		}
	}
	return "", false
}

// identityFnFromSessions returns an IdentityFn that reads the session cookie.
// This is injected at router construction time (replaces the test-only fake from 05-03).
func identityFnFromSessions(ss *SessionStore) IdentityFn {
	return func(r *http.Request) (string, bool) {
		return ss.GetUserID(r.Cookies())
	}
}

// --------------------------------------------------------------------------
// OAuth HTTP handlers
// --------------------------------------------------------------------------

// oauthHandlers wraps the provider and session store for the login/callback routes.
type oauthHandlers struct {
	provider OAuthProvider
	sessions *SessionStore
}

// loginHandler handles GET /oauth/login.
// Generates a CSRF state nonce, stores it in a short-lived httpOnly cookie,
// then redirects the browser to the provider's authorization URL.
func (o *oauthHandlers) loginHandler(w http.ResponseWriter, r *http.Request) {
	state, err := o.sessions.createState()
	if err != nil {
		http.Error(w, "internal error generating state", http.StatusInternalServerError)
		return
	}

	// Short-lived state cookie (httpOnly, SameSite=Lax per security requirements).
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    state,
		Path:     "/",
		MaxAge:   int(stateTTL.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, o.provider.AuthCodeURL(state), http.StatusFound)
}

// callbackHandler handles GET /oauth/callback.
// Validates the state parameter against the cookie (T-05-13), exchanges the code
// for a token, fetches the GitHub user ID, creates a session, then DISCARDS the
// token. The session cookie is set; the token is never persisted.
func (o *oauthHandlers) callbackHandler(w http.ResponseWriter, r *http.Request) {
	// --- State validation (T-05-13 CSRF prevention) ---
	cookieState, err := r.Cookie(stateCookieName)
	if err != nil || cookieState.Value == "" {
		http.Error(w, "missing state cookie", http.StatusBadRequest)
		return
	}
	urlState := r.URL.Query().Get("state")
	if urlState == "" || urlState != cookieState.Value {
		http.Error(w, "state mismatch: possible CSRF", http.StatusBadRequest)
		return
	}
	if !o.sessions.validateState(urlState) {
		http.Error(w, "state invalid or expired", http.StatusBadRequest)
		return
	}

	// Clear the state cookie (consumed).
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	// --- Exchange code for token ---
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code parameter", http.StatusBadRequest)
		return
	}

	token, err := o.provider.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "token exchange failed", http.StatusBadGateway)
		return
	}

	// --- Fetch GitHub user ID then DISCARD the token (T-05-14) ---
	userID, err := o.provider.GetUserID(r.Context(), token)
	// token is no longer referenced after this line; it goes out of scope
	// and is garbage collected. It is never passed to the store.
	token = "" // explicit zero to make intent clear
	_ = token

	if err != nil {
		http.Error(w, "failed to get user ID", http.StatusBadGateway)
		return
	}

	// --- Create session (stores userID only, not token) ---
	sid, err := o.sessions.createSession(userID)
	if err != nil {
		http.Error(w, "session creation failed", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sid,
		Path:     "/",
		MaxAge:   int(sessionTTL.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/", http.StatusFound)
}
