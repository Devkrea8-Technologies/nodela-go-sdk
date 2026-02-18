package models

// SupportedInvoiceCurrencies is the complete set of ISO 4217 currency codes
// accepted by the Nodela invoice API. Map keys are uppercase three-letter codes
// (e.g. "USD", "EUR", "NGN"). The map values are empty structs; use a membership
// check to test support:
//
//	if _, ok := models.SupportedInvoiceCurrencies["GBP"]; ok {
//	    // GBP is supported
//	}
//
// This map is used internally by [Invoices.Create] to validate and reject
// unsupported currencies before making a network call. Callers may also consult
// it at runtime to enumerate or display the supported currency list.
//
// Supported currencies by region:
//
//	Americas:    USD, CAD, MXN, BRL, ARS, CLP, COP, PEN, JMD, TTD
//	Europe:      EUR, GBP, CHF, SEK, NOK, DKK, PLN, CZK, HUF, RON,
//	             BGN, HRK, ISK, TRY, RUB, UAH
//	Africa:      NGN, ZAR, KES, GHS, EGP, MAD, TZS, UGX, XOF, XAF, ETB
//	Asia:        JPY, CNY, INR, KRW, IDR, MYR, THB, PHP, VND, SGD,
//	             HKD, TWD, BDT, PKR, LKR
//	Middle East: AED, SAR, QAR, KWD, BHD, OMR, ILS, JOD
//	Oceania:     AUD, NZD, FJD
var SupportedInvoiceCurrencies = map[string]struct{}{
	// Americas
	"USD": {}, "CAD": {}, "MXN": {}, "BRL": {}, "ARS": {}, "CLP": {}, "COP": {},
	"PEN": {}, "JMD": {}, "TTD": {},
	// Europe
	"EUR": {}, "GBP": {}, "CHF": {}, "SEK": {}, "NOK": {}, "DKK": {}, "PLN": {},
	"CZK": {}, "HUF": {}, "RON": {}, "BGN": {}, "HRK": {}, "ISK": {}, "TRY": {},
	"RUB": {}, "UAH": {},
	// Africa
	"NGN": {}, "ZAR": {}, "KES": {}, "GHS": {}, "EGP": {}, "MAD": {}, "TZS": {},
	"UGX": {}, "XOF": {}, "XAF": {}, "ETB": {},
	// Asia
	"JPY": {}, "CNY": {}, "INR": {}, "KRW": {}, "IDR": {}, "MYR": {}, "THB": {},
	"PHP": {}, "VND": {}, "SGD": {}, "HKD": {}, "TWD": {}, "BDT": {}, "PKR": {},
	"LKR": {},
	// Middle East
	"AED": {}, "SAR": {}, "QAR": {}, "KWD": {}, "BHD": {}, "OMR": {}, "ILS": {},
	"JOD": {},
	// Oceania
	"AUD": {}, "NZD": {}, "FJD": {},
}

// InvoiceCustomer contains optional customer identity information that can be
// associated with an invoice at creation time. When provided, Nodela may
// pre-fill checkout form fields and use the email address for payment
// confirmation notifications.
type InvoiceCustomer struct {
	// Email is the customer's email address. Used for payment notifications and
	// pre-filling the checkout form.
	Email string `json:"email"`

	// Name is the customer's display name shown on the checkout page. Optional;
	// omitted from the request when nil.
	Name *string `json:"name,omitempty"`
}

// CreateInvoiceParams holds the parameters for creating a new invoice via the
// Nodela API. Amount and Currency are required; all other fields are optional.
//
// Example — minimal invoice:
//
//	params := models.CreateInvoiceParams{
//	    Amount:   49.99,
//	    Currency: "USD",
//	}
//
// Example — full invoice with customer and redirect URLs:
//
//	successURL := "https://example.com/success"
//	cancelURL  := "https://example.com/cancel"
//	title      := "Order #1042"
//	params := models.CreateInvoiceParams{
//	    Amount:     49.99,
//	    Currency:   "USD",
//	    SuccessURL: &successURL,
//	    CancelURL:  &cancelURL,
//	    Title:      &title,
//	    Customer:   &models.InvoiceCustomer{Email: "alice@example.com"},
//	}
type CreateInvoiceParams struct {
	// Amount is the invoice total expressed in the Currency. Must be a positive value.
	Amount float64 `json:"amount"`

	// Currency is the ISO 4217 currency code for the invoice amount (e.g. "USD", "EUR").
	// The value is case-insensitive; the SDK normalises it to uppercase before sending
	// the request. Must be present in [SupportedInvoiceCurrencies] or [Invoices.Create]
	// will return a validation error without making a network call.
	Currency string `json:"currency"`

	// SuccessURL is the URL that Nodela will redirect the customer to after a
	// successful payment. Optional.
	SuccessURL *string `json:"success_url,omitempty"`

	// CancelURL is the URL that Nodela will redirect the customer to if they abandon
	// or cancel the checkout flow. Optional.
	CancelURL *string `json:"cancel_url,omitempty"`

	// WebhookURL is the HTTPS endpoint that Nodela will POST real-time payment event
	// notifications to. Use this to receive server-side confirmation of payment without
	// polling. Optional.
	WebhookURL *string `json:"webhook_url,omitempty"`

	// Reference is a caller-defined identifier for the invoice — for example an order
	// ID or internal transaction reference. It is stored by Nodela and returned
	// as-is on subsequent calls, making it easy to correlate Nodela invoices with
	// records in your own system. Optional.
	Reference *string `json:"reference,omitempty"`

	// Customer contains optional customer identity information to associate with the
	// invoice. When provided, Nodela may pre-fill the checkout form and use the email
	// address for payment notifications. Optional.
	Customer *InvoiceCustomer `json:"customer,omitempty"`

	// Title is a short, human-readable label for the invoice displayed prominently on
	// the Nodela checkout page (e.g. "Order #1042" or "Pro Plan Subscription"). Optional.
	Title *string `json:"title,omitempty"`

	// Description provides additional context about the invoice shown beneath the title
	// on the checkout page (e.g. item details or billing period). Optional.
	Description *string `json:"description,omitempty"`
}

// InvoiceErrorDetail holds the structured error information returned by the
// Nodela API within an invoice response envelope when the Success field is false.
// It gives callers both a machine-readable code and a human-readable message.
type InvoiceErrorDetail struct {
	// Code is the machine-readable error code returned by the API.
	Code string `json:"code"`

	// Message is the human-readable description of the error returned by the API.
	Message string `json:"message"`
}

// CreateInvoiceData contains the invoice details returned by the Nodela API after
// a successful invoice creation request. The most important field for most callers
// is CheckoutURL, which is the URL to redirect the end user to in order to complete
// payment.
type CreateInvoiceData struct {
	// ID is the unique internal identifier assigned to this invoice record by Nodela.
	ID string `json:"id"`

	// InvoiceID is the public-facing invoice identifier used in subsequent API calls.
	// Pass this value to [Invoices.Verify] to retrieve the current invoice status.
	InvoiceID string `json:"invoice_id"`

	// OriginalAmount is the invoice amount as originally specified by the caller,
	// preserved as a string to avoid floating-point rounding issues.
	OriginalAmount string `json:"original_amount"`

	// OriginalCurrency is the ISO 4217 currency code that was provided in the
	// creation request.
	OriginalCurrency string `json:"original_currency"`

	// Amount is the equivalent amount in the settlement currency used by Nodela,
	// represented as a string.
	Amount string `json:"amount"`

	// Currency is the ISO 4217 settlement currency code used by Nodela to process
	// this invoice.
	Currency string `json:"currency"`

	// ExchangeRate is the conversion rate applied when OriginalCurrency differs from
	// Currency. Nil when no conversion was necessary.
	ExchangeRate *string `json:"exchange_rate,omitempty"`

	// WebhookURL is the webhook endpoint registered for this invoice, echoed back
	// from the creation request. Nil when no webhook was provided.
	WebhookURL *string `json:"webhook_url,omitempty"`

	// Customer holds the customer information associated with this invoice, echoed
	// back from the creation request. Nil when no customer was provided.
	Customer *InvoiceCustomer `json:"customer,omitempty"`

	// CheckoutURL is the URL to redirect the end user to in order to complete payment.
	// This is the primary output of invoice creation; send the customer here.
	CheckoutURL string `json:"checkout_url"`

	// Status is the current lifecycle status of the invoice (e.g. "pending", "paid",
	// "expired"). May be absent on freshly created invoices.
	Status *string `json:"status,omitempty"`

	// CreatedAt is the RFC 3339 timestamp of when the invoice was created on the server.
	CreatedAt string `json:"created_at"`
}

// CreateInvoiceResponse is the top-level JSON response envelope returned by the
// invoice creation endpoint. Always check Success before accessing Data.
//
// When Success is true, Data contains the created invoice details including the
// CheckoutURL. When Success is false, Error contains the API error detail.
type CreateInvoiceResponse struct {
	// Success indicates whether the API request completed successfully.
	// When false, inspect Error for details; when true, inspect Data.
	Success bool `json:"success"`

	// Data contains the created invoice details. Present only when Success is true.
	Data *CreateInvoiceData `json:"data,omitempty"`

	// Error contains the structured error information from the API.
	// Present only when Success is false.
	Error *InvoiceErrorDetail `json:"error,omitempty"`
}

// InvoicePayment represents the on-chain cryptocurrency payment record associated
// with a paid or partially-paid invoice. It captures the blockchain-level details
// of how the invoice was settled, as recorded by the Nodela payment processor.
type InvoicePayment struct {
	// ID is the unique identifier for this payment record within Nodela.
	ID string `json:"id"`

	// Network is the blockchain network on which the payment was made
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

	// PayerEmail is the email address of the payer, captured during the checkout flow
	// if the customer provided one.
	PayerEmail string `json:"payer_email"`

	// CreatedAt is the RFC 3339 timestamp of when this payment record was created.
	CreatedAt string `json:"created_at"`
}

// VerifyInvoiceData contains the full invoice details returned when verifying an
// invoice's status. It includes all fields from creation as well as live status
// information. When payment has been received, the Payment field is populated with
// the associated on-chain payment record.
type VerifyInvoiceData struct {
	// ID is the unique internal identifier for this invoice record.
	ID string `json:"id"`

	// InvoiceID is the public-facing invoice identifier, matching the value used
	// in the verification request.
	InvoiceID string `json:"invoice_id"`

	// Reference is the caller-supplied reference string from invoice creation, if
	// one was provided. Nil when no reference was set.
	Reference *string `json:"reference,omitempty"`

	// OriginalAmount is the invoice amount as specified at creation time, preserved
	// as a string to avoid floating-point rounding issues.
	OriginalAmount string `json:"original_amount"`

	// OriginalCurrency is the ISO 4217 currency code provided at creation time.
	OriginalCurrency string `json:"original_currency"`

	// Amount is the equivalent amount in the settlement currency.
	Amount float64 `json:"amount"`

	// Currency is the ISO 4217 settlement currency code.
	Currency string `json:"currency"`

	// ExchangeRate is the conversion rate applied between OriginalCurrency and
	// Currency. Nil when both currencies are the same and no conversion was needed.
	ExchangeRate *float64 `json:"exchange_rate,omitempty"`

	// Title is the short human-readable label set on the invoice at creation time.
	// Nil when no title was provided.
	Title *string `json:"title,omitempty"`

	// Description is the longer description set on the invoice at creation time.
	// Nil when no description was provided.
	Description *string `json:"description,omitempty"`

	// Status is the current lifecycle status of the invoice. Common values include
	// "pending", "paid", "expired", and "cancelled".
	Status string `json:"status"`

	// Paid reports whether the invoice has been fully paid. When true, the Payment
	// field will contain the associated on-chain payment details.
	Paid bool `json:"paid"`

	// Customer holds the customer information associated with the invoice, if any
	// was provided at creation time. Nil when no customer was set.
	Customer *InvoiceCustomer `json:"customer,omitempty"`

	// CreatedAt is the RFC 3339 timestamp of when the invoice was originally created.
	CreatedAt string `json:"created_at"`

	// Payment holds the on-chain payment details when a payment has been recorded for
	// this invoice. Nil when the invoice has not yet been paid.
	Payment *InvoicePayment `json:"payment,omitempty"`
}

// VerifyInvoiceResponse is the top-level JSON response envelope returned by the
// invoice verification endpoint. Always check Success before accessing Data.
//
// When Success is true, Data contains the full invoice state including Paid status
// and, if applicable, the on-chain Payment record. When Success is false, Error
// contains the API error detail.
type VerifyInvoiceResponse struct {
	// Success indicates whether the API request completed successfully.
	// When false, inspect Error for details; when true, inspect Data.
	Success bool `json:"success"`

	// Data contains the full invoice details. Present only when Success is true.
	Data *VerifyInvoiceData `json:"data,omitempty"`

	// Error contains the structured error information from the API.
	// Present only when Success is false.
	Error *InvoiceErrorDetail `json:"error,omitempty"`
}
