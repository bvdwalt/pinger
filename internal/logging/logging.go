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
	reqForDump := req.Clone(req.Context())
	reqForDump.Header = redactHeader(req.Header)
	requestDump, _ := httputil.DumpRequestOut(reqForDump, false)
	slog.Debug("[HTTP Request]", "request", string(requestDump))

	// Perform the actual request
	resp, err := lt.transport.RoundTrip(req)

	if err != nil {
		slog.Error("[HTTP Error]", "method", req.Method, "url", req.URL, "error", err)
		return resp, err
	}

	respForDump := *resp
	respForDump.Header = redactHeader(resp.Header)
	responseDump, _ := httputil.DumpResponse(&respForDump, false)
	slog.Debug("[HTTP Response]", "response", string(responseDump))

	return resp, err
}

// redactHeader returns a copy of h with every header value partially redacted.
func redactHeader(h http.Header) http.Header {
	redacted := make(http.Header, len(h))
	for name, values := range h {
		redactedValues := make([]string, len(values))
		for i, v := range values {
			redactedValues[i] = redactValue(v)
		}
		redacted[name] = redactedValues
	}
	return redacted
}

// redactValue keeps a short prefix of v and masks the rest.
func redactValue(v string) string {
	const prefixLen = 4
	if len(v) <= prefixLen {
		return "***"
	}
	return v[:prefixLen] + "***"
}
