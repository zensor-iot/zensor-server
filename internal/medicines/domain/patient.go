package domain

import (
	"time"
	"zensor-server/internal/infra/utils"

	shareddomain "zensor-server/internal/shared_kernel/domain"
)

// Patient is a family member that receives medication, human or animal.
type Patient struct {
	ID        shareddomain.ID          `json:"id"`
	Version   shareddomain.Version     `json:"version"`
	TenantID  shareddomain.ID          `json:"tenant_id"`
	Name      shareddomain.Name        `json:"name"`
	Kind      PatientKind              `json:"kind"`
	Notes     shareddomain.Description `json:"notes"`
	CreatedAt utils.Time               `json:"created_at"`
	UpdatedAt utils.Time               `json:"updated_at"`
	DeletedAt *utils.Time              `json:"deleted_at,omitempty"`
}

func (p *Patient) IsDeleted() bool {
	return p.DeletedAt != nil
}

func (p *Patient) SoftDelete() {
	now := utils.Time{Time: time.Now()}
	p.DeletedAt = &now
	p.UpdatedAt = now
}

type patientBuilder struct {
	actions []func(*Patient) error
}

func NewPatientBuilder() *patientBuilder {
	return &patientBuilder{}
}

func (b *patientBuilder) WithTenantID(value string) *patientBuilder {
	b.actions = append(b.actions, func(p *Patient) error {
		p.TenantID = shareddomain.ID(value)
		return nil
	})
	return b
}

func (b *patientBuilder) WithName(value string) *patientBuilder {
	b.actions = append(b.actions, func(p *Patient) error {
		p.Name = shareddomain.Name(value)
		return nil
	})
	return b
}

func (b *patientBuilder) WithKind(value PatientKind) *patientBuilder {
	b.actions = append(b.actions, func(p *Patient) error {
		p.Kind = value
		return nil
	})
	return b
}

func (b *patientBuilder) WithNotes(value string) *patientBuilder {
	b.actions = append(b.actions, func(p *Patient) error {
		p.Notes = shareddomain.Description(value)
		return nil
	})
	return b
}

func (b *patientBuilder) Build() (Patient, error) {
	now := utils.Time{Time: time.Now()}
	result := Patient{
		ID:        shareddomain.ID(utils.GenerateUUID()),
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	for _, action := range b.actions {
		if err := action(&result); err != nil {
			return Patient{}, err
		}
	}

	if result.TenantID == "" {
		return Patient{}, ErrTenantIDRequired
	}
	if result.Name == "" {
		return Patient{}, ErrPatientNameRequired
	}
	if err := result.Kind.Validate(); err != nil {
		return Patient{}, err
	}

	return result, nil
}
