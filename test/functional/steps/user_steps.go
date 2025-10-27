package steps

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type User struct {
	ID      string   `json:"id"`
	Tenants []string `json:"tenants"`
}

func (fc *FeatureContext) iAssociateUserWithTenants(userID string) error {
	tenantIDs := []string{fc.tenantID}
	if len(fc.tenantIDs) > 1 {
		tenantIDs = fc.tenantIDs[:2]
	}

	resp, err := fc.apiDriver.AssociateUserWithTenants(userID, tenantIDs)
	fc.require.NoError(err)
	fc.response = resp
	fc.userID = userID
	return nil
}

func (fc *FeatureContext) userIsAssociatedWithTenants(userID string) error {
	tenantIDs := fc.tenantIDs
	if len(tenantIDs) == 0 {
		tenantIDs = []string{fc.tenantID}
	}

	resp, err := fc.apiDriver.AssociateUserWithTenants(userID, tenantIDs)
	fc.require.NoError(err)
	fc.require.Equal(http.StatusOK, resp.StatusCode)

	fc.userID = userID
	return nil
}

func (fc *FeatureContext) iGetTheUser(userID string) error {
	resp, err := fc.apiDriver.GetUser(userID)
	fc.require.NoError(err)
	fc.response = resp
	return nil
}

func (fc *FeatureContext) theResponseShouldContainTheUserWithId(userID string) error {
	bodyBytes, err := io.ReadAll(fc.response.Body)
	fc.require.NoError(err)

	var data User
	err = json.Unmarshal(bodyBytes, &data)
	fc.require.NoError(err)
	fc.require.Equal(userID, data.ID)

	fc.responseData = map[string]any{"id": data.ID, "tenants": data.Tenants}
	return nil
}

func (fc *FeatureContext) theResponseShouldContainExactlyTenants(count int) error {
	if fc.responseData != nil {
		data := fc.responseData
		if tenantsInterface, ok := data["tenants"]; ok {
			if tenantsArray, ok := tenantsInterface.([]any); ok {
				var tenants []string
				for _, t := range tenantsArray {
					if tenant, ok := t.(string); ok {
						tenants = append(tenants, tenant)
					}
				}
				fc.require.Len(tenants, count)
				return nil
			}
			if tenants, ok := tenantsInterface.([]string); ok {
				fc.require.Len(tenants, count)
				return nil
			}
		}
	}

	bodyBytes, err := io.ReadAll(fc.response.Body)
	fc.require.NoError(err)

	var data User
	err = json.Unmarshal(bodyBytes, &data)
	fc.require.NoError(err)
	fc.require.Len(data.Tenants, count)
	return nil
}

func (fc *FeatureContext) iUpdateUserWithDifferentTenants(userID string) error {
	tenantIDs := fc.tenantIDs
	if len(tenantIDs) < 3 {
		tenantIDs = fc.tenantIDs
	}

	resp, err := fc.apiDriver.AssociateUserWithTenants(userID, tenantIDs)
	fc.require.NoError(err)
	fc.response = resp
	return err
}

func (fc *FeatureContext) userIsAssociatedWithTenantsCount(userID string, count string) error {
	tenantIDs := fc.tenantIDs
	resp, err := fc.apiDriver.AssociateUserWithTenants(userID, tenantIDs)
	fc.require.NoError(err)
	fc.require.Equal(http.StatusOK, resp.StatusCode)

	time.Sleep(50 * time.Millisecond)
	fc.userID = userID
	return nil
}

func (fc *FeatureContext) iAssociateUserWithEmptyTenantList(userID string) error {
	resp, err := fc.apiDriver.AssociateUserWithTenants(userID, []string{})
	fc.require.NoError(err)
	fc.response = resp
	return nil
}

func (fc *FeatureContext) iAttemptToAssociateUserWithNonExistentTenant(userID string) error {
	resp, err := fc.apiDriver.AssociateUserWithTenants(userID, []string{"non-existent-tenant-id"})
	fc.require.NoError(err)
	fc.response = resp
	return err
}

func (fc *FeatureContext) iAttemptToAssociateUserWithMixedTenantList(userID string) error {
	tenantIDs := []string{fc.tenantID, "invalid-tenant-id"}
	resp, err := fc.apiDriver.AssociateUserWithTenants(userID, tenantIDs)
	fc.require.NoError(err)
	fc.response = resp
	return err
}

func (fc *FeatureContext) anotherTenantExistsWithNameAndEmail(name, email string) error {
	resp, err := fc.apiDriver.CreateTenant(name, email, "Another test tenant")
	fc.require.NoError(err)
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	fc.require.NoError(err)

	var data map[string]any
	err = json.Unmarshal(bodyBytes, &data)
	fc.require.NoError(err)

	tenantID := data["id"].(string)
	fc.tenantIDs = append(fc.tenantIDs, tenantID)
	fc.tenantID = tenantID

	time.Sleep(50 * time.Millisecond)
	return nil
}

func (fc *FeatureContext) aThirdTenantExistsWithNameAndEmail(name, email string) error {
	return fc.anotherTenantExistsWithNameAndEmail(name, email)
}
