package steps

import (
	"net/http"
)

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
	var rulesList map[string][]map[string]any
	err := fc.decodeBody(fc.response.Body, &rulesList)
	fc.require.NoError(err)

	found := false
	for _, rule := range rulesList["evaluation_rules"] {
		if rule["id"] == fc.evaluationRuleID {
			found = true
			break
		}
	}
	fc.require.True(found, "Evaluation rule not found in list")
	return nil
}
