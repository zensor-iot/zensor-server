package domain

import (
	"time"
	"zensor-server/internal/infra/utils"
	sharedDomain "zensor-server/internal/shared_kernel/domain"
)

type ActivityType struct {
	ID           sharedDomain.ID
	Name         sharedDomain.Name
	DisplayName  sharedDomain.DisplayName
	Description  sharedDomain.Description
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
		d.Name = sharedDomain.Name(value)
		return nil
	})
	return b
}

func (b *activityTypeBuilder) WithDisplayName(value string) *activityTypeBuilder {
	b.actions = append(b.actions, func(d *ActivityType) error {
		d.DisplayName = sharedDomain.DisplayName(value)
		return nil
	})
	return b
}

func (b *activityTypeBuilder) WithDescription(value string) *activityTypeBuilder {
	b.actions = append(b.actions, func(d *ActivityType) error {
		d.Description = sharedDomain.Description(value)
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
		ID:        sharedDomain.ID(utils.GenerateUUID()),
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
		Name:         sharedDomain.Name(ActivityTypeWaterSystem),
		DisplayName:  sharedDomain.DisplayName("Water System Maintenance"),
		Description:  sharedDomain.Description("Maintenance tasks for water filtration and treatment systems"),
		IsPredefined: true,
		Fields: []FieldDefinition{
			{Name: sharedDomain.Name("service_type"), DisplayName: sharedDomain.DisplayName("Service Type"), Type: FieldTypeText, IsRequired: true},
			{Name: sharedDomain.Name("contact"), DisplayName: sharedDomain.DisplayName("Contact"), Type: FieldTypeText, IsRequired: false},
			{Name: sharedDomain.Name("service_provider"), DisplayName: sharedDomain.DisplayName("Service Provider"), Type: FieldTypeText, IsRequired: false},
			{Name: sharedDomain.Name("cost"), DisplayName: sharedDomain.DisplayName("Cost"), Type: FieldTypeNumber, IsRequired: false},
		},
	},
	ActivityTypeCar: {
		Name:         sharedDomain.Name(ActivityTypeCar),
		DisplayName:  sharedDomain.DisplayName("Car Maintenance"),
		Description:  sharedDomain.Description("Automotive maintenance and service tracking"),
		IsPredefined: true,
		Fields: []FieldDefinition{
			{Name: sharedDomain.Name("service_type"), DisplayName: sharedDomain.DisplayName("Service Type"), Type: FieldTypeText, IsRequired: true},
			{Name: sharedDomain.Name("mileage"), DisplayName: sharedDomain.DisplayName("Mileage at Last Service"), Type: FieldTypeNumber, IsRequired: false},
			{Name: sharedDomain.Name("service_provider"), DisplayName: sharedDomain.DisplayName("Service Provider"), Type: FieldTypeText, IsRequired: false},
			{Name: sharedDomain.Name("cost"), DisplayName: sharedDomain.DisplayName("Cost"), Type: FieldTypeNumber, IsRequired: false},
		},
	},
	ActivityTypePets: {
		Name:         sharedDomain.Name(ActivityTypePets),
		DisplayName:  sharedDomain.DisplayName("Pet Care"),
		Description:  sharedDomain.Description("Pet health care and grooming tracking"),
		IsPredefined: true,
		Fields: []FieldDefinition{
			{Name: sharedDomain.Name("service_type"), DisplayName: sharedDomain.DisplayName("Service Type"), Type: FieldTypeText, IsRequired: true},
			{Name: sharedDomain.Name("provider_name"), DisplayName: sharedDomain.DisplayName("Veterinary/Groomer Name"), Type: FieldTypeText, IsRequired: false},
			{Name: sharedDomain.Name("pet_name"), DisplayName: sharedDomain.DisplayName("Pet Name"), Type: FieldTypeText, IsRequired: false},
			{Name: sharedDomain.Name("medicine"), DisplayName: sharedDomain.DisplayName("Medicine"), Type: FieldTypeText, IsRequired: false},
			{Name: sharedDomain.Name("cost"), DisplayName: sharedDomain.DisplayName("Cost"), Type: FieldTypeNumber, IsRequired: false},
		},
	},
}
