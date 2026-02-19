package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
)

func main() {
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	client := &http.Client{
		Timeout:   time.Duration(config.TimeoutSeconds) * time.Second,
		Transport: NewLoggingTransport(config.EnableHttpLogging),
	}

	c := cron.New()

	log.Printf("Scheduling Pinger with cron: '%s'", config.Schedule)
	for _, endpoint := range config.Endpoints {
		ep := endpoint

		// Run immediately on startup
		pingFunc := func() {
			pingEndpoint(client, ep, config.APIKeyHeaderName, config.APIKey, config.UserAgent)
		}
		go pingFunc()

		_, err := c.AddFunc(config.Schedule, func() {
			pingFunc()
		})
		if err != nil {
			log.Printf("Failed to schedule %s: %v", ep.Name, err)
			continue
		}
		log.Printf("Scheduled: %s", ep.Name)
	}

	c.Start()
	log.Println("Pinger started. Press Ctrl+C to exit.")
	log.Println("...")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	ctx := c.Stop()
	<-ctx.Done()
	log.Println("Shutdown complete")
}

func pingEndpoint(client *http.Client, ep Endpoint, apiKeyHeaderName, apiKey, userAgent string) {
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
