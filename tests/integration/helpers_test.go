package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"nodela.co/go-sdk/pkg/client"
)

// testTransport redirects every outbound request to the given test server URL
// while preserving the original path and query string.
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

// newTestClient wires a *client.Client to an httptest.Server backed by handler.
// The server is closed when the test ends.
func newTestClient(t *testing.T, handler http.Handler) (*client.Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	transport := &testTransport{serverURL: server.URL}
	c := client.NewClient("integration-test-key", client.WithHTTPClient(&http.Client{Transport: transport}))
	return c, server
}

// writeJSON writes v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// ptr returns a pointer to v.
func ptr[T any](v T) *T { return &v }
