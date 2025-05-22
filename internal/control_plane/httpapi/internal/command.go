package internal

import (
	"encoding/json"
	"zensor-server/internal/infra/utils"
)

type CommandSendRequest struct {
	Priority string                      `json:"priority"`
	Sequence []CommandSendPayloadRequest `json:"sequence"`
}

type CommandSendPayloadRequest struct {
	WaitFor  utils.Duration `json:"wait_for"`
	Priority string         `json:"priority"`
	Index    uint8          `json:"index"`
	Value    uint8          `json:"value"`
}

func (c *CommandSendRequest) UnmarshalJSON(data []byte) error {
	type Alias CommandSendRequest
	defaults := &Alias{
		Priority: "NORMAL",
	}

	if err := json.Unmarshal(data, defaults); err != nil {
		return err
	}

	*c = CommandSendRequest(*defaults)
	return nil
}

type CommandResponse struct {
	ID       string `json:"id"`
	Priority string `json:"priority"`
}
