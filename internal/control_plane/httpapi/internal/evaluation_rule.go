package internal

type EvaluationRuleCreateRequest struct {
	Description    string  `json:"description"`
	Metric         string  `json:"metric"`
	LowerThreshold float64 `json:"lower_threshold"`
	UpperThreshold float64 `json:"upper_threshold"`
	Enabled        bool    `json:"enabled"`
}
