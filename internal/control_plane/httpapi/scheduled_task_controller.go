package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"
	"zensor-server/internal/control_plane/httpapi/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/httpserver"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/domain"
)

const (
	createScheduledTaskErrMessage = "failed to create scheduled task"
	updateScheduledTaskErrMessage = "failed to update scheduled task"
	getScheduledTaskErrMessage    = "failed to get scheduled task"
	listScheduledTaskErrMessage   = "failed to list scheduled tasks"
	deleteScheduledTaskErrMessage = "failed to delete scheduled task"
)

func NewScheduledTaskController(
	service usecases.ScheduledTaskService,
	deviceService usecases.DeviceService,
	tenantService usecases.TenantService,
	taskService usecases.TaskService,
) *ScheduledTaskController {
	return &ScheduledTaskController{
		service:       service,
		deviceService: deviceService,
		tenantService: tenantService,
		taskService:   taskService,
	}
}

var _ httpserver.Controller = &ScheduledTaskController{}

type ScheduledTaskController struct {
	service       usecases.ScheduledTaskService
	deviceService usecases.DeviceService
	tenantService usecases.TenantService
	taskService   usecases.TaskService
}

func (c *ScheduledTaskController) AddRoutes(router *http.ServeMux) {
	router.Handle("POST /v1/tenants/{tenant_id}/devices/{device_id}/scheduled-tasks", c.create())
	router.Handle("GET /v1/tenants/{tenant_id}/devices/{device_id}/scheduled-tasks", c.list())
	router.Handle("GET /v1/tenants/{tenant_id}/devices/{device_id}/scheduled-tasks/{id}", c.get())
	router.Handle("PUT /v1/tenants/{tenant_id}/devices/{device_id}/scheduled-tasks/{id}", c.update())
	router.Handle("DELETE /v1/tenants/{tenant_id}/devices/{device_id}/scheduled-tasks/{id}", c.delete())
	router.Handle("GET /v1/tenants/{tenant_id}/devices/{device_id}/scheduled-tasks/{id}/tasks", c.getTasksByScheduledTask())
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
		builder := domain.NewScheduledTaskBuilder().
			WithTenant(tenant).
			WithDevice(device).
			WithCommandTemplates(commandTemplates).
			WithIsActive(body.IsActive)

		if body.Scheduling != nil {
			schedulingConfig := body.Scheduling.ToSchedulingConfiguration()

			if schedulingConfig.Type == domain.SchedulingTypeCron && body.Scheduling.Schedule != nil {
				builder = builder.WithSchedule(*body.Scheduling.Schedule)
			}

			builder = builder.WithScheduling(schedulingConfig)
		} else if body.Schedule != "" {
			builder = builder.WithSchedule(body.Schedule)
		}

		scheduledTask, err := builder.Build()
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

		responseCommands := make([]internal.CommandSendPayloadRequest, len(scheduledTask.CommandTemplates))
		for i, template := range scheduledTask.CommandTemplates {
			responseCommands[i] = internal.CommandSendPayloadRequest{
				Index:    uint8(template.Payload.Index),
				Value:    uint8(template.Payload.Value),
				Priority: string(template.Priority),
				WaitFor:  utils.Duration(template.WaitFor),
			}
		}

		var nextExecution *time.Time
		if scheduledTask.Scheduling.Type == domain.SchedulingTypeInterval {
			nextExec, err := scheduledTask.CalculateNextExecution("UTC") // TODO: Get tenant timezone
			if err == nil {
				nextExecution = &nextExec
			}
		}

		response := internal.ScheduledTaskResponse{
			ID:         scheduledTask.ID.String(),
			DeviceID:   scheduledTask.Device.ID.String(),
			Commands:   responseCommands,
			Schedule:   scheduledTask.Schedule,
			Scheduling: internal.FromSchedulingConfiguration(scheduledTask.Scheduling, nextExecution),
			IsActive:   scheduledTask.IsActive,
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

		params := httpserver.ExtractPaginationParams(r)
		pagination := usecases.Pagination{Limit: params.Limit, Offset: (params.Page - 1) * params.Limit}

		scheduledTasks, total, err := c.service.FindAllByTenantAndDevice(r.Context(), domain.ID(tenantID), domain.ID(deviceID), pagination)
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

		httpserver.ReplyWithPaginatedData(w, http.StatusOK, responses, total, params)
	}
}

func (c *ScheduledTaskController) get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.PathValue("tenant_id")
		deviceID := r.PathValue("device_id")
		id := r.PathValue("id")

		scheduledTask, err := c.service.GetByID(r.Context(), domain.ID(id))
		if err != nil {
			if errors.Is(err, usecases.ErrScheduledTaskNotFound) {
				http.Error(w, getScheduledTaskErrMessage, http.StatusNotFound)
				return
			}
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

func (c *ScheduledTaskController) getTasksByScheduledTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.PathValue("tenant_id")
		deviceID := r.PathValue("device_id")
		scheduledTaskID := r.PathValue("id")

		// Verify tenant exists
		_, err := c.tenantService.GetTenant(r.Context(), domain.ID(tenantID))
		if err != nil {
			slog.Error("get tenant failed", slog.String("error", err.Error()))
			http.Error(w, "failed to get tasks", http.StatusInternalServerError)
			return
		}

		// Verify device exists and belongs to tenant
		_, err = c.deviceService.GetDevice(r.Context(), domain.ID(deviceID))
		if err != nil {
			slog.Error("get device failed", slog.String("error", err.Error()))
			http.Error(w, "failed to get tasks", http.StatusInternalServerError)
			return
		}

		// Verify scheduled task exists and belongs to tenant and device
		scheduledTask, err := c.service.GetByID(r.Context(), domain.ID(scheduledTaskID))
		if err != nil {
			slog.Error("get scheduled task failed", slog.String("error", err.Error()))
			http.Error(w, "failed to get tasks", http.StatusInternalServerError)
			return
		}

		if scheduledTask.Tenant.ID != domain.ID(tenantID) || scheduledTask.Device.ID != domain.ID(deviceID) {
			http.Error(w, "failed to get tasks", http.StatusNotFound)
			return
		}

		// Extract pagination parameters
		params := httpserver.ExtractPaginationParams(r)
		pagination := usecases.Pagination{Limit: params.Limit, Offset: (params.Page - 1) * params.Limit}

		// Get tasks by scheduled task
		tasks, total, err := c.taskService.FindAllByScheduledTask(r.Context(), domain.ID(scheduledTaskID), pagination)
		if err != nil {
			slog.Error("get tasks by scheduled task failed", slog.String("error", err.Error()))
			http.Error(w, "failed to get tasks", http.StatusInternalServerError)
			return
		}

		// Convert domain tasks to API responses
		responses := make([]internal.TaskResponse, len(tasks))
		for i, task := range tasks {
			// Convert domain commands to API command responses
			commandResponses := make([]internal.TaskCommandResponse, len(task.Commands))
			for j, cmd := range task.Commands {
				var sentAt *string
				if !cmd.SentAt.IsZero() {
					sentAtStr := cmd.SentAt.Time.Format("2006-01-02T15:04:05Z07:00")
					sentAt = &sentAtStr
				}

				commandResponses[j] = internal.TaskCommandResponse{
					ID:            cmd.ID.String(),
					Index:         uint8(cmd.Payload.Index),
					Value:         uint8(cmd.Payload.Value),
					Port:          uint8(cmd.Port),
					Priority:      string(cmd.Priority),
					DispatchAfter: cmd.DispatchAfter.Time.Format("2006-01-02T15:04:05Z07:00"),
					Ready:         cmd.Ready,
					Sent:          cmd.Sent,
					SentAt:        sentAt,
				}
			}

			responses[i] = internal.TaskResponse{
				ID:        task.ID.String(),
				Commands:  commandResponses,
				CreatedAt: task.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			}
		}

		// Return paginated response
		httpserver.ReplyWithPaginatedData(w, http.StatusOK, responses, total, params)
	}
}

func (c *ScheduledTaskController) delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.PathValue("tenant_id")
		deviceID := r.PathValue("device_id")
		id := r.PathValue("id")

		// Get existing scheduled task to verify ownership
		scheduledTask, err := c.service.GetByID(r.Context(), domain.ID(id))
		if err != nil {
			if errors.Is(err, usecases.ErrScheduledTaskNotFound) {
				http.Error(w, deleteScheduledTaskErrMessage, http.StatusNotFound)
				return
			}
			slog.Error("get scheduled task failed", slog.String("error", err.Error()))
			http.Error(w, deleteScheduledTaskErrMessage, http.StatusInternalServerError)
			return
		}

		// Verify the scheduled task belongs to the tenant and device
		if scheduledTask.Tenant.ID != domain.ID(tenantID) || scheduledTask.Device.ID != domain.ID(deviceID) {
			http.Error(w, deleteScheduledTaskErrMessage, http.StatusNotFound)
			return
		}

		// Check if already deleted
		if scheduledTask.IsDeleted() {
			http.Error(w, deleteScheduledTaskErrMessage, http.StatusConflict)
			return
		}

		err = c.service.Delete(r.Context(), domain.ID(id))
		if err != nil {
			slog.Error("delete scheduled task failed", slog.String("error", err.Error()))
			http.Error(w, deleteScheduledTaskErrMessage, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
