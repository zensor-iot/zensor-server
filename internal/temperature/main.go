package temperature

import "log/slog"

type temperatureHandler struct {
	lowLevelCounter uint
}

func CreateHandler() *temperatureHandler {
	return &temperatureHandler{
		lowLevelCounter: 0,
	}
}

func (h *temperatureHandler) Push(msg string) {
	slog.Info("message receive in temperature analyzer", "msg", msg)

	if h.isCold() {
		slog.Info("is cold")
	}
}

func (h *temperatureHandler) isCold() bool {
	return false
}
