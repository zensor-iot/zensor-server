package domain

import "time"

// Session is an authenticated browser session backed by the session store.
type Session struct {
	ID        string    `json:"id"`
	UserID    ID        `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// IsExpired reports whether the session has passed its expiry at the given time.
func (s Session) IsExpired(now time.Time) bool {
	return now.After(s.ExpiresAt)
}
