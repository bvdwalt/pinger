package main

import (
	"log"
	"log/slog"
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
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.ParsedLogLevel})
	slog.SetDefault(slog.New(handler))

	client := &http.Client{
		Timeout:   time.Duration(cfg.TimeoutSeconds) * time.Second,
		Transport: logging.NewLoggingTransport(),
	}

	c := cron.New()

	slog.Info("Scheduling pinger", "cron", cfg.Schedule)
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
			slog.Error("Failed to schedule", "endpoint", ep, "err", err)
			continue
		}
		slog.Info("Scheduled", "endpoint", ep.Name)
	}

	c.Start()
	log.Println("Pinger started. Press Ctrl+C to exit.")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	ctx := c.Stop()
	<-ctx.Done()
	log.Println("Shutdown complete")
}
