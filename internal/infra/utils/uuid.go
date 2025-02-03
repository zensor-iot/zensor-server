package utils

import (
	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/google/uuid"
)

func GenerateUUID() string {
	return uuid.NewString()
}

func GenerateHEX(size int) string {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return strings.Repeat("0", size)
	}
	return hex.EncodeToString(bytes)
}
