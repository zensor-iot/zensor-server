package driver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type APIDriver struct {
	baseURL string
	client  *http.Client
}

func NewAPIDriver(baseURL string) *APIDriver {
	return &APIDriver{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (d *APIDriver) CreateTenant(name, email, description string) (*http.Response, error) {
	reqBody, err := json.Marshal(map[string]any{
		"name":        name,
		"email":       email,
		"description": description,
	})
	if err != nil {
		panic(err)
	}
	return d.client.Post(fmt.Sprintf("%s/v1/tenants", d.baseURL), "application/json", bytes.NewBuffer(reqBody))
}

func (d *APIDriver) GetTenant(id string) (*http.Response, error) {
	return d.client.Get(fmt.Sprintf("%s/v1/tenants/%s", d.baseURL, id))
}

func (d *APIDriver) ListTenants() (*http.Response, error) {
	return d.client.Get(fmt.Sprintf("%s/v1/tenants", d.baseURL))
}

func (d *APIDriver) UpdateTenant(id, newName string) (*http.Response, error) {
	reqBody, err := json.Marshal(map[string]any{"name": newName})
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/v1/tenants/%s", d.baseURL, id), bytes.NewBuffer(reqBody))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	return d.client.Do(req)
}

func (d *APIDriver) DeactivateTenant(id string) (*http.Response, error) {
	return d.client.Post(fmt.Sprintf("%s/v1/tenants/%s/deactivate", d.baseURL, id), "application/json", nil)
}

func (d *APIDriver) ActivateTenant(id string) (*http.Response, error) {
	return d.client.Post(fmt.Sprintf("%s/v1/tenants/%s/activate", d.baseURL, id), "application/json", nil)
}

func (d *APIDriver) SoftDeleteTenant(id string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/v1/tenants/%s", d.baseURL, id), nil)
	if err != nil {
		panic(err)
	}
	return d.client.Do(req)
}

func (d *APIDriver) CreateDevice(name, displayName string) (*http.Response, error) {
	reqBody, err := json.Marshal(map[string]any{"name": name, "display_name": displayName})
	if err != nil {
		panic(err)
	}
	return d.client.Post(fmt.Sprintf("%s/v1/devices", d.baseURL), "application/json", bytes.NewBuffer(reqBody))
}

func (d *APIDriver) ListDevices() (*http.Response, error) {
	return d.client.Get(fmt.Sprintf("%s/v1/devices", d.baseURL))
}

func (d *APIDriver) GetDevice(id string) (*http.Response, error) {
	return d.client.Get(fmt.Sprintf("%s/v1/devices/%s", d.baseURL, id))
}

func (d *APIDriver) UpdateDevice(id, newDisplayName string) (*http.Response, error) {
	reqBody, err := json.Marshal(map[string]any{"display_name": newDisplayName})
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/v1/devices/%s", d.baseURL, id), bytes.NewBuffer(reqBody))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	return d.client.Do(req)
}

func (d *APIDriver) CreateTask(deviceID string) (*http.Response, error) {
	reqBody, err := json.Marshal(map[string]any{
		"commands": []map[string]any{
			{"index": 1, "value": 100},
		},
	})
	if err != nil {
		panic(err)
	}
	return d.client.Post(fmt.Sprintf("%s/v1/devices/%s/tasks", d.baseURL, deviceID), "application/json", bytes.NewBuffer(reqBody))
}

func (d *APIDriver) CreateScheduledTask(tenantID, deviceID, schedule string) (*http.Response, error) {
	reqBody, err := json.Marshal(map[string]any{
		"schedule": schedule,
		"commands": []map[string]any{
			{"index": 1, "value": 200, "priority": "NORMAL", "wait_for": "0s"},
		},
		"is_active": true,
	})
	if err != nil {
		panic(err)
	}
	return d.client.Post(fmt.Sprintf("%s/v1/tenants/%s/devices/%s/scheduled-tasks", d.baseURL, tenantID, deviceID), "application/json", bytes.NewBuffer(reqBody))
}

func (d *APIDriver) ListScheduledTasks(tenantID, deviceID string) (*http.Response, error) {
	return d.client.Get(fmt.Sprintf("%s/v1/tenants/%s/devices/%s/scheduled-tasks", d.baseURL, tenantID, deviceID))
}

func (d *APIDriver) GetScheduledTask(tenantID, deviceID, scheduledTaskID string) (*http.Response, error) {
	return d.client.Get(fmt.Sprintf("%s/v1/tenants/%s/devices/%s/scheduled-tasks/%s", d.baseURL, tenantID, deviceID, scheduledTaskID))
}

func (d *APIDriver) UpdateScheduledTask(tenantID, deviceID, scheduledTaskID, newSchedule string) (*http.Response, error) {
	reqBody, err := json.Marshal(map[string]any{"schedule": &newSchedule})
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/v1/tenants/%s/devices/%s/scheduled-tasks/%s", d.baseURL, tenantID, deviceID, scheduledTaskID), bytes.NewBuffer(reqBody))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	return d.client.Do(req)
}

func (d *APIDriver) DeleteScheduledTask(tenantID, deviceID, scheduledTaskID string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/v1/tenants/%s/devices/%s/scheduled-tasks/%s", d.baseURL, tenantID, deviceID, scheduledTaskID), nil)
	if err != nil {
		panic(err)
	}
	return d.client.Do(req)
}

func (d *APIDriver) CreateScheduledTaskWithJSON(tenantID, deviceID, requestBody string) (*http.Response, error) {
	return d.client.Post(fmt.Sprintf("%s/v1/tenants/%s/devices/%s/scheduled-tasks", d.baseURL, tenantID, deviceID), "application/json", strings.NewReader(requestBody))
}

func (d *APIDriver) UpdateScheduledTaskWithJSON(tenantID, deviceID, scheduledTaskID, requestBody string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/v1/tenants/%s/devices/%s/scheduled-tasks/%s", d.baseURL, tenantID, deviceID, scheduledTaskID), strings.NewReader(requestBody))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	return d.client.Do(req)
}

func (d *APIDriver) CreateEvaluationRule(deviceID string) (*http.Response, error) {
	reqBody, err := json.Marshal(map[string]any{
		"description": "test rule",
		"kind":        "threshold",
		"parameters": []map[string]any{
			{"key": "threshold", "value": 25},
		},
	})
	if err != nil {
		panic(err)
	}
	return d.client.Post(fmt.Sprintf("%s/v1/devices/%s/evaluation-rules", d.baseURL, deviceID), "application/json", bytes.NewBuffer(reqBody))
}

func (d *APIDriver) ListEvaluationRules(deviceID string) (*http.Response, error) {
	return d.client.Get(fmt.Sprintf("%s/v1/devices/%s/evaluation-rules", d.baseURL, deviceID))
}

func (d *APIDriver) GetTasksByScheduledTask(tenantID, deviceID, scheduledTaskID string, page, limit int) (*http.Response, error) {
	url := fmt.Sprintf("%s/v1/tenants/%s/devices/%s/scheduled-tasks/%s/tasks", d.baseURL, tenantID, deviceID, scheduledTaskID)
	if page > 0 || limit > 0 {
		url += fmt.Sprintf("?page=%d&limit=%d", page, limit)
	}
	return d.client.Get(url)
}

func (d *APIDriver) CreateTaskFromScheduledTask(tenantID, deviceID, scheduledTaskID string) (*http.Response, error) {
	reqBody, err := json.Marshal(map[string]any{
		"scheduled_task_id": scheduledTaskID,
		"commands": []map[string]any{
			{"index": 1, "value": 100, "priority": "NORMAL", "wait_for": "0s"},
		},
	})
	if err != nil {
		panic(err)
	}
	return d.client.Post(fmt.Sprintf("%s/v1/tenants/%s/devices/%s/scheduled-tasks/%s/tasks", d.baseURL, tenantID, deviceID, scheduledTaskID), "application/json", bytes.NewBuffer(reqBody))
}

func (d *APIDriver) CreateTenantConfiguration(tenantID, timezone string) (*http.Response, error) {
	reqBody, err := json.Marshal(map[string]any{
		"timezone": timezone,
	})
	if err != nil {
		panic(err)
	}
	return d.client.Post(fmt.Sprintf("%s/v1/tenants/%s/configuration", d.baseURL, tenantID), "application/json", bytes.NewBuffer(reqBody))
}

func (d *APIDriver) GetTenantConfiguration(tenantID string) (*http.Response, error) {
	return d.client.Get(fmt.Sprintf("%s/v1/tenants/%s/configuration", d.baseURL, tenantID))
}

func (d *APIDriver) UpdateTenantConfiguration(tenantID, timezone string) (*http.Response, error) {
	reqBody, err := json.Marshal(map[string]any{
		"timezone": timezone,
	})
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/v1/tenants/%s/configuration", d.baseURL, tenantID), bytes.NewBuffer(reqBody))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	return d.client.Do(req)
}

func (d *APIDriver) GetHealthz() (*http.Response, error) {
	return d.client.Get(fmt.Sprintf("%s/healthz", d.baseURL))
}
