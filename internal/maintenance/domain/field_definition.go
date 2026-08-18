package domain

import (
	shareddomain "zensor-server/internal/shared_kernel/domain"
)

type FieldDefinition struct {
	Name         shareddomain.Name        `json:"name"`
	DisplayName  shareddomain.DisplayName `json:"display_name"`
	Type         FieldType                `json:"type"`
	IsRequired   bool                     `json:"is_required"`
	DefaultValue *any                     `json:"default_value,omitempty"`
}

type FieldType string

const (
	FieldTypeText    FieldType = "text"
	FieldTypeNumber  FieldType = "number"
	FieldTypeDate    FieldType = "date"
	FieldTypeBoolean FieldType = "boolean"
)
