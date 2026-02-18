package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"nodela.co/go-sdk/pkg/errors"
)

// Get executes an authenticated HTTP GET request to the given API path and
// deserialises the JSON response body into result.
//
// path must be relative to the base API URL (e.g. "/v1/invoices/inv_xxxx/verify").
// result must be a non-nil pointer to the struct that the response body should be
// decoded into.
//
// Returns a [*errors.SDKError] if the request cannot be built or the response cannot
// be read or deserialised. Returns a [*errors.APIError] if the server responds with
// a 4xx or 5xx HTTP status code.
func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
	return c.doRequest(ctx, RequestOptions{
		Method: http.MethodGet,
		Path:   path,
	}, result)
}

// Post executes an authenticated HTTP POST request to the given API path,
// serialising body as JSON and deserialising the JSON response body into result.
//
// path must be relative to the base API URL (e.g. "/v1/invoices").
// body is marshalled to JSON and sent with a Content-Type: application/json header;
// pass nil to send a request with no body.
// result must be a non-nil pointer to the struct that the response body should be
// decoded into.
//
// Returns a [*errors.SDKError] if body cannot be marshalled, the request cannot be
// built, or the response cannot be read or deserialised. Returns a [*errors.APIError]
// if the server responds with a 4xx or 5xx HTTP status code.
func (c *Client) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.doRequest(ctx, RequestOptions{
		Method: http.MethodPost,
		Path:   path,
		Body:   body,
	}, result)
}

// RequestOptions configures a single HTTP request executed by [Client.doRequest].
// It is consumed internally by the [Client.Get] and [Client.Post] helpers and is
// not intended for direct use by callers of the SDK.
type RequestOptions struct {
	// Method is the HTTP method to use (e.g. http.MethodGet, http.MethodPost).
	Method string

	// Path is the API path relative to the base URL (e.g. "/v1/invoices").
	// It must begin with a leading slash.
	Path string

	// QueryParams holds URL query string parameters to append to the request URL.
	// When empty or nil, no query string is appended.
	QueryParams url.Values

	// Body is the request payload. When non-nil it is marshalled to JSON and sent
	// as the request body with a Content-Type: application/json header. When nil,
	// the request is sent without a body.
	Body interface{}

	// Headers holds additional HTTP headers to merge into the outgoing request.
	// These are applied after the default Authorization, Content-Type, and
	// User-Agent headers, so they may override those defaults if needed.
	Headers map[string]string
}

// doRequest is the central request dispatcher for the client. It constructs and
// sends an HTTP request according to opts, handles authentication headers, reads
// the response body, and either decodes a successful response into result or
// converts an error response into a typed error.
//
// The request lifecycle is:
//  1. Build the full URL from baseURL + opts.Path (+ optional query string).
//  2. Marshal opts.Body to JSON if non-nil.
//  3. Construct the *http.Request with the context and set standard headers.
//  4. Apply any per-request headers from opts.Headers.
//  5. Execute the request via the configured httpClient.
//  6. Read the response body in full.
//  7. On a 4xx/5xx status, delegate to handleErrorResponse.
//  8. On success, unmarshal the body into result if result is non-nil.
func (c *Client) doRequest(ctx context.Context, opts RequestOptions, result interface{}) error {
	fullURL := fmt.Sprintf("%s%s", baseURL, opts.Path)
	if len(opts.QueryParams) > 0 {
		fullURL = fmt.Sprintf("%s%s", fullURL, opts.QueryParams.Encode())
	}

	var bodyReader io.Reader
	if opts.Body != nil {
		jsonBody, err := json.Marshal(opts.Body)
		if err != nil {
			return errors.NewSDKError(errors.ErrCodeSerialization, "failed to marshal request body", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, opts.Method, fullURL, bodyReader)
	if err != nil {
		return errors.NewSDKError(errors.ErrCodeRequest, "failed to create request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("User-Agent", c.userAgent)

	for key, value := range opts.Headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.NewSDKError(errors.ErrCodeNetwork, "request failed", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.NewSDKError(errors.ErrCodeResponse, "failed to read response body", err)
	}

	if resp.StatusCode >= 400 {
		return c.handleErrorResponse(resp.StatusCode, bodyBytes)
	}

	if result != nil {
		if err := json.Unmarshal(bodyBytes, result); err != nil {
			return errors.NewSDKError(errors.ErrCodeDeserialization, "failed to unmarshal response", err)
		}
	}

	return nil
}

// handleErrorResponse converts a non-2xx API response into an [*errors.APIError].
//
// It attempts to decode the response body as a JSON object containing "error",
// "message", and "code" fields. If decoding fails (e.g. the body is plain text or
// empty), it falls back to using the raw body as the error code so that no
// information is silently discarded.
func (c *Client) handleErrorResponse(statusCode int, body []byte) error {
	var apiError struct {
		Error   string `json:"error"`
		Message string `json:"message"`
		Code    string `json:"code"`
	}

	if err := json.Unmarshal(body, &apiError); err != nil {
		return errors.NewAPIError(statusCode, "unknown error", string(body))
	}

	return errors.NewAPIError(statusCode, apiError.Message, apiError.Code)
}
