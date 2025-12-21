package steps

import (
	"regexp"
)

// Healthz endpoint step implementations

func (fc *FeatureContext) iCallTheHealthzEndpoint() error {
	response, err := fc.apiDriver.GetHealthz()
	if err != nil {
		return err
	}
	fc.response = response
	return nil
}

func (fc *FeatureContext) theResponseShouldContainStatusInformation() error {
	var data map[string]any
	err := fc.decodeBody(fc.response.Body, &data)
	fc.require.NoError(err)

	// Check that all required fields are present
	fc.require.Contains(data, "status", "Status should be present")
	fc.require.Contains(data, "version", "Version should be present")
	fc.require.Contains(data, "commit_hash", "Commit hash should be present")

	// Validate status field
	status, ok := data["status"].(string)
	fc.require.True(ok, "Status should be a string")
	fc.require.Equal("success", status, "Status should be 'success'")

	fc.responseData = data
	return nil
}

func (fc *FeatureContext) theResponseShouldContainVersionInformation() error {
	version, ok := fc.responseData["version"].(string)
	fc.require.True(ok, "version should be a string")
	fc.require.NotEmpty(version, "version should not be empty")

	return nil
}

func (fc *FeatureContext) theResponseShouldContainCommitHashInformation() error {
	commitHash, ok := fc.responseData["commit_hash"].(string)
	fc.require.True(ok, "commit_hash should be a string")
	fc.require.NotEmpty(commitHash, "commit_hash should not be empty")

	// Accept either a valid git commit hash or "unknown" for development
	if commitHash != "unknown" {
		// Basic commit hash validation (should be alphanumeric and at least 7 characters)
		commitHashRegex := regexp.MustCompile(`^[a-f0-9]{7,}$`)
		fc.require.True(commitHashRegex.MatchString(commitHash), "COMMIT_HASH should be a valid git commit hash")
	}

	return nil
}
