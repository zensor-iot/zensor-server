package dto_test

import (
	"encoding/json"
	"zensor-server/internal/data_plane/dto"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("TTNMessage", func() {
	ginkgo.Context("WithCorrelationIDs", func() {
		var ttnMsg dto.TTNMessage

		ginkgo.When("creating a TTN message with correlation IDs", func() {
			ginkgo.BeforeEach(func() {
				ttnMsg = dto.TTNMessage{
					Downlinks: []dto.TTNMessageDownlink{
						{
							FPort:          15,
							FrmPayload:     []byte{1, 2, 3},
							Priority:       "NORMAL",
							CorrelationIDs: []string{"zensor:cmd-123", "cmd-456"},
						},
					},
				}
			})

			ginkgo.It("should marshal and unmarshal correctly", func() {
				// Marshal to JSON to verify the structure
				jsonData, err := json.Marshal(ttnMsg)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Verify the JSON contains correlation_ids
				jsonStr := string(jsonData)
				gomega.Expect(jsonStr).To(gomega.ContainSubstring("correlation_ids"))
				gomega.Expect(jsonStr).To(gomega.ContainSubstring("zensor:cmd-123"))
				gomega.Expect(jsonStr).To(gomega.ContainSubstring("cmd-456"))

				// Unmarshal back to verify round-trip
				var unmarshaledMsg dto.TTNMessage
				err = json.Unmarshal(jsonData, &unmarshaledMsg)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Verify the correlation IDs are preserved
				gomega.Expect(unmarshaledMsg.Downlinks).To(gomega.HaveLen(1))
				gomega.Expect(unmarshaledMsg.Downlinks[0].CorrelationIDs).To(gomega.Equal([]string{"zensor:cmd-123", "cmd-456"}))
			})
		})
	})

	ginkgo.Context("WithoutCorrelationIDs", func() {
		var ttnMsg dto.TTNMessage

		ginkgo.When("creating a TTN message without correlation IDs", func() {
			ginkgo.BeforeEach(func() {
				ttnMsg = dto.TTNMessage{
					Downlinks: []dto.TTNMessageDownlink{
						{
							FPort:      15,
							FrmPayload: []byte{1, 2, 3},
							Priority:   "NORMAL",
						},
					},
				}
			})

			ginkgo.It("should handle backward compatibility", func() {
				// Marshal to JSON to verify the structure
				jsonData, err := json.Marshal(ttnMsg)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Verify the JSON doesn't contain correlation_ids (since it's empty)
				jsonStr := string(jsonData)
				gomega.Expect(jsonStr).NotTo(gomega.ContainSubstring("correlation_ids"))

				// Unmarshal back to verify round-trip
				var unmarshaledMsg dto.TTNMessage
				err = json.Unmarshal(jsonData, &unmarshaledMsg)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Verify the correlation IDs are empty
				gomega.Expect(unmarshaledMsg.Downlinks).To(gomega.HaveLen(1))
				gomega.Expect(unmarshaledMsg.Downlinks[0].CorrelationIDs).To(gomega.BeNil())
			})
		})
	})
})
