package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"zensor-server/internal/infra/httpserver"
	"zensor-server/internal/maintenance/httpapi/internal"
	"zensor-server/internal/maintenance/usecases"
	shareddomain "zensor-server/internal/shared_kernel/domain"
)

const (
	getExecutionErrMessage              = "failed to get maintenance execution"
	markCompletedErrMessage             = "failed to mark execution as completed"
	executionNotFoundErrMessage         = "maintenance execution not found"
	executionAlreadyCompletedErrMessage = "maintenance execution is already completed"
)

func NewMaintenanceExecutionController(
	service usecases.MaintenanceExecutionService,
	activityService usecases.MaintenanceActivityService,
) *MaintenanceExecutionController {
	return &MaintenanceExecutionController{
		service:         service,
		activityService: activityService,
	}
}

var _ httpserver.Controller = &MaintenanceExecutionController{}

type MaintenanceExecutionController struct {
	service         usecases.MaintenanceExecutionService
	activityService usecases.MaintenanceActivityService
}

func (c *MaintenanceExecutionController) AddRoutes(router *http.ServeMux) {
	router.Handle("GET /v1/maintenance/executions", c.listExecutions())
	router.Handle("GET /v1/maintenance/executions/{id}", c.getExecution())
	router.Handle("POST /v1/maintenance/executions/{id}/complete", c.markCompleted())
}

func (c *MaintenanceExecutionController) listExecutions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		activityID := r.URL.Query().Get("activity_id")
		if activityID == "" {
			http.Error(w, "activity_id is required", http.StatusBadRequest)
			return
		}

		paginationParams := httpserver.ExtractPaginationParams(r)
		pagination := usecases.Pagination{
			Limit:  paginationParams.Limit,
			Offset: (paginationParams.Page - 1) * paginationParams.Limit,
		}

		executions, total, err := c.service.ListExecutionsByActivity(r.Context(), shareddomain.ID(activityID), pagination)
		if err != nil {
			slog.Error("listing maintenance executions", slog.String("error", err.Error()))
			http.Error(w, "failed to list maintenance executions", http.StatusInternalServerError)
			return
		}

		executionResponses := make([]internal.MaintenanceExecutionResponse, len(executions))
		for i, execution := range executions {
			executionResponses[i] = internal.ToMaintenanceExecutionResponse(execution)
		}

		httpserver.ReplyWithPaginatedData(w, http.StatusOK, executionResponses, total, paginationParams)
	}
}

func (c *MaintenanceExecutionController) getExecution() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		execution, err := c.service.GetExecution(r.Context(), shareddomain.ID(id))
		if err != nil {
			if errors.Is(err, usecases.ErrMaintenanceExecutionNotFound) {
				http.Error(w, executionNotFoundErrMessage, http.StatusNotFound)
				return
			}
			slog.Error("getting maintenance execution", slog.String("error", err.Error()))
			http.Error(w, getExecutionErrMessage, http.StatusInternalServerError)
			return
		}

		response := internal.ToMaintenanceExecutionResponse(execution)
		httpserver.ReplyJSONResponse(w, http.StatusOK, response)
	}
}

func (c *MaintenanceExecutionController) markCompleted() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var body internal.MaintenanceExecutionCompleteRequest
		err := httpserver.DecodeJSONBody(r, &body)
		if err != nil {
			http.Error(w, markCompletedErrMessage, http.StatusBadRequest)
			return
		}

		err = c.service.MarkExecutionCompleted(r.Context(), shareddomain.ID(id), body.CompletedBy)
		if err != nil {
			if errors.Is(err, usecases.ErrMaintenanceExecutionNotFound) {
				http.Error(w, executionNotFoundErrMessage, http.StatusNotFound)
				return
			}
			slog.Error("marking execution as completed", slog.String("error", err.Error()))
			http.Error(w, markCompletedErrMessage, http.StatusInternalServerError)
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusOK, nil)
	}
}
