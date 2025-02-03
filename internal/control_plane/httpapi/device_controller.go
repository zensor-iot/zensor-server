package httpapi

import (
	"net/http"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/control_plane/httpapi/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/httpserver"
)

const (
	createDeviceErrMessage = "failed to create device"
)

func NewDeviceController(service usecases.DeviceService) *DeviceController {
	return &DeviceController{
		service,
	}
}

var _ httpserver.Controller = &DeviceController{}

type DeviceController struct {
	service usecases.DeviceService
}

func (c *DeviceController) AddRoutes(router *http.ServeMux) {
	router.Handle("GET /devices", c.listDevices())
	router.Handle("POST /devices", c.createDevice())
}

func (c *DeviceController) listDevices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := c.service.AllDevices(r.Context())
		if err != nil {
			http.Error(w, "failed to list devices", http.StatusInternalServerError)
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusOK, result)
	}
}

func (c *DeviceController) createDevice() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body internal.DeviceCreateRequest
		err := httpserver.DecodeJSONBody(r, &body)
		if err != nil {
			http.Error(w, createDeviceErrMessage, http.StatusBadRequest)
			return
		}

		device, err := domain.NewDeviceBuilder().
			WithName(body.Name).
			Build()
		if err != nil {
			http.Error(w, createDeviceErrMessage, http.StatusInternalServerError)
			return
		}

		err = c.service.CreateDevice(r.Context(), device)
		if err != nil {
			http.Error(w, createDeviceErrMessage, http.StatusInternalServerError)
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusCreated, nil)
	}
}
