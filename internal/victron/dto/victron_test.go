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
				topic := "N/d41243b4e8e4/battery/512/Dc/0/Voltage"
				serviceType, instance, path, ok := dto.ParseVictronTopic(topic)
				Expect(ok).To(BeTrue())
				Expect(serviceType).To(Equal("battery"))
				Expect(instance).To(Equal(512))
				Expect(path).To(Equal("Dc/0/Voltage"))
			})
		})

		When("topic has multiple path segments", func() {
			It("should join all path segments after instance", func() {
				topic := "N/d41243b4e8e4/solarcharger/279/Yield/Power"
				serviceType, instance, path, ok := dto.ParseVictronTopic(topic)
				Expect(ok).To(BeTrue())
				Expect(serviceType).To(Equal("solarcharger"))
				Expect(instance).To(Equal(279))
				Expect(path).To(Equal("Yield/Power"))
			})
		})

		When("topic is a simple path like Soc", func() {
			It("should parse with empty path", func() {
				topic := "N/d41243b4e8e4/battery/512/Soc"
				serviceType, instance, path, ok := dto.ParseVictronTopic(topic)
				Expect(ok).To(BeTrue())
				Expect(serviceType).To(Equal("battery"))
				Expect(instance).To(Equal(512))
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
				_, _, _, ok := dto.ParseVictronTopic("N/d41243b4e8e4/battery/abc/Voltage")
				Expect(ok).To(BeFalse())
			})
		})
	})

	Describe("ParseTelemetry", func() {
		When("payload is valid JSON with value and text", func() {
			It("should parse successfully", func() {
				topic := "N/d41243b4e8e4/battery/512/Dc/0/Voltage"
				payload := []byte(`{"value": 12.8, "text": "12.8 V"}`)
				telemetry, err := dto.ParseTelemetry(topic, payload)
				Expect(err).NotTo(HaveOccurred())
				Expect(telemetry.PortalID).To(Equal("d41243b4e8e4"))
				Expect(telemetry.ServiceType).To(Equal("battery"))
				Expect(telemetry.Instance).To(Equal(512))
				Expect(telemetry.Path).To(Equal("Dc/0/Voltage"))
				Expect(telemetry.Value.Value).To(Equal(12.8))
				Expect(telemetry.Value.Text).To(Equal("12.8 V"))
			})
		})

		When("payload is invalid JSON", func() {
			It("should return an error", func() {
				_, err := dto.ParseTelemetry("N/d41243b4e8e4/battery/512/Soc", []byte(`not json`))
				Expect(err).To(HaveOccurred())
			})
		})

		When("topic is invalid", func() {
			It("should return an error", func() {
				_, err := dto.ParseTelemetry("invalid", []byte(`{"value": 1}`))
				Expect(err).To(HaveOccurred())
			})
		})

		When("payload value is a string, such as a serial number or firmware version", func() {
			It("should parse successfully, keeping the string in Text and leaving Value at zero", func() {
				topic := "N/d41243b4e8e4/system/0/Serial"
				payload := []byte(`{"value": "d41243b4e8e4"}`)
				telemetry, err := dto.ParseTelemetry(topic, payload)
				Expect(err).NotTo(HaveOccurred())
				Expect(telemetry.Value.Value).To(Equal(0.0))
				Expect(telemetry.Value.Text).To(Equal("d41243b4e8e4"))
				Expect(telemetry.Value.IsNumeric()).To(BeFalse())
			})
		})

		When("payload value is a JSON number", func() {
			It("should mark the value as numeric", func() {
				topic := "N/d41243b4e8e4/battery/512/Soc"
				payload := []byte(`{"value": 87}`)
				telemetry, err := dto.ParseTelemetry(topic, payload)
				Expect(err).NotTo(HaveOccurred())
				Expect(telemetry.Value.Value).To(Equal(87.0))
				Expect(telemetry.Value.IsNumeric()).To(BeTrue())
			})
		})

		When("payload value is null, as Victron sends for an unavailable/disconnected path", func() {
			It("should parse successfully with a zero value", func() {
				topic := "N/d41243b4e8e4/battery/512/Soc"
				payload := []byte(`{"value": null}`)
				telemetry, err := dto.ParseTelemetry(topic, payload)
				Expect(err).NotTo(HaveOccurred())
				Expect(telemetry.Value.Value).To(Equal(0.0))
				Expect(telemetry.Value.IsNumeric()).To(BeFalse())
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
