# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-02-25

### Added

- Initial release of the RenderScreenshot Go SDK
- `Client` with functional options (`WithBaseURL`, `WithTimeout`, `WithMaxRetries`, etc.)
- `Take` and `TakeJSON` methods for capturing screenshots
- `GenerateURL` for signed URL generation (HMAC-SHA256)
- `Batch` and `BatchAdvanced` for batch screenshot processing
- `GetBatch` for checking batch job status
- `Presets`, `Preset`, `Devices`, and `Usage` metadata methods
- `CacheManager` with `Get`, `Delete`, `Purge`, `PurgeURL`, `PurgeBefore`, `PurgePattern`
- `TakeOptions` fluent builder with full API coverage (viewport, capture, wait, blocking, page manipulation, browser emulation, network, cache, PDF, storage)
- `VerifyWebhook` with HMAC-SHA256 verification and timing-safe comparison
- `ParseWebhook` and `ExtractWebhookHeaders` utilities
- `Error` type with `IsRetryable()` method and helper functions (`IsNotFound`, `IsRateLimited`, `IsAuthentication`, `IsValidation`)
- Automatic retry with exponential backoff and jitter
- Zero external dependencies (standard library only)
- CI pipeline with Go 1.21, 1.22, 1.23 matrix testing
