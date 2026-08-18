package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"zensor-server/internal/shared_kernel/usecases"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const googleUserInfoURL = "https://openidconnect.googleapis.com/v1/userinfo"

var ErrUserInfoRequestFailed = errors.New("userinfo request failed")

type GoogleOAuthProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewGoogleOAuthProvider(config GoogleOAuthProviderConfig) *GoogleOAuthProvider {
	return &GoogleOAuthProvider{
		config: &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RedirectURL:  config.RedirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
		userInfoURL: googleUserInfoURL,
	}
}

var _ usecases.OAuthProvider = (*GoogleOAuthProvider)(nil)

// GoogleOAuthProvider implements the authorization-code flow against Google.
type GoogleOAuthProvider struct {
	config      *oauth2.Config
	userInfoURL string
}

func (p *GoogleOAuthProvider) AuthCodeURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func (p *GoogleOAuthProvider) ExchangeCode(ctx context.Context, code string) (usecases.OAuthIdentity, error) {
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return usecases.OAuthIdentity{}, fmt.Errorf("exchanging code: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, p.userInfoURL, nil)
	if err != nil {
		return usecases.OAuthIdentity{}, fmt.Errorf("building userinfo request: %w", err)
	}

	response, err := p.config.Client(ctx, token).Do(request)
	if err != nil {
		return usecases.OAuthIdentity{}, fmt.Errorf("fetching userinfo: %w", err)
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			slog.Warn("failed to close userinfo response body", slog.String("error", err.Error()))
		}
	}()

	if response.StatusCode != http.StatusOK {
		return usecases.OAuthIdentity{}, fmt.Errorf("%w: status %d", ErrUserInfoRequestFailed, response.StatusCode)
	}

	var userInfo struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
	}
	if err := json.NewDecoder(response.Body).Decode(&userInfo); err != nil {
		return usecases.OAuthIdentity{}, fmt.Errorf("decoding userinfo: %w", err)
	}

	return usecases.OAuthIdentity{
		Email:         userInfo.Email,
		Name:          userInfo.Name,
		EmailVerified: userInfo.EmailVerified,
	}, nil
}
