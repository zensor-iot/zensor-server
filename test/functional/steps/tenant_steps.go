package steps

import (
	"fmt"
	"net/http"
	"time"
)

// Tenant represents a tenant entity in the response
type Tenant struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// Tenant step implementations
func (fc *FeatureContext) iCreateANewTenantWithNameAndEmail(name, email string) error {
	resp, err := fc.apiDriver.CreateTenant(name, email, "A test tenant")
	fc.require.NoError(err)
	fc.response = resp
	return nil
}

func (fc *FeatureContext) aTenantExistsWithNameAndEmail(name, email string) error {
	resp, err := fc.apiDriver.CreateTenant(name, email, "A test tenant")
	fc.require.NoError(err)
	fc.require.Equal(http.StatusCreated, resp.StatusCode)

	var data map[string]any
	err = fc.decodeBody(resp.Body, &data)
	fc.require.NoError(err)
	fc.tenantID = data["id"].(string)
	return nil
}

func (fc *FeatureContext) aDeactivatedTenantExistsWithNameAndEmail(name, email string) error {
	err := fc.aTenantExistsWithNameAndEmail(name, email)
	time.Sleep(50 * time.Millisecond) // wait for replication to complete
	fc.require.NoError(err)
	resp, err := fc.apiDriver.DeactivateTenant(fc.tenantID)
	fc.require.NoError(err)
	fc.require.Equal(http.StatusNoContent, resp.StatusCode)
	return nil
}

func (fc *FeatureContext) iGetTheTenantByItsID() error {
	resp, err := fc.apiDriver.GetTenant(fc.tenantID)
	fc.require.NoError(err)
	fc.response = resp
	return err
}

func (fc *FeatureContext) theResponseShouldContainTheTenantWithName(name string) error {
	var data map[string]any
	err := fc.decodeBody(fc.response.Body, &data)
	fc.require.NoError(err)
	fc.require.Equal(name, data["name"])
	return nil
}

func (fc *FeatureContext) iListAllTenants() error {
	resp, err := fc.apiDriver.ListTenants()
	fc.require.NoError(err)
	fc.response = resp
	return err
}

func (fc *FeatureContext) theListShouldContainTheTenantWithName(name string) error {
	items, err := fc.decodePaginatedResponse(fc.response)
	fc.require.NoError(err)

	found := false
	for _, item := range items {
		if item["name"] == name {
			found = true
			break
		}
	}
	fc.require.True(found, fmt.Sprintf("Tenant with name %s not found in list", name))
	return nil
}

func (fc *FeatureContext) iUpdateTheTenantWithANewName(newName string) error {
	resp, err := fc.apiDriver.UpdateTenant(fc.tenantID, newName)
	fc.require.NoError(err)
	fc.response = resp
	return err
}

func (fc *FeatureContext) iDeactivateTheTenant() error {
	resp, err := fc.apiDriver.DeactivateTenant(fc.tenantID)
	fc.require.NoError(err)
	fc.response = resp
	return err
}

func (fc *FeatureContext) iActivateTheTenant() error {
	resp, err := fc.apiDriver.ActivateTenant(fc.tenantID)
	fc.require.NoError(err)
	fc.response = resp
	return err
}

func (fc *FeatureContext) iSoftDeleteTheTenant() error {
	resp, err := fc.apiDriver.SoftDeleteTenant(fc.tenantID)
	fc.require.NoError(err)
	fc.response = resp
	return err
}

func (fc *FeatureContext) theTenantShouldBeSoftDeleted() error {
	var data map[string]any
	err := fc.decodeBody(fc.response.Body, &data)
	fc.require.NoError(err)
	fc.require.NotNil(data["deleted_at"], "Tenant should be soft deleted")
	return nil
}
