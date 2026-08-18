// Package httpserver provides the HTTP server, middleware, and helpers.
package httpserver

import "net/http"

type Controller interface {
	AddRoutes(*http.ServeMux)
}
