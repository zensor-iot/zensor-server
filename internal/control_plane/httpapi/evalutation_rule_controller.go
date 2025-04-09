package httpapi

import (
	"fmt"
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
) *EvaluationRuleController {
	return &EvaluationRuleController{
		evaluationRuleService: evaluationRuleService,
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

		evaluationRule, err := domain.NewEvaluationRuleBuilder().
			WithDescription(body.Description).
			WithMetric(body.Metric).
			WithLowerThreshold(body.LowerThreshold).
			WithUpperThreshold(body.UpperThreshold).
			Build()
		if err != nil {
			http.Error(w, createEvaluationRuleErrMessage, http.StatusBadRequest)
			return
		}

		device.AddEvaluationRule(evaluationRule)

		fmt.Printf("*** %+v\n", device)

		httpserver.ReplyJSONResponse(w, http.StatusNotImplemented, nil)
	}
}
