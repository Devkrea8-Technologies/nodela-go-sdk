# Getting Started

This guide walks you from zero to a working Nodela integration in Go. By the end you will have installed the SDK, initialized a client, created a payment invoice, and handled the response.

---

## Prerequisites

- **Go 1.22 or later** installed on your machine. Verify with:
  ```bash
  go version
  # go version go1.22.x ...
  ```
- A **Nodela account** with an API key. Sign up at [nodela.co](https://nodela.co) if you don't have one.

---

## Step 1 — Install the SDK

Run the following command inside your Go module:

```bash
go get nodela.co/go-sdk
```

This adds the SDK to your `go.mod` and `go.sum`. The SDK has no runtime dependencies beyond the Go standard library.

---

## Step 2 — Initialize the Client

Import the client package and create a single `Client` instance. The client is safe for concurrent use, so create one and reuse it for the lifetime of your application.

```go
package main

import (
    "os"

    nodela "nodela.co/go-sdk/pkg/client"
)

func main() {
    client := nodela.NewClient(os.Getenv("NODELA_API_KEY"))

    // client.Invoices    — invoice operations
    // client.Transactions — transaction operations
}
```

> **Security:** Never hard-code your API key in source files. Load it from an environment variable, a secrets manager (e.g., AWS Secrets Manager, HashiCorp Vault), or a `.env` file that is excluded from version control.

---

## Step 3 — Create Your First Invoice

An invoice generates a hosted checkout page. Redirect your customer to the returned `CheckoutURL` to complete payment.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    nodela "nodela.co/go-sdk/pkg/client"
    "nodela.co/go-sdk/pkg/models"
)

func main() {
    client := nodela.NewClient(os.Getenv("NODELA_API_KEY"))

    resp, err := client.Invoices.Create(context.Background(), models.CreateInvoiceParams{
        Amount:     50.00,
        Currency:   "USD",
        SuccessURL: "https://yourapp.com/success",
        CancelURL:  "https://yourapp.com/cancel",
    })
    if err != nil {
        log.Fatalf("failed to create invoice: %v", err)
    }

    fmt.Println("Invoice ID:  ", resp.Data.InvoiceID)
    fmt.Println("Checkout URL:", resp.Data.CheckoutURL)
    fmt.Println("Status:      ", resp.Data.Status)
}
```

Redirect your user to `resp.Data.CheckoutURL` to complete the payment.

---

## Step 4 — Verify Payment Status

After the customer pays (or is redirected back to your `SuccessURL`), call `Verify` to confirm payment on the server side.

```go
status, err := client.Invoices.Verify(context.Background(), resp.Data.InvoiceID)
if err != nil {
    log.Fatalf("failed to verify invoice: %v", err)
}

if status.Data.Paid {
    fmt.Println("Payment confirmed!")
    fmt.Println("Network:", status.Data.Payment.Network)
    fmt.Println("Token:  ", status.Data.Payment.Token)
} else {
    fmt.Println("Payment pending. Status:", status.Data.Status)
}
```

> **Note:** Do not rely solely on URL redirects for payment confirmation. Always verify payment status server-side, either by polling `Verify` or by listening for a webhook event at your `WebhookURL`.

---

## Step 5 — List Transactions

Retrieve a paginated list of all settled transactions on your account:

```go
txResp, err := client.Transactions.List(context.Background(), nil)
if err != nil {
    log.Fatalf("failed to list transactions: %v", err)
}

for _, tx := range txResp.Data.Transactions {
    fmt.Printf("%s  %s %.2f  paid=%v\n",
        tx.ID, tx.Currency, tx.Amount, tx.Paid)
}

fmt.Printf("Page %d of %d — %d total\n",
    txResp.Data.Pagination.Page,
    txResp.Data.Pagination.TotalPages,
    txResp.Data.Pagination.Total,
)
```

---

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    nodela "nodela.co/go-sdk/pkg/client"
    nodela_errors "nodela.co/go-sdk/pkg/errors"
    "nodela.co/go-sdk/pkg/models"
    "errors"
)

func main() {
    client := nodela.NewClient(os.Getenv("NODELA_API_KEY"))

    // 1. Create invoice
    invoice, err := client.Invoices.Create(context.Background(), models.CreateInvoiceParams{
        Amount:      100.00,
        Currency:    "NGN",
        Reference:   "order_001",
        Title:       "Premium Plan",
        Description: "12-month premium subscription",
        SuccessURL:  "https://yourapp.com/success",
        CancelURL:   "https://yourapp.com/cancel",
        WebhookURL:  "https://yourapp.com/webhooks/nodela",
        Customer: &models.InvoiceCustomer{
            Email: "user@example.com",
            Name:  "John Doe",
        },
    })
    if err != nil {
        log.Fatalf("create invoice: %v", err)
    }

    fmt.Println("Send customer to:", invoice.Data.CheckoutURL)

    // 2. Poll for payment (in production, use webhooks instead)
    invoiceID := invoice.Data.InvoiceID
    for i := 0; i < 10; i++ {
        time.Sleep(5 * time.Second)

        status, err := client.Invoices.Verify(context.Background(), invoiceID)
        if err != nil {
            var apiErr *nodela_errors.APIError
            if errors.As(err, &apiErr) && nodela_errors.IsNotFoundError(err) {
                log.Fatal("Invoice not found")
            }
            log.Printf("verify error: %v", err)
            continue
        }

        if status.Data.Paid {
            fmt.Printf("PAID! Tx hash: %v\n", status.Data.Payment.TxHash)
            break
        }

        fmt.Printf("Still pending... (%d/10)\n", i+1)
    }
}
```

---

## Next Steps

- [Authentication](./authentication.md) — learn how API keys are applied and best security practices.
- [Configuration](./configuration.md) — customize timeouts and the underlying HTTP client.
- [Invoices](./invoices.md) — full invoice API reference with all parameters and response fields.
- [Transactions](./transactions.md) — pagination, filtering, and working with transaction data.
- [Error Handling](./error-handling.md) — understand every error type and how to handle them.
