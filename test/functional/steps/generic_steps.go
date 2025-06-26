package steps

import (
	"strconv"
	"strings"
	"time"
)

// Generic step implementations
func (fc *FeatureContext) waitForDuration(duration string) error {
	// Parse duration string (e.g., "250ms", "1s", "500ms")
	duration = strings.TrimSpace(duration)

	// Handle common time units
	var d time.Duration

	if strings.HasSuffix(duration, "ms") {
		// Parse milliseconds
		msStr := strings.TrimSuffix(duration, "ms")
		ms, err := strconv.Atoi(msStr)
		if err != nil {
			return err
		}
		d = time.Duration(ms) * time.Millisecond
	} else if strings.HasSuffix(duration, "s") {
		// Parse seconds
		sStr := strings.TrimSuffix(duration, "s")
		s, err := strconv.Atoi(sStr)
		if err != nil {
			return err
		}
		d = time.Duration(s) * time.Second
	} else {
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
	fc.tenantID = data["id"].(string)
	fc.responseData = data
	return nil
}
