package httpserver

import "net/http"

type Controller interface {
	AddRoutes(*http.ServeMux)
}
