package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/control_plane/httpapi/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/httpserver"
	"zensor-server/internal/infra/utils"
)

const (
	createScheduledTaskErrMessage = "failed to create scheduled task"
	updateScheduledTaskErrMessage = "failed to update scheduled task"
	getScheduledTaskErrMessage    = "failed to get scheduled task"
	listScheduledTaskErrMessage   = "failed to list scheduled tasks"
)

func NewScheduledTaskController(
	service usecases.ScheduledTaskService,
	deviceService usecases.DeviceService,
	tenantService usecases.TenantService,
) *ScheduledTaskController {
	return &ScheduledTaskController{
		service:       service,
		deviceService: deviceService,
		tenantService: tenantService,
	}
}

var _ httpserver.Controller = &ScheduledTaskController{}

type ScheduledTaskController struct {
	service       usecases.ScheduledTaskService
	deviceService usecases.DeviceService
	tenantService usecases.TenantService
}

func (c *ScheduledTaskController) AddRoutes(router *http.ServeMux) {
	router.Handle("POST /v1/tenants/{tenant_id}/devices/{device_id}/scheduled-tasks", c.create())
	router.Handle("GET /v1/tenants/{tenant_id}/devices/{device_id}/scheduled-tasks", c.list())
	router.Handle("GET /v1/tenants/{tenant_id}/devices/{device_id}/scheduled-tasks/{id}", c.get())
	router.Handle("PUT /v1/tenants/{tenant_id}/devices/{device_id}/scheduled-tasks/{id}", c.update())
}

func (c *ScheduledTaskController) create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.PathValue("tenant_id")
		deviceID := r.PathValue("device_id")

		// Verify tenant exists
		tenant, err := c.tenantService.GetTenant(r.Context(), domain.ID(tenantID))
		if err != nil {
			slog.Error("get tenant failed", slog.String("error", err.Error()))
			http.Error(w, createScheduledTaskErrMessage, http.StatusInternalServerError)
			return
		}

		var body internal.ScheduledTaskCreateRequest
		err = httpserver.DecodeJSONBody(r, &body)
		if err != nil {
			slog.Error("decoding json body", slog.String("error", err.Error()))
			http.Error(w, createScheduledTaskErrMessage, http.StatusBadRequest)
			return
		}

		// Get device and verify it belongs to the tenant
		device, err := c.deviceService.GetDevice(r.Context(), domain.ID(deviceID))
		if err != nil {
			slog.Error("get device failed", slog.String("error", err.Error()))
			http.Error(w, createScheduledTaskErrMessage, http.StatusInternalServerError)
			return
		}

		// Create command templates from the request
		commandTemplates := make([]domain.CommandTemplate, len(body.Commands))
		for i, item := range body.Commands {
			templateBuilder := domain.NewCommandTemplateBuilder()
			template, err := templateBuilder.
				WithDevice(device).
				WithPayload(domain.CommandPayload{
					Index: domain.Index(item.Index),
					Value: domain.CommandValue(item.Value),
				}).
				WithPriority(domain.CommandPriority(item.Priority)).
				WithWaitFor(time.Duration(item.WaitFor)).
				Build()
			if err != nil {
				slog.Error("build command template", slog.String("error", err.Error()))
				http.Error(w, createScheduledTaskErrMessage, http.StatusInternalServerError)
				return
			}

			commandTemplates[i] = template
		}

		// Create the scheduled task with command templates
		scheduledTask, err := domain.NewScheduledTaskBuilder().
			WithTenant(tenant).
			WithDevice(device).
			WithCommandTemplates(commandTemplates).
			WithSchedule(body.Schedule).
			WithIsActive(body.IsActive).
			Build()
		if err != nil {
			slog.Error("build scheduled task", slog.String("error", err.Error()))
			http.Error(w, createScheduledTaskErrMessage, http.StatusInternalServerError)
			return
		}

		err = c.service.Create(r.Context(), scheduledTask)
		if err != nil {
			slog.Error("create scheduled task failed", slog.String("error", err.Error()))
			http.Error(w, createScheduledTaskErrMessage, http.StatusInternalServerError)
			return
		}

		// Convert command templates to response format
		responseCommands := make([]internal.CommandSendPayloadRequest, len(scheduledTask.CommandTemplates))
		for i, template := range scheduledTask.CommandTemplates {
			responseCommands[i] = internal.CommandSendPayloadRequest{
				Index:    uint8(template.Payload.Index),
				Value:    uint8(template.Payload.Value),
				Priority: string(template.Priority),
				WaitFor:  utils.Duration(template.WaitFor),
			}
		}

		response := internal.ScheduledTaskResponse{
			ID:       scheduledTask.ID.String(),
			DeviceID: scheduledTask.Device.ID.String(),
			Commands: responseCommands,
			Schedule: scheduledTask.Schedule,
			IsActive: scheduledTask.IsActive,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

func (c *ScheduledTaskController) list() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.PathValue("tenant_id")
		deviceID := r.PathValue("device_id")

		scheduledTasks, err := c.service.FindAllByTenantAndDevice(r.Context(), domain.ID(tenantID), domain.ID(deviceID))
		if err != nil {
			slog.Error("list scheduled tasks failed", slog.String("error", err.Error()))
			http.Error(w, listScheduledTaskErrMessage, http.StatusInternalServerError)
			return
		}

		responses := make([]internal.ScheduledTaskResponse, len(scheduledTasks))
		for i, scheduledTask := range scheduledTasks {
			// Convert domain command templates to API commands
			apiCommands := make([]internal.CommandSendPayloadRequest, len(scheduledTask.CommandTemplates))
			for j, template := range scheduledTask.CommandTemplates {
				apiCommands[j] = internal.CommandSendPayloadRequest{
					Index:    uint8(template.Payload.Index),
					Value:    uint8(template.Payload.Value),
					Priority: string(template.Priority),
					WaitFor:  utils.Duration(template.WaitFor),
				}
			}

			responses[i] = internal.ScheduledTaskResponse{
				ID:       scheduledTask.ID.String(),
				DeviceID: scheduledTask.Device.ID.String(),
				Commands: apiCommands,
				Schedule: scheduledTask.Schedule,
				IsActive: scheduledTask.IsActive,
			}
		}

		response := internal.ScheduledTaskListResponse{
			ScheduledTasks: responses,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func (c *ScheduledTaskController) get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.PathValue("tenant_id")
		deviceID := r.PathValue("device_id")
		id := r.PathValue("id")

		scheduledTask, err := c.service.GetByID(r.Context(), domain.ID(id))
		if err != nil {
			slog.Error("get scheduled task failed", slog.String("error", err.Error()))
			http.Error(w, getScheduledTaskErrMessage, http.StatusInternalServerError)
			return
		}

		// Verify the scheduled task belongs to the tenant and device
		if scheduledTask.Tenant.ID != domain.ID(tenantID) || scheduledTask.Device.ID != domain.ID(deviceID) {
			http.Error(w, getScheduledTaskErrMessage, http.StatusNotFound)
			return
		}

		// Convert command templates to response format
		responseCommands := make([]internal.CommandSendPayloadRequest, len(scheduledTask.CommandTemplates))
		for i, template := range scheduledTask.CommandTemplates {
			responseCommands[i] = internal.CommandSendPayloadRequest{
				Index:    uint8(template.Payload.Index),
				Value:    uint8(template.Payload.Value),
				Priority: string(template.Priority),
				WaitFor:  utils.Duration(template.WaitFor),
			}
		}

		response := internal.ScheduledTaskResponse{
			ID:       scheduledTask.ID.String(),
			DeviceID: scheduledTask.Device.ID.String(),
			Commands: responseCommands,
			Schedule: scheduledTask.Schedule,
			IsActive: scheduledTask.IsActive,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func (c *ScheduledTaskController) update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.PathValue("tenant_id")
		deviceID := r.PathValue("device_id")
		id := r.PathValue("id")

		// Get existing scheduled task
		scheduledTask, err := c.service.GetByID(r.Context(), domain.ID(id))
		if err != nil {
			slog.Error("get scheduled task failed", slog.String("error", err.Error()))
			http.Error(w, updateScheduledTaskErrMessage, http.StatusInternalServerError)
			return
		}

		// Verify the scheduled task belongs to the tenant and device
		if scheduledTask.Tenant.ID != domain.ID(tenantID) || scheduledTask.Device.ID != domain.ID(deviceID) {
			http.Error(w, updateScheduledTaskErrMessage, http.StatusNotFound)
			return
		}

		var body internal.ScheduledTaskUpdateRequest
		err = httpserver.DecodeJSONBody(r, &body)
		if err != nil {
			slog.Error("decoding json body", slog.String("error", err.Error()))
			http.Error(w, updateScheduledTaskErrMessage, http.StatusBadRequest)
			return
		}

		// Update fields if provided
		if body.Schedule != nil {
			scheduledTask.Schedule = *body.Schedule
		}
		if body.IsActive != nil {
			scheduledTask.IsActive = *body.IsActive
		}
		if body.Commands != nil {
			// Convert API commands to domain command templates
			device, err := c.deviceService.GetDevice(r.Context(), scheduledTask.Device.ID)
			if err != nil {
				slog.Error("get device failed", slog.String("error", err.Error()))
				http.Error(w, updateScheduledTaskErrMessage, http.StatusInternalServerError)
				return
			}

			commandTemplates := make([]domain.CommandTemplate, len(*body.Commands))
			for i, item := range *body.Commands {
				templateBuilder := domain.NewCommandTemplateBuilder()
				template, err := templateBuilder.
					WithDevice(device).
					WithPayload(domain.CommandPayload{
						Index: domain.Index(item.Index),
						Value: domain.CommandValue(item.Value),
					}).
					WithPriority(domain.CommandPriority(item.Priority)).
					WithWaitFor(time.Duration(item.WaitFor)).
					Build()
				if err != nil {
					slog.Error("build command template", slog.String("error", err.Error()))
					http.Error(w, updateScheduledTaskErrMessage, http.StatusInternalServerError)
					return
				}

				commandTemplates[i] = template
			}

			scheduledTask.CommandTemplates = commandTemplates
		}

		err = c.service.Update(r.Context(), scheduledTask)
		if err != nil {
			slog.Error("update scheduled task failed", slog.String("error", err.Error()))
			http.Error(w, updateScheduledTaskErrMessage, http.StatusInternalServerError)
			return
		}

		// Convert command templates to response format
		responseCommands := make([]internal.CommandSendPayloadRequest, len(scheduledTask.CommandTemplates))
		for i, template := range scheduledTask.CommandTemplates {
			responseCommands[i] = internal.CommandSendPayloadRequest{
				Index:    uint8(template.Payload.Index),
				Value:    uint8(template.Payload.Value),
				Priority: string(template.Priority),
				WaitFor:  utils.Duration(template.WaitFor),
			}
		}

		response := internal.ScheduledTaskResponse{
			ID:       scheduledTask.ID.String(),
			DeviceID: scheduledTask.Device.ID.String(),
			Commands: responseCommands,
			Schedule: scheduledTask.Schedule,
			IsActive: scheduledTask.IsActive,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
