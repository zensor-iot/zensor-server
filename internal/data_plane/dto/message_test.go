package dto_test

import (
	"encoding/json"
	"zensor-server/internal/data_plane/dto"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Envelop", func() {
	ginkgo.Context("UnmarshalJSON", func() {
		var payload string
		var envelop dto.Envelop
		var err error

		ginkgo.JustBeforeEach(func() {
			envelop = dto.Envelop{}
			err = json.Unmarshal([]byte(payload), &envelop)
		})

		ginkgo.When("decoded payload contains arrays of sensor data", func() {
			ginkgo.BeforeEach(func() {
				payload = `{"uplink_message":{"decoded_payload":{"temperature":[{"index":0,"value":21.5},{"index":1,"value":22.5}]}}}`
			})

			ginkgo.It("should decode every reading", func() {
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(envelop.UplinkMessage.DecodedPayload["temperature"]).To(gomega.Equal([]dto.SensorData{
					{Index: 0, Value: 21.5},
					{Index: 1, Value: 22.5},
				}))
			})
		})

		ginkgo.When("decoded payload contains a scalar number", func() {
			ginkgo.BeforeEach(func() {
				payload = `{"uplink_message":{"decoded_payload":{"temperature":23.75}}}`
			})

			ginkgo.It("should decode it as a single reading", func() {
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(envelop.UplinkMessage.DecodedPayload["temperature"]).To(gomega.Equal([]dto.SensorData{
					{Index: 0, Value: 23.75},
				}))
			})
		})

		ginkgo.When("decoded payload contains a boolean", func() {
			ginkgo.BeforeEach(func() {
				payload = `{"uplink_message":{"decoded_payload":{"relay":true}}}`
			})

			ginkgo.It("should decode it as a numeric reading", func() {
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(envelop.UplinkMessage.DecodedPayload["relay"]).To(gomega.Equal([]dto.SensorData{
					{Index: 0, Value: 1},
				}))
			})
		})

		ginkgo.When("decoded payload contains an array of numbers", func() {
			ginkgo.BeforeEach(func() {
				payload = `{"uplink_message":{"decoded_payload":{"waterFlow":[10,20]}}}`
			})

			ginkgo.It("should index each element by position", func() {
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(envelop.UplinkMessage.DecodedPayload["waterFlow"]).To(gomega.Equal([]dto.SensorData{
					{Index: 0, Value: 10},
					{Index: 1, Value: 20},
				}))
			})
		})

		ginkgo.When("decoded payload mixes supported and unsupported entries", func() {
			ginkgo.BeforeEach(func() {
				payload = `{"end_device_ids":{"device_id":"device-1"},"uplink_message":{"port":15,"decoded_payload":{"temperature":23.75,"metadata":{"firmware":"1.2.3"},"label":"ok"}}}`
			})

			ginkgo.It("should keep the supported entries and skip the rest", func() {
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(envelop.EndDeviceIDs.DeviceID).To(gomega.Equal("device-1"))
				gomega.Expect(envelop.UplinkMessage.Port).To(gomega.Equal(uint8(15)))
				gomega.Expect(envelop.UplinkMessage.DecodedPayload).To(gomega.HaveLen(1))
				gomega.Expect(envelop.UplinkMessage.DecodedPayload["temperature"]).To(gomega.Equal([]dto.SensorData{
					{Index: 0, Value: 23.75},
				}))
			})
		})

		ginkgo.When("decoded payload is absent", func() {
			ginkgo.BeforeEach(func() {
				payload = `{"uplink_message":{"port":15,"frm_payload":"AQID"}}`
			})

			ginkgo.It("should decode the remaining fields", func() {
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(envelop.UplinkMessage.Port).To(gomega.Equal(uint8(15)))
				gomega.Expect(envelop.UplinkMessage.RawPayload).To(gomega.Equal([]byte{1, 2, 3}))
				gomega.Expect(envelop.UplinkMessage.DecodedPayload).To(gomega.BeNil())
			})
		})
	})
})
