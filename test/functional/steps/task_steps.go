package steps

// Task step implementations
func (fc *FeatureContext) iCreateATaskForTheDevice() error {
	resp, err := fc.apiDriver.CreateTask(fc.deviceID)
	fc.require.NoError(err)
	fc.response = resp
	return err
}

func (fc *FeatureContext) theResponseShouldContainTheTaskDetails() error {
	var data map[string]any
	err := fc.decodeBody(fc.response.Body, &data)
	fc.require.NoError(err)
	fc.require.NotEmpty(data["id"])
	fc.responseData = data
	return nil
}
