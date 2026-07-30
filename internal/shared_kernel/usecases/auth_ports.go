package usecases

import (
	"context"
	"errors"
	"zensor-server/internal/shared_kernel/domain"
)

//go:generate mockgen -source=auth_ports.go -destination=../../../test/unit/doubles/shared_kernel/usecases/auth_ports_mock.go -package=usecases -mock_names=AllowedUserRepository=MockAllowedUserRepository,SessionStore=MockSessionStore,OAuthProvider=MockOAuthProvider,AuthService=MockAuthService

var (
	ErrAllowedUserNotFound   = errors.New("allowed user not found")
	ErrAllowedUserDuplicated = errors.New("allowed user already exists")
	ErrEmailNotAllowed       = errors.New("email is not allowed")
	ErrEmailNotVerified      = errors.New("email is not verified")
	ErrSessionNotFound       = errors.New("session not found")
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
	GetSession(ctx context.Context, sessionID string) (domain.Session, error)
	Logout(ctx context.Context, sessionID string) error
	ListAllowedUsers(ctx context.Context) ([]domain.AllowedUser, error)
	AddAllowedUser(ctx context.Context, email string, isAdmin bool) (domain.AllowedUser, error)
	UpdateAllowedUser(ctx context.Context, id domain.ID, isAdmin bool) (domain.AllowedUser, error)
	RemoveAllowedUser(ctx context.Context, id domain.ID) error
	BootstrapAdmin(ctx context.Context, email string) error
}
