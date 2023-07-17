package rest

import (
	"net/http"

	"zensor-server/internal/persistence"

	"zensor-server/internal/logger"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
)

type RestServer interface {
	Run()
}

type RestServerWrapper struct {
	server          *http.Server
	eventRepository persistence.EventRepository
}

func (s *RestServerWrapper) Run() {
	logger.Info("initializing...")
	s.server.ListenAndServe()
}

func NewRestServer(eventRepository persistence.EventRepository) RestServer {
	router := mux.NewRouter().StrictSlash(true)
	httpServer := &http.Server{Addr: ":3000", Handler: cors.Default().Handler(router)}
	restServer := &RestServerWrapper{httpServer, eventRepository}
	router.HandleFunc("/healthz", getHealthz).Methods("GET")
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")
	router.HandleFunc("/devices", getDevices).Methods("GET")
	router.HandleFunc("/events", restServer.getEvents).Methods("GET")
	return restServer
}

func getHealthz(w http.ResponseWriter, r *http.Request) {
	output := map[string]string{"status": "success"}
	replyJSONResponse(w, output)
}

func getDevices(w http.ResponseWriter, r *http.Request) {
	output := []string{}
	replyJSONResponse(w, output)
}

func (s *RestServerWrapper) getEvents(w http.ResponseWriter, r *http.Request) {
	output := s.eventRepository.GetEvents()
	replyJSONResponse(w, output)
}
