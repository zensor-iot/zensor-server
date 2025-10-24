package node_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestNode(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Node Suite")
}

var _ = ginkgo.BeforeEach(func() {
	// Discard slog output during tests
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
})
