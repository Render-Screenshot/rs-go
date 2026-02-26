package renderscreenshot

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

func createTestSignature(timestamp, payload, secret string) string {
	signedPayload := fmt.Sprintf("%s.%s", timestamp, payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyWebhookValid(t *testing.T) {
	secret := "whsec_test_secret"
	payload := `{"type":"screenshot.completed","id":"evt_123","data":{"url":"https://example.com"}}`
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	signature := createTestSignature(timestamp, payload, secret)

	if !VerifyWebhook(payload, signature, timestamp, secret, DefaultTolerance) {
		t.Error("expected valid webhook to verify")
	}
}

func TestVerifyWebhookInvalidSignature(t *testing.T) {
	secret := "whsec_test_secret"
	payload := `{"type":"screenshot.completed"}`
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	if VerifyWebhook(payload, "sha256=invalid", timestamp, secret, DefaultTolerance) {
		t.Error("expected invalid signature to fail")
	}
}

func TestVerifyWebhookExpiredTimestamp(t *testing.T) {
	secret := "whsec_test_secret"
	payload := `{"type":"screenshot.completed"}`
	// 10 minutes ago - beyond 5 minute tolerance
	timestamp := fmt.Sprintf("%d", time.Now().Unix()-600)
	signature := createTestSignature(timestamp, payload, secret)

	if VerifyWebhook(payload, signature, timestamp, secret, DefaultTolerance) {
		t.Error("expected expired timestamp to fail")
	}
}

func TestVerifyWebhookFutureTimestamp(t *testing.T) {
	secret := "whsec_test_secret"
	payload := `{"type":"screenshot.completed"}`
	// 10 minutes in the future
	timestamp := fmt.Sprintf("%d", time.Now().Unix()+600)
	signature := createTestSignature(timestamp, payload, secret)

	if VerifyWebhook(payload, signature, timestamp, secret, DefaultTolerance) {
		t.Error("expected future timestamp to fail")
	}
}

func TestVerifyWebhookEmptyParams(t *testing.T) {
	tests := []struct {
		name      string
		payload   string
		signature string
		timestamp string
		secret    string
	}{
		{"empty payload", "", "sha256=abc", "123", "secret"},
		{"empty signature", "payload", "", "123", "secret"},
		{"empty timestamp", "payload", "sha256=abc", "", "secret"},
		{"empty secret", "payload", "sha256=abc", "123", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if VerifyWebhook(tt.payload, tt.signature, tt.timestamp, tt.secret, DefaultTolerance) {
				t.Error("expected verification to fail with empty param")
			}
		})
	}
}

func TestVerifyWebhookInvalidTimestamp(t *testing.T) {
	if VerifyWebhook("payload", "sha256=abc", "not-a-number", "secret", DefaultTolerance) {
		t.Error("expected invalid timestamp to fail")
	}
}

func TestVerifyWebhookCustomTolerance(t *testing.T) {
	secret := "whsec_test_secret"
	payload := `{"type":"screenshot.completed"}`
	// 2 minutes ago
	timestamp := fmt.Sprintf("%d", time.Now().Unix()-120)
	signature := createTestSignature(timestamp, payload, secret)

	// Should pass with 3-minute tolerance
	if !VerifyWebhook(payload, signature, timestamp, secret, 3*time.Minute) {
		t.Error("expected verification to pass with 3-minute tolerance")
	}

	// Should fail with 1-minute tolerance
	if VerifyWebhook(payload, signature, timestamp, secret, 1*time.Minute) {
		t.Error("expected verification to fail with 1-minute tolerance")
	}
}

func TestVerifyWebhookTimingSafe(t *testing.T) {
	secret := "whsec_test_secret"
	payload := `{"type":"screenshot.completed"}`
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	validSig := createTestSignature(timestamp, payload, secret)

	// Valid signature should pass
	if !VerifyWebhook(payload, validSig, timestamp, secret, DefaultTolerance) {
		t.Error("expected valid signature to pass")
	}

	// Tampered signature (same length) should fail
	tampered := validSig[:len(validSig)-1] + "0"
	if VerifyWebhook(payload, tampered, timestamp, secret, DefaultTolerance) {
		t.Error("expected tampered signature to fail")
	}
}

func TestParseWebhook(t *testing.T) {
	payload := `{
		"type": "screenshot.completed",
		"id": "evt_123",
		"timestamp": 1700000000,
		"data": {
			"url": "https://example.com",
			"image_url": "https://cdn.example.com/img.png"
		}
	}`

	event, err := ParseWebhook(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.Event != "screenshot.completed" {
		t.Errorf("Event = %q, want screenshot.completed", event.Event)
	}
	if event.ID != "evt_123" {
		t.Errorf("ID = %q, want evt_123", event.ID)
	}
	if event.Timestamp != 1700000000 {
		t.Errorf("Timestamp = %d, want 1700000000", event.Timestamp)
	}
	if event.Data["url"] != "https://example.com" {
		t.Errorf("Data.url = %v", event.Data["url"])
	}
}

func TestParseWebhookEventField(t *testing.T) {
	payload := `{"event": "batch.completed", "id": "evt_456"}`
	event, err := ParseWebhook(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.Event != "batch.completed" {
		t.Errorf("Event = %q, want batch.completed", event.Event)
	}
}

func TestParseWebhookInvalidJSON(t *testing.T) {
	_, err := ParseWebhook("not-json{")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.Code != CodeInvalidRequest {
		t.Errorf("Code = %q, want %q", apiErr.Code, CodeInvalidRequest)
	}
}

func TestExtractWebhookHeaders(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		wantSig string
		wantTS  string
		wantID  string
	}{
		{
			name: "standard headers",
			headers: map[string]string{
				"X-Webhook-Signature": "sha256=abc",
				"X-Webhook-Timestamp": "12345",
				"X-Webhook-ID":        "evt_123",
			},
			wantSig: "sha256=abc",
			wantTS:  "12345",
			wantID:  "evt_123",
		},
		{
			name: "lowercase headers",
			headers: map[string]string{
				"x-webhook-signature": "sha256=def",
				"x-webhook-timestamp": "67890",
				"x-webhook-id":        "evt_456",
			},
			wantSig: "sha256=def",
			wantTS:  "67890",
			wantID:  "evt_456",
		},
		{
			name: "underscore headers",
			headers: map[string]string{
				"x_webhook_signature": "sha256=ghi",
				"x_webhook_timestamp": "11111",
				"x_webhook_id":        "evt_789",
			},
			wantSig: "sha256=ghi",
			wantTS:  "11111",
			wantID:  "evt_789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := ExtractWebhookHeaders(tt.headers)
			if h.Signature != tt.wantSig {
				t.Errorf("Signature = %q, want %q", h.Signature, tt.wantSig)
			}
			if h.Timestamp != tt.wantTS {
				t.Errorf("Timestamp = %q, want %q", h.Timestamp, tt.wantTS)
			}
			if h.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", h.ID, tt.wantID)
			}
		})
	}
}

func TestExtractWebhookHeadersEmpty(t *testing.T) {
	h := ExtractWebhookHeaders(map[string]string{})
	if h.Signature != "" || h.Timestamp != "" || h.ID != "" {
		t.Error("expected all empty values for empty headers")
	}
}
