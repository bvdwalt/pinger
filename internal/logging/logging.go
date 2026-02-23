package logging

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
)

// Transport wraps http.RoundTripper to log HTTP requests and responses
type Transport struct {
	transport http.RoundTripper
}

// NewLoggingTransport creates a new LoggingTransport with the default HTTP transport
func NewLoggingTransport() *Transport {
	return &Transport{
		transport: http.DefaultTransport,
	}
}

// RoundTrip implements the http.RoundTripper interface and logs requests/responses
func (lt *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	requestDump, _ := httputil.DumpRequestOut(req, false)
	slog.Debug("[HTTP Request]", "request", string(requestDump))

	// Perform the actual request
	resp, err := lt.transport.RoundTrip(req)

	if err != nil {
		slog.Error("[HTTP Error]", "method", req.Method, "url", req.URL, "error", err)
		return resp, err
	}

	responseDump, _ := httputil.DumpResponse(resp, false)
	slog.Debug("[HTTP Response]", "response", string(responseDump))

	return resp, err
}
