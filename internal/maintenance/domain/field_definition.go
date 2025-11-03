package domain

import (
	sharedDomain "zensor-server/internal/shared_kernel/domain"
)

type FieldDefinition struct {
	Name         sharedDomain.Name
	DisplayName  sharedDomain.DisplayName
	Type         FieldType
	IsRequired   bool
	DefaultValue *any
}

type FieldType string

const (
	FieldTypeText    FieldType = "text"
	FieldTypeNumber  FieldType = "number"
	FieldTypeDate    FieldType = "date"
	FieldTypeBoolean FieldType = "boolean"
)
