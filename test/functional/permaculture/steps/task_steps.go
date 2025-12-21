package steps

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

func (fc *FeatureContext) theResponseShouldContainCommandDetails() error {
	if fc.responseData == nil {
		var data map[string]any
		err := fc.decodeBody(fc.response.Body, &data)
		fc.require.NoError(err)
		fc.responseData = data
	}

	fc.require.Contains(fc.responseData, "commands")
	commands, ok := fc.responseData["commands"].([]any)
	fc.require.True(ok, "commands should be an array")
	fc.require.NotEmpty(commands, "commands array should not be empty")

	for i, cmd := range commands {
		command, ok := cmd.(map[string]any)
		fc.require.True(ok, "command %d should be an object", i)

		fc.require.Contains(command, "id", "command %d should have id", i)
		fc.require.Contains(command, "index", "command %d should have index", i)
		fc.require.Contains(command, "value", "command %d should have value", i)
		fc.require.Contains(command, "priority", "command %d should have priority", i)
		fc.require.Contains(command, "dispatch_after", "command %d should have dispatch_after", i)
		fc.require.Contains(command, "ready", "command %d should have ready", i)
		fc.require.Contains(command, "sent", "command %d should have sent", i)
		fc.require.Contains(command, "port", "command %d should have port", i)

		fc.require.IsType("", command["id"], "command %d id should be string", i)
		fc.require.IsType(float64(0), command["index"], "command %d index should be number", i)
		fc.require.IsType(float64(0), command["value"], "command %d value should be number", i)
		fc.require.IsType("", command["priority"], "command %d priority should be string", i)
		fc.require.IsType("", command["dispatch_after"], "command %d dispatch_after should be string", i)
		fc.require.IsType(false, command["ready"], "command %d ready should be boolean", i)
		fc.require.IsType(false, command["sent"], "command %d sent should be boolean", i)
		fc.require.IsType(float64(0), command["port"], "command %d port should be number", i)
	}

	return nil
}

