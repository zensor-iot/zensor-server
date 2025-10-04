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

	ginkgo.Context("ParseExecutionTime", func() {
		ginkgo.When("parsing valid time strings", func() {
			ginkgo.It("should parse 24-hour format correctly", func() {
				hour, minute, err := utils.ParseExecutionTime("02:00")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(hour).To(gomega.Equal(2))
				gomega.Expect(minute).To(gomega.Equal(0))

				hour, minute, err = utils.ParseExecutionTime("14:30")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(hour).To(gomega.Equal(14))
				gomega.Expect(minute).To(gomega.Equal(30))

				hour, minute, err = utils.ParseExecutionTime("23:59")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(hour).To(gomega.Equal(23))
				gomega.Expect(minute).To(gomega.Equal(59))
			})
		})

		ginkgo.When("parsing invalid time strings", func() {
			ginkgo.It("should return error for invalid format", func() {
				_, _, err := utils.ParseExecutionTime("2:00:30")
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("execution time must be in HH:MM format"))

				_, _, err = utils.ParseExecutionTime("25:00")
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("hour must be between 0 and 23"))

				_, _, err = utils.ParseExecutionTime("12:60")
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("minute must be between 0 and 59"))

				_, _, err = utils.ParseExecutionTime("invalid")
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("execution time must be in HH:MM format"))
			})
		})
	})
})
