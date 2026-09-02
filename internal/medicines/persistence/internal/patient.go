// Package internal holds the database rows of the medicines module. It is an
// internal package so the ORM shape never leaks past the module boundary.
package internal

import (
	"zensor-server/internal/infra/utils"

	medicinesDomain "zensor-server/internal/medicines/domain"
	shareddomain "zensor-server/internal/shared_kernel/domain"
)

type Patient struct {
	ID        string      `json:"id" gorm:"primaryKey"`
	Version   int         `json:"version"`
	TenantID  string      `json:"tenant_id" gorm:"index;not null"`
	Name      string      `json:"name" gorm:"not null"`
	Kind      string      `json:"kind" gorm:"index;not null"`
	Notes     string      `json:"notes"`
	CreatedAt utils.Time  `json:"created_at"`
	UpdatedAt utils.Time  `json:"updated_at"`
	DeletedAt *utils.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func (Patient) TableName() string {
	return "medicine_patients"
}

func (m Patient) ToDomain() medicinesDomain.Patient {
	return medicinesDomain.Patient{
		ID:        shareddomain.ID(m.ID),
		Version:   shareddomain.Version(m.Version),
		TenantID:  shareddomain.ID(m.TenantID),
		Name:      shareddomain.Name(m.Name),
		Kind:      medicinesDomain.PatientKind(m.Kind),
		Notes:     shareddomain.Description(m.Notes),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: m.DeletedAt,
	}
}

func FromPatient(value medicinesDomain.Patient) Patient {
	return Patient{
		ID:        value.ID.String(),
		Version:   int(value.Version),
		TenantID:  value.TenantID.String(),
		Name:      string(value.Name),
		Kind:      string(value.Kind),
		Notes:     string(value.Notes),
		CreatedAt: value.CreatedAt,
		UpdatedAt: value.UpdatedAt,
		DeletedAt: value.DeletedAt,
	}
}
