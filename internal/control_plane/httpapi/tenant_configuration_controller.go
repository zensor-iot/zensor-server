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
	createTenantConfigurationErrMessage   = "failed to create tenant configuration"
	getTenantConfigurationErrMessage      = "failed to get tenant configuration"
	updateTenantConfigurationErrMessage   = "failed to update tenant configuration"
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
	router.Handle("POST /v1/tenants/{id}/configuration", c.createTenantConfiguration())
	router.Handle("PUT /v1/tenants/{id}/configuration", c.updateTenantConfiguration())
}

func (c *TenantConfigurationController) getTenantConfiguration() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.PathValue("id")
		if tenantID == "" {
			http.Error(w, "tenant id is required", http.StatusBadRequest)
			return
		}

		config, err := c.service.GetTenantConfiguration(r.Context(), domain.ID(tenantID))
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

func (c *TenantConfigurationController) createTenantConfiguration() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.PathValue("id")
		if tenantID == "" {
			http.Error(w, "tenant id is required", http.StatusBadRequest)
			return
		}

		var body internal.TenantConfigurationCreateRequest
		err := httpserver.DecodeJSONBody(r, &body)
		if err != nil {
			slog.Error("decoding create tenant configuration request", slog.String("error", err.Error()))
			http.Error(w, createTenantConfigurationErrMessage, http.StatusBadRequest)
			return
		}

		config, err := domain.NewTenantConfigurationBuilder().
			WithTenantID(domain.ID(tenantID)).
			WithTimezone(body.Timezone).
			Build()
		if err != nil {
			slog.Error("building tenant configuration", slog.String("error", err.Error()))
			http.Error(w, invalidTimezoneErrMessage, http.StatusBadRequest)
			return
		}

		err = c.service.CreateTenantConfiguration(r.Context(), config)
		if errors.Is(err, usecases.ErrTenantConfigurationAlreadyExists) {
			http.Error(w, createTenantConfigurationErrMessage, http.StatusConflict)
			return
		}
		if err != nil {
			slog.Error("creating tenant configuration", slog.String("error", err.Error()))
			http.Error(w, createTenantConfigurationErrMessage, http.StatusInternalServerError)
			return
		}

		response := internal.ToTenantConfigurationResponse(config)
		httpserver.ReplyJSONResponse(w, http.StatusCreated, response)
	}
}

func (c *TenantConfigurationController) updateTenantConfiguration() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.PathValue("id")
		if tenantID == "" {
			http.Error(w, "tenant id is required", http.StatusBadRequest)
			return
		}

		var body internal.TenantConfigurationUpdateRequest
		err := httpserver.DecodeJSONBody(r, &body)
		if err != nil {
			slog.Error("decoding update tenant configuration request", slog.String("error", err.Error()))
			http.Error(w, updateTenantConfigurationErrMessage, http.StatusBadRequest)
			return
		}

		// Create a configuration object with the update data (version handled internally)
		config := domain.TenantConfiguration{
			TenantID: domain.ID(tenantID),
			Timezone: body.Timezone,
		}

		err = c.service.UpdateTenantConfiguration(r.Context(), config)
		if errors.Is(err, usecases.ErrTenantConfigurationNotFound) {
			http.Error(w, tenantConfigurationNotFoundErrMessage, http.StatusNotFound)
			return
		}
		if errors.Is(err, usecases.ErrInvalidTimezone) {
			http.Error(w, invalidTimezoneErrMessage, http.StatusBadRequest)
			return
		}
		if err != nil {
			slog.Error("updating tenant configuration", slog.String("error", err.Error()))
			http.Error(w, updateTenantConfigurationErrMessage, http.StatusInternalServerError)
			return
		}

		// Get updated configuration to return
		config, err = c.service.GetTenantConfiguration(r.Context(), domain.ID(tenantID))
		if err != nil {
			slog.Error("getting updated tenant configuration", slog.String("error", err.Error()))
			http.Error(w, getTenantConfigurationErrMessage, http.StatusInternalServerError)
			return
		}

		response := internal.ToTenantConfigurationResponse(config)
		httpserver.ReplyJSONResponse(w, http.StatusOK, response)
	}
}
