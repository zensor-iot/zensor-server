package httpapi

import (
	"log/slog"
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
	router.Handle("GET /v1/devices", c.listDevices())
	router.Handle("POST /v1/devices", c.createDevice())
	router.Handle("POST /v1/devices/{id}/commands", c.sendCommand())
}

func (c *DeviceController) listDevices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := c.service.AllDevices(r.Context())
		if err != nil {
			http.Error(w, "service all devices", http.StatusInternalServerError)
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

func (c *DeviceController) sendCommand() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var body internal.CommandSendRequest
		err := httpserver.DecodeJSONBody(r, &body)
		if err != nil {
			slog.Error("decoding json body", slog.String("error", err.Error()))
			http.Error(w, createDeviceErrMessage, http.StatusBadRequest)
			return
		}

		cmd := domain.Command{
			Device: domain.Device{ID: domain.ID(id)},
			Payload: domain.CommandPayload{
				Index: domain.Index(body.Payload.Index),
				Value: domain.CommandValue(body.Payload.Value),
			},
			Priority: domain.CommandPriority(body.Priority),
		}

		err = c.service.QueueCommand(r.Context(), cmd)
		if err != nil {
			slog.Error("queue command failed", slog.String("error", err.Error()))
			http.Error(w, "queue command failed", http.StatusInternalServerError)
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusCreated, nil)
	}
}
