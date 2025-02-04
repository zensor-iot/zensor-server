package internal

import "encoding/json"

type CommandSendRequest struct {
	Port       uint8  `json:"port"`
	RawPayload string `json:"raw_payload"`
	Priority   string `json:"priority"`
}

func (c *CommandSendRequest) UnmarshalJSON(data []byte) error {
	type Alias CommandSendRequest
	defaults := &Alias{
		Priority: "NORMAL",
		Port:     0,
	}

	if err := json.Unmarshal(data, defaults); err != nil {
		return err
	}

	*c = CommandSendRequest(*defaults)
	return nil
}
