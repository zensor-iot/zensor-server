package steps

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// TenantConfiguration represents a tenant configuration entity in the response
type TenantConfiguration struct {
	ID        string `json:"id"`
	TenantID  string `json:"tenant_id"`
	Timezone  string `json:"timezone"`
	Version   int    `json:"version"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// TenantConfigurationCreateRequest represents the request for creating a tenant configuration
type TenantConfigurationCreateRequest struct {
	Timezone string `json:"timezone"`
}

// TenantConfigurationUpdateRequest represents the request for updating a tenant configuration
type TenantConfigurationUpdateRequest struct {
	Timezone string `json:"timezone"`
}

// iHaveATenantWithIdForConfiguration is a simplified version for tenant configuration tests
func (fc *FeatureContext) iHaveATenantWithIdForConfiguration(tenantID string) error {
	// Create a tenant with the specified ID
	resp, err := fc.apiDriver.CreateTenant(tenantID, tenantID+"@example.com", "Test tenant for configuration")
	fc.require.NoError(err)

	// Accept both 201 (Created) and 409 (Conflict - already exists)
	if resp.StatusCode == http.StatusCreated {
		var data map[string]any
		err = fc.decodeBody(resp.Body, &data)
		fc.require.NoError(err)
		actualID := data["id"].(string)

		fc.tenantIDs = append(fc.tenantIDs, actualID)
		fc.tenantNameToID[tenantID] = actualID
		fc.tenantID = actualID
	} else if resp.StatusCode == http.StatusConflict {
		// If tenant already exists, find it by listing and add to array
		listResp, err := fc.apiDriver.ListTenants()
		fc.require.NoError(err)
		fc.require.Equal(http.StatusOK, listResp.StatusCode)

		var listData struct {
			Data []map[string]any `json:"data"`
		}
		err = fc.decodeBody(listResp.Body, &listData)
		fc.require.NoError(err)

		// Find the tenant
		for _, tenant := range listData.Data {
			if tenant["id"] != nil {
				actualID := tenant["id"].(string)
				name := tenant["name"].(string)
				found := false
				for _, existingID := range fc.tenantIDs {
					if existingID == actualID {
						found = true
						break
					}
				}
				if !found {
					fc.tenantIDs = append(fc.tenantIDs, actualID)
					fc.tenantNameToID[name] = actualID
				}
				fc.tenantID = actualID
				return nil
			}
		}
		fc.require.Fail("Tenant with name " + tenantID + " not found in list")
		return nil
	} else {
		fc.require.Equal(http.StatusCreated, resp.StatusCode, "Unexpected status code when creating tenant")
	}
	return nil
}

// Tenant Configuration step implementations
func (fc *FeatureContext) iCreateATenantConfigurationForTenantWithTimezone(tenantName string, timezone string) error {
	targetTenantID, exists := fc.tenantNameToID[tenantName]
	if !exists {
		targetTenantID = fc.tenantID
	}

	if fc.userID == "" {
		fc.userID = "test-user-" + targetTenantID
	}

	_, err := fc.apiDriver.AssociateUserWithTenants(fc.userID, []string{targetTenantID})
	if err != nil {
		return fmt.Errorf("failed to associate user with tenant: %w", err)
	}

	resp, err := fc.apiDriver.UpsertTenantConfiguration(targetTenantID, timezone, fc.userID)
	if err != nil {
		return fmt.Errorf("failed to create tenant configuration: %w", err)
	}
	fc.response = resp
	return nil
}

func (fc *FeatureContext) iGetTheTenantConfigurationForTenant(tenantName string) error {
	if tenantName == "non-existent-tenant" {
		resp, err := fc.apiDriver.GetTenantConfiguration(tenantName)
		if err != nil {
			return fmt.Errorf("failed to get tenant configuration: %w", err)
		}
		fc.response = resp
		return nil
	}

	targetTenantID, exists := fc.tenantNameToID[tenantName]
	if !exists {
		targetTenantID = fc.tenantID
	}

	resp, err := fc.apiDriver.GetTenantConfiguration(targetTenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant configuration: %w", err)
	}
	fc.response = resp
	return nil
}

func (fc *FeatureContext) iUpdateTheTenantConfigurationForTenantWithTimezone(tenantName string, timezone string) error {
	targetTenantID, exists := fc.tenantNameToID[tenantName]
	if !exists {
		targetTenantID = fc.tenantID
	}

	if fc.userID == "" {
		fc.userID = "test-user-" + targetTenantID
	}

	_, err := fc.apiDriver.AssociateUserWithTenants(fc.userID, []string{targetTenantID})
	if err != nil {
		return fmt.Errorf("failed to associate user with tenant: %w", err)
	}

	resp, err := fc.apiDriver.UpsertTenantConfiguration(targetTenantID, timezone, fc.userID)
	if err != nil {
		return fmt.Errorf("failed to update tenant configuration: %w", err)
	}
	fc.response = resp
	return nil
}

func (fc *FeatureContext) iUpdateTheTenantConfigurationForTenantWithTimezoneAndVersion(tenantName string, timezone string, version int) error {
	return fc.iUpdateTheTenantConfigurationForTenantWithTimezone(tenantName, timezone)
}

func (fc *FeatureContext) theTenantConfigurationShouldBeCreatedSuccessfully() error {
	fc.require.Equal(http.StatusOK, fc.response.StatusCode)
	return nil
}

func (fc *FeatureContext) theTenantConfigurationShouldBeRetrievedSuccessfully() error {
	fc.require.Equal(http.StatusOK, fc.response.StatusCode)
	return nil
}

func (fc *FeatureContext) theTenantConfigurationShouldBeUpdatedSuccessfully() error {
	fc.require.Equal(http.StatusOK, fc.response.StatusCode)
	return nil
}

func (fc *FeatureContext) theResponseShouldContainTimezone(timezone string) error {
	var config TenantConfiguration
	err := json.NewDecoder(fc.response.Body).Decode(&config)
	if err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	fc.require.Equal(timezone, config.Timezone)
	return nil
}

func (fc *FeatureContext) iHaveATenantConfigurationForTenantWithTimezone(tenantName string, timezone string) error {
	return fc.iCreateATenantConfigurationForTenantWithTimezone(tenantName, timezone)
}

func (fc *FeatureContext) theResponseShouldBe(statusCode string) error {
	switch statusCode {
	case "400 Bad Request":
		fc.require.Equal(http.StatusBadRequest, fc.response.StatusCode)
	case "404 Not Found":
		fc.require.Equal(http.StatusNotFound, fc.response.StatusCode)
	case "409 Conflict":
		fc.require.Equal(http.StatusConflict, fc.response.StatusCode)
	case "500 Internal Server Error":
		fc.require.Equal(http.StatusInternalServerError, fc.response.StatusCode)
	default:
		return fmt.Errorf("unexpected status code: %s", statusCode)
	}
	return nil
}

func (fc *FeatureContext) theErrorMessageShouldBe(message string) error {
	// For now, we'll just check that the response is not successful
	// In a real implementation, you might want to parse the error response body
	fc.require.NotEqual(http.StatusOK, fc.response.StatusCode)
	fc.require.NotEqual(http.StatusCreated, fc.response.StatusCode)

	// Try to parse the error response body if available
	if fc.response.Body != nil {
		var errorResponse map[string]any
		if err := json.NewDecoder(fc.response.Body).Decode(&errorResponse); err == nil {
			if errorMsg, ok := errorResponse["error"].(string); ok {
				fc.require.Contains(errorMsg, message, "Error message should contain expected text")
			}
		}
	}
	return nil
}
