package renderscreenshot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL    = "https://api.renderscreenshot.com"
	defaultTimeout    = 30 * time.Second
	defaultRetryDelay = 1.0  // seconds
	maxRetryDelay     = 30.0 // seconds
)

// httpClient is the internal HTTP wrapper for API requests.
type httpClient struct {
	apiKey     string
	baseURL    string
	timeout    time.Duration
	maxRetries int
	retryDelay float64
	client     *http.Client
	userAgent  string
}

// httpResponse wraps an HTTP response with parsed data.
type httpResponse struct {
	Body    []byte
	Headers http.Header
}

func newHTTPClient(apiKey, baseURL string, timeout time.Duration, maxRetries int, retryDelay float64) *httpClient {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if timeout == 0 {
		timeout = defaultTimeout
	}
	if retryDelay == 0 {
		retryDelay = defaultRetryDelay
	}

	return &httpClient{
		apiKey:     apiKey,
		baseURL:    strings.TrimRight(baseURL, "/"),
		timeout:    timeout,
		maxRetries: maxRetries,
		retryDelay: retryDelay,
		client:     &http.Client{Timeout: timeout},
		userAgent:  fmt.Sprintf("renderscreenshot-go/%s", Version),
	}
}

func (c *httpClient) get(path string, params, headers map[string]string) (map[string]interface{}, error) {
	return c.requestJSON(http.MethodGet, path, params, nil, headers)
}

func (c *httpClient) getBinary(path string, params, headers map[string]string) (*httpResponse, error) {
	return c.requestBinary(http.MethodGet, path, params, nil, headers)
}

func (c *httpClient) post(path string, body interface{}, headers map[string]string) (map[string]interface{}, error) {
	return c.requestJSON(http.MethodPost, path, nil, body, headers)
}

func (c *httpClient) postBinary(path string, body interface{}, headers map[string]string) (*httpResponse, error) {
	return c.requestBinary(http.MethodPost, path, nil, body, headers)
}

func (c *httpClient) delete(path string, params, headers map[string]string) (map[string]interface{}, error) {
	return c.requestJSON(http.MethodDelete, path, params, nil, headers)
}

func (c *httpClient) requestJSON(method, path string, params map[string]string, body interface{}, headers map[string]string) (map[string]interface{}, error) {
	respBody, _, err := c.doWithRetry(method, path, params, body, headers)
	if err != nil {
		return nil, err
	}

	if len(respBody) == 0 {
		return map[string]interface{}{}, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		// If not valid JSON, return raw body as string
		return map[string]interface{}{"body": string(respBody)}, nil
	}
	return result, nil
}

func (c *httpClient) requestBinary(method, path string, params map[string]string, body interface{}, headers map[string]string) (*httpResponse, error) {
	respBody, respHeaders, err := c.doWithRetry(method, path, params, body, headers)
	if err != nil {
		return nil, err
	}

	return &httpResponse{
		Body:    respBody,
		Headers: respHeaders,
	}, nil
}

func (c *httpClient) doWithRetry(method, path string, params map[string]string, body interface{}, headers map[string]string) ([]byte, http.Header, error) {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		respBody, respHeaders, err := c.doRequest(method, path, params, body, headers)
		if err == nil {
			return respBody, respHeaders, nil
		}

		lastErr = err
		apiErr, ok := err.(*Error)
		if !ok || !apiErr.IsRetryable() || attempt >= c.maxRetries {
			return nil, nil, err
		}

		delay := c.calculateDelay(apiErr, attempt)
		time.Sleep(time.Duration(delay * float64(time.Second)))
	}

	return nil, nil, lastErr
}

func (c *httpClient) doRequest(method, path string, params map[string]string, body interface{}, extraHeaders map[string]string) ([]byte, http.Header, error) {
	reqURL := c.baseURL + path

	// Add query params
	if len(params) > 0 {
		parts := make([]string, 0, len(params))
		for k, v := range params {
			parts = append(parts, fmt.Sprintf("%s=%s", k, v))
		}
		reqURL += "?" + strings.Join(parts, "&")
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, nil, &Error{Message: "failed to marshal request body", Code: CodeInvalidRequest}
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, reqURL, bodyReader)
	if err != nil {
		return nil, nil, &Error{Message: "failed to create request: " + err.Error(), Code: CodeConnectionError}
	}

	// Set standard headers
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set extra headers
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		if isTimeoutError(err) {
			return nil, nil, &Error{Message: "Request timed out", Code: CodeTimeout, HTTPStatus: 408}
		}
		return nil, nil, &Error{Message: "Failed to connect to server: " + err.Error(), Code: CodeConnectionError}
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, &Error{Message: "failed to read response body: " + err.Error(), Code: CodeConnectionError}
	}

	if resp.StatusCode >= 400 {
		retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
		requestID := resp.Header.Get("X-Request-Id")

		var bodyMap map[string]interface{}
		if err := json.Unmarshal(respBody, &bodyMap); err != nil {
			bodyMap = map[string]interface{}{}
		}

		return nil, nil, errorFromResponse(resp.StatusCode, bodyMap, retryAfter, requestID)
	}

	return respBody, resp.Header, nil
}

func (c *httpClient) calculateDelay(err *Error, attempt int) float64 {
	// Use retry_after if available (from rate limit responses)
	if err.RetryAfter > 0 {
		return float64(err.RetryAfter)
	}

	// Exponential backoff with jitter: base_delay * 2^attempt + random jitter
	calculated := c.retryDelay * math.Pow(2, float64(attempt))
	jitter := rand.Float64() * c.retryDelay * 0.5 //nolint:gosec // weak randomness is fine for jitter
	delay := calculated + jitter

	if delay > maxRetryDelay {
		delay = maxRetryDelay
	}

	return delay
}

func parseRetryAfter(value string) int {
	if value == "" {
		return 0
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return n
}

func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "deadline exceeded")
}
