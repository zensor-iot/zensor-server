package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"time"
	"zensor-server/internal/shared_kernel/domain"
	"zensor-server/internal/control_plane/httpapi/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/httpserver"
	"zensor-server/internal/infra/utils"
)

const (
	createTaskErrMessage     = "failed to create task"
	commandOverlapErrMessage = "command overlap detected with existing pending commands"
)

func NewTaskController(
	service usecases.TaskService,
	deviceService usecases.DeviceService,
) *TaskController {
	return &TaskController{
		service:       service,
		deviceService: deviceService,
	}
}

var _ httpserver.Controller = &TaskController{}

type TaskController struct {
	service       usecases.TaskService
	deviceService usecases.DeviceService
}

func (c *TaskController) AddRoutes(router *http.ServeMux) {
	router.Handle("POST /v1/devices/{id}/tasks", c.create())
	router.Handle("GET /v1/devices/{id}/tasks", c.getByDevice())
}

func (c *TaskController) create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		device, err := c.deviceService.GetDevice(r.Context(), domain.ID(id))
		if err != nil {
			slog.Error("get device failed", slog.String("error", err.Error()))
			http.Error(w, createTaskErrMessage, http.StatusInternalServerError)
			return

		}

		var body internal.TaskCreateRequest
		err = httpserver.DecodeJSONBody(r, &body)
		if err != nil {
			slog.Error("decoding json body", slog.String("error", err.Error()))
			http.Error(w, createTaskErrMessage, http.StatusBadRequest)
			return
		}

		cmds := make([]domain.Command, len(body.Commands))
		for i, item := range body.Commands {
			cmdBuilder := domain.NewCommandBuilder()
			cmd, err := cmdBuilder.
				WithDevice(device).
				WithPayload(domain.CommandPayload{
					Index: domain.Index(item.Index),
					Value: domain.CommandValue(item.Value),
				}).
				WithPriority(domain.CommandPriority(item.Priority)).
				WithDispatchAfter(utils.Time{Time: time.Now().Add(time.Duration(item.WaitFor))}).
				Build()
			if err != nil {
				slog.Error("build command", slog.String("error", err.Error()))
				http.Error(w, createTaskErrMessage, http.StatusInternalServerError)
				return
			}

			cmds[i] = cmd
		}

		task, err := domain.NewTaskBuilder().
			WithDevice(device).
			WithCommands(cmds).
			Build()
		if err != nil {
			slog.Error("build task", slog.String("error", err.Error()))
			http.Error(w, createTaskErrMessage, http.StatusInternalServerError)
			return
		}

		// Set the Task field on each command so they have the correct task reference
		for i := range task.Commands {
			task.Commands[i].Task = task
		}

		err = c.service.Create(r.Context(), task)
		if errors.Is(err, usecases.ErrCommandOverlap) {
			http.Error(w, commandOverlapErrMessage, http.StatusConflict)
			return
		}
		if err != nil {
			slog.Error("create task failed", slog.String("error", err.Error()))
			http.Error(w, createTaskErrMessage, http.StatusInternalServerError)
			return
		}

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

		response := internal.TaskResponse{
			ID:        task.ID.String(),
			Commands:  commandResponses,
			CreatedAt: task.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		}

		httpserver.ReplyJSONResponse(w, http.StatusCreated, response)
	}
}

func (c *TaskController) getByDevice() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		params := httpserver.ExtractPaginationParams(r)
		pagination := usecases.Pagination{Limit: params.Limit, Offset: (params.Page - 1) * params.Limit}

		tasks, total, err := c.service.FindAllByDevice(r.Context(), domain.ID(id), pagination)
		if err != nil {
			slog.Error("get tasks by device failed", slog.String("error", err.Error()))
			http.Error(w, "failed to get tasks", http.StatusInternalServerError)
			return
		}

		responses := make([]internal.TaskResponse, len(tasks))
		for i, task := range tasks {
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

		httpserver.ReplyWithPaginatedData(w, http.StatusOK, responses, total, params)
	}
}
