package ping

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/bvdwalt/pinger/internal/config"
)

func Execute(client *http.Client, ep config.Endpoint, apiKeyHeaderName, apiKey, userAgent string) {
	start := time.Now()

	req, err := http.NewRequest(ep.Method, ep.URL, nil)
	if err != nil {
		log.Printf("[%s] Failed to create request: %v", ep.Name, err)
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
		log.Printf("[%s] Failed with %v", ep.Name, err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("[%s] Failed to close response body: %v", ep.Name, err)
		}
	}(resp.Body)

	duration := time.Since(start)
	log.Printf("[%-30s] %-6s Status: %-9s Duration: %v",
		ep.Name, ep.Method, resp.Status, duration)
}
