package dto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTTNMessage_WithCorrelationIDs(t *testing.T) {
	// Create a TTN message with correlation IDs
	ttnMsg := TTNMessage{
		Downlinks: []TTNMessageDownlink{
			{
				FPort:          15,
				FrmPayload:     []byte{1, 2, 3},
				Priority:       "NORMAL",
				CorrelationIDs: []string{"zensor:cmd-123", "cmd-456"},
			},
		},
	}

	// Marshal to JSON to verify the structure
	jsonData, err := json.Marshal(ttnMsg)
	assert.NoError(t, err)

	// Verify the JSON contains correlation_ids
	jsonStr := string(jsonData)
	assert.Contains(t, jsonStr, "correlation_ids")
	assert.Contains(t, jsonStr, "zensor:cmd-123")
	assert.Contains(t, jsonStr, "cmd-456")

	// Unmarshal back to verify round-trip
	var unmarshaledMsg TTNMessage
	err = json.Unmarshal(jsonData, &unmarshaledMsg)
	assert.NoError(t, err)

	// Verify the correlation IDs are preserved
	assert.Equal(t, 1, len(unmarshaledMsg.Downlinks))
	assert.Equal(t, []string{"zensor:cmd-123", "cmd-456"}, unmarshaledMsg.Downlinks[0].CorrelationIDs)
}

func TestTTNMessage_WithoutCorrelationIDs(t *testing.T) {
	// Create a TTN message without correlation IDs (backward compatibility)
	ttnMsg := TTNMessage{
		Downlinks: []TTNMessageDownlink{
			{
				FPort:      15,
				FrmPayload: []byte{1, 2, 3},
				Priority:   "NORMAL",
			},
		},
	}

	// Marshal to JSON to verify the structure
	jsonData, err := json.Marshal(ttnMsg)
	assert.NoError(t, err)

	// Verify the JSON doesn't contain correlation_ids (since it's empty)
	jsonStr := string(jsonData)
	assert.NotContains(t, jsonStr, "correlation_ids")

	// Unmarshal back to verify round-trip
	var unmarshaledMsg TTNMessage
	err = json.Unmarshal(jsonData, &unmarshaledMsg)
	assert.NoError(t, err)

	// Verify the correlation IDs are empty
	assert.Equal(t, 1, len(unmarshaledMsg.Downlinks))
	assert.Nil(t, unmarshaledMsg.Downlinks[0].CorrelationIDs)
}
