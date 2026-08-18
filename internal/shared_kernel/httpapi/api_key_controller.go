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

func NewAPIKeyController(service usecases.APIKeyService) *APIKeyController {
	return &APIKeyController{
		service: service,
	}
}

var _ httpserver.Controller = &APIKeyController{}

type APIKeyController struct {
	service usecases.APIKeyService
}

func (c *APIKeyController) AddRoutes(router *http.ServeMux) {
	router.Handle("POST /v1/admin/api-keys", c.create())
	router.Handle("GET /v1/admin/api-keys", c.list())
	router.Handle("DELETE /v1/admin/api-keys/{id}", c.revoke())
}

func (c *APIKeyController) create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body internal.APIKeyCreateRequest
		if err := httpserver.DecodeJSONBody(r, &body); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		createdBy := domain.ID(r.Header.Get("X-User-ID"))
		key, plaintext, err := c.service.Create(r.Context(), body.Name, createdBy)
		if errors.Is(err, usecases.ErrAPIKeyDuplicated) {
			http.Error(w, "api key name already exists", http.StatusConflict)
			return
		}
		if errors.Is(err, domain.ErrAPIKeyNameRequired) {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}
		if err != nil {
			slog.Error("creating api key", slog.String("error", err.Error()))
			http.Error(w, "failed to create api key", http.StatusInternalServerError)
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusCreated, internal.ToAPIKeyCreatedResponse(key, plaintext))
	}
}

func (c *APIKeyController) list() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		keys, err := c.service.List(r.Context())
		if err != nil {
			slog.Error("listing api keys", slog.String("error", err.Error()))
			http.Error(w, "failed to list api keys", http.StatusInternalServerError)
			return
		}

		response := make([]internal.APIKeyResponse, 0, len(keys))
		for _, key := range keys {
			response = append(response, internal.ToAPIKeyResponse(key))
		}

		httpserver.ReplyJSONResponse(w, http.StatusOK, response)
	}
}

func (c *APIKeyController) revoke() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}

		err := c.service.Revoke(r.Context(), domain.ID(id))
		if errors.Is(err, usecases.ErrAPIKeyNotFound) {
			http.Error(w, "api key not found", http.StatusNotFound)
			return
		}
		if err != nil {
			slog.Error("revoking api key", slog.String("error", err.Error()))
			http.Error(w, "failed to revoke api key", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
