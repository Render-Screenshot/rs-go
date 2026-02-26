package renderscreenshot

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

// Client is the RenderScreenshot API client.
type Client struct {
	http        *httpClient
	signingKey  string
	publicKeyID string
	cache       *CacheManager
}

// Option is a functional option for configuring the Client.
type Option func(*clientConfig)

type clientConfig struct {
	baseURL     string
	timeout     time.Duration
	signingKey  string
	publicKeyID string
	maxRetries  int
	retryDelay  float64
}

// WithBaseURL sets a custom API base URL.
func WithBaseURL(baseURL string) Option {
	return func(c *clientConfig) {
		c.baseURL = baseURL
	}
}

// WithTimeout sets the HTTP request timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *clientConfig) {
		c.timeout = timeout
	}
}

// WithSigningKey sets the secret key for signed URL generation (rs_secret_*).
func WithSigningKey(key string) Option {
	return func(c *clientConfig) {
		c.signingKey = key
	}
}

// WithPublicKeyID sets the public key ID for signed URL generation (rs_pub_*).
func WithPublicKeyID(id string) Option {
	return func(c *clientConfig) {
		c.publicKeyID = id
	}
}

// WithMaxRetries sets the maximum number of retries for retryable errors.
func WithMaxRetries(n int) Option {
	return func(c *clientConfig) {
		c.maxRetries = n
	}
}

// WithRetryDelay sets the base delay between retries in seconds.
func WithRetryDelay(delay float64) Option {
	return func(c *clientConfig) {
		c.retryDelay = delay
	}
}

// New creates a new RenderScreenshot client.
func New(apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		return nil, &Error{Message: "Invalid or missing API key", HTTPStatus: 401, Code: CodeUnauthorized}
	}

	cfg := &clientConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	return &Client{
		http:        newHTTPClient(apiKey, cfg.baseURL, cfg.timeout, cfg.maxRetries, cfg.retryDelay),
		signingKey:  cfg.signingKey,
		publicKeyID: cfg.publicKeyID,
	}, nil
}

// Take captures a screenshot and returns the binary image/PDF data.
func (c *Client) Take(_ context.Context, options *TakeOptions) ([]byte, error) {
	params := options.ToParams()
	resp, err := c.http.postBinary("/v1/screenshot", params, nil)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// TakeJSON captures a screenshot and returns the JSON response with metadata.
func (c *Client) TakeJSON(_ context.Context, options *TakeOptions) (*ScreenshotResponse, error) {
	params := options.ToParams()
	result, err := c.http.post("/v1/screenshot", params, map[string]string{"Accept": "application/json"})
	if err != nil {
		return nil, err
	}
	return parseScreenshotResponse(result), nil
}

// GenerateURL creates a signed URL for client-side use without exposing the API key.
func (c *Client) GenerateURL(options *TakeOptions, expiresAt time.Time, signingKey, publicKeyID string) (string, error) {
	secret := signingKey
	if secret == "" {
		secret = c.signingKey
	}
	keyID := publicKeyID
	if keyID == "" {
		keyID = c.publicKeyID
	}

	if secret == "" || keyID == "" {
		return "", &Error{
			Message:    "Signed URLs require signing_key (rs_secret_*) and public_key_id (rs_pub_*). Pass them to New() or to GenerateURL directly.",
			HTTPStatus: 400,
			Code:       CodeInvalidRequest,
		}
	}

	// Build params in alphabetical order
	signParams := map[string]string{
		"expires": fmt.Sprintf("%d", expiresAt.Unix()),
		"key_id":  keyID,
	}

	// Add options as flat params
	flatMap := options.toFlatMap()
	for k, v := range flatMap {
		signParams[k] = v
	}

	// Sort keys for deterministic signature
	keys := make([]string, 0, len(signParams))
	for k := range signParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build query string
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, url.QueryEscape(signParams[k])))
	}
	queryString := strings.Join(parts, "&")

	// Sign with HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(queryString))
	signature := hex.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("%s/v1/screenshot?%s&signature=%s", c.http.baseURL, queryString, signature), nil
}

// Batch processes multiple URLs with the same options.
func (c *Client) Batch(_ context.Context, urls []string, options *TakeOptions) (*BatchResponse, error) {
	body := map[string]interface{}{
		"urls": urls,
	}
	if options != nil {
		body["options"] = options.ToParams()
	}

	result, err := c.http.post("/v1/batch", body, nil)
	if err != nil {
		return nil, err
	}
	return parseBatchResponse(result), nil
}

// BatchAdvanced processes multiple URLs with per-URL options.
func (c *Client) BatchAdvanced(_ context.Context, requests []BatchRequest) (*BatchResponse, error) {
	formatted := make([]map[string]interface{}, 0, len(requests))
	for _, req := range requests {
		entry := map[string]interface{}{
			"url": req.URL,
		}
		if req.Options != nil {
			params := req.Options.ToParams()
			for k, v := range params {
				if k != "url" {
					entry[k] = v
				}
			}
		}
		formatted = append(formatted, entry)
	}

	result, err := c.http.post("/v1/batch", map[string]interface{}{"requests": formatted}, nil)
	if err != nil {
		return nil, err
	}
	return parseBatchResponse(result), nil
}

// GetBatch retrieves the status of a batch job.
func (c *Client) GetBatch(_ context.Context, batchID string) (*BatchResponse, error) {
	result, err := c.http.get("/v1/batch/"+batchID, nil, nil)
	if err != nil {
		return nil, err
	}
	return parseBatchResponse(result), nil
}

// Presets lists all available screenshot presets.
func (c *Client) Presets(_ context.Context) ([]PresetInfo, error) {
	result, err := c.http.get("/v1/presets", nil, nil)
	if err != nil {
		return nil, err
	}

	presetData := result
	if p, ok := result["presets"]; ok {
		if arr, ok := p.([]interface{}); ok {
			return parsePresets(arr), nil
		}
	}
	_ = presetData
	return nil, nil
}

// Preset retrieves a specific preset by ID.
func (c *Client) Preset(_ context.Context, id string) (*PresetInfo, error) {
	result, err := c.http.get("/v1/presets/"+id, nil, nil)
	if err != nil {
		return nil, err
	}
	return parsePresetInfo(result), nil
}

// Devices lists all available device presets.
func (c *Client) Devices(_ context.Context) ([]DeviceInfo, error) {
	result, err := c.http.get("/v1/devices", nil, nil)
	if err != nil {
		return nil, err
	}

	if d, ok := result["devices"]; ok {
		if arr, ok := d.([]interface{}); ok {
			return parseDevices(arr), nil
		}
	}
	return nil, nil
}

// Usage retrieves account usage and credits information.
func (c *Client) Usage(_ context.Context) (*UsageInfo, error) {
	result, err := c.http.get("/v1/usage", nil, nil)
	if err != nil {
		return nil, err
	}
	return parseUsageInfo(result), nil
}

// Cache returns the CacheManager for cache operations.
func (c *Client) Cache() *CacheManager {
	if c.cache == nil {
		c.cache = NewCacheManager(c.http)
	}
	return c.cache
}

// response parsing helpers

func parseScreenshotResponse(m map[string]interface{}) *ScreenshotResponse {
	r := &ScreenshotResponse{}
	if v, ok := m["id"].(string); ok {
		r.ID = v
	}
	if v, ok := m["status"].(string); ok {
		r.Status = v
	}
	if img, ok := m["image"].(map[string]interface{}); ok {
		if v, ok := img["url"].(string); ok {
			r.Image.URL = v
		}
		if v, ok := img["width"].(float64); ok {
			r.Image.Width = int(v)
		}
		if v, ok := img["height"].(float64); ok {
			r.Image.Height = int(v)
		}
	}
	if cache, ok := m["cache"].(map[string]interface{}); ok {
		if v, ok := cache["hit"].(bool); ok {
			r.Cache.Hit = v
		}
		if v, ok := cache["key"].(string); ok {
			r.Cache.Key = v
		}
	}
	return r
}

func parseBatchResponse(m map[string]interface{}) *BatchResponse {
	r := &BatchResponse{}
	if v, ok := m["id"].(string); ok {
		r.ID = v
	}
	if v, ok := m["status"].(string); ok {
		r.Status = v
	}
	if v, ok := m["total"].(float64); ok {
		r.Total = int(v)
	}
	if v, ok := m["completed"].(float64); ok {
		r.Completed = int(v)
	}
	if v, ok := m["failed"].(float64); ok {
		r.Failed = int(v)
	}
	if results, ok := m["results"].([]interface{}); ok {
		for _, item := range results {
			entry, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			br := BatchResult{}
			if v, ok := entry["url"].(string); ok {
				br.URL = v
			}
			if v, ok := entry["status"].(string); ok {
				br.Status = v
			}
			if v, ok := entry["image_url"].(string); ok {
				br.ImageURL = v
			}
			if v, ok := entry["error"].(string); ok {
				br.Error = v
			}
			r.Results = append(r.Results, br)
		}
	}
	return r
}

func parsePresets(arr []interface{}) []PresetInfo {
	presets := make([]PresetInfo, 0, len(arr))
	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		presets = append(presets, *parsePresetInfo(m))
	}
	return presets
}

func parsePresetInfo(m map[string]interface{}) *PresetInfo {
	p := &PresetInfo{}
	if v, ok := m["id"].(string); ok {
		p.ID = v
	}
	if v, ok := m["name"].(string); ok {
		p.Name = v
	}
	if v, ok := m["width"].(float64); ok {
		p.Width = int(v)
	}
	if v, ok := m["height"].(float64); ok {
		p.Height = int(v)
	}
	return p
}

func parseDevices(arr []interface{}) []DeviceInfo {
	devices := make([]DeviceInfo, 0, len(arr))
	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		d := DeviceInfo{}
		if v, ok := m["id"].(string); ok {
			d.ID = v
		}
		if v, ok := m["name"].(string); ok {
			d.Name = v
		}
		if v, ok := m["width"].(float64); ok {
			d.Width = int(v)
		}
		if v, ok := m["height"].(float64); ok {
			d.Height = int(v)
		}
		devices = append(devices, d)
	}
	return devices
}

func parseUsageInfo(m map[string]interface{}) *UsageInfo {
	u := &UsageInfo{}
	if v, ok := m["credits"].(float64); ok {
		u.Credits = int(v)
	}
	if v, ok := m["used"].(float64); ok {
		u.Used = int(v)
	}
	if v, ok := m["remaining"].(float64); ok {
		u.Remaining = int(v)
	}
	if v, ok := m["period_start"].(string); ok {
		u.PeriodStart = v
	}
	if v, ok := m["period_end"].(string); ok {
		u.PeriodEnd = v
	}
	return u
}
