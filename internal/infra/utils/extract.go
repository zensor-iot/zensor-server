package utils

import (
	"fmt"

	"github.com/thoas/go-funk"
)

// ExtractStringValue uses go-funk to extract a string value from a message using the specified property name
func ExtractStringValue(msg any, propertyName string) string {
	if propertyName == "" {
		return ""
	}

	value := funk.Get(msg, propertyName)
	if value != nil {
		if strVal, ok := value.(string); ok {
			return strVal
		}
		// Convert other types to string
		return fmt.Sprintf("%v", value)
	}

	return ""
}

// ExtractFloat64Value uses go-funk to extract a float64 value from a message using the specified property name
func ExtractFloat64Value(msg any, propertyName string) float64 {
	if propertyName == "" {
		return 0.0
	}

	value := funk.Get(msg, propertyName)
	if value != nil {
		if floatVal, ok := value.(float64); ok {
			return floatVal
		}
		if intVal, ok := value.(int); ok {
			return float64(intVal)
		}
		if int32Val, ok := value.(int32); ok {
			return float64(int32Val)
		}
		if int64Val, ok := value.(int64); ok {
			return float64(int64Val)
		}
		if float32Val, ok := value.(float32); ok {
			return float64(float32Val)
		}
	}

	return 0.0
}
