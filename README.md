# Nodela Go SDK

[![Go Version](https://img.shields.io/badge/go-1.22+-blue.svg)](https://golang.org/dl/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
[![Go Reference](https://pkg.go.dev/badge/nodela.co/go-sdk.svg)](https://pkg.go.dev/nodela.co/go-sdk)

The official Go client library for the [Nodela](https://nodela.co) payment API. Accept global payments, create invoices, and reconcile transactions — all from idiomatic Go code.

---

## Table of Contents

- [Requirements](#requirements)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Authentication](#authentication)
- [Configuration](#configuration)
- [Invoices](#invoices)
  - [Create an Invoice](#create-an-invoice)
  - [Verify an Invoice](#verify-an-invoice)
- [Transactions](#transactions)
  - [List Transactions](#list-transactions)
  - [Pagination](#pagination)
- [Error Handling](#error-handling)
- [Supported Currencies](#supported-currencies)
- [Contributing](#contributing)
- [License](#license)

---

## Requirements

- Go **1.22** or later
- A Nodela account and API key — sign up at [nodela.co](https://nodela.co)

---

## Installation

```bash
go get nodela.co/go-sdk
```

---

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    nodela "nodela.co/go-sdk/pkg/client"
)

func main() {
    client := nodela.NewClient("your-api-key")

    invoice, err := client.Invoices.Create(context.Background(), nodela.CreateInvoiceParams{
        Amount:     100.00,
        Currency:   "USD",
        SuccessURL: "https://yourapp.com/success",
        CancelURL:  "https://yourapp.com/cancel",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Checkout URL:", invoice.Data.CheckoutURL)
}
```

---

## Documentation

For full API reference, guides, and examples, see the [docs](docs/) folder:

| Page | Description |
|---|---|
| [Getting Started](docs/getting-started.md) | Installation, API keys, and first steps |
| [Authentication](docs/authentication.md) | How API key authentication works and best practices |
| [Configuration](docs/configuration.md) | SDK options, timeouts, and custom HTTP clients |
| [Invoices](docs/invoices.md) | Create payment invoices and verify status |
| [Transactions](docs/transactions.md) | List and paginate through transaction history |
| [Error Handling](docs/error-handling.md) | Structured error types and pattern matching |
| [Supported Currencies](docs/supported-currencies.md) | Full list of 60+ supported fiat currencies |
| [API Reference](docs/api-reference.md) | Complete reference for every exported type and method |

---

## Authentication

All API requests are authenticated with your Nodela **API key** via the `Authorization: Bearer <key>` header. The SDK handles this automatically — just pass your key to `NewClient`.

```go
client := nodela.NewClient("ndk_live_xxxxxxxxxxxxxxxxxxxx")
```

> **Keep your API key secret.** Do not hard-code it in source files. Load it from an environment variable or secrets manager:

```go
import "os"

client := nodela.NewClient(os.Getenv("NODELA_API_KEY"))
```

---

## Configuration

`NewClient` accepts optional functional options after the API key.

### Custom Timeout

The default HTTP timeout is **30 seconds**. Override it with `WithTimeout`:

```go
import "time"

client := nodela.NewClient(
    os.Getenv("NODELA_API_KEY"),
    nodela.WithTimeout(60 * time.Second),
)
```

### Custom HTTP Client

Provide your own `*http.Client` for proxy support, custom TLS settings, or transport-level instrumentation:

```go
import "net/http"

transport := &http.Transport{
    // ... custom settings
}

httpClient := &http.Client{Transport: transport}

client := nodela.NewClient(
    os.Getenv("NODELA_API_KEY"),
    nodela.WithHTTPClient(httpClient),
)
```

### Reusing the Client

The client is safe for concurrent use. Create **one instance** per application and reuse it across goroutines:

```go
var nodelaClient *nodela.Client

func init() {
    nodelaClient = nodela.NewClient(os.Getenv("NODELA_API_KEY"))
}
```

---

## Invoices

### Create an Invoice

Generate a hosted checkout page that your customer pays through.

```go
import (
    "context"
    nodela "nodela.co/go-sdk/pkg/client"
    "nodela.co/go-sdk/pkg/models"
)

params := models.CreateInvoiceParams{
    // Required fields
    Amount:   250.00,
    Currency: "NGN",

    // Optional redirect URLs
    SuccessURL: "https://yourapp.com/payment/success",
    CancelURL:  "https://yourapp.com/payment/cancel",

    // Optional webhook for server-side confirmation
    WebhookURL: "https://yourapp.com/webhooks/nodela",

    // Optional metadata
    Reference:   "order_8675309",
    Title:       "Pro Plan Subscription",
    Description: "Monthly subscription to Acme Pro",

    // Optional customer info
    Customer: &models.InvoiceCustomer{
        Email: "jane@example.com",
        Name:  "Jane Doe",
    },
}

resp, err := client.Invoices.Create(context.Background(), params)
if err != nil {
    // Handle error — see Error Handling section
    log.Fatal(err)
}

fmt.Println("Invoice ID:   ", resp.Data.InvoiceID)
fmt.Println("Checkout URL: ", resp.Data.CheckoutURL)
fmt.Println("Status:       ", resp.Data.Status)
```

**Currency is case-insensitive** — `"usd"`, `"USD"`, and `"Usd"` are all accepted and normalized to uppercase.

### Verify an Invoice

Poll or webhook-trigger a status check on an existing invoice.

```go
resp, err := client.Invoices.Verify(context.Background(), "inv_xxxxxxxxxxxxxxxxx")
if err != nil {
    var apiErr *errors.APIError
    if errors.As(err, &apiErr) && nodela_errors.IsNotFoundError(err) {
        fmt.Println("Invoice not found")
        return
    }
    log.Fatal(err)
}

if resp.Data.Paid {
    fmt.Println("Payment confirmed!")
    fmt.Println("Network:    ", resp.Data.Payment.Network)
    fmt.Println("Token:      ", resp.Data.Payment.Token)
    fmt.Println("Tx Hashes:  ", resp.Data.Payment.TxHash)
} else {
    fmt.Println("Awaiting payment. Status:", resp.Data.Status)
}
```

---

## Transactions

### List Transactions

Retrieve settled transactions for your account. All fields are optional — calling with `nil` returns all transactions.

```go
resp, err := client.Transactions.List(context.Background(), nil)
if err != nil {
    log.Fatal(err)
}

for _, tx := range resp.Data.Transactions {
    fmt.Printf("ID: %s | Amount: %.2f %s | Paid: %v\n",
        tx.ID, tx.Amount, tx.Currency, tx.Paid)
}

fmt.Printf("Page %d of %d (Total: %d)\n",
    resp.Data.Pagination.Page,
    resp.Data.Pagination.TotalPages,
    resp.Data.Pagination.Total,
)
```

### Pagination

Use `Page` and `Limit` to walk through large result sets:

```go
import "nodela.co/go-sdk/pkg/models"

page := 1
limit := 20

for {
    resp, err := client.Transactions.List(context.Background(), &models.ListTransactionsParams{
        Page:  &page,
        Limit: &limit,
    })
    if err != nil {
        log.Fatal(err)
    }

    for _, tx := range resp.Data.Transactions {
        fmt.Println(tx.ID, tx.Amount, tx.Currency)
    }

    if !resp.Data.Pagination.HasMore {
        break
    }
    page++
}
```

---

## Error Handling

The SDK returns two distinct error types.

### SDKError — Client-side validation errors

Returned before any network request is made (e.g., unsupported currency).

```go
import (
    nodela_errors "nodela.co/go-sdk/pkg/errors"
    "errors"
)

_, err := client.Invoices.Create(ctx, models.CreateInvoiceParams{
    Amount:   100,
    Currency: "XYZ", // Not supported
})

var sdkErr *nodela_errors.SDKError
if errors.As(err, &sdkErr) {
    fmt.Println("Validation error:", sdkErr.Message)
    fmt.Println("Error code:     ", sdkErr.Code)
}
```

### APIError — Server-side errors

Returned when the Nodela API responds with a non-2xx status code.

```go
var apiErr *nodela_errors.APIError
if errors.As(err, &apiErr) {
    fmt.Println("HTTP status:", apiErr.StatusCode)
    fmt.Println("Message:    ", apiErr.Message)
    fmt.Println("Code:       ", apiErr.Code)
}
```

### Predicate Helpers

```go
if nodela_errors.IsNotFoundError(err) {
    // 404 — invoice or resource does not exist
}

if nodela_errors.IsAuthenticatedError(err) {
    // 401 — invalid or missing API key
}

if nodela_errors.IsRateLimitError(err) {
    // 429 — too many requests, back off and retry
}
```

### Error Code Reference

| Code | Description |
| --- | --- |
| `ErrCodeUnknown` | Unexpected internal condition |
| `ErrCodeNetwork` | Transport failure (DNS, timeout, connection refused) |
| `ErrCodeSerialization` | Failed to marshal request body |
| `ErrCodeDeserialization` | Failed to unmarshal response body |
| `ErrCodeRequest` | Failed to construct HTTP request |
| `ErrCodeResponse` | Failed to read HTTP response |
| `ErrCodeValidation` | SDK-side input validation failed |
| `ErrCodeAuthentication` | HTTP 401 — invalid credentials |
| `ErrCodeAuthorization` | HTTP 403 — insufficient permissions |
| `ErrCodeNotFound` | HTTP 404 — resource not found |
| `ErrCodeRateLimit` | HTTP 429 — rate limit exceeded |

---

## Supported Currencies

The SDK validates currencies against 63 supported ISO 4217 codes before sending a request.

| Region | Currencies |
| --- | --- |
| Americas | USD, CAD, MXN, BRL, ARS, CLP, COP, PEN, JMD, TTD |
| Europe | EUR, GBP, CHF, SEK, NOK, DKK, PLN, CZK, HUF, RON, BGN, HRK, ISK, TRY, RUB, UAH |
| Africa | NGN, ZAR, KES, GHS, EGP, MAD, TZS, UGX, XOF, XAF, ETB |
| Asia | JPY, CNY, INR, KRW, IDR, MYR, THB, PHP, VND, SGD, HKD, TWD, BDT, PKR, LKR |
| Middle East | AED, SAR, QAR, KWD, BHD, OMR, ILS, JOD |
| Oceania | AUD, NZD, FJD |

Passing an unsupported currency returns an `SDKError` with code `ErrCodeValidation` — no HTTP request is made.

---

## Contributing

We welcome contributions! Please read [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines on how to set up the development environment, run tests, manage changelogs, and submit pull requests.

---

## License

This project is licensed under the **MIT License** — see [LICENSE](./LICENSE) for details.

Copyright © 2026 Devkrea8 Technologies
