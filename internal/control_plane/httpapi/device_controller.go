package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"time"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/control_plane/httpapi/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/httpserver"
	"zensor-server/internal/infra/utils"
)

const (
	createDeviceErrMessage           = "failed to create device"
	createDeviceDuplicatedErrMessage = "the device already exists"
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
	router.Handle("GET /v1/devices/{id}", c.getDevice())
	router.Handle("POST /v1/devices", c.createDevice())
	router.Handle("PUT /v1/devices/{id}", c.updateDevice())
	router.Handle("POST /v1/devices/{id}/commands", c.sendCommand())
}

func (c *DeviceController) listDevices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := c.service.AllDevices(r.Context())
		if err != nil {
			http.Error(w, "service all devices", http.StatusInternalServerError)
			return
		}

		response := internal.ToDeviceListResponse(result)
		httpserver.ReplyJSONResponse(w, http.StatusOK, response)
	}
}

func (c *DeviceController) getDevice() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		device, err := c.service.GetDevice(r.Context(), domain.ID(id))
		if err != nil {
			if errors.Is(err, usecases.ErrDeviceNotFound) {
				http.Error(w, "device not found", http.StatusNotFound)
				return
			}
			http.Error(w, "failed to get device", http.StatusInternalServerError)
			return
		}

		response := internal.ToDeviceResponse(device)
		httpserver.ReplyJSONResponse(w, http.StatusOK, response)
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

		displayName := body.DisplayName
		if displayName == "" {
			displayName = body.Name // Default display name to name if not provided
		}

		// Start building the device
		builder := domain.NewDeviceBuilder().
			WithName(body.Name).
			WithDisplayName(displayName)

		// Add LoRaWAN parameters if provided
		if body.AppEUI != nil && *body.AppEUI != "" {
			builder = builder.WithAppEUI(*body.AppEUI)
		}
		if body.DevEUI != nil && *body.DevEUI != "" {
			builder = builder.WithDevEUI(*body.DevEUI)
		}
		if body.AppKey != nil && *body.AppKey != "" {
			builder = builder.WithAppKey(*body.AppKey)
		}

		device, err := builder.Build()
		if err != nil {
			http.Error(w, createDeviceErrMessage, http.StatusInternalServerError)
			return
		}

		err = c.service.CreateDevice(r.Context(), device)
		if errors.Is(err, usecases.ErrDeviceDuplicated) {
			http.Error(w, createDeviceDuplicatedErrMessage, http.StatusConflict)
			return
		}

		if err != nil {
			http.Error(w, createDeviceErrMessage, http.StatusInternalServerError)
			return
		}

		response := internal.ToDeviceResponse(device)
		httpserver.ReplyJSONResponse(w, http.StatusCreated, response)
	}
}

func (c *DeviceController) updateDevice() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var body internal.DeviceUpdateRequest
		err := httpserver.DecodeJSONBody(r, &body)
		if err != nil {
			http.Error(w, "failed to decode request body", http.StatusBadRequest)
			return
		}

		err = c.service.UpdateDeviceDisplayName(r.Context(), domain.ID(id), body.DisplayName)
		if errors.Is(err, usecases.ErrDeviceNotFound) {
			http.Error(w, "device not found", http.StatusNotFound)
			return
		}

		if err != nil {
			http.Error(w, "failed to update device", http.StatusInternalServerError)
			return
		}

		// Get the updated device to return it
		device, err := c.service.GetDevice(r.Context(), domain.ID(id))
		if err != nil {
			http.Error(w, "failed to get updated device", http.StatusInternalServerError)
			return
		}

		response := internal.ToDeviceResponse(device)
		httpserver.ReplyJSONResponse(w, http.StatusOK, response)
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

		builder := domain.NewCommandBuilder()

		cmdSequence := domain.CommandSequence{
			Commands: make([]domain.Command, len(body.Sequence)),
		}
		for i, item := range body.Sequence {
			cmd, err := builder.
				WithDevice(domain.Device{ID: domain.ID(id)}).
				WithPayload(domain.CommandPayload{
					Index: domain.Index(item.Index),
					Value: domain.CommandValue(item.Value),
				}).
				WithPriority(domain.CommandPriority(body.Priority)).
				WithDispatchAfter(utils.Time{Time: time.Now().Add(time.Duration(item.WaitFor))}).
				Build()
			if err != nil {
				slog.Error("build command", slog.String("error", err.Error()))
				http.Error(w, "queue command failed", http.StatusInternalServerError)
				return
			}
			cmdSequence.Commands[i] = cmd
		}

		err = c.service.QueueCommandSequence(r.Context(), cmdSequence)
		if err != nil {
			slog.Error("queue command sequence", slog.String("error", err.Error()))
			http.Error(w, "queue command sequence failed", http.StatusInternalServerError)
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusCreated, nil)
	}
}
