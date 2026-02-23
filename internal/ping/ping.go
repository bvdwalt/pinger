package ping

import (
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/bvdwalt/pinger/internal/config"
)

func Execute(client *http.Client, ep config.Endpoint, apiKeyHeaderName, apiKey, userAgent string) {
	start := time.Now()

	req, err := http.NewRequest(ep.Method, ep.URL, nil)
	if err != nil {
		slog.Error("Failed to create request", "endpoint", ep.Name, "error", err)
		return
	}

	// Add API key header if provided
	if apiKey != "" {
		req.Header.Set(apiKeyHeaderName, apiKey)
	}

	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Failed request", "endpoint", ep.Name, "error", err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			slog.Error("Failed to close response body", "endpoint", ep.Name, "error", err)
		}
	}(resp.Body)

	duration := time.Since(start)
	slog.Info("Request complete", "endpoint", ep.Name, "duration", duration, "status", resp.StatusCode)
}
