package domain

import "slices"

type User struct {
	ID      ID
	Tenants []ID
}

func (u *User) AddTenant(tenantID ID) {
	u.Tenants = append(u.Tenants, tenantID)
}

func (u *User) RemoveTenant(tenantID ID) {
	u.Tenants = slices.DeleteFunc(u.Tenants, func(t ID) bool {
		return t == tenantID
	})
}

func (u *User) HasTenant(tenantID ID) bool {
	return slices.Contains(u.Tenants, tenantID)
}

func (u *User) SetTenants(tenantIDs []ID) {
	u.Tenants = make([]ID, len(tenantIDs))
	copy(u.Tenants, tenantIDs)
}
