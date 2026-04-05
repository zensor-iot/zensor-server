package notification_test

import (
	"context"
	"path/filepath"

	"zensor-server/internal/infra/notification"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("FCM dummy credentials fixture", func() {
	ginkgo.It("loads testdata/fcm_service_account_dummy.json for startup-only use", func() {
		path := filepath.Join("testdata", "fcm_service_account_dummy.json")

		_, err := notification.NewFCMClient(context.Background(), notification.FCMConfig{
			ProjectID:          "ci-dummy-project",
			ServiceAccountPath: path,
		})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})
})
