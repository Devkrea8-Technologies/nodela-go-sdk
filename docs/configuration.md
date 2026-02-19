# Configuration

The Nodela Go SDK uses the **functional options pattern** for configuration. All settings are passed as variadic `Option` arguments to `NewClient` after the API key. This keeps the constructor signature stable — adding new options is always backwards-compatible.

---

## Default Settings

When you call `NewClient` with only an API key, the SDK uses these defaults:

| Setting | Default Value |
| --- | --- |
| Base URL | `https://api.nodela.co` |
| HTTP timeout | 30 seconds |
| HTTP client | `&http.Client{}` with the default transport |
| User-Agent | `nodela-go-sdk/1.0.1` |

---

## Available Options

### `WithTimeout(duration time.Duration)`

Overrides the default 30-second timeout for all HTTP requests made by the client.

```go
import (
    "time"
    nodela "nodela.co/go-sdk/pkg/client"
)

client := nodela.NewClient(
    os.Getenv("NODELA_API_KEY"),
    nodela.WithTimeout(60 * time.Second),
)
```

**When to increase the timeout:**
- Long-running operations on slower networks.
- Environments with high API latency (e.g., cross-region calls).

**When to decrease the timeout:**
- Latency-sensitive paths where a slow response should fail fast.
- Serverless functions with strict execution limits.

The timeout applies to the entire round trip: DNS resolution, TCP handshake, TLS handshake, request write, server processing, and response read.

---

### `WithHTTPClient(client *http.Client)`

Replaces the default HTTP client entirely. This gives you full control over the underlying transport layer.

```go
import (
    "net/http"
    nodela "nodela.co/go-sdk/pkg/client"
)

httpClient := &http.Client{
    Timeout: 45 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}

client := nodela.NewClient(
    os.Getenv("NODELA_API_KEY"),
    nodela.WithHTTPClient(httpClient),
)
```

**Common use cases:**

#### Custom TLS Configuration

```go
import (
    "crypto/tls"
    "net/http"
)

tlsConfig := &tls.Config{
    MinVersion: tls.VersionTLS13,
}

transport := &http.Transport{
    TLSClientConfig: tlsConfig,
}

httpClient := &http.Client{Transport: transport}

client := nodela.NewClient(apiKey, nodela.WithHTTPClient(httpClient))
```

#### HTTP Proxy

```go
import (
    "net/http"
    "net/url"
)

proxyURL, _ := url.Parse("http://proxy.corp.example.com:8080")

transport := &http.Transport{
    Proxy: http.ProxyURL(proxyURL),
}

httpClient := &http.Client{Transport: transport}
client := nodela.NewClient(apiKey, nodela.WithHTTPClient(httpClient))
```

#### Request Tracing / Instrumentation

```go
import (
    "net/http"
    "net/http/httptrace"
    "context"
)

// Wrap the transport with your tracing middleware
type tracingTransport struct {
    base http.RoundTripper
}

func (t *tracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    // Start span, record metrics, etc.
    return t.base.RoundTrip(req)
}

httpClient := &http.Client{
    Transport: &tracingTransport{base: http.DefaultTransport},
}

client := nodela.NewClient(apiKey, nodela.WithHTTPClient(httpClient))
```

> **Note:** If you set a `Timeout` on the custom `*http.Client`, that timeout takes precedence over any `WithTimeout` option you pass separately. To avoid confusion, use one or the other, not both.

---

## Combining Options

Options are applied in the order they are passed. All options are independent and can be combined freely:

```go
import (
    "net/http"
    "time"
    nodela "nodela.co/go-sdk/pkg/client"
)

customTransport := &http.Transport{
    MaxIdleConnsPerHost: 20,
}

client := nodela.NewClient(
    os.Getenv("NODELA_API_KEY"),
    nodela.WithTimeout(45 * time.Second),
    nodela.WithHTTPClient(&http.Client{Transport: customTransport}),
)
```

---

## Multiple Clients

You can create multiple independent clients — for example, one per environment:

```go
liveClient := nodela.NewClient(os.Getenv("NODELA_LIVE_KEY"))
testClient := nodela.NewClient(
    os.Getenv("NODELA_TEST_KEY"),
    nodela.WithTimeout(10 * time.Second),
)
```

Each client maintains its own HTTP client, API key, and service instances. They do not share state.

---

## Thread Safety

The `Client` struct and all of its service fields (`Invoices`, `Transactions`) are safe for concurrent use by multiple goroutines. Create one client per application and share it freely:

```go
var (
    nodelaClient *nodela.Client
    once         sync.Once
)

func getClient() *nodela.Client {
    once.Do(func() {
        nodelaClient = nodela.NewClient(os.Getenv("NODELA_API_KEY"))
    })
    return nodelaClient
}
```

---

## Next Steps

- [Invoices](./invoices.md) — create and verify payment invoices.
- [Transactions](./transactions.md) — list and paginate settled transactions.
- [Error Handling](./error-handling.md) — handle errors from configuration mistakes (e.g., bad API key).
