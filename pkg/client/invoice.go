package client

import (
	"context"
	"fmt"
	"strings"

	"nodela.co/go-sdk/pkg/errors"
	"nodela.co/go-sdk/pkg/models"
)

// invoiceBasePath is the root API path for all invoice endpoints.
const invoiceBasePath = "/v1/invoices"

// Invoices provides methods for creating and verifying Nodela payment invoices.
//
// Access this service via the [Client.Invoices] field on a configured [Client].
// Do not construct Invoices directly.
//
//	resp, err := c.Invoices.Create(ctx, models.CreateInvoiceParams{
//	    Amount:   49.99,
//	    Currency: "USD",
//	})
type Invoices struct {
	client *Client
}

// Create submits a new payment invoice to the Nodela API and returns the created
// invoice details, including a CheckoutURL to redirect the customer to.
//
// # Currency validation
//
// The Currency field of params is validated against [models.SupportedInvoiceCurrencies]
// before any network call is made. Currency matching is case-insensitive — "usd",
// "USD", and "Usd" are all accepted — and the value is normalised to uppercase in
// the outgoing request. If the currency is not supported, Create returns immediately
// with a [*errors.SDKError] carrying code [errors.ErrCodeValidation].
//
// # Return value
//
// On success, the returned [models.CreateInvoiceResponse].Data.CheckoutURL contains
// the Nodela hosted checkout URL. Redirect the customer (or open the URL in a
// WebView) to initiate payment. Use the InvoiceID from the same response with
// [Invoices.Verify] to poll for payment confirmation.
//
// Example:
//
//	successURL := "https://example.com/success"
//	resp, err := c.Invoices.Create(ctx, models.CreateInvoiceParams{
//	    Amount:     99.00,
//	    Currency:   "USD",
//	    SuccessURL: &successURL,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Pay here:", resp.Data.CheckoutURL)
func (s *Invoices) Create(ctx context.Context, params models.CreateInvoiceParams) (*models.CreateInvoiceResponse, error) {
	upper := strings.ToUpper(params.Currency)
	if _, ok := models.SupportedInvoiceCurrencies[upper]; !ok {
		return nil, errors.NewSDKError(
			errors.ErrCodeValidation,
			fmt.Sprintf("unsupported currency: %q", params.Currency),
			nil,
		)
	}
	params.Currency = upper

	var result models.CreateInvoiceResponse
	if err := s.client.Post(ctx, invoiceBasePath, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Verify retrieves the current status of an invoice by its InvoiceID and returns
// the full invoice details including payment confirmation.
//
// # When to call Verify
//
// Call Verify after redirecting a customer to the CheckoutURL returned by [Invoices.Create]
// to determine whether payment has been completed. It is also appropriate to call
// Verify in response to a Nodela webhook notification to fetch the authoritative
// server-side state.
//
// # invoiceID parameter
//
// invoiceID must be the InvoiceID field from a [models.CreateInvoiceData] response
// (not the internal ID field). The InvoiceID is the public-facing identifier
// used in all subsequent calls.
//
// # Determining payment status
//
// The returned [models.VerifyInvoiceData].Paid field is the canonical boolean
// indicator of whether the invoice has been fully paid. When Paid is true, the
// Payment field will be populated with the on-chain payment details including the
// blockchain network, token, transaction hash(es), and payer information.
//
// Example:
//
//	resp, err := c.Invoices.Verify(ctx, "inv_xxxx")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if resp.Data.Paid {
//	    fmt.Printf("Payment confirmed on %s\n", resp.Data.Payment.Network)
//	} else {
//	    fmt.Printf("Invoice status: %s\n", resp.Data.Status)
//	}
func (s *Invoices) Verify(ctx context.Context, invoiceID string) (*models.VerifyInvoiceResponse, error) {
	path := fmt.Sprintf("%s/%s/verify", invoiceBasePath, invoiceID)

	var result models.VerifyInvoiceResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
