package httpserver

import (
	"net/http"
)

// Example usage of the pagination helper methods
func ExamplePaginationUsage() {
	// This is an example handler showing how to use the pagination helpers
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Extract pagination parameters from query string
		// e.g., ?page=2&limit=20
		params := ExtractPaginationParams(r)

		// Your business logic to fetch data with pagination
		// This would typically involve:
		// 1. Calculating offset: (params.Page - 1) * params.Limit
		// 2. Using params.Limit for the limit
		// 3. Getting total count for pagination metadata

		// Example data (replace with your actual data fetching logic)
		data := []string{"item1", "item2", "item3"}
		total := 100 // Total number of items in the database

		// Respond with paginated data
		ReplyWithPaginatedData(w, http.StatusOK, data, total, params)
	}

	_ = handler // Suppress unused variable warning
}

// Example of how to use pagination with database queries
func ExampleDatabasePagination() {
	// This is a conceptual example showing how pagination would work with database queries
	/*
		handler := func(w http.ResponseWriter, r *http.Request) {
			params := ExtractPaginationParams(r)

			// Calculate offset for database query
			offset := (params.Page - 1) * params.Limit

			// Example database query with pagination
			// var items []YourEntity
			// var total int64
			//
			// // Get total count
			// db.Model(&YourEntity{}).Count(&total)
			//
			// // Get paginated data
			// db.Offset(offset).Limit(params.Limit).Find(&items)

			// Convert to response format
			// responseData := convertToResponseFormat(items)

			// Reply with paginated response
			// ReplyWithPaginatedData(w, http.StatusOK, responseData, int(total), params)
		}
	*/
}
