package node

import (
	"net"
	"sync"

	"github.com/google/uuid"
)

// Node represents the current application node with its metadata
type Node struct {
	ID         string
	IPAddress  string
	Version    string
	CommitHash string
}

var Version = "development"
var CommitHash = "unknown"

var (
	nodeID     string
	nodeIDOnce sync.Once
	nodeIP     string
	nodeIPOnce sync.Once
)

// GetNodeInfo returns the current node information
func GetNodeInfo() *Node {
	return &Node{
		ID:         getNodeID(),
		IPAddress:  getNodeIPAddress(),
		Version:    Version,
		CommitHash: CommitHash,
	}
}

// getNodeID returns the current node ID
func getNodeID() string {
	nodeIDOnce.Do(func() {
		nodeID = generateNodeID()
	})
	return nodeID
}

// getNodeIPAddress returns the current node IP address
func getNodeIPAddress() string {
	nodeIPOnce.Do(func() {
		nodeIP = getNodeIPAddressInternal()
	})
	return nodeIP
}

func generateNodeID() string {
	return uuid.New().String()
}

func getNodeIPAddressInternal() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
