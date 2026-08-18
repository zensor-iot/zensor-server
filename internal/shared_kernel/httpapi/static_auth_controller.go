package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"zensor-server/internal/infra/httpserver"
	"zensor-server/internal/shared_kernel/httpapi/internal"
	"zensor-server/internal/shared_kernel/usecases"
)

// NewStaticAuthController builds the controller for static auth mode: a single
// hardcoded admin user authenticated via username/password. Local dev only.
func NewStaticAuthController(service usecases.StaticAuthService) *StaticAuthController {
	return &StaticAuthController{
		service: service,
	}
}

var _ httpserver.Controller = &StaticAuthController{}

type StaticAuthController struct {
	service usecases.StaticAuthService
}

func (c *StaticAuthController) AddRoutes(router *http.ServeMux) {
	router.Handle("GET /auth/mode", c.mode())
	router.Handle("POST /auth/login", c.login())
	router.Handle("POST /auth/logout", c.logout())
	router.Handle("GET /v1/me", c.currentUser())
}

func (c *StaticAuthController) mode() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpserver.ReplyJSONResponse(w, http.StatusOK, internal.AuthModeResponse{Mode: "static"})
	}
}

func (c *StaticAuthController) login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body internal.StaticLoginRequest
		if err := httpserver.DecodeJSONBody(r, &body); err != nil {
			httpserver.ReplyWithError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		session, err := c.service.Login(r.Context(), body.Username, body.Password)
		if errors.Is(err, usecases.ErrInvalidCredentials) {
			httpserver.ReplyWithError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		if err != nil {
			slog.Error("static login failed", slog.String("error", err.Error()))
			httpserver.ReplyWithError(w, http.StatusInternalServerError, "login failed")
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

		httpserver.ReplyJSONResponse(w, http.StatusOK, internal.ToCurrentUserResponse(session))
	}
}

func (c *StaticAuthController) logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, err := r.Cookie(httpserver.SessionCookieName)
		if err == nil && sessionCookie.Value != "" {
			if err := c.service.Logout(r.Context(), sessionCookie.Value); err != nil {
				slog.Error("logging out", slog.String("error", err.Error()))
			}
		}

		clearCookie(w, httpserver.SessionCookieName, isRequestSecure(r))
		w.WriteHeader(http.StatusNoContent)
	}
}

func (c *StaticAuthController) currentUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, err := r.Cookie(httpserver.SessionCookieName)
		if err != nil || sessionCookie.Value == "" {
			httpserver.ReplyWithError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		session, err := c.service.GetSession(r.Context(), sessionCookie.Value)
		if err != nil {
			httpserver.ReplyWithError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusOK, internal.ToCurrentUserResponse(session))
	}
}
