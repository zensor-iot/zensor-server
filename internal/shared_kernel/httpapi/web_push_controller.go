package httpapi

import (
	"net/http"

	"zensor-server/internal/infra/httpserver"
)

const vapidKeyNotConfiguredErrMessage = "web push is not configured"

// NewWebPushController exposes the VAPID public key so browsers can subscribe to web push.
func NewWebPushController(vapidPublicKey string) *WebPushController {
	return &WebPushController{
		vapidPublicKey: vapidPublicKey,
	}
}

var _ httpserver.Controller = &WebPushController{}

type WebPushController struct {
	vapidPublicKey string
}

func (c *WebPushController) AddRoutes(router *http.ServeMux) {
	router.Handle("GET /v1/push/vapid-public-key", c.getVAPIDPublicKey())
}

func (c *WebPushController) getVAPIDPublicKey() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if c.vapidPublicKey == "" {
			http.Error(w, vapidKeyNotConfiguredErrMessage, http.StatusNotFound)
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusOK, map[string]string{"public_key": c.vapidPublicKey})
	}
}
