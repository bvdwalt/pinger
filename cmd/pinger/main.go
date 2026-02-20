package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bvdwalt/pinger/internal/config"
	"github.com/bvdwalt/pinger/internal/logging"
	"github.com/bvdwalt/pinger/internal/ping"
	"github.com/robfig/cron/v3"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	client := &http.Client{
		Timeout:   time.Duration(cfg.TimeoutSeconds) * time.Second,
		Transport: logging.NewLoggingTransport(cfg.EnableHttpLogging),
	}

	c := cron.New()

	log.Printf("Scheduling Pinger with cron: '%s'", cfg.Schedule)
	for _, endpoint := range cfg.Endpoints {
		ep := endpoint

		// Run immediately on startup
		pingFunc := func() {
			ping.Execute(client, ep, cfg.APIKeyHeaderName, cfg.APIKey, cfg.UserAgent)
		}
		go pingFunc()

		_, err := c.AddFunc(cfg.Schedule, func() {
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
