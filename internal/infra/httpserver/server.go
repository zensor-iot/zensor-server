package httpserver

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"

	_ "net/http/pprof"
)

type Server interface {
	Run()
	Shutdown()
}

var _ Server = &StandardServer{}

type StandardServer struct {
	server *http.Server
}

func (s *StandardServer) Run() {
	s.server.ListenAndServe()
}

func (s *StandardServer) Shutdown() {
	s.server.Shutdown(context.Background())
}

func NewServer(controllers ...Controller) *StandardServer {
	router := http.NewServeMux()

	// Configure CORS for development
	c := cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
			"https://portal.zensor-iot.net",
		},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
		},
		ExposedHeaders: []string{
			"Link",
		},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})

	server := &StandardServer{
		&http.Server{
			Addr:    ":3000",
			Handler: c.Handler(router),
		},
	}

	router.Handle("GET /healthz", getHealthz())
	router.Handle("GET /metrics", promhttp.Handler())

	for _, controller := range controllers {
		controller.AddRoutes(router)
	}

	return server
}

func getHealthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		output := map[string]string{"status": "success"}
		ReplyJSONResponse(w, http.StatusOK, output)
	}
}
