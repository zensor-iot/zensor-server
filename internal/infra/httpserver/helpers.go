package httpserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
)

type ErrorResponse struct {
	Message string `json:"message,omitempty"`
}

// PaginatedResponse represents a paginated response with data and pagination metadata
type PaginatedResponse struct {
	Data       any `json:"data"`
	Pagination struct {
		Page       int `json:"page"`
		Limit      int `json:"limit"`
		Total      int `json:"total"`
		TotalPages int `json:"total_pages"`
	} `json:"pagination"`
}

// PaginationParams represents pagination parameters extracted from query parameters
type PaginationParams struct {
	Page  int
	Limit int
}

// DefaultPaginationParams returns default pagination parameters
func DefaultPaginationParams() PaginationParams {
	return PaginationParams{
		Page:  1,
		Limit: 10,
	}
}

// ExtractPaginationParams extracts pagination parameters from request query parameters
func ExtractPaginationParams(r *http.Request) PaginationParams {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	return PaginationParams{
		Page:  page,
		Limit: limit,
	}
}

// ReplyWithPaginatedData responds with a paginated set of entities
func ReplyWithPaginatedData(w http.ResponseWriter, statusCode int, data any, total int, params PaginationParams) {
	totalPages := (total + params.Limit - 1) / params.Limit // Ceiling division

	response := PaginatedResponse{
		Data: data,
	}
	response.Pagination.Page = params.Page
	response.Pagination.Limit = params.Limit
	response.Pagination.Total = total
	response.Pagination.TotalPages = totalPages

	ReplyJSONResponse(w, statusCode, response)
}

func ReplyWithError(w http.ResponseWriter, statusCode int, errMsg string) {
	errResponse := &ErrorResponse{
		Message: errMsg,
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errResponse)
}

func ReplyJSONResponse(w http.ResponseWriter, statusCode int, output any) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(output)
}

func DecodeJSONBody(r *http.Request, placeholder any) error {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("reading request body: %w", err)
	}

	if err := json.Unmarshal(reqBody, placeholder); err != nil {
		return fmt.Errorf("marshaling json: %w", err)
	}

	return nil
}

func GetQueryParamMapKeyValue(r *http.Request, name string) (string, string) {
	queryVal := r.URL.Query().Get(name)
	pattern := regexp.MustCompile(`(\w+[\w \-_.]+):(\w+[\w \-_.]+)`)
	kv := pattern.FindStringSubmatch(queryVal)
	if len(kv) < 3 {
		return "", ""
	}

	return kv[1], kv[2]
}
