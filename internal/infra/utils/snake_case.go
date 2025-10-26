package utils

import (
	"strings"
	"unicode"
)

// ToSnakeCase converts a camelCase or PascalCase string to snake_case.
func ToSnakeCase(s string) string {
	if s == "" {
		return s
	}

	var result strings.Builder
	result.Grow(len(s) + len(s)/2)

	runes := []rune(s)
	for i, r := range runes {
		if i == 0 {
			result.WriteRune(unicode.ToLower(r))
			continue
		}

		if unicode.IsUpper(r) {
			if unicode.IsLower(runes[i-1]) || unicode.IsDigit(runes[i-1]) {
				result.WriteRune('_')
			} else if unicode.IsUpper(runes[i-1]) {
				if i+1 == len(runes) || unicode.IsUpper(runes[i+1]) {
					result.WriteRune('_')
				} else if unicode.IsLower(runes[i+1]) {
					result.WriteRune('_')
				}
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}
