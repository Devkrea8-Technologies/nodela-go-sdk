// Package errors defines the structured error types used throughout the Nodela Go SDK.
//
// Two distinct error types are provided:
//
//   - [SDKError]: a client-side error that originates within the SDK itself, such as
//     request body serialization failures, input validation errors, or network-level
//     failures. These errors are always wrapped in SDKError so callers have a single
//     type to inspect for SDK-originated problems.
//
//   - [APIError]: an error response received from the Nodela API server, carrying the
//     HTTP status code and the machine-readable error code from the response body.
//
// Error codes are defined as package-level string constants (e.g. [ErrCodeValidation],
// [ErrCodeNotFound]) so callers can programmatically distinguish error categories.
//
// Convenience predicates — [IsNotFoundError], [IsAuthenticatedError], and
// [IsRateLimitError] — allow callers to quickly identify the most common failure
// modes without manually inspecting HTTP status codes or performing type assertions.
package errors

import "fmt"

// Error code constants categorise the origin of an [SDKError]. They are returned
// in the Code field and can be compared directly against the values below.
const (
	// ErrCodeUnknown is used when the error cannot be placed into a more specific
	// category. It is a safe fallback and should be treated as an unexpected condition.
	ErrCodeUnknown = "UNKNOWN"

	// ErrCodeNetwork indicates a failure at the transport layer — for example a DNS
	// resolution failure, connection refused, or a request that timed out before the
	// server responded.
	ErrCodeNetwork = "NETWORK_ERROR"

	// ErrCodeSerialization indicates that marshalling the request body to JSON failed.
	// This typically means the value passed as the request params contains a type that
	// the standard encoding/json package cannot serialise.
	ErrCodeSerialization = "SERIALIZATION_ERROR"

	// ErrCodeDeserialization indicates that unmarshalling the API response body failed.
	// This can happen if the server returns an unexpected payload shape or if the
	// response body is truncated.
	ErrCodeDeserialization = "DESERIALIZATION_ERROR"

	// ErrCodeRequest indicates that constructing the HTTP request itself failed — for
	// example due to a malformed URL or a cancelled context supplied before the request
	// was created.
	ErrCodeRequest = "REQUEST_ERROR"

	// ErrCodeResponse indicates a failure while reading or processing the HTTP response
	// after it was received — for example an error reading the response body stream.
	ErrCodeResponse = "RESPONSE_ERROR"

	// ErrCodeValidation indicates that input provided by the caller failed SDK-side
	// validation before a network request was ever made. For example, passing an
	// unsupported currency code to [Invoices.Create] returns this code.
	ErrCodeValidation = "VALIDATION_ERROR"

	// ErrCodeAuthentication indicates the request was rejected by the API because the
	// provided credentials were missing or invalid (HTTP 401).
	ErrCodeAuthentication = "AUTHENTICATION_ERROR"

	// ErrCodeAuthorization indicates the authenticated caller does not have permission
	// to perform the requested operation (HTTP 403).
	ErrCodeAuthorization = "AUTHORIZATION_ERROR"

	// ErrCodeNotFound indicates the requested resource does not exist on the server
	// (HTTP 404).
	ErrCodeNotFound = "NOT_FOUND"

	// ErrCodeRateLimit indicates the caller has exceeded the Nodela API's rate limit
	// (HTTP 429). Callers should implement exponential backoff before retrying.
	ErrCodeRateLimit = "RATE_LIMIT_ERROR"
)

// SDKError represents a client-side error that originates within the SDK, either
// before a request is sent (e.g. validation) or after a response is received
// (e.g. deserialisation failure). It is distinct from [APIError], which represents
// an error response explicitly returned by the Nodela API server.
//
// SDKError wraps an optional underlying error (accessible via [SDKError.Unwrap]) so
// it is fully compatible with [errors.Is] and [errors.As] from the standard library.
//
// Example:
//
//	_, err := client.Invoices.Create(ctx, params)
//	var sdkErr *errors.SDKError
//	if errors.As(err, &sdkErr) {
//	    log.Printf("SDK error [%s]: %s", sdkErr.Code, sdkErr.Message)
//	}
type SDKError struct {
	// Code is a machine-readable error category. Compare against the ErrCode*
	// constants in this package to handle specific error kinds.
	Code string

	// Message is a human-readable description of what went wrong. Suitable for
	// logging but not intended for display to end users.
	Message string

	// Err is the underlying error that caused this SDKError, if any. May be nil
	// for errors that originate purely within the SDK (e.g. validation failures).
	Err error
}

// Error implements the [error] interface.
//
// When an underlying error is present the message is formatted as:
//
//	[CODE] message: underlying error
//
// When there is no underlying error it is formatted as:
//
//	[CODE] message
func (e *SDKError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error wrapped by this SDKError, enabling
// compatibility with [errors.Is] and [errors.As] from the standard library.
// Returns nil if no underlying error was provided at construction time.
func (e *SDKError) Unwrap() error {
	return e.Err
}

// NewSDKError constructs a new [SDKError] with the given code, human-readable
// message, and optional underlying error. Pass nil for err if there is no
// underlying cause.
//
// This function is primarily used internally by the SDK. Callers should use
// [errors.As] to inspect errors returned by SDK methods rather than constructing
// SDKErrors directly.
func NewSDKError(code, message string, err error) *SDKError {
	return &SDKError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// APIError represents an HTTP error response returned by the Nodela API. It
// captures the HTTP status code alongside the machine-readable error code and
// human-readable message extracted from the JSON response body.
//
// Use [errors.As] to unwrap an APIError from any error returned by SDK methods:
//
//	_, err := client.Invoices.Verify(ctx, "inv_123")
//	var apiErr *errors.APIError
//	if errors.As(err, &apiErr) {
//	    log.Printf("API returned %d: %s (%s)", apiErr.StatusCode, apiErr.Message, apiErr.Code)
//	}
//
// For the most common status codes, prefer the predicate helpers ([IsNotFoundError],
// [IsAuthenticatedError], [IsRateLimitError]) over manual status code checks.
type APIError struct {
	// StatusCode is the HTTP status code returned by the API (e.g. 400, 401, 404, 429).
	StatusCode int

	// Message is the human-readable error description extracted from the API
	// response body. May be empty if the response body could not be decoded.
	Message string

	// Code is the machine-readable error code extracted from the API response body.
	// May be the raw response body string if JSON decoding failed.
	Code string
}

// Error implements the [error] interface, formatting the error as:
//
//	API error (status <StatusCode>): <Message> [<Code>]
func (e *APIError) Error() string {
	return fmt.Sprintf("API error (status %d): %s [%s]", e.StatusCode, e.Message, e.Code)
}

// NewAPIError constructs a new [APIError] with the given HTTP status code,
// human-readable message, and machine-readable code from the API response body.
//
// This is called internally by the SDK whenever the Nodela API returns a 4xx or
// 5xx HTTP response. Callers should not need to construct APIErrors directly.
func NewAPIError(statusCode int, message, code string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		Code:       code,
	}
}

// IsNotFoundError reports whether err is an [*APIError] with HTTP status 404,
// indicating that the requested resource does not exist on the server.
//
// Example:
//
//	_, err := client.Invoices.Verify(ctx, invoiceID)
//	if errors.IsNotFoundError(err) {
//	    // The invoice ID does not exist; handle accordingly.
//	}
func IsNotFoundError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 404
	}
	return false
}

// IsAuthenticatedError reports whether err is an [*APIError] with HTTP status 401,
// indicating that the request was rejected because the API key is missing, expired,
// or otherwise invalid. Verify that the API key passed to [NewClient] is correct.
func IsAuthenticatedError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 401
	}
	return false
}

// IsRateLimitError reports whether err is an [*APIError] with HTTP status 429,
// indicating that the caller has exceeded the Nodela API's rate limit.
// When this returns true, callers should implement exponential backoff before retrying.
func IsRateLimitError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 429
	}
	return false
}
