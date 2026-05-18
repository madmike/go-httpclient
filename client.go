package httpclient

import (
	"net"
	"net/http"
	"time"
)

// NewClient creates a production-ready HTTP client with:
// - 10 second request timeout
// - Connection pooling (keep-alive)
// - TCP keep-alive
// Suitable for inter-service communication.
func NewClient() *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     32,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}
}

// Global default client for inter-service calls
var defaultClient = NewClient()

// Default returns the global HTTP client for inter-service calls.
func Default() *http.Client {
	return defaultClient
}
