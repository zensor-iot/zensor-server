package httpserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
)

type ErrorResponse struct {
	Message string `json:"message,omitempty"`
}

func ReplyWithError(w http.ResponseWriter, statusCode int, errMsg string) {
	errResponse := &ErrorResponse{
		Message: errMsg,
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errResponse)
}

func ReplyJSONResponse(w http.ResponseWriter, statusCode int, output interface{}) {
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

func GetPathParam(r *http.Request, name string) string {
	vars := mux.Vars(r)
	return vars[name]
}

func GetQueryParam(r *http.Request, name string) string {
	val := r.URL.Query().Get(name)
	return val
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
