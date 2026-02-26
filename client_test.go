package renderscreenshot

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client, err := New("rs_live_test_key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClientEmptyKey(t *testing.T) {
	_, err := New("")
	if err == nil {
		t.Fatal("expected error for empty API key")
	}

	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.Code != CodeUnauthorized {
		t.Errorf("Code = %q, want %q", apiErr.Code, CodeUnauthorized)
	}
}

func TestNewClientWithOptions(t *testing.T) {
	client, err := New("rs_live_test_key",
		WithBaseURL("https://custom.api.com"),
		WithTimeout(60*time.Second),
		WithSigningKey("rs_secret_123"),
		WithPublicKeyID("rs_pub_456"),
		WithMaxRetries(3),
		WithRetryDelay(2.0),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.http.baseURL != "https://custom.api.com" {
		t.Errorf("baseURL = %q, want https://custom.api.com", client.http.baseURL)
	}
	if client.http.timeout != 60*time.Second {
		t.Errorf("timeout = %v, want 60s", client.http.timeout)
	}
	if client.signingKey != "rs_secret_123" {
		t.Errorf("signingKey = %q, want rs_secret_123", client.signingKey)
	}
	if client.publicKeyID != "rs_pub_456" {
		t.Errorf("publicKeyID = %q, want rs_pub_456", client.publicKeyID)
	}
}

func TestClientTake(t *testing.T) {
	imageData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/screenshot" {
			t.Errorf("path = %q, want /v1/screenshot", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}

		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["url"] != "https://example.com" {
			t.Errorf("url = %v, want https://example.com", body["url"])
		}

		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(imageData)
	}))
	defer server.Close()

	client, _ := New("rs_live_test", WithBaseURL(server.URL))
	data, err := client.Take(context.Background(),
		URL("https://example.com").Width(1200).Format(FormatPNG))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) != len(imageData) {
		t.Errorf("expected %d bytes, got %d", len(imageData), len(data))
	}
}

func TestClientTakeJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("expected Accept: application/json, got %s", r.Header.Get("Accept"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     "req_123",
			"status": "completed",
			"image": map[string]interface{}{
				"url":    "https://cdn.example.com/img.png",
				"width":  1200.0,
				"height": 630.0,
			},
			"cache": map[string]interface{}{
				"hit": true,
				"key": "cache_abc",
			},
		})
	}))
	defer server.Close()

	client, _ := New("rs_live_test", WithBaseURL(server.URL))
	resp, err := client.TakeJSON(context.Background(),
		URL("https://example.com").Preset("og_card"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "req_123" {
		t.Errorf("ID = %q, want req_123", resp.ID)
	}
	if resp.Status != "completed" {
		t.Errorf("Status = %q, want completed", resp.Status)
	}
	if resp.Image.URL != "https://cdn.example.com/img.png" {
		t.Errorf("Image.URL = %q", resp.Image.URL)
	}
	if resp.Image.Width != 1200 {
		t.Errorf("Image.Width = %d, want 1200", resp.Image.Width)
	}
	if resp.Image.Height != 630 {
		t.Errorf("Image.Height = %d, want 630", resp.Image.Height)
	}
	if !resp.Cache.Hit {
		t.Error("Cache.Hit = false, want true")
	}
	if resp.Cache.Key != "cache_abc" {
		t.Errorf("Cache.Key = %q, want cache_abc", resp.Cache.Key)
	}
}

func TestClientGenerateURL(t *testing.T) {
	client, _ := New("rs_live_test",
		WithBaseURL("https://api.renderscreenshot.com"),
		WithSigningKey("rs_secret_test123"),
		WithPublicKeyID("rs_pub_test456"),
	)

	expires := time.Unix(1700000000, 0)
	signedURL, err := client.GenerateURL(
		URL("https://example.com").Preset("og_card"),
		expires, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(signedURL, "https://api.renderscreenshot.com/v1/screenshot?") {
		t.Errorf("URL should start with base URL, got %q", signedURL)
	}
	if !strings.Contains(signedURL, "signature=") {
		t.Error("URL should contain signature")
	}
	if !strings.Contains(signedURL, "key_id=rs_pub_test456") {
		t.Error("URL should contain key_id")
	}
	if !strings.Contains(signedURL, "expires=1700000000") {
		t.Error("URL should contain expires")
	}
}

func TestClientGenerateURLNoCredentials(t *testing.T) {
	client, _ := New("rs_live_test")
	_, err := client.GenerateURL(URL("https://example.com"), time.Now(), "", "")
	if err == nil {
		t.Fatal("expected error when no signing credentials")
	}
}

func TestClientGenerateURLOverrideCredentials(t *testing.T) {
	client, _ := New("rs_live_test",
		WithBaseURL("https://api.renderscreenshot.com"))

	expires := time.Unix(1700000000, 0)
	signedURL, err := client.GenerateURL(
		URL("https://example.com"),
		expires,
		"rs_secret_override",
		"rs_pub_override",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(signedURL, "key_id=rs_pub_override") {
		t.Error("URL should contain overridden key_id")
	}
}

func TestClientBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/batch" {
			t.Errorf("path = %q, want /v1/batch", r.URL.Path)
		}

		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)

		urls := body["urls"].([]interface{})
		if len(urls) != 2 {
			t.Errorf("expected 2 URLs, got %d", len(urls))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":        "batch_123",
			"status":    "queued",
			"total":     2.0,
			"completed": 0.0,
			"failed":    0.0,
			"results":   []interface{}{},
		})
	}))
	defer server.Close()

	client, _ := New("rs_live_test", WithBaseURL(server.URL))
	resp, err := client.Batch(context.Background(),
		[]string{"https://site1.com", "https://site2.com"},
		URL("").Preset("og_card"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "batch_123" {
		t.Errorf("ID = %q, want batch_123", resp.ID)
	}
	if resp.Status != "queued" {
		t.Errorf("Status = %q, want queued", resp.Status)
	}
	if resp.Total != 2 {
		t.Errorf("Total = %d, want 2", resp.Total)
	}
}

func TestClientBatchAdvanced(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)

		requests := body["requests"].([]interface{})
		if len(requests) != 2 {
			t.Errorf("expected 2 requests, got %d", len(requests))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     "batch_456",
			"status": "queued",
			"total":  2.0,
		})
	}))
	defer server.Close()

	client, _ := New("rs_live_test", WithBaseURL(server.URL))
	resp, err := client.BatchAdvanced(context.Background(), []BatchRequest{
		{URL: "https://site1.com", Options: URL("").Preset("og_card")},
		{URL: "https://site2.com", Options: URL("").Device("iphone_14_pro")},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "batch_456" {
		t.Errorf("ID = %q, want batch_456", resp.ID)
	}
}

func TestClientGetBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/batch/batch_789" {
			t.Errorf("path = %q, want /v1/batch/batch_789", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":        "batch_789",
			"status":    "completed",
			"total":     2.0,
			"completed": 2.0,
			"failed":    0.0,
			"results": []interface{}{
				map[string]interface{}{
					"url":       "https://site1.com",
					"status":    "completed",
					"image_url": "https://cdn.example.com/1.png",
				},
				map[string]interface{}{
					"url":       "https://site2.com",
					"status":    "completed",
					"image_url": "https://cdn.example.com/2.png",
				},
			},
		})
	}))
	defer server.Close()

	client, _ := New("rs_live_test", WithBaseURL(server.URL))
	resp, err := client.GetBatch(context.Background(), "batch_789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "completed" {
		t.Errorf("Status = %q, want completed", resp.Status)
	}
	if len(resp.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(resp.Results))
	}
	if resp.Results[0].ImageURL != "https://cdn.example.com/1.png" {
		t.Errorf("Results[0].ImageURL = %q", resp.Results[0].ImageURL)
	}
}

func TestClientPresets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/presets" {
			t.Errorf("path = %q, want /v1/presets", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"presets": []interface{}{
				map[string]interface{}{"id": "og_card", "name": "Open Graph Card", "width": 1200.0, "height": 630.0},
				map[string]interface{}{"id": "twitter_card", "name": "Twitter Card", "width": 1200.0, "height": 600.0},
			},
		})
	}))
	defer server.Close()

	client, _ := New("rs_live_test", WithBaseURL(server.URL))
	presets, err := client.Presets(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(presets) != 2 {
		t.Fatalf("expected 2 presets, got %d", len(presets))
	}
	if presets[0].ID != "og_card" {
		t.Errorf("presets[0].ID = %q, want og_card", presets[0].ID)
	}
	if presets[0].Width != 1200 {
		t.Errorf("presets[0].Width = %d, want 1200", presets[0].Width)
	}
}

func TestClientPreset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/presets/og_card" {
			t.Errorf("path = %q, want /v1/presets/og_card", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "og_card", "name": "Open Graph Card", "width": 1200.0, "height": 630.0,
		})
	}))
	defer server.Close()

	client, _ := New("rs_live_test", WithBaseURL(server.URL))
	preset, err := client.Preset(context.Background(), "og_card")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if preset.ID != "og_card" {
		t.Errorf("ID = %q, want og_card", preset.ID)
	}
}

func TestClientDevices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"devices": []interface{}{
				map[string]interface{}{"id": "iphone_14_pro", "name": "iPhone 14 Pro", "width": 393.0, "height": 852.0},
			},
		})
	}))
	defer server.Close()

	client, _ := New("rs_live_test", WithBaseURL(server.URL))
	devices, err := client.Devices(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}
	if devices[0].ID != "iphone_14_pro" {
		t.Errorf("devices[0].ID = %q, want iphone_14_pro", devices[0].ID)
	}
}

func TestClientUsage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"credits":      10000.0,
			"used":         2500.0,
			"remaining":    7500.0,
			"period_start": "2026-01-01T00:00:00Z",
			"period_end":   "2026-02-01T00:00:00Z",
		})
	}))
	defer server.Close()

	client, _ := New("rs_live_test", WithBaseURL(server.URL))
	usage, err := client.Usage(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if usage.Credits != 10000 {
		t.Errorf("Credits = %d, want 10000", usage.Credits)
	}
	if usage.Used != 2500 {
		t.Errorf("Used = %d, want 2500", usage.Used)
	}
	if usage.Remaining != 7500 {
		t.Errorf("Remaining = %d, want 7500", usage.Remaining)
	}
}

func TestClientCache(t *testing.T) {
	client, _ := New("rs_live_test")
	cache1 := client.Cache()
	cache2 := client.Cache()
	if cache1 != cache2 {
		t.Error("Cache() should return the same instance")
	}
}
