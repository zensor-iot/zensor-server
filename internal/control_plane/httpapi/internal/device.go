package internal

type DeviceListResponse struct {
	Data []DeviceResponse `json:"data"`
}

type DeviceResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	AppEUI string `json:"app_eui"`
	DevEUI string `json:"dev_eui"`
	AppKey string `json:"app_key"`
}

type DeviceCreateRequest struct {
	Name string `json:"name"`
}
