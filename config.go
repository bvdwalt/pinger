package main

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Endpoint struct {
	Name   string         `yaml:"name"`
	URL    string         `yaml:"url"`
	Method string         `yaml:"method"`
	Iterations   []Iterations `yaml:"iterations,omitempty"`
}

type Iterations struct {
	Name string `yaml:"name"`
	ID   string `yaml:"id"`
}

type Config struct {
	APIKeyHeaderName string     `yaml:"api-key-header-name"`
	APIKey           string     `yaml:"api-key-value"`
	TimeoutSeconds   int        `yaml:"timeout-seconds"`
	Schedule         string     `yaml:"schedule"`
	Endpoints        []Endpoint `yaml:"endpoints"`
}

func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Expand endpoints with iterations
	expandedEndpoints := []Endpoint{}
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

	return &config, nil
}
