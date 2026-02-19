package unit

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	sdkerrors "nodela.co/go-sdk/pkg/errors"
	"nodela.co/go-sdk/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeCreateInvoiceResponse(data *models.CreateInvoiceData) models.CreateInvoiceResponse {
	return models.CreateInvoiceResponse{Success: true, Data: data}
}

func makeVerifyInvoiceResponse(data *models.VerifyInvoiceData) models.VerifyInvoiceResponse {
	return models.VerifyInvoiceResponse{Success: true, Data: data}
}

func sampleCreateData() *models.CreateInvoiceData {
	return &models.CreateInvoiceData{
		ID:               "inv_abc",
		InvoiceID:        "INV-001",
		OriginalAmount:   "100.00",
		OriginalCurrency: "USD",
		Amount:           "0.0025",
		Currency:         "BTC",
		CheckoutURL:      "https://checkout.nodela.co/inv_abc",
		CreatedAt:        "2024-01-01T00:00:00Z",
	}
}

func sampleVerifyData(paid bool) *models.VerifyInvoiceData {
	status := "pending"
	if paid {
		status = "completed"
	}
	return &models.VerifyInvoiceData{
		ID:               "inv_abc",
		InvoiceID:        "INV-001",
		OriginalAmount:   "100.00",
		OriginalCurrency: "USD",
		Amount:           100.00,
		Currency:         "USD",
		Status:           status,
		Paid:             paid,
		CreatedAt:        "2024-01-01T00:00:00Z",
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// ---------------------------------------------------------------------------
// Create — success cases
// ---------------------------------------------------------------------------

func TestInvoices_Create_Success_MinimalParams(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/invoices", r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		writeJSON(w, 200, makeCreateInvoiceResponse(sampleCreateData()))
	})

	c := newTestClient(t, handler)
	resp, err := c.Invoices.Create(context.Background(), models.CreateInvoiceParams{
		Amount:   100.00,
		Currency: "USD",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
	require.NotNil(t, resp.Data)
	assert.Equal(t, "inv_abc", resp.Data.ID)
	assert.Equal(t, "https://checkout.nodela.co/inv_abc", resp.Data.CheckoutURL)
}

func TestInvoices_Create_Success_LowercaseCurrencyIsNormalized(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body models.CreateInvoiceParams
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		// The SDK must uppercase the currency before sending
		assert.Equal(t, "GBP", body.Currency)
		writeJSON(w, 200, makeCreateInvoiceResponse(sampleCreateData()))
	})

	c := newTestClient(t, handler)
	_, err := c.Invoices.Create(context.Background(), models.CreateInvoiceParams{
		Amount:   50.00,
		Currency: "gbp", // lowercase
	})
	require.NoError(t, err)
}

func TestInvoices_Create_Success_MixedCaseCurrencyIsNormalized(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body models.CreateInvoiceParams
		_ = json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "EUR", body.Currency)
		writeJSON(w, 200, makeCreateInvoiceResponse(sampleCreateData()))
	})

	c := newTestClient(t, handler)
	_, err := c.Invoices.Create(context.Background(), models.CreateInvoiceParams{
		Amount:   200.00,
		Currency: "Eur",
	})
	require.NoError(t, err)
}

func TestInvoices_Create_Success_WithAllOptionalFields(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body models.CreateInvoiceParams
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))

		assert.Equal(t, 75.50, body.Amount)
		assert.Equal(t, "NGN", body.Currency)
		require.NotNil(t, body.Customer)
		assert.Equal(t, "customer@example.com", body.Customer.Email)
		require.NotNil(t, body.Title)
		assert.Equal(t, "Test Purchase", *body.Title)

		writeJSON(w, 201, makeCreateInvoiceResponse(sampleCreateData()))
	})

	c := newTestClient(t, handler)
	resp, err := c.Invoices.Create(context.Background(), models.CreateInvoiceParams{
		Amount:      75.50,
		Currency:    "NGN",
		Customer:    &models.InvoiceCustomer{Email: "customer@example.com", Name: ptr("Customer")},
		Title:       ptr("Test Purchase"),
		Description: ptr("A test purchase description"),
		SuccessURL:  ptr("https://example.com/success"),
		CancelURL:   ptr("https://example.com/cancel"),
		WebhookURL:  ptr("https://example.com/webhook"),
		Reference:   ptr("ref-789"),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestInvoices_Create_Success_VariousCurrencies(t *testing.T) {
	currencies := []string{
		"USD", "EUR", "GBP", "NGN", "JPY", "AED", "AUD", "BRL", "KES",
	}

	for _, currency := range currencies {
		currency := currency
		t.Run(currency, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				writeJSON(w, 200, makeCreateInvoiceResponse(sampleCreateData()))
			})
			c := newTestClient(t, handler)
			_, err := c.Invoices.Create(context.Background(), models.CreateInvoiceParams{
				Amount:   10.00,
				Currency: currency,
			})
			assert.NoError(t, err, "currency %q should be accepted", currency)
		})
	}
}

// ---------------------------------------------------------------------------
// Create — validation errors (no HTTP call)
// ---------------------------------------------------------------------------

func TestInvoices_Create_InvalidCurrency_ReturnsValidationError(t *testing.T) {
	// Handler should never be called for a validation error
	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(500)
	})

	c := newTestClient(t, handler)
	resp, err := c.Invoices.Create(context.Background(), models.CreateInvoiceParams{
		Amount:   100,
		Currency: "XYZ", // not supported
	})

	assert.False(t, handlerCalled, "server should NOT be called for validation errors")
	require.Error(t, err)
	assert.Nil(t, resp)

	var sdkErr *sdkerrors.SDKError
	require.ErrorAs(t, err, &sdkErr)
	assert.Equal(t, sdkerrors.ErrCodeValidation, sdkErr.Code)
	assert.Contains(t, sdkErr.Message, "XYZ")
}

func TestInvoices_Create_EmptyCurrency_ReturnsValidationError(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	_, err := c.Invoices.Create(context.Background(), models.CreateInvoiceParams{
		Amount:   10,
		Currency: "",
	})
	require.Error(t, err)
	var sdkErr *sdkerrors.SDKError
	require.ErrorAs(t, err, &sdkErr)
	assert.Equal(t, sdkerrors.ErrCodeValidation, sdkErr.Code)
}

// ---------------------------------------------------------------------------
// Create — API / HTTP error cases
// ---------------------------------------------------------------------------

func TestInvoices_Create_APIError_401_Unauthorized(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 401, map[string]string{
			"error":   "Unauthorized",
			"message": "invalid API key",
			"code":    "AUTHENTICATION_ERROR",
		})
	})

	c := newTestClient(t, handler)
	resp, err := c.Invoices.Create(context.Background(), models.CreateInvoiceParams{
		Amount: 10, Currency: "USD",
	})

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, sdkerrors.IsAuthenticatedError(err))

	var apiErr *sdkerrors.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 401, apiErr.StatusCode)
}

func TestInvoices_Create_APIError_400_BadRequest(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 400, map[string]string{
			"error":   "Bad Request",
			"message": "amount must be greater than zero",
			"code":    "VALIDATION_ERROR",
		})
	})

	c := newTestClient(t, handler)
	resp, err := c.Invoices.Create(context.Background(), models.CreateInvoiceParams{
		Amount: 0, Currency: "USD",
	})

	require.Error(t, err)
	assert.Nil(t, resp)

	var apiErr *sdkerrors.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 400, apiErr.StatusCode)
	assert.Equal(t, "amount must be greater than zero", apiErr.Message)
}

func TestInvoices_Create_APIError_429_RateLimit(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 429, map[string]string{
			"error":   "Too Many Requests",
			"message": "rate limit exceeded",
			"code":    "RATE_LIMIT_ERROR",
		})
	})

	c := newTestClient(t, handler)
	_, err := c.Invoices.Create(context.Background(), models.CreateInvoiceParams{
		Amount: 10, Currency: "USD",
	})

	require.Error(t, err)
	assert.True(t, sdkerrors.IsRateLimitError(err))
}

func TestInvoices_Create_APIError_500_InternalServer(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 500, map[string]string{
			"error":   "Internal Server Error",
			"message": "unexpected error",
			"code":    "SERVER_ERROR",
		})
	})

	c := newTestClient(t, handler)
	_, err := c.Invoices.Create(context.Background(), models.CreateInvoiceParams{
		Amount: 10, Currency: "USD",
	})

	require.Error(t, err)
	var apiErr *sdkerrors.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 500, apiErr.StatusCode)
}

func TestInvoices_Create_MalformedJSONResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{not valid json`))
	})

	c := newTestClient(t, handler)
	_, err := c.Invoices.Create(context.Background(), models.CreateInvoiceParams{
		Amount: 10, Currency: "USD",
	})
	require.Error(t, err)
}

func TestInvoices_Create_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, makeCreateInvoiceResponse(sampleCreateData()))
	}))

	_, err := c.Invoices.Create(ctx, models.CreateInvoiceParams{
		Amount: 10, Currency: "USD",
	})
	require.Error(t, err)
}

func TestInvoices_Create_ErrorResponseWithUnknownBody(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
		_, _ = w.Write([]byte("Service Unavailable"))
	})

	c := newTestClient(t, handler)
	_, err := c.Invoices.Create(context.Background(), models.CreateInvoiceParams{
		Amount: 10, Currency: "USD",
	})
	require.Error(t, err)
	var apiErr *sdkerrors.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 503, apiErr.StatusCode)
}

// ---------------------------------------------------------------------------
// Verify — success cases
// ---------------------------------------------------------------------------

func TestInvoices_Verify_Success_PaidInvoice(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v1/invoices/INV-001/verify", r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		data := sampleVerifyData(true)
		data.Payment = &models.InvoicePayment{
			ID:              "pay_xyz",
			Network:         "polygon",
			Token:           "USDC",
			Address:         "0xABC",
			Amount:          100.00,
			Status:          "confirmed",
			TxHash:          []string{"0xDEF"},
			TransactionType: "onchain",
			PayerEmail:      "payer@example.com",
			CreatedAt:       "2024-01-01T01:00:00Z",
		}
		writeJSON(w, 200, makeVerifyInvoiceResponse(data))
	})

	c := newTestClient(t, handler)
	resp, err := c.Invoices.Verify(context.Background(), "INV-001")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
	require.NotNil(t, resp.Data)
	assert.True(t, resp.Data.Paid)
	assert.Equal(t, "completed", resp.Data.Status)
	require.NotNil(t, resp.Data.Payment)
	assert.Equal(t, "pay_xyz", resp.Data.Payment.ID)
	assert.Equal(t, []string{"0xDEF"}, resp.Data.Payment.TxHash)
}

func TestInvoices_Verify_Success_UnpaidInvoice(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, makeVerifyInvoiceResponse(sampleVerifyData(false)))
	})

	c := newTestClient(t, handler)
	resp, err := c.Invoices.Verify(context.Background(), "INV-002")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Data.Paid)
	assert.Equal(t, "pending", resp.Data.Status)
	assert.Nil(t, resp.Data.Payment)
}

func TestInvoices_Verify_PathContainsInvoiceID(t *testing.T) {
	invoiceID := "my-special-invoice-id"
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, invoiceID)
		assert.Contains(t, r.URL.Path, "verify")
		writeJSON(w, 200, makeVerifyInvoiceResponse(sampleVerifyData(false)))
	})

	c := newTestClient(t, handler)
	_, err := c.Invoices.Verify(context.Background(), invoiceID)
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Verify — error cases
// ---------------------------------------------------------------------------

func TestInvoices_Verify_NotFound_404(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 404, map[string]string{
			"error":   "Not Found",
			"message": "invoice not found",
			"code":    "NOT_FOUND",
		})
	})

	c := newTestClient(t, handler)
	resp, err := c.Invoices.Verify(context.Background(), "non-existent")

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, sdkerrors.IsNotFoundError(err))

	var apiErr *sdkerrors.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestInvoices_Verify_Unauthorized_401(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 401, map[string]string{
			"message": "unauthorized",
			"code":    "AUTHENTICATION_ERROR",
		})
	})

	c := newTestClient(t, handler)
	_, err := c.Invoices.Verify(context.Background(), "inv_abc")

	require.Error(t, err)
	assert.True(t, sdkerrors.IsAuthenticatedError(err))
}

func TestInvoices_Verify_RateLimit_429(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 429, map[string]string{
			"message": "too many requests",
			"code":    "RATE_LIMIT_ERROR",
		})
	})

	c := newTestClient(t, handler)
	_, err := c.Invoices.Verify(context.Background(), "inv_abc")

	require.Error(t, err)
	assert.True(t, sdkerrors.IsRateLimitError(err))
}

func TestInvoices_Verify_ServerError_500(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 500, map[string]string{
			"message": "internal server error",
			"code":    "SERVER_ERROR",
		})
	})

	c := newTestClient(t, handler)
	_, err := c.Invoices.Verify(context.Background(), "inv_abc")

	require.Error(t, err)
	var apiErr *sdkerrors.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 500, apiErr.StatusCode)
}

func TestInvoices_Verify_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, makeVerifyInvoiceResponse(sampleVerifyData(true)))
	}))

	_, err := c.Invoices.Verify(ctx, "inv_abc")
	require.Error(t, err)
}
