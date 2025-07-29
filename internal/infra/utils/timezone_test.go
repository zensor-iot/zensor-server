package utils_test

import (
	"zensor-server/internal/infra/utils"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Timezone", func() {
	ginkgo.Context("ValidateTimezone", func() {
		ginkgo.When("validating timezones", func() {
			ginkgo.It("should validate UTC timezone", func() {
				err := utils.ValidateTimezone("UTC")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.It("should validate America/New_York timezone", func() {
				err := utils.ValidateTimezone("America/New_York")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.It("should validate Europe/London timezone", func() {
				err := utils.ValidateTimezone("Europe/London")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.It("should validate Asia/Tokyo timezone", func() {
				err := utils.ValidateTimezone("Asia/Tokyo")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.It("should validate Australia/Sydney timezone", func() {
				err := utils.ValidateTimezone("Australia/Sydney")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.It("should validate EST timezone", func() {
				err := utils.ValidateTimezone("EST")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})

		ginkgo.When("validating invalid timezones", func() {
			ginkgo.It("should return error for empty timezone", func() {
				err := utils.ValidateTimezone("")
				gomega.Expect(err).To(gomega.HaveOccurred())
			})

			ginkgo.It("should return error for PST timezone", func() {
				err := utils.ValidateTimezone("PST")
				gomega.Expect(err).To(gomega.HaveOccurred())
			})

			ginkgo.It("should return error for random string timezone", func() {
				err := utils.ValidateTimezone("Invalid/Timezone/Name")
				gomega.Expect(err).To(gomega.HaveOccurred())
			})

			ginkgo.It("should return error for timezone with spaces", func() {
				err := utils.ValidateTimezone("America/New York")
				gomega.Expect(err).To(gomega.HaveOccurred())
			})
		})
	})

	ginkgo.Context("IsValidTimezone", func() {
		ginkgo.When("checking valid timezones", func() {
			ginkgo.It("should return true for UTC timezone", func() {
				result := utils.IsValidTimezone("UTC")
				gomega.Expect(result).To(gomega.BeTrue())
			})

			ginkgo.It("should return true for America/New_York timezone", func() {
				result := utils.IsValidTimezone("America/New_York")
				gomega.Expect(result).To(gomega.BeTrue())
			})

			ginkgo.It("should return true for EST timezone", func() {
				result := utils.IsValidTimezone("EST")
				gomega.Expect(result).To(gomega.BeTrue())
			})
		})

		ginkgo.When("checking invalid timezones", func() {
			ginkgo.It("should return false for empty timezone", func() {
				result := utils.IsValidTimezone("")
				gomega.Expect(result).To(gomega.BeFalse())
			})

			ginkgo.It("should return false for PST timezone", func() {
				result := utils.IsValidTimezone("PST")
				gomega.Expect(result).To(gomega.BeFalse())
			})
		})
	})
})
