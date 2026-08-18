package o11y_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"zensor-server/internal/infra/o11y"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("TeeHandler", func() {
	ginkgo.Context("NewTeeHandler", func() {
		ginkgo.When("logging a record", func() {
			ginkgo.It("should deliver it to every underlying handler", func() {
				var bufA, bufB bytes.Buffer
				handlerA := slog.NewTextHandler(&bufA, &slog.HandlerOptions{Level: slog.LevelInfo})
				handlerB := slog.NewTextHandler(&bufB, &slog.HandlerOptions{Level: slog.LevelInfo})

				logger := slog.New(o11y.NewTeeHandler(slog.LevelInfo, handlerA, handlerB))
				logger.Info("hello", slog.String("key", "value"))

				gomega.Expect(bufA.String()).To(gomega.ContainSubstring("hello"))
				gomega.Expect(bufA.String()).To(gomega.ContainSubstring("key=value"))
				gomega.Expect(bufB.String()).To(gomega.ContainSubstring("hello"))
				gomega.Expect(bufB.String()).To(gomega.ContainSubstring("key=value"))
			})
		})

		ginkgo.When("logging below the configured level", func() {
			ginkgo.It("should not deliver the record to any handler", func() {
				var bufA bytes.Buffer
				handlerA := slog.NewTextHandler(&bufA, &slog.HandlerOptions{Level: slog.LevelDebug})

				logger := slog.New(o11y.NewTeeHandler(slog.LevelInfo, handlerA))
				logger.Debug("invisible")

				gomega.Expect(bufA.String()).To(gomega.BeEmpty())
			})
		})

		ginkgo.When("using WithAttrs and WithGroup", func() {
			ginkgo.It("should propagate attributes and groups to every handler", func() {
				var bufA, bufB bytes.Buffer
				handlerA := slog.NewTextHandler(&bufA, &slog.HandlerOptions{Level: slog.LevelInfo})
				handlerB := slog.NewTextHandler(&bufB, &slog.HandlerOptions{Level: slog.LevelInfo})

				logger := slog.New(o11y.NewTeeHandler(slog.LevelInfo, handlerA, handlerB)).
					With(slog.String("version", "1.2.3")).
					WithGroup("req")
				logger.Info("done", slog.String("id", "42"))

				for _, out := range []string{bufA.String(), bufB.String()} {
					gomega.Expect(out).To(gomega.ContainSubstring("version=1.2.3"))
					gomega.Expect(out).To(gomega.ContainSubstring("req.id=42"))
				}
			})
		})

		ginkgo.When("checking Enabled", func() {
			ginkgo.It("should report enabled only at or above the configured level", func() {
				handler := o11y.NewTeeHandler(slog.LevelWarn, slog.NewTextHandler(&strings.Builder{}, nil))

				gomega.Expect(handler.Enabled(context.Background(), slog.LevelInfo)).To(gomega.BeFalse())
				gomega.Expect(handler.Enabled(context.Background(), slog.LevelWarn)).To(gomega.BeTrue())
				gomega.Expect(handler.Enabled(context.Background(), slog.LevelError)).To(gomega.BeTrue())
			})
		})
	})
})
