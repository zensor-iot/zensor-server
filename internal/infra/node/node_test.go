package node_test

import (
	"net"
	"zensor-server/internal/infra/node"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Node", func() {
	ginkgo.Context("GetNodeInfo", func() {
		ginkgo.It("should return node information with all fields", func() {
			nodeInfo := node.GetNodeInfo()

			gomega.Expect(nodeInfo).ToNot(gomega.BeNil())
			gomega.Expect(nodeInfo.ID).ToNot(gomega.BeEmpty())
			gomega.Expect(nodeInfo.IPAddress).ToNot(gomega.BeEmpty())
			gomega.Expect(nodeInfo.Version).ToNot(gomega.BeEmpty())
			gomega.Expect(nodeInfo.CommitHash).ToNot(gomega.BeEmpty())
		})

		ginkgo.It("should return a valid UUID for node ID", func() {
			nodeInfo := node.GetNodeInfo()
			gomega.Expect(nodeInfo.ID).ToNot(gomega.BeEmpty())
			gomega.Expect(len(nodeInfo.ID)).To(gomega.Equal(36)) // UUID length
		})

		ginkgo.It("should return the same node ID on multiple calls (singleton)", func() {
			nodeInfo1 := node.GetNodeInfo()
			nodeInfo2 := node.GetNodeInfo()
			gomega.Expect(nodeInfo1.ID).To(gomega.Equal(nodeInfo2.ID))
		})

		ginkgo.It("should return a valid IP address", func() {
			nodeInfo := node.GetNodeInfo()
			gomega.Expect(nodeInfo.IPAddress).ToNot(gomega.BeEmpty())

			// Should be a valid IP address
			ip := net.ParseIP(nodeInfo.IPAddress)
			gomega.Expect(ip).ToNot(gomega.BeNil())
		})

		ginkgo.It("should return the same IP address on multiple calls (singleton)", func() {
			nodeInfo1 := node.GetNodeInfo()
			nodeInfo2 := node.GetNodeInfo()
			gomega.Expect(nodeInfo1.IPAddress).To(gomega.Equal(nodeInfo2.IPAddress))
		})

		ginkgo.It("should return version and commit hash", func() {
			nodeInfo := node.GetNodeInfo()
			gomega.Expect(nodeInfo.Version).To(gomega.BeAssignableToTypeOf(""))
			gomega.Expect(nodeInfo.Version).To(gomega.Not(gomega.BeEmpty()))
			gomega.Expect(nodeInfo.CommitHash).To(gomega.BeAssignableToTypeOf(""))
			gomega.Expect(nodeInfo.CommitHash).To(gomega.Not(gomega.BeEmpty()))
		})
	})
})
