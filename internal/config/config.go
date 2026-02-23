package config

import (
	"log/slog"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Endpoint struct {
	Name       string       `yaml:"name"`
	URL        string       `yaml:"url"`
	Method     string       `yaml:"method"`
	Iterations []Iterations `yaml:"iterations,omitempty"`
}

type Iterations struct {
	Name string `yaml:"name"`
	ID   string `yaml:"id"`
}

type Config struct {
	APIKeyHeaderName string     `yaml:"api-key-header-name"`
	APIKey           string     `yaml:"api-key-value"`
	UserAgent        string     `yaml:"user-agent"`
	TimeoutSeconds   int        `yaml:"timeout-seconds"`
	Schedule         string     `yaml:"schedule"`
	Endpoints        []Endpoint `yaml:"endpoints"`
	LogLevel         string     `yaml:"log-level"`
	ParsedLogLevel   slog.Level `yaml:"-"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Expand endpoints with iterations
	var expandedEndpoints []Endpoint
	for _, ep := range config.Endpoints {
		if len(ep.Iterations) > 0 {
			for _, iteration := range ep.Iterations {
				expandedEndpoints = append(expandedEndpoints, Endpoint{
					Name:   strings.Replace(ep.Name, "{name}", iteration.Name, 1),
					URL:    strings.Replace(ep.URL, "{id}", iteration.ID, 1),
					Method: ep.Method,
				})
			}
		} else {
			expandedEndpoints = append(expandedEndpoints, ep)
		}
	}
	config.Endpoints = expandedEndpoints

	level, ok := parseLogLevel(config.LogLevel)
	if !ok && config.LogLevel != "" {
		slog.Warn("Invalid log level, defaulting to info", "log-level", config.LogLevel)
	}
	config.ParsedLogLevel = level

	return &config, nil
}

// ParseLogLevel maps a config string to slog.Level.
func parseLogLevel(value string) (slog.Level, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug, true
	case "info", "":
		return slog.LevelInfo, value != ""
	case "warn", "warning":
		return slog.LevelWarn, true
	case "error":
		return slog.LevelError, true
	default:
		return slog.LevelInfo, false
	}
}
