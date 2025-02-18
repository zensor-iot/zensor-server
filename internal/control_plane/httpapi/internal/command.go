package internal

import "encoding/json"

type CommandSendRequest struct {
	Priority string                    `json:"priority"`
	Payload  CommandSendPayloadRequest `json:"payload"`
}

type CommandSendPayloadRequest struct {
	Index uint8 `json:"index"`
	Value uint8 `json:"value"`
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
