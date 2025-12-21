package steps

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Device represents a device entity in the response
type Device struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	DisplayName   string `json:"display_name"`
	TenantID      string `json:"tenant_id"`
	LastMessageAt string `json:"last_message_at,omitempty"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// Device step implementations
func (fc *FeatureContext) iCreateANewDeviceWithNameAndDisplayName(name, displayName string) error {
	resp, err := fc.apiDriver.CreateDevice(name, displayName)
	fc.require.NoError(err)
	fc.response = resp
	return nil
}

func (fc *FeatureContext) aDeviceExistsWithName(name string) error {
	resp, err := fc.apiDriver.CreateDevice(name, name)
	fc.require.NoError(err)

	if resp.StatusCode == http.StatusConflict {
		// Device already exists, try to get it by listing all devices
		listResp, err := fc.apiDriver.ListDevices()
		fc.require.NoError(err)
		fc.require.Equal(http.StatusOK, listResp.StatusCode)

		var listData struct {
			Data []map[string]any `json:"data"`
		}
		err = fc.decodeBody(listResp.Body, &listData)
		fc.require.NoError(err)

		for _, device := range listData.Data {
			if device["name"] == name {
				fc.deviceID = device["id"].(string)
				return nil
			}
		}
		fc.require.Fail("Device with name " + name + " not found in list")
		return nil
	}

	fc.require.Equal(http.StatusCreated, resp.StatusCode)

	var data map[string]any
	err = fc.decodeBody(resp.Body, &data)
	fc.require.NoError(err)
	fc.deviceID = data["id"].(string)
	return nil
}

func (fc *FeatureContext) theResponseShouldContainTheDeviceDetails() error {
	var data map[string]any
	err := fc.decodeBody(fc.response.Body, &data)
	fc.require.NoError(err)
	fc.require.NotEmpty(data["id"])
	fc.deviceID = data["id"].(string)
	fc.responseData = data
	return nil
}

func (fc *FeatureContext) iListAllDevices() error {
	resp, err := fc.apiDriver.ListDevices()
	fc.require.NoError(err)
	fc.response = resp
	return err
}

func (fc *FeatureContext) theListShouldContainTheDeviceWithName(name string) error {
	body, err := io.ReadAll(fc.response.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var paginatedResp PaginatedResponse[Device]
	if err := json.Unmarshal(body, &paginatedResp); err != nil {
		return fmt.Errorf("failed to decode paginated response: %w", err)
	}

	found := false
	for _, device := range paginatedResp.Data {
		if device.Name == name {
			found = true
			break
		}
	}
	fc.require.True(found, fmt.Sprintf("Device with name %s not found in list", name))
	return nil
}

func (fc *FeatureContext) iUpdateTheDeviceWithANewDisplayName(newDisplayName string) error {
	resp, err := fc.apiDriver.UpdateDevice(fc.deviceID, newDisplayName)
	fc.require.NoError(err)
	fc.response = resp
	return err
}

func (fc *FeatureContext) iGetTheDeviceByItsID() error {
	resp, err := fc.apiDriver.GetDevice(fc.deviceID)
	fc.require.NoError(err)
	fc.response = resp
	return err
}

func (fc *FeatureContext) theResponseShouldContainTheDeviceWithDisplayName(displayName string) error {
	var data map[string]any
	err := fc.decodeBody(fc.response.Body, &data)
	fc.require.NoError(err)
	fc.require.Equal(displayName, data["display_name"])
	return nil
}
