package driver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"time"
)

type APIDriver struct {
	baseURL string
	client  *http.Client
	ctx     context.Context
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

	return nil
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
	req, err := http.NewRequestWithContext(d.ctx, http.MethodPut, fmt.Sprintf("%s/v1/tenants/%s", d.baseURL, id), bytes.NewBuffer(reqBody))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	return d.client.Do(req)
}

func (d *APIDriver) DeactivateTenant(id string) (*http.Response, error) {
	return d.post(fmt.Sprintf("%s/v1/tenants/%s/deactivate", d.baseURL, id), nil)
}

func (d *APIDriver) ActivateTenant(id string) (*http.Response, error) {
	return d.post(fmt.Sprintf("%s/v1/tenants/%s/activate", d.baseURL, id), nil)
}

func (d *APIDriver) SoftDeleteTenant(id string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(d.ctx, http.MethodDelete, fmt.Sprintf("%s/v1/tenants/%s", d.baseURL, id), nil)
	if err != nil {
		panic(err)
	}
	return d.client.Do(req)
}

func (d *APIDriver) CreateMaintenanceActivity(tenantID, typeName, name, description string, schedule map[string]any, notificationDaysBefore []int, fields []map[string]any) (*http.Response, error) {
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
	return d.post(d.baseURL+"/v1/maintenance/activities", body)
}

func (d *APIDriver) ListMaintenanceActivities(tenantID string, page, limit int) (*http.Response, error) {
	url := fmt.Sprintf("%s/v1/maintenance/activities?tenant_id=%s", d.baseURL, tenantID)
	if page > 0 || limit > 0 {
		url += fmt.Sprintf("&page=%d&limit=%d", page, limit)
	}
	return d.get(url)
}

func (d *APIDriver) GetMaintenanceActivity(id string) (*http.Response, error) {
	return d.get(fmt.Sprintf("%s/v1/maintenance/activities/%s", d.baseURL, id))
}

func (d *APIDriver) UpdateMaintenanceActivity(id string, updates map[string]any) (*http.Response, error) {
	reqBody, err := json.Marshal(updates)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequestWithContext(d.ctx, http.MethodPut, fmt.Sprintf("%s/v1/maintenance/activities/%s", d.baseURL, id), bytes.NewBuffer(reqBody))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	return d.client.Do(req)
}

func (d *APIDriver) DeleteMaintenanceActivity(id string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(d.ctx, http.MethodDelete, fmt.Sprintf("%s/v1/maintenance/activities/%s", d.baseURL, id), nil)
	if err != nil {
		panic(err)
	}
	return d.client.Do(req)
}

func (d *APIDriver) ActivateMaintenanceActivity(id string) (*http.Response, error) {
	return d.post(fmt.Sprintf("%s/v1/maintenance/activities/%s/activate", d.baseURL, id), nil)
}

func (d *APIDriver) DeactivateMaintenanceActivity(id string) (*http.Response, error) {
	return d.post(fmt.Sprintf("%s/v1/maintenance/activities/%s/deactivate", d.baseURL, id), nil)
}

func (d *APIDriver) ListMaintenanceExecutions(activityID string, page, limit int) (*http.Response, error) {
	url := fmt.Sprintf("%s/v1/maintenance/executions?activity_id=%s", d.baseURL, activityID)
	if page > 0 || limit > 0 {
		url += fmt.Sprintf("&page=%d&limit=%d", page, limit)
	}
	return d.get(url)
}

func (d *APIDriver) GetMaintenanceExecution(id string) (*http.Response, error) {
	return d.get(fmt.Sprintf("%s/v1/maintenance/executions/%s", d.baseURL, id))
}

func (d *APIDriver) MarkMaintenanceExecutionCompleted(id, completedBy string) (*http.Response, error) {
	return d.MarkMaintenanceExecutionCompletedWithFieldValues(id, completedBy, nil)
}

func (d *APIDriver) MarkMaintenanceExecutionCompletedWithFieldValues(id, completedBy string, fieldValues map[string]any) (*http.Response, error) {
	body := map[string]any{
		"completed_by": completedBy,
	}
	if fieldValues != nil {
		body["field_values"] = fieldValues
	}
	reqBody, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	return d.post(fmt.Sprintf("%s/v1/maintenance/executions/%s/complete", d.baseURL, id), reqBody)
}

func (d *APIDriver) GetVAPIDPublicKey() (*http.Response, error) {
	return d.get(d.baseURL + "/v1/push/vapid-public-key")
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
	return d.post(d.baseURL+"/v1/maintenance/executions", body)
}
