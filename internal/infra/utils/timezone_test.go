package utils

import (
	"testing"
)

func TestValidateTimezone(t *testing.T) {
	tests := []struct {
		name      string
		timezone  string
		expectErr bool
	}{
		{
			name:      "valid timezone - UTC",
			timezone:  "UTC",
			expectErr: false,
		},
		{
			name:      "valid timezone - America/New_York",
			timezone:  "America/New_York",
			expectErr: false,
		},
		{
			name:      "valid timezone - Europe/London",
			timezone:  "Europe/London",
			expectErr: false,
		},
		{
			name:      "valid timezone - Asia/Tokyo",
			timezone:  "Asia/Tokyo",
			expectErr: false,
		},
		{
			name:      "valid timezone - Australia/Sydney",
			timezone:  "Australia/Sydney",
			expectErr: false,
		},
		{
			name:      "valid timezone - EST",
			timezone:  "EST",
			expectErr: false,
		},
		{
			name:      "empty timezone",
			timezone:  "",
			expectErr: true,
		},
		{
			name:      "invalid timezone - PST",
			timezone:  "PST",
			expectErr: true,
		},
		{
			name:      "invalid timezone - random string",
			timezone:  "Invalid/Timezone/Name",
			expectErr: true,
		},
		{
			name:      "invalid timezone - with spaces",
			timezone:  "America/New York",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimezone(tt.timezone)
			if tt.expectErr && err == nil {
				t.Errorf("ValidateTimezone() expected error for timezone '%s', but got none", tt.timezone)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("ValidateTimezone() unexpected error for timezone '%s': %v", tt.timezone, err)
			}
		})
	}
}

func TestIsValidTimezone(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
		expected bool
	}{
		{
			name:     "valid timezone - UTC",
			timezone: "UTC",
			expected: true,
		},
		{
			name:     "valid timezone - America/New_York",
			timezone: "America/New_York",
			expected: true,
		},
		{
			name:     "valid timezone - EST",
			timezone: "EST",
			expected: true,
		},
		{
			name:     "empty timezone",
			timezone: "",
			expected: false,
		},
		{
			name:     "invalid timezone - PST",
			timezone: "PST",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidTimezone(tt.timezone)
			if result != tt.expected {
				t.Errorf("IsValidTimezone() = %v, expected %v for timezone '%s'", result, tt.expected, tt.timezone)
			}
		})
	}
}
