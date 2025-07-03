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
	resp, err := fc.apiDriver.ListScheduledTasks(fc.tenantID, fc.deviceID)
	fc.require.NoError(err)
	fc.response = resp
	return err
}

func (fc *FeatureContext) theListShouldContainOurScheduledTask() error {
	body, err := io.ReadAll(fc.response.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var paginatedResp struct {
		Data       []ScheduledTask `json:"data"`
		Pagination struct {
			Page       int `json:"page"`
			Limit      int `json:"limit"`
			Total      int `json:"total"`
			TotalPages int `json:"total_pages"`
		} `json:"pagination"`
	}
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
	fc.updatedSchedule = newSchedule
	resp, err := fc.apiDriver.UpdateScheduledTask(fc.tenantID, fc.deviceID, fc.scheduledTaskID, newSchedule)
	fc.require.NoError(err)
	fc.response = resp
	return err
}

func (fc *FeatureContext) iGetTheScheduledTaskByItsID() error {
	resp, err := fc.apiDriver.GetScheduledTask(fc.tenantID, fc.deviceID, fc.scheduledTaskID)
	fc.require.NoError(err)
	fc.response = resp
	return err
}

func (fc *FeatureContext) theResponseShouldContainTheScheduledTaskWithTheNewSchedule(schedule string) error {
	var data map[string]any
	err := fc.decodeBody(fc.response.Body, &data)
	fc.require.NoError(err)
	fc.require.Equal(schedule, data["schedule"])
	return nil
}

// New high-level step definitions for scheduled task tasks feature

func (fc *FeatureContext) thereAreTasksCreatedFromScheduledTask(count int, scheduledTaskID string) error {
	// Create multiple tasks for the scheduled task
	for i := 0; i < count; i++ {
		resp, err := fc.apiDriver.CreateTaskFromScheduledTask(fc.tenantID, fc.deviceID, scheduledTaskID)
		fc.require.NoError(err)
		fc.require.Equal(http.StatusCreated, resp.StatusCode)
	}
	return nil
}

func (fc *FeatureContext) iRetrieveTheFirstTasksForScheduledTask(limit int, scheduledTaskID string) error {
	resp, err := fc.apiDriver.GetTasksByScheduledTask(fc.tenantID, fc.deviceID, scheduledTaskID, 1, limit)
	fc.require.NoError(err)
	fc.response = resp
	return nil
}

func (fc *FeatureContext) iRetrievePageWithTasksForScheduledTask(page, limit int, scheduledTaskID string) error {
	resp, err := fc.apiDriver.GetTasksByScheduledTask(fc.tenantID, fc.deviceID, scheduledTaskID, page, limit)
	fc.require.NoError(err)
	fc.response = resp
	return nil
}

func (fc *FeatureContext) iRetrieveTasksForScheduledTask(scheduledTaskID string) error {
	resp, err := fc.apiDriver.GetTasksByScheduledTask(fc.tenantID, fc.deviceID, scheduledTaskID, 0, 0)
	fc.require.NoError(err)
	fc.response = resp
	return nil
}

func (fc *FeatureContext) iShouldReceiveTasks(count int) error {
	var paginatedResp struct {
		Data       []map[string]any `json:"data"`
		Pagination struct {
			Page       int `json:"page"`
			Limit      int `json:"limit"`
			Total      int `json:"total"`
			TotalPages int `json:"total_pages"`
		} `json:"pagination"`
	}

	err := fc.decodeBody(fc.response.Body, &paginatedResp)
	fc.require.NoError(err)
	fc.require.Len(paginatedResp.Data, count)

	// Store the response data for reuse
	fc.responseData = map[string]any{
		"data":       paginatedResp.Data,
		"pagination": paginatedResp.Pagination,
	}
	return nil
}

func (fc *FeatureContext) theTasksShouldBeSortedByCreationDateInDescendingOrder() error {
	data, ok := fc.responseData["data"].([]map[string]any)
	fc.require.True(ok, "Response data should contain tasks")

	// Check if tasks are sorted by created_at in descending order
	for i := 0; i < len(data)-1; i++ {
		currentCreatedAt := data[i]["created_at"].(string)
		nextCreatedAt := data[i+1]["created_at"].(string)
		fc.require.GreaterOrEqual(currentCreatedAt, nextCreatedAt, "Tasks should be sorted by created_at in descending order")
	}
	return nil
}

func (fc *FeatureContext) paginationInformationShouldBeIncluded() error {
	pagination, ok := fc.responseData["pagination"]
	fc.require.True(ok, "Response should contain pagination information")
	fc.require.NotNil(pagination)
	return nil
}

func (fc *FeatureContext) thePaginationShouldIndicatePage(page int) error {
	paginationData, ok := fc.responseData["pagination"]
	fc.require.True(ok, "Response should contain pagination information")

	// Convert to map to access the page field
	paginationMap, ok := paginationData.(map[string]any)
	fc.require.True(ok, "Pagination should be a map")

	pageValue, ok := paginationMap["page"].(float64) // JSON numbers are unmarshaled as float64
	fc.require.True(ok, "Pagination should have a page field")
	fc.require.Equal(page, int(pageValue))
	return nil
}

func (fc *FeatureContext) iTryToRetrieveTasksForNonExistentScheduledTask(scheduledTaskID string) error {
	resp, err := fc.apiDriver.GetTasksByScheduledTask(fc.tenantID, fc.deviceID, scheduledTaskID, 0, 0)
	fc.require.NoError(err)
	fc.response = resp
	return nil
}

func (fc *FeatureContext) iTryToRetrieveTasksForScheduledTaskUsingInvalidTenant(scheduledTaskID, invalidTenant string) error {
	resp, err := fc.apiDriver.GetTasksByScheduledTask(invalidTenant, fc.deviceID, scheduledTaskID, 0, 0)
	fc.require.NoError(err)
	fc.response = resp
	return nil
}

func (fc *FeatureContext) iTryToRetrieveTasksForScheduledTaskUsingInvalidDevice(scheduledTaskID, invalidDevice string) error {
	resp, err := fc.apiDriver.GetTasksByScheduledTask(fc.tenantID, invalidDevice, scheduledTaskID, 0, 0)
	fc.require.NoError(err)
	fc.response = resp
	return nil
}

func (fc *FeatureContext) theOperationShouldFailWithAnError() error {
	fc.require.NotEqual(http.StatusOK, fc.response.StatusCode)
	fc.require.True(fc.response.StatusCode >= 400, "Expected error status code (4xx or 5xx)")
	return nil
}

// Background step definitions for scheduled task tasks feature

func (fc *FeatureContext) aTenantWithId(tenantID string) error {
	// Create a tenant with the specified ID
	resp, err := fc.apiDriver.CreateTenant(tenantID, tenantID+"@example.com", "Test tenant for scheduled task tasks")
	fc.require.NoError(err)

	// Accept both 201 (Created) and 409 (Conflict - already exists)
	if resp.StatusCode == http.StatusCreated {
		var data map[string]any
		err = fc.decodeBody(resp.Body, &data)
		fc.require.NoError(err)
		fc.tenantID = data["id"].(string)
	} else if resp.StatusCode == http.StatusConflict {
		// If tenant already exists, try to get it
		getResp, err := fc.apiDriver.GetTenant(tenantID)
		fc.require.NoError(err)
		fc.require.Equal(http.StatusOK, getResp.StatusCode)

		var data map[string]any
		err = fc.decodeBody(getResp.Body, &data)
		fc.require.NoError(err)
		fc.tenantID = data["id"].(string)
	} else {
		fc.require.Equal(http.StatusCreated, resp.StatusCode, "Unexpected status code when creating tenant")
	}
	return nil
}

func (fc *FeatureContext) aDeviceWithIdBelongingToTenant(deviceID, tenantID string) error {
	// Create a device with the specified ID
	resp, err := fc.apiDriver.CreateDevice(deviceID, deviceID+" Display Name")
	fc.require.NoError(err)

	// Accept both 201 (Created) and 409 (Conflict - already exists)
	if resp.StatusCode == http.StatusCreated {
		var data map[string]any
		err = fc.decodeBody(resp.Body, &data)
		fc.require.NoError(err)
		fc.deviceID = data["id"].(string)
	} else if resp.StatusCode == http.StatusConflict {
		// If device already exists, try to get it
		getResp, err := fc.apiDriver.GetDevice(deviceID)
		fc.require.NoError(err)
		fc.require.Equal(http.StatusOK, getResp.StatusCode)

		var data map[string]any
		err = fc.decodeBody(getResp.Body, &data)
		fc.require.NoError(err)
		fc.deviceID = data["id"].(string)
	} else {
		fc.require.Equal(http.StatusCreated, resp.StatusCode, "Unexpected status code when creating device")
	}
	return nil
}

func (fc *FeatureContext) aScheduledTaskWithIdForDeviceWithSchedule(scheduledTaskID, deviceID, schedule string) error {
	// Create a scheduled task with the specified ID
	resp, err := fc.apiDriver.CreateScheduledTask(fc.tenantID, fc.deviceID, schedule)
	fc.require.NoError(err)

	// Accept both 201 (Created) and 409 (Conflict - already exists)
	if resp.StatusCode == http.StatusCreated {
		var data map[string]any
		err = fc.decodeBody(resp.Body, &data)
		fc.require.NoError(err)
		fc.scheduledTaskID = data["id"].(string)
	} else if resp.StatusCode == http.StatusConflict {
		// If scheduled task already exists, we need to find it in the list
		listResp, err := fc.apiDriver.ListScheduledTasks(fc.tenantID, fc.deviceID)
		fc.require.NoError(err)
		fc.require.Equal(http.StatusOK, listResp.StatusCode)

		var paginatedResp struct {
			Data []map[string]any `json:"data"`
		}
		err = fc.decodeBody(listResp.Body, &paginatedResp)
		fc.require.NoError(err)
		fc.require.NotEmpty(paginatedResp.Data, "Should have at least one scheduled task")

		// Use the first scheduled task found
		fc.scheduledTaskID = paginatedResp.Data[0]["id"].(string)
	} else {
		fc.require.Equal(http.StatusCreated, resp.StatusCode, "Unexpected status code when creating scheduled task")
	}
	return nil
}

func (fc *FeatureContext) iDeleteTheScheduledTask() error {
	resp, err := fc.apiDriver.DeleteScheduledTask(fc.tenantID, fc.deviceID, fc.scheduledTaskID)
	fc.require.NoError(err)
	fc.response = resp
	return err
}

func (fc *FeatureContext) iTryToGetTheScheduledTaskByItsID() error {
	resp, err := fc.apiDriver.GetScheduledTask(fc.tenantID, fc.deviceID, fc.scheduledTaskID)
	fc.require.NoError(err)
	fc.response = resp
	return err
}
