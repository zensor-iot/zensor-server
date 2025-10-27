package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"zensor-server/internal/control_plane/httpapi/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/httpserver"
	"zensor-server/internal/shared_kernel/domain"
)

const (
	associateTenantsErrMessage = "failed to associate tenants with user"
	getUserErrMessage          = "failed to get user"
	invalidTenantErrMessage    = "one or more tenants are invalid"
)

func NewUserController(service usecases.UserService) *UserController {
	return &UserController{
		service: service,
	}
}

var _ httpserver.Controller = &UserController{}

type UserController struct {
	service usecases.UserService
}

func (c *UserController) AddRoutes(router *http.ServeMux) {
	router.Handle("PUT /v1/users/{id}", c.associateTenants())
	router.Handle("GET /v1/users/{id}", c.getUser())
}

func (c *UserController) associateTenants() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var body internal.UserUpdateRequest
		err := httpserver.DecodeJSONBody(r, &body)
		if err != nil {
			http.Error(w, associateTenantsErrMessage, http.StatusBadRequest)
			return
		}

		tenantIDs := make([]domain.ID, len(body.Tenants))
		for i, tenantIDStr := range body.Tenants {
			tenantIDs[i] = domain.ID(tenantIDStr)
		}

		err = c.service.AssociateTenants(r.Context(), domain.ID(id), tenantIDs)
		if errors.Is(err, usecases.ErrTenantNotFound) {
			http.Error(w, invalidTenantErrMessage, http.StatusBadRequest)
			return
		}
		if err != nil {
			slog.Error("associating tenants", slog.String("error", err.Error()))
			http.Error(w, associateTenantsErrMessage, http.StatusInternalServerError)
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusOK, nil)
	}
}

func (c *UserController) getUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		user, err := c.service.GetUser(r.Context(), domain.ID(id))
		if errors.Is(err, usecases.ErrUserNotFound) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		if err != nil {
			slog.Error("getting user", slog.String("error", err.Error()))
			http.Error(w, getUserErrMessage, http.StatusInternalServerError)
			return
		}

		response := internal.ToUserResponse(user)
		httpserver.ReplyJSONResponse(w, http.StatusOK, response)
	}
}
