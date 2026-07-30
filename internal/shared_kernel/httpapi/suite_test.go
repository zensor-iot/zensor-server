package httpapi_test

import (
	"io"
	"log/slog"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHTTPAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Shared Kernel HTTPAPI Suite")
}

var _ = BeforeEach(func() {
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
})
