package client

import (
	"context"
	"fmt"
	"net/url"

	"nodela.co/go-sdk/pkg/models"
)

// transactionBasePath is the root API path for all transaction endpoints.
const transactionBasePath = "/v1/transactions"

// Transactions provides methods for interacting with the Nodela transaction API.
//
// Access this service via the [Client.Transactions] field on a configured [Client].
// Do not construct Transactions directly.
//
//	resp, err := c.Transactions.List(ctx, nil)
type Transactions struct {
	client *Client
}

// List retrieves a paginated list of settled transactions from the Nodela API.
//
// # Pagination
//
// Pass a *[models.ListTransactionsParams] to control which page and how many items
// per page are returned. Pass nil to use the server's defaults (typically page 1
// with the server's default page size).
//
// Only non-nil fields in params are sent as query parameters — this allows you to
// set just Page without overriding the server's default Limit, or vice versa.
//
// The returned [models.Pagination] struct (at resp.Data.Pagination) contains:
//   - Page — the current page number
//   - Limit — the effective page size used by the server
//   - Total — total number of transactions across all pages
//   - TotalPages — total number of available pages
//   - HasMore — whether at least one more page exists after the current one
//
// Use HasMore to drive a pagination loop rather than comparing Page to TotalPages
// directly, as this is more robust to concurrent writes between pages.
//
// # Example — fetch the first page with server defaults
//
//	resp, err := c.Transactions.List(ctx, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, tx := range resp.Data.Transactions {
//	    fmt.Printf("%s  %s %.2f %s\n", tx.CreatedAt, tx.InvoiceID, tx.Amount, tx.Currency)
//	}
//
// # Example — iterate all pages
//
//	page  := 1
//	limit := 50
//	for {
//	    resp, err := c.Transactions.List(ctx, &models.ListTransactionsParams{
//	        Page:  &page,
//	        Limit: &limit,
//	    })
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    process(resp.Data.Transactions)
//	    if !resp.Data.Pagination.HasMore {
//	        break
//	    }
//	    page++
//	}
func (s *Transactions) List(ctx context.Context, params *models.ListTransactionsParams) (*models.ListTransactionsResponse, error) {
	path := transactionBasePath

	if params != nil {
		q := url.Values{}
		if params.Page != nil {
			q.Set("page", fmt.Sprintf("%d", *params.Page))
		}
		if params.Limit != nil {
			q.Set("limit", fmt.Sprintf("%d", *params.Limit))
		}
		if len(q) > 0 {
			path = fmt.Sprintf("%s?%s", path, q.Encode())
		}
	}

	var result models.ListTransactionsResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
