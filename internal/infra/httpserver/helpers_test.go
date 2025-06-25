package httpserver

import (
	"net/http"
	"net/url"
	"testing"
)

func TestExtractPaginationParams(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected PaginationParams
	}{
		{
			name:     "default values when no query params",
			query:    "",
			expected: PaginationParams{Page: 1, Limit: 10},
		},
		{
			name:     "valid page and limit",
			query:    "page=2&limit=20",
			expected: PaginationParams{Page: 2, Limit: 20},
		},
		{
			name:     "invalid page defaults to 1",
			query:    "page=0&limit=10",
			expected: PaginationParams{Page: 1, Limit: 10},
		},
		{
			name:     "invalid limit defaults to 10",
			query:    "page=1&limit=0",
			expected: PaginationParams{Page: 1, Limit: 10},
		},
		{
			name:     "limit too high defaults to 10",
			query:    "page=1&limit=150",
			expected: PaginationParams{Page: 1, Limit: 10},
		},
		{
			name:     "only page parameter",
			query:    "page=3",
			expected: PaginationParams{Page: 3, Limit: 10},
		},
		{
			name:     "only limit parameter",
			query:    "limit=25",
			expected: PaginationParams{Page: 1, Limit: 25},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{
				URL: &url.URL{},
			}
			if tt.query != "" {
				req.URL.RawQuery = tt.query
			}

			result := ExtractPaginationParams(req)
			if result != tt.expected {
				t.Errorf("ExtractPaginationParams() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDefaultPaginationParams(t *testing.T) {
	params := DefaultPaginationParams()
	expected := PaginationParams{Page: 1, Limit: 10}

	if params != expected {
		t.Errorf("DefaultPaginationParams() = %v, want %v", params, expected)
	}
}
