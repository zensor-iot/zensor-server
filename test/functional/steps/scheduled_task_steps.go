package steps

import (
	"net/http"
)

// Scheduled Task step implementations
func (fc *FeatureContext) aScheduledTaskExistsForTheTenantAndDeviceWithSchedule(schedule string) error {
	err := fc.aTenantExistsWithNameAndEmail("st-tenant", "st-tenant@example.com")
	fc.require.NoError(err)
	err = fc.aDeviceExistsWithName("st-device")
	fc.require.NoError(err)

	resp, err := fc.apiDriver.CreateScheduledTask(fc.tenantID, fc.deviceID, schedule)
	fc.require.NoError(err)
	fc.require.Equal(http.StatusCreated, resp.StatusCode)

	var data map[string]any
	err = fc.decodeBody(resp.Body, &data)
	fc.require.NoError(err)
	fc.scheduledTaskID = data["id"].(string)
	return nil
}

func (fc *FeatureContext) iCreateAScheduledTaskForTheTenantAndDeviceWithSchedule(schedule string) error {
	resp, err := fc.apiDriver.CreateScheduledTask(fc.tenantID, fc.deviceID, schedule)
	fc.require.NoError(err)
	fc.response = resp
	return err
}

func (fc *FeatureContext) theResponseShouldContainTheScheduledTaskDetails() error {
	var data map[string]any
	err := fc.decodeBody(fc.response.Body, &data)
	fc.require.NoError(err)
	fc.require.NotEmpty(data["id"])
	fc.scheduledTaskID = data["id"].(string)
	fc.responseData = data
	return nil
}

func (fc *FeatureContext) iListAllScheduledTasksForTheTenant() error {
	resp, err := fc.apiDriver.ListScheduledTasks(fc.tenantID)
	fc.require.NoError(err)
	fc.response = resp
	return err
}

func (fc *FeatureContext) theListShouldContainOurScheduledTask() error {
	var tasksList map[string][]map[string]any
	err := fc.decodeBody(fc.response.Body, &tasksList)
	fc.require.NoError(err)

	found := false
	for _, task := range tasksList["scheduled_tasks"] {
		if task["id"] == fc.scheduledTaskID {
			found = true
			break
		}
	}
	fc.require.True(found, "Scheduled task not found in list")
	return nil
}

func (fc *FeatureContext) iUpdateTheScheduledTaskWithANewSchedule(newSchedule string) error {
	resp, err := fc.apiDriver.UpdateScheduledTask(fc.tenantID, fc.scheduledTaskID, newSchedule)
	fc.require.NoError(err)
	fc.response = resp
	return err
}

func (fc *FeatureContext) theResponseShouldContainTheScheduledTaskWithTheNewSchedule() error {
	var data map[string]any
	err := fc.decodeBody(fc.response.Body, &data)
	fc.require.NoError(err)
	fc.require.Equal(fc.scheduledTaskID, data["id"])
	fc.require.Equal(fc.responseData["schedule"], data["schedule"])
	return nil
}
