package ping

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bvdwalt/pinger/internal/config"
)

// Helper to create a test client
func testClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}

// Helper to create a test endpoint
func testEndpoint(url, method string) config.Endpoint {
	return config.Endpoint{Name: "Test", URL: url, Method: method}
}

// TestPingEndpointAPIKeyHandling covers API key scenarios with table-driven tests
func TestPingEndpointAPIKeyHandling(t *testing.T) {
	tests := []struct {
		name          string
		headerName    string
		apiKey        string
		userAgent     string
		shouldExist   bool
		expectedValue string
	}{
		{"With API key", "X-API-Key", "secret-123", "Pinger", true, "secret-123"},
		{"Without API key", "X-API-Key", "", "Pinger", false, ""},
		{"Custom header", "Authorization", "Bearer token", "Pinger", true, "Bearer token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var headerValue string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				headerValue = r.Header.Get(tt.headerName)
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			Execute(testClient(5*time.Millisecond), testEndpoint(server.URL, "GET"), tt.headerName, tt.apiKey, tt.userAgent)

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

			Execute(testClient(5*time.Millisecond), testEndpoint(server.URL, method), "", "", "Pinger")
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
			Execute(testClient(5*time.Millisecond), testEndpoint(server.URL, "GET"), "", "", "Pinger")
		})
	}
}

// TestPingEndpointErrorCases verifies graceful error handling
func TestPingEndpointErrorCases(t *testing.T) {
	closedAddr := func(t *testing.T) string {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("failed to create listener: %v", err)
		}
		addr := listener.Addr().String()
		if err := listener.Close(); err != nil {
			t.Fatalf("failed to close listener: %v", err)
		}
		return "http://" + addr
	}

	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer slowServer.Close()

	tests := []struct {
		name    string
		url     string
		timeout time.Duration
	}{
		{"Invalid URL", "://invalid", 5 * time.Millisecond},
		{"Connection refused", closedAddr(t), 10 * time.Millisecond},
		{"Timeout", slowServer.URL, 1 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should handle errors gracefully without panicking
			Execute(testClient(tt.timeout), testEndpoint(tt.url, "GET"), "", "", "Pinger")
		})
	}
}
