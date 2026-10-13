package usecases

import (
	"context"
	"errors"
	"zensor-server/internal/shared_kernel/domain"
)

//go:generate mockgen -source=auth_ports.go -destination=../../../test/unit/doubles/shared_kernel/usecases/auth_ports_mock.go -package=usecases -mock_names=AllowedUserRepository=MockAllowedUserRepository,SessionStore=MockSessionStore,AuthHandoffStore=MockAuthHandoffStore,OAuthProvider=MockOAuthProvider,AuthService=MockAuthService

var (
	ErrAllowedUserNotFound   = errors.New("allowed user not found")
	ErrAllowedUserDuplicated = errors.New("allowed user already exists")
	ErrEmailNotAllowed       = errors.New("email is not allowed")
	ErrEmailNotVerified      = errors.New("email is not verified")
	ErrSessionNotFound       = errors.New("session not found")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrAuthHandoffNotFound   = errors.New("auth handoff not found")
)

// OAuthIdentity is the identity obtained from the OAuth provider after code exchange.
type OAuthIdentity struct {
	Email         string
	Name          string
	EmailVerified bool
}

// OAuthProvider abstracts the OAuth2 authorization-code flow against an identity provider.
type OAuthProvider interface {
	AuthCodeURL(state string) string
	ExchangeCode(ctx context.Context, code string) (OAuthIdentity, error)
}

// SessionStore persists authenticated sessions, indexed by session ID and user ID.
type SessionStore interface {
	Create(ctx context.Context, session domain.Session) error
	Get(ctx context.Context, sessionID string) (domain.Session, error)
	Delete(ctx context.Context, sessionID string) error
	DeleteByUser(ctx context.Context, userID domain.ID) error
}

// AuthHandoffStore keeps the handoffs mobile logins leave for the app, keyed by
// the hash of their single-use code.
type AuthHandoffStore interface {
	Save(ctx context.Context, codeHash string, handoff domain.AuthHandoff) error
	// Consume returns the handoff and removes it, so a code works only once.
	Consume(ctx context.Context, codeHash string) (domain.AuthHandoff, error)
}

type AllowedUserRepository interface {
	Create(context.Context, domain.AllowedUser) error
	GetByID(context.Context, domain.ID) (domain.AllowedUser, error)
	GetByEmail(context.Context, string) (domain.AllowedUser, error)
	FindAll(context.Context) ([]domain.AllowedUser, error)
	Update(context.Context, domain.AllowedUser) error
	Delete(context.Context, domain.ID) error
}

type AuthService interface {
	AuthCodeURL(state string) string
	HandleCallback(ctx context.Context, code string) (domain.Session, error)
	// HandleMobileCallback authenticates like HandleCallback but, instead of a
	// session, returns a single-use code the app trades in ExchangeMobileCode.
	HandleMobileCallback(ctx context.Context, code string) (string, error)
	ExchangeMobileCode(ctx context.Context, code string) (domain.Session, error)
	GetSession(ctx context.Context, sessionID string) (domain.Session, error)
	Logout(ctx context.Context, sessionID string) error
	ListAllowedUsers(ctx context.Context) ([]domain.AllowedUser, error)
	AddAllowedUser(ctx context.Context, email string, isAdmin bool) (domain.AllowedUser, error)
	UpdateAllowedUser(ctx context.Context, id domain.ID, isAdmin bool) (domain.AllowedUser, error)
	RemoveAllowedUser(ctx context.Context, id domain.ID) error
	BootstrapAdmin(ctx context.Context, email string) error
}

// StaticAuthService authenticates a single, config-defined admin user against a
// fixed username/password pair. Intended for local dev only, never production.
type StaticAuthService interface {
	Login(ctx context.Context, username, password string) (domain.Session, error)
	GetSession(ctx context.Context, sessionID string) (domain.Session, error)
	Logout(ctx context.Context, sessionID string) error
}
