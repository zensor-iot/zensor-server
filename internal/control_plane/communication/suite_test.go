package communication_test

import (
	"io"
	"log/slog"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCommunication(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Communication Suite")
}

var _ = BeforeEach(func() {
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
})
