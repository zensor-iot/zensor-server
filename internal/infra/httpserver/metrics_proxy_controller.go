package httpserver

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// MetricsProxyController exposes the Prometheus-compatible query API of a
// VictoriaMetrics instance to the SPA under /v1/metrics/*. Every request
// below that prefix is reverse-proxied to <baseURL>/api/v1/* so the browser
// never talks to VictoriaMetrics directly (no CORS, no exposed internals).
type MetricsProxyController struct {
	baseURL string
	proxy   *httputil.ReverseProxy
}

func NewMetricsProxyController(baseURL string) *MetricsProxyController {
	target, err := url.Parse(baseURL)
	if err != nil {
		panic(err)
	}

	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.Out.URL.Scheme = target.Scheme
			pr.Out.URL.Host = target.Host
			pr.Out.URL.Path = joinURLPath(target.Path, "/api/v1/"+strings.TrimPrefix(pr.In.URL.Path, "/v1/metrics/"))
			pr.Out.URL.RawQuery = pr.In.URL.RawQuery
			pr.SetXForwarded()
		},
		FlushInterval: time.Second,
	}

	return &MetricsProxyController{
		baseURL: baseURL,
		proxy:   proxy,
	}
}

var _ Controller = (*MetricsProxyController)(nil)

func (c *MetricsProxyController) AddRoutes(router *http.ServeMux) {
	router.Handle("/v1/metrics/", c.proxy)
}

// joinURLPath joins two URL paths without losing trailing slashes, mirroring
// net/http/httputil's singleJoiningSlash helper.
func joinURLPath(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
