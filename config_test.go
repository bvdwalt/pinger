package main

import (
	"os"
	"testing"
)

// createTempTestFile creates a temporary test file with the given content
// and automatically schedules cleanup using t.Cleanup
func createTempTestFile(t *testing.T, filename, content string) {
	t.Helper()
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	t.Cleanup(func() {
		err := os.Remove(filename)
		if err != nil {
			t.Logf("Failed to remove test file: %v", err)
		}
	})
}

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
	createTempTestFile(t, tmpFile, "invalid: yaml: content: [")

	_, err := loadConfig(tmpFile)
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
	createTempTestFile(t, tmpFile, content)

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

func TestConfigWithNoEndpoints(t *testing.T) {
	tmpFile := "test-no-endpoints.yaml"
	content := `schedule: "*/5 * * * *"
timeout-seconds: 30
api-key-header-name: "x-api-key"
api-key-value: "test-key"
endpoints: []
`
	createTempTestFile(t, tmpFile, content)

	config, err := loadConfig(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(config.Endpoints) != 0 {
		t.Errorf("Expected 0 endpoints, got %d", len(config.Endpoints))
	}
}

func TestEndpointExpansionWithMultipleIterations(t *testing.T) {
	tmpFile := "test-multiple-expansions.yaml"
	content := `schedule: "*/5 * * * *"
timeout-seconds: 30
api-key-header-name: "x-api-key"
api-key-value: "test-key"
endpoints:
  - name: Service {name}
    url: https://api.example.com/{id}/health
    method: POST
    iterations:
      - name: Alpha
        id: alpha-001
      - name: Beta
        id: beta-002
      - name: Gamma
        id: gamma-003
`
	createTempTestFile(t, tmpFile, content)

	config, err := loadConfig(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(config.Endpoints) != 3 {
		t.Errorf("Expected 3 endpoints after expansion, got %d", len(config.Endpoints))
	}

	expectedNames := []string{"Service Alpha", "Service Beta", "Service Gamma"}
	for i, expectedName := range expectedNames {
		if config.Endpoints[i].Name != expectedName {
			t.Errorf("Endpoint %d: expected name '%s', got '%s'", i, expectedName, config.Endpoints[i].Name)
		}
	}
}

func TestMixedEndpointsWithAndWithoutIterations(t *testing.T) {
	tmpFile := "test-mixed-endpoints.yaml"
	content := `schedule: "*/5 * * * *"
timeout-seconds: 30
api-key-header-name: "x-api-key"
api-key-value: "test-key"
endpoints:
  - name: Static Endpoint
    url: https://example.com/health
    method: GET
  - name: Dynamic {name}
    url: https://example.com/{id}/status
    method: POST
    iterations:
      - name: Instance1
        id: inst1
      - name: Instance2
        id: inst2
`
	createTempTestFile(t, tmpFile, content)

	config, err := loadConfig(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Should have 3 total endpoints: 1 static + 2 expanded
	if len(config.Endpoints) != 3 {
		t.Errorf("Expected 3 endpoints, got %d", len(config.Endpoints))
	}

	if config.Endpoints[0].Name != "Static Endpoint" {
		t.Errorf("Expected 'Static Endpoint', got '%s'", config.Endpoints[0].Name)
	}
	if config.Endpoints[1].Name != "Dynamic Instance1" {
		t.Errorf("Expected 'Dynamic Instance1', got '%s'", config.Endpoints[1].Name)
	}
	if config.Endpoints[2].Name != "Dynamic Instance2" {
		t.Errorf("Expected 'Dynamic Instance2', got '%s'", config.Endpoints[2].Name)
	}
}

func TestConfigWithSpecialCharacters(t *testing.T) {
	tmpFile := "test-special-chars.yaml"
	content := `schedule: "*/5 * * * *"
timeout-seconds: 30
api-key-header-name: "x-api-key"
api-key-value: "test-key-!@#$%"
endpoints:
  - name: API with special chars
    url: "https://example.com/path?query=value&other=123"
    method: GET
`
	createTempTestFile(t, tmpFile, content)

	config, err := loadConfig(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.APIKey != "test-key-!@#$%" {
		t.Errorf("Expected APIKey with special chars, got '%s'", config.APIKey)
	}
	if config.Endpoints[0].URL != "https://example.com/path?query=value&other=123" {
		t.Errorf("Expected URL with special chars, got '%s'", config.Endpoints[0].URL)
	}
}

func TestConfigDefaultValues(t *testing.T) {
	tmpFile := "test-defaults.yaml"
	content := `schedule: "0 0 * * *"
endpoints:
  - name: Endpoint
    url: https://example.com
    method: GET
`
	createTempTestFile(t, tmpFile, content)

	config, err := loadConfig(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Fields not specified should be empty/zero
	if config.APIKeyHeaderName != "" {
		t.Errorf("Expected empty APIKeyHeaderName, got '%s'", config.APIKeyHeaderName)
	}
	if config.APIKey != "" {
		t.Errorf("Expected empty APIKey, got '%s'", config.APIKey)
	}
	if config.TimeoutSeconds != 0 {
		t.Errorf("Expected 0 TimeoutSeconds, got %d", config.TimeoutSeconds)
	}
}
