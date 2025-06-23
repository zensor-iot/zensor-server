package steps

import (
	"fmt"
	"net/http"
)

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
	fc.require.Equal(http.StatusCreated, resp.StatusCode)

	var data map[string]interface{}
	err = fc.decodeBody(resp.Body, &data)
	fc.require.NoError(err)
	fc.deviceID = data["id"].(string)
	return nil
}

func (fc *FeatureContext) theResponseShouldContainTheDeviceDetails() error {
	var data map[string]interface{}
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
	var devicesList map[string][]map[string]interface{}
	err := fc.decodeBody(fc.response.Body, &devicesList)
	fc.require.NoError(err)

	found := false
	for _, device := range devicesList["devices"] {
		if device["name"] == name {
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

func (fc *FeatureContext) theResponseShouldContainTheDeviceWithDisplayName(displayName string) error {
	var data map[string]interface{}
	err := fc.decodeBody(fc.response.Body, &data)
	fc.require.NoError(err)
	fc.require.Equal(displayName, data["display_name"])
	return nil
}
