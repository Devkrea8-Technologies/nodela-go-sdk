package models

// ListTransactionsParams holds optional query parameters for the transaction list
// endpoint. All fields are optional; pass nil to [Transactions.List] to use the
// server's default page (page 1, server-default limit).
//
// Use pointer fields so the SDK can distinguish between "not set" and an explicit
// zero value — only non-nil fields are appended to the request query string.
//
// Example — fetch page 2 with 50 items per page:
//
//	page  := 2
//	limit := 50
//	resp, err := c.Transactions.List(ctx, &models.ListTransactionsParams{
//	    Page:  &page,
//	    Limit: &limit,
//	})
type ListTransactionsParams struct {
	// Page is the 1-based page number to retrieve.
	// Omitted from the request query string when nil, causing the server to return
	// the first page.
	Page *int `json:"page,omitempty"`

	// Limit is the maximum number of transactions to return per page.
	// Omitted from the request query string when nil, causing the server to apply
	// its own default page size.
	Limit *int `json:"limit,omitempty"`
}

// TransactionCustomer holds the customer identity information recorded at the
// time a transaction was completed. Unlike [InvoiceCustomer], all fields are
// non-pointer strings because a settled transaction always has a known customer.
type TransactionCustomer struct {
	// Email is the customer's email address as recorded at payment time.
	Email string `json:"email"`

	// Name is the customer's display name as recorded at payment time.
	Name string `json:"name"`
}

// TransactionPayment holds the on-chain payment details recorded for a settled
// transaction. It is structurally similar to [InvoicePayment] but appears in the
// context of the transaction list endpoint rather than the invoice verification
// endpoint.
type TransactionPayment struct {
	// ID is the unique identifier for this payment record within Nodela.
	ID string `json:"id"`

	// Network is the blockchain network on which the payment was settled
	// (e.g. "ethereum", "polygon", "bsc").
	Network string `json:"network"`

	// Token is the cryptocurrency token used for payment (e.g. "USDC", "USDT", "ETH").
	Token string `json:"token"`

	// Address is the on-chain wallet address that received the payment funds.
	Address string `json:"address"`

	// Amount is the total token amount received for this payment.
	Amount float64 `json:"amount"`

	// Status is the current confirmation status of the on-chain payment
	// (e.g. "confirmed", "pending", "failed").
	Status string `json:"status"`

	// TxHash contains the blockchain transaction hash(es) associated with this payment.
	// A slice is used because a single Nodela payment may occasionally span more than
	// one on-chain transaction.
	TxHash []string `json:"tx_hash"`

	// TransactionType describes the settlement mechanism used
	// (e.g. "onchain", "offchain").
	TransactionType string `json:"transaction_type"`

	// PayerEmail is the email address of the payer, captured during the checkout flow.
	PayerEmail string `json:"payer_email"`

	// CreatedAt is the RFC 3339 timestamp of when this payment record was created.
	CreatedAt string `json:"created_at"`
}

// Transaction represents a fully settled payment transaction as recorded by Nodela.
// Each Transaction corresponds to a paid invoice and contains both the invoice
// metadata (amount, currency, customer) and the associated on-chain payment details.
//
// Transactions are returned by [Transactions.List] inside a [ListTransactionsData]
// envelope along with pagination metadata.
type Transaction struct {
	// ID is the unique internal identifier for this transaction within Nodela.
	ID string `json:"id"`

	// InvoiceID is the identifier of the invoice that this transaction settles.
	// Use this to correlate a transaction with a previously created invoice.
	InvoiceID string `json:"invoice_id"`

	// Reference is the caller-supplied reference string from the original invoice
	// creation request. Useful for correlating Nodela transactions with records in
	// your own system (e.g. order IDs). Empty string when no reference was set.
	Reference string `json:"reference"`

	// OriginalAmount is the invoice amount in the currency originally requested by
	// the caller (see OriginalCurrency).
	OriginalAmount float64 `json:"original_amount"`

	// OriginalCurrency is the ISO 4217 currency code originally requested at invoice
	// creation time.
	OriginalCurrency string `json:"original_currency"`

	// Amount is the settled amount in the Nodela settlement currency (see Currency).
	Amount float64 `json:"amount"`

	// Currency is the ISO 4217 settlement currency code used by Nodela to process
	// this transaction.
	Currency string `json:"currency"`

	// ExchangeRate is the conversion rate applied to convert OriginalCurrency to
	// Currency. Will be 1.0 when no conversion was necessary.
	ExchangeRate float64 `json:"exchange_rate"`

	// Title is the short human-readable label that was set on the originating invoice,
	// if any. Empty string when no title was provided.
	Title string `json:"title"`

	// Description is the longer description that was set on the originating invoice,
	// if any. Empty string when no description was provided.
	Description string `json:"description"`

	// Status is the lifecycle status of the transaction (e.g. "completed").
	Status string `json:"status"`

	// Paid reports whether the underlying invoice was fully paid.
	Paid bool `json:"paid"`

	// Customer holds the customer information recorded at payment time.
	Customer TransactionCustomer `json:"customer"`

	// CreatedAt is the RFC 3339 timestamp of when this transaction record was created.
	CreatedAt string `json:"created_at"`

	// Payment holds the on-chain payment details for this transaction.
	Payment TransactionPayment `json:"payment"`
}

// Pagination contains the cursor metadata returned by list endpoints, allowing
// callers to determine the current position within a result set and whether
// additional pages are available.
type Pagination struct {
	// Page is the current 1-based page number returned by the server.
	Page int `json:"page"`

	// Limit is the maximum number of items per page as applied by the server.
	Limit int `json:"limit"`

	// Total is the total number of items across all pages at the current filter.
	Total int `json:"total"`

	// TotalPages is the total number of pages available given the current Limit.
	TotalPages int `json:"total_pages"`

	// HasMore reports whether there is at least one additional page after the current
	// one. Use this to drive pagination loops instead of comparing Page to TotalPages.
	HasMore bool `json:"has_more"`
}

// ListTransactionsData holds the payload of a successful transaction list response.
// It bundles the current page of [Transaction] records with the [Pagination] metadata
// needed to fetch subsequent pages.
type ListTransactionsData struct {
	// Transactions is the slice of Transaction records for the requested page.
	// May be empty when no transactions exist or when the requested page is beyond
	// the last available page.
	Transactions []Transaction `json:"transactions"`

	// Pagination contains the server-side pagination metadata for this response,
	// including total counts and whether more pages are available.
	Pagination Pagination `json:"pagination"`
}

// ListTransactionsResponse is the top-level JSON response envelope returned by the
// transaction list endpoint. Always check Success before accessing Data.
//
// Example — iterate all pages:
//
//	page  := 1
//	limit := 50
//	for {
//	    resp, err := c.Transactions.List(ctx, &models.ListTransactionsParams{
//	        Page: &page, Limit: &limit,
//	    })
//	    if err != nil {
//	        break
//	    }
//	    process(resp.Data.Transactions)
//	    if !resp.Data.Pagination.HasMore {
//	        break
//	    }
//	    page++
//	}
type ListTransactionsResponse struct {
	// Success indicates whether the API request completed successfully.
	// When false the response body will not contain meaningful Data.
	Success bool `json:"success"`

	// Data contains the page of transactions and the associated pagination metadata.
	// Only meaningful when Success is true.
	Data ListTransactionsData `json:"data"`
}
