package driver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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

func (d *APIDriver) CreateMaintenanceActivity(tenantID, typeName, name, description, schedule string, notificationDaysBefore []int, fields []map[string]any) (*http.Response, error) {
	reqBody := map[string]any{
		"tenant_id":                tenantID,
		"type_name":                typeName,
		"name":                     name,
		"description":              description,
		"schedule":                 schedule,
		"notification_days_before": notificationDaysBefore,
		"fields":                   fields,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		panic(err)
	}
	return d.client.Post(fmt.Sprintf("%s/v1/maintenance/activities", d.baseURL), "application/json", bytes.NewBuffer(body))
}

func (d *APIDriver) ListMaintenanceActivities(tenantID string, page, limit int) (*http.Response, error) {
	url := fmt.Sprintf("%s/v1/maintenance/activities?tenant_id=%s", d.baseURL, tenantID)
	if page > 0 || limit > 0 {
		url += fmt.Sprintf("&page=%d&limit=%d", page, limit)
	}
	return d.client.Get(url)
}

func (d *APIDriver) GetMaintenanceActivity(id string) (*http.Response, error) {
	return d.client.Get(fmt.Sprintf("%s/v1/maintenance/activities/%s", d.baseURL, id))
}

func (d *APIDriver) UpdateMaintenanceActivity(id string, updates map[string]any) (*http.Response, error) {
	reqBody, err := json.Marshal(updates)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/v1/maintenance/activities/%s", d.baseURL, id), bytes.NewBuffer(reqBody))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	return d.client.Do(req)
}

func (d *APIDriver) DeleteMaintenanceActivity(id string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/v1/maintenance/activities/%s", d.baseURL, id), nil)
	if err != nil {
		panic(err)
	}
	return d.client.Do(req)
}

func (d *APIDriver) ActivateMaintenanceActivity(id string) (*http.Response, error) {
	return d.client.Post(fmt.Sprintf("%s/v1/maintenance/activities/%s/activate", d.baseURL, id), "application/json", nil)
}

func (d *APIDriver) DeactivateMaintenanceActivity(id string) (*http.Response, error) {
	return d.client.Post(fmt.Sprintf("%s/v1/maintenance/activities/%s/deactivate", d.baseURL, id), "application/json", nil)
}

func (d *APIDriver) ListMaintenanceExecutions(activityID string, page, limit int) (*http.Response, error) {
	url := fmt.Sprintf("%s/v1/maintenance/executions?activity_id=%s", d.baseURL, activityID)
	if page > 0 || limit > 0 {
		url += fmt.Sprintf("&page=%d&limit=%d", page, limit)
	}
	return d.client.Get(url)
}

func (d *APIDriver) GetMaintenanceExecution(id string) (*http.Response, error) {
	return d.client.Get(fmt.Sprintf("%s/v1/maintenance/executions/%s", d.baseURL, id))
}

func (d *APIDriver) MarkMaintenanceExecutionCompleted(id, completedBy string) (*http.Response, error) {
	reqBody, err := json.Marshal(map[string]any{
		"completed_by": completedBy,
	})
	if err != nil {
		panic(err)
	}
	return d.client.Post(fmt.Sprintf("%s/v1/maintenance/executions/%s/complete", d.baseURL, id), "application/json", bytes.NewBuffer(reqBody))
}

func (d *APIDriver) CreateMaintenanceExecution(activityID string, scheduledDate time.Time, fieldValues map[string]any) (*http.Response, error) {
	reqBody := map[string]any{
		"activity_id":    activityID,
		"scheduled_date": scheduledDate.Format(time.RFC3339),
		"field_values":   fieldValues,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		panic(err)
	}
	return d.client.Post(fmt.Sprintf("%s/v1/maintenance/executions", d.baseURL), "application/json", bytes.NewBuffer(body))
}
