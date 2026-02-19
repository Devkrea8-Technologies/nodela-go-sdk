# API Reference

Complete reference for every exported type, function, method, constant, and field in the Nodela Go SDK.

**Package:** `nodela.co/go-sdk`

**Go minimum version:** 1.22

---

## Table of Contents

- [Package: client](#package-client)
  - [Client](#client)
  - [NewClient](#newclient)
  - [Option](#option)
  - [WithTimeout](#withtimeout)
  - [WithHTTPClient](#withhttpclient)
  - [Invoices](#invoices-service)
  - [Transactions](#transactions-service)
- [Package: models](#package-models)
  - [CreateInvoiceParams](#createinvoiceparams)
  - [InvoiceCustomer](#invoicecustomer)
  - [CreateInvoiceResponse](#createinvoiceresponse)
  - [CreateInvoiceData](#createinvoicedata)
  - [InvoiceErrorDetail](#invoiceerrordetail)
  - [VerifyInvoiceResponse](#verifyinvoiceresponse)
  - [VerifyInvoiceData](#verifyinvoicedata)
  - [InvoicePayment](#invoicepayment)
  - [ListTransactionsParams](#listtransactionsparams)
  - [ListTransactionsResponse](#listtransactionsresponse)
  - [ListTransactionsData](#listtransactionsdata)
  - [Transaction](#transaction)
  - [TransactionCustomer](#transactioncustomer)
  - [TransactionPayment](#transactionpayment)
  - [Pagination](#pagination)
  - [SupportedInvoiceCurrencies](#supportedinvoicecurrencies)
  - [Metadata](#metadata)
  - [PaginationParams](#paginationparams)
  - [PaginatedResponse](#paginatedresponse)
- [Package: errors](#package-errors)
  - [SDKError](#sdkerror)
  - [APIError](#apierror)
  - [Error Code Constants](#error-code-constants)
  - [NewSDKError](#newsdkerror)
  - [NewAPIError](#newapierror)
  - [IsNotFoundError](#isnotfounderror)
  - [IsAuthenticatedError](#isauthenticatederror)
  - [IsRateLimitError](#isratelimiterror)

---

## Package: client

Import path: `nodela.co/go-sdk/pkg/client`

---

### Client

The root SDK object. Holds configuration and provides access to all API services.

```go
type Client struct {
    Invoices     *Invoices
    Transactions *Transactions
    // unexported fields: apiKey, httpClient, userAgent
}
```

| Field | Type | Description |
| --- | --- | --- |
| `Invoices` | `*Invoices` | Invoice service — Create and Verify operations |
| `Transactions` | `*Transactions` | Transaction service — List operation |

**Thread safety:** The `Client` and all its service fields are safe for concurrent use. Create one instance per application.

---

### NewClient

```go
func NewClient(apiKey string, opts ...Option) *Client
```

Creates and returns a new `Client` configured with the provided API key and options.

**Parameters:**

| Parameter | Type | Description |
| --- | --- | --- |
| `apiKey` | `string` | Your Nodela API key. Applied as `Authorization: Bearer <apiKey>` on every request |
| `opts` | `...Option` | Zero or more configuration options |

**Returns:** `*Client`

**Default configuration:**
- Base URL: `https://api.nodela.co`
- HTTP timeout: 30 seconds
- HTTP client: default `net/http` client

**Example:**

```go
import (
    "os"
    nodela "nodela.co/go-sdk/pkg/client"
)

client := nodela.NewClient(os.Getenv("NODELA_API_KEY"))
```

---

### Option

```go
type Option func(*Client)
```

A functional option that modifies a `Client` during construction. Pass options as variadic arguments to `NewClient`.

---

### WithTimeout

```go
func WithTimeout(timeout time.Duration) Option
```

Sets the HTTP request timeout. Overrides the default 30-second timeout.

**Parameters:**

| Parameter | Type | Description |
| --- | --- | --- |
| `timeout` | `time.Duration` | Maximum duration for a single HTTP round trip |

**Example:**

```go
import "time"

client := nodela.NewClient(apiKey, nodela.WithTimeout(60*time.Second))
```

---

### WithHTTPClient

```go
func WithHTTPClient(httpClient *http.Client) Option
```

Replaces the default `*http.Client` used for all requests. Use this to configure a custom transport, proxy, TLS settings, or instrumentation.

**Parameters:**

| Parameter | Type | Description |
| --- | --- | --- |
| `httpClient` | `*http.Client` | The custom HTTP client to use |

**Example:**

```go
import "net/http"

custom := &http.Client{
    Transport: &http.Transport{MaxIdleConnsPerHost: 20},
}
client := nodela.NewClient(apiKey, nodela.WithHTTPClient(custom))
```

---

### Invoices service

```go
type Invoices struct { /* unexported */ }
```

Provides invoice-related API operations. Accessed via `client.Invoices`.

#### `Invoices.Create`

```go
func (s *Invoices) Create(ctx context.Context, params models.CreateInvoiceParams) (*models.CreateInvoiceResponse, error)
```

Creates a new payment invoice. Performs currency validation before making any HTTP request.

**Parameters:**

| Parameter | Type | Description |
| --- | --- | --- |
| `ctx` | `context.Context` | Request context for cancellation and deadlines |
| `params` | `models.CreateInvoiceParams` | Invoice configuration — see [CreateInvoiceParams](#createinvoiceparams) |

**Returns:**

| Return | Description |
| --- | --- |
| `*models.CreateInvoiceResponse` | Invoice data including `CheckoutURL` and `InvoiceID` |
| `error` | `*SDKError` for validation failures; `*APIError` for server errors |

**Errors:**

| Error | Condition |
| --- | --- |
| `*SDKError` (ErrCodeValidation) | `params.Currency` is not in `SupportedInvoiceCurrencies` |
| `*SDKError` (ErrCodeNetwork) | Network failure before or during the request |
| `*SDKError` (ErrCodeSerialization) | Failed to marshal `params` to JSON |
| `*SDKError` (ErrCodeDeserialization) | Failed to unmarshal API response |
| `*APIError` (401) | Invalid or missing API key |
| `*APIError` (400) | Malformed request body |
| `*APIError` (429) | Rate limit exceeded |
| `*APIError` (5xx) | Nodela server error |

**HTTP details:**
- Method: `POST`
- Path: `/v1/invoices`
- Content-Type: `application/json`
- Auth: `Authorization: Bearer <apiKey>`

---

#### `Invoices.Verify`

```go
func (s *Invoices) Verify(ctx context.Context, invoiceID string) (*models.VerifyInvoiceResponse, error)
```

Retrieves the current state of an invoice. Use this to confirm payment before fulfilling an order.

**Parameters:**

| Parameter | Type | Description |
| --- | --- | --- |
| `ctx` | `context.Context` | Request context |
| `invoiceID` | `string` | The `InvoiceID` from a `CreateInvoiceResponse` |

**Returns:**

| Return | Description |
| --- | --- |
| `*models.VerifyInvoiceResponse` | Full invoice state including `Paid` flag and optional `Payment` details |
| `error` | `*APIError` for server errors |

**Errors:**

| Error | Condition |
| --- | --- |
| `*APIError` (404) | Invoice not found |
| `*APIError` (401) | Invalid or missing API key |
| `*APIError` (429) | Rate limit exceeded |
| `*APIError` (5xx) | Nodela server error |

**HTTP details:**
- Method: `GET`
- Path: `/v1/invoices/{invoiceID}`
- Auth: `Authorization: Bearer <apiKey>`

---

### Transactions service

```go
type Transactions struct { /* unexported */ }
```

Provides transaction listing. Accessed via `client.Transactions`.

#### `Transactions.List`

```go
func (s *Transactions) List(ctx context.Context, params *models.ListTransactionsParams) (*models.ListTransactionsResponse, error)
```

Returns a paginated list of settled transactions.

**Parameters:**

| Parameter | Type | Description |
| --- | --- | --- |
| `ctx` | `context.Context` | Request context |
| `params` | `*models.ListTransactionsParams` | Pagination parameters — pass `nil` for API defaults |

**Returns:**

| Return | Description |
| --- | --- |
| `*models.ListTransactionsResponse` | Transactions slice and pagination metadata |
| `error` | `*APIError` for server errors |

**Errors:**

| Error | Condition |
| --- | --- |
| `*APIError` (401) | Invalid or missing API key |
| `*APIError` (403) | Key lacks permission to list transactions |
| `*APIError` (429) | Rate limit exceeded |
| `*APIError` (5xx) | Nodela server error |

**HTTP details:**
- Method: `GET`
- Path: `/v1/transactions`
- Query params: `page`, `limit` (only set when non-nil in `params`)
- Auth: `Authorization: Bearer <apiKey>`

---

## Package: models

Import path: `nodela.co/go-sdk/pkg/models`

---

### CreateInvoiceParams

Request parameters for `Invoices.Create`.

```go
type CreateInvoiceParams struct {
    Amount      float64           `json:"amount"`
    Currency    string            `json:"currency"`
    SuccessURL  string            `json:"success_url,omitempty"`
    CancelURL   string            `json:"cancel_url,omitempty"`
    WebhookURL  string            `json:"webhook_url,omitempty"`
    Reference   string            `json:"reference,omitempty"`
    Title       string            `json:"title,omitempty"`
    Description string            `json:"description,omitempty"`
    Customer    *InvoiceCustomer  `json:"customer,omitempty"`
}
```

| Field | Type | JSON key | Required | Description |
| --- | --- | --- | --- | --- |
| `Amount` | `float64` | `amount` | Yes | Charge amount in the specified currency |
| `Currency` | `string` | `currency` | Yes | ISO 4217 code (case-insensitive, normalized to uppercase) |
| `SuccessURL` | `string` | `success_url` | No | Redirect URL after successful payment |
| `CancelURL` | `string` | `cancel_url` | No | Redirect URL if customer cancels |
| `WebhookURL` | `string` | `webhook_url` | No | URL for payment confirmation POST callback |
| `Reference` | `string` | `reference` | No | Your internal order/reference identifier |
| `Title` | `string` | `title` | No | Short title shown on the checkout page |
| `Description` | `string` | `description` | No | Longer description shown on the checkout page |
| `Customer` | `*InvoiceCustomer` | `customer` | No | Customer email and optional name |

---

### InvoiceCustomer

Customer information attached to an invoice.

```go
type InvoiceCustomer struct {
    Email string `json:"email"`
    Name  string `json:"name,omitempty"`
}
```

| Field | Type | JSON key | Required | Description |
| --- | --- | --- | --- | --- |
| `Email` | `string` | `email` | Yes (when Customer set) | Customer's email address |
| `Name` | `string` | `name` | No | Customer's display name |

---

### CreateInvoiceResponse

Top-level response from `Invoices.Create`.

```go
type CreateInvoiceResponse struct {
    Success bool                 `json:"success"`
    Data    CreateInvoiceData    `json:"data"`
    Error   *InvoiceErrorDetail  `json:"error,omitempty"`
}
```

| Field | Type | Description |
| --- | --- | --- |
| `Success` | `bool` | `true` on success |
| `Data` | `CreateInvoiceData` | Invoice details |
| `Error` | `*InvoiceErrorDetail` | Populated on API-level errors (alongside a non-nil `error` return) |

---

### CreateInvoiceData

Invoice record returned from `Invoices.Create`.

```go
type CreateInvoiceData struct {
    ID               string           `json:"id"`
    InvoiceID        string           `json:"invoice_id"`
    OriginalAmount   float64          `json:"original_amount"`
    OriginalCurrency string           `json:"original_currency"`
    Amount           float64          `json:"amount"`
    Currency         string           `json:"currency"`
    ExchangeRate     float64          `json:"exchange_rate"`
    WebhookURL       string           `json:"webhook_url"`
    Customer         *InvoiceCustomer `json:"customer,omitempty"`
    CheckoutURL      string           `json:"checkout_url"`
    Status           string           `json:"status"`
    CreatedAt        time.Time        `json:"created_at"`
}
```

| Field | Type | Description |
| --- | --- | --- |
| `ID` | `string` | Internal record ID |
| `InvoiceID` | `string` | Public identifier — use with `Invoices.Verify` |
| `OriginalAmount` | `float64` | Amount from the request |
| `OriginalCurrency` | `string` | Currency from the request (uppercased) |
| `Amount` | `float64` | Settlement amount (may differ due to FX) |
| `Currency` | `string` | Settlement currency |
| `ExchangeRate` | `float64` | Applied exchange rate |
| `WebhookURL` | `string` | Webhook URL echoed from the request |
| `Customer` | `*InvoiceCustomer` | Customer details echoed from the request |
| `CheckoutURL` | `string` | Hosted payment page URL — redirect the customer here |
| `Status` | `string` | Invoice status (e.g., `pending`) |
| `CreatedAt` | `time.Time` | Creation timestamp |

---

### InvoiceErrorDetail

Application-level error detail returned in an API error response body.

```go
type InvoiceErrorDetail struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}
```

| Field | Type | Description |
| --- | --- | --- |
| `Code` | `string` | Application error code from the server |
| `Message` | `string` | Human-readable error message |

---

### VerifyInvoiceResponse

Top-level response from `Invoices.Verify`.

```go
type VerifyInvoiceResponse struct {
    Success bool                `json:"success"`
    Data    VerifyInvoiceData   `json:"data"`
    Error   *InvoiceErrorDetail `json:"error,omitempty"`
}
```

---

### VerifyInvoiceData

Full invoice state returned from `Invoices.Verify`. Contains all `CreateInvoiceData` fields plus:

```go
type VerifyInvoiceData struct {
    // All CreateInvoiceData fields embedded
    ID               string           `json:"id"`
    InvoiceID        string           `json:"invoice_id"`
    OriginalAmount   float64          `json:"original_amount"`
    OriginalCurrency string           `json:"original_currency"`
    Amount           float64          `json:"amount"`
    Currency         string           `json:"currency"`
    ExchangeRate     float64          `json:"exchange_rate"`
    WebhookURL       string           `json:"webhook_url"`
    Customer         *InvoiceCustomer `json:"customer,omitempty"`
    CheckoutURL      string           `json:"checkout_url"`
    Status           string           `json:"status"`
    CreatedAt        time.Time        `json:"created_at"`

    // Additional fields
    Reference        string           `json:"reference"`
    Paid             bool             `json:"paid"`
    Payment          *InvoicePayment  `json:"payment,omitempty"`
}
```

| Additional Field | Type | Description |
| --- | --- | --- |
| `Reference` | `string` | Your reference ID echoed from the original invoice |
| `Paid` | `bool` | `true` when payment is confirmed |
| `Payment` | `*InvoicePayment` | On-chain payment details — only non-nil when `Paid` is `true` |

---

### InvoicePayment

On-chain payment details. Only present in `VerifyInvoiceData` when `Paid == true`.

```go
type InvoicePayment struct {
    ID              string    `json:"id"`
    Network         string    `json:"network"`
    Token           string    `json:"token"`
    Address         string    `json:"address"`
    Amount          float64   `json:"amount"`
    Status          string    `json:"status"`
    TxHash          []string  `json:"tx_hash"`
    TransactionType string    `json:"transaction_type"`
    PayerEmail      string    `json:"payer_email"`
    CreatedAt       time.Time `json:"created_at"`
}
```

| Field | Type | Description |
| --- | --- | --- |
| `ID` | `string` | Internal payment record ID |
| `Network` | `string` | Blockchain network (e.g., `ethereum`, `tron`, `bsc`) |
| `Token` | `string` | Token symbol (e.g., `USDT`, `USDC`) |
| `Address` | `string` | Recipient wallet address |
| `Amount` | `float64` | Amount received on-chain |
| `Status` | `string` | On-chain payment status |
| `TxHash` | `[]string` | Blockchain transaction hash(es) |
| `TransactionType` | `string` | Type of on-chain transaction |
| `PayerEmail` | `string` | Payer's email (if provided at checkout) |
| `CreatedAt` | `time.Time` | Payment creation timestamp |

---

### ListTransactionsParams

Optional pagination parameters for `Transactions.List`. Use pointer fields to distinguish "not set" from zero values.

```go
type ListTransactionsParams struct {
    Page  *int `json:"page,omitempty"`
    Limit *int `json:"limit,omitempty"`
}
```

| Field | Type | Description |
| --- | --- | --- |
| `Page` | `*int` | Page number (1-based). `nil` = use API default (1) |
| `Limit` | `*int` | Records per page. `nil` = use API default |

---

### ListTransactionsResponse

Top-level response from `Transactions.List`.

```go
type ListTransactionsResponse struct {
    Success bool                 `json:"success"`
    Data    ListTransactionsData `json:"data"`
}
```

---

### ListTransactionsData

Transaction list and pagination metadata.

```go
type ListTransactionsData struct {
    Transactions []Transaction `json:"transactions"`
    Pagination   Pagination    `json:"pagination"`
}
```

| Field | Type | Description |
| --- | --- | --- |
| `Transactions` | `[]Transaction` | Slice of transaction records for this page |
| `Pagination` | `Pagination` | Pagination metadata |

---

### Transaction

A settled payment transaction.

```go
type Transaction struct {
    ID               string              `json:"id"`
    InvoiceID        string              `json:"invoice_id"`
    Reference        string              `json:"reference"`
    OriginalAmount   float64             `json:"original_amount"`
    OriginalCurrency string              `json:"original_currency"`
    Amount           float64             `json:"amount"`
    Currency         string              `json:"currency"`
    ExchangeRate     float64             `json:"exchange_rate"`
    Title            string              `json:"title"`
    Description      string              `json:"description"`
    Status           string              `json:"status"`
    Paid             bool                `json:"paid"`
    Customer         TransactionCustomer `json:"customer"`
    CreatedAt        time.Time           `json:"created_at"`
    Payment          TransactionPayment  `json:"payment"`
}
```

| Field | Type | Description |
| --- | --- | --- |
| `ID` | `string` | Internal transaction record ID |
| `InvoiceID` | `string` | Source invoice ID |
| `Reference` | `string` | Your reference ID from the original invoice |
| `OriginalAmount` | `float64` | Invoice amount before FX |
| `OriginalCurrency` | `string` | Invoice currency before FX |
| `Amount` | `float64` | Settled amount |
| `Currency` | `string` | Settlement currency |
| `ExchangeRate` | `float64` | Applied exchange rate |
| `Title` | `string` | Title from the original invoice |
| `Description` | `string` | Description from the original invoice |
| `Status` | `string` | Transaction status |
| `Paid` | `bool` | Always `true` for transactions |
| `Customer` | `TransactionCustomer` | Customer information |
| `CreatedAt` | `time.Time` | Transaction creation timestamp |
| `Payment` | `TransactionPayment` | On-chain payment details |

---

### TransactionCustomer

Customer information on a settled transaction.

```go
type TransactionCustomer struct {
    Email string `json:"email"`
    Name  string `json:"name"`
}
```

---

### TransactionPayment

On-chain payment details for a settled transaction.

```go
type TransactionPayment struct {
    ID              string    `json:"id"`
    Network         string    `json:"network"`
    Token           string    `json:"token"`
    Address         string    `json:"address"`
    Amount          float64   `json:"amount"`
    Status          string    `json:"status"`
    TxHash          []string  `json:"tx_hash"`
    TransactionType string    `json:"transaction_type"`
    PayerEmail      string    `json:"payer_email"`
    CreatedAt       time.Time `json:"created_at"`
}
```

Fields are identical to [InvoicePayment](#invoicepayment).

---

### Pagination

Metadata about the current page in a paginated response.

```go
type Pagination struct {
    Page       int  `json:"page"`
    Limit      int  `json:"limit"`
    Total      int  `json:"total"`
    TotalPages int  `json:"total_pages"`
    HasMore    bool `json:"has_more"`
}
```

| Field | Type | Description |
| --- | --- | --- |
| `Page` | `int` | Current page number (1-based) |
| `Limit` | `int` | Records per page used for this response |
| `Total` | `int` | Total records across all pages |
| `TotalPages` | `int` | Total number of pages |
| `HasMore` | `bool` | `true` when additional pages exist |

---

### SupportedInvoiceCurrencies

A `map[string]bool` of all valid ISO 4217 currency codes (uppercase keys, `true` values). Used internally by `Invoices.Create` for validation. Can be inspected directly:

```go
import "nodela.co/go-sdk/pkg/models"

_, ok := models.SupportedInvoiceCurrencies["NGN"] // true
_, ok  = models.SupportedInvoiceCurrencies["XYZ"] // false
```

Contains 63 entries. See [Supported Currencies](./supported-currencies.md) for the full list.

---

### Metadata

Shared timestamp fields (currently available for internal use).

```go
type Metadata struct {
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

---

### PaginationParams

Generic pagination input parameters.

```go
type PaginationParams struct {
    Page    *int `json:"page,omitempty"`
    PerPage *int `json:"per_page,omitempty"`
}
```

---

### PaginatedResponse

Generic paginated response wrapper (used internally).

```go
type PaginatedResponse[T any] struct {
    Data       []T `json:"data"`
    Page       int `json:"page"`
    PerPage    int `json:"per_page"`
    Total      int `json:"total"`
    TotalPages int `json:"total_pages"`
}
```

---

## Package: errors

Import path: `nodela.co/go-sdk/pkg/errors`

---

### SDKError

Returned for client-side errors: input validation failures and transport-level problems.

```go
type SDKError struct {
    Code    string
    Message string
    Err     error
}

func (e *SDKError) Error() string  // "[CODE] message" or "[CODE] message: underlying"
func (e *SDKError) Unwrap() error  // Returns Err for errors chain compatibility
```

| Field | Type | Description |
| --- | --- | --- |
| `Code` | `string` | One of the `ErrCode*` constants |
| `Message` | `string` | Human-readable description |
| `Err` | `error` | Underlying cause (may be `nil`) |

**Check with `errors.As`:**
```go
var sdkErr *nodela_errors.SDKError
if errors.As(err, &sdkErr) {
    fmt.Println(sdkErr.Code, sdkErr.Message)
}
```

---

### APIError

Returned when the Nodela API responds with a non-2xx HTTP status code.

```go
type APIError struct {
    StatusCode int
    Message    string
    Code       string
}

func (e *APIError) Error() string  // "API error (status XXX): message [code]"
```

| Field | Type | Description |
| --- | --- | --- |
| `StatusCode` | `int` | HTTP status code (e.g., 401, 404, 429, 500) |
| `Message` | `string` | Human-readable error message from the API |
| `Code` | `string` | Application error code from the API response body |

**Check with `errors.As`:**
```go
var apiErr *nodela_errors.APIError
if errors.As(err, &apiErr) {
    fmt.Println(apiErr.StatusCode, apiErr.Message)
}
```

---

### Error Code Constants

Defined in `pkg/errors/errors.go`. Used as the `Code` field in `SDKError`.

```go
const (
    ErrCodeUnknown          = "UNKNOWN"
    ErrCodeNetwork          = "NETWORK"
    ErrCodeSerialization    = "SERIALIZATION"
    ErrCodeDeserialization  = "DESERIALIZATION"
    ErrCodeRequest          = "REQUEST"
    ErrCodeResponse         = "RESPONSE"
    ErrCodeValidation       = "VALIDATION"
    ErrCodeAuthentication   = "AUTHENTICATION"
    ErrCodeAuthorization    = "AUTHORIZATION"
    ErrCodeNotFound         = "NOT_FOUND"
    ErrCodeRateLimit        = "RATE_LIMIT"
)
```

| Constant | Value | Description |
| --- | --- | --- |
| `ErrCodeUnknown` | `"UNKNOWN"` | Unexpected internal condition |
| `ErrCodeNetwork` | `"NETWORK"` | DNS, timeout, or connection failure |
| `ErrCodeSerialization` | `"SERIALIZATION"` | Failed to JSON-marshal the request body |
| `ErrCodeDeserialization` | `"DESERIALIZATION"` | Failed to JSON-unmarshal the response body |
| `ErrCodeRequest` | `"REQUEST"` | Failed to construct the `*http.Request` |
| `ErrCodeResponse` | `"RESPONSE"` | Failed to read the HTTP response body |
| `ErrCodeValidation` | `"VALIDATION"` | Input validation failed (e.g., unsupported currency) |
| `ErrCodeAuthentication` | `"AUTHENTICATION"` | HTTP 401 — invalid or missing credentials |
| `ErrCodeAuthorization` | `"AUTHORIZATION"` | HTTP 403 — insufficient permissions |
| `ErrCodeNotFound` | `"NOT_FOUND"` | HTTP 404 — resource not found |
| `ErrCodeRateLimit` | `"RATE_LIMIT"` | HTTP 429 — rate limit exceeded |

---

### NewSDKError

```go
func NewSDKError(code, message string, err error) *SDKError
```

Constructs a new `SDKError`. The `err` argument may be `nil`.

```go
sdkErr := nodela_errors.NewSDKError(
    nodela_errors.ErrCodeValidation,
    "unsupported currency: XYZ",
    nil,
)
```

---

### NewAPIError

```go
func NewAPIError(statusCode int, message, code string) *APIError
```

Constructs a new `APIError` from an HTTP response.

---

### IsNotFoundError

```go
func IsNotFoundError(err error) bool
```

Returns `true` if `err` is an `*APIError` with `StatusCode == 404`.

```go
if nodela_errors.IsNotFoundError(err) {
    fmt.Println("Resource not found")
}
```

---

### IsAuthenticatedError

```go
func IsAuthenticatedError(err error) bool
```

Returns `true` if `err` is an `*APIError` with `StatusCode == 401`.

```go
if nodela_errors.IsAuthenticatedError(err) {
    log.Fatal("Invalid API key")
}
```

---

### IsRateLimitError

```go
func IsRateLimitError(err error) bool
```

Returns `true` if `err` is an `*APIError` with `StatusCode == 429`.

```go
if nodela_errors.IsRateLimitError(err) {
    time.Sleep(5 * time.Second)
    // retry
}
```

---

## Base URL

All requests are sent to:

```
https://api.nodela.co
```

This is a compiled-in constant and cannot be overridden via options.
