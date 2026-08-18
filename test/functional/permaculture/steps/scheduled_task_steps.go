package steps

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/cucumber/godog"
)

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

func (fc *FeatureContext) aScheduledTaskExistsForTheTenantAndDeviceWithSchedule(schedule string) error {
	err := fc.aTenantExistsWithNameAndEmail("st-tenant", "st-tenant@example.com")
	fc.require.NoError(err)
	err = fc.aDeviceExistsWithName("st-device")
	fc.require.NoError(err)

	resp, err := fc.apiDriver.CreateScheduledTask(fc.tenantID, fc.deviceID, schedule)
	fc.require.NoError(err)
	defer resp.Body.Close()
	fc.require.Equal(http.StatusCreated, resp.StatusCode)

	var data map[string]any
	err = fc.decodeBody(resp.Body, &data)
	fc.require.NoError(err)
	id, ok := data["id"].(string)
	fc.require.True(ok, "Scheduled task id should be a string")
	fc.scheduledTaskID = id
	return nil
}

func (fc *FeatureContext) iCreateAScheduledTaskForTheTenantAndDeviceWithSchedule(schedule string) error {
	resp, err := fc.apiDriver.CreateScheduledTask(fc.tenantID, fc.deviceID, schedule)
	fc.require.NoError(err)
	if err := fc.bufferResponseBody(resp); err != nil {
		return err
	}
	fc.response = resp
	return err
}

func (fc *FeatureContext) theResponseShouldContainTheScheduledTaskDetails() error {
	var data map[string]any
	err := fc.decodeBody(fc.response.Body, &data)
	fc.require.NoError(err)
	fc.require.NotEmpty(data["id"])
	id, ok := data["id"].(string)
	fc.require.True(ok, "Scheduled task id should be a string")
	fc.scheduledTaskID = id
	fc.responseData = data
	return nil
}

func (fc *FeatureContext) iListAllScheduledTasksForTheTenant() error {
	resp, err := fc.apiDriver.ListScheduledTasks(fc.tenantID, fc.deviceID)
	fc.require.NoError(err)
	if err := fc.bufferResponseBody(resp); err != nil {
		return err
	}
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
	if err := fc.bufferResponseBody(resp); err != nil {
		return err
	}
	fc.response = resp
	return err
}

func (fc *FeatureContext) iGetTheScheduledTaskByItsID() error {
	resp, err := fc.apiDriver.GetScheduledTask(fc.tenantID, fc.deviceID, fc.scheduledTaskID)
	fc.require.NoError(err)
	if err := fc.bufferResponseBody(resp); err != nil {
		return err
	}
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

func (fc *FeatureContext) thereAreTasksCreatedFromScheduledTask(count int, scheduledTaskID string) error {
	for range count {
		resp, err := fc.apiDriver.CreateTaskFromScheduledTask(fc.tenantID, fc.deviceID, scheduledTaskID)
		fc.require.NoError(err)
		fc.require.Equal(http.StatusCreated, resp.StatusCode)
		resp.Body.Close()
	}
	return nil
}

func (fc *FeatureContext) iRetrieveTheFirstTasksForScheduledTask(limit int, scheduledTaskID string) error {
	resp, err := fc.apiDriver.GetTasksByScheduledTask(fc.tenantID, fc.deviceID, scheduledTaskID, 1, limit)
	fc.require.NoError(err)
	if err := fc.bufferResponseBody(resp); err != nil {
		return err
	}
	fc.response = resp
	return nil
}

func (fc *FeatureContext) iRetrievePageWithTasksForScheduledTask(page, limit int, scheduledTaskID string) error {
	resp, err := fc.apiDriver.GetTasksByScheduledTask(fc.tenantID, fc.deviceID, scheduledTaskID, page, limit)
	fc.require.NoError(err)
	if err := fc.bufferResponseBody(resp); err != nil {
		return err
	}
	fc.response = resp
	return nil
}

func (fc *FeatureContext) iRetrieveTasksForScheduledTask(scheduledTaskID string) error {
	resp, err := fc.apiDriver.GetTasksByScheduledTask(fc.tenantID, fc.deviceID, scheduledTaskID, 0, 0)
	fc.require.NoError(err)
	if err := fc.bufferResponseBody(resp); err != nil {
		return err
	}
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

	fc.responseData = map[string]any{
		"data":       paginatedResp.Data,
		"pagination": paginatedResp.Pagination,
	}
	return nil
}

func (fc *FeatureContext) theTasksShouldBeSortedByCreationDateInDescendingOrder() error {
	data, ok := fc.responseData["data"].([]map[string]any)
	fc.require.True(ok, "Response data should contain tasks")

	for i := range len(data) - 1 {
		currentCreatedAt, ok := data[i]["created_at"].(string)
		if !ok {
			return errors.New("task created_at is not a string")
		}
		nextCreatedAt, ok := data[i+1]["created_at"].(string)
		if !ok {
			return errors.New("task created_at is not a string")
		}
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

	paginationMap, ok := paginationData.(map[string]any)
	fc.require.True(ok, "Pagination should be a map")

	pageValue, ok := paginationMap["page"].(float64)
	fc.require.True(ok, "Pagination should have a page field")
	fc.require.Equal(page, int(pageValue))
	return nil
}

func (fc *FeatureContext) iTryToRetrieveTasksForNonExistentScheduledTask(scheduledTaskID string) error {
	resp, err := fc.apiDriver.GetTasksByScheduledTask(fc.tenantID, fc.deviceID, scheduledTaskID, 0, 0)
	fc.require.NoError(err)
	if err := fc.bufferResponseBody(resp); err != nil {
		return err
	}
	fc.response = resp
	return nil
}

func (fc *FeatureContext) iTryToRetrieveTasksForScheduledTaskUsingInvalidTenant(scheduledTaskID, invalidTenant string) error {
	resp, err := fc.apiDriver.GetTasksByScheduledTask(invalidTenant, fc.deviceID, scheduledTaskID, 0, 0)
	fc.require.NoError(err)
	if err := fc.bufferResponseBody(resp); err != nil {
		return err
	}
	fc.response = resp
	return nil
}

func (fc *FeatureContext) iTryToRetrieveTasksForScheduledTaskUsingInvalidDevice(scheduledTaskID, invalidDevice string) error {
	resp, err := fc.apiDriver.GetTasksByScheduledTask(fc.tenantID, invalidDevice, scheduledTaskID, 0, 0)
	fc.require.NoError(err)
	if err := fc.bufferResponseBody(resp); err != nil {
		return err
	}
	fc.response = resp
	return nil
}

func (fc *FeatureContext) theOperationShouldFailWithAnError() error {
	fc.require.NotEqual(http.StatusOK, fc.response.StatusCode)
	fc.require.GreaterOrEqual(fc.response.StatusCode, 400, "Expected error status code (4xx or 5xx)")
	return nil
}

func (fc *FeatureContext) aTenantWithId(tenantID string) error {
	resp, err := fc.apiDriver.CreateTenant(tenantID, tenantID+"@example.com", "Test tenant for scheduled task tasks")
	fc.require.NoError(err)
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusCreated:
		var data map[string]any
		err = fc.decodeBody(resp.Body, &data)
		fc.require.NoError(err)
		id, ok := data["id"].(string)
		fc.require.True(ok, "Tenant id should be a string")
		fc.tenantID = id
	case http.StatusConflict:
		listResp, err := fc.apiDriver.ListTenants()
		fc.require.NoError(err)
		defer listResp.Body.Close()
		fc.require.Equal(http.StatusOK, listResp.StatusCode)

		var listData struct {
			Data []map[string]any `json:"data"`
		}
		err = fc.decodeBody(listResp.Body, &listData)
		fc.require.NoError(err)

		for _, tenant := range listData.Data {
			if tenant["name"] == tenantID {
				id, ok := tenant["id"].(string)
				if !ok {
					return errors.New("tenant id is not a string")
				}
				fc.tenantID = id
				return nil
			}
		}
		fc.require.Fail("Tenant with name " + tenantID + " not found in list")
		return nil
	default:
		fc.require.Equal(http.StatusCreated, resp.StatusCode, "Unexpected status code when creating tenant")
	}
	return nil
}

func (fc *FeatureContext) aDeviceWithIdBelongingToTenant(deviceID, tenantID string) error {
	resp, err := fc.apiDriver.CreateDevice(deviceID, deviceID+" Display Name")
	fc.require.NoError(err)
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusCreated:
		var data map[string]any
		err = fc.decodeBody(resp.Body, &data)
		fc.require.NoError(err)
		id, ok := data["id"].(string)
		fc.require.True(ok, "Device id should be a string")
		fc.deviceID = id
	case http.StatusConflict:
		listResp, err := fc.apiDriver.ListDevices()
		fc.require.NoError(err)
		defer listResp.Body.Close()
		fc.require.Equal(http.StatusOK, listResp.StatusCode)

		var listData struct {
			Data []map[string]any `json:"data"`
		}
		err = fc.decodeBody(listResp.Body, &listData)
		fc.require.NoError(err)

		for _, device := range listData.Data {
			if device["name"] == deviceID {
				id, ok := device["id"].(string)
				if !ok {
					return errors.New("device id is not a string")
				}
				fc.deviceID = id
				return nil
			}
		}
		fc.require.Fail("Device with name " + deviceID + " not found in list")
		return nil
	default:
		fc.require.Equal(http.StatusCreated, resp.StatusCode, "Unexpected status code when creating device")
	}
	return nil
}

func (fc *FeatureContext) aScheduledTaskWithIdForDeviceWithSchedule(scheduledTaskID, deviceID, schedule string) error {
	resp, err := fc.apiDriver.CreateScheduledTask(fc.tenantID, fc.deviceID, schedule)
	fc.require.NoError(err)
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusCreated:
		var data map[string]any
		err = fc.decodeBody(resp.Body, &data)
		fc.require.NoError(err)
		id, ok := data["id"].(string)
		fc.require.True(ok, "Scheduled task id should be a string")
		fc.scheduledTaskID = id
	case http.StatusConflict:
		listResp, err := fc.apiDriver.ListScheduledTasks(fc.tenantID, fc.deviceID)
		fc.require.NoError(err)
		defer listResp.Body.Close()
		fc.require.Equal(http.StatusOK, listResp.StatusCode)

		var paginatedResp struct {
			Data []map[string]any `json:"data"`
		}
		err = fc.decodeBody(listResp.Body, &paginatedResp)
		fc.require.NoError(err)
		fc.require.NotEmpty(paginatedResp.Data, "Should have at least one scheduled task")

		scheduledTaskID, ok := paginatedResp.Data[0]["id"].(string)
		fc.require.True(ok, "Scheduled task id should be a string")
		fc.scheduledTaskID = scheduledTaskID
	default:
		fc.require.Equal(http.StatusCreated, resp.StatusCode, "Unexpected status code when creating scheduled task")
	}
	return nil
}

func (fc *FeatureContext) iDeleteTheScheduledTask() error {
	resp, err := fc.apiDriver.DeleteScheduledTask(fc.tenantID, fc.deviceID, fc.scheduledTaskID)
	fc.require.NoError(err)
	if err := fc.bufferResponseBody(resp); err != nil {
		return err
	}
	fc.response = resp
	return err
}

func (fc *FeatureContext) iTryToGetTheScheduledTaskByItsID() error {
	resp, err := fc.apiDriver.GetScheduledTask(fc.tenantID, fc.deviceID, fc.scheduledTaskID)
	fc.require.NoError(err)
	if err := fc.bufferResponseBody(resp); err != nil {
		return err
	}
	fc.response = resp
	return err
}

func (fc *FeatureContext) iCreateAScheduledTaskWith(table *godog.Table) error {
	params := make(map[string]string)
	for _, row := range table.Rows[1:] {
		params[row.Cells[0].Value] = row.Cells[1].Value
	}

	requestBody, err := fc.buildScheduledTaskRequestFromParams(params)
	if err != nil {
		return err
	}

	resp, err := fc.apiDriver.CreateScheduledTaskWithJSON(fc.tenantID, fc.deviceID, requestBody)
	fc.require.NoError(err)
	if err := fc.bufferResponseBody(resp); err != nil {
		return err
	}
	fc.response = resp
	return err
}

func (fc *FeatureContext) iUpdateTheScheduledTaskWith(table *godog.Table) error {
	params := make(map[string]string)
	for _, row := range table.Rows[1:] {
		params[row.Cells[0].Value] = row.Cells[1].Value
	}

	requestBody, err := fc.buildScheduledTaskUpdateRequestFromParams(params)
	if err != nil {
		return err
	}

	resp, err := fc.apiDriver.UpdateScheduledTaskWithJSON(fc.tenantID, fc.deviceID, fc.scheduledTaskID, requestBody)
	fc.require.NoError(err)
	if err := fc.bufferResponseBody(resp); err != nil {
		return err
	}
	fc.response = resp
	return err
}

func (fc *FeatureContext) buildScheduledTaskRequestFromParams(params map[string]string) (string, error) {
	request := map[string]interface{}{
		"commands": []map[string]interface{}{
			{
				"index":    fc.parseUint8(params["command_index"]),
				"value":    fc.parseUint8(params["command_value"]),
				"priority": params["command_priority"],
				"wait_for": params["command_wait_for"],
			},
		},
		"is_active": fc.parseBool(params["is_active"]),
	}

	if schedulingType, exists := params["scheduling_type"]; exists {
		scheduling := map[string]interface{}{
			"type": schedulingType,
		}

		if initialDay, exists := params["initial_day"]; exists {
			scheduling["initial_day"] = initialDay + "T00:00:00Z"
		}
		if dayInterval, exists := params["day_interval"]; exists {
			scheduling["day_interval"] = fc.parseInt(dayInterval)
		}
		if executionTime, exists := params["execution_time"]; exists {
			scheduling["execution_time"] = executionTime
		}

		request["scheduling"] = scheduling
	}

	return fc.marshalJSON(request)
}

func (fc *FeatureContext) buildScheduledTaskUpdateRequestFromParams(params map[string]string) (string, error) {
	request := make(map[string]interface{})

	if schedulingType, exists := params["scheduling_type"]; exists {
		scheduling := map[string]interface{}{
			"type": schedulingType,
		}

		if initialDay, exists := params["initial_day"]; exists {
			scheduling["initial_day"] = initialDay + "T00:00:00Z"
		}
		if dayInterval, exists := params["day_interval"]; exists {
			scheduling["day_interval"] = fc.parseInt(dayInterval)
		}
		if executionTime, exists := params["execution_time"]; exists {
			scheduling["execution_time"] = executionTime
		}

		request["scheduling"] = scheduling
	}

	return fc.marshalJSON(request)
}

func (fc *FeatureContext) parseUint8(s string) uint8 {
	val, _ := strconv.ParseUint(s, 10, 8)
	return uint8(val)
}

func (fc *FeatureContext) parseInt(s string) int {
	val, _ := strconv.Atoi(s)
	return val
}

func (fc *FeatureContext) parseBool(s string) bool {
	return s == "true"
}

func (fc *FeatureContext) marshalJSON(data interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

func (fc *FeatureContext) theResponseShouldContainTheScheduledTaskDetailsWithIntervalScheduling() error {
	var data map[string]any
	err := fc.decodeBody(fc.response.Body, &data)
	fc.require.NoError(err)

	scheduling, exists := data["scheduling"]
	fc.require.True(exists, "Response should contain scheduling configuration")

	schedulingMap, ok := scheduling.(map[string]any)
	fc.require.True(ok, "Scheduling should be a map")
	fc.require.Equal("interval", schedulingMap["type"], "Scheduling type should be 'interval'")

	fc.require.NotNil(schedulingMap["initial_day"], "Initial day should be present")
	fc.require.NotNil(schedulingMap["day_interval"], "Day interval should be present")
	fc.require.NotNil(schedulingMap["execution_time"], "Execution time should be present")
	fc.require.NotNil(schedulingMap["next_execution"], "Next execution time should be calculated")

	id, ok := data["id"].(string)
	fc.require.True(ok, "Scheduled task id should be a string")
	fc.scheduledTaskID = id
	fc.responseData = data
	return nil
}

func (fc *FeatureContext) theResponseShouldContainTheScheduledTaskWithNextExecutionTime() error {
	var data map[string]any
	err := fc.decodeBody(fc.response.Body, &data)
	fc.require.NoError(err)

	scheduling, exists := data["scheduling"]
	fc.require.True(exists, "Response should contain scheduling configuration")

	schedulingMap, ok := scheduling.(map[string]any)
	fc.require.True(ok, "Scheduling should be a map")
	nextExecution := schedulingMap["next_execution"]
	fc.require.NotNil(nextExecution, "Next execution time should be present")

	id, ok := data["id"].(string)
	fc.require.True(ok, "Scheduled task id should be a string")
	fc.scheduledTaskID = id
	fc.responseData = data
	return nil
}

func (fc *FeatureContext) theResponseShouldContainTheScheduledTaskDetailsWith3DayInterval() error {
	var data map[string]any
	err := fc.decodeBody(fc.response.Body, &data)
	fc.require.NoError(err)

	scheduling, ok := data["scheduling"].(map[string]any)
	fc.require.True(ok, "Scheduling should be a map")
	fc.require.Equal("interval", scheduling["type"])
	fc.require.InDelta(float64(3), scheduling["day_interval"], 0.0001, "Day interval should be 3")
	fc.require.Equal("15:00", scheduling["execution_time"], "Execution time should be 15:00")

	id, ok := data["id"].(string)
	fc.require.True(ok, "Scheduled task id should be a string")
	fc.scheduledTaskID = id
	fc.responseData = data
	return nil
}

func (fc *FeatureContext) theResponseShouldContainAnErrorAboutMissingInitialDay() error {
	var data map[string]any
	err := fc.decodeBody(fc.response.Body, &data)
	fc.require.NoError(err)

	errorMsg, exists := data["error"]
	fc.require.True(exists, "Response should contain an error")
	fc.require.Contains(errorMsg, "initial_day", "Error should mention missing initial_day")

	return nil
}

func (fc *FeatureContext) theResponseShouldContainTheUpdatedScheduledTaskWithIntervalScheduling() error {
	var data map[string]any
	err := fc.decodeBody(fc.response.Body, &data)
	fc.require.NoError(err)

	scheduling, ok := data["scheduling"].(map[string]any)
	fc.require.True(ok, "Scheduling should be a map")
	fc.require.Equal("interval", scheduling["type"])
	fc.require.InDelta(float64(5), scheduling["day_interval"], 0.0001, "Day interval should be 5")
	fc.require.Equal("10:30", scheduling["execution_time"], "Execution time should be 10:30")

	return nil
}
