package httpserver

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

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
			"X-User-ID",
			"X-User-Name",
			"X-User-Email",
		},
		ExposedHeaders: []string{
			"Link",
		},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})

	// Create middleware
	tracingMiddleware := createTracingMiddleware()
	userHeaderMiddleware := createUserHeaderMiddleware()
	metricsMiddleware := MetricsMiddleware()

	server := &StandardServer{
		&http.Server{
			Addr: ":3000",
			Handler: c.Handler(
				metricsMiddleware(
					tracingMiddleware(
						userHeaderMiddleware(router),
					),
				),
			),
		},
	}

	router.Handle("GET /healthz", getHealthz())
	router.Handle("GET /metrics", promhttp.Handler())

	for _, controller := range controllers {
		controller.AddRoutes(router)
	}

	return server
}

// createUserHeaderMiddleware creates a middleware that validates user headers and adds them as span attributes
func createUserHeaderMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the current span from the request context
			span := GetSpanFromContext(r)

			// Check for required user headers
			userID := r.Header.Get("X-User-ID")
			userName := r.Header.Get("X-User-Name")
			userEmail := r.Header.Get("X-User-Email")

			// Add user information as span attributes if present
			if userID != "" {
				span.SetAttributes(attribute.String("user.id", userID))
			}
			if userName != "" {
				span.SetAttributes(attribute.String("user.name", userName))
			}
			if userEmail != "" {
				span.SetAttributes(attribute.String("user.email", userEmail))
			}

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// createTracingMiddleware creates a middleware that adds OpenTelemetry tracing to all requests
func createTracingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a new span for this request
			tracer := otel.Tracer("zensor-server")
			ctx, span := tracer.Start(r.Context(), "http.request",
				trace.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.url", r.URL.String()),
					attribute.String("http.user_agent", r.UserAgent()),
					attribute.String("http.remote_addr", r.RemoteAddr),
				),
			)
			defer span.End()

			// Add the span to the request context
			r = r.WithContext(ctx)

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

func getHealthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the current span from the request context
		span := GetSpanFromContext(r)
		span.SetAttributes(attribute.String("endpoint", "healthz"))

		output := map[string]string{"status": "success"}
		ReplyJSONResponse(w, http.StatusOK, output)
	}
}
