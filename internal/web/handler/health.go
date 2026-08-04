package handler

import (
	"net/http"
	"time"

	"github.com/wavefnd/wave-platform/internal/health"
	"github.com/wavefnd/wave-platform/internal/xmlcodec"
)

type HealthHandler struct {
	DatabaseCheck func() error
}

func (handler HealthHandler) Live(writer http.ResponseWriter, request *http.Request) {
	response := health.Status{
		Status:    "healthy",
		Ready:     true,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Database:  "not-checked",
	}

	if err := xmlcodec.Write(writer, http.StatusOK, response); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func (handler HealthHandler) Ready(writer http.ResponseWriter, request *http.Request) {
	response := health.Status{
		Status:    "healthy",
		Ready:     true,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Database:  "available",
	}

	status := http.StatusOK
	if handler.DatabaseCheck == nil || handler.DatabaseCheck() != nil {
		response.Status = "degraded"
		response.Ready = false
		response.Database = "unavailable"
		status = http.StatusServiceUnavailable
	}

	if err := xmlcodec.Write(writer, status, response); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}
