package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
	"zensor-server/internal/infra/httpserver"
	"zensor-server/internal/shared_kernel/domain"
	"zensor-server/internal/shared_kernel/httpapi/internal"
	"zensor-server/internal/shared_kernel/usecases"
)

// StateCookieName is the short-lived cookie carrying the OAuth CSRF state.
const StateCookieName = "zensor_oauth_state"

// ClientCookieName is the short-lived cookie recording which client started the
// login. The callback comes back from Google carrying nothing that says so.
const ClientCookieName = "zensor_oauth_client"

const (
	stateCookieTTL          = 10 * time.Minute
	accessDeniedRedirectURL = "/ui/access-denied"
	appRedirectURL          = "/ui/"
)

// NewAuthController builds the Google OAuth controller. mobileRedirectURI is
// where the mobile app is sent back to after login (e.g. zensor://auth/callback);
// it comes from configuration and is never read from the request, so no other
// app can ask to receive somebody else's login. Empty turns mobile login off.
func NewAuthController(service usecases.AuthService, mobileRedirectURI string) *AuthController {
	return &AuthController{
		service:           service,
		mobileRedirectURI: strings.TrimSpace(mobileRedirectURI),
	}
}

var _ httpserver.Controller = &AuthController{}

type AuthController struct {
	service           usecases.AuthService
	mobileRedirectURI string
}

func (c *AuthController) AddRoutes(router *http.ServeMux) {
	router.Handle("GET /auth/mode", c.mode())
	router.Handle("GET /auth/login", c.login())
	router.Handle("GET /auth/callback", c.callback())
	router.Handle("POST /auth/mobile/exchange", c.exchangeMobileCode())
	router.Handle("POST /auth/logout", c.logout())
	router.Handle("GET /v1/me", c.currentUser())
	router.Handle("GET /v1/admin/allowed-users", c.listAllowedUsers())
	router.Handle("POST /v1/admin/allowed-users", c.addAllowedUser())
	router.Handle("PUT /v1/admin/allowed-users/{id}", c.updateAllowedUser())
	router.Handle("DELETE /v1/admin/allowed-users/{id}", c.removeAllowedUser())
}

func (c *AuthController) mode() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpserver.ReplyJSONResponse(w, http.StatusOK, internal.AuthModeResponse{Mode: "google"})
	}
}

func (c *AuthController) login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		client := domain.ParseAuthClient(r.URL.Query().Get("client"))
		if client == domain.AuthClientMobile && c.mobileRedirectURI == "" {
			http.Error(w, "mobile login is not configured", http.StatusServiceUnavailable)
			return
		}

		state, err := generateState()
		if err != nil {
			slog.Error("generating oauth state", slog.String("error", err.Error()))
			http.Error(w, "failed to start login", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{ // #nosec G124 -- Secure is set conditionally on TLS/X-Forwarded-Proto
			Name:     StateCookieName,
			Value:    state,
			Path:     "/",
			MaxAge:   int(stateCookieTTL.Seconds()),
			HttpOnly: true,
			Secure:   isRequestSecure(r),
			SameSite: http.SameSiteLaxMode,
		})

		http.SetCookie(w, &http.Cookie{ // #nosec G124 -- Secure is set conditionally on TLS/X-Forwarded-Proto
			Name:     ClientCookieName,
			Value:    string(client),
			Path:     "/",
			MaxAge:   int(stateCookieTTL.Seconds()),
			HttpOnly: true,
			Secure:   isRequestSecure(r),
			SameSite: http.SameSiteLaxMode,
		})

		http.Redirect(w, r, c.service.AuthCodeURL(state), http.StatusFound)
	}
}

func (c *AuthController) callback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		client := domain.AuthClientWeb
		if clientCookie, err := r.Cookie(ClientCookieName); err == nil {
			client = domain.ParseAuthClient(clientCookie.Value)
		}
		clearCookie(w, ClientCookieName, isRequestSecure(r))
		if client == domain.AuthClientMobile && c.mobileRedirectURI == "" {
			client = domain.AuthClientWeb
		}

		stateCookie, err := r.Cookie(StateCookieName)
		if err != nil || stateCookie.Value == "" {
			slog.Warn("oauth callback without state cookie")
			c.fail(w, r, client, http.StatusBadRequest, "invalid_state", "invalid oauth state")
			return
		}

		clearCookie(w, StateCookieName, isRequestSecure(r))

		if r.URL.Query().Get("state") != stateCookie.Value {
			slog.Warn("oauth callback state mismatch")
			c.fail(w, r, client, http.StatusBadRequest, "invalid_state", "invalid oauth state")
			return
		}

		// Google reports a refused consent as a query parameter, not a status.
		if providerErr := r.URL.Query().Get("error"); providerErr != "" && client == domain.AuthClientMobile {
			c.redirectToApp(w, r, url.Values{"error": {"cancelled"}})
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			c.fail(w, r, client, http.StatusBadRequest, "invalid_request", "missing authorization code")
			return
		}

		if client == domain.AuthClientMobile {
			c.completeMobileLogin(w, r, code)
			return
		}

		session, err := c.service.HandleCallback(r.Context(), code)
		if errors.Is(err, usecases.ErrEmailNotAllowed) || errors.Is(err, usecases.ErrEmailNotVerified) {
			http.Redirect(w, r, accessDeniedRedirectURL, http.StatusFound)
			return
		}
		if err != nil {
			slog.Error("handling oauth callback", slog.String("error", err.Error()))
			http.Error(w, "login failed", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{ // #nosec G124 -- Secure is set conditionally on TLS/X-Forwarded-Proto
			Name:     httpserver.SessionCookieName,
			Value:    session.ID,
			Path:     "/",
			Expires:  session.ExpiresAt,
			HttpOnly: true,
			Secure:   isRequestSecure(r),
			SameSite: http.SameSiteLaxMode,
		})

		http.Redirect(w, r, appRedirectURL, http.StatusFound)
	}
}

// completeMobileLogin sends the browser back into the app with a single-use
// code instead of a cookie: the cookie would only live in the throwaway
// browser the login ran in.
func (c *AuthController) completeMobileLogin(w http.ResponseWriter, r *http.Request, code string) {
	handoffCode, err := c.service.HandleMobileCallback(r.Context(), code)
	if errors.Is(err, usecases.ErrEmailNotAllowed) || errors.Is(err, usecases.ErrEmailNotVerified) {
		c.redirectToApp(w, r, url.Values{"error": {"access_denied"}})
		return
	}
	if err != nil {
		slog.Error("handling mobile oauth callback", slog.String("error", err.Error()))
		c.redirectToApp(w, r, url.Values{"error": {"sign_in_failed"}})
		return
	}

	c.redirectToApp(w, r, url.Values{"code": {handoffCode}})
}

// exchangeMobileCode trades the single-use code from the redirect for the
// bearer session token. It is public by definition: the code stands in for
// authentication, and it is worth one exchange.
func (c *AuthController) exchangeMobileCode() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body internal.MobileExchangeRequest
		if err := httpserver.DecodeJSONBody(r, &body); err != nil {
			httpserver.ReplyWithError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		session, err := c.service.ExchangeMobileCode(r.Context(), strings.TrimSpace(body.Code))
		if errors.Is(err, usecases.ErrAuthHandoffNotFound) {
			httpserver.ReplyWithError(w, http.StatusUnauthorized, "invalid or expired code")
			return
		}
		if err != nil {
			slog.Error("exchanging mobile code", slog.String("error", err.Error()))
			httpserver.ReplyWithError(w, http.StatusInternalServerError, "login failed")
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusOK, internal.ToMobileSessionResponse(session))
	}
}

// fail ends a callback that cannot continue: the app is sent back with the
// reason so it is not left on a web page it cannot act on.
func (c *AuthController) fail(w http.ResponseWriter, r *http.Request, client domain.AuthClient, status int, reason, message string) {
	if client == domain.AuthClientMobile {
		c.redirectToApp(w, r, url.Values{"error": {reason}})
		return
	}
	http.Error(w, message, status)
}

// redirectToApp appends the outcome to the configured mobile redirect URI.
func (c *AuthController) redirectToApp(w http.ResponseWriter, r *http.Request, params url.Values) {
	separator := "?"
	if strings.Contains(c.mobileRedirectURI, "?") {
		separator = "&"
	}
	http.Redirect(w, r, c.mobileRedirectURI+separator+params.Encode(), http.StatusFound)
}

func (c *AuthController) logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if token := httpserver.SessionTokenFrom(r); token != "" {
			if err := c.service.Logout(r.Context(), token); err != nil {
				slog.Error("logging out", slog.String("error", err.Error()))
			}
		}

		clearCookie(w, httpserver.SessionCookieName, isRequestSecure(r))
		w.WriteHeader(http.StatusNoContent)
	}
}

func (c *AuthController) currentUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := c.sessionFromRequest(r)
		if !ok {
			httpserver.ReplyWithError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusOK, internal.ToCurrentUserResponse(session))
	}
}

func (c *AuthController) listAllowedUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := c.service.ListAllowedUsers(r.Context())
		if err != nil {
			slog.Error("listing allowed users", slog.String("error", err.Error()))
			http.Error(w, "failed to list allowed users", http.StatusInternalServerError)
			return
		}

		response := make([]internal.AllowedUserResponse, 0, len(users))
		for _, user := range users {
			response = append(response, internal.ToAllowedUserResponse(user))
		}

		httpserver.ReplyJSONResponse(w, http.StatusOK, response)
	}
}

func (c *AuthController) addAllowedUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body internal.AllowedUserCreateRequest
		if err := httpserver.DecodeJSONBody(r, &body); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		user, err := c.service.AddAllowedUser(r.Context(), body.Email, body.IsAdmin)
		if errors.Is(err, usecases.ErrAllowedUserDuplicated) {
			http.Error(w, "email already allowed", http.StatusConflict)
			return
		}
		if err != nil {
			slog.Warn("adding allowed user", slog.String("error", err.Error()))
			http.Error(w, "invalid allowed user", http.StatusBadRequest)
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusCreated, internal.ToAllowedUserResponse(user))
	}
}

func (c *AuthController) updateAllowedUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}

		var body internal.AllowedUserUpdateRequest
		if err := httpserver.DecodeJSONBody(r, &body); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		user, err := c.service.UpdateAllowedUser(r.Context(), domain.ID(id), body.IsAdmin)
		if errors.Is(err, usecases.ErrAllowedUserNotFound) {
			http.Error(w, "allowed user not found", http.StatusNotFound)
			return
		}
		if err != nil {
			slog.Error("updating allowed user", slog.String("error", err.Error()))
			http.Error(w, "failed to update allowed user", http.StatusInternalServerError)
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusOK, internal.ToAllowedUserResponse(user))
	}
}

func (c *AuthController) removeAllowedUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}

		if err := c.service.RemoveAllowedUser(r.Context(), domain.ID(id)); err != nil {
			slog.Error("removing allowed user", slog.String("error", err.Error()))
			http.Error(w, "failed to remove allowed user", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (c *AuthController) sessionFromRequest(r *http.Request) (domain.Session, bool) {
	token := httpserver.SessionTokenFrom(r)
	if token == "" {
		return domain.Session{}, false
	}

	session, err := c.service.GetSession(r.Context(), token)
	if err != nil {
		return domain.Session{}, false
	}

	return session, true
}

func clearCookie(w http.ResponseWriter, name string, secure bool) {
	http.SetCookie(w, &http.Cookie{ // #nosec G124 -- Secure is set conditionally on TLS/X-Forwarded-Proto
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func isRequestSecure(r *http.Request) bool {
	return r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
}

func generateState() (string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("reading random bytes: %w", err)
	}

	return hex.EncodeToString(buffer), nil
}
