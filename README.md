# RenderScreenshot Go SDK

The official Go SDK for [RenderScreenshot](https://renderscreenshot.com) - a developer-friendly screenshot API for capturing web pages programmatically.

## Installation

```bash
go get github.com/Render-Screenshot/rs-go
```

## Requirements

- Go 1.21 or higher
- Zero external dependencies (standard library only)

## Quick Start

```go
package main

import (
	"context"
	"os"

	rs "github.com/Render-Screenshot/rs-go"
)

func main() {
	// Create a client
	client, err := rs.New("rs_live_your_api_key")
	if err != nil {
		panic(err)
	}

	// Take a screenshot
	data, err := client.Take(context.Background(),
		rs.URL("https://example.com").
			Width(1200).
			Height(630).
			Format(rs.FormatPNG))
	if err != nil {
		panic(err)
	}

	// Save to file
	os.WriteFile("screenshot.png", data, 0644)
}
```

## Configuration

### Client Options

```go
client, err := rs.New("rs_live_your_api_key",
	rs.WithBaseURL("https://custom.api.com"),
	rs.WithTimeout(60 * time.Second),
	rs.WithMaxRetries(3),
	rs.WithRetryDelay(2.0),
	rs.WithSigningKey("rs_secret_your_key"),
	rs.WithPublicKeyID("rs_pub_your_id"),
)
```

## Usage

### Taking Screenshots

#### Binary Response

```go
// Get screenshot as binary data
data, err := client.Take(ctx,
	rs.URL("https://example.com").Preset("og_card"))
```

#### JSON Response (with metadata)

```go
// Get screenshot URL and metadata
resp, err := client.TakeJSON(ctx,
	rs.URL("https://example.com").Preset("og_card"))

fmt.Println(resp.Image.URL)    // CDN URL
fmt.Println(resp.Image.Width)  // 1200
fmt.Println(resp.Image.Height) // 630
fmt.Println(resp.Cache.Key)    // Cache key
```

### Screenshot Options

The `TakeOptions` type provides a fluent builder for configuring screenshots:

```go
opts := rs.URL("https://example.com").
	// Viewport
	Width(1920).
	Height(1080).
	Scale(2).
	Mobile().
	// Capture
	FullPage().
	Element("#main-content").
	Format(rs.FormatWebP).
	Quality(90).
	// Wait conditions
	WaitFor(rs.WaitNetworkIdle).
	Delay(500).
	WaitForSelector(".loaded").
	// Blocking
	BlockAds().
	BlockTrackers().
	BlockCookieBanners().
	// Browser emulation
	DarkMode().
	Timezone("America/New_York").
	Locale("en-US")
```

### Using Presets

```go
// Social card presets
opts := rs.URL("https://example.com").Preset("og_card") // 1200x630

// Device presets
opts = rs.URL("https://example.com").Device("iphone_14_pro")
```

### PDF Generation

```go
opts := rs.URL("https://example.com").
	Format(rs.FormatPDF).
	PDFPaperSize(rs.PaperA4).
	PDFLandscape().
	PDFPrintBackground().
	PDFMarginUniform("1cm")

data, err := client.Take(ctx, opts)
os.WriteFile("document.pdf", data, 0644)
```

### Batch Processing

```go
// Simple batch (same options for all URLs)
resp, err := client.Batch(ctx,
	[]string{"https://example1.com", "https://example2.com"},
	rs.URL("").Preset("og_card"))

for _, result := range resp.Results {
	fmt.Printf("%s: %s\n", result.URL, result.Status)
}

// Advanced batch (per-URL options)
resp, err = client.BatchAdvanced(ctx, []rs.BatchRequest{
	{URL: "https://example1.com", Options: rs.URL("").Preset("og_card")},
	{URL: "https://example2.com", Options: rs.URL("").Device("iphone_14_pro")},
})
```

### Signed URLs

Generate signed URLs for client-side use without exposing your API key:

```go
client, _ := rs.New("rs_live_key",
	rs.WithSigningKey("rs_secret_your_key"),
	rs.WithPublicKeyID("rs_pub_your_id"))

opts := rs.URL("https://example.com").Preset("og_card")
signedURL, err := client.GenerateURL(opts, time.Now().Add(24*time.Hour), "", "")

// Use in HTML: <img src="signedURL" />
```

### Cache Management

```go
cache := client.Cache()

// Get cached screenshot
data, err := cache.Get(ctx, "cache_key_123")

// Delete cached entry
deleted, err := cache.Delete(ctx, "cache_key_123")

// Bulk purge
result, err := cache.Purge(ctx, []string{"key1", "key2", "key3"})

// Purge by URL pattern
result, err = cache.PurgeURL(ctx, "https://example.com/*")

// Purge by date
result, err = cache.PurgeBefore(ctx, time.Now().Add(-7*24*time.Hour))

// Purge by storage path pattern
result, err = cache.PurgePattern(ctx, "screenshots/2024/01/*")
```

### Presets and Devices

```go
// List all presets
presets, err := client.Presets(ctx)
for _, p := range presets {
	fmt.Printf("%s: %s\n", p.ID, p.Name)
}

// Get specific preset
preset, err := client.Preset(ctx, "og_card")

// List all devices
devices, err := client.Devices(ctx)
for _, d := range devices {
	fmt.Printf("%s: %s (%dx%d)\n", d.ID, d.Name, d.Width, d.Height)
}
```

### Webhook Verification

```go
func webhookHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	payload := string(body)

	// Extract headers
	headers := rs.ExtractWebhookHeaders(map[string]string{
		"X-Webhook-Signature": r.Header.Get("X-Webhook-Signature"),
		"X-Webhook-Timestamp": r.Header.Get("X-Webhook-Timestamp"),
		"X-Webhook-ID":        r.Header.Get("X-Webhook-ID"),
	})

	// Verify signature
	if !rs.VerifyWebhook(payload, headers.Signature, headers.Timestamp,
		os.Getenv("WEBHOOK_SECRET"), rs.DefaultTolerance) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Parse event
	event, err := rs.ParseWebhook(payload)
	if err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	switch event.Event {
	case "screenshot.completed":
		// Handle completed screenshot
	case "screenshot.failed":
		// Handle failed screenshot
	case "batch.completed":
		// Handle completed batch
	}

	w.WriteHeader(http.StatusOK)
}
```

## Error Handling

All API errors are returned as `*renderscreenshot.Error`:

```go
data, err := client.Take(ctx, rs.URL("https://example.com"))
if err != nil {
	var apiErr *rs.Error
	if errors.As(err, &apiErr) {
		fmt.Println("Status:", apiErr.HTTPStatus)
		fmt.Println("Code:", apiErr.Code)
		fmt.Println("Message:", apiErr.Message)
		fmt.Println("Request ID:", apiErr.RequestID)

		if apiErr.IsRetryable() {
			// Safe to retry
		}
		if apiErr.RetryAfter > 0 {
			// Wait this many seconds before retrying
		}
	}
}

// Helper functions for error checking
if rs.IsNotFound(err) { /* 404 */ }
if rs.IsRetryable(err) { /* transient error */ }
if rs.IsRateLimited(err) { /* 429 */ }
if rs.IsAuthentication(err) { /* 401 */ }
if rs.IsValidation(err) { /* 400/422 */ }
```

Error properties:
- `HTTPStatus` - HTTP status code
- `Code` - Error code from API
- `Message` - Human-readable message
- `RequestID` - Request ID for support
- `RetryAfter` - Seconds to wait (rate limits)
- `IsRetryable()` - Whether the error can be retried

## Complete Options Reference

| Category | Methods |
|----------|---------|
| **Viewport** | `Width`, `Height`, `Scale`, `Mobile` |
| **Capture** | `FullPage`, `Element`, `Format`, `Quality` |
| **Wait** | `WaitFor`, `Delay`, `WaitForSelector`, `WaitForTimeout` |
| **Preset** | `Preset`, `Device` |
| **Blocking** | `BlockAds`, `BlockTrackers`, `BlockCookieBanners`, `BlockChatWidgets`, `BlockURLs`, `BlockResources` |
| **Page** | `InjectScript`, `InjectStyle`, `Click`, `Hide`, `Remove` |
| **Browser** | `DarkMode`, `ReducedMotion`, `SetMediaType`, `UserAgent`, `Timezone`, `Locale`, `SetGeolocation` |
| **Network** | `Headers`, `Cookies`, `AuthBasic`, `AuthBearer`, `BypassCSP` |
| **Cache** | `CacheTTL`, `CacheRefresh` |
| **PDF** | `PDFPaperSize`, `PDFWidth`, `PDFHeight`, `PDFLandscape`, `PDFMarginUniform`, `PDFMarginSides`, `PDFScale`, `PDFPrintBackground`, `PDFPageRanges`, `PDFHeader`, `PDFFooter`, `PDFFitOnePage`, `PDFPreferCSSPageSize` |
| **Storage** | `StorageEnabled`, `StoragePath`, `StorageACL` |

## Development

```bash
go test -v -race ./...     # Run tests
go vet ./...               # Run vet
golangci-lint run          # Run linter
```

### Pre-commit Hooks

This project uses pre-commit hooks to prevent secrets from being committed:

```bash
# Install pre-commit (if not already installed)
brew install pre-commit    # macOS
pip install pre-commit     # or via pip

# Install the hooks
pre-commit install

# Run manually on all files
pre-commit run --all-files
```

## Contributing

Bug reports and pull requests are welcome on GitHub at https://github.com/Render-Screenshot/rs-go.

## License

Available as open source under the terms of the [MIT License](https://opensource.org/licenses/MIT).
