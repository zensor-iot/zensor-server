package steps

import (
	"fmt"
	"net/http"
)

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
	var tenantList map[string][]map[string]any
	err := fc.decodeBody(fc.response.Body, &tenantList)
	fc.require.NoError(err)

	found := false
	for _, tenant := range tenantList["tenants"] {
		if tenant["name"] == name {
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
