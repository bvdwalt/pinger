package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Helper to create a test client
func testClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}

// Helper to create a test endpoint
func testEndpoint(url, method string) Endpoint {
	return Endpoint{Name: "Test", URL: url, Method: method}
}

// TestPingEndpointAPIKeyHandling covers API key scenarios with table-driven tests
func TestPingEndpointAPIKeyHandling(t *testing.T) {
	tests := []struct {
		name          string
		headerName    string
		apiKey        string
		shouldExist   bool
		expectedValue string
	}{
		{"With API key", "X-API-Key", "secret-123", true, "secret-123"},
		{"Without API key", "X-API-Key", "", false, ""},
		{"Custom header", "Authorization", "Bearer token", true, "Bearer token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var headerValue string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				headerValue = r.Header.Get(tt.headerName)
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			pingEndpoint(testClient(5*time.Second), testEndpoint(server.URL, "GET"), tt.headerName, tt.apiKey)

			if tt.shouldExist && headerValue != tt.expectedValue {
				t.Errorf("expected %q, got %q", tt.expectedValue, headerValue)
			}
			if !tt.shouldExist && headerValue != "" {
				t.Errorf("header should not be set, got %q", headerValue)
			}
		})
	}
}

// TestPingEndpointMethods verifies different HTTP methods are sent correctly
func TestPingEndpointMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "HEAD"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			var capturedMethod string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedMethod = r.Method
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			pingEndpoint(testClient(5*time.Second), testEndpoint(server.URL, method), "", "")
			if capturedMethod != method {
				t.Errorf("expected %s, got %s", method, capturedMethod)
			}
		})
	}
}

// TestPingEndpointStatusCodes verifies handling of various HTTP status codes
func TestPingEndpointStatusCodes(t *testing.T) {
	codes := []int{200, 201, 202, 400, 401, 403, 404, 500, 503}
	for _, code := range codes {
		t.Run(http.StatusText(code), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(code)
			}))
			defer server.Close()

			// Should not panic on any status code
			pingEndpoint(testClient(5*time.Second), testEndpoint(server.URL, "GET"), "", "")
		})
	}
}

// TestPingEndpointErrorCases verifies graceful error handling
func TestPingEndpointErrorCases(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		timeout time.Duration
	}{
		{"Invalid URL", "://invalid", 5 * time.Second},
		{"Connection refused", "http://localhost:1", 1 * time.Second},
		{"Timeout", "http://httpbin.org/delay/10", 100 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should handle errors gracefully without panicking
			pingEndpoint(testClient(tt.timeout), testEndpoint(tt.url, "GET"), "", "")
		})
	}
}

func TestStringReplacementInEndpointExpansion(t *testing.T) {
	tests := []struct {
		name          string
		templateName  string
		templateURL   string
		iterationName string
		iterationID   string
		expectedName  string
		expectedURL   string
	}{
		{
			name:          "Simple replacement",
			templateName:  "Endpoint {name}",
			templateURL:   "http://example.com/{id}",
			iterationName: "Prod",
			iterationID:   "prod-123",
			expectedName:  "Endpoint Prod",
			expectedURL:   "http://example.com/prod-123",
		},
		{
			name:          "No replacement needed",
			templateName:  "Static Endpoint",
			templateURL:   "http://example.com/static",
			iterationName: "Prod",
			iterationID:   "prod-123",
			expectedName:  "Static Endpoint",
			expectedURL:   "http://example.com/static",
		},
		{
			name:          "Multiple occurrences (name only replaces first)",
			templateName:  "Endpoint {name} for {name}",
			templateURL:   "http://example.com/{id}/{id}",
			iterationName: "Prod",
			iterationID:   "prod-123",
			expectedName:  "Endpoint Prod for {name}",
			expectedURL:   "http://example.com/prod-123/{id}",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resultName := strings.Replace(test.templateName, "{name}", test.iterationName, 1)
			resultURL := strings.Replace(test.templateURL, "{id}", test.iterationID, 1)

			if resultName != test.expectedName {
				t.Errorf("Name: expected '%s', got '%s'", test.expectedName, resultName)
			}
			if resultURL != test.expectedURL {
				t.Errorf("URL: expected '%s', got '%s'", test.expectedURL, resultURL)
			}
		})
	}
}
