package renderscreenshot

// ImageFormat represents supported output image formats.
type ImageFormat string

// Supported image formats.
const (
	FormatPNG  ImageFormat = "png"
	FormatJPEG ImageFormat = "jpeg"
	FormatWebP ImageFormat = "webp"
	FormatPDF  ImageFormat = "pdf"
)

// WaitCondition represents supported page load wait conditions.
type WaitCondition string

// Supported wait conditions.
const (
	WaitLoad             WaitCondition = "load"
	WaitDOMContentLoaded WaitCondition = "domcontentloaded"
	WaitNetworkIdle      WaitCondition = "networkidle"
)

// MediaType represents CSS media type emulation values.
type MediaType string

// Supported media types.
const (
	MediaScreen MediaType = "screen"
	MediaPrint  MediaType = "print"
)

// PaperSize represents supported PDF paper sizes.
type PaperSize string

// Supported PDF paper sizes.
const (
	PaperA3     PaperSize = "a3"
	PaperA4     PaperSize = "a4"
	PaperA5     PaperSize = "a5"
	PaperLegal  PaperSize = "legal"
	PaperLetter PaperSize = "letter"
	PaperLedger PaperSize = "ledger"
)

// StorageACL represents storage access control values.
type StorageACL string

// Supported storage ACL values.
const (
	ACLPublicRead StorageACL = "public-read"
	ACLPrivate    StorageACL = "private"
)

// Cookie represents a browser cookie to set before capture.
type Cookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain,omitempty"`
	Path     string `json:"path,omitempty"`
	Secure   bool   `json:"secure,omitempty"`
	HTTPOnly bool   `json:"httpOnly,omitempty"`
}

// Geolocation represents a geographic location for browser emulation.
type Geolocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Accuracy  int     `json:"accuracy,omitempty"`
}

// ScreenshotResponse represents the JSON response from a screenshot request.
type ScreenshotResponse struct {
	ID     string         `json:"id"`
	Status string         `json:"status"`
	Image  ImageInfo      `json:"image"`
	Cache  CacheInfo      `json:"cache"`
	Error  *ErrorResponse `json:"error,omitempty"`
}

// ImageInfo contains details about the captured image.
type ImageInfo struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// CacheInfo contains cache status from a screenshot response.
type CacheInfo struct {
	Hit bool   `json:"hit"`
	Key string `json:"key"`
}

// BatchResponse represents the response from a batch screenshot request.
type BatchResponse struct {
	ID        string         `json:"id"`
	Status    string         `json:"status"`
	Total     int            `json:"total"`
	Completed int            `json:"completed"`
	Failed    int            `json:"failed"`
	Results   []BatchResult  `json:"results"`
	Error     *ErrorResponse `json:"error,omitempty"`
}

// BatchResult represents a single result in a batch response.
type BatchResult struct {
	URL      string `json:"url"`
	Status   string `json:"status"`
	ImageURL string `json:"image_url"`
	Error    string `json:"error,omitempty"`
}

// BatchRequest represents a single request in an advanced batch.
type BatchRequest struct {
	URL     string       `json:"url"`
	Options *TakeOptions `json:"-"`
}

// PresetInfo contains information about a screenshot preset.
type PresetInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// DeviceInfo contains information about a device preset.
type DeviceInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// UsageInfo contains account usage and credits information.
type UsageInfo struct {
	Credits     int    `json:"credits"`
	Used        int    `json:"used"`
	Remaining   int    `json:"remaining"`
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`
}

// WebhookEvent represents a parsed webhook event.
type WebhookEvent struct {
	Event     string                 `json:"event"`
	ID        string                 `json:"id"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// PurgeResult represents the result of a cache purge operation.
type PurgeResult struct {
	Purged int      `json:"purged"`
	Keys   []string `json:"keys,omitempty"`
}

// ErrorResponse represents an error from the API.
type ErrorResponse struct {
	Message   string `json:"message"`
	Code      string `json:"code"`
	RequestID string `json:"request_id,omitempty"`
}

// PDFMargin represents PDF margin settings as either a uniform string or per-side values.
type PDFMargin struct {
	uniform string
	Top     string
	Right   string
	Bottom  string
	Left    string
}

// UniformMargin creates a PDFMargin with the same value on all sides.
func UniformMargin(value string) PDFMargin {
	return PDFMargin{uniform: value}
}

// SidesMargin creates a PDFMargin with individual side values.
func SidesMargin(top, right, bottom, left string) PDFMargin {
	return PDFMargin{Top: top, Right: right, Bottom: bottom, Left: left}
}

// isUniform returns true if this margin uses a uniform value.
func (m *PDFMargin) isUniform() bool {
	return m.uniform != ""
}

// toAPI converts the margin to its API representation.
func (m *PDFMargin) toAPI() interface{} {
	if m.isUniform() {
		return m.uniform
	}
	result := map[string]string{}
	if m.Top != "" {
		result["top"] = m.Top
	}
	if m.Right != "" {
		result["right"] = m.Right
	}
	if m.Bottom != "" {
		result["bottom"] = m.Bottom
	}
	if m.Left != "" {
		result["left"] = m.Left
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
