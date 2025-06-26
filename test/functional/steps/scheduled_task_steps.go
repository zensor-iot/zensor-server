package steps

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ScheduledTask represents a scheduled task entity in the response
type ScheduledTask struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenant_id"`
	DeviceID    string `json:"device_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Schedule    string `json:"schedule"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

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
	body, err := io.ReadAll(fc.response.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var paginatedResp PaginatedResponse[ScheduledTask]
	if err := json.Unmarshal(body, &paginatedResp); err != nil {
		return fmt.Errorf("failed to decode paginated response: %w", err)
	}

	found := false
	for _, task := range paginatedResp.Data {
		if task.ID == fc.scheduledTaskID {
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

func (fc *FeatureContext) iGetTheScheduledTaskByItsID() error {
	resp, err := fc.apiDriver.GetScheduledTask(fc.tenantID, fc.scheduledTaskID)
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
