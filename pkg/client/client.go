// Package client provides the core HTTP client, functional configuration options,
// and high-level service accessors for the Nodela payment API.
//
// # Getting started
//
// Create a [Client] with [NewClient], passing your API key. The returned client
// exposes two service objects — [Client.Invoices] and [Client.Transactions] —
// that group the available API operations:
//
//	import (
//	    "context"
//	    "nodela.co/go-sdk/pkg/client"
//	    "nodela.co/go-sdk/pkg/models"
//	)
//
//	c := client.NewClient("nod_live_xxxx")
//
//	// Create an invoice
//	resp, err := c.Invoices.Create(context.Background(), models.CreateInvoiceParams{
//	    Amount:   49.99,
//	    Currency: "USD",
//	})
//
//	// Verify payment status
//	status, err := c.Invoices.Verify(context.Background(), resp.Data.InvoiceID)
//
//	// List transactions
//	txns, err := c.Transactions.List(context.Background(), nil)
//
// # Configuration
//
// Use functional options to customise the client at construction time:
//
//	c := client.NewClient(apiKey,
//	    client.WithTimeout(10*time.Second),
//	)
//
// # Concurrency
//
// A single Client instance is safe for concurrent use by multiple goroutines.
// Create one client and reuse it across the lifetime of your application rather
// than constructing a new client for every request.
//
// # Error handling
//
// All service methods return a typed error. Use errors.As to distinguish between
// SDK-side errors ([errors.SDKError]) and API-side errors ([errors.APIError]):
//
//	_, err := c.Invoices.Create(ctx, params)
//	var apiErr *sdkerrors.APIError
//	if errors.As(err, &apiErr) {
//	    log.Printf("API error %d: %s", apiErr.StatusCode, apiErr.Message)
//	}
package client

import (
	"net/http"
	"time"
)

// baseURL is the root endpoint for all Nodela API requests.
const baseURL = "https://api.nodela.co"

// Client is the root API client for the Nodela Go SDK. It manages authentication,
// HTTP transport configuration, and provides access to the Invoices and Transactions
// service objects through its exported fields.
//
// A single Client is safe for concurrent use by multiple goroutines. Construct one
// with [NewClient] and reuse it throughout your application.
type Client struct {
	apiKey     string
	httpClient *http.Client
	userAgent  string

	// Invoices provides methods for creating and verifying Nodela payment invoices.
	// See [Invoices.Create] and [Invoices.Verify].
	Invoices *Invoices

	// Transactions provides methods for listing settled Nodela transactions.
	// See [Transactions.List].
	Transactions *Transactions
}

// Option is a functional option that modifies the configuration of a [Client].
// Options are passed to [NewClient] and applied in order after the client is
// initialised with its defaults.
//
// Use the constructor functions [WithTimeout] and [WithHTTPClient] to obtain
// Option values rather than constructing them directly.
type Option func(*Client)

// NewClient creates and returns a fully initialised Nodela API client authenticated
// with the provided API key. The key is sent as a Bearer token in the Authorization
// header of every outgoing request.
//
// By default the client uses a 30-second HTTP timeout and the standard
// net/http.DefaultTransport. Pass functional [Option] values to override these
// defaults.
//
// Example — default client:
//
//	c := client.NewClient("nod_live_xxxx")
//
// Example — client with a shorter timeout:
//
//	c := client.NewClient("nod_live_xxxx", client.WithTimeout(10*time.Second))
func NewClient(apiKey string, opts ...Option) *Client {
	client := &Client{
		apiKey:    apiKey,
		userAgent: "Nodela/1.0.0",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(client)
	}

	client.Invoices = &Invoices{client: client}
	client.Transactions = &Transactions{client: client}

	return client
}

// WithTimeout returns an [Option] that sets the HTTP request timeout applied to
// all requests made by the client. The timeout covers the full round-trip
// including connection, sending the request body, and reading the response body.
//
// The default timeout is 30 seconds. Pass a shorter duration for latency-sensitive
// applications or a longer one if you expect large response payloads.
//
// Example:
//
//	c := client.NewClient(apiKey, client.WithTimeout(10*time.Second))
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithHTTPClient returns an [Option] that replaces the client's default
// *http.Client with the one provided. This is useful for:
//
//   - Injecting a custom [http.RoundTripper] (e.g. for request logging or proxying)
//   - Supplying a mock HTTP client in unit tests
//   - Sharing a pre-configured transport across multiple SDK clients
//
// When providing a custom client, ensure its Timeout is set appropriately;
// the [WithTimeout] option has no effect when used alongside [WithHTTPClient]
// unless the supplied client also has a matching timeout.
//
// Example:
//
//	custom := &http.Client{Transport: myLoggingTransport}
//	c := client.NewClient(apiKey, client.WithHTTPClient(custom))
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}
