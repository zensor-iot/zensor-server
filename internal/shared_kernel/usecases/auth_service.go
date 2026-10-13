package usecases

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"zensor-server/internal/shared_kernel/domain"
)

func NewAuthService(
	repository AllowedUserRepository,
	sessions SessionStore,
	handoffs AuthHandoffStore,
	provider OAuthProvider,
	sessionTTL time.Duration,
) *SimpleAuthService {
	return &SimpleAuthService{
		repository: repository,
		sessions:   sessions,
		handoffs:   handoffs,
		provider:   provider,
		sessionTTL: sessionTTL,
	}
}

var _ AuthService = (*SimpleAuthService)(nil)

type SimpleAuthService struct {
	repository AllowedUserRepository
	sessions   SessionStore
	handoffs   AuthHandoffStore
	provider   OAuthProvider
	sessionTTL time.Duration
}

func (s *SimpleAuthService) AuthCodeURL(state string) string {
	return s.provider.AuthCodeURL(state)
}

func (s *SimpleAuthService) HandleCallback(ctx context.Context, code string) (domain.Session, error) {
	user, name, err := s.authenticate(ctx, code)
	if err != nil {
		return domain.Session{}, err
	}

	return s.openSession(ctx, user.ID, user.Email, name, user.IsAdmin)
}

func (s *SimpleAuthService) HandleMobileCallback(ctx context.Context, code string) (string, error) {
	user, name, err := s.authenticate(ctx, code)
	if err != nil {
		return "", err
	}

	handoffCode, err := domain.NewAuthHandoffCode()
	if err != nil {
		return "", fmt.Errorf("generating handoff code: %w", err)
	}

	handoff := domain.AuthHandoff{
		UserID:    user.ID,
		Email:     user.Email,
		Name:      name,
		IsAdmin:   user.IsAdmin,
		ExpiresAt: time.Now().Add(domain.AuthHandoffTTL),
	}
	if err := s.handoffs.Save(ctx, domain.HashAuthHandoffCode(handoffCode), handoff); err != nil {
		return "", fmt.Errorf("saving auth handoff: %w", err)
	}

	return handoffCode, nil
}

func (s *SimpleAuthService) ExchangeMobileCode(ctx context.Context, code string) (domain.Session, error) {
	if code == "" {
		return domain.Session{}, ErrAuthHandoffNotFound
	}

	handoff, err := s.handoffs.Consume(ctx, domain.HashAuthHandoffCode(code))
	if errors.Is(err, ErrAuthHandoffNotFound) {
		return domain.Session{}, ErrAuthHandoffNotFound
	}
	if err != nil {
		return domain.Session{}, fmt.Errorf("consuming auth handoff: %w", err)
	}

	if handoff.IsExpired(time.Now()) {
		return domain.Session{}, ErrAuthHandoffNotFound
	}

	return s.openSession(ctx, handoff.UserID, handoff.Email, handoff.Name, handoff.IsAdmin)
}

// authenticate exchanges the provider code and checks the identity against the
// allowlist, recording the login. It returns the allowed user and the display
// name the provider reported.
func (s *SimpleAuthService) authenticate(ctx context.Context, code string) (domain.AllowedUser, string, error) {
	identity, err := s.provider.ExchangeCode(ctx, code)
	if err != nil {
		return domain.AllowedUser{}, "", fmt.Errorf("exchanging authorization code: %w", err)
	}

	if !identity.EmailVerified {
		return domain.AllowedUser{}, "", ErrEmailNotVerified
	}

	email := domain.NormalizeEmail(identity.Email)
	user, err := s.repository.GetByEmail(ctx, email)
	if errors.Is(err, ErrAllowedUserNotFound) {
		slog.Warn("login denied for email not on allowlist", slog.String("email", email))
		return domain.AllowedUser{}, "", ErrEmailNotAllowed
	}
	if err != nil {
		return domain.AllowedUser{}, "", fmt.Errorf("looking up allowed user: %w", err)
	}

	user.RecordLogin(identity.Name)
	if err := s.repository.Update(ctx, user); err != nil {
		return domain.AllowedUser{}, "", fmt.Errorf("recording login: %w", err)
	}

	return user, identity.Name, nil
}

func (s *SimpleAuthService) openSession(ctx context.Context, userID domain.ID, email, name string, isAdmin bool) (domain.Session, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return domain.Session{}, fmt.Errorf("generating session id: %w", err)
	}

	now := time.Now()
	session := domain.Session{
		ID:        sessionID,
		UserID:    userID,
		Email:     email,
		Name:      name,
		IsAdmin:   isAdmin,
		CreatedAt: now,
		ExpiresAt: now.Add(s.sessionTTL),
	}

	if err := s.sessions.Create(ctx, session); err != nil {
		return domain.Session{}, fmt.Errorf("creating session: %w", err)
	}

	slog.Info("user logged in", slog.String("email", email), slog.String("user_id", userID.String()))

	return session, nil
}

func (s *SimpleAuthService) GetSession(ctx context.Context, sessionID string) (domain.Session, error) {
	session, err := s.sessions.Get(ctx, sessionID)
	if errors.Is(err, ErrSessionNotFound) {
		return domain.Session{}, ErrSessionNotFound
	}
	if err != nil {
		return domain.Session{}, fmt.Errorf("getting session: %w", err)
	}

	return session, nil
}

func (s *SimpleAuthService) Logout(ctx context.Context, sessionID string) error {
	if err := s.sessions.Delete(ctx, sessionID); err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}

	return nil
}

func (s *SimpleAuthService) ListAllowedUsers(ctx context.Context) ([]domain.AllowedUser, error) {
	users, err := s.repository.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing allowed users: %w", err)
	}

	return users, nil
}

func (s *SimpleAuthService) AddAllowedUser(ctx context.Context, email string, isAdmin bool) (domain.AllowedUser, error) {
	user, err := domain.NewAllowedUserBuilder().
		WithEmail(email).
		WithIsAdmin(isAdmin).
		Build()
	if err != nil {
		return domain.AllowedUser{}, fmt.Errorf("building allowed user: %w", err)
	}

	_, err = s.repository.GetByEmail(ctx, user.Email)
	if err == nil {
		return domain.AllowedUser{}, ErrAllowedUserDuplicated
	}
	if !errors.Is(err, ErrAllowedUserNotFound) {
		return domain.AllowedUser{}, fmt.Errorf("checking existing allowed user: %w", err)
	}

	if err := s.repository.Create(ctx, user); err != nil {
		return domain.AllowedUser{}, fmt.Errorf("creating allowed user: %w", err)
	}

	slog.Info("allowed user added", slog.String("email", user.Email), slog.Bool("is_admin", user.IsAdmin))

	return user, nil
}

func (s *SimpleAuthService) UpdateAllowedUser(ctx context.Context, id domain.ID, isAdmin bool) (domain.AllowedUser, error) {
	user, err := s.repository.GetByID(ctx, id)
	if errors.Is(err, ErrAllowedUserNotFound) {
		return domain.AllowedUser{}, ErrAllowedUserNotFound
	}
	if err != nil {
		return domain.AllowedUser{}, fmt.Errorf("getting allowed user: %w", err)
	}

	user.IsAdmin = isAdmin
	user.UpdatedAt = time.Now()

	if err := s.repository.Update(ctx, user); err != nil {
		return domain.AllowedUser{}, fmt.Errorf("updating allowed user: %w", err)
	}

	return user, nil
}

func (s *SimpleAuthService) RemoveAllowedUser(ctx context.Context, id domain.ID) error {
	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting allowed user: %w", err)
	}

	if err := s.sessions.DeleteByUser(ctx, id); err != nil {
		return fmt.Errorf("revoking user sessions: %w", err)
	}

	slog.Info("allowed user removed", slog.String("user_id", id.String()))

	return nil
}

func (s *SimpleAuthService) BootstrapAdmin(ctx context.Context, email string) error {
	if email == "" {
		return nil
	}

	normalized := domain.NormalizeEmail(email)
	existing, err := s.repository.GetByEmail(ctx, normalized)
	if errors.Is(err, ErrAllowedUserNotFound) {
		user, err := domain.NewAllowedUserBuilder().
			WithEmail(normalized).
			WithIsAdmin(true).
			Build()
		if err != nil {
			return fmt.Errorf("building bootstrap admin: %w", err)
		}

		if err := s.repository.Create(ctx, user); err != nil {
			return fmt.Errorf("creating bootstrap admin: %w", err)
		}

		slog.Info("bootstrap admin created", slog.String("email", normalized))
		return nil
	}
	if err != nil {
		return fmt.Errorf("getting bootstrap admin: %w", err)
	}

	if existing.IsAdmin {
		return nil
	}

	existing.IsAdmin = true
	existing.UpdatedAt = time.Now()
	if err := s.repository.Update(ctx, existing); err != nil {
		return fmt.Errorf("promoting bootstrap admin: %w", err)
	}

	slog.Info("bootstrap admin promoted", slog.String("email", normalized))

	return nil
}

func generateSessionID() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("reading random bytes: %w", err)
	}

	return hex.EncodeToString(buffer), nil
}
