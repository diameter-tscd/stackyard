package infrastructure

import (
	"net/http"
	"test-go/config"
	"time"
)

type HttpManager struct {
	Services []config.ExternalService
	Client   *http.Client
}

func NewHttpManager(cfg config.ExternalConfig) *HttpManager {
	return &HttpManager{
		Services: cfg.Services,
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (h *HttpManager) GetStatus() []map[string]interface{} {
	results := []map[string]interface{}{}

	for _, svc := range h.Services {
		start := time.Now()
		resp, err := h.Client.Get(svc.URL)
		latency := time.Since(start).Milliseconds()

		status := "down"
		statusCode := 0
		if err == nil {
			statusCode = resp.StatusCode
			resp.Body.Close()
			if statusCode >= 200 && statusCode < 300 {
				status = "up"
			} else {
				status = "degraded"
			}
		}

		results = append(results, map[string]interface{}{
			"name":        svc.Name,
			"url":         svc.URL,
			"status":      status,
			"status_code": statusCode,
			"latency_ms":  latency,
		})
	}

	return results
}
