package steps

import (
	"fmt"
	"net/http"
	"time"
)

func (fc *FeatureContext) iCreateAMaintenanceActivityForTenantWithTypeAndName(typeName, name string) error {
	fields := []map[string]any{}
	if typeName == "custom_filter" {
		fields = []map[string]any{
			{
				"name":         "filter_type",
				"display_name": "Filter Type",
				"type":         "text",
				"is_required":  true,
			},
		}
	}
	resp, err := fc.apiDriver.CreateMaintenanceActivity(
		fc.tenantID,
		typeName,
		name,
		"Test maintenance activity",
		"0 0 * * *",
		[]int{1, 3},
		fields,
	)
	fc.require.NoError(err)
	fc.response = resp

	if resp.StatusCode == http.StatusCreated {
		var data map[string]any
		err = fc.decodeBody(resp.Body, &data)
		fc.require.NoError(err)
		fc.responseData = data
		if id, ok := data["id"].(string); ok {
			fc.maintenanceActivityID = id
		}
	}

	return nil
}

func (fc *FeatureContext) iCreateAMaintenanceActivityForTenantWithCustomTypeAndName(customTypeName, name string) error {
	fields := []map[string]any{
		{
			"name":         "custom_field",
			"display_name": "Custom Field",
			"type":         "text",
			"is_required":  false,
		},
	}
	resp, err := fc.apiDriver.CreateMaintenanceActivity(
		fc.tenantID,
		"custom",
		name,
		"Test custom maintenance activity",
		"0 0 * * *",
		[]int{1},
		fields,
	)
	fc.require.NoError(err)
	fc.response = resp

	if resp.StatusCode == http.StatusCreated {
		var data map[string]any
		err = fc.decodeBody(resp.Body, &data)
		fc.require.NoError(err)
		fc.responseData = data
		if id, ok := data["id"].(string); ok {
			fc.maintenanceActivityID = id
		}
	}

	return nil
}

func (fc *FeatureContext) aMaintenanceActivityExistsForTenantWithTypeAndName(typeName, name string) error {
	return fc.iCreateAMaintenanceActivityForTenantWithTypeAndName(typeName, name)
}

func (fc *FeatureContext) aDeactivatedMaintenanceActivityExistsForTenantWithTypeAndName(typeName, name string) error {
	err := fc.aMaintenanceActivityExistsForTenantWithTypeAndName(typeName, name)
	fc.require.NoError(err)
	time.Sleep(50 * time.Millisecond)
	resp, err := fc.apiDriver.DeactivateMaintenanceActivity(fc.maintenanceActivityID)
	fc.require.NoError(err)
	fc.require.Equal(http.StatusOK, resp.StatusCode)
	return nil
}

func (fc *FeatureContext) iListAllMaintenanceActivitiesForTheTenant() error {
	resp, err := fc.apiDriver.ListMaintenanceActivities(fc.tenantID, 0, 0)
	fc.require.NoError(err)
	fc.response = resp
	return nil
}

func (fc *FeatureContext) theListShouldContainTheMaintenanceActivityWithName(name string) error {
	items, err := fc.decodePaginatedResponse(fc.response)
	fc.require.NoError(err)

	found := false
	for _, item := range items {
		if item["name"] == name {
			found = true
			break
		}
	}
	fc.require.True(found, fmt.Sprintf("Maintenance activity with name %s not found in list", name))
	return nil
}

func (fc *FeatureContext) iGetTheMaintenanceActivityByItsID() error {
	resp, err := fc.apiDriver.GetMaintenanceActivity(fc.maintenanceActivityID)
	fc.require.NoError(err)
	fc.response = resp
	fc.responseData = nil
	return nil
}

func (fc *FeatureContext) theResponseShouldContainTheMaintenanceActivityWithName(name string) error {
	var data map[string]any
	if fc.responseData != nil {
		data = fc.responseData
	} else {
		err := fc.decodeBody(fc.response.Body, &data)
		fc.require.NoError(err)
		fc.responseData = data
	}
	fc.require.Equal(name, data["name"])
	return nil
}

func (fc *FeatureContext) theResponseShouldContainTheMaintenanceActivityDetails() error {
	var data map[string]any
	if fc.responseData != nil {
		data = fc.responseData
	} else {
		err := fc.decodeBody(fc.response.Body, &data)
		fc.require.NoError(err)
		fc.responseData = data
	}
	fc.require.NotEmpty(data["id"])
	fc.require.NotEmpty(data["name"])
	fc.require.NotEmpty(data["type_name"])
	fc.require.NotEmpty(data["tenant_id"])
	return nil
}

func (fc *FeatureContext) iUpdateTheMaintenanceActivityWithName(newName string) error {
	updates := map[string]any{
		"name": newName,
	}
	resp, err := fc.apiDriver.UpdateMaintenanceActivity(fc.maintenanceActivityID, updates)
	fc.require.NoError(err)
	fc.response = resp
	fc.responseData = nil
	return nil
}

func (fc *FeatureContext) iActivateTheMaintenanceActivity() error {
	resp, err := fc.apiDriver.ActivateMaintenanceActivity(fc.maintenanceActivityID)
	fc.require.NoError(err)
	fc.response = resp
	fc.responseData = nil
	return nil
}

func (fc *FeatureContext) iDeactivateTheMaintenanceActivity() error {
	resp, err := fc.apiDriver.DeactivateMaintenanceActivity(fc.maintenanceActivityID)
	fc.require.NoError(err)
	fc.response = resp
	fc.responseData = nil
	return nil
}

func (fc *FeatureContext) theResponseShouldContainAnActiveMaintenanceActivity() error {
	var data map[string]any
	if fc.responseData != nil {
		data = fc.responseData
	} else {
		err := fc.decodeBody(fc.response.Body, &data)
		fc.require.NoError(err)
		fc.responseData = data
	}
	fc.require.True(data["is_active"].(bool), "Maintenance activity should be active")
	return nil
}

func (fc *FeatureContext) theResponseShouldContainAnInactiveMaintenanceActivity() error {
	var data map[string]any
	if fc.responseData != nil {
		data = fc.responseData
	} else {
		err := fc.decodeBody(fc.response.Body, &data)
		fc.require.NoError(err)
		fc.responseData = data
	}
	fc.require.False(data["is_active"].(bool), "Maintenance activity should be inactive")
	return nil
}

func (fc *FeatureContext) iDeleteTheMaintenanceActivity() error {
	resp, err := fc.apiDriver.DeleteMaintenanceActivity(fc.maintenanceActivityID)
	fc.require.NoError(err)
	fc.response = resp
	return nil
}

func (fc *FeatureContext) thereAreMaintenanceExecutionsForTheActivity(count int) error {
	for i := 0; i < count; i++ {
		scheduledDate := time.Now().Add(time.Duration(i+1) * 24 * time.Hour)
		fieldValues := map[string]any{
			"service_type": fmt.Sprintf("Service %d", i+1),
		}

		resp, err := fc.createMaintenanceExecution(fc.maintenanceActivityID, scheduledDate, fieldValues)
		fc.require.NoError(err)

		if resp.StatusCode == http.StatusCreated {
			var data map[string]any
			err = fc.decodeBody(resp.Body, &data)
			if err == nil {
				if id, ok := data["id"].(string); ok {
					if fc.maintenanceExecutionIDs == nil {
						fc.maintenanceExecutionIDs = []string{}
					}
					fc.maintenanceExecutionIDs = append(fc.maintenanceExecutionIDs, id)
					if fc.maintenanceExecutionID == "" {
						fc.maintenanceExecutionID = id
					}
				}
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

func (fc *FeatureContext) aMaintenanceExecutionExistsForTheActivity() error {
	scheduledDate := time.Now().Add(24 * time.Hour)
	fieldValues := map[string]any{
		"service_type": "Regular Service",
	}

	resp, err := fc.createMaintenanceExecution(fc.maintenanceActivityID, scheduledDate, fieldValues)
	fc.require.NoError(err)
	fc.require.Equal(http.StatusCreated, resp.StatusCode)

	var data map[string]any
	err = fc.decodeBody(resp.Body, &data)
	fc.require.NoError(err)
	if id, ok := data["id"].(string); ok {
		if fc.maintenanceExecutionIDs == nil {
			fc.maintenanceExecutionIDs = []string{}
		}
		fc.maintenanceExecutionIDs = append(fc.maintenanceExecutionIDs, id)
		if fc.maintenanceExecutionID == "" {
			fc.maintenanceExecutionID = id
		}
	}

	return nil
}

func (fc *FeatureContext) anOverdueMaintenanceExecutionExistsForTheActivity() error {
	scheduledDate := time.Now().Add(-24 * time.Hour)
	fieldValues := map[string]any{
		"service_type": "Overdue Service",
	}

	resp, err := fc.createMaintenanceExecution(fc.maintenanceActivityID, scheduledDate, fieldValues)
	fc.require.NoError(err)
	fc.require.Equal(http.StatusCreated, resp.StatusCode)

	var data map[string]any
	err = fc.decodeBody(resp.Body, &data)
	fc.require.NoError(err)
	if id, ok := data["id"].(string); ok {
		fc.maintenanceExecutionID = id
		if fc.maintenanceExecutionIDs == nil {
			fc.maintenanceExecutionIDs = []string{}
		}
		fc.maintenanceExecutionIDs = append(fc.maintenanceExecutionIDs, id)
	}

	return nil
}

func (fc *FeatureContext) createMaintenanceExecution(activityID string, scheduledDate time.Time, fieldValues map[string]any) (*http.Response, error) {
	return fc.apiDriver.CreateMaintenanceExecution(activityID, scheduledDate, fieldValues)
}

func (fc *FeatureContext) iListAllMaintenanceExecutionsForTheActivity() error {
	resp, err := fc.apiDriver.ListMaintenanceExecutions(fc.maintenanceActivityID, 0, 0)
	fc.require.NoError(err)
	fc.response = resp
	return nil
}

func (fc *FeatureContext) iShouldReceiveExecutions(count int) error {
	items, err := fc.decodePaginatedResponse(fc.response)
	fc.require.NoError(err)
	fc.require.Len(items, count, fmt.Sprintf("Expected %d executions, got %d", count, len(items)))
	return nil
}

func (fc *FeatureContext) iGetTheMaintenanceExecutionByItsID() error {
	resp, err := fc.apiDriver.GetMaintenanceExecution(fc.maintenanceExecutionID)
	fc.require.NoError(err)
	fc.response = resp
	fc.responseData = nil
	return nil
}

func (fc *FeatureContext) theResponseShouldContainTheMaintenanceExecutionDetails() error {
	var data map[string]any
	if fc.responseData != nil {
		data = fc.responseData
	} else {
		err := fc.decodeBody(fc.response.Body, &data)
		fc.require.NoError(err)
		fc.responseData = data
	}
	fc.require.NotEmpty(data["id"])
	fc.require.NotEmpty(data["activity_id"])
	fc.require.NotEmpty(data["scheduled_date"])
	return nil
}

func (fc *FeatureContext) iMarkTheMaintenanceExecutionAsCompletedBy(completedBy string) error {
	resp, err := fc.apiDriver.MarkMaintenanceExecutionCompleted(fc.maintenanceExecutionID, completedBy)
	fc.require.NoError(err)
	fc.response = resp
	fc.responseData = nil
	return nil
}

func (fc *FeatureContext) theResponseShouldContainACompletedMaintenanceExecution() error {
	var data map[string]any
	if fc.responseData != nil {
		data = fc.responseData
	} else {
		err := fc.decodeBody(fc.response.Body, &data)
		fc.require.NoError(err)
		fc.responseData = data
	}
	fc.require.NotNil(data["completed_at"], "Maintenance execution should be completed")
	fc.require.NotNil(data["completed_by"], "Maintenance execution should have completed_by")
	return nil
}

func (fc *FeatureContext) theResponseShouldContainCompletedBy(completedBy string) error {
	var data map[string]any
	if fc.responseData != nil {
		data = fc.responseData
	} else {
		err := fc.decodeBody(fc.response.Body, &data)
		fc.require.NoError(err)
		fc.responseData = data
	}
	fc.require.Equal(completedBy, data["completed_by"])
	return nil
}

func (fc *FeatureContext) theResponseShouldContainAnOverdueMaintenanceExecution() error {
	var data map[string]any
	if fc.responseData != nil {
		data = fc.responseData
	} else {
		err := fc.decodeBody(fc.response.Body, &data)
		fc.require.NoError(err)
		fc.responseData = data
	}
	fc.require.True(data["is_overdue"].(bool), "Maintenance execution should be overdue")
	return nil
}
