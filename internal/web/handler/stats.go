package handler

import (
	"net/http"

	"github.com/wavefnd/wave-platform/internal/platformstats"
	"github.com/wavefnd/wave-platform/internal/xmlcodec"
)

type StatsHandler struct {
	Service *platformstats.Service
}

func (handler StatsHandler) Get(writer http.ResponseWriter, request *http.Request) {
	if handler.Service == nil {
		http.Error(writer, "platform statistics are unavailable", http.StatusServiceUnavailable)
		return
	}
	snapshot, err := handler.Service.Snapshot(request.Context())
	if err != nil {
		http.Error(writer, "failed to load platform statistics", http.StatusInternalServerError)
		return
	}
	if err := xmlcodec.Write(writer, http.StatusOK, snapshot); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}
