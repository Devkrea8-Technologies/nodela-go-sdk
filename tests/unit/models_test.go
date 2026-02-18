package unit

import (
	"encoding/json"
	"testing"

	"nodela.co/go-sdk/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// SupportedInvoiceCurrencies
// ---------------------------------------------------------------------------

func TestSupportedInvoiceCurrencies_NotEmpty(t *testing.T) {
	assert.NotEmpty(t, models.SupportedInvoiceCurrencies)
}

func TestSupportedInvoiceCurrencies_Americas(t *testing.T) {
	americas := []string{"USD", "CAD", "MXN", "BRL", "ARS", "CLP", "COP", "PEN", "JMD", "TTD"}
	for _, c := range americas {
		_, ok := models.SupportedInvoiceCurrencies[c]
		assert.True(t, ok, "expected Americas currency %q to be supported", c)
	}
}

func TestSupportedInvoiceCurrencies_Europe(t *testing.T) {
	europe := []string{
		"EUR", "GBP", "CHF", "SEK", "NOK", "DKK", "PLN",
		"CZK", "HUF", "RON", "BGN", "HRK", "ISK", "TRY",
		"RUB", "UAH",
	}
	for _, c := range europe {
		_, ok := models.SupportedInvoiceCurrencies[c]
		assert.True(t, ok, "expected European currency %q to be supported", c)
	}
}

func TestSupportedInvoiceCurrencies_Africa(t *testing.T) {
	africa := []string{
		"NGN", "ZAR", "KES", "GHS", "EGP", "MAD", "TZS",
		"UGX", "XOF", "XAF", "ETB",
	}
	for _, c := range africa {
		_, ok := models.SupportedInvoiceCurrencies[c]
		assert.True(t, ok, "expected African currency %q to be supported", c)
	}
}

func TestSupportedInvoiceCurrencies_Asia(t *testing.T) {
	asia := []string{
		"JPY", "CNY", "INR", "KRW", "IDR", "MYR", "THB",
		"PHP", "VND", "SGD", "HKD", "TWD", "BDT", "PKR", "LKR",
	}
	for _, c := range asia {
		_, ok := models.SupportedInvoiceCurrencies[c]
		assert.True(t, ok, "expected Asian currency %q to be supported", c)
	}
}

func TestSupportedInvoiceCurrencies_MiddleEast(t *testing.T) {
	me := []string{"AED", "SAR", "QAR", "KWD", "BHD", "OMR", "ILS", "JOD"}
	for _, c := range me {
		_, ok := models.SupportedInvoiceCurrencies[c]
		assert.True(t, ok, "expected Middle East currency %q to be supported", c)
	}
}

func TestSupportedInvoiceCurrencies_Oceania(t *testing.T) {
	oceania := []string{"AUD", "NZD", "FJD"}
	for _, c := range oceania {
		_, ok := models.SupportedInvoiceCurrencies[c]
		assert.True(t, ok, "expected Oceania currency %q to be supported", c)
	}
}

func TestSupportedInvoiceCurrencies_DoesNotContainInvalidCodes(t *testing.T) {
	invalid := []string{"", "XX", "USDT", "BTC", "ETH", "xtz", "usd"}
	for _, c := range invalid {
		_, ok := models.SupportedInvoiceCurrencies[c]
		assert.False(t, ok, "expected %q NOT to be a supported currency", c)
	}
}

// ---------------------------------------------------------------------------
// InvoiceCustomer
// ---------------------------------------------------------------------------

func TestInvoiceCustomer_JSON_WithName(t *testing.T) {
	customer := models.InvoiceCustomer{
		Email: "alice@example.com",
		Name:  ptr("Alice"),
	}

	data, err := json.Marshal(customer)
	require.NoError(t, err)

	var out models.InvoiceCustomer
	require.NoError(t, json.Unmarshal(data, &out))

	assert.Equal(t, "alice@example.com", out.Email)
	require.NotNil(t, out.Name)
	assert.Equal(t, "Alice", *out.Name)
}

func TestInvoiceCustomer_JSON_WithoutName(t *testing.T) {
	customer := models.InvoiceCustomer{Email: "bob@example.com"}

	data, err := json.Marshal(customer)
	require.NoError(t, err)

	// omitempty: Name must be absent from JSON
	assert.NotContains(t, string(data), "name")

	var out models.InvoiceCustomer
	require.NoError(t, json.Unmarshal(data, &out))
	assert.Equal(t, "bob@example.com", out.Email)
	assert.Nil(t, out.Name)
}

// ---------------------------------------------------------------------------
// CreateInvoiceParams
// ---------------------------------------------------------------------------

func TestCreateInvoiceParams_JSON_MinimalParams(t *testing.T) {
	params := models.CreateInvoiceParams{
		Amount:   100.50,
		Currency: "USD",
	}

	data, err := json.Marshal(params)
	require.NoError(t, err)

	jsonStr := string(data)
	assert.Contains(t, jsonStr, `"amount":100.5`)
	assert.Contains(t, jsonStr, `"currency":"USD"`)
	// Optional fields must be omitted
	assert.NotContains(t, jsonStr, "success_url")
	assert.NotContains(t, jsonStr, "cancel_url")
	assert.NotContains(t, jsonStr, "webhook_url")
	assert.NotContains(t, jsonStr, "reference")
	assert.NotContains(t, jsonStr, "customer")
	assert.NotContains(t, jsonStr, "title")
	assert.NotContains(t, jsonStr, "description")
}

func TestCreateInvoiceParams_JSON_AllFields(t *testing.T) {
	successURL := "https://example.com/success"
	cancelURL := "https://example.com/cancel"
	webhookURL := "https://example.com/webhook"
	ref := "order-123"
	title := "Test Order"
	desc := "Payment for test order"

	params := models.CreateInvoiceParams{
		Amount:      250.00,
		Currency:    "GBP",
		SuccessURL:  &successURL,
		CancelURL:   &cancelURL,
		WebhookURL:  &webhookURL,
		Reference:   &ref,
		Customer:    &models.InvoiceCustomer{Email: "user@example.com", Name: ptr("User")},
		Title:       &title,
		Description: &desc,
	}

	data, err := json.Marshal(params)
	require.NoError(t, err)

	jsonStr := string(data)
	assert.Contains(t, jsonStr, `"amount":250`)
	assert.Contains(t, jsonStr, `"currency":"GBP"`)
	assert.Contains(t, jsonStr, `"success_url"`)
	assert.Contains(t, jsonStr, `"cancel_url"`)
	assert.Contains(t, jsonStr, `"webhook_url"`)
	assert.Contains(t, jsonStr, `"reference":"order-123"`)
	assert.Contains(t, jsonStr, `"customer"`)
	assert.Contains(t, jsonStr, `"title":"Test Order"`)
	assert.Contains(t, jsonStr, `"description":"Payment for test order"`)
}

// ---------------------------------------------------------------------------
// CreateInvoiceResponse
// ---------------------------------------------------------------------------

func TestCreateInvoiceResponse_Unmarshal_Success(t *testing.T) {
	raw := `{
		"success": true,
		"data": {
			"id": "inv_abc123",
			"invoice_id": "INV-001",
			"original_amount": "100.00",
			"original_currency": "USD",
			"amount": "0.0025",
			"currency": "BTC",
			"exchange_rate": "40000",
			"checkout_url": "https://checkout.nodela.co/inv_abc123",
			"status": "pending",
			"created_at": "2024-01-01T00:00:00Z"
		}
	}`

	var resp models.CreateInvoiceResponse
	require.NoError(t, json.Unmarshal([]byte(raw), &resp))

	assert.True(t, resp.Success)
	require.NotNil(t, resp.Data)
	assert.Equal(t, "inv_abc123", resp.Data.ID)
	assert.Equal(t, "INV-001", resp.Data.InvoiceID)
	assert.Equal(t, "100.00", resp.Data.OriginalAmount)
	assert.Equal(t, "USD", resp.Data.OriginalCurrency)
	assert.Equal(t, "0.0025", resp.Data.Amount)
	assert.Equal(t, "BTC", resp.Data.Currency)
	require.NotNil(t, resp.Data.ExchangeRate)
	assert.Equal(t, "40000", *resp.Data.ExchangeRate)
	assert.Equal(t, "https://checkout.nodela.co/inv_abc123", resp.Data.CheckoutURL)
	require.NotNil(t, resp.Data.Status)
	assert.Equal(t, "pending", *resp.Data.Status)
	assert.Nil(t, resp.Error)
}

func TestCreateInvoiceResponse_Unmarshal_Error(t *testing.T) {
	raw := `{
		"success": false,
		"error": {
			"code": "VALIDATION_ERROR",
			"message": "amount must be positive"
		}
	}`

	var resp models.CreateInvoiceResponse
	require.NoError(t, json.Unmarshal([]byte(raw), &resp))

	assert.False(t, resp.Success)
	assert.Nil(t, resp.Data)
	require.NotNil(t, resp.Error)
	assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	assert.Equal(t, "amount must be positive", resp.Error.Message)
}

// ---------------------------------------------------------------------------
// VerifyInvoiceResponse
// ---------------------------------------------------------------------------

func TestVerifyInvoiceResponse_Unmarshal_Paid(t *testing.T) {
	raw := `{
		"success": true,
		"data": {
			"id": "inv_abc123",
			"invoice_id": "INV-001",
			"original_amount": "100.00",
			"original_currency": "USD",
			"amount": 0.0025,
			"currency": "BTC",
			"exchange_rate": 40000.0,
			"status": "completed",
			"paid": true,
			"created_at": "2024-01-01T00:00:00Z",
			"payment": {
				"id": "pay_xyz",
				"network": "bitcoin",
				"token": "BTC",
				"address": "1A2B3C4D",
				"amount": 0.0025,
				"status": "confirmed",
				"tx_hash": ["abc123def456"],
				"transaction_type": "onchain",
				"payer_email": "payer@example.com",
				"created_at": "2024-01-01T01:00:00Z"
			}
		}
	}`

	var resp models.VerifyInvoiceResponse
	require.NoError(t, json.Unmarshal([]byte(raw), &resp))

	assert.True(t, resp.Success)
	require.NotNil(t, resp.Data)
	assert.Equal(t, "inv_abc123", resp.Data.ID)
	assert.Equal(t, "completed", resp.Data.Status)
	assert.True(t, resp.Data.Paid)
	require.NotNil(t, resp.Data.Payment)
	assert.Equal(t, "pay_xyz", resp.Data.Payment.ID)
	assert.Equal(t, []string{"abc123def456"}, resp.Data.Payment.TxHash)
}

func TestVerifyInvoiceResponse_Unmarshal_Unpaid(t *testing.T) {
	raw := `{
		"success": true,
		"data": {
			"id": "inv_abc123",
			"invoice_id": "INV-001",
			"original_amount": "100.00",
			"original_currency": "USD",
			"amount": 0.0025,
			"currency": "BTC",
			"status": "pending",
			"paid": false,
			"created_at": "2024-01-01T00:00:00Z"
		}
	}`

	var resp models.VerifyInvoiceResponse
	require.NoError(t, json.Unmarshal([]byte(raw), &resp))

	assert.True(t, resp.Success)
	require.NotNil(t, resp.Data)
	assert.Equal(t, "pending", resp.Data.Status)
	assert.False(t, resp.Data.Paid)
	assert.Nil(t, resp.Data.Payment)
}

func TestVerifyInvoiceResponse_Unmarshal_WithOptionalFields(t *testing.T) {
	ref := "order-456"
	title := "Subscription"
	raw := `{
		"success": true,
		"data": {
			"id": "inv_xyz",
			"invoice_id": "INV-002",
			"reference": "order-456",
			"original_amount": "50.00",
			"original_currency": "EUR",
			"amount": 55.0,
			"currency": "USD",
			"title": "Subscription",
			"status": "pending",
			"paid": false,
			"created_at": "2024-06-01T00:00:00Z"
		}
	}`

	var resp models.VerifyInvoiceResponse
	require.NoError(t, json.Unmarshal([]byte(raw), &resp))

	require.NotNil(t, resp.Data.Reference)
	assert.Equal(t, ref, *resp.Data.Reference)
	require.NotNil(t, resp.Data.Title)
	assert.Equal(t, title, *resp.Data.Title)
}

// ---------------------------------------------------------------------------
// ListTransactionsParams
// ---------------------------------------------------------------------------

func TestListTransactionsParams_JSON_Empty(t *testing.T) {
	params := models.ListTransactionsParams{}
	data, err := json.Marshal(params)
	require.NoError(t, err)

	// Both optional fields must be omitted
	assert.NotContains(t, string(data), "page")
	assert.NotContains(t, string(data), "limit")
}

func TestListTransactionsParams_JSON_WithValues(t *testing.T) {
	params := models.ListTransactionsParams{
		Page:  ptr(2),
		Limit: ptr(25),
	}
	data, err := json.Marshal(params)
	require.NoError(t, err)

	jsonStr := string(data)
	assert.Contains(t, jsonStr, `"page":2`)
	assert.Contains(t, jsonStr, `"limit":25`)
}

// ---------------------------------------------------------------------------
// ListTransactionsResponse
// ---------------------------------------------------------------------------

func TestListTransactionsResponse_Unmarshal(t *testing.T) {
	raw := `{
		"success": true,
		"data": {
			"transactions": [
				{
					"id": "txn_001",
					"invoice_id": "INV-001",
					"reference": "ref-001",
					"original_amount": 99.99,
					"original_currency": "USD",
					"amount": 99.99,
					"currency": "USD",
					"exchange_rate": 1.0,
					"title": "Order 001",
					"description": "Test transaction",
					"status": "completed",
					"paid": true,
					"customer": {"email": "cust@example.com", "name": "Customer"},
					"created_at": "2024-01-01T00:00:00Z",
					"payment": {
						"id": "pay_001",
						"network": "polygon",
						"token": "USDC",
						"address": "0xABC",
						"amount": 99.99,
						"status": "confirmed",
						"tx_hash": ["0xDEF"],
						"transaction_type": "onchain",
						"payer_email": "payer@example.com",
						"created_at": "2024-01-01T01:00:00Z"
					}
				}
			],
			"pagination": {
				"page": 1,
				"limit": 10,
				"total": 1,
				"total_pages": 1,
				"has_more": false
			}
		}
	}`

	var resp models.ListTransactionsResponse
	require.NoError(t, json.Unmarshal([]byte(raw), &resp))

	assert.True(t, resp.Success)
	require.Len(t, resp.Data.Transactions, 1)

	txn := resp.Data.Transactions[0]
	assert.Equal(t, "txn_001", txn.ID)
	assert.Equal(t, "INV-001", txn.InvoiceID)
	assert.Equal(t, 99.99, txn.OriginalAmount)
	assert.True(t, txn.Paid)
	assert.Equal(t, "cust@example.com", txn.Customer.Email)
	assert.Equal(t, "Customer", txn.Customer.Name)
	assert.Equal(t, []string{"0xDEF"}, txn.Payment.TxHash)

	pag := resp.Data.Pagination
	assert.Equal(t, 1, pag.Page)
	assert.Equal(t, 10, pag.Limit)
	assert.Equal(t, 1, pag.Total)
	assert.Equal(t, 1, pag.TotalPages)
	assert.False(t, pag.HasMore)
}

func TestListTransactionsResponse_Unmarshal_EmptyList(t *testing.T) {
	raw := `{
		"success": true,
		"data": {
			"transactions": [],
			"pagination": {
				"page": 1,
				"limit": 10,
				"total": 0,
				"total_pages": 0,
				"has_more": false
			}
		}
	}`

	var resp models.ListTransactionsResponse
	require.NoError(t, json.Unmarshal([]byte(raw), &resp))

	assert.True(t, resp.Success)
	assert.Empty(t, resp.Data.Transactions)
	assert.Equal(t, 0, resp.Data.Pagination.Total)
}

func TestListTransactionsResponse_Unmarshal_MultipleTransactions(t *testing.T) {
	raw := `{
		"success": true,
		"data": {
			"transactions": [
				{"id": "t1", "invoice_id": "i1", "reference": "", "original_amount": 10.0, "original_currency": "USD", "amount": 10.0, "currency": "USD", "exchange_rate": 1.0, "title": "", "description": "", "status": "completed", "paid": true, "customer": {"email": "a@a.com", "name": "A"}, "created_at": "", "payment": {"id": "p1", "network": "polygon", "token": "USDC", "address": "0x1", "amount": 10.0, "status": "confirmed", "tx_hash": [], "transaction_type": "onchain", "payer_email": "", "created_at": ""}},
				{"id": "t2", "invoice_id": "i2", "reference": "", "original_amount": 20.0, "original_currency": "EUR", "amount": 22.0, "currency": "USD", "exchange_rate": 1.1, "title": "", "description": "", "status": "pending", "paid": false, "customer": {"email": "b@b.com", "name": "B"}, "created_at": "", "payment": {"id": "p2", "network": "ethereum", "token": "ETH", "address": "0x2", "amount": 0.01, "status": "pending", "tx_hash": [], "transaction_type": "onchain", "payer_email": "", "created_at": ""}}
			],
			"pagination": {
				"page": 1, "limit": 10, "total": 2, "total_pages": 1, "has_more": false
			}
		}
	}`

	var resp models.ListTransactionsResponse
	require.NoError(t, json.Unmarshal([]byte(raw), &resp))

	assert.Len(t, resp.Data.Transactions, 2)
	assert.Equal(t, 2, resp.Data.Pagination.Total)
}

// ---------------------------------------------------------------------------
// Pagination
// ---------------------------------------------------------------------------

func TestPagination_HasMore_True(t *testing.T) {
	raw := `{"page":1,"limit":5,"total":12,"total_pages":3,"has_more":true}`
	var p models.Pagination
	require.NoError(t, json.Unmarshal([]byte(raw), &p))

	assert.Equal(t, 1, p.Page)
	assert.Equal(t, 5, p.Limit)
	assert.Equal(t, 12, p.Total)
	assert.Equal(t, 3, p.TotalPages)
	assert.True(t, p.HasMore)
}

func TestPagination_HasMore_False(t *testing.T) {
	raw := `{"page":3,"limit":5,"total":12,"total_pages":3,"has_more":false}`
	var p models.Pagination
	require.NoError(t, json.Unmarshal([]byte(raw), &p))

	assert.Equal(t, 3, p.Page)
	assert.False(t, p.HasMore)
}
