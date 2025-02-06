package dto

type TTNMessage struct {
	Downlinks []TTNMessageDownlink `json:"downlinks"`
}

type TTNMessageDownlink struct {
	FPort      uint8  `json:"f_port"`
	FrmPayload []byte `json:"frm_payload"`
	Priority   string `json:"priority"`
}
