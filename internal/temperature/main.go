package temperature

import (
	"zensor-server/internal/logger"
)

type temperatureHandler struct {
	l               logger.Logger
	lowLevelCounter uint
}

func CreateHandler() *temperatureHandler {
	return &temperatureHandler{
		lowLevelCounter: 0,
	}
}

func (h *temperatureHandler) Push(msg string) {
	logger.Info("message receive in temperature analyzer", "msg", msg)

	if h.isCold() {
		logger.Info("is cold")
	}
}

func (h *temperatureHandler) isCold() bool {
	return false
}
