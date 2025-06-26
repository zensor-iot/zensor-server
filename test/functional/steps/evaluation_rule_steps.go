package steps

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// EvaluationRule represents an evaluation rule entity in the response
type EvaluationRule struct {
	ID          string `json:"id"`
	DeviceID    string `json:"device_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Condition   string `json:"condition"`
	Action      string `json:"action"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// Evaluation Rule step implementations
func (fc *FeatureContext) anEvaluationRuleExistsForTheDevice() error {
	err := fc.aTenantExistsWithNameAndEmail("er-tenant", "er-tenant@example.com")
	fc.require.NoError(err)
	err = fc.aDeviceExistsWithName("er-device")
	fc.require.NoError(err)

	resp, err := fc.apiDriver.CreateEvaluationRule(fc.deviceID)
	fc.require.NoError(err)
	fc.require.Equal(http.StatusCreated, resp.StatusCode)

	var data map[string]any
	err = fc.decodeBody(resp.Body, &data)
	fc.require.NoError(err)
	fc.evaluationRuleID = data["id"].(string)
	return nil
}

func (fc *FeatureContext) iCreateAnEvaluationRuleForTheDevice() error {
	resp, err := fc.apiDriver.CreateEvaluationRule(fc.deviceID)
	fc.require.NoError(err)
	fc.response = resp
	return err
}

func (fc *FeatureContext) theResponseShouldContainTheEvaluationRuleDetails() error {
	var data map[string]any
	err := fc.decodeBody(fc.response.Body, &data)
	fc.require.NoError(err)
	fc.require.NotEmpty(data["id"])
	fc.evaluationRuleID = data["id"].(string)
	fc.responseData = data
	return nil
}

func (fc *FeatureContext) iListAllEvaluationRulesForTheDevice() error {
	resp, err := fc.apiDriver.ListEvaluationRules(fc.deviceID)
	fc.require.NoError(err)
	fc.response = resp
	return err
}

func (fc *FeatureContext) theListShouldContainOurEvaluationRule() error {
	body, err := io.ReadAll(fc.response.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var paginatedResp PaginatedResponse[EvaluationRule]
	if err := json.Unmarshal(body, &paginatedResp); err != nil {
		return fmt.Errorf("failed to decode paginated response: %w", err)
	}

	found := false
	for _, rule := range paginatedResp.Data {
		if rule.ID == fc.evaluationRuleID {
			found = true
			break
		}
	}
	fc.require.True(found, "Evaluation rule not found in list")
	return nil
}
