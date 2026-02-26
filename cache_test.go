package renderscreenshot

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCacheGet(t *testing.T) {
	imageData := []byte{0x89, 0x50, 0x4E, 0x47}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/cache/key123" {
			t.Errorf("path = %q, want /v1/cache/key123", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(imageData)
	}))
	defer server.Close()

	cm := NewCacheManager(newHTTPClient("test_key", server.URL, 10*time.Second, 0, 1.0))
	data, err := cm.Get(context.Background(), "key123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) != 4 {
		t.Errorf("expected 4 bytes, got %d", len(data))
	}
}

func TestCacheGetNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Not found",
				"code":    "not_found",
			},
		})
	}))
	defer server.Close()

	cm := NewCacheManager(newHTTPClient("test_key", server.URL, 10*time.Second, 0, 1.0))
	data, err := cm.Get(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data != nil {
		t.Errorf("expected nil for not found, got %v", data)
	}
}

func TestCacheDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/cache/key123" {
			t.Errorf("path = %q, want /v1/cache/key123", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"deleted": true})
	}))
	defer server.Close()

	cm := NewCacheManager(newHTTPClient("test_key", server.URL, 10*time.Second, 0, 1.0))
	deleted, err := cm.Delete(context.Background(), "key123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Error("expected deleted to be true")
	}
}

func TestCacheDeleteNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Not found",
				"code":    "not_found",
			},
		})
	}))
	defer server.Close()

	cm := NewCacheManager(newHTTPClient("test_key", server.URL, 10*time.Second, 0, 1.0))
	deleted, err := cm.Delete(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted {
		t.Error("expected deleted to be false for not found")
	}
}

func TestCachePurge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/cache/purge" {
			t.Errorf("path = %q, want /v1/cache/purge", r.URL.Path)
		}

		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)

		keys := body["keys"].([]interface{})
		if len(keys) != 2 {
			t.Errorf("expected 2 keys, got %d", len(keys))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"purged": 2.0,
			"keys":   []interface{}{"key1", "key2"},
		})
	}))
	defer server.Close()

	cm := NewCacheManager(newHTTPClient("test_key", server.URL, 10*time.Second, 0, 1.0))
	result, err := cm.Purge(context.Background(), []string{"key1", "key2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Purged != 2 {
		t.Errorf("Purged = %d, want 2", result.Purged)
	}
	if len(result.Keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(result.Keys))
	}
}

func TestCachePurgeURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)

		if body["url"] != "https://example.com/*" {
			t.Errorf("url = %v, want https://example.com/*", body["url"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"purged": 5.0})
	}))
	defer server.Close()

	cm := NewCacheManager(newHTTPClient("test_key", server.URL, 10*time.Second, 0, 1.0))
	result, err := cm.PurgeURL(context.Background(), "https://example.com/*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Purged != 5 {
		t.Errorf("Purged = %d, want 5", result.Purged)
	}
}

func TestCachePurgeBefore(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)

		if body["before"] == nil {
			t.Error("expected before field")
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"purged": 10.0})
	}))
	defer server.Close()

	cm := NewCacheManager(newHTTPClient("test_key", server.URL, 10*time.Second, 0, 1.0))
	result, err := cm.PurgeBefore(context.Background(), time.Now().Add(-7*24*time.Hour))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Purged != 10 {
		t.Errorf("Purged = %d, want 10", result.Purged)
	}
}

func TestCachePurgePattern(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)

		if body["pattern"] != "screenshots/2024/01/*" {
			t.Errorf("pattern = %v, want screenshots/2024/01/*", body["pattern"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"purged": 3.0})
	}))
	defer server.Close()

	cm := NewCacheManager(newHTTPClient("test_key", server.URL, 10*time.Second, 0, 1.0))
	result, err := cm.PurgePattern(context.Background(), "screenshots/2024/01/*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Purged != 3 {
		t.Errorf("Purged = %d, want 3", result.Purged)
	}
}
