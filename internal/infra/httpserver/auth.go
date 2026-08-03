package httpserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"zensor-server/internal/shared_kernel/domain"

	"go.opentelemetry.io/otel/attribute"
)

// SessionCookieName is the HttpOnly cookie carrying the authenticated session ID.
const SessionCookieName = "zensor_session"

// ErrNoSession signals the resolver could not find a valid session for the given ID.
var ErrNoSession = errors.New("no valid session")

// SessionResolver looks up a session by ID; implementations must return an error
// for missing, expired, or revoked sessions.
type SessionResolver interface {
	GetSession(ctx context.Context, sessionID string) (domain.Session, error)
}

// APIKeyResolver validates a plaintext bearer API key; implementations must
// return an error for unknown or revoked keys.
type APIKeyResolver interface {
	Validate(ctx context.Context, rawKey string) (domain.APIKey, error)
}

var publicPathPrefixes = []string{"/auth/", "/ui/"}

var publicExactPaths = map[string]struct{}{
	"/healthz": {},
	"/metrics": {},
}

const adminPathPrefix = "/v1/admin/"

// NewAuthMiddleware enforces authentication on /v1/* and /ws/* routes, first
// via session cookie and otherwise via bearer API key. It always strips
// client-provided X-User-* headers and re-populates them from the resolved
// identity so downstream controllers keep working unchanged.
func NewAuthMiddleware(sessions SessionResolver, apiKeys APIKeyResolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Del("X-User-ID")
			r.Header.Del("X-User-Name")
			r.Header.Del("X-User-Email")

			session, sessionAuthenticated := resolveSession(sessions, r)
			var apiKey domain.APIKey
			apiKeyAuthenticated := false
			if !sessionAuthenticated {
				apiKey, apiKeyAuthenticated = resolveAPIKey(apiKeys, r)
			}

			span := GetSpanFromContext(r)
			switch {
			case sessionAuthenticated:
				r.Header.Set("X-User-ID", session.UserID.String())
				r.Header.Set("X-User-Name", session.Name)
				r.Header.Set("X-User-Email", session.Email)

				span.SetAttributes(
					attribute.String("user.id", session.UserID.String()),
					attribute.String("user.name", session.Name),
					attribute.String("user.email", session.Email),
				)
			case apiKeyAuthenticated:
				r.Header.Set("X-User-ID", apiKey.ID.String())
				r.Header.Set("X-User-Name", apiKey.Name)
				r.Header.Set("X-User-Email", "")

				span.SetAttributes(
					attribute.String("user.id", apiKey.ID.String()),
					attribute.String("user.name", apiKey.Name),
					attribute.String("user.email", ""),
				)
			}

			if isPublicPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			if !sessionAuthenticated && !apiKeyAuthenticated {
				ReplyWithError(w, http.StatusUnauthorized, "authentication required")
				return
			}

			if strings.HasPrefix(r.URL.Path, adminPathPrefix) {
				if apiKeyAuthenticated {
					ReplyWithError(w, http.StatusForbidden, "admin access required")
					return
				}
				if !session.IsAdmin {
					ReplyWithError(w, http.StatusForbidden, "admin access required")
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func resolveAPIKey(resolver APIKeyResolver, r *http.Request) (domain.APIKey, bool) {
	const bearerPrefix = "Bearer "

	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, bearerPrefix) {
		return domain.APIKey{}, false
	}

	rawKey := strings.TrimSpace(strings.TrimPrefix(header, bearerPrefix))
	if rawKey == "" {
		return domain.APIKey{}, false
	}

	key, err := resolver.Validate(r.Context(), rawKey)
	if err != nil {
		slog.Warn("api key validation failed", slog.String("error", err.Error()))
		return domain.APIKey{}, false
	}

	return key, true
}

func resolveSession(resolver SessionResolver, r *http.Request) (domain.Session, bool) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil || cookie.Value == "" {
		return domain.Session{}, false
	}

	session, err := resolver.GetSession(r.Context(), cookie.Value)
	if err != nil {
		return domain.Session{}, false
	}

	return session, true
}

func isPublicPath(path string) bool {
	if _, found := publicExactPaths[path]; found {
		return true
	}

	for _, prefix := range publicPathPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}
