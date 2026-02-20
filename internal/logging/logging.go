package logging

import (
	"log"
	"net/http"
	"net/http/httputil"
)

// Transport wraps http.RoundTripper to log HTTP requests and responses
type Transport struct {
	transport     http.RoundTripper
	enableLogging bool
}

// NewLoggingTransport creates a new LoggingTransport with the default HTTP transport
func NewLoggingTransport(enableLogging bool) *Transport {
	return &Transport{
		transport:     http.DefaultTransport,
		enableLogging: enableLogging,
	}
}

// RoundTrip implements the http.RoundTripper interface and logs requests/responses
func (lt *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if lt.enableLogging {
		requestDump, _ := httputil.DumpRequestOut(req, false)
		log.Printf("[HTTP Request]\n%s", string(requestDump))
	}

	// Perform the actual request
	resp, err := lt.transport.RoundTrip(req)

	if err != nil {
		if lt.enableLogging {
			log.Printf("[HTTP Error] %s %s: %v", req.Method, req.URL, err)
		}
		return resp, err
	}

	if lt.enableLogging {
		responseDump, _ := httputil.DumpResponse(resp, false)
		log.Printf("[HTTP Response]\n%s", string(responseDump))
	}

	return resp, err
}
