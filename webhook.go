package renderscreenshot

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	// SignatureHeader is the HTTP header containing the webhook signature.
	SignatureHeader = "X-Webhook-Signature"
	// TimestampHeader is the HTTP header containing the webhook timestamp.
	TimestampHeader = "X-Webhook-Timestamp"
	// IDHeader is the HTTP header containing the webhook event ID.
	IDHeader = "X-Webhook-ID"
	// DefaultTolerance is the default maximum age for webhook timestamps (5 minutes).
	DefaultTolerance = 300 * time.Second
)

// VerifyWebhook verifies a webhook signature using HMAC-SHA256.
// It checks the timestamp is within the tolerance window and performs
// a timing-safe comparison of the signature.
func VerifyWebhook(payload, signature, timestamp, secret string, tolerance time.Duration) bool {
	if payload == "" || signature == "" || timestamp == "" || secret == "" {
		return false
	}

	if tolerance == 0 {
		tolerance = DefaultTolerance
	}

	// Parse and validate timestamp
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}

	age := time.Now().Unix() - ts
	if age < 0 {
		age = -age
	}
	if age > int64(tolerance.Seconds()) {
		return false
	}

	// Compute expected signature: sha256=HMAC-SHA256("timestamp.payload", secret)
	signedPayload := fmt.Sprintf("%s.%s", timestamp, payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	expectedHash := hex.EncodeToString(mac.Sum(nil))
	expected := "sha256=" + expectedHash

	// Timing-safe comparison
	return subtle.ConstantTimeCompare([]byte(expected), []byte(signature)) == 1
}

// ParseWebhook parses a webhook payload into a WebhookEvent.
func ParseWebhook(payload string) (*WebhookEvent, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		return nil, &Error{
			Message:    "Invalid webhook payload: " + err.Error(),
			HTTPStatus: 400,
			Code:       CodeInvalidRequest,
		}
	}

	event := &WebhookEvent{
		Data: map[string]interface{}{},
	}

	// Event type from "type" or "event" field
	if v, ok := data["type"].(string); ok {
		event.Event = v
	} else if v, ok := data["event"].(string); ok {
		event.Event = v
	}

	if v, ok := data["id"].(string); ok {
		event.ID = v
	}

	if v, ok := data["timestamp"].(float64); ok {
		event.Timestamp = int64(v)
	}

	if v, ok := data["data"].(map[string]interface{}); ok {
		event.Data = v
	}

	return event, nil
}

// WebhookHeaders contains the extracted webhook header values.
type WebhookHeaders struct {
	Signature string
	Timestamp string
	ID        string
}

// ExtractWebhookHeaders extracts the signature, timestamp, and ID from request headers.
// Supports multiple header naming conventions (standard HTTP, lowercase, underscored).
func ExtractWebhookHeaders(headers map[string]string) WebhookHeaders {
	normalized := normalizeHeaders(headers)

	return WebhookHeaders{
		Signature: normalized["x-webhook-signature"],
		Timestamp: normalized["x-webhook-timestamp"],
		ID:        normalized["x-webhook-id"],
	}
}

func normalizeHeaders(headers map[string]string) map[string]string {
	result := make(map[string]string, len(headers))
	for k, v := range headers {
		normalized := strings.ToLower(strings.ReplaceAll(k, "_", "-"))
		result[normalized] = v
	}
	return result
}
