package domain

import "time"

// Session is an authenticated browser session backed by the session store.
type Session struct {
	ID        string
	UserID    ID
	Email     string
	Name      string
	IsAdmin   bool
	CreatedAt time.Time
	ExpiresAt time.Time
}

// IsExpired reports whether the session has passed its expiry at the given time.
func (s Session) IsExpired(now time.Time) bool {
	return now.After(s.ExpiresAt)
}
