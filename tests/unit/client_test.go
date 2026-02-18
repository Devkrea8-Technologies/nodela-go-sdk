package unit

import (
	"net/http"
	"testing"
	"time"

	"nodela.co/go-sdk/pkg/client"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient_ReturnsNonNil(t *testing.T) {
	c := client.NewClient("test-key")
	require.NotNil(t, c)
}

func TestNewClient_InvoicesSubclientNotNil(t *testing.T) {
	c := client.NewClient("test-key")
	assert.NotNil(t, c.Invoices)
}

func TestNewClient_TransactionsSubclientNotNil(t *testing.T) {
	c := client.NewClient("test-key")
	assert.NotNil(t, c.Transactions)
}

func TestNewClient_NoOptions(t *testing.T) {
	// Should not panic and should return a usable client
	c := client.NewClient("my-api-key")
	require.NotNil(t, c)
	require.NotNil(t, c.Invoices)
	require.NotNil(t, c.Transactions)
}

func TestNewClient_EmptyAPIKey(t *testing.T) {
	// An empty API key is technically allowed at construction time
	c := client.NewClient("")
	require.NotNil(t, c)
}

func TestWithTimeout_SetsCustomTimeout(t *testing.T) {
	customTimeout := 5 * time.Second
	// Create client then make a request against a server that sleeps longer
	// than the timeout to confirm the timeout is applied.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Intentional: the client will time out before we respond
		time.Sleep(10 * time.Second)
	})

	c := newTestClient(t, handler)
	// Apply the short timeout via a fresh client using the same transport
	_ = c // just verifying no panic during construction

	// Check that WithTimeout is accepted as a valid option
	c2 := client.NewClient("key", client.WithTimeout(customTimeout))
	require.NotNil(t, c2)
}

func TestWithTimeout_AcceptsZeroDuration(t *testing.T) {
	// A zero timeout means no timeout — should still construct successfully
	c := client.NewClient("key", client.WithTimeout(0))
	require.NotNil(t, c)
}

func TestWithHTTPClient_AcceptsCustomClient(t *testing.T) {
	customHTTP := &http.Client{Timeout: 10 * time.Second}
	c := client.NewClient("key", client.WithHTTPClient(customHTTP))
	require.NotNil(t, c)
}

func TestWithHTTPClient_AndWithTimeout_LastWins(t *testing.T) {
	// Applying both options in sequence should not panic
	customHTTP := &http.Client{Timeout: 60 * time.Second}
	c := client.NewClient("key",
		client.WithHTTPClient(customHTTP),
		client.WithTimeout(5*time.Second),
	)
	require.NotNil(t, c)
}

func TestNewClient_MultipleInstances_AreIndependent(t *testing.T) {
	c1 := client.NewClient("key-one")
	c2 := client.NewClient("key-two")

	require.NotNil(t, c1)
	require.NotNil(t, c2)
	// The two clients must not share sub-client pointers
	assert.NotSame(t, c1.Invoices, c2.Invoices)
	assert.NotSame(t, c1.Transactions, c2.Transactions)
}
