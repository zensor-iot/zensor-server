package domain

import (
	"time"
	"zensor-server/internal/infra/utils"
	shareddomain "zensor-server/internal/shared_kernel/domain"
)

type ActivityType struct {
	ID           shareddomain.ID
	Name         shareddomain.Name
	DisplayName  shareddomain.DisplayName
	Description  shareddomain.Description
	IsPredefined bool
	Fields       []FieldDefinition
	CreatedAt    time.Time
}

func NewActivityTypeBuilder() *activityTypeBuilder {
	return &activityTypeBuilder{}
}

type activityTypeBuilder struct {
	actions []activityTypeHandler
}

type activityTypeHandler func(v *ActivityType) error

func (b *activityTypeBuilder) WithName(value string) *activityTypeBuilder {
	b.actions = append(b.actions, func(d *ActivityType) error {
		d.Name = shareddomain.Name(value)
		return nil
	})
	return b
}

func (b *activityTypeBuilder) WithDisplayName(value string) *activityTypeBuilder {
	b.actions = append(b.actions, func(d *ActivityType) error {
		d.DisplayName = shareddomain.DisplayName(value)
		return nil
	})
	return b
}

func (b *activityTypeBuilder) WithDescription(value string) *activityTypeBuilder {
	b.actions = append(b.actions, func(d *ActivityType) error {
		d.Description = shareddomain.Description(value)
		return nil
	})
	return b
}

func (b *activityTypeBuilder) WithIsPredefined(value bool) *activityTypeBuilder {
	b.actions = append(b.actions, func(d *ActivityType) error {
		d.IsPredefined = value
		return nil
	})
	return b
}

func (b *activityTypeBuilder) WithFields(value []FieldDefinition) *activityTypeBuilder {
	b.actions = append(b.actions, func(d *ActivityType) error {
		d.Fields = value
		return nil
	})
	return b
}

func (b *activityTypeBuilder) Build() (ActivityType, error) {
	result := ActivityType{
		ID:        shareddomain.ID(utils.GenerateUUID()),
		Fields:    make([]FieldDefinition, 0),
		CreatedAt: time.Now(),
	}

	for _, a := range b.actions {
		if err := a(&result); err != nil {
			return ActivityType{}, err
		}
	}

	return result, nil
}

const (
	ActivityTypeWaterSystem = "water_system"
	ActivityTypeCar         = "car"
	ActivityTypePets        = "pets"
	ActivityTypeCustom      = "custom"
)

var PredefinedActivityTypes = map[string]ActivityType{
	ActivityTypeWaterSystem: {
		Name:         shareddomain.Name(ActivityTypeWaterSystem),
		DisplayName:  shareddomain.DisplayName("Water System Maintenance"),
		Description:  shareddomain.Description("Maintenance tasks for water filtration and treatment systems"),
		IsPredefined: true,
		Fields: []FieldDefinition{
			{Name: shareddomain.Name("service_type"), DisplayName: shareddomain.DisplayName("Service Type"), Type: FieldTypeText, IsRequired: true},
			{Name: shareddomain.Name("contact"), DisplayName: shareddomain.DisplayName("Contact"), Type: FieldTypeText, IsRequired: false},
			{Name: shareddomain.Name("service_provider"), DisplayName: shareddomain.DisplayName("Service Provider"), Type: FieldTypeText, IsRequired: false},
			{Name: shareddomain.Name("cost"), DisplayName: shareddomain.DisplayName("Cost"), Type: FieldTypeNumber, IsRequired: false},
		},
	},
	ActivityTypeCar: {
		Name:         shareddomain.Name(ActivityTypeCar),
		DisplayName:  shareddomain.DisplayName("Car Maintenance"),
		Description:  shareddomain.Description("Automotive maintenance and service tracking"),
		IsPredefined: true,
		Fields: []FieldDefinition{
			{Name: shareddomain.Name("service_type"), DisplayName: shareddomain.DisplayName("Service Type"), Type: FieldTypeText, IsRequired: true},
			{Name: shareddomain.Name("mileage"), DisplayName: shareddomain.DisplayName("Mileage at Last Service"), Type: FieldTypeNumber, IsRequired: false},
			{Name: shareddomain.Name("service_provider"), DisplayName: shareddomain.DisplayName("Service Provider"), Type: FieldTypeText, IsRequired: false},
			{Name: shareddomain.Name("cost"), DisplayName: shareddomain.DisplayName("Cost"), Type: FieldTypeNumber, IsRequired: false},
		},
	},
	ActivityTypePets: {
		Name:         shareddomain.Name(ActivityTypePets),
		DisplayName:  shareddomain.DisplayName("Pet Care"),
		Description:  shareddomain.Description("Pet health care and grooming tracking"),
		IsPredefined: true,
		Fields: []FieldDefinition{
			{Name: shareddomain.Name("service_type"), DisplayName: shareddomain.DisplayName("Service Type"), Type: FieldTypeText, IsRequired: true},
			{Name: shareddomain.Name("provider_name"), DisplayName: shareddomain.DisplayName("Veterinary/Groomer Name"), Type: FieldTypeText, IsRequired: false},
			{Name: shareddomain.Name("pet_name"), DisplayName: shareddomain.DisplayName("Pet Name"), Type: FieldTypeText, IsRequired: false},
			{Name: shareddomain.Name("medicine"), DisplayName: shareddomain.DisplayName("Medicine"), Type: FieldTypeText, IsRequired: false},
			{Name: shareddomain.Name("cost"), DisplayName: shareddomain.DisplayName("Cost"), Type: FieldTypeNumber, IsRequired: false},
		},
	},
}
