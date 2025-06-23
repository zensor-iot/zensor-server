package steps

import (
	"encoding/json"
	"io"
)

// Generic step implementations
func (fc *FeatureContext) theServiceIsRunning() error {
	return nil
}

func (fc *FeatureContext) theResponseStatusCodeShouldBe(code int) error {
	fc.require.Equal(code, fc.response.StatusCode, "Unexpected status code")
	return nil
}

func (fc *FeatureContext) decodeBody(body io.ReadCloser, v interface{}) error {
	defer body.Close()
	return json.NewDecoder(body).Decode(v)
}

func (fc *FeatureContext) theResponseShouldContainTheTenantDetails() error {
	var data map[string]interface{}
	err := fc.decodeBody(fc.response.Body, &data)
	fc.require.NoError(err)
	fc.require.NotEmpty(data["id"])
	fc.tenantID = data["id"].(string)
	fc.responseData = data
	return nil
}
