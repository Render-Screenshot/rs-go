package renderscreenshot

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPClientGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test_key" {
			t.Errorf("expected Bearer test_key, got %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("User-Agent") == "" {
			t.Error("expected User-Agent header")
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
		})
	}))
	defer server.Close()

	client := newHTTPClient("test_key", server.URL, 10*time.Second, 0, 1.0)
	result, err := client.get("/test", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("expected status ok, got %v", result["status"])
	}
}

func TestHTTPClientPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json, got %s", r.Header.Get("Content-Type"))
		}

		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     "req_123",
			"status": "completed",
		})
	}))
	defer server.Close()

	client := newHTTPClient("test_key", server.URL, 10*time.Second, 0, 1.0)
	result, err := client.post("/screenshot", map[string]interface{}{"url": "https://example.com"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["id"] != "req_123" {
		t.Errorf("expected id req_123, got %v", result["id"])
	}
}

func TestHTTPClientPostBinary(t *testing.T) {
	imageData := []byte{0x89, 0x50, 0x4E, 0x47} // PNG magic bytes

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(imageData)
	}))
	defer server.Close()

	client := newHTTPClient("test_key", server.URL, 10*time.Second, 0, 1.0)
	resp, err := client.postBinary("/screenshot", map[string]interface{}{"url": "https://example.com"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Body) != 4 {
		t.Errorf("expected 4 bytes, got %d", len(resp.Body))
	}
}

func TestHTTPClientDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"deleted": true})
	}))
	defer server.Close()

	client := newHTTPClient("test_key", server.URL, 10*time.Second, 0, 1.0)
	result, err := client.delete("/cache/key1", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["deleted"] != true {
		t.Errorf("expected deleted true, got %v", result["deleted"])
	}
}

func TestHTTPClientErrorResponse(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		body       map[string]interface{}
		retryAfter string
		wantCode   ErrorCode
	}{
		{
			name:   "400 validation",
			status: 400,
			body: map[string]interface{}{
				"error": map[string]interface{}{
					"message": "Invalid URL",
					"code":    "invalid_url",
				},
			},
			wantCode: CodeInvalidURL,
		},
		{
			name:     "401 unauthorized",
			status:   401,
			body:     map[string]interface{}{},
			wantCode: CodeUnauthorized,
		},
		{
			name:     "404 not found",
			status:   404,
			body:     map[string]interface{}{},
			wantCode: CodeNotFound,
		},
		{
			name:       "429 rate limited",
			status:     429,
			body:       map[string]interface{}{},
			retryAfter: "30",
			wantCode:   CodeRateLimited,
		},
		{
			name:     "500 server error",
			status:   500,
			body:     map[string]interface{}{},
			wantCode: CodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.retryAfter != "" {
					w.Header().Set("Retry-After", tt.retryAfter)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				_ = json.NewEncoder(w).Encode(tt.body)
			}))
			defer server.Close()

			client := newHTTPClient("test_key", server.URL, 10*time.Second, 0, 1.0)
			_, err := client.get("/test", nil, nil)
			if err == nil {
				t.Fatal("expected error")
			}

			apiErr, ok := err.(*Error)
			if !ok {
				t.Fatalf("expected *Error, got %T", err)
			}
			if apiErr.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q", apiErr.Code, tt.wantCode)
			}
			if apiErr.HTTPStatus != tt.status {
				t.Errorf("HTTPStatus = %d, want %d", apiErr.HTTPStatus, tt.status)
			}
		})
	}
}

func TestHTTPClientRetryAfterHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(429)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer server.Close()

	client := newHTTPClient("test_key", server.URL, 10*time.Second, 0, 1.0)
	_, err := client.get("/test", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}

	apiErr := err.(*Error)
	if apiErr.RetryAfter != 60 {
		t.Errorf("RetryAfter = %d, want 60", apiErr.RetryAfter)
	}
}

func TestHTTPClientRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"message": "Internal error",
					"code":    "internal_error",
				},
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})
	}))
	defer server.Close()

	// Use very small retry delay for testing
	client := newHTTPClient("test_key", server.URL, 10*time.Second, 3, 0.01)
	result, err := client.get("/test", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error after retries: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("expected status ok, got %v", result["status"])
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestHTTPClientNoRetryOnNonRetryable(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(401)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Unauthorized",
				"code":    "unauthorized",
			},
		})
	}))
	defer server.Close()

	client := newHTTPClient("test_key", server.URL, 10*time.Second, 3, 0.01)
	_, err := client.get("/test", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt (no retry), got %d", attempts)
	}
}

func TestHTTPClientExtraHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("expected Accept header, got %s", r.Header.Get("Accept"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})
	}))
	defer server.Close()

	client := newHTTPClient("test_key", server.URL, 10*time.Second, 0, 1.0)
	_, err := client.get("/test", nil, map[string]string{"Accept": "application/json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHTTPClientQueryParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("key") != "value" {
			t.Errorf("expected query param key=value, got %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})
	}))
	defer server.Close()

	client := newHTTPClient("test_key", server.URL, 10*time.Second, 0, 1.0)
	_, err := client.get("/test", map[string]string{"key": "value"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHTTPClientUserAgent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		if ua != "renderscreenshot-go/"+Version {
			t.Errorf("User-Agent = %q, want renderscreenshot-go/%s", ua, Version)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer server.Close()

	client := newHTTPClient("test_key", server.URL, 10*time.Second, 0, 1.0)
	_, _ = client.get("/test", nil, nil)
}

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"30", 30},
		{"abc", 0},
		{"60", 60},
	}

	for _, tt := range tests {
		got := parseRetryAfter(tt.input)
		if got != tt.want {
			t.Errorf("parseRetryAfter(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestCalculateDelay(t *testing.T) {
	client := newHTTPClient("key", "", 0, 3, 1.0)

	// With retry_after, should use that value
	errWithRetry := &Error{RetryAfter: 60}
	delay := client.calculateDelay(errWithRetry, 0)
	if delay != 60.0 {
		t.Errorf("expected delay 60, got %f", delay)
	}

	// Without retry_after, should use exponential backoff
	errNoRetry := &Error{}
	delay0 := client.calculateDelay(errNoRetry, 0)
	if delay0 < 1.0 || delay0 > 1.5 {
		t.Errorf("attempt 0 delay should be ~1.0-1.5, got %f", delay0)
	}

	delay1 := client.calculateDelay(errNoRetry, 1)
	if delay1 < 2.0 || delay1 > 2.5 {
		t.Errorf("attempt 1 delay should be ~2.0-2.5, got %f", delay1)
	}
}

func TestHTTPClientEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	client := newHTTPClient("test_key", server.URL, 10*time.Second, 0, 1.0)
	result, err := client.get("/test", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}
