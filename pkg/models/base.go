// Package models defines the request parameter types and response data types
// used by the Nodela Go SDK. Every type in this package maps directly to a JSON
// structure accepted or returned by the Nodela REST API.
//
// Types are organised into three broad groups:
//
//   - Parameter types (e.g. [CreateInvoiceParams], [ListTransactionsParams]):
//     populated by the caller and passed to service methods.
//
//   - Response types (e.g. [CreateInvoiceResponse], [ListTransactionsResponse]):
//     returned by service methods after a successful API call.
//
//   - Shared / embedded types (e.g. [InvoiceCustomer], [Pagination]):
//     reused across multiple request or response types.
package models

import "time"

// Metadata holds common timestamp fields returned by the Nodela API on resources
// that track both creation and last-modification times.
//
// Not all API resources include Metadata; refer to the specific response type
// (e.g. [CreateInvoiceData]) for the timestamp format used by that endpoint.
type Metadata struct {
	// CreatedAt is the time the resource was first created on the server.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the time the resource was most recently modified on the server.
	UpdatedAt time.Time `json:"updated_at"`
}

// PaginationParams holds optional page-based pagination parameters that can be
// embedded in or passed alongside list request bodies. Zero values are omitted
// from JSON serialisation.
//
// For endpoints that accept pagination via query string parameters (e.g.
// [ListTransactionsParams]), prefer using that type directly.
type PaginationParams struct {
	// Page is the 1-based page number to retrieve. When zero, the server returns
	// the first page.
	Page int `json:"page,omitempty"`

	// PerPage is the maximum number of items to include in each page.
	// When zero, the server applies its own default page size.
	PerPage int `json:"per_page,omitempty"`
}

// PaginatedResponse is a generic response envelope used by paginated list
// endpoints. The Data field holds the actual items for the requested page;
// its concrete type depends on the specific endpoint called.
//
// Prefer the strongly-typed response wrappers (e.g. [ListTransactionsResponse])
// over this type where possible.
type PaginatedResponse struct {
	// Data holds the items for the current page. The concrete type varies by
	// endpoint and must be type-asserted by the caller.
	Data interface{} `json:"data"`

	// Page is the current 1-based page number.
	Page int `json:"page"`

	// PerPage is the number of items requested per page.
	PerPage int `json:"per_page"`

	// Total is the total number of items across all pages.
	Total int `json:"total"`

	// TotalPages is the total number of pages available at the current PerPage size.
	TotalPages int `json:"total_pages"`
}
