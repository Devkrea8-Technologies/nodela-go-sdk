package unit

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"nodela.co/go-sdk/pkg/client"
)

// testTransport rewrites every outbound request so it hits the given test
// server instead of the real API host, while preserving path and query.
type testTransport struct {
	serverURL string
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	target, _ := url.Parse(t.serverURL)
	req = req.Clone(req.Context())
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	return http.DefaultTransport.RoundTrip(req)
}

// newTestClient creates a *client.Client wired to a fresh httptest.Server
// that will call handler for every request.  The server is closed
// automatically when the test completes.
func newTestClient(t *testing.T, handler http.Handler) *client.Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	transport := &testTransport{serverURL: server.URL}
	c := client.NewClient("test-api-key", client.WithHTTPClient(&http.Client{Transport: transport}))
	return c
}

// ptr returns a pointer to the supplied value — handy for optional model fields.
func ptr[T any](v T) *T { return &v }
