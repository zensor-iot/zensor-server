package replication

import (
	"io"
	"log/slog"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestReplication(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Replication Suite")
}

var _ = ginkgo.BeforeEach(func() {
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
})
