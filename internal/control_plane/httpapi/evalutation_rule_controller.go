package httpapi

import (
	"net/http"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/httpserver"
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
}

func (c *EvaluationRuleController) AddRoutes(router *http.ServeMux) {

}
