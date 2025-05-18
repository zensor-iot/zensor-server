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

type EvaluationRuleResponse struct {
	Device string `json:"device"`
}
