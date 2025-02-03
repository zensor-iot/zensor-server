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
	server := &StandardServer{
		&http.Server{
			Addr:    ":3000",
			Handler: cors.Default().Handler(router),
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
