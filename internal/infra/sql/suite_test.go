package sql_test

import (
	"io"
	"log/slog"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestSql(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "SQL Suite")
}

var _ = ginkgo.BeforeEach(func() {
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
})
