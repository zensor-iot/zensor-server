package domain

import (
	shareddomain "zensor-server/internal/shared_kernel/domain"
)

type FieldDefinition struct {
	Name         shareddomain.Name
	DisplayName  shareddomain.DisplayName
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
