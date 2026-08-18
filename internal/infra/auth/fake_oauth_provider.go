// Package auth provides authentication providers and session stores.
package auth

import (
	"context"
	"fmt"
	"zensor-server/internal/shared_kernel/usecases"
)

func NewFakeOAuthProvider(callbackURL string) *FakeOAuthProvider {
	return &FakeOAuthProvider{
		callbackURL: callbackURL,
	}
}

var _ usecases.OAuthProvider = (*FakeOAuthProvider)(nil)

// FakeOAuthProvider short-circuits the OAuth flow for ENV=local runs: the auth URL
// points straight back at the callback and any code resolves to a fixed dev identity.
type FakeOAuthProvider struct {
	callbackURL string
}

func (p *FakeOAuthProvider) AuthCodeURL(state string) string {
	return fmt.Sprintf("%s?code=local-dev&state=%s", p.callbackURL, state)
}

func (p *FakeOAuthProvider) ExchangeCode(_ context.Context, _ string) (usecases.OAuthIdentity, error) {
	return usecases.OAuthIdentity{
		Email:         "dev@localhost",
		Name:          "Local Dev",
		EmailVerified: true,
	}, nil
}
