package utils

import (
	"strings"
	"unicode"
)

// ToSnakeCase converts a camelCase or PascalCase string to snake_case.
// It handles various edge cases including consecutive uppercase letters,
// numbers, and mixed case scenarios.
//
// Examples:
//   - "camelCase" -> "camel_case"
//   - "PascalCase" -> "pascal_case"
//   - "XMLHttpRequest" -> "xml_http_request"
//   - "version2Update" -> "version2_update"
//   - "snake_case" -> "snake_case" (unchanged)
func ToSnakeCase(s string) string {
	if s == "" {
		return s
	}

	var result strings.Builder
	result.Grow(len(s) + len(s)/2) // Pre-allocate capacity

	runes := []rune(s)
	for i, r := range runes {
		// Handle first character
		if i == 0 {
			result.WriteRune(unicode.ToLower(r))
			continue
		}

		// Check if current character is uppercase
		if unicode.IsUpper(r) {
			// Check if previous character is lowercase or a number
			// This handles cases like "camelCase" -> "camel_case"
			if unicode.IsLower(runes[i-1]) || unicode.IsDigit(runes[i-1]) {
				result.WriteRune('_')
			} else if i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
				// Check if next character is lowercase
				// This handles cases like "XMLHttpRequest" -> "xml_http_request"
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}
