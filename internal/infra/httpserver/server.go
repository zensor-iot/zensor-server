package httpserver

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"zensor-server/internal/infra/node"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
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
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

func (s *StandardServer) Shutdown() {
	if err := s.server.Shutdown(context.Background()); err != nil {
		panic(err)
	}
}

func NewServer(controllers ...Controller) *StandardServer {
	router := http.NewServeMux()

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

func createUserHeaderMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span := GetSpanFromContext(r)

			userID := r.Header.Get("X-User-ID")
			userName := r.Header.Get("X-User-Name")
			userEmail := r.Header.Get("X-User-Email")

			if userID != "" {
				span.SetAttributes(attribute.String("user.id", userID))
			}
			if userName != "" {
				span.SetAttributes(attribute.String("user.name", userName))
			}
			if userEmail != "" {
				span.SetAttributes(attribute.String("user.email", userEmail))
			}

			next.ServeHTTP(w, r)
		})
	}
}

func createTracingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			propagator := b3.New()
			ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			tracer := otel.Tracer("zensor-server")
			ctx, span := tracer.Start(ctx, "http.request",
				trace.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.url", r.URL.String()),
					attribute.String("http.user_agent", r.UserAgent()),
					attribute.String("http.remote_addr", r.RemoteAddr),
					attribute.String("span.kind", "server"),
					attribute.String("component", "http-server"),
				),
			)
			defer span.End()

			r = r.WithContext(ctx)

			propagator.Inject(ctx, propagation.HeaderCarrier(w.Header()))

			wrapped := &statusCodeResponseWriter{ResponseWriter: w}

			next.ServeHTTP(wrapped, r)

			span.SetAttributes(attribute.Int("http.status_code", wrapped.statusCode))

			setTraceStatus(span, wrapped.statusCode)
		})
	}
}

func setTraceStatus(span trace.Span, statusCode int) {
	switch {
	case statusCode >= 200 && statusCode < 300:
		span.SetStatus(codes.Ok, "")
	case statusCode >= 500 && statusCode < 600:
		span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", statusCode))
	default:
		span.SetStatus(codes.Unset, "")
	}
}

type statusCodeResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusCodeResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusCodeResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("underlying ResponseWriter does not support hijacking")
}

func getHealthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		span := GetSpanFromContext(r)
		span.SetAttributes(attribute.String("endpoint", "healthz"))

		nodeInfo := node.GetNodeInfo()
		output := HealthzResponse{
			Status:     "success",
			Version:    nodeInfo.Version,
			CommitHash: nodeInfo.CommitHash,
		}
		ReplyJSONResponse(w, http.StatusOK, output)
	}
}

type HealthzResponse struct {
	Status     string `json:"status"`
	Version    string `json:"version"`
	CommitHash string `json:"commit_hash"`
}
