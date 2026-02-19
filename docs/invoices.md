# Invoices

The Invoices service lets you create hosted payment pages and check their status. Access it via `client.Invoices`.

---

## Overview

The invoice flow works as follows:

1. Your server calls **`Create`** with the amount, currency, and optional metadata.
2. The API returns a **`CheckoutURL`** — redirect your customer there.
3. The customer completes payment on the hosted Nodela checkout page.
4. Nodela sends a **webhook** to your `WebhookURL` (if configured) and redirects the customer to `SuccessURL`.
5. Your server calls **`Verify`** with the invoice ID to confirm payment before fulfilling the order.

---

## Methods

### `Create`

```go
func (s *Invoices) Create(ctx context.Context, params models.CreateInvoiceParams) (*models.CreateInvoiceResponse, error)
```

Creates a new payment invoice. Returns a `CheckoutURL` for the hosted payment page.

#### Parameters — `models.CreateInvoiceParams`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Amount` | `float64` | Yes | Amount to charge in the specified currency |
| `Currency` | `string` | Yes | ISO 4217 currency code (case-insensitive). See [Supported Currencies](./supported-currencies.md) |
| `SuccessURL` | `string` | No | URL to redirect the customer after successful payment |
| `CancelURL` | `string` | No | URL to redirect the customer if they cancel payment |
| `WebhookURL` | `string` | No | URL where Nodela sends a POST request on payment confirmation |
| `Reference` | `string` | No | Your internal order or reference ID |
| `Title` | `string` | No | Short title shown on the checkout page |
| `Description` | `string` | No | Longer description shown on the checkout page |
| `Customer` | `*InvoiceCustomer` | No | Customer email and optional name |

#### `InvoiceCustomer` fields

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Email` | `string` | Yes (if Customer set) | Customer's email address |
| `Name` | `string` | No | Customer's display name |

#### Return value — `*models.CreateInvoiceResponse`

| Field | Type | Description |
| --- | --- | --- |
| `Success` | `bool` | `true` when the invoice was created successfully |
| `Data` | `CreateInvoiceData` | Invoice details (see below) |
| `Error` | `*InvoiceErrorDetail` | Populated on failure (non-nil `error` return also set) |

#### `CreateInvoiceData` fields

| Field | Type | Description |
| --- | --- | --- |
| `ID` | `string` | Internal record ID |
| `InvoiceID` | `string` | Public invoice identifier — use this for `Verify` calls |
| `OriginalAmount` | `float64` | Amount as passed in the request |
| `OriginalCurrency` | `string` | Currency as passed in the request (uppercased) |
| `Amount` | `float64` | Settled amount (may differ due to FX) |
| `Currency` | `string` | Settlement currency |
| `ExchangeRate` | `float64` | Exchange rate applied (if currency converted) |
| `WebhookURL` | `string` | Echo of the webhook URL provided |
| `Customer` | `*InvoiceCustomer` | Echo of customer details |
| `CheckoutURL` | `string` | Hosted payment page URL — redirect the customer here |
| `Status` | `string` | Current invoice status (e.g., `pending`) |
| `CreatedAt` | `time.Time` | Invoice creation timestamp |

#### Examples

**Minimal — amount and currency only:**

```go
resp, err := client.Invoices.Create(context.Background(), models.CreateInvoiceParams{
    Amount:   100.00,
    Currency: "USD",
})
if err != nil {
    log.Fatal(err)
}

fmt.Println(resp.Data.CheckoutURL)
```

**Full — all optional fields:**

```go
resp, err := client.Invoices.Create(context.Background(), models.CreateInvoiceParams{
    Amount:      5000.00,
    Currency:    "NGN",
    SuccessURL:  "https://shop.example.com/orders/123/success",
    CancelURL:   "https://shop.example.com/cart",
    WebhookURL:  "https://shop.example.com/webhooks/payment",
    Reference:   "order_00123",
    Title:       "Order #00123",
    Description: "2× Widget Pro + 1× Widget Lite",
    Customer: &models.InvoiceCustomer{
        Email: "buyer@example.com",
        Name:  "Amara Obi",
    },
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Invoice %s created. Redirect to: %s\n",
    resp.Data.InvoiceID, resp.Data.CheckoutURL)
```

**Currency is case-insensitive:**

```go
// All three are equivalent — currency is normalized to "USD" internally
client.Invoices.Create(ctx, models.CreateInvoiceParams{Amount: 10, Currency: "usd"})
client.Invoices.Create(ctx, models.CreateInvoiceParams{Amount: 10, Currency: "USD"})
client.Invoices.Create(ctx, models.CreateInvoiceParams{Amount: 10, Currency: "Usd"})
```

**Unsupported currency — returns `SDKError` before any HTTP call:**

```go
_, err := client.Invoices.Create(ctx, models.CreateInvoiceParams{
    Amount:   10,
    Currency: "XYZ",
})

var sdkErr *nodela_errors.SDKError
if errors.As(err, &sdkErr) {
    // sdkErr.Code    == nodela_errors.ErrCodeValidation
    // sdkErr.Message == "unsupported currency: XYZ"
}
```

---

### `Verify`

```go
func (s *Invoices) Verify(ctx context.Context, invoiceID string) (*models.VerifyInvoiceResponse, error)
```

Retrieves the current status of an invoice. Use this to confirm payment before fulfilling an order.

#### Parameters

| Parameter | Type | Description |
| --- | --- | --- |
| `ctx` | `context.Context` | Request context (for cancellation and deadlines) |
| `invoiceID` | `string` | The `InvoiceID` from a `CreateInvoiceResponse` |

#### Return value — `*models.VerifyInvoiceResponse`

| Field | Type | Description |
| --- | --- | --- |
| `Success` | `bool` | `true` when the request succeeded |
| `Data` | `VerifyInvoiceData` | Full invoice state |
| `Error` | `*InvoiceErrorDetail` | Populated on API-level failure |

#### `VerifyInvoiceData` fields

All fields from `CreateInvoiceData`, plus:

| Field | Type | Description |
| --- | --- | --- |
| `Reference` | `string` | Your reference ID echoed back |
| `Paid` | `bool` | `true` when payment has been confirmed on-chain |
| `Payment` | `*InvoicePayment` | Payment details — only populated when `Paid` is `true` |

#### `InvoicePayment` fields

| Field | Type | Description |
| --- | --- | --- |
| `ID` | `string` | Internal payment record ID |
| `Network` | `string` | Blockchain network used (e.g., `ethereum`, `tron`) |
| `Token` | `string` | Token/asset used for payment (e.g., `USDT`, `USDC`) |
| `Address` | `string` | On-chain wallet address that received funds |
| `Amount` | `float64` | Amount received on-chain |
| `Status` | `string` | Payment status |
| `TxHash` | `[]string` | One or more blockchain transaction hashes |
| `TransactionType` | `string` | Type of on-chain transaction |
| `PayerEmail` | `string` | Payer's email (if provided at checkout) |
| `CreatedAt` | `time.Time` | When the payment was recorded |

#### Examples

**Check if paid:**

```go
resp, err := client.Invoices.Verify(context.Background(), "inv_xxxxxxxxx")
if err != nil {
    log.Fatal(err)
}

if resp.Data.Paid {
    fmt.Println("Payment confirmed!")
    fmt.Printf("Network: %s | Token: %s\n",
        resp.Data.Payment.Network,
        resp.Data.Payment.Token,
    )
    for i, hash := range resp.Data.Payment.TxHash {
        fmt.Printf("Tx[%d]: %s\n", i, hash)
    }
} else {
    fmt.Printf("Still pending (status: %s)\n", resp.Data.Status)
}
```

**Handle 404 — invoice not found:**

```go
resp, err := client.Invoices.Verify(context.Background(), "inv_doesnotexist")
if err != nil {
    if nodela_errors.IsNotFoundError(err) {
        fmt.Println("Invoice does not exist")
        return
    }
    log.Fatal(err)
}
```

**Polling pattern (use webhooks in production):**

```go
func waitForPayment(ctx context.Context, client *nodela.Client, invoiceID string) error {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            resp, err := client.Invoices.Verify(ctx, invoiceID)
            if err != nil {
                return fmt.Errorf("verify: %w", err)
            }
            if resp.Data.Paid {
                fmt.Printf("Paid! Tx: %v\n", resp.Data.Payment.TxHash)
                return nil
            }
        }
    }
}

// Usage
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
defer cancel()

if err := waitForPayment(ctx, client, invoiceID); err != nil {
    log.Fatal(err)
}
```

---

## Webhook Events

When a payment is confirmed, Nodela sends a `POST` request to your `WebhookURL` with a JSON body describing the payment event. Always **verify the invoice server-side** after receiving a webhook — do not trust the webhook payload alone.

Webhook-driven confirmation:

```go
http.HandleFunc("/webhooks/nodela", func(w http.ResponseWriter, r *http.Request) {
    // Parse webhook payload
    var payload struct {
        InvoiceID string `json:"invoice_id"`
        Event     string `json:"event"`
    }
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "bad request", http.StatusBadRequest)
        return
    }

    if payload.Event == "payment.confirmed" {
        // Always re-verify server-side
        resp, err := client.Invoices.Verify(r.Context(), payload.InvoiceID)
        if err != nil || !resp.Data.Paid {
            // Payment not actually confirmed — ignore or log
            w.WriteHeader(http.StatusOK)
            return
        }

        // Fulfill the order
        fulfillOrder(payload.InvoiceID)
    }

    w.WriteHeader(http.StatusOK)
})
```

---

## Error Scenarios

| Error | When | How to handle |
| --- | --- | --- |
| `SDKError` (ErrCodeValidation) | Unsupported currency | Fix the currency code; see [Supported Currencies](./supported-currencies.md) |
| `APIError` (401) | Invalid API key | Verify `NODELA_API_KEY`; see [Authentication](./authentication.md) |
| `APIError` (400) | Malformed request | Check field types and required fields |
| `APIError` (404) | Invoice not found on `Verify` | The invoice ID does not exist or belongs to another account |
| `APIError` (429) | Rate limit exceeded | Back off and retry with exponential delay |
| `SDKError` (ErrCodeNetwork) | Network failure | Check connectivity; retry with backoff |

---

## Next Steps

- [Transactions](./transactions.md) — list settled transactions for reconciliation.
- [Error Handling](./error-handling.md) — handle every error type correctly.
- [Supported Currencies](./supported-currencies.md) — full list of valid currency codes.
- [API Reference](./api-reference.md) — complete type definitions.
