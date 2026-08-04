package handler

import (
	"encoding/xml"
	"net/http"
)

type PlatformHandler struct {
	Environment string
	Version     string
}

type PlatformResponse struct {
	XMLName xml.Name `xml:"https://wave-lang.dev/ns/platform/api/v1 platform"`

	Name        string `xml:"name"`
	Status      string `xml:"status"`
	Version     string `xml:"version"`
	Environment string `xml:"environment"`
}

func (handler PlatformHandler) Status(
	writer http.ResponseWriter,
	request *http.Request,
) {
	response := PlatformResponse{
		Name:        "Wave Platform",
		Status:      "online",
		Version:     handler.Version,
		Environment: handler.Environment,
	}

	data, err := xml.MarshalIndent(response, "", "    ")
	if err != nil {
		http.Error(writer, "failed to encode response", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/xml; charset=utf-8")
	writer.WriteHeader(http.StatusOK)

	_, _ = writer.Write([]byte(xml.Header))
	_, _ = writer.Write(data)
}
