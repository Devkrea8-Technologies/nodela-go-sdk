package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"nodela.co/go-sdk/pkg/models"
	sdkerrors "nodela.co/go-sdk/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Full create → verify flow
// ---------------------------------------------------------------------------

// TestInvoices_CreateThenVerify simulates the typical payment lifecycle:
// create an invoice, then verify its status.
func TestInvoices_CreateThenVerify_Lifecycle(t *testing.T) {
	const invoiceID = "INV-lifecycle-001"

	// Track how many times each endpoint is called
	var createCalls, verifyCalls atomic.Int32

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/invoices":
			createCalls.Add(1)

			var body models.CreateInvoiceParams
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, 199.99, body.Amount)
			assert.Equal(t, "USD", body.Currency)

			writeJSON(w, 201, models.CreateInvoiceResponse{
				Success: true,
				Data: &models.CreateInvoiceData{
					ID:               invoiceID,
					InvoiceID:        invoiceID,
					OriginalAmount:   "199.99",
					OriginalCurrency: "USD",
					Amount:           "0.005",
					Currency:         "BTC",
					CheckoutURL:      "https://checkout.nodela.co/" + invoiceID,
					CreatedAt:        time.Now().Format(time.RFC3339),
				},
			})

		case r.Method == http.MethodGet && r.URL.Path == "/v1/invoices/"+invoiceID+"/verify":
			verifyCalls.Add(1)
			writeJSON(w, 200, models.VerifyInvoiceResponse{
				Success: true,
				Data: &models.VerifyInvoiceData{
					ID:               invoiceID,
					InvoiceID:        invoiceID,
					OriginalAmount:   "199.99",
					OriginalCurrency: "USD",
					Amount:           0.005,
					Currency:         "BTC",
					Status:           "completed",
					Paid:             true,
					CreatedAt:        time.Now().Format(time.RFC3339),
					Payment: &models.InvoicePayment{
						ID:              "pay-001",
						Network:         "bitcoin",
						Token:           "BTC",
						Address:         "bc1qXXXXXXXX",
						Amount:          0.005,
						Status:          "confirmed",
						TxHash:          []string{"a1b2c3d4e5f6"},
						TransactionType: "onchain",
						PayerEmail:      "buyer@example.com",
						CreatedAt:       time.Now().Format(time.RFC3339),
					},
				},
			})

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	c, _ := newTestClient(t, handler)

	// Step 1: Create the invoice
	createResp, err := c.Invoices.Create(context.Background(), models.CreateInvoiceParams{
		Amount:   199.99,
		Currency: "USD",
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	assert.True(t, createResp.Success)
	require.NotNil(t, createResp.Data)
	assert.Equal(t, invoiceID, createResp.Data.ID)
	assert.NotEmpty(t, createResp.Data.CheckoutURL)

	// Step 2: Verify the invoice using the returned ID
	verifyResp, err := c.Invoices.Verify(context.Background(), createResp.Data.ID)
	require.NoError(t, err)
	require.NotNil(t, verifyResp)
	assert.True(t, verifyResp.Data.Paid)
	assert.Equal(t, "completed", verifyResp.Data.Status)
	require.NotNil(t, verifyResp.Data.Payment)
	assert.Equal(t, []string{"a1b2c3d4e5f6"}, verifyResp.Data.Payment.TxHash)

	// Verify each endpoint was called exactly once
	assert.Equal(t, int32(1), createCalls.Load())
	assert.Equal(t, int32(1), verifyCalls.Load())
}

// ---------------------------------------------------------------------------
// Create with all optional fields
// ---------------------------------------------------------------------------

func TestInvoices_Create_WithFullCustomer(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body models.CreateInvoiceParams
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))

		require.NotNil(t, body.Customer)
		assert.Equal(t, "jane@example.com", body.Customer.Email)
		require.NotNil(t, body.Customer.Name)
		assert.Equal(t, "Jane Doe", *body.Customer.Name)
		require.NotNil(t, body.Reference)
		assert.Equal(t, "order-XYZ", *body.Reference)

		writeJSON(w, 201, models.CreateInvoiceResponse{
			Success: true,
			Data: &models.CreateInvoiceData{
				ID:               "inv-customer",
				InvoiceID:        "INV-CUST",
				OriginalAmount:   "50.00",
				OriginalCurrency: "EUR",
				Amount:           "54.75",
				Currency:         "USD",
				CheckoutURL:      "https://checkout.nodela.co/inv-customer",
				CreatedAt:        "2024-01-01T00:00:00Z",
				Customer:         &models.InvoiceCustomer{Email: "jane@example.com", Name: ptr("Jane Doe")},
			},
		})
	})

	c, _ := newTestClient(t, handler)
	resp, err := c.Invoices.Create(context.Background(), models.CreateInvoiceParams{
		Amount:   50.00,
		Currency: "EUR",
		Customer: &models.InvoiceCustomer{
			Email: "jane@example.com",
			Name:  ptr("Jane Doe"),
		},
		Reference:   ptr("order-XYZ"),
		SuccessURL:  ptr("https://shop.example.com/success"),
		CancelURL:   ptr("https://shop.example.com/cancel"),
		WebhookURL:  ptr("https://shop.example.com/webhook"),
		Title:       ptr("Order #XYZ"),
		Description: ptr("Premium subscription"),
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
	require.NotNil(t, resp.Data)
	require.NotNil(t, resp.Data.Customer)
	assert.Equal(t, "jane@example.com", resp.Data.Customer.Email)
}

// ---------------------------------------------------------------------------
// Create — authorization header is always sent
// ---------------------------------------------------------------------------

func TestInvoices_Create_AuthorizationHeaderPresent(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer integration-test-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		writeJSON(w, 201, models.CreateInvoiceResponse{
			Success: true,
			Data: &models.CreateInvoiceData{
				ID: "inv-auth", InvoiceID: "INV-AUTH",
				OriginalAmount: "10.00", OriginalCurrency: "USD",
				Amount: "10.00", Currency: "USD",
				CheckoutURL: "https://checkout.nodela.co/inv-auth",
				CreatedAt:   "2024-01-01T00:00:00Z",
			},
		})
	})

	c, _ := newTestClient(t, handler)
	_, err := c.Invoices.Create(context.Background(), models.CreateInvoiceParams{
		Amount: 10, Currency: "USD",
	})
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Verify — poll until paid simulation
// ---------------------------------------------------------------------------

func TestInvoices_Verify_PollUntilPaid(t *testing.T) {
	var callCount atomic.Int32

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		n := callCount.Load()

		paid := n >= 3
		status := "pending"
		if paid {
			status = "completed"
		}

		var payment *models.InvoicePayment
		if paid {
			payment = &models.InvoicePayment{
				ID: "pay-poll", Network: "polygon", Token: "USDC",
				Address: "0xPOLL", Amount: 25.0, Status: "confirmed",
				TxHash: []string{"0xABCDEF"}, TransactionType: "onchain",
				PayerEmail: "poller@example.com", CreatedAt: "2024-01-01T01:00:00Z",
			}
		}

		writeJSON(w, 200, models.VerifyInvoiceResponse{
			Success: true,
			Data: &models.VerifyInvoiceData{
				ID: "inv-poll", InvoiceID: "INV-POLL",
				OriginalAmount: "25.00", OriginalCurrency: "USD",
				Amount: 25.0, Currency: "USD",
				Status: status, Paid: paid,
				CreatedAt: "2024-01-01T00:00:00Z",
				Payment:   payment,
			},
		})
	})

	c, _ := newTestClient(t, handler)
	ctx := context.Background()

	// Poll until paid (max 5 attempts)
	var lastResp *models.VerifyInvoiceResponse
	for i := 0; i < 5; i++ {
		resp, err := c.Invoices.Verify(ctx, "inv-poll")
		require.NoError(t, err)
		lastResp = resp
		if resp.Data.Paid {
			break
		}
	}

	require.NotNil(t, lastResp)
	assert.True(t, lastResp.Data.Paid)
	assert.Equal(t, "completed", lastResp.Data.Status)
	assert.Equal(t, int32(3), callCount.Load(), "should take exactly 3 polls to get a paid response")
}

// ---------------------------------------------------------------------------
// Error propagation
// ---------------------------------------------------------------------------

func TestInvoices_Create_InvalidCurrency_NeverHitsServer(t *testing.T) {
	serverCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true
		w.WriteHeader(500)
	})

	c, _ := newTestClient(t, handler)
	_, err := c.Invoices.Create(context.Background(), models.CreateInvoiceParams{
		Amount: 10, Currency: "INVALID",
	})

	require.Error(t, err)
	assert.False(t, serverCalled)

	var sdkErr *sdkerrors.SDKError
	require.ErrorAs(t, err, &sdkErr)
	assert.Equal(t, sdkerrors.ErrCodeValidation, sdkErr.Code)
}

func TestInvoices_Create_ServerReturns422_ReturnsAPIError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 422, map[string]string{
			"error":   "Unprocessable Entity",
			"message": "amount must be a positive number",
			"code":    "VALIDATION_ERROR",
		})
	})

	c, _ := newTestClient(t, handler)
	_, err := c.Invoices.Create(context.Background(), models.CreateInvoiceParams{
		Amount: 100, Currency: "USD",
	})

	require.Error(t, err)
	var apiErr *sdkerrors.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 422, apiErr.StatusCode)
	assert.Equal(t, "amount must be a positive number", apiErr.Message)
}

func TestInvoices_Verify_NotFound_IsDetectable(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 404, map[string]string{
			"message": "invoice not found",
			"code":    "NOT_FOUND",
		})
	})

	c, _ := newTestClient(t, handler)
	_, err := c.Invoices.Verify(context.Background(), "ghost-invoice")

	require.Error(t, err)
	assert.True(t, sdkerrors.IsNotFoundError(err))
}

// ---------------------------------------------------------------------------
// Context deadline respected
// ---------------------------------------------------------------------------

func TestInvoices_Create_ContextDeadline(t *testing.T) {
	// Handler that responds very slowly
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			return
		case <-time.After(5 * time.Second):
			writeJSON(w, 200, models.CreateInvoiceResponse{Success: true})
		}
	})

	c, _ := newTestClient(t, handler)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := c.Invoices.Create(ctx, models.CreateInvoiceParams{
		Amount: 10, Currency: "USD",
	})
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Currency normalisation is idempotent
// ---------------------------------------------------------------------------

func TestInvoices_Create_CurrencyNormalisationVariants(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"usd", "USD"},
		{"Usd", "USD"},
		{"USD", "USD"},
		{"ngn", "NGN"},
		{"Ngn", "NGN"},
		{"eur", "EUR"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body models.CreateInvoiceParams
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				assert.Equal(t, tc.expected, body.Currency)
				writeJSON(w, 200, models.CreateInvoiceResponse{
					Success: true,
					Data: &models.CreateInvoiceData{
						ID: "inv-norm", InvoiceID: "INV-NORM",
						OriginalAmount: "10.00", OriginalCurrency: tc.expected,
						Amount: "10.00", Currency: tc.expected,
						CheckoutURL: "https://checkout.nodela.co/inv-norm",
						CreatedAt:   "2024-01-01T00:00:00Z",
					},
				})
			})
			c, _ := newTestClient(t, handler)
			resp, err := c.Invoices.Create(context.Background(), models.CreateInvoiceParams{
				Amount: 10, Currency: tc.input,
			})
			require.NoError(t, err)
			assert.NotNil(t, resp)
		})
	}
}
