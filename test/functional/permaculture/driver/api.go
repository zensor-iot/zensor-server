// Package driver provides the HTTP driver used by the permaculture functional tests.
package driver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"strings"
)

type APIDriver struct {
	baseURL   string
	client    *http.Client
	userEmail string
	ctx       context.Context
}

func NewAPIDriver(baseURL string) *APIDriver {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	return &APIDriver{
		baseURL: baseURL,
		client:  &http.Client{Jar: jar},
		ctx:     context.Background(),
	}
}

func (d *APIDriver) get(url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(d.ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return d.client.Do(req)
}

func (d *APIDriver) post(url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(d.ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return d.client.Do(req)
}

func (d *APIDriver) request(method, url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(d.ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return d.client.Do(req)
}

func (d *APIDriver) Login() (err error) {
	resp, err := d.get(d.baseURL + "/auth/login")
	if err != nil {
		return fmt.Errorf("performing login flow: %w", err)
	}
	if err := resp.Body.Close(); err != nil {
		return fmt.Errorf("closing login response: %w", err)
	}

	meResp, err := d.get(d.baseURL + "/v1/me")
	if err != nil {
		return fmt.Errorf("fetching current user: %w", err)
	}
	defer func() {
		if cerr := meResp.Body.Close(); cerr != nil {
			err = fmt.Errorf("closing current user response: %w", cerr)
		}
	}()

	if meResp.StatusCode != http.StatusOK {
		return fmt.Errorf("login did not establish a session: /v1/me returned %d", meResp.StatusCode)
	}

	var me struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(meResp.Body).Decode(&me); err != nil {
		return fmt.Errorf("decoding current user: %w", err)
	}

	d.userEmail = me.Email
	return nil
}

func (d *APIDriver) CurrentUserEmail() string {
	return d.userEmail
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
	return d.post(d.baseURL+"/v1/tenants", reqBody)
}

func (d *APIDriver) GetTenant(id string) (*http.Response, error) {
	return d.get(fmt.Sprintf("%s/v1/tenants/%s", d.baseURL, id))
}

func (d *APIDriver) ListTenants() (*http.Response, error) {
	return d.get(d.baseURL + "/v1/tenants")
}

func (d *APIDriver) UpdateTenant(id, newName string) (*http.Response, error) {
	reqBody, err := json.Marshal(map[string]any{"name": newName})
	if err != nil {
		panic(err)
	}
	return d.request(http.MethodPut, fmt.Sprintf("%s/v1/tenants/%s", d.baseURL, id), reqBody)
}

func (d *APIDriver) DeactivateTenant(id string) (*http.Response, error) {
	return d.post(fmt.Sprintf("%s/v1/tenants/%s/deactivate", d.baseURL, id), nil)
}

func (d *APIDriver) ActivateTenant(id string) (*http.Response, error) {
	return d.post(fmt.Sprintf("%s/v1/tenants/%s/activate", d.baseURL, id), nil)
}

func (d *APIDriver) SoftDeleteTenant(id string) (*http.Response, error) {
	return d.request(http.MethodDelete, fmt.Sprintf("%s/v1/tenants/%s", d.baseURL, id), nil)
}

func (d *APIDriver) CreateDevice(name, displayName string) (*http.Response, error) {
	reqBody, err := json.Marshal(map[string]any{"name": name, "display_name": displayName})
	if err != nil {
		panic(err)
	}
	return d.post(d.baseURL+"/v1/devices", reqBody)
}

func (d *APIDriver) ListDevices() (*http.Response, error) {
	return d.get(d.baseURL + "/v1/devices")
}

func (d *APIDriver) GetDevice(id string) (*http.Response, error) {
	return d.get(fmt.Sprintf("%s/v1/devices/%s", d.baseURL, id))
}

func (d *APIDriver) UpdateDevice(id, newDisplayName string) (*http.Response, error) {
	reqBody, err := json.Marshal(map[string]any{"display_name": newDisplayName})
	if err != nil {
		panic(err)
	}
	return d.request(http.MethodPut, fmt.Sprintf("%s/v1/devices/%s", d.baseURL, id), reqBody)
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
	return d.post(fmt.Sprintf("%s/v1/devices/%s/evaluation-rules", d.baseURL, deviceID), reqBody)
}

func (d *APIDriver) ListEvaluationRules(deviceID string) (*http.Response, error) {
	return d.get(fmt.Sprintf("%s/v1/devices/%s/evaluation-rules", d.baseURL, deviceID))
}

func (d *APIDriver) CreateTenantConfiguration(tenantID, timezone string) (*http.Response, error) {
	return d.UpsertTenantConfiguration(tenantID, timezone, "")
}

func (d *APIDriver) UpsertTenantConfiguration(tenantID, timezone, userID string) (*http.Response, error) {
	reqBody, err := json.Marshal(map[string]any{
		"timezone": timezone,
	})
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequestWithContext(d.ctx, http.MethodPut, fmt.Sprintf("%s/v1/tenants/%s/configuration", d.baseURL, tenantID), bytes.NewBuffer(reqBody))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if userID != "" {
		req.Header.Set("X-User-Email", userID)
	}
	return d.client.Do(req)
}

func (d *APIDriver) GetTenantConfiguration(tenantID string) (*http.Response, error) {
	return d.get(fmt.Sprintf("%s/v1/tenants/%s/configuration", d.baseURL, tenantID))
}

func (d *APIDriver) UpdateTenantConfiguration(tenantID, timezone string) (*http.Response, error) {
	return d.UpsertTenantConfiguration(tenantID, timezone, "")
}

func (d *APIDriver) GetHealthz() (*http.Response, error) {
	return d.get(d.baseURL + "/healthz")
}

func (d *APIDriver) AssociateUserWithTenants(userID string, tenantIDs []string) (*http.Response, error) {
	reqBody, err := json.Marshal(map[string]any{
		"tenants": tenantIDs,
	})
	if err != nil {
		panic(err)
	}
	return d.request(http.MethodPut, fmt.Sprintf("%s/v1/users/%s", d.baseURL, userID), reqBody)
}

func (d *APIDriver) GetUser(userID string) (*http.Response, error) {
	return d.get(fmt.Sprintf("%s/v1/users/%s", d.baseURL, userID))
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
	return d.post(fmt.Sprintf("%s/v1/devices/%s/tasks", d.baseURL, deviceID), reqBody)
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
	return d.post(fmt.Sprintf("%s/v1/tenants/%s/devices/%s/scheduled-tasks", d.baseURL, tenantID, deviceID), reqBody)
}

func (d *APIDriver) ListScheduledTasks(tenantID, deviceID string) (*http.Response, error) {
	return d.get(fmt.Sprintf("%s/v1/tenants/%s/devices/%s/scheduled-tasks", d.baseURL, tenantID, deviceID))
}

func (d *APIDriver) GetScheduledTask(tenantID, deviceID, scheduledTaskID string) (*http.Response, error) {
	return d.get(fmt.Sprintf("%s/v1/tenants/%s/devices/%s/scheduled-tasks/%s", d.baseURL, tenantID, deviceID, scheduledTaskID))
}

func (d *APIDriver) UpdateScheduledTask(tenantID, deviceID, scheduledTaskID, newSchedule string) (*http.Response, error) {
	reqBody, err := json.Marshal(map[string]any{"schedule": &newSchedule})
	if err != nil {
		panic(err)
	}
	return d.request(http.MethodPut, fmt.Sprintf("%s/v1/tenants/%s/devices/%s/scheduled-tasks/%s", d.baseURL, tenantID, deviceID, scheduledTaskID), reqBody)
}

func (d *APIDriver) DeleteScheduledTask(tenantID, deviceID, scheduledTaskID string) (*http.Response, error) {
	return d.request(http.MethodDelete, fmt.Sprintf("%s/v1/tenants/%s/devices/%s/scheduled-tasks/%s", d.baseURL, tenantID, deviceID, scheduledTaskID), nil)
}

func (d *APIDriver) CreateScheduledTaskWithJSON(tenantID, deviceID, requestBody string) (*http.Response, error) {
	return d.post(fmt.Sprintf("%s/v1/tenants/%s/devices/%s/scheduled-tasks", d.baseURL, tenantID, deviceID), []byte(requestBody))
}

func (d *APIDriver) UpdateScheduledTaskWithJSON(tenantID, deviceID, scheduledTaskID, requestBody string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(d.ctx, http.MethodPut, fmt.Sprintf("%s/v1/tenants/%s/devices/%s/scheduled-tasks/%s", d.baseURL, tenantID, deviceID, scheduledTaskID), strings.NewReader(requestBody))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	return d.client.Do(req)
}

func (d *APIDriver) GetTasksByScheduledTask(tenantID, deviceID, scheduledTaskID string, page, limit int) (*http.Response, error) {
	url := fmt.Sprintf("%s/v1/tenants/%s/devices/%s/scheduled-tasks/%s/tasks", d.baseURL, tenantID, deviceID, scheduledTaskID)
	if page > 0 || limit > 0 {
		url += fmt.Sprintf("?page=%d&limit=%d", page, limit)
	}
	return d.get(url)
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
	return d.post(fmt.Sprintf("%s/v1/tenants/%s/devices/%s/scheduled-tasks/%s/tasks", d.baseURL, tenantID, deviceID, scheduledTaskID), reqBody)
}
