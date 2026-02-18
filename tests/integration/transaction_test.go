package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"nodela.co/go-sdk/pkg/models"
	sdkerrors "nodela.co/go-sdk/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func buildTransaction(id int) models.Transaction {
	sid := fmt.Sprintf("txn-%03d", id)
	return models.Transaction{
		ID:               sid,
		InvoiceID:        fmt.Sprintf("INV-%03d", id),
		Reference:        fmt.Sprintf("ref-%03d", id),
		OriginalAmount:   float64(id) * 10,
		OriginalCurrency: "USD",
		Amount:           float64(id) * 10,
		Currency:         "USD",
		ExchangeRate:     1.0,
		Title:            fmt.Sprintf("Invoice %d", id),
		Description:      fmt.Sprintf("Description for invoice %d", id),
		Status:           "completed",
		Paid:             true,
		Customer:         models.TransactionCustomer{Email: fmt.Sprintf("user%d@example.com", id), Name: fmt.Sprintf("User%d", id)},
		CreatedAt:        time.Now().Format(time.RFC3339),
		Payment: models.TransactionPayment{
			ID:              fmt.Sprintf("pay-%03d", id),
			Network:         "polygon",
			Token:           "USDC",
			Address:         fmt.Sprintf("0x%d", id),
			Amount:          float64(id) * 10,
			Status:          "confirmed",
			TxHash:          []string{fmt.Sprintf("0xhash%d", id)},
			TransactionType: "onchain",
			PayerEmail:      fmt.Sprintf("payer%d@example.com", id),
			CreatedAt:       time.Now().Format(time.RFC3339),
		},
	}
}

// ---------------------------------------------------------------------------
// List without params
// ---------------------------------------------------------------------------

func TestTransactions_List_NilParams_ReturnsAll(t *testing.T) {
	txns := []models.Transaction{buildTransaction(1), buildTransaction(2), buildTransaction(3)}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v1/transactions", r.URL.Path)
		assert.Equal(t, "Bearer integration-test-key", r.Header.Get("Authorization"))
		assert.Empty(t, r.URL.RawQuery, "nil params must not produce any query string")

		writeJSON(w, 200, models.ListTransactionsResponse{
			Success: true,
			Data: models.ListTransactionsData{
				Transactions: txns,
				Pagination:   models.Pagination{Page: 1, Limit: 10, Total: 3, TotalPages: 1, HasMore: false},
			},
		})
	})

	c, _ := newTestClient(t, handler)
	resp, err := c.Transactions.List(context.Background(), nil)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Len(t, resp.Data.Transactions, 3)
	assert.Equal(t, "txn-001", resp.Data.Transactions[0].ID)
}

// ---------------------------------------------------------------------------
// Pagination: full multi-page walk
// ---------------------------------------------------------------------------

func TestTransactions_List_PaginatedWalk(t *testing.T) {
	// Simulate 25 transactions served in pages of 10
	allTxns := make([]models.Transaction, 25)
	for i := range allTxns {
		allTxns[i] = buildTransaction(i + 1)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageStr := r.URL.Query().Get("page")
		limitStr := r.URL.Query().Get("limit")

		page, _ := strconv.Atoi(pageStr)
		limit, _ := strconv.Atoi(limitStr)
		if page == 0 {
			page = 1
		}
		if limit == 0 {
			limit = 10
		}

		start := (page - 1) * limit
		end := start + limit
		if end > len(allTxns) {
			end = len(allTxns)
		}

		totalPages := (len(allTxns) + limit - 1) / limit
		hasMore := page < totalPages

		writeJSON(w, 200, models.ListTransactionsResponse{
			Success: true,
			Data: models.ListTransactionsData{
				Transactions: allTxns[start:end],
				Pagination: models.Pagination{
					Page:       page,
					Limit:      limit,
					Total:      len(allTxns),
					TotalPages: totalPages,
					HasMore:    hasMore,
				},
			},
		})
	})

	c, _ := newTestClient(t, handler)
	ctx := context.Background()

	// Collect all transactions by walking all pages
	var collected []models.Transaction
	page := 1
	limit := 10
	for {
		resp, err := c.Transactions.List(ctx, &models.ListTransactionsParams{
			Page: ptr(page), Limit: ptr(limit),
		})
		require.NoError(t, err)
		collected = append(collected, resp.Data.Transactions...)

		if !resp.Data.Pagination.HasMore {
			break
		}
		page++
	}

	assert.Len(t, collected, 25, "should collect all 25 transactions across 3 pages")
	assert.Equal(t, "txn-001", collected[0].ID)
	assert.Equal(t, "txn-025", collected[24].ID)
}

// ---------------------------------------------------------------------------
// Specific page / limit combinations
// ---------------------------------------------------------------------------

func TestTransactions_List_FirstPageWithLimit5(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "5", r.URL.Query().Get("limit"))

		txns := make([]models.Transaction, 5)
		for i := range txns {
			txns[i] = buildTransaction(i + 1)
		}

		writeJSON(w, 200, models.ListTransactionsResponse{
			Success: true,
			Data: models.ListTransactionsData{
				Transactions: txns,
				Pagination:   models.Pagination{Page: 1, Limit: 5, Total: 20, TotalPages: 4, HasMore: true},
			},
		})
	})

	c, _ := newTestClient(t, handler)
	resp, err := c.Transactions.List(context.Background(), &models.ListTransactionsParams{
		Page: ptr(1), Limit: ptr(5),
	})

	require.NoError(t, err)
	assert.Len(t, resp.Data.Transactions, 5)
	assert.True(t, resp.Data.Pagination.HasMore)
	assert.Equal(t, 4, resp.Data.Pagination.TotalPages)
}

func TestTransactions_List_LastPage(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "3", r.URL.Query().Get("page"))

		txns := []models.Transaction{buildTransaction(21), buildTransaction(22)}
		writeJSON(w, 200, models.ListTransactionsResponse{
			Success: true,
			Data: models.ListTransactionsData{
				Transactions: txns,
				Pagination:   models.Pagination{Page: 3, Limit: 10, Total: 22, TotalPages: 3, HasMore: false},
			},
		})
	})

	c, _ := newTestClient(t, handler)
	resp, err := c.Transactions.List(context.Background(), &models.ListTransactionsParams{
		Page: ptr(3), Limit: ptr(10),
	})

	require.NoError(t, err)
	assert.Len(t, resp.Data.Transactions, 2)
	assert.False(t, resp.Data.Pagination.HasMore)
}

// ---------------------------------------------------------------------------
// Empty result set
// ---------------------------------------------------------------------------

func TestTransactions_List_EmptyAccount(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, models.ListTransactionsResponse{
			Success: true,
			Data: models.ListTransactionsData{
				Transactions: []models.Transaction{},
				Pagination:   models.Pagination{Page: 1, Limit: 10, Total: 0, TotalPages: 0, HasMore: false},
			},
		})
	})

	c, _ := newTestClient(t, handler)
	resp, err := c.Transactions.List(context.Background(), nil)

	require.NoError(t, err)
	assert.Empty(t, resp.Data.Transactions)
	assert.Equal(t, 0, resp.Data.Pagination.Total)
}

// ---------------------------------------------------------------------------
// Data integrity
// ---------------------------------------------------------------------------

func TestTransactions_List_TransactionFieldsAreComplete(t *testing.T) {
	txn := buildTransaction(42)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, models.ListTransactionsResponse{
			Success: true,
			Data: models.ListTransactionsData{
				Transactions: []models.Transaction{txn},
				Pagination:   models.Pagination{Page: 1, Limit: 10, Total: 1, TotalPages: 1},
			},
		})
	})

	c, _ := newTestClient(t, handler)
	resp, err := c.Transactions.List(context.Background(), nil)

	require.NoError(t, err)
	require.Len(t, resp.Data.Transactions, 1)

	got := resp.Data.Transactions[0]
	assert.Equal(t, "txn-042", got.ID)
	assert.Equal(t, "INV-042", got.InvoiceID)
	assert.Equal(t, 420.0, got.OriginalAmount)
	assert.Equal(t, "USD", got.OriginalCurrency)
	assert.Equal(t, 420.0, got.Amount)
	assert.Equal(t, 1.0, got.ExchangeRate)
	assert.True(t, got.Paid)
	assert.Equal(t, "user42@example.com", got.Customer.Email)
	assert.Equal(t, "User42", got.Customer.Name)
	assert.Equal(t, "polygon", got.Payment.Network)
	assert.Equal(t, "USDC", got.Payment.Token)
	assert.Equal(t, []string{"0xhash42"}, got.Payment.TxHash)
}

// ---------------------------------------------------------------------------
// Request format validation
// ---------------------------------------------------------------------------

func TestTransactions_List_OnlyPageParam(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		assert.Equal(t, "7", query.Get("page"))
		assert.Empty(t, query.Get("limit"), "limit must not appear when not set")
		writeJSON(w, 200, models.ListTransactionsResponse{
			Success: true,
			Data:    models.ListTransactionsData{Transactions: nil, Pagination: models.Pagination{Page: 7}},
		})
	})

	c, _ := newTestClient(t, handler)
	_, err := c.Transactions.List(context.Background(), &models.ListTransactionsParams{Page: ptr(7)})
	require.NoError(t, err)
}

func TestTransactions_List_OnlyLimitParam(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		assert.Empty(t, query.Get("page"), "page must not appear when not set")
		assert.Equal(t, "50", query.Get("limit"))
		writeJSON(w, 200, models.ListTransactionsResponse{
			Success: true,
			Data:    models.ListTransactionsData{Transactions: nil, Pagination: models.Pagination{Limit: 50}},
		})
	})

	c, _ := newTestClient(t, handler)
	_, err := c.Transactions.List(context.Background(), &models.ListTransactionsParams{Limit: ptr(50)})
	require.NoError(t, err)
}

func TestTransactions_List_JSONRequestBody_NotSent(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// GET requests must not have a body
		var body interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		// Either EOF (no body) or nil body — both acceptable
		assert.True(t, err != nil || body == nil, "GET /v1/transactions must not have a body")
		writeJSON(w, 200, models.ListTransactionsResponse{
			Success: true,
			Data:    models.ListTransactionsData{Transactions: []models.Transaction{}},
		})
	})

	c, _ := newTestClient(t, handler)
	_, err := c.Transactions.List(context.Background(), nil)
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Error handling
// ---------------------------------------------------------------------------

func TestTransactions_List_401_Unauthorized(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 401, map[string]string{
			"message": "invalid or missing API key",
			"code":    "AUTHENTICATION_ERROR",
		})
	})

	c, _ := newTestClient(t, handler)
	resp, err := c.Transactions.List(context.Background(), nil)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, sdkerrors.IsAuthenticatedError(err))

	var apiErr *sdkerrors.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 401, apiErr.StatusCode)
}

func TestTransactions_List_429_RateLimit(t *testing.T) {
	var attempts atomic.Int32
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		writeJSON(w, 429, map[string]string{
			"message": "rate limit exceeded, please retry after 60 seconds",
			"code":    "RATE_LIMIT_ERROR",
		})
	})

	c, _ := newTestClient(t, handler)
	_, err := c.Transactions.List(context.Background(), nil)

	require.Error(t, err)
	assert.True(t, sdkerrors.IsRateLimitError(err))
	// SDK should not auto-retry — exactly one attempt
	assert.Equal(t, int32(1), attempts.Load())
}

func TestTransactions_List_500_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 500, map[string]string{
			"message": "internal server error",
			"code":    "SERVER_ERROR",
		})
	})

	c, _ := newTestClient(t, handler)
	_, err := c.Transactions.List(context.Background(), nil)

	require.Error(t, err)
	var apiErr *sdkerrors.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 500, apiErr.StatusCode)
}

func TestTransactions_List_503_ServiceUnavailable(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
		_, _ = w.Write([]byte("Service Unavailable"))
	})

	c, _ := newTestClient(t, handler)
	_, err := c.Transactions.List(context.Background(), nil)

	require.Error(t, err)
	var apiErr *sdkerrors.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 503, apiErr.StatusCode)
}

func TestTransactions_List_ContextTimeout(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			return
		case <-time.After(5 * time.Second):
			writeJSON(w, 200, models.ListTransactionsResponse{Success: true})
		}
	})

	c, _ := newTestClient(t, handler)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := c.Transactions.List(ctx, nil)
	require.Error(t, err)
}
