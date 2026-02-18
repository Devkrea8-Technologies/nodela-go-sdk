package unit

import (
	"errors"
	"testing"

	sdkerrors "nodela.co/go-sdk/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Error code constants
// ---------------------------------------------------------------------------

func TestErrorCodeConstants(t *testing.T) {
	assert.Equal(t, "UNKNOWN", sdkerrors.ErrCodeUnknown)
	assert.Equal(t, "NETWORK_ERROR", sdkerrors.ErrCodeNetwork)
	assert.Equal(t, "SERIALIZATION_ERROR", sdkerrors.ErrCodeSerialization)
	assert.Equal(t, "DESERIALIZATION_ERROR", sdkerrors.ErrCodeDeserialization)
	assert.Equal(t, "REQUEST_ERROR", sdkerrors.ErrCodeRequest)
	assert.Equal(t, "RESPONSE_ERROR", sdkerrors.ErrCodeResponse)
	assert.Equal(t, "VALIDATION_ERROR", sdkerrors.ErrCodeValidation)
	assert.Equal(t, "AUTHENTICATION_ERROR", sdkerrors.ErrCodeAuthentication)
	assert.Equal(t, "AUTHORIZATION_ERROR", sdkerrors.ErrCodeAuthorization)
	assert.Equal(t, "NOT_FOUND", sdkerrors.ErrCodeNotFound)
	assert.Equal(t, "RATE_LIMIT_ERROR", sdkerrors.ErrCodeRateLimit)
}

// ---------------------------------------------------------------------------
// SDKError
// ---------------------------------------------------------------------------

func TestNewSDKError_FieldsSet(t *testing.T) {
	underlying := errors.New("underlying cause")
	err := sdkerrors.NewSDKError(sdkerrors.ErrCodeNetwork, "request failed", underlying)

	require.NotNil(t, err)
	assert.Equal(t, sdkerrors.ErrCodeNetwork, err.Code)
	assert.Equal(t, "request failed", err.Message)
	assert.Equal(t, underlying, err.Err)
}

func TestSDKError_Error_WithUnderlyingError(t *testing.T) {
	underlying := errors.New("dial tcp: connection refused")
	err := sdkerrors.NewSDKError(sdkerrors.ErrCodeNetwork, "request failed", underlying)

	msg := err.Error()
	assert.Contains(t, msg, "NETWORK_ERROR")
	assert.Contains(t, msg, "request failed")
	assert.Contains(t, msg, "dial tcp: connection refused")
}

func TestSDKError_Error_WithoutUnderlyingError(t *testing.T) {
	err := sdkerrors.NewSDKError(sdkerrors.ErrCodeValidation, "unsupported currency: \"XYZ\"", nil)

	msg := err.Error()
	assert.Contains(t, msg, "VALIDATION_ERROR")
	assert.Contains(t, msg, "unsupported currency")
	// Should not contain a trailing colon/error when Err is nil
	assert.NotContains(t, msg, "<nil>")
}

func TestSDKError_Unwrap_ReturnsUnderlyingError(t *testing.T) {
	underlying := errors.New("root cause")
	err := sdkerrors.NewSDKError(sdkerrors.ErrCodeRequest, "outer message", underlying)

	assert.Equal(t, underlying, err.Unwrap())
}

func TestSDKError_Unwrap_ReturnsNilWhenNoUnderlying(t *testing.T) {
	err := sdkerrors.NewSDKError(sdkerrors.ErrCodeValidation, "bad currency", nil)
	assert.Nil(t, err.Unwrap())
}

func TestSDKError_ImplementsErrorInterface(t *testing.T) {
	var err error = sdkerrors.NewSDKError(sdkerrors.ErrCodeUnknown, "unknown", nil)
	assert.NotNil(t, err)
	assert.NotEmpty(t, err.Error())
}

func TestSDKError_ErrorsIs_WorksWithUnwrap(t *testing.T) {
	sentinel := errors.New("sentinel")
	wrapped := sdkerrors.NewSDKError(sdkerrors.ErrCodeNetwork, "wrapped", sentinel)

	// errors.Is should find the sentinel through Unwrap
	assert.True(t, errors.Is(wrapped, sentinel))
}

func TestSDKError_AllCodes(t *testing.T) {
	codes := []string{
		sdkerrors.ErrCodeUnknown,
		sdkerrors.ErrCodeNetwork,
		sdkerrors.ErrCodeSerialization,
		sdkerrors.ErrCodeDeserialization,
		sdkerrors.ErrCodeRequest,
		sdkerrors.ErrCodeResponse,
		sdkerrors.ErrCodeValidation,
		sdkerrors.ErrCodeAuthentication,
		sdkerrors.ErrCodeAuthorization,
		sdkerrors.ErrCodeNotFound,
		sdkerrors.ErrCodeRateLimit,
	}

	for _, code := range codes {
		err := sdkerrors.NewSDKError(code, "test message", nil)
		assert.Contains(t, err.Error(), code, "error string should contain the code")
	}
}

// ---------------------------------------------------------------------------
// APIError
// ---------------------------------------------------------------------------

func TestNewAPIError_FieldsSet(t *testing.T) {
	err := sdkerrors.NewAPIError(404, "invoice not found", "NOT_FOUND")

	require.NotNil(t, err)
	assert.Equal(t, 404, err.StatusCode)
	assert.Equal(t, "invoice not found", err.Message)
	assert.Equal(t, "NOT_FOUND", err.Code)
}

func TestAPIError_Error_ContainsAllParts(t *testing.T) {
	err := sdkerrors.NewAPIError(401, "invalid API key", "AUTHENTICATION_ERROR")
	msg := err.Error()

	assert.Contains(t, msg, "401")
	assert.Contains(t, msg, "invalid API key")
	assert.Contains(t, msg, "AUTHENTICATION_ERROR")
}

func TestAPIError_ImplementsErrorInterface(t *testing.T) {
	var err error = sdkerrors.NewAPIError(500, "internal server error", "SERVER_ERROR")
	assert.NotNil(t, err)
	assert.NotEmpty(t, err.Error())
}

func TestAPIError_Various_StatusCodes(t *testing.T) {
	cases := []struct {
		status  int
		message string
		code    string
	}{
		{400, "bad request", "BAD_REQUEST"},
		{401, "unauthorized", "AUTHENTICATION_ERROR"},
		{403, "forbidden", "AUTHORIZATION_ERROR"},
		{404, "not found", "NOT_FOUND"},
		{422, "unprocessable entity", "VALIDATION_ERROR"},
		{429, "rate limit exceeded", "RATE_LIMIT_ERROR"},
		{500, "internal server error", "SERVER_ERROR"},
		{503, "service unavailable", "SERVICE_UNAVAILABLE"},
	}

	for _, tc := range cases {
		err := sdkerrors.NewAPIError(tc.status, tc.message, tc.code)
		assert.Equal(t, tc.status, err.StatusCode)
		assert.Equal(t, tc.message, err.Message)
		assert.Equal(t, tc.code, err.Code)
		assert.Contains(t, err.Error(), tc.message)
	}
}

// ---------------------------------------------------------------------------
// Helper predicates
// ---------------------------------------------------------------------------

func TestIsNotFoundError_TrueFor404(t *testing.T) {
	err := sdkerrors.NewAPIError(404, "not found", "NOT_FOUND")
	assert.True(t, sdkerrors.IsNotFoundError(err))
}

func TestIsNotFoundError_FalseForOtherStatus(t *testing.T) {
	statuses := []int{200, 400, 401, 403, 429, 500}
	for _, s := range statuses {
		err := sdkerrors.NewAPIError(s, "msg", "CODE")
		assert.False(t, sdkerrors.IsNotFoundError(err), "expected false for status %d", s)
	}
}

func TestIsNotFoundError_FalseForNonAPIError(t *testing.T) {
	err := errors.New("plain error")
	assert.False(t, sdkerrors.IsNotFoundError(err))
}

func TestIsNotFoundError_FalseForNil(t *testing.T) {
	assert.False(t, sdkerrors.IsNotFoundError(nil))
}

func TestIsAuthenticatedError_TrueFor401(t *testing.T) {
	err := sdkerrors.NewAPIError(401, "unauthorized", "AUTH_ERROR")
	assert.True(t, sdkerrors.IsAuthenticatedError(err))
}

func TestIsAuthenticatedError_FalseForOtherStatus(t *testing.T) {
	statuses := []int{200, 400, 403, 404, 429, 500}
	for _, s := range statuses {
		err := sdkerrors.NewAPIError(s, "msg", "CODE")
		assert.False(t, sdkerrors.IsAuthenticatedError(err), "expected false for status %d", s)
	}
}

func TestIsAuthenticatedError_FalseForNonAPIError(t *testing.T) {
	err := errors.New("plain error")
	assert.False(t, sdkerrors.IsAuthenticatedError(err))
}

func TestIsRateLimitError_TrueFor429(t *testing.T) {
	err := sdkerrors.NewAPIError(429, "too many requests", "RATE_LIMIT_ERROR")
	assert.True(t, sdkerrors.IsRateLimitError(err))
}

func TestIsRateLimitError_FalseForOtherStatus(t *testing.T) {
	statuses := []int{200, 400, 401, 404, 500}
	for _, s := range statuses {
		err := sdkerrors.NewAPIError(s, "msg", "CODE")
		assert.False(t, sdkerrors.IsRateLimitError(err), "expected false for status %d", s)
	}
}

func TestIsRateLimitError_FalseForNonAPIError(t *testing.T) {
	err := errors.New("plain error")
	assert.False(t, sdkerrors.IsRateLimitError(err))
}

func TestIsRateLimitError_FalseForSDKError(t *testing.T) {
	err := sdkerrors.NewSDKError(sdkerrors.ErrCodeNetwork, "msg", nil)
	assert.False(t, sdkerrors.IsRateLimitError(err))
}
