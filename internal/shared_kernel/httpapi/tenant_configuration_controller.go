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
	upsertTenantConfigurationErrMessage   = "failed to upsert tenant configuration"
	getTenantConfigurationErrMessage      = "failed to get tenant configuration"
	tenantConfigurationNotFoundErrMessage = "tenant configuration not found"
	invalidTimezoneErrMessage             = "invalid timezone"
)

func NewTenantConfigurationController(service usecases.TenantConfigurationService) *TenantConfigurationController {
	return &TenantConfigurationController{
		service: service,
	}
}

var _ httpserver.Controller = &TenantConfigurationController{}

type TenantConfigurationController struct {
	service usecases.TenantConfigurationService
}

func (c *TenantConfigurationController) AddRoutes(router *http.ServeMux) {
	router.Handle("GET /v1/tenants/{id}/configuration", c.getTenantConfiguration())
	router.Handle("PUT /v1/tenants/{id}/configuration", c.upsertTenantConfiguration())
}

func (c *TenantConfigurationController) getTenantConfiguration() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.PathValue("id")
		if tenantID == "" {
			http.Error(w, "tenant id is required", http.StatusBadRequest)
			return
		}

		tenant := domain.Tenant{ID: domain.ID(tenantID)}
		config, err := c.service.GetTenantConfiguration(r.Context(), tenant)
		if errors.Is(err, usecases.ErrTenantConfigurationNotFound) {
			http.Error(w, tenantConfigurationNotFoundErrMessage, http.StatusNotFound)
			return
		}
		if err != nil {
			slog.Error("getting tenant configuration", slog.String("error", err.Error()))
			http.Error(w, getTenantConfigurationErrMessage, http.StatusInternalServerError)
			return
		}

		response := internal.ToTenantConfigurationResponse(config)
		httpserver.ReplyJSONResponse(w, http.StatusOK, response)
	}
}

func (c *TenantConfigurationController) upsertTenantConfiguration() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.PathValue("id")
		if tenantID == "" {
			http.Error(w, "tenant id is required", http.StatusBadRequest)
			return
		}

		userEmail := r.Header.Get("X-User-Email")
		if userEmail == "" {
			slog.Error("missing user email in auth header")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var body internal.TenantConfigurationUpdateRequest
		err := httpserver.DecodeJSONBody(r, &body)
		if err != nil {
			slog.Error("decoding upsert tenant configuration request", slog.String("error", err.Error()))
			http.Error(w, upsertTenantConfigurationErrMessage, http.StatusBadRequest)
			return
		}

		builder := domain.NewTenantConfigurationBuilder().
			WithTenantID(domain.ID(tenantID)).
			WithTimezone(body.Timezone)

		if body.NotificationEmail != nil && *body.NotificationEmail != "" {
			builder = builder.WithNotificationEmail(*body.NotificationEmail)
		}

		config, err := builder.Build()
		if err != nil {
			slog.Error("building tenant configuration", slog.String("error", err.Error()))
			http.Error(w, invalidTimezoneErrMessage, http.StatusBadRequest)
			return
		}

		resultConfig, err := c.service.UpsertTenantConfiguration(r.Context(), userEmail, config)
		if errors.Is(err, usecases.ErrUserNotFound) {
			slog.Warn("user not found")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if errors.Is(err, usecases.ErrForbiddenTenantConfigurationAccess) {
			slog.Warn("forbidden access to tenant configuration")
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if errors.Is(err, usecases.ErrInvalidTimezone) {
			http.Error(w, invalidTimezoneErrMessage, http.StatusBadRequest)
			return
		}
		if err != nil {
			slog.Error("upserting tenant configuration", slog.String("error", err.Error()))
			http.Error(w, upsertTenantConfigurationErrMessage, http.StatusInternalServerError)
			return
		}

		response := internal.ToTenantConfigurationResponse(resultConfig)
		httpserver.ReplyJSONResponse(w, http.StatusOK, response)
	}
}
