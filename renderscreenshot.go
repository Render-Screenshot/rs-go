// Package renderscreenshot provides the official Go SDK for the RenderScreenshot API.
//
// RenderScreenshot is a developer-friendly screenshot API for capturing web pages
// programmatically. This SDK provides a type-safe, idiomatic Go client with support
// for screenshots, PDF generation, batch processing, cache management, signed URLs,
// and webhook verification.
//
// # Quick Start
//
//	client := renderscreenshot.New("rs_live_your_api_key")
//
//	data, err := client.Take(ctx, renderscreenshot.URL("https://example.com").
//		Width(1200).
//		Height(630).
//		Format(renderscreenshot.FormatPNG))
//
// Zero external dependencies — uses only the Go standard library.
package renderscreenshot

// Version is the current SDK version.
const Version = "1.0.0"
