package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"zensor-server/internal/shared_kernel/httpapi/internal"
	"zensor-server/internal/shared_kernel/usecases"
	"zensor-server/internal/infra/httpserver"
	"zensor-server/internal/shared_kernel/domain"
)

const (
	registerTokenErrMessage   = "failed to register push token"
	unregisterTokenErrMessage = "failed to unregister push token"
)

func NewPushTokenController(service usecases.PushTokenService) *PushTokenController {
	return &PushTokenController{
		service: service,
	}
}

var _ httpserver.Controller = &PushTokenController{}

type PushTokenController struct {
	service usecases.PushTokenService
}

func (c *PushTokenController) AddRoutes(router *http.ServeMux) {
	router.Handle("POST /v1/users/{id}/push-tokens", c.registerToken())
	router.Handle("DELETE /v1/users/{id}/push-tokens", c.unregisterToken())
}

func (c *PushTokenController) registerToken() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("id")
		var body internal.PushTokenRegistrationRequest
		err := httpserver.DecodeJSONBody(r, &body)
		if err != nil {
			http.Error(w, registerTokenErrMessage, http.StatusBadRequest)
			return
		}

		err = c.service.RegisterToken(r.Context(), domain.ID(userID), body.Token, body.Platform)
		if err != nil {
			slog.Error("registering push token", slog.String("error", err.Error()))
			http.Error(w, registerTokenErrMessage, http.StatusInternalServerError)
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusOK, nil)
	}
}

func (c *PushTokenController) unregisterToken() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body internal.PushTokenUnregistrationRequest
		err := httpserver.DecodeJSONBody(r, &body)
		if err != nil {
			http.Error(w, unregisterTokenErrMessage, http.StatusBadRequest)
			return
		}

		err = c.service.UnregisterToken(r.Context(), body.Token)
		if errors.Is(err, usecases.ErrPushTokenNotFound) {
			httpserver.ReplyJSONResponse(w, http.StatusNotFound, map[string]string{"error": "push token not found"})
			return
		}
		if err != nil {
			slog.Error("unregistering push token", slog.String("error", err.Error()))
			http.Error(w, unregisterTokenErrMessage, http.StatusInternalServerError)
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusOK, nil)
	}
}
