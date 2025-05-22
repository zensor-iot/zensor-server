package httpapi

import (
	"log/slog"
	"net/http"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/control_plane/httpapi/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/httpserver"
	"zensor-server/internal/infra/utils"
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

		response := internal.EvaluationRuleSetResponse{
			EvaluationRules: make([]internal.EvaluationRuleResponse, len(items)),
		}

		for i, item := range items {
			response.EvaluationRules[i] = internal.EvaluationRuleResponse{
				Device:      device.ID.String(),
				Description: item.Description,
				Kind:        item.Kind,
				Parameters: utils.Map(item.Parameters, func(inputParam domain.EvaluationRuleParameter) internal.EvaluationRuleParametersResponse {
					return internal.EvaluationRuleParametersResponse{
						Key:   inputParam.Key,
						Value: inputParam.Value,
					}
				}),
				Enabled: item.Enabled,
			}
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
			slog.Debug("decoding body failed", slog.String("error", err.Error()))
			http.Error(w, createEvaluationRuleErrMessage, http.StatusBadRequest)
			return
		}

		params := make([]domain.EvaluationRuleParameter, len(body.Parameters))
		for i, p := range body.Parameters {
			params[i] = domain.EvaluationRuleParameter{Key: p.Key, Value: p.Value}
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

		err = c.evaluationRuleService.AddToDevice(r.Context(), device, evaluationRule)
		if err != nil {
			http.Error(w, createEvaluationRuleErrMessage, http.StatusInternalServerError)
			return
		}

		response := internal.EvaluationRuleResponse{
			Device:      device.ID.String(),
			Description: evaluationRule.Description,
			Kind:        evaluationRule.Kind,
			Parameters: utils.Map(evaluationRule.Parameters, func(inputParam domain.EvaluationRuleParameter) internal.EvaluationRuleParametersResponse {
				return internal.EvaluationRuleParametersResponse{
					Key:   inputParam.Key,
					Value: inputParam.Value,
				}
			}),
			Enabled: evaluationRule.Enabled,
		}

		httpserver.ReplyJSONResponse(w, http.StatusCreated, response)
	}
}
