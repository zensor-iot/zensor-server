package wire

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"zensor-server/internal/infra/auth"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/domain"
	"zensor-server/internal/shared_kernel/httpapi"
	"zensor-server/internal/shared_kernel/persistence"
	"zensor-server/internal/shared_kernel/usecases"

	"github.com/redis/go-redis/v9"
)

const localDevEmail = "dev@localhost"

// AuthComponents bundles everything the server needs to enforce authentication.
type AuthComponents struct {
	Controller *httpapi.AuthController
	Service    usecases.AuthService
}

// InitializeAuthComponents wires the auth stack and bootstraps the admin allowlist entry.
func InitializeAuthComponents() (*AuthComponents, error) {
	appConfig := provideAppConfig()
	orm := provideDatabase(appConfig)

	repository, err := persistence.NewAllowedUserRepository(orm)
	if err != nil {
		return nil, fmt.Errorf("creating allowed user repository: %w", err)
	}

	env, ok := os.LookupEnv("ENV")
	if !ok {
		env = "production"
	}

	var sessionStore usecases.SessionStore
	var provider usecases.OAuthProvider
	if env == "local" {
		sessionStore = auth.NewMemorySessionStore()
		provider = auth.NewFakeOAuthProvider("http://localhost:3000/auth/callback")
	} else {
		client := redis.NewClient(&redis.Options{
			Addr:     appConfig.Redis.Addr,
			Password: appConfig.Redis.Password,
			DB:       appConfig.Redis.DB,
		})
		sessionStore = auth.NewRedisSessionStore(client)
		provider = auth.NewGoogleOAuthProvider(auth.GoogleOAuthProviderConfig{
			ClientID:     appConfig.Auth.Google.ClientID,
			ClientSecret: appConfig.Auth.Google.ClientSecret,
			RedirectURL:  appConfig.Auth.Google.RedirectURL,
		})
	}

	service := usecases.NewAuthService(repository, sessionStore, provider, appConfig.Auth.SessionTTL)

	ctx := context.Background()
	if err := service.BootstrapAdmin(ctx, appConfig.Auth.BootstrapAdminEmail); err != nil {
		return nil, fmt.Errorf("bootstrapping admin: %w", err)
	}

	if env == "local" {
		if err := seedLocalDevUser(ctx, repository); err != nil {
			return nil, fmt.Errorf("seeding local dev user: %w", err)
		}
	}

	return &AuthComponents{
		Controller: httpapi.NewAuthController(service),
		Service:    service,
	}, nil
}

// StaticAuthComponents bundles everything the server needs for static auth mode:
// a single hardcoded admin user, no OAuth provider or allowlist required.
type StaticAuthComponents struct {
	Controller *httpapi.StaticAuthController
	Service    usecases.StaticAuthService
}

// InitializeStaticAuthComponents wires static auth mode. Local dev only.
func InitializeStaticAuthComponents() (*StaticAuthComponents, error) {
	appConfig := provideAppConfig()

	env, ok := os.LookupEnv("ENV")
	if !ok {
		env = "production"
	}

	var sessionStore usecases.SessionStore
	if env == "local" {
		sessionStore = auth.NewMemorySessionStore()
	} else {
		client := redis.NewClient(&redis.Options{
			Addr:     appConfig.Redis.Addr,
			Password: appConfig.Redis.Password,
			DB:       appConfig.Redis.DB,
		})
		sessionStore = auth.NewRedisSessionStore(client)
	}

	service := usecases.NewStaticAuthService(
		sessionStore,
		appConfig.Auth.SessionTTL,
		appConfig.Auth.Static.Username,
		appConfig.Auth.Static.Password,
	)

	return &StaticAuthComponents{
		Controller: httpapi.NewStaticAuthController(service),
		Service:    service,
	}, nil
}

func seedLocalDevUser(ctx context.Context, repository usecases.AllowedUserRepository) error {
	_, err := repository.GetByEmail(ctx, localDevEmail)
	if err == nil {
		return nil
	}
	if !errors.Is(err, usecases.ErrAllowedUserNotFound) {
		return fmt.Errorf("checking local dev user: %w", err)
	}

	now := time.Now()
	devUser := domain.AllowedUser{
		ID:          domain.ID(utils.GenerateUUID()),
		Email:       localDevEmail,
		DisplayName: "Local Dev",
		IsAdmin:     true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return repository.Create(ctx, devUser)
}
