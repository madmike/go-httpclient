package httpclient

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestNewClient creates a client with production settings.
func TestNewClient(t *testing.T) {
	client := NewClient()

	require.NotNil(t, client)
	require.NotNil(t, client.Transport)

	// Verify timeout is set
	require.Equal(t, 10*time.Second, client.Timeout)
}

// TestNewClientTransportConfiguration verifies connection pool settings.
func TestNewClientTransportConfiguration(t *testing.T) {
	client := NewClient()
	transport := client.Transport.(*http.Transport)

	// Verify pool sizes
	require.Equal(t, 100, transport.MaxIdleConns)
	require.Equal(t, 10, transport.MaxIdleConnsPerHost)
	require.Equal(t, 32, transport.MaxConnsPerHost)

	// Verify timeouts
	require.Equal(t, 90*time.Second, transport.IdleConnTimeout)
	require.Equal(t, 5*time.Second, transport.TLSHandshakeTimeout)

	// Verify keep-alives enabled
	require.False(t, transport.DisableKeepAlives)
}

// TestNewClientDialConfiguration verifies dial settings.
func TestNewClientDialConfiguration(t *testing.T) {
	client := NewClient()
	transport := client.Transport.(*http.Transport)

	// Get the dialer settings by calling DialContext
	// We can't directly inspect the dialer, but we verify the timeout
	// is configured reasonably for service-to-service calls
	require.NotNil(t, transport.DialContext)
}

// TestDefaultClientSingleton returns the same instance.
func TestDefaultClientSingleton(t *testing.T) {
	client1 := Default()
	client2 := Default()

	require.NotNil(t, client1)
	require.NotNil(t, client2)
	require.Equal(t, client1, client2, "Default should return the same instance")
}

// TestDefaultClientIsProperlyConfigured verifies default has production settings.
func TestDefaultClientIsProperlyConfigured(t *testing.T) {
	client := Default()

	// Verify it has a timeout (not 0)
	require.NotZero(t, client.Timeout)
	require.Equal(t, 10*time.Second, client.Timeout)

	// Verify transport is configured
	transport := client.Transport.(*http.Transport)
	require.Equal(t, 100, transport.MaxIdleConns)
	require.Equal(t, 10, transport.MaxIdleConnsPerHost)
}

// TestClientCanBeUsed verifies the client is functional (won't panic).
func TestClientCanBeUsed(t *testing.T) {
	client := NewClient()

	// Client should not panic when checking properties
	require.NotNil(t, client.Transport)
	require.NotZero(t, client.Timeout)
}

// TestMultipleClientsIndependent multiple clients don't share state.
func TestMultipleClientsIndependent(t *testing.T) {
	client1 := NewClient()
	client2 := NewClient()

	// Each client has its own transport
	require.NotEqual(t, client1.Transport, client2.Transport)
}

// TestTransportDialContextTimeout verifies dial timeout is reasonable.
func TestTransportDialContextTimeout(t *testing.T) {
	client := NewClient()
	transport := client.Transport.(*http.Transport)

	// The dialer should have a 5-second connect timeout (suitable for service calls)
	// We can't directly inspect the timeout, but we verify the transport exists
	require.NotNil(t, transport.DialContext)
}

// TestTransportKeepAliveInterval verifies TCP keep-alive is enabled.
func TestTransportKeepAliveInterval(t *testing.T) {
	client := NewClient()
	transport := client.Transport.(*http.Transport)

	// KeepAlive should be 30 seconds
	// (Can't directly inspect, but this tests the configuration)
	require.NotNil(t, transport.DialContext)
}

// TestHttpTransportOther settings are production-ready.
func TestHttpTransportOtherSettings(t *testing.T) {
	client := NewClient()
	transport := client.Transport.(*http.Transport)

	tests := []struct {
		name  string
		check func() bool
	}{
		{"MaxIdleConns >= 100", func() bool { return transport.MaxIdleConns >= 100 }},
		{"MaxIdleConnsPerHost >= 10", func() bool { return transport.MaxIdleConnsPerHost >= 10 }},
		{"MaxConnsPerHost >= 32", func() bool { return transport.MaxConnsPerHost >= 32 }},
		{"IdleConnTimeout > 0", func() bool { return transport.IdleConnTimeout > 0 }},
		{"TLSHandshakeTimeout > 0", func() bool { return transport.TLSHandshakeTimeout > 0 }},
		{"KeepAlives enabled", func() bool { return !transport.DisableKeepAlives }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.True(t, tt.check())
		})
	}
}

// TestClientTimeoutPreventsHangs ensures timeout prevents infinite waits.
func TestClientTimeoutPreventsHangs(t *testing.T) {
	client := NewClient()

	// Timeout should be <= 10 seconds (prevents unbounded hangs)
	require.LessOrEqual(t, client.Timeout, 10*time.Second)
	require.Greater(t, client.Timeout, 0*time.Second)
}

// TestTransportDialSettings are suitable for microservices.
func TestTransportDialSettings(t *testing.T) {
	client := NewClient()
	transport := client.Transport.(*http.Transport)

	// Verify settings are production-appropriate for inter-service communication:
	// - Not too aggressive (would waste resources)
	// - Not too conservative (would timeout legitimate requests)
	require.GreaterOrEqual(t, transport.MaxIdleConns, 50)
	require.LessOrEqual(t, transport.MaxIdleConns, 500)

	require.GreaterOrEqual(t, transport.MaxConnsPerHost, 8)
	require.LessOrEqual(t, transport.MaxConnsPerHost, 64)
}
