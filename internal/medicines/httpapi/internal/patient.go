// Package internal holds the wire shapes of the medicines HTTP API, kept
// unimportable from outside the module.
package internal

import (
	"time"

	medicinesDomain "zensor-server/internal/medicines/domain"
)

type PatientResponse struct {
	ID        string    `json:"id"`
	Version   int       `json:"version"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	Kind      string    `json:"kind"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PatientCreateRequest struct {
	TenantID string `json:"tenant_id"`
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Notes    string `json:"notes"`
}

type PatientUpdateRequest struct {
	Name  *string `json:"name,omitempty"`
	Kind  *string `json:"kind,omitempty"`
	Notes *string `json:"notes,omitempty"`
}

func ToPatientResponse(patient medicinesDomain.Patient) PatientResponse {
	return PatientResponse{
		ID:        patient.ID.String(),
		Version:   int(patient.Version),
		TenantID:  patient.TenantID.String(),
		Name:      string(patient.Name),
		Kind:      string(patient.Kind),
		Notes:     string(patient.Notes),
		CreatedAt: patient.CreatedAt.Time,
		UpdatedAt: patient.UpdatedAt.Time,
	}
}
