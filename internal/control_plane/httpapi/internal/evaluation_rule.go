package internal

type EvaluationRuleCreateRequest struct {
	Description string                                  `json:"description"`
	Kind        string                                  `json:"kind"`
	Parameters  []EvaluationRuleParametersCreateRequest `json:"parameters"`
}

type EvaluationRuleParametersCreateRequest struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

type EvaluationRuleSetResponse struct {
	EvaluationRules []EvaluationRuleResponse `json:"evaluation_rules"`
}

type EvaluationRuleResponse struct {
	ID          string                             `json:"id"`
	Device      string                             `json:"device"`
	Description string                             `json:"description"`
	Kind        string                             `json:"kind"`
	Parameters  []EvaluationRuleParametersResponse `json:"parameters"`
	Enabled     bool                               `json:"enabled"`
}

type EvaluationRuleParametersResponse struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}
