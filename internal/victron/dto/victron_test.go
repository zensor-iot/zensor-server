package dto_test

import (
	"zensor-server/internal/victron/dto"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Victron DTO", func() {

	Describe("ParseVictronTopic", func() {
		When("topic is a valid battery voltage path", func() {
			It("should parse service type, instance, and path", func() {
				topic := "N/abc123/system/0/Battery/0/Dc/0/Voltage"
				serviceType, instance, path, ok := dto.ParseVictronTopic(topic)
				Expect(ok).To(BeTrue())
				Expect(serviceType).To(Equal("Battery"))
				Expect(instance).To(Equal(0))
				Expect(path).To(Equal("Dc/0/Voltage"))
			})
		})

		When("topic has multiple path segments", func() {
			It("should join all path segments after instance", func() {
				topic := "N/abc123/system/0/SolarCharger/1/Dc/0/Power"
				serviceType, instance, path, ok := dto.ParseVictronTopic(topic)
				Expect(ok).To(BeTrue())
				Expect(serviceType).To(Equal("SolarCharger"))
				Expect(instance).To(Equal(1))
				Expect(path).To(Equal("Dc/0/Power"))
			})
		})

		When("topic is a simple path like Soc", func() {
			It("should parse with empty path", func() {
				topic := "N/abc123/system/0/Battery/0/Soc"
				serviceType, instance, path, ok := dto.ParseVictronTopic(topic)
				Expect(ok).To(BeTrue())
				Expect(serviceType).To(Equal("Battery"))
				Expect(instance).To(Equal(0))
				Expect(path).To(Equal("Soc"))
			})
		})

		When("topic does not match the pattern", func() {
			It("should return not ok", func() {
				_, _, _, ok := dto.ParseVictronTopic("invalid/topic")
				Expect(ok).To(BeFalse())
			})
		})

		When("topic has non-numeric instance", func() {
			It("should return not ok", func() {
				_, _, _, ok := dto.ParseVictronTopic("N/abc123/system/0/Battery/abc/Voltage")
				Expect(ok).To(BeFalse())
			})
		})
	})

	Describe("ParseTelemetry", func() {
		When("payload is valid JSON with value and text", func() {
			It("should parse successfully", func() {
				topic := "N/abc123/system/0/Battery/0/Dc/0/Voltage"
				payload := []byte(`{"value": 12.8, "text": "12.8 V"}`)
				telemetry, err := dto.ParseTelemetry(topic, payload)
				Expect(err).NotTo(HaveOccurred())
				Expect(telemetry.PortalID).To(Equal("abc123"))
				Expect(telemetry.ServiceType).To(Equal("Battery"))
				Expect(telemetry.Instance).To(Equal(0))
				Expect(telemetry.Path).To(Equal("Dc/0/Voltage"))
				Expect(telemetry.Value.Value).To(Equal(12.8))
				Expect(telemetry.Value.Text).To(Equal("12.8 V"))
			})
		})

		When("payload is invalid JSON", func() {
			It("should return an error", func() {
				_, err := dto.ParseTelemetry("N/abc123/system/0/Battery/0/Soc", []byte(`not json`))
				Expect(err).To(HaveOccurred())
			})
		})

		When("topic is invalid", func() {
			It("should return an error", func() {
				_, err := dto.ParseTelemetry("invalid", []byte(`{"value": 1}`))
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("ToSnakeName", func() {
		It("should convert PascalCase to snake_case", func() {
			Expect(dto.ToSnakeName("Battery")).To(Equal("battery"))
			Expect(dto.ToSnakeName("SolarCharger")).To(Equal("solar_charger"))
			Expect(dto.ToSnakeName("PvInverter")).To(Equal("pv_inverter"))
			Expect(dto.ToSnakeName("AcLoad")).To(Equal("ac_load"))
		})
	})
})
