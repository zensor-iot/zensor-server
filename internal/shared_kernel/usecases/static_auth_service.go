package usecases

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"zensor-server/internal/shared_kernel/domain"
)

// staticAdminUserID identifies the single admin user issued by static auth mode.
const staticAdminUserID domain.ID = "static-admin"

func NewStaticAuthService(sessions SessionStore, sessionTTL time.Duration, username, password string) *SimpleStaticAuthService {
	return &SimpleStaticAuthService{
		sessions:   sessions,
		sessionTTL: sessionTTL,
		username:   username,
		password:   password,
	}
}

var _ StaticAuthService = (*SimpleStaticAuthService)(nil)

// SimpleStaticAuthService authenticates a single hardcoded admin user, configured
// via ZENSOR_SERVER_AUTH_STATIC_USERNAME/PASSWORD. Local dev only, never production.
type SimpleStaticAuthService struct {
	sessions   SessionStore
	sessionTTL time.Duration
	username   string
	password   string
}

func (s *SimpleStaticAuthService) Login(ctx context.Context, username, password string) (domain.Session, error) {
	usernameMatches := subtle.ConstantTimeCompare([]byte(username), []byte(s.username)) == 1
	passwordMatches := subtle.ConstantTimeCompare([]byte(password), []byte(s.password)) == 1
	if !usernameMatches || !passwordMatches {
		return domain.Session{}, ErrInvalidCredentials
	}

	sessionID, err := generateSessionID()
	if err != nil {
		return domain.Session{}, fmt.Errorf("generating session id: %w", err)
	}

	now := time.Now()
	session := domain.Session{
		ID:        sessionID,
		UserID:    staticAdminUserID,
		Email:     "admin@localhost",
		Name:      "Admin",
		IsAdmin:   true,
		CreatedAt: now,
		ExpiresAt: now.Add(s.sessionTTL),
	}

	if err := s.sessions.Create(ctx, session); err != nil {
		return domain.Session{}, fmt.Errorf("creating session: %w", err)
	}

	slog.Info("static admin user logged in")

	return session, nil
}

func (s *SimpleStaticAuthService) GetSession(ctx context.Context, sessionID string) (domain.Session, error) {
	session, err := s.sessions.Get(ctx, sessionID)
	if errors.Is(err, ErrSessionNotFound) {
		return domain.Session{}, ErrSessionNotFound
	}
	if err != nil {
		return domain.Session{}, fmt.Errorf("getting session: %w", err)
	}

	return session, nil
}

func (s *SimpleStaticAuthService) Logout(ctx context.Context, sessionID string) error {
	if err := s.sessions.Delete(ctx, sessionID); err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}

	return nil
}
