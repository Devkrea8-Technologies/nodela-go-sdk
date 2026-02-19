package unit

import (
	"context"
	"encoding/json"
	"fmt"
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

func makePaginatedResponse(txns []models.Transaction, page, limit, total, totalPages int, hasMore bool) models.ListTransactionsResponse {
	return models.ListTransactionsResponse{
		Success: true,
		Data: models.ListTransactionsData{
			Transactions: txns,
			Pagination: models.Pagination{
				Page:       page,
				Limit:      limit,
				Total:      total,
				TotalPages: totalPages,
				HasMore:    hasMore,
			},
		},
	}
}

func sampleTransaction(id string) models.Transaction {
	return models.Transaction{
		ID:               id,
		InvoiceID:        "INV-" + id,
		Reference:        "ref-" + id,
		OriginalAmount:   50.00,
		OriginalCurrency: "USD",
		Amount:           50.00,
		Currency:         "USD",
		ExchangeRate:     1.0,
		Title:            "Test Invoice",
		Description:      "Test description",
		Status:           "completed",
		Paid:             true,
		Customer:         models.TransactionCustomer{Email: "user@example.com", Name: "User"},
		CreatedAt:        "2024-01-01T00:00:00Z",
		Payment: models.TransactionPayment{
			ID:              "pay-" + id,
			Network:         "polygon",
			Token:           "USDC",
			Address:         "0xABC",
			Amount:          50.00,
			Status:          "confirmed",
			TxHash:          []string{"0xDEF"},
			TransactionType: "onchain",
			PayerEmail:      "payer@example.com",
			CreatedAt:       "2024-01-01T01:00:00Z",
		},
	}
}

// ---------------------------------------------------------------------------
// List — success cases
// ---------------------------------------------------------------------------

func TestTransactions_List_NilParams_AllTransactions(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v1/transactions", r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
		// No query params when params is nil
		assert.Empty(t, r.URL.RawQuery)

		resp := makePaginatedResponse(
			[]models.Transaction{sampleTransaction("t1"), sampleTransaction("t2")},
			1, 10, 2, 1, false,
		)
		writeJSON(w, 200, resp)
	})

	c := newTestClient(t, handler)
	resp, err := c.Transactions.List(context.Background(), nil)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Len(t, resp.Data.Transactions, 2)
}

func TestTransactions_List_EmptyParams_NoQueryString(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Both Page and Limit are nil — no query params expected
		assert.Empty(t, r.URL.RawQuery)
		writeJSON(w, 200, makePaginatedResponse(nil, 1, 10, 0, 0, false))
	})

	c := newTestClient(t, handler)
	params := &models.ListTransactionsParams{} // both Page and Limit are nil
	_, err := c.Transactions.List(context.Background(), params)
	require.NoError(t, err)
}

func TestTransactions_List_WithPage(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		assert.Empty(t, r.URL.Query().Get("limit"))
		writeJSON(w, 200, makePaginatedResponse(
			[]models.Transaction{sampleTransaction("t3")},
			2, 10, 3, 1, false,
		))
	})

	c := newTestClient(t, handler)
	resp, err := c.Transactions.List(context.Background(), &models.ListTransactionsParams{
		Page: ptr(2),
	})

	require.NoError(t, err)
	assert.Equal(t, 2, resp.Data.Pagination.Page)
}

func TestTransactions_List_WithLimit(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.URL.Query().Get("page"))
		assert.Equal(t, "25", r.URL.Query().Get("limit"))
		writeJSON(w, 200, makePaginatedResponse(
			[]models.Transaction{sampleTransaction("t4")},
			1, 25, 1, 1, false,
		))
	})

	c := newTestClient(t, handler)
	resp, err := c.Transactions.List(context.Background(), &models.ListTransactionsParams{
		Limit: ptr(25),
	})

	require.NoError(t, err)
	assert.Equal(t, 25, resp.Data.Pagination.Limit)
}

func TestTransactions_List_WithPageAndLimit(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "3", r.URL.Query().Get("page"))
		assert.Equal(t, "5", r.URL.Query().Get("limit"))
		writeJSON(w, 200, makePaginatedResponse(
			[]models.Transaction{sampleTransaction("t5")},
			3, 5, 11, 3, false,
		))
	})

	c := newTestClient(t, handler)
	resp, err := c.Transactions.List(context.Background(), &models.ListTransactionsParams{
		Page:  ptr(3),
		Limit: ptr(5),
	})

	require.NoError(t, err)
	assert.Equal(t, 3, resp.Data.Pagination.Page)
	assert.Equal(t, 5, resp.Data.Pagination.Limit)
}

func TestTransactions_List_EmptyResult(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, makePaginatedResponse([]models.Transaction{}, 1, 10, 0, 0, false))
	})

	c := newTestClient(t, handler)
	resp, err := c.Transactions.List(context.Background(), nil)

	require.NoError(t, err)
	assert.Empty(t, resp.Data.Transactions)
	assert.Equal(t, 0, resp.Data.Pagination.Total)
	assert.False(t, resp.Data.Pagination.HasMore)
}

func TestTransactions_List_Pagination_HasMore(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var txns []models.Transaction
		for i := 0; i < 10; i++ {
			txns = append(txns, sampleTransaction(fmt.Sprintf("t%d", i)))
		}
		writeJSON(w, 200, makePaginatedResponse(txns, 1, 10, 50, 5, true))
	})

	c := newTestClient(t, handler)
	resp, err := c.Transactions.List(context.Background(), &models.ListTransactionsParams{
		Page: ptr(1), Limit: ptr(10),
	})

	require.NoError(t, err)
	assert.Len(t, resp.Data.Transactions, 10)
	assert.True(t, resp.Data.Pagination.HasMore)
	assert.Equal(t, 5, resp.Data.Pagination.TotalPages)
}

func TestTransactions_List_TransactionData_IsPopulated(t *testing.T) {
	txn := sampleTransaction("txn-full")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, makePaginatedResponse([]models.Transaction{txn}, 1, 10, 1, 1, false))
	})

	c := newTestClient(t, handler)
	resp, err := c.Transactions.List(context.Background(), nil)

	require.NoError(t, err)
	require.Len(t, resp.Data.Transactions, 1)
	got := resp.Data.Transactions[0]
	assert.Equal(t, "txn-full", got.ID)
	assert.Equal(t, "INV-txn-full", got.InvoiceID)
	assert.Equal(t, 50.00, got.OriginalAmount)
	assert.True(t, got.Paid)
	assert.Equal(t, "user@example.com", got.Customer.Email)
	assert.Equal(t, []string{"0xDEF"}, got.Payment.TxHash)
	assert.Equal(t, "polygon", got.Payment.Network)
}

// ---------------------------------------------------------------------------
// List — error cases
// ---------------------------------------------------------------------------

func TestTransactions_List_APIError_401(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 401, map[string]string{
			"message": "invalid API key",
			"code":    "AUTHENTICATION_ERROR",
		})
	})

	c := newTestClient(t, handler)
	resp, err := c.Transactions.List(context.Background(), nil)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, sdkerrors.IsAuthenticatedError(err))
}

func TestTransactions_List_APIError_403(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 403, map[string]string{
			"message": "access forbidden",
			"code":    "AUTHORIZATION_ERROR",
		})
	})

	c := newTestClient(t, handler)
	_, err := c.Transactions.List(context.Background(), nil)

	require.Error(t, err)
	var apiErr *sdkerrors.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 403, apiErr.StatusCode)
}

func TestTransactions_List_APIError_429_RateLimit(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 429, map[string]string{
			"message": "rate limit exceeded",
			"code":    "RATE_LIMIT_ERROR",
		})
	})

	c := newTestClient(t, handler)
	_, err := c.Transactions.List(context.Background(), nil)

	require.Error(t, err)
	assert.True(t, sdkerrors.IsRateLimitError(err))
}

func TestTransactions_List_APIError_500(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 500, map[string]string{
			"message": "internal error",
			"code":    "SERVER_ERROR",
		})
	})

	c := newTestClient(t, handler)
	_, err := c.Transactions.List(context.Background(), nil)

	require.Error(t, err)
	var apiErr *sdkerrors.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 500, apiErr.StatusCode)
}

func TestTransactions_List_MalformedJSON(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{broken`))
	})

	c := newTestClient(t, handler)
	_, err := c.Transactions.List(context.Background(), nil)
	require.Error(t, err)
}

func TestTransactions_List_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, makePaginatedResponse(nil, 1, 10, 0, 0, false))
	}))

	_, err := c.Transactions.List(ctx, nil)
	require.Error(t, err)
}

func TestTransactions_List_RequestBodyIsNotSent(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// GET requests should have no body
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		assert.Nil(t, body, "GET /v1/transactions must not send a request body")
		writeJSON(w, 200, makePaginatedResponse(nil, 1, 10, 0, 0, false))
	})

	c := newTestClient(t, handler)
	_, err := c.Transactions.List(context.Background(), nil)
	require.NoError(t, err)
}
