package node

import (
	"context"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Node represents the current application node with its metadata.
type Node struct {
	ID         string
	IPAddress  string
	Version    string
	CommitHash string
}

var (
	Version    = "development"
	CommitHash = "unknown"
)

var (
	nodeID     string
	nodeIDOnce sync.Once
	nodeIP     string
	nodeIPOnce sync.Once
)

// GetNodeInfo returns the current node information.
func GetNodeInfo() *Node {
	return &Node{
		ID:         getNodeID(),
		IPAddress:  getNodeIPAddress(),
		Version:    Version,
		CommitHash: CommitHash,
	}
}

// getNodeID returns the current node ID.
func getNodeID() string {
	nodeIDOnce.Do(func() {
		nodeID = generateNodeID()
	})
	return nodeID
}

// getNodeIPAddress returns the current node IP address.
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
	dialer := &net.Dialer{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := dialer.DialContext(ctx, "udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer func() {
		if err := conn.Close(); err != nil {
			slog.Warn("failed to close UDP connection", slog.String("error", err.Error()))
		}
	}()

	localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return "127.0.0.1"
	}
	return localAddr.IP.String()
}
