package main

import (
	"os"
	"testing"
)

func TestLoadConfigExample(t *testing.T) {
	config, err := loadConfig("config-example.yaml")
	if err != nil {
		t.Fatalf("Failed to load config-example.yaml: %v", err)
	}

	if config == nil {
		t.Fatal("Config is nil")
	}

	if config.Schedule == "" {
		t.Error("Schedule should not be empty")
	}

	if config.TimeoutSeconds <= 0 {
		t.Error("TimeoutSeconds should be positive")
	}

	if config.APIKeyHeaderName == "" {
		t.Error("APIKeyHeaderName should not be empty")
	}

	if len(config.Endpoints) == 0 {
		t.Error("Should have at least one endpoint")
	}

	// Verify each endpoint has required fields
	for i, ep := range config.Endpoints {
		if ep.Name == "" {
			t.Errorf("Endpoint %d: Name should not be empty", i)
		}
		if ep.URL == "" {
			t.Errorf("Endpoint %d: URL should not be empty", i)
		}
		if ep.Method == "" {
			t.Errorf("Endpoint %d: Method should not be empty", i)
		}
	}
}

func TestLoadConfig(t *testing.T) {
	config, err := loadConfig("config.yaml")
	if err != nil {
		t.Fatalf("Failed to load config.yaml: %v", err)
	}

	if config == nil {
		t.Fatal("Config is nil")
	}

	if len(config.Endpoints) == 0 {
		t.Error("Should have at least one endpoint")
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := loadConfig("nonexistent.yaml")
	if err == nil {
		t.Error("Expected error when loading non-existent file")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	// Create a temporary invalid YAML file
	tmpFile := "test-invalid.yaml"
	err := os.WriteFile(tmpFile, []byte("invalid: yaml: content: ["), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	_, err = loadConfig(tmpFile)
	if err == nil {
		t.Error("Expected error when loading invalid YAML")
	}
}

func TestEndpointExpansion(t *testing.T) {
	// Create a temporary config with iterations
	tmpFile := "test-expansion.yaml"
	content := `schedule: "*/5 * * * *"
timeout-seconds: 30
api-key-header-name: "x-api-key"
api-key-value: "test-key"
endpoints:
  - name: Test ({name})
    url: https://example.com/{id}
    method: GET
    iterations:
      - name: Org1
        id: id1
      - name: Org2
        id: id2
`
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	config, err := loadConfig(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Should have 2 endpoints after expansion
	if len(config.Endpoints) != 2 {
		t.Errorf("Expected 2 endpoints after expansion, got %d", len(config.Endpoints))
	}

	// Verify names were replaced
	if config.Endpoints[0].Name != "Test (Org1)" {
		t.Errorf("Expected 'Test (Org1)', got '%s'", config.Endpoints[0].Name)
	}
	if config.Endpoints[1].Name != "Test (Org2)" {
		t.Errorf("Expected 'Test (Org2)', got '%s'", config.Endpoints[1].Name)
	}

	// Verify URLs were replaced
	if config.Endpoints[0].URL != "https://example.com/id1" {
		t.Errorf("Expected 'https://example.com/id1', got '%s'", config.Endpoints[0].URL)
	}
	if config.Endpoints[1].URL != "https://example.com/id2" {
		t.Errorf("Expected 'https://example.com/id2', got '%s'", config.Endpoints[1].URL)
	}
}
