# Error Handling

The Nodela Go SDK uses two distinct error types that map cleanly to the standard Go `error` interface. Understanding the difference between them is essential to writing robust integrations.

---

## Two Error Types

| Type | Package | When returned | Example cause |
| --- | --- | --- | --- |
| `*SDKError` | `nodela.co/go-sdk/pkg/errors` | Before any HTTP request | Unsupported currency, bad input |
| `*APIError` | `nodela.co/go-sdk/pkg/errors` | After HTTP response with non-2xx status | 401 Unauthorized, 404 Not Found |

Both types implement the standard `error` interface and are compatible with `errors.As`, `errors.Is`, and `errors.Unwrap`.

---

## SDKError — Client-Side Errors

`SDKError` is returned for problems detected within the SDK itself, before any network request is made. This includes input validation failures and internal transport issues.

### Definition

```go
type SDKError struct {
    Code    string // One of the ErrCode* constants
    Message string // Human-readable description
    Err     error  // Underlying error (may be nil)
}

func (e *SDKError) Error() string // "[CODE] message" or "[CODE] message: underlying"
func (e *SDKError) Unwrap() error // Returns Err for errors chain compatibility
```

### Checking for SDKError

```go
import (
    "errors"
    nodela_errors "nodela.co/go-sdk/pkg/errors"
)

_, err := client.Invoices.Create(ctx, models.CreateInvoiceParams{
    Amount:   100,
    Currency: "XYZ",
})

var sdkErr *nodela_errors.SDKError
if errors.As(err, &sdkErr) {
    fmt.Println("Code:   ", sdkErr.Code)    // "VALIDATION"
    fmt.Println("Message:", sdkErr.Message) // "unsupported currency: XYZ"

    if sdkErr.Err != nil {
        fmt.Println("Cause:", sdkErr.Err)
    }
}
```

---

## APIError — Server-Side Errors

`APIError` is returned when the Nodela API responds with an HTTP status code outside the 2xx range. It captures the HTTP status, the API-level error message, and any application-level code returned by the server.

### Definition

```go
type APIError struct {
    StatusCode int    // HTTP status code (e.g., 401, 404, 429)
    Message    string // Human-readable error message from the API
    Code       string // Application error code from the API response body
}

func (e *APIError) Error() string // "API error (status XXX): message [code]"
```

### Checking for APIError

```go
var apiErr *nodela_errors.APIError
if errors.As(err, &apiErr) {
    fmt.Printf("HTTP status: %d\n", apiErr.StatusCode)
    fmt.Printf("Message:     %s\n", apiErr.Message)
    fmt.Printf("Code:        %s\n", apiErr.Code)
}
```

---

## Error Codes

### SDKError codes

These constants are defined in `pkg/errors/errors.go`:

| Constant | Value | Description |
| --- | --- | --- |
| `ErrCodeUnknown` | `"UNKNOWN"` | Unexpected internal condition |
| `ErrCodeNetwork` | `"NETWORK"` | Transport failure — DNS, timeout, connection refused |
| `ErrCodeSerialization` | `"SERIALIZATION"` | Failed to marshal the request body to JSON |
| `ErrCodeDeserialization` | `"DESERIALIZATION"` | Failed to unmarshal the response body from JSON |
| `ErrCodeRequest` | `"REQUEST"` | Failed to construct the `*http.Request` |
| `ErrCodeResponse` | `"RESPONSE"` | Failed to read the HTTP response body |
| `ErrCodeValidation` | `"VALIDATION"` | SDK-side input validation failed (e.g., bad currency) |
| `ErrCodeAuthentication` | `"AUTHENTICATION"` | HTTP 401 — invalid or missing API credentials |
| `ErrCodeAuthorization` | `"AUTHORIZATION"` | HTTP 403 — valid key but insufficient permissions |
| `ErrCodeNotFound` | `"NOT_FOUND"` | HTTP 404 — resource does not exist |
| `ErrCodeRateLimit` | `"RATE_LIMIT"` | HTTP 429 — too many requests |

---

## Predicate Helpers

The SDK provides three convenience functions for the most common API error checks. These work on any `error` value — you don't need to unwrap it yourself.

```go
// Returns true for APIError with StatusCode == 404
nodela_errors.IsNotFoundError(err)

// Returns true for APIError with StatusCode == 401
nodela_errors.IsAuthenticatedError(err)

// Returns true for APIError with StatusCode == 429
nodela_errors.IsRateLimitError(err)
```

### Usage example

```go
resp, err := client.Invoices.Verify(ctx, invoiceID)
if err != nil {
    switch {
    case nodela_errors.IsNotFoundError(err):
        fmt.Println("Invoice does not exist")
    case nodela_errors.IsAuthenticatedError(err):
        fmt.Println("Invalid API key — check NODELA_API_KEY")
    case nodela_errors.IsRateLimitError(err):
        fmt.Println("Rate limited — back off and retry")
    default:
        fmt.Printf("Unexpected error: %v\n", err)
    }
    return
}
```

---

## Handling All Error Scenarios

### Full error-handling pattern

```go
import (
    "context"
    "errors"
    "fmt"
    "log"
    "time"

    nodela "nodela.co/go-sdk/pkg/client"
    nodela_errors "nodela.co/go-sdk/pkg/errors"
    "nodela.co/go-sdk/pkg/models"
)

func createInvoice(client *nodela.Client, params models.CreateInvoiceParams) (*models.CreateInvoiceResponse, error) {
    resp, err := client.Invoices.Create(context.Background(), params)
    if err != nil {
        var sdkErr *nodela_errors.SDKError
        var apiErr *nodela_errors.APIError

        switch {
        case errors.As(err, &sdkErr):
            switch sdkErr.Code {
            case nodela_errors.ErrCodeValidation:
                // Input error — do not retry
                return nil, fmt.Errorf("invalid parameters: %s", sdkErr.Message)
            case nodela_errors.ErrCodeNetwork:
                // Transient — safe to retry
                return nil, fmt.Errorf("network error, retry: %w", err)
            default:
                return nil, fmt.Errorf("sdk error [%s]: %s", sdkErr.Code, sdkErr.Message)
            }

        case errors.As(err, &apiErr):
            switch {
            case nodela_errors.IsAuthenticatedError(err):
                // Fatal — wrong key
                log.Fatal("NODELA_API_KEY is invalid")
            case nodela_errors.IsRateLimitError(err):
                // Transient — back off
                return nil, fmt.Errorf("rate limited, retry after delay: %w", err)
            case apiErr.StatusCode >= 500:
                // Transient server error
                return nil, fmt.Errorf("server error %d, retry: %w", apiErr.StatusCode, err)
            default:
                return nil, fmt.Errorf("api error %d: %s", apiErr.StatusCode, apiErr.Message)
            }

        default:
            return nil, fmt.Errorf("unexpected error: %w", err)
        }
    }

    return resp, nil
}
```

---

## Retry Strategy

The SDK does **not** retry automatically. You are responsible for implementing retry logic appropriate to your use case.

**Retriable errors** (transient):
- `ErrCodeNetwork` — connection refused, timeout, DNS failure
- `APIError` with `StatusCode >= 500` — Nodela server errors
- `APIError` with `StatusCode == 429` — rate limit exceeded

**Non-retriable errors** (permanent):
- `ErrCodeValidation` — fix the input
- `APIError` with `StatusCode == 400` — bad request; fix the payload
- `APIError` with `StatusCode == 401` — fix the API key
- `APIError` with `StatusCode == 404` — resource does not exist

### Exponential backoff example

```go
import (
    "context"
    "fmt"
    "math"
    "time"
)

func withRetry(ctx context.Context, maxAttempts int, fn func() error) error {
    for attempt := 1; attempt <= maxAttempts; attempt++ {
        err := fn()
        if err == nil {
            return nil
        }

        // Do not retry non-transient errors
        var sdkErr *nodela_errors.SDKError
        if errors.As(err, &sdkErr) && sdkErr.Code == nodela_errors.ErrCodeValidation {
            return err
        }

        var apiErr *nodela_errors.APIError
        if errors.As(err, &apiErr) && apiErr.StatusCode < 500 && apiErr.StatusCode != 429 {
            return err
        }

        if attempt == maxAttempts {
            return fmt.Errorf("failed after %d attempts: %w", maxAttempts, err)
        }

        // Exponential backoff: 1s, 2s, 4s, 8s…
        delay := time.Duration(math.Pow(2, float64(attempt-1))) * time.Second
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(delay):
        }
    }
    return nil
}

// Usage
err := withRetry(ctx, 4, func() error {
    _, err := client.Invoices.Create(ctx, params)
    return err
})
```

---

## Context Cancellation

All SDK methods accept a `context.Context`. If the context is cancelled or its deadline is exceeded before the request completes, the SDK returns a network-level error wrapping `context.Canceled` or `context.DeadlineExceeded`.

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

_, err := client.Invoices.Create(ctx, params)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        fmt.Println("Request timed out")
    }
    if errors.Is(err, context.Canceled) {
        fmt.Println("Request was cancelled")
    }
}
```

---

## Error Chain Compatibility

`SDKError` implements `Unwrap()`, making it compatible with the standard `errors` package:

```go
// errors.Is traverses the chain
sentinel := fmt.Errorf("root cause")
sdkErr := nodela_errors.NewSDKError(nodela_errors.ErrCodeNetwork, "connection failed", sentinel)

fmt.Println(errors.Is(sdkErr, sentinel)) // true
```

---

## Next Steps

- [Invoices](./invoices.md) — all error scenarios for invoice operations.
- [Transactions](./transactions.md) — all error scenarios for transaction listing.
- [API Reference](./api-reference.md) — full type definitions for `SDKError` and `APIError`.
