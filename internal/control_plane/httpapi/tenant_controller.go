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
	createTenantErrMessage          = "failed to create tenant"
	tenantNotFoundErrMessage        = "tenant not found"
	tenantDuplicatedErrMessage      = "tenant already exists"
	tenantSoftDeletedErrMessage     = "tenant is soft deleted"
	tenantVersionConflictErrMessage = "tenant version conflict"
	updateTenantErrMessage          = "failed to update tenant"
	softDeleteTenantErrMessage      = "failed to soft delete tenant"
	listTenantsErrMessage           = "failed to list tenants"
	getTenantErrMessage             = "failed to get tenant"
)

func NewTenantController(service usecases.TenantService) *TenantController {
	return &TenantController{
		service: service,
	}
}

var _ httpserver.Controller = &TenantController{}

type TenantController struct {
	service usecases.TenantService
}

func (c *TenantController) AddRoutes(router *http.ServeMux) {
	router.Handle("GET /v1/tenants", c.listTenants())
	router.Handle("POST /v1/tenants", c.createTenant())
	router.Handle("GET /v1/tenants/{id}", c.getTenant())
	router.Handle("PUT /v1/tenants/{id}", c.updateTenant())
	router.Handle("DELETE /v1/tenants/{id}", c.softDeleteTenant())
	router.Handle("POST /v1/tenants/{id}/activate", c.activateTenant())
	router.Handle("POST /v1/tenants/{id}/deactivate", c.deactivateTenant())
	router.Handle("POST /v1/tenants/{id}/adopt", c.adoptDevice())
	router.Handle("GET /v1/tenants/{id}/devices", c.listTenantDevices())
}

func (c *TenantController) listTenants() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		includeDeleted := false
		if r.URL.Query().Get("include_deleted") == "true" {
			includeDeleted = true
		}

		params := httpserver.ExtractPaginationParams(r)
		pagination := usecases.Pagination{Limit: params.Limit, Offset: (params.Page - 1) * params.Limit}

		tenants, total, err := c.service.ListTenants(r.Context(), includeDeleted, pagination)
		if err != nil {
			slog.Error("listing tenants", slog.String("error", err.Error()))
			http.Error(w, listTenantsErrMessage, http.StatusInternalServerError)
			return
		}

		responses := make([]internal.TenantResponse, len(tenants))
		for i, tenant := range tenants {
			responses[i] = internal.ToTenantResponse(tenant)
		}

		httpserver.ReplyWithPaginatedData(w, http.StatusOK, responses, total, params)
	}
}

func (c *TenantController) createTenant() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body internal.TenantCreateRequest
		err := httpserver.DecodeJSONBody(r, &body)
		if err != nil {
			slog.Error("decoding create tenant request", slog.String("error", err.Error()))
			http.Error(w, createTenantErrMessage, http.StatusBadRequest)
			return
		}

		tenant, err := domain.NewTenantBuilder().
			WithName(body.Name).
			WithEmail(body.Email).
			WithDescription(body.Description).
			Build()
		if err != nil {
			slog.Error("building tenant", slog.String("error", err.Error()))
			http.Error(w, createTenantErrMessage, http.StatusInternalServerError)
			return
		}

		err = c.service.CreateTenant(r.Context(), tenant)
		if errors.Is(err, usecases.ErrTenantDuplicated) {
			http.Error(w, tenantDuplicatedErrMessage, http.StatusConflict)
			return
		}
		if err != nil {
			slog.Error("creating tenant", slog.String("error", err.Error()))
			http.Error(w, createTenantErrMessage, http.StatusInternalServerError)
			return
		}

		response := internal.ToTenantResponse(tenant)
		httpserver.ReplyJSONResponse(w, http.StatusCreated, response)
	}
}

func (c *TenantController) getTenant() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "tenant id is required", http.StatusBadRequest)
			return
		}

		tenant, err := c.service.GetTenant(r.Context(), domain.ID(id))
		if errors.Is(err, usecases.ErrTenantNotFound) {
			http.Error(w, tenantNotFoundErrMessage, http.StatusNotFound)
			return
		}
		if err != nil {
			slog.Error("getting tenant", slog.String("error", err.Error()))
			http.Error(w, getTenantErrMessage, http.StatusInternalServerError)
			return
		}

		response := internal.ToTenantResponse(tenant)
		httpserver.ReplyJSONResponse(w, http.StatusOK, response)
	}
}

func (c *TenantController) updateTenant() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "tenant id is required", http.StatusBadRequest)
			return
		}

		var body internal.TenantUpdateRequest
		err := httpserver.DecodeJSONBody(r, &body)
		if err != nil {
			slog.Error("decoding update tenant request", slog.String("error", err.Error()))
			http.Error(w, updateTenantErrMessage, http.StatusBadRequest)
			return
		}

		// Create a tenant object with the update data
		tenant := domain.Tenant{
			ID:          domain.ID(id),
			Name:        body.Name,
			Email:       body.Email,
			Description: body.Description,
			Version:     body.Version,
		}

		err = c.service.UpdateTenant(r.Context(), tenant)
		if errors.Is(err, usecases.ErrTenantNotFound) {
			http.Error(w, tenantNotFoundErrMessage, http.StatusNotFound)
			return
		}
		if errors.Is(err, usecases.ErrTenantDuplicated) {
			http.Error(w, tenantDuplicatedErrMessage, http.StatusConflict)
			return
		}
		if errors.Is(err, usecases.ErrTenantSoftDeleted) {
			http.Error(w, tenantSoftDeletedErrMessage, http.StatusConflict)
			return
		}
		if errors.Is(err, usecases.ErrTenantVersionConflict) {
			http.Error(w, tenantVersionConflictErrMessage, http.StatusConflict)
			return
		}
		if err != nil {
			slog.Error("updating tenant", slog.String("error", err.Error()))
			http.Error(w, updateTenantErrMessage, http.StatusInternalServerError)
			return
		}

		// Get updated tenant to return
		tenant, err = c.service.GetTenant(r.Context(), domain.ID(id))
		if err != nil {
			slog.Error("getting updated tenant", slog.String("error", err.Error()))
			http.Error(w, getTenantErrMessage, http.StatusInternalServerError)
			return
		}

		response := internal.ToTenantResponse(tenant)
		httpserver.ReplyJSONResponse(w, http.StatusOK, response)
	}
}

func (c *TenantController) softDeleteTenant() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "tenant id is required", http.StatusBadRequest)
			return
		}

		err := c.service.SoftDeleteTenant(r.Context(), domain.ID(id))
		if errors.Is(err, usecases.ErrTenantNotFound) {
			http.Error(w, tenantNotFoundErrMessage, http.StatusNotFound)
			return
		}
		if errors.Is(err, usecases.ErrTenantSoftDeleted) {
			http.Error(w, tenantSoftDeletedErrMessage, http.StatusConflict)
			return
		}
		if err != nil {
			slog.Error("soft deleting tenant", slog.String("error", err.Error()))
			http.Error(w, softDeleteTenantErrMessage, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (c *TenantController) activateTenant() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "tenant id is required", http.StatusBadRequest)
			return
		}

		err := c.service.ActivateTenant(r.Context(), domain.ID(id))
		if errors.Is(err, usecases.ErrTenantNotFound) {
			http.Error(w, tenantNotFoundErrMessage, http.StatusNotFound)
			return
		}
		if errors.Is(err, usecases.ErrTenantSoftDeleted) {
			http.Error(w, tenantSoftDeletedErrMessage, http.StatusConflict)
			return
		}
		if err != nil {
			slog.Error("activating tenant", slog.String("error", err.Error()))
			http.Error(w, "failed to activate tenant", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (c *TenantController) deactivateTenant() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "tenant id is required", http.StatusBadRequest)
			return
		}

		err := c.service.DeactivateTenant(r.Context(), domain.ID(id))
		if errors.Is(err, usecases.ErrTenantNotFound) {
			http.Error(w, tenantNotFoundErrMessage, http.StatusNotFound)
			return
		}
		if errors.Is(err, usecases.ErrTenantSoftDeleted) {
			http.Error(w, tenantSoftDeletedErrMessage, http.StatusConflict)
			return
		}
		if err != nil {
			slog.Error("deactivating tenant", slog.String("error", err.Error()))
			http.Error(w, "failed to deactivate tenant", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (c *TenantController) adoptDevice() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.PathValue("id")
		if tenantID == "" {
			http.Error(w, "tenant id is required", http.StatusBadRequest)
			return
		}

		var body internal.TenantAdoptDeviceRequest
		err := httpserver.DecodeJSONBody(r, &body)
		if err != nil {
			slog.Error("decoding adopt device request", slog.String("error", err.Error()))
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if body.DeviceID == "" {
			http.Error(w, "device_id is required", http.StatusBadRequest)
			return
		}

		err = c.service.AdoptDevice(r.Context(), domain.ID(tenantID), domain.ID(body.DeviceID))
		if errors.Is(err, usecases.ErrTenantNotFound) {
			http.Error(w, tenantNotFoundErrMessage, http.StatusNotFound)
			return
		}
		if errors.Is(err, usecases.ErrTenantSoftDeleted) {
			http.Error(w, tenantSoftDeletedErrMessage, http.StatusConflict)
			return
		}
		if errors.Is(err, usecases.ErrDeviceNotFound) {
			http.Error(w, "device not found", http.StatusNotFound)
			return
		}
		if err != nil {
			slog.Error("adopting device to tenant", slog.String("error", err.Error()))
			http.Error(w, "failed to adopt device", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (c *TenantController) listTenantDevices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.PathValue("id")
		if tenantID == "" {
			http.Error(w, "tenant id is required", http.StatusBadRequest)
			return
		}

		params := httpserver.ExtractPaginationParams(r)
		pagination := usecases.Pagination{Limit: params.Limit, Offset: (params.Page - 1) * params.Limit}

		devices, total, err := c.service.ListTenantDevices(r.Context(), domain.ID(tenantID), pagination)
		if errors.Is(err, usecases.ErrTenantNotFound) {
			http.Error(w, tenantNotFoundErrMessage, http.StatusNotFound)
			return
		}
		if errors.Is(err, usecases.ErrTenantSoftDeleted) {
			http.Error(w, tenantSoftDeletedErrMessage, http.StatusConflict)
			return
		}
		if err != nil {
			slog.Error("listing tenant devices", slog.String("error", err.Error()))
			http.Error(w, "failed to list tenant devices", http.StatusInternalServerError)
			return
		}

		responses := make([]internal.DeviceResponse, len(devices))
		for i, device := range devices {
			responses[i] = internal.ToDeviceResponse(device)
		}

		httpserver.ReplyWithPaginatedData(w, http.StatusOK, responses, total, params)
	}
}
