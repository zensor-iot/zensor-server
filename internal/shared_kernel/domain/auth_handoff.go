package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"
)

// AuthClient is which client started an OAuth login. It decides how the
// callback hands the session over: the web UI receives an HttpOnly cookie, the
// mobile app a single-use code it trades for a bearer token.
type AuthClient string

const (
	AuthClientWeb    AuthClient = "web"
	AuthClientMobile AuthClient = "mobile"
)

// ParseAuthClient reads the client from a query parameter or cookie. Anything
// unrecognised is the web UI, which follows a plain link and never names itself.
func ParseAuthClient(raw string) AuthClient {
	if AuthClient(raw) == AuthClientMobile {
		return AuthClientMobile
	}
	return AuthClientWeb
}

const (
	authHandoffCodeBytes = 32
	// AuthHandoffTTL is short on purpose: the app is waiting for the redirect
	// and exchanges the code right away.
	AuthHandoffTTL = 2 * time.Minute
)

// AuthHandoff is what a mobile login leaves behind for the app to collect with
// its single-use code.
//
// The session token never travels in the redirect: any app on the device can
// claim a custom URL scheme, so the redirect only carries a code worth one
// exchange. It holds the user rather than a session, so the session is minted
// at the exchange and an abandoned login leaves no live session behind.
type AuthHandoff struct {
	UserID    ID        `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	IsAdmin   bool      `json:"is_admin"`
	ExpiresAt time.Time `json:"expires_at"`
}

// IsExpired reports whether the handoff has passed its expiry at the given time.
func (h AuthHandoff) IsExpired(now time.Time) bool {
	return !now.Before(h.ExpiresAt)
}

// NewAuthHandoffCode returns a random single-use code for the redirect to the app.
func NewAuthHandoffCode() (string, error) {
	buffer := make([]byte, authHandoffCodeBytes)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("reading random bytes: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

// HashAuthHandoffCode is the key a handoff is stored under; the plaintext code
// exists only in the redirect that carried it.
func HashAuthHandoffCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}
