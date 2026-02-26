package renderscreenshot

import (
	"context"
	"time"
)

// CacheManager provides operations for managing cached screenshots.
type CacheManager struct {
	http *httpClient
}

// NewCacheManager creates a new CacheManager with the given HTTP client.
func NewCacheManager(http *httpClient) *CacheManager {
	return &CacheManager{http: http}
}

// Get retrieves a cached screenshot by key. Returns nil if not found.
func (cm *CacheManager) Get(_ context.Context, key string) ([]byte, error) {
	resp, err := cm.http.getBinary("/v1/cache/"+key, nil, nil)
	if err != nil {
		if IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return resp.Body, nil
}

// Delete removes a single cached entry. Returns true if deleted, false if not found.
func (cm *CacheManager) Delete(_ context.Context, key string) (bool, error) {
	_, err := cm.http.delete("/v1/cache/"+key, nil, nil)
	if err != nil {
		if IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Purge removes multiple cache entries by keys.
func (cm *CacheManager) Purge(_ context.Context, keys []string) (*PurgeResult, error) {
	result, err := cm.http.post("/v1/cache/purge", map[string]interface{}{"keys": keys}, nil)
	if err != nil {
		return nil, err
	}
	return parsePurgeResult(result), nil
}

// PurgeURL removes cache entries matching a URL pattern (glob syntax).
func (cm *CacheManager) PurgeURL(_ context.Context, pattern string) (*PurgeResult, error) {
	result, err := cm.http.post("/v1/cache/purge", map[string]interface{}{"url": pattern}, nil)
	if err != nil {
		return nil, err
	}
	return parsePurgeResult(result), nil
}

// PurgeBefore removes cache entries older than the given time.
func (cm *CacheManager) PurgeBefore(_ context.Context, before time.Time) (*PurgeResult, error) {
	dateStr := before.UTC().Format(time.RFC3339)
	result, err := cm.http.post("/v1/cache/purge", map[string]interface{}{"before": dateStr}, nil)
	if err != nil {
		return nil, err
	}
	return parsePurgeResult(result), nil
}

// PurgePattern removes cache entries matching a storage path pattern.
func (cm *CacheManager) PurgePattern(_ context.Context, pattern string) (*PurgeResult, error) {
	result, err := cm.http.post("/v1/cache/purge", map[string]interface{}{"pattern": pattern}, nil)
	if err != nil {
		return nil, err
	}
	return parsePurgeResult(result), nil
}

func parsePurgeResult(m map[string]interface{}) *PurgeResult {
	r := &PurgeResult{}
	if v, ok := m["purged"].(float64); ok {
		r.Purged = int(v)
	}
	if v, ok := m["keys"].([]interface{}); ok {
		for _, k := range v {
			if s, ok := k.(string); ok {
				r.Keys = append(r.Keys, s)
			}
		}
	}
	return r
}
