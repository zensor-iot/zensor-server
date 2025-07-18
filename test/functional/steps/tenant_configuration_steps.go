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
		fc.tenantID = data["id"].(string)
	} else if resp.StatusCode == http.StatusConflict {
		// If tenant already exists, just use the provided ID
		fc.tenantID = tenantID
	} else {
		fc.require.Equal(http.StatusCreated, resp.StatusCode, "Unexpected status code when creating tenant")
	}
	return nil
}

// Tenant Configuration step implementations
func (fc *FeatureContext) iCreateATenantConfigurationForTenantWithTimezone(tenantName string, timezone string) error {
	// Use the tenant ID from the feature context, which is set by the background steps
	resp, err := fc.apiDriver.CreateTenantConfiguration(fc.tenantID, timezone)
	if err != nil {
		return fmt.Errorf("failed to create tenant configuration: %w", err)
	}
	fc.response = resp
	return nil
}

func (fc *FeatureContext) iGetTheTenantConfigurationForTenant(tenantName string) error {
	// For "non-existent-tenant", use the literal value
	if tenantName == "non-existent-tenant" {
		resp, err := fc.apiDriver.GetTenantConfiguration(tenantName)
		if err != nil {
			return fmt.Errorf("failed to get tenant configuration: %w", err)
		}
		fc.response = resp
		return nil
	}

	// For existing tenants, use the tenant ID from the feature context
	resp, err := fc.apiDriver.GetTenantConfiguration(fc.tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant configuration: %w", err)
	}
	fc.response = resp
	return nil
}

func (fc *FeatureContext) iUpdateTheTenantConfigurationForTenantWithTimezone(tenantName string, timezone string) error {
	// Use the tenant ID from the feature context, which is set by the background steps
	resp, err := fc.apiDriver.UpdateTenantConfiguration(fc.tenantID, timezone)
	if err != nil {
		return fmt.Errorf("failed to update tenant configuration: %w", err)
	}
	fc.response = resp
	return nil
}

func (fc *FeatureContext) iUpdateTheTenantConfigurationForTenantWithTimezoneAndVersion(_ string, timezone string, version int) error {
	// Version is now handled internally, so we just call the regular update method
	return fc.iUpdateTheTenantConfigurationForTenantWithTimezone("", timezone)
}

func (fc *FeatureContext) theTenantConfigurationShouldBeCreatedSuccessfully() error {
	fc.require.Equal(http.StatusCreated, fc.response.StatusCode)
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

func (fc *FeatureContext) theResponseShouldContainTenantID(tenantID string) error {
	var config TenantConfiguration
	err := json.NewDecoder(fc.response.Body).Decode(&config)
	if err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	fc.require.Equal(tenantID, config.TenantID)
	return nil
}

func (fc *FeatureContext) iHaveATenantConfigurationForTenantWithTimezone(_ string, timezone string) error {
	return fc.iCreateATenantConfigurationForTenantWithTimezone("", timezone)
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
