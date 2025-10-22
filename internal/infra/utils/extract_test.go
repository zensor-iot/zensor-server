package utils_test

import (
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ExtractFloat64Value", func() {
	Context("with different message types", func() {
		It("should extract value from struct with Value field", func() {
			msg := struct {
				Value float64
			}{Value: 42.5}

			result := utils.ExtractFloat64Value(msg, "Value")
			Expect(result).To(Equal(42.5))
		})

		It("should extract value from map with Value key", func() {
			msg := map[string]interface{}{
				"Value": 25.3,
			}

			result := utils.ExtractFloat64Value(msg, "Value")
			Expect(result).To(Equal(25.3))
		})

		It("should extract value from nested struct", func() {
			msg := struct {
				Data struct {
					Value float64
				}
			}{
				Data: struct {
					Value float64
				}{Value: 10.0},
			}

			result := utils.ExtractFloat64Value(msg, "Data.Value")
			Expect(result).To(Equal(10.0))
		})

		It("should extract value from broker message with nested data", func() {
			msg := async.BrokerMessage{
				Event: "sensor_data",
				Value: map[string]interface{}{
					"data": map[string]interface{}{
						"Value": 15.7,
					},
				},
			}

			result := utils.ExtractFloat64Value(msg, "Value.data.Value")
			Expect(result).To(Equal(15.7))
		})

		It("should return 0.0 for empty property name", func() {
			msg := struct {
				Value float64
			}{Value: 42.5}

			result := utils.ExtractFloat64Value(msg, "")
			Expect(result).To(Equal(0.0))
		})

		It("should return 0.0 when property not found", func() {
			msg := struct {
				Other float64
			}{Other: 42.5}

			result := utils.ExtractFloat64Value(msg, "Value")
			Expect(result).To(Equal(0.0))
		})
	})
})

var _ = Describe("ExtractStringValue", func() {
	Context("with different message types", func() {
		It("should extract string value from struct", func() {
			msg := struct {
				DeviceID string
			}{DeviceID: "device123"}

			result := utils.ExtractStringValue(msg, "DeviceID")
			Expect(result).To(Equal("device123"))
		})

		It("should extract string value from map", func() {
			msg := map[string]interface{}{
				"device_id": "device456",
			}

			result := utils.ExtractStringValue(msg, "device_id")
			Expect(result).To(Equal("device456"))
		})

		It("should convert non-string values to string", func() {
			msg := map[string]interface{}{
				"count": 42,
			}

			result := utils.ExtractStringValue(msg, "count")
			Expect(result).To(Equal("42"))
		})

		It("should return empty string for empty property name", func() {
			msg := struct {
				DeviceID string
			}{DeviceID: "device123"}

			result := utils.ExtractStringValue(msg, "")
			Expect(result).To(Equal(""))
		})

		It("should return empty string when property not found", func() {
			msg := struct {
				Other string
			}{Other: "value"}

			result := utils.ExtractStringValue(msg, "DeviceID")
			Expect(result).To(Equal(""))
		})
	})
})
