package renderscreenshot

import (
	"errors"
	"testing"
)

func TestErrorMessage(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "with request ID",
			err:  &Error{Message: "Rate limit exceeded", HTTPStatus: 429, Code: CodeRateLimited, RequestID: "req_123"},
			want: "renderscreenshot: Rate limit exceeded (status=429, code=rate_limited, request_id=req_123)",
		},
		{
			name: "with HTTP status",
			err:  &Error{Message: "Not found", HTTPStatus: 404, Code: CodeNotFound},
			want: "renderscreenshot: Not found (status=404, code=not_found)",
		},
		{
			name: "without HTTP status",
			err:  &Error{Message: "Connection failed", Code: CodeConnectionError},
			want: "renderscreenshot: Connection failed (code=connection_error)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestErrorIsRetryable(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want bool
	}{
		{"rate limited", &Error{Code: CodeRateLimited}, true},
		{"timeout", &Error{Code: CodeTimeout}, true},
		{"render failed", &Error{Code: CodeRenderFailed}, true},
		{"internal error", &Error{Code: CodeInternalError}, true},
		{"connection error", &Error{Code: CodeConnectionError}, true},
		{"5xx status", &Error{HTTPStatus: 502, Code: "bad_gateway"}, true},
		{"validation error", &Error{Code: CodeInvalidRequest}, false},
		{"auth error", &Error{Code: CodeUnauthorized}, false},
		{"not found", &Error{Code: CodeNotFound}, false},
		{"forbidden", &Error{Code: CodeForbidden}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.IsRetryable()
			if got != tt.want {
				t.Errorf("IsRetryable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	if IsNotFound(errors.New("random error")) {
		t.Error("IsNotFound should return false for non-Error types")
	}
	if !IsNotFound(&Error{HTTPStatus: 404, Code: CodeNotFound}) {
		t.Error("IsNotFound should return true for 404 errors")
	}
	if IsNotFound(&Error{HTTPStatus: 400, Code: CodeInvalidRequest}) {
		t.Error("IsNotFound should return false for 400 errors")
	}
}

func TestIsRetryableHelper(t *testing.T) {
	if IsRetryable(errors.New("random error")) {
		t.Error("IsRetryable should return false for non-Error types")
	}
	if !IsRetryable(&Error{Code: CodeRateLimited}) {
		t.Error("IsRetryable should return true for rate limited errors")
	}
	if IsRetryable(&Error{Code: CodeUnauthorized}) {
		t.Error("IsRetryable should return false for auth errors")
	}
}

func TestIsRateLimited(t *testing.T) {
	if IsRateLimited(errors.New("random error")) {
		t.Error("IsRateLimited should return false for non-Error types")
	}
	if !IsRateLimited(&Error{HTTPStatus: 429, Code: CodeRateLimited}) {
		t.Error("IsRateLimited should return true for 429 errors")
	}
}

func TestIsAuthentication(t *testing.T) {
	if IsAuthentication(errors.New("random error")) {
		t.Error("IsAuthentication should return false for non-Error types")
	}
	if !IsAuthentication(&Error{HTTPStatus: 401, Code: CodeUnauthorized}) {
		t.Error("IsAuthentication should return true for 401 errors")
	}
}

func TestIsValidation(t *testing.T) {
	if IsValidation(errors.New("random error")) {
		t.Error("IsValidation should return false for non-Error types")
	}
	if !IsValidation(&Error{HTTPStatus: 400, Code: CodeInvalidRequest}) {
		t.Error("IsValidation should return true for 400 errors")
	}
	if IsValidation(&Error{HTTPStatus: 422, Code: CodeRenderFailed}) {
		t.Error("IsValidation should return false for render_failed 422")
	}
	if !IsValidation(&Error{HTTPStatus: 422, Code: CodeInvalidRequest}) {
		t.Error("IsValidation should return true for non-render_failed 422")
	}
}

func TestErrorFromResponse(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		body       map[string]interface{}
		retryAfter int
		requestID  string
		wantCode   ErrorCode
		wantMsg    string
	}{
		{
			name:   "400 with error body",
			status: 400,
			body: map[string]interface{}{
				"error": map[string]interface{}{
					"message": "Invalid URL provided",
					"code":    "invalid_url",
				},
			},
			wantCode: CodeInvalidURL,
			wantMsg:  "Invalid URL provided",
		},
		{
			name:     "401 empty body",
			status:   401,
			body:     map[string]interface{}{},
			wantCode: CodeUnauthorized,
			wantMsg:  "HTTP 401 error",
		},
		{
			name:   "422 render_failed",
			status: 422,
			body: map[string]interface{}{
				"error": map[string]interface{}{
					"message": "Screenshot rendering failed",
					"code":    "render_failed",
				},
			},
			wantCode: CodeRenderFailed,
			wantMsg:  "Screenshot rendering failed",
		},
		{
			name:       "429 with retry_after",
			status:     429,
			body:       map[string]interface{}{},
			retryAfter: 30,
			wantCode:   CodeRateLimited,
		},
		{
			name:      "500 with request_id in body",
			status:    500,
			body:      map[string]interface{}{"request_id": "req_abc"},
			requestID: "",
			wantCode:  CodeInternalError,
		},
		{
			name:      "request_id from header takes precedence",
			status:    500,
			body:      map[string]interface{}{"request_id": "req_body"},
			requestID: "req_header",
			wantCode:  CodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errorFromResponse(tt.status, tt.body, tt.retryAfter, tt.requestID)
			if err.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q", err.Code, tt.wantCode)
			}
			if tt.wantMsg != "" && err.Message != tt.wantMsg {
				t.Errorf("Message = %q, want %q", err.Message, tt.wantMsg)
			}
			if err.HTTPStatus != tt.status {
				t.Errorf("HTTPStatus = %d, want %d", err.HTTPStatus, tt.status)
			}
			if tt.retryAfter != 0 && err.RetryAfter != tt.retryAfter {
				t.Errorf("RetryAfter = %d, want %d", err.RetryAfter, tt.retryAfter)
			}
			if tt.requestID != "" && err.RequestID != tt.requestID {
				t.Errorf("RequestID = %q, want %q", err.RequestID, tt.requestID)
			}
		})
	}
}
