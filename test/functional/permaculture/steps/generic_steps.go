package steps

import (
	"strconv"
	"strings"
	"time"
)

// Generic step implementations.
func (fc *FeatureContext) waitForDuration(duration string) error {
	// Parse duration string (e.g., "250ms", "1s", "500ms")
	duration = strings.TrimSpace(duration)

	// Handle common time units
	var d time.Duration

	switch {
	case strings.HasSuffix(duration, "ms"):
		// Parse milliseconds
		msStr := strings.TrimSuffix(duration, "ms")
		ms, err := strconv.Atoi(msStr)
		if err != nil {
			return err
		}
		d = time.Duration(ms) * time.Millisecond
	case strings.HasSuffix(duration, "s"):
		// Parse seconds
		sStr := strings.TrimSuffix(duration, "s")
		s, err := strconv.Atoi(sStr)
		if err != nil {
			return err
		}
		d = time.Duration(s) * time.Second
	default:
		// Try to parse as milliseconds by default
		ms, err := strconv.Atoi(duration)
		if err != nil {
			return err
		}
		d = time.Duration(ms) * time.Millisecond
	}

	time.Sleep(d)
	return nil
}

func (fc *FeatureContext) theResponseStatusCodeShouldBe(code int) error {
	fc.require.Equal(code, fc.response.StatusCode, "Unexpected status code")
	return nil
}

func (fc *FeatureContext) theResponseShouldContainTheTenantDetails() error {
	var data map[string]any
	err := fc.decodeBody(fc.response.Body, &data)
	fc.require.NoError(err)
	fc.require.NotEmpty(data["id"])
	id, ok := data["id"].(string)
	fc.require.True(ok, "Tenant id should be a string")
	fc.tenantID = id
	fc.responseData = data
	return nil
}
