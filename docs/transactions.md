# Transactions

The Transactions service lets you retrieve settled transactions on your Nodela account. Use it for reconciliation, reporting, and auditing. Access it via `client.Transactions`.

---

## Overview

A **transaction** represents a completed, settled payment. Once a customer pays an invoice and the payment is confirmed on-chain, a transaction record is created. Transactions are immutable — they cannot be modified after creation.

Key differences between invoices and transactions:

| | Invoice | Transaction |
| --- | --- | --- |
| Created by | Your server (via `Create`) | Nodela (on payment confirmation) |
| Lifecycle | `pending` → `paid` | Always settled |
| Use case | Generate checkout page | Reconciliation, reporting |
| API call | `Invoices.Create` / `Invoices.Verify` | `Transactions.List` |

---

## Methods

### `List`

```go
func (s *Transactions) List(ctx context.Context, params *models.ListTransactionsParams) (*models.ListTransactionsResponse, error)
```

Returns a paginated list of settled transactions. Passing `nil` as `params` returns the first page with the API's default page size.

#### Parameters — `*models.ListTransactionsParams`

Both fields are **optional pointers**. Omitting a field (leaving it `nil`) tells the SDK not to include it in the query string, allowing the API to use its own defaults.

| Field | Type | Description |
| --- | --- | --- |
| `Page` | `*int` | Page number (1-based). Default: `1` |
| `Limit` | `*int` | Maximum number of transactions per page. Default: API default |

Use the `ptr()` helper for inline pointer creation:

```go
func ptr[T any](v T) *T { return &v }

params := &models.ListTransactionsParams{
    Page:  ptr(2),
    Limit: ptr(25),
}
```

#### Return value — `*models.ListTransactionsResponse`

| Field | Type | Description |
| --- | --- | --- |
| `Success` | `bool` | `true` when the request succeeded |
| `Data` | `ListTransactionsData` | Transactions and pagination metadata |

#### `ListTransactionsData` fields

| Field | Type | Description |
| --- | --- | --- |
| `Transactions` | `[]Transaction` | Slice of transaction records for this page |
| `Pagination` | `Pagination` | Pagination metadata |

#### `Pagination` fields

| Field | Type | Description |
| --- | --- | --- |
| `Page` | `int` | Current page number |
| `Limit` | `int` | Records per page used for this response |
| `Total` | `int` | Total number of transactions across all pages |
| `TotalPages` | `int` | Total number of pages |
| `HasMore` | `bool` | `true` if there are additional pages after this one |

#### `Transaction` fields

| Field | Type | Description |
| --- | --- | --- |
| `ID` | `string` | Internal transaction record ID |
| `InvoiceID` | `string` | The invoice that triggered this transaction |
| `Reference` | `string` | Your reference ID (from the original invoice) |
| `OriginalAmount` | `float64` | Amount from the original invoice |
| `OriginalCurrency` | `string` | Currency from the original invoice |
| `Amount` | `float64` | Settled amount (may differ due to FX) |
| `Currency` | `string` | Settlement currency |
| `ExchangeRate` | `float64` | Exchange rate applied |
| `Title` | `string` | Title from the original invoice |
| `Description` | `string` | Description from the original invoice |
| `Status` | `string` | Transaction status (always settled) |
| `Paid` | `bool` | Always `true` for transactions |
| `Customer` | `TransactionCustomer` | Customer details |
| `CreatedAt` | `time.Time` | When the transaction was created |
| `Payment` | `TransactionPayment` | On-chain payment details |

#### `TransactionCustomer` fields

| Field | Type | Description |
| --- | --- | --- |
| `Email` | `string` | Customer email |
| `Name` | `string` | Customer name |

#### `TransactionPayment` fields

| Field | Type | Description |
| --- | --- | --- |
| `ID` | `string` | Internal payment record ID |
| `Network` | `string` | Blockchain network (e.g., `ethereum`, `tron`, `bsc`) |
| `Token` | `string` | Token/asset symbol (e.g., `USDT`, `USDC`) |
| `Address` | `string` | Recipient wallet address |
| `Amount` | `float64` | Amount received on-chain |
| `Status` | `string` | On-chain payment status |
| `TxHash` | `[]string` | One or more blockchain transaction hashes |
| `TransactionType` | `string` | Type of blockchain transaction |
| `PayerEmail` | `string` | Payer's email (if provided) |
| `CreatedAt` | `time.Time` | When the payment was recorded |

---

## Examples

### List all transactions (no pagination params)

```go
resp, err := client.Transactions.List(context.Background(), nil)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Showing %d of %d transactions (page %d of %d)\n",
    len(resp.Data.Transactions),
    resp.Data.Pagination.Total,
    resp.Data.Pagination.Page,
    resp.Data.Pagination.TotalPages,
)

for _, tx := range resp.Data.Transactions {
    fmt.Printf("[%s] %s %.2f %s | Invoice: %s | Ref: %s\n",
        tx.CreatedAt.Format("2006-01-02"),
        tx.ID,
        tx.Amount,
        tx.Currency,
        tx.InvoiceID,
        tx.Reference,
    )
}
```

### Fetch a specific page

```go
func ptr[T any](v T) *T { return &v }

resp, err := client.Transactions.List(context.Background(), &models.ListTransactionsParams{
    Page:  ptr(3),
    Limit: ptr(10),
})
if err != nil {
    log.Fatal(err)
}
```

### Walk all pages

```go
func allTransactions(ctx context.Context, client *nodela.Client) ([]models.Transaction, error) {
    var all []models.Transaction
    page := 1
    limit := 50

    for {
        resp, err := client.Transactions.List(ctx, &models.ListTransactionsParams{
            Page:  &page,
            Limit: &limit,
        })
        if err != nil {
            return nil, fmt.Errorf("page %d: %w", page, err)
        }

        all = append(all, resp.Data.Transactions...)

        if !resp.Data.Pagination.HasMore {
            break
        }
        page++
    }

    return all, nil
}
```

### Access payment details

```go
resp, err := client.Transactions.List(context.Background(), nil)
if err != nil {
    log.Fatal(err)
}

for _, tx := range resp.Data.Transactions {
    fmt.Printf("Invoice: %s\n", tx.InvoiceID)
    fmt.Printf("  Network:  %s\n", tx.Payment.Network)
    fmt.Printf("  Token:    %s\n", tx.Payment.Token)
    fmt.Printf("  Amount:   %f\n", tx.Payment.Amount)
    fmt.Printf("  Tx hash:  %v\n", tx.Payment.TxHash)
    fmt.Printf("  Customer: %s <%s>\n", tx.Customer.Name, tx.Customer.Email)
    fmt.Println()
}
```

### Export to CSV

```go
import (
    "encoding/csv"
    "os"
    "strconv"
)

func exportCSV(transactions []models.Transaction, filename string) error {
    f, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer f.Close()

    w := csv.NewWriter(f)
    defer w.Flush()

    // Header
    w.Write([]string{
        "ID", "InvoiceID", "Reference", "Amount", "Currency",
        "OriginalAmount", "OriginalCurrency", "Customer", "Email",
        "Network", "Token", "TxHash", "CreatedAt",
    })

    for _, tx := range transactions {
        w.Write([]string{
            tx.ID,
            tx.InvoiceID,
            tx.Reference,
            strconv.FormatFloat(tx.Amount, 'f', 2, 64),
            tx.Currency,
            strconv.FormatFloat(tx.OriginalAmount, 'f', 2, 64),
            tx.OriginalCurrency,
            tx.Customer.Name,
            tx.Customer.Email,
            tx.Payment.Network,
            tx.Payment.Token,
            strings.Join(tx.Payment.TxHash, ";"),
            tx.CreatedAt.Format(time.RFC3339),
        })
    }

    return w.Error()
}
```

---

## Pagination Details

### How it works

The API uses **page-based pagination**:

- `Page` is 1-indexed (page 1 is the first page).
- `Limit` controls the maximum number of records returned per page.
- `HasMore` is `true` when there are more pages after the current one.
- `TotalPages` lets you pre-calculate the number of requests needed.
- `Total` gives the full record count regardless of paging.

### Only setting Page (no Limit)

```go
resp, _ := client.Transactions.List(ctx, &models.ListTransactionsParams{
    Page: ptr(2),
    // Limit not set — API uses its default
})
```

### Only setting Limit (no Page)

```go
resp, _ := client.Transactions.List(ctx, &models.ListTransactionsParams{
    Limit: ptr(5),
    // Page not set — API defaults to page 1
})
```

### Empty result set

When there are no transactions, `Transactions` is an empty slice (not `nil`). Check `len`:

```go
if len(resp.Data.Transactions) == 0 {
    fmt.Println("No transactions found")
}
```

---

## Error Scenarios

| Error | When | How to handle |
| --- | --- | --- |
| `APIError` (401) | Invalid API key | Verify `NODELA_API_KEY` |
| `APIError` (403) | Key lacks permission to read transactions | Use a key with transaction read access |
| `APIError` (429) | Rate limit exceeded | Back off and retry |
| `APIError` (500) | Nodela server error | Retry with exponential backoff |
| `SDKError` (ErrCodeNetwork) | Network failure | Check connectivity and retry |

---

## Next Steps

- [Error Handling](./error-handling.md) — handle API and network errors.
- [Invoices](./invoices.md) — create invoices that generate transactions.
- [API Reference](./api-reference.md) — full type definitions.
