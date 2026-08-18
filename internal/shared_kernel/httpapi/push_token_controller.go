package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"zensor-server/internal/infra/httpserver"
	"zensor-server/internal/shared_kernel/domain"
	"zensor-server/internal/shared_kernel/httpapi/internal"
	"zensor-server/internal/shared_kernel/usecases"
)

const (
	registerTokenErrMessage    = "failed to register push token"
	unregisterTokenErrMessage  = "failed to unregister push token"
	broadcastPushErrMessage    = "failed to broadcast push notification"
	invalidBroadcastErrMessage = "invalid push broadcast request"
)

// NewPushTokenController registers push-token HTTP routes including user-targeted push broadcast.
func NewPushTokenController(service usecases.PushTokenService, sender usecases.UserPushMessageSender) *PushTokenController {
	return &PushTokenController{
		service: service,
		sender:  sender,
	}
}

var _ httpserver.Controller = &PushTokenController{}

type PushTokenController struct {
	service usecases.PushTokenService
	sender  usecases.UserPushMessageSender
}

func (c *PushTokenController) AddRoutes(router *http.ServeMux) {
	router.Handle("POST /v1/users/{id}/push-tokens", c.registerToken())
	router.Handle("DELETE /v1/users/{id}/push-tokens", c.unregisterToken())
	router.Handle("POST /v1/users/{id}/push-message", c.broadcastPush())
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

func (c *PushTokenController) broadcastPush() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("id")
		var body internal.UserPushBroadcastRequest
		err := httpserver.DecodeJSONBody(r, &body)
		if err != nil {
			httpserver.ReplyJSONResponse(w, http.StatusBadRequest, map[string]string{"error": invalidBroadcastErrMessage})
			return
		}

		content := usecases.UserPushBroadcastContent{
			Title:    body.Title,
			Body:     body.Body,
			DeepLink: body.DeepLink,
		}
		err = c.sender.SendBroadcastToUser(r.Context(), domain.ID(userID), content)
		if errors.Is(err, usecases.ErrUserPushBroadcastBodyRequired) {
			httpserver.ReplyJSONResponse(w, http.StatusBadRequest, map[string]string{"error": invalidBroadcastErrMessage})
			return
		}
		if errors.Is(err, usecases.ErrPushTokenNotFound) {
			httpserver.ReplyJSONResponse(w, http.StatusNotFound, map[string]string{"error": "push token not found"})
			return
		}
		if err != nil {
			slog.Error("broadcasting push notification", slog.String("error", err.Error()))
			httpserver.ReplyJSONResponse(w, http.StatusInternalServerError, map[string]string{"error": broadcastPushErrMessage})
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusOK, nil)
	}
}
