package renderscreenshot

import "fmt"

// ErrorCode represents API error codes.
type ErrorCode string

// API error codes.
const (
	CodeInvalidURL      ErrorCode = "invalid_url"
	CodeInvalidRequest  ErrorCode = "invalid_request"
	CodeMissingRequired ErrorCode = "missing_required"
	CodeUnauthorized    ErrorCode = "unauthorized"
	CodeInvalidAPIKey   ErrorCode = "invalid_api_key"
	CodeExpiredSig      ErrorCode = "expired_signature"
	CodeForbidden       ErrorCode = "forbidden"
	CodeNoCredits       ErrorCode = "insufficient_credits"
	CodeNotFound        ErrorCode = "not_found"
	CodeRateLimited     ErrorCode = "rate_limited"
	CodeTimeout         ErrorCode = "timeout"
	CodeRenderFailed    ErrorCode = "render_failed"
	CodeInternalError   ErrorCode = "internal_error"
	CodeConnectionError ErrorCode = "connection_error"
)

// Error represents an API error from RenderScreenshot.
type Error struct {
	Message    string
	HTTPStatus int
	Code       ErrorCode
	RequestID  string
	RetryAfter int
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("renderscreenshot: %s (status=%d, code=%s, request_id=%s)", e.Message, e.HTTPStatus, e.Code, e.RequestID)
	}
	if e.HTTPStatus != 0 {
		return fmt.Sprintf("renderscreenshot: %s (status=%d, code=%s)", e.Message, e.HTTPStatus, e.Code)
	}
	return fmt.Sprintf("renderscreenshot: %s (code=%s)", e.Message, e.Code)
}

// IsRetryable returns true if the error represents a transient failure that can be retried.
func (e *Error) IsRetryable() bool {
	switch e.Code {
	case CodeRateLimited, CodeTimeout, CodeRenderFailed, CodeInternalError, CodeConnectionError:
		return true
	}
	if e.HTTPStatus >= 500 && e.HTTPStatus < 600 {
		return true
	}
	return false
}

// IsNotFound returns true if the error represents a 404 not found response.
func IsNotFound(err error) bool {
	e, ok := err.(*Error)
	if !ok {
		return false
	}
	return e.HTTPStatus == 404 || e.Code == CodeNotFound
}

// IsRetryable returns true if the error is retryable.
func IsRetryable(err error) bool {
	e, ok := err.(*Error)
	if !ok {
		return false
	}
	return e.IsRetryable()
}

// IsRateLimited returns true if the error represents a rate limit response.
func IsRateLimited(err error) bool {
	e, ok := err.(*Error)
	if !ok {
		return false
	}
	return e.Code == CodeRateLimited || e.HTTPStatus == 429
}

// IsAuthentication returns true if the error represents an authentication failure.
func IsAuthentication(err error) bool {
	e, ok := err.(*Error)
	if !ok {
		return false
	}
	return e.HTTPStatus == 401
}

// IsValidation returns true if the error represents a validation failure.
func IsValidation(err error) bool {
	e, ok := err.(*Error)
	if !ok {
		return false
	}
	return e.HTTPStatus == 400 || (e.HTTPStatus == 422 && e.Code != CodeRenderFailed)
}

// errorFromResponse creates an Error from an HTTP response status and body.
func errorFromResponse(httpStatus int, body map[string]interface{}, retryAfter int, requestID string) *Error {
	message := fmt.Sprintf("HTTP %d error", httpStatus)
	var code ErrorCode

	if errObj, ok := body["error"]; ok {
		if errMap, ok := errObj.(map[string]interface{}); ok {
			if msg, ok := errMap["message"].(string); ok {
				message = msg
			}
			if c, ok := errMap["code"].(string); ok {
				code = ErrorCode(c)
			}
			if rid, ok := errMap["request_id"].(string); ok && requestID == "" {
				requestID = rid
			}
		}
	}

	if requestID == "" {
		if rid, ok := body["request_id"].(string); ok {
			requestID = rid
		}
	}

	if code == "" {
		switch {
		case httpStatus == 400:
			code = CodeInvalidRequest
		case httpStatus == 401:
			code = CodeUnauthorized
		case httpStatus == 403:
			code = CodeForbidden
		case httpStatus == 404:
			code = CodeNotFound
		case httpStatus == 408:
			code = CodeTimeout
		case httpStatus == 422:
			code = CodeInvalidRequest
		case httpStatus == 429:
			code = CodeRateLimited
		case httpStatus >= 500:
			code = CodeInternalError
		}
	}

	return &Error{
		Message:    message,
		HTTPStatus: httpStatus,
		Code:       code,
		RequestID:  requestID,
		RetryAfter: retryAfter,
	}
}
