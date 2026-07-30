package httpserver

import (
	"context"
	"errors"
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

var publicPathPrefixes = []string{"/auth/", "/ui/"}

var publicExactPaths = map[string]struct{}{
	"/healthz": {},
	"/metrics": {},
}

const adminPathPrefix = "/v1/admin/"

// NewAuthMiddleware enforces session authentication on /v1/* and /ws/* routes.
// It always strips client-provided X-User-* headers and re-populates them from
// the session so downstream controllers keep working unchanged.
func NewAuthMiddleware(resolver SessionResolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Del("X-User-ID")
			r.Header.Del("X-User-Name")
			r.Header.Del("X-User-Email")

			session, authenticated := resolveSession(resolver, r)
			if authenticated {
				r.Header.Set("X-User-ID", session.UserID.String())
				r.Header.Set("X-User-Name", session.Name)
				r.Header.Set("X-User-Email", session.Email)

				span := GetSpanFromContext(r)
				span.SetAttributes(
					attribute.String("user.id", session.UserID.String()),
					attribute.String("user.name", session.Name),
					attribute.String("user.email", session.Email),
				)
			}

			if isPublicPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			if !authenticated {
				ReplyWithError(w, http.StatusUnauthorized, "authentication required")
				return
			}

			if strings.HasPrefix(r.URL.Path, adminPathPrefix) && !session.IsAdmin {
				ReplyWithError(w, http.StatusForbidden, "admin access required")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
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
