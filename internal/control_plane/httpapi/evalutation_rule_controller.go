package httpapi

import (
	"log/slog"
	"net/http"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/control_plane/httpapi/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/httpserver"
)

const (
	createEvaluationRuleErrMessage = "failed to create evaluation rule"
)

func NewEvaluationRuleController(
	evaluationRuleService usecases.EvaluationRuleService,
	deviceService usecases.DeviceService,
) *EvaluationRuleController {
	return &EvaluationRuleController{
		evaluationRuleService: evaluationRuleService,
		deviceService:         deviceService,
	}
}

var _ httpserver.Controller = (*EvaluationRuleController)(nil)

type EvaluationRuleController struct {
	evaluationRuleService usecases.EvaluationRuleService
	deviceService         usecases.DeviceService
}

func (c *EvaluationRuleController) AddRoutes(router *http.ServeMux) {
	router.Handle("GET /v1/devices/{id}/evaluation-rules", c.listEvaluationRules())
	router.Handle("POST /v1/devices/{id}/evaluation-rules", c.craeteEvaluationRule())
}

func (c *EvaluationRuleController) listEvaluationRules() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		device := domain.Device{ID: domain.ID(id)}

		items, err := c.evaluationRuleService.FindAllByDevice(r.Context(), device)
		if err != nil {
			http.Error(w, "list evaluation rules failed", http.StatusInternalServerError)
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusOK, items)
	}
}

func (c *EvaluationRuleController) craeteEvaluationRule() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		device, err := c.deviceService.GetDevice(r.Context(), domain.ID(id))
		if err != nil {
			http.Error(w, "device not found", http.StatusNotFound)
			return
		}

		var body internal.EvaluationRuleCreateRequest
		if err := httpserver.DecodeJSONBody(r, &body); err != nil {
			http.Error(w, createDeviceErrMessage, http.StatusBadRequest)
			return
		}

		params := make([]domain.EvaluetionRuleParameter, len(body.Parameters))
		for i, p := range body.Parameters {
			params[i] = domain.EvaluetionRuleParameter{Key: p.Key, Value: p.Value}
		}

		evaluationRule, err := domain.NewEvaluationRuleBuilder().
			WithDescription(body.Description).
			WithKind(body.Kind).
			WithParameters(params...).
			Build()

		if err != nil {
			slog.Warn(createEvaluationRuleErrMessage, slog.String("error", err.Error()))
			http.Error(w, createEvaluationRuleErrMessage, http.StatusBadRequest)
			return
		}

		device.AddEvaluationRule(evaluationRule)

		httpserver.ReplyJSONResponse(w, http.StatusNotImplemented, nil)
	}
}
