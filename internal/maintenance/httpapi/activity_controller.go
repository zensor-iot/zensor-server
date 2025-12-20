package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"zensor-server/internal/infra/httpserver"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
	"zensor-server/internal/maintenance/httpapi/internal"
	"zensor-server/internal/maintenance/usecases"
	shareddomain "zensor-server/internal/shared_kernel/domain"
)

const (
	createActivityErrMessage         = "failed to create maintenance activity"
	getActivityErrMessage            = "failed to get maintenance activity"
	updateActivityErrMessage         = "failed to update maintenance activity"
	deleteActivityErrMessage         = "failed to delete maintenance activity"
	activateActivityErrMessage       = "failed to activate maintenance activity"
	deactivateActivityErrMessage     = "failed to deactivate maintenance activity"
	activityNotFoundErrMessage       = "maintenance activity not found"
	activityAlreadyDeletedErrMessage = "maintenance activity is already deleted"
)

func NewActivityController(service usecases.ActivityService) *ActivityController {
	return &ActivityController{
		service: service,
	}
}

var _ httpserver.Controller = &ActivityController{}

type ActivityController struct {
	service usecases.ActivityService
}

func (c *ActivityController) AddRoutes(router *http.ServeMux) {
	router.Handle("GET /v1/maintenance/activities", c.listActivities())
	router.Handle("GET /v1/maintenance/activities/{id}", c.getActivity())
	router.Handle("POST /v1/maintenance/activities", c.createActivity())
	router.Handle("PUT /v1/maintenance/activities/{id}", c.updateActivity())
	router.Handle("DELETE /v1/maintenance/activities/{id}", c.deleteActivity())
	router.Handle("POST /v1/maintenance/activities/{id}/activate", c.activateActivity())
	router.Handle("POST /v1/maintenance/activities/{id}/deactivate", c.deactivateActivity())
}

func (c *ActivityController) listActivities() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.URL.Query().Get("tenant_id")
		if tenantID == "" {
			http.Error(w, "tenant_id is required", http.StatusBadRequest)
			return
		}

		paginationParams := httpserver.ExtractPaginationParams(r)
		pagination := usecases.Pagination{
			Limit:  paginationParams.Limit,
			Offset: (paginationParams.Page - 1) * paginationParams.Limit,
		}

		activities, total, err := c.service.ListActivitiesByTenant(r.Context(), shareddomain.ID(tenantID), pagination)
		if err != nil {
			slog.Error("listing maintenance activities", slog.String("error", err.Error()))
			http.Error(w, "failed to list maintenance activities", http.StatusInternalServerError)
			return
		}

		activityResponses := make([]internal.ActivityResponse, len(activities))
		for i, activity := range activities {
			activityResponses[i] = internal.ToActivityResponse(activity)
		}

		httpserver.ReplyWithPaginatedData(w, http.StatusOK, activityResponses, total, paginationParams)
	}
}

func (c *ActivityController) getActivity() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		activity, err := c.service.GetActivity(r.Context(), shareddomain.ID(id))
		if err != nil {
			if errors.Is(err, usecases.ErrActivityNotFound) {
				http.Error(w, activityNotFoundErrMessage, http.StatusNotFound)
				return
			}
			slog.Error("getting maintenance activity", slog.String("error", err.Error()))
			http.Error(w, getActivityErrMessage, http.StatusInternalServerError)
			return
		}

		response := internal.ToActivityResponse(activity)
		httpserver.ReplyJSONResponse(w, http.StatusOK, response)
	}
}

func (c *ActivityController) createActivity() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body internal.ActivityCreateRequest
		err := httpserver.DecodeJSONBody(r, &body)
		if err != nil {
			http.Error(w, createActivityErrMessage, http.StatusBadRequest)
			return
		}

		var activityType maintenanceDomain.ActivityType
		if predefinedType, found := maintenanceDomain.PredefinedActivityTypes[body.TypeName]; found {
			activityType = predefinedType
		} else {
			activityType, _ = maintenanceDomain.NewActivityTypeBuilder().
				WithName(body.TypeName).
				WithDisplayName(body.TypeName).
				WithDescription("").
				WithIsPredefined(false).
				Build()
		}

		fields := []maintenanceDomain.FieldDefinition{}
		for _, fieldReq := range body.Fields {
			field := maintenanceDomain.FieldDefinition{
				Name:        shareddomain.Name(fieldReq.Name),
				DisplayName: shareddomain.DisplayName(fieldReq.DisplayName),
				Type:        maintenanceDomain.FieldType(fieldReq.Type),
				IsRequired:  fieldReq.IsRequired,
			}
			if fieldReq.DefaultValue != nil {
				defaultValue := interface{}(*fieldReq.DefaultValue)
				field.DefaultValue = &defaultValue
			}
			fields = append(fields, field)
		}

		if len(fields) == 0 && activityType.IsPredefined {
			fields = activityType.Fields
		}

		activity, err := maintenanceDomain.NewActivityBuilder().
			WithTenantID(shareddomain.ID(body.TenantID)).
			WithType(activityType).
			WithName(body.Name).
			WithDescription(body.Description).
			WithSchedule(body.Schedule).
			WithNotificationDaysBefore(body.NotificationDaysBefore).
			WithFields(fields).
			Build()

		if err != nil {
			http.Error(w, createActivityErrMessage, http.StatusBadRequest)
			return
		}

		err = c.service.CreateActivity(r.Context(), activity)
		if err != nil {
			slog.Error("creating maintenance activity", slog.String("error", err.Error()))
			http.Error(w, createActivityErrMessage, http.StatusInternalServerError)
			return
		}

		response := internal.ToActivityResponse(activity)
		httpserver.ReplyJSONResponse(w, http.StatusCreated, response)
	}
}

func (c *ActivityController) updateActivity() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var body internal.ActivityUpdateRequest
		err := httpserver.DecodeJSONBody(r, &body)
		if err != nil {
			http.Error(w, updateActivityErrMessage, http.StatusBadRequest)
			return
		}

		activity, err := c.service.GetActivity(r.Context(), shareddomain.ID(id))
		if err != nil {
			if errors.Is(err, usecases.ErrActivityNotFound) {
				http.Error(w, activityNotFoundErrMessage, http.StatusNotFound)
				return
			}
			http.Error(w, updateActivityErrMessage, http.StatusInternalServerError)
			return
		}

		if body.Name != nil {
			activity.Name = shareddomain.Name(*body.Name)
		}
		if body.Description != nil {
			activity.Description = shareddomain.Description(*body.Description)
		}
		if body.Schedule != nil {
			activity.Schedule = maintenanceDomain.Schedule(*body.Schedule)
		}
		if body.NotificationDaysBefore != nil {
			activity.NotificationDaysBefore = maintenanceDomain.Days(*body.NotificationDaysBefore)
		}
		if body.Fields != nil {
			fields := []maintenanceDomain.FieldDefinition{}
			for _, fieldReq := range *body.Fields {
				field := maintenanceDomain.FieldDefinition{
					Name:        shareddomain.Name(fieldReq.Name),
					DisplayName: shareddomain.DisplayName(fieldReq.DisplayName),
					Type:        maintenanceDomain.FieldType(fieldReq.Type),
					IsRequired:  fieldReq.IsRequired,
				}
				if fieldReq.DefaultValue != nil {
					defaultValue := interface{}(*fieldReq.DefaultValue)
					field.DefaultValue = &defaultValue
				}
				fields = append(fields, field)
			}
			activity.Fields = fields
		}

		err = c.service.UpdateActivity(r.Context(), activity)
		if err != nil {
			slog.Error("updating maintenance activity", slog.String("error", err.Error()))
			http.Error(w, updateActivityErrMessage, http.StatusInternalServerError)
			return
		}

		response := internal.ToActivityResponse(activity)
		httpserver.ReplyJSONResponse(w, http.StatusOK, response)
	}
}

func (c *ActivityController) deleteActivity() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		err := c.service.DeleteActivity(r.Context(), shareddomain.ID(id))
		if err != nil {
			if errors.Is(err, usecases.ErrActivityNotFound) {
				http.Error(w, activityNotFoundErrMessage, http.StatusNotFound)
				return
			}
			slog.Error("deleting maintenance activity", slog.String("error", err.Error()))
			http.Error(w, deleteActivityErrMessage, http.StatusInternalServerError)
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusNoContent, nil)
	}
}

func (c *ActivityController) activateActivity() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		err := c.service.ActivateActivity(r.Context(), shareddomain.ID(id))
		if err != nil {
			if errors.Is(err, usecases.ErrActivityNotFound) {
				http.Error(w, activityNotFoundErrMessage, http.StatusNotFound)
				return
			}
			slog.Error("activating maintenance activity", slog.String("error", err.Error()))
			http.Error(w, activateActivityErrMessage, http.StatusInternalServerError)
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusOK, nil)
	}
}

func (c *ActivityController) deactivateActivity() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		err := c.service.DeactivateActivity(r.Context(), shareddomain.ID(id))
		if err != nil {
			if errors.Is(err, usecases.ErrActivityNotFound) {
				http.Error(w, activityNotFoundErrMessage, http.StatusNotFound)
				return
			}
			slog.Error("deactivating maintenance activity", slog.String("error", err.Error()))
			http.Error(w, deactivateActivityErrMessage, http.StatusInternalServerError)
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusOK, nil)
	}
}
