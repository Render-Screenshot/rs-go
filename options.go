package renderscreenshot

import (
	"fmt"
	"net/url"
	"strings"
)

// TakeOptions is a mutable fluent builder for screenshot options.
// All setter methods return the same pointer for chaining.
type TakeOptions struct {
	// Source
	url  string
	html string

	// Preset
	preset string
	device string

	// Viewport
	width  int
	height int
	scale  float64
	mobile *bool

	// Capture
	fullPage *bool
	element  string
	format   ImageFormat
	quality  int

	// Wait
	waitFor         WaitCondition
	delay           int
	waitForSelector string
	waitForTimeout  int

	// Blocking
	blockAds           *bool
	blockTrackers      *bool
	blockCookieBanners *bool
	blockChatWidgets   *bool
	blockURLs          []string
	blockResources     []string

	// Page manipulation
	injectScript string
	injectStyle  string
	click        string
	hide         []string
	remove       []string

	// Browser emulation
	darkMode      *bool
	reducedMotion *bool
	mediaType     MediaType
	userAgent     string
	timezone      string
	locale        string
	geolocation   *Geolocation

	// Network
	headers    map[string]string
	cookies    []Cookie
	authBasic  *basicAuth
	authBearer string
	bypassCSP  *bool

	// Cache
	cacheTTL     int
	cacheRefresh *bool

	// PDF
	pdfPaperSize       PaperSize
	pdfWidth           string
	pdfHeight          string
	pdfLandscape       *bool
	pdfMargin          *PDFMargin
	pdfScale           float64
	pdfPrintBackground *bool
	pdfPageRanges      string
	pdfHeader          string
	pdfFooter          string
	pdfFitOnePage      *bool
	pdfPreferCSSSize   *bool

	// Storage
	storageEnabled *bool
	storagePath    string
	storageACL     StorageACL
}

type basicAuth struct {
	username string
	password string
}

// URL creates a new TakeOptions with the given URL.
func URL(u string) *TakeOptions {
	return &TakeOptions{url: u}
}

// HTML creates a new TakeOptions with the given HTML content.
func HTML(html string) *TakeOptions {
	return &TakeOptions{html: html}
}

// FromConfig creates a new TakeOptions from a map of key-value pairs.
func FromConfig(config map[string]interface{}) *TakeOptions {
	opts := &TakeOptions{}

	if v, ok := config["url"].(string); ok {
		opts.url = v
	}
	if v, ok := config["html"].(string); ok {
		opts.html = v
	}
	if v, ok := config["preset"].(string); ok {
		opts.preset = v
	}
	if v, ok := config["device"].(string); ok {
		opts.device = v
	}
	if v, ok := config["width"].(int); ok {
		opts.width = v
	}
	if v, ok := config["height"].(int); ok {
		opts.height = v
	}
	if v, ok := config["format"].(string); ok {
		opts.format = ImageFormat(v)
	}

	return opts
}

// Preset sets the screenshot preset.
func (o *TakeOptions) Preset(value string) *TakeOptions {
	o.preset = value
	return o
}

// Device sets the device preset.
func (o *TakeOptions) Device(value string) *TakeOptions {
	o.device = value
	return o
}

// Width sets the viewport width.
func (o *TakeOptions) Width(value int) *TakeOptions {
	o.width = value
	return o
}

// Height sets the viewport height.
func (o *TakeOptions) Height(value int) *TakeOptions {
	o.height = value
	return o
}

// Scale sets the device pixel ratio.
func (o *TakeOptions) Scale(value float64) *TakeOptions {
	o.scale = value
	return o
}

// Mobile enables or disables mobile viewport emulation.
func (o *TakeOptions) Mobile(value ...bool) *TakeOptions {
	v := true
	if len(value) > 0 {
		v = value[0]
	}
	o.mobile = &v
	return o
}

// FullPage enables or disables full page capture.
func (o *TakeOptions) FullPage(value ...bool) *TakeOptions {
	v := true
	if len(value) > 0 {
		v = value[0]
	}
	o.fullPage = &v
	return o
}

// Element sets a CSS selector to capture a specific element.
func (o *TakeOptions) Element(selector string) *TakeOptions {
	o.element = selector
	return o
}

// Format sets the output image format.
func (o *TakeOptions) Format(value ImageFormat) *TakeOptions {
	o.format = value
	return o
}

// Quality sets the output quality (0-100, JPEG/WebP only).
func (o *TakeOptions) Quality(value int) *TakeOptions {
	o.quality = value
	return o
}

// WaitFor sets the page load wait condition.
func (o *TakeOptions) WaitFor(value WaitCondition) *TakeOptions {
	o.waitFor = value
	return o
}

// Delay sets an additional delay in milliseconds after page load.
func (o *TakeOptions) Delay(value int) *TakeOptions {
	o.delay = value
	return o
}

// WaitForSelector sets a CSS selector to wait for before capture.
func (o *TakeOptions) WaitForSelector(selector string) *TakeOptions {
	o.waitForSelector = selector
	return o
}

// WaitForTimeout sets the maximum wait timeout in milliseconds.
func (o *TakeOptions) WaitForTimeout(value int) *TakeOptions {
	o.waitForTimeout = value
	return o
}

// BlockAds enables or disables ad blocking.
func (o *TakeOptions) BlockAds(value ...bool) *TakeOptions {
	v := true
	if len(value) > 0 {
		v = value[0]
	}
	o.blockAds = &v
	return o
}

// BlockTrackers enables or disables tracker blocking.
func (o *TakeOptions) BlockTrackers(value ...bool) *TakeOptions {
	v := true
	if len(value) > 0 {
		v = value[0]
	}
	o.blockTrackers = &v
	return o
}

// BlockCookieBanners enables or disables cookie banner blocking.
func (o *TakeOptions) BlockCookieBanners(value ...bool) *TakeOptions {
	v := true
	if len(value) > 0 {
		v = value[0]
	}
	o.blockCookieBanners = &v
	return o
}

// BlockChatWidgets enables or disables chat widget blocking.
func (o *TakeOptions) BlockChatWidgets(value ...bool) *TakeOptions {
	v := true
	if len(value) > 0 {
		v = value[0]
	}
	o.blockChatWidgets = &v
	return o
}

// BlockURLs sets URL patterns to block.
func (o *TakeOptions) BlockURLs(patterns []string) *TakeOptions {
	o.blockURLs = patterns
	return o
}

// BlockResources sets resource types to block (e.g., "font", "media", "image").
func (o *TakeOptions) BlockResources(types []string) *TakeOptions {
	o.blockResources = types
	return o
}

// InjectScript sets JavaScript to inject before capture.
func (o *TakeOptions) InjectScript(script string) *TakeOptions {
	o.injectScript = script
	return o
}

// InjectStyle sets CSS to inject before capture.
func (o *TakeOptions) InjectStyle(style string) *TakeOptions {
	o.injectStyle = style
	return o
}

// Click sets a CSS selector to click before capture.
func (o *TakeOptions) Click(selector string) *TakeOptions {
	o.click = selector
	return o
}

// Hide sets CSS selectors to visually hide before capture.
func (o *TakeOptions) Hide(selectors []string) *TakeOptions {
	o.hide = selectors
	return o
}

// Remove sets CSS selectors to remove from the DOM before capture.
func (o *TakeOptions) Remove(selectors []string) *TakeOptions {
	o.remove = selectors
	return o
}

// DarkMode enables or disables dark mode emulation.
func (o *TakeOptions) DarkMode(value ...bool) *TakeOptions {
	v := true
	if len(value) > 0 {
		v = value[0]
	}
	o.darkMode = &v
	return o
}

// ReducedMotion enables or disables reduced motion emulation.
func (o *TakeOptions) ReducedMotion(value ...bool) *TakeOptions {
	v := true
	if len(value) > 0 {
		v = value[0]
	}
	o.reducedMotion = &v
	return o
}

// SetMediaType sets the CSS media type emulation.
func (o *TakeOptions) SetMediaType(value MediaType) *TakeOptions {
	o.mediaType = value
	return o
}

// UserAgent sets a custom user agent string.
func (o *TakeOptions) UserAgent(value string) *TakeOptions {
	o.userAgent = value
	return o
}

// Timezone sets the timezone for browser emulation (IANA format).
func (o *TakeOptions) Timezone(value string) *TakeOptions {
	o.timezone = value
	return o
}

// Locale sets the locale for browser emulation (e.g., "en-US").
func (o *TakeOptions) Locale(value string) *TakeOptions {
	o.locale = value
	return o
}

// SetGeolocation sets the geolocation for browser emulation.
func (o *TakeOptions) SetGeolocation(lat, lon float64, accuracy ...int) *TakeOptions {
	geo := &Geolocation{Latitude: lat, Longitude: lon}
	if len(accuracy) > 0 {
		geo.Accuracy = accuracy[0]
	}
	o.geolocation = geo
	return o
}

// Headers sets custom HTTP headers.
func (o *TakeOptions) Headers(value map[string]string) *TakeOptions {
	o.headers = value
	return o
}

// Cookies sets browser cookies.
func (o *TakeOptions) Cookies(value []Cookie) *TakeOptions {
	o.cookies = value
	return o
}

// AuthBasic sets HTTP basic authentication credentials.
func (o *TakeOptions) AuthBasic(username, password string) *TakeOptions {
	o.authBasic = &basicAuth{username: username, password: password}
	return o
}

// AuthBearer sets a bearer token for authentication.
func (o *TakeOptions) AuthBearer(token string) *TakeOptions {
	o.authBearer = token
	return o
}

// BypassCSP enables or disables CSP bypass.
func (o *TakeOptions) BypassCSP(value ...bool) *TakeOptions {
	v := true
	if len(value) > 0 {
		v = value[0]
	}
	o.bypassCSP = &v
	return o
}

// CacheTTL sets the cache TTL in seconds.
func (o *TakeOptions) CacheTTL(value int) *TakeOptions {
	o.cacheTTL = value
	return o
}

// CacheRefresh forces a cache refresh.
func (o *TakeOptions) CacheRefresh(value ...bool) *TakeOptions {
	v := true
	if len(value) > 0 {
		v = value[0]
	}
	o.cacheRefresh = &v
	return o
}

// PDFPaperSize sets the PDF paper size.
func (o *TakeOptions) PDFPaperSize(value PaperSize) *TakeOptions {
	o.pdfPaperSize = value
	return o
}

// PDFWidth sets a custom PDF width (e.g., "210mm", "8.5in").
func (o *TakeOptions) PDFWidth(value string) *TakeOptions {
	o.pdfWidth = value
	return o
}

// PDFHeight sets a custom PDF height (e.g., "297mm", "11in").
func (o *TakeOptions) PDFHeight(value string) *TakeOptions {
	o.pdfHeight = value
	return o
}

// PDFLandscape enables or disables landscape PDF orientation.
func (o *TakeOptions) PDFLandscape(value ...bool) *TakeOptions {
	v := true
	if len(value) > 0 {
		v = value[0]
	}
	o.pdfLandscape = &v
	return o
}

// PDFMarginUniform sets a uniform PDF margin on all sides (e.g., "2cm").
func (o *TakeOptions) PDFMarginUniform(value string) *TakeOptions {
	m := UniformMargin(value)
	o.pdfMargin = &m
	return o
}

// PDFMarginSides sets individual PDF margins for each side.
func (o *TakeOptions) PDFMarginSides(top, right, bottom, left string) *TakeOptions {
	m := SidesMargin(top, right, bottom, left)
	o.pdfMargin = &m
	return o
}

// PDFScale sets the PDF rendering scale (0.1-2.0).
func (o *TakeOptions) PDFScale(value float64) *TakeOptions {
	o.pdfScale = value
	return o
}

// PDFPrintBackground enables or disables printing background graphics.
func (o *TakeOptions) PDFPrintBackground(value ...bool) *TakeOptions {
	v := true
	if len(value) > 0 {
		v = value[0]
	}
	o.pdfPrintBackground = &v
	return o
}

// PDFPageRanges sets the page ranges to include (e.g., "1-5", "1,3,5").
func (o *TakeOptions) PDFPageRanges(value string) *TakeOptions {
	o.pdfPageRanges = value
	return o
}

// PDFHeader sets HTML content for the PDF header.
func (o *TakeOptions) PDFHeader(value string) *TakeOptions {
	o.pdfHeader = value
	return o
}

// PDFFooter sets HTML content for the PDF footer.
func (o *TakeOptions) PDFFooter(value string) *TakeOptions {
	o.pdfFooter = value
	return o
}

// PDFFitOnePage enables or disables fitting content to one page.
func (o *TakeOptions) PDFFitOnePage(value ...bool) *TakeOptions {
	v := true
	if len(value) > 0 {
		v = value[0]
	}
	o.pdfFitOnePage = &v
	return o
}

// PDFPreferCSSPageSize enables or disables preferring CSS page size.
func (o *TakeOptions) PDFPreferCSSPageSize(value ...bool) *TakeOptions {
	v := true
	if len(value) > 0 {
		v = value[0]
	}
	o.pdfPreferCSSSize = &v
	return o
}

// StorageEnabled enables or disables cloud storage.
func (o *TakeOptions) StorageEnabled(value ...bool) *TakeOptions {
	v := true
	if len(value) > 0 {
		v = value[0]
	}
	o.storageEnabled = &v
	return o
}

// StoragePath sets the storage path pattern (e.g., "screenshots/{date}/{hash}.{ext}").
func (o *TakeOptions) StoragePath(value string) *TakeOptions {
	o.storagePath = value
	return o
}

// StorageACL sets the storage access control level.
func (o *TakeOptions) StorageACL(value StorageACL) *TakeOptions {
	o.storageACL = value
	return o
}

// ToParams converts the options to nested JSON params for POST requests.
// The structure matches the API's expected JSON format.
func (o *TakeOptions) ToParams() map[string]interface{} {
	result := map[string]interface{}{}

	// Top-level params
	if o.url != "" {
		result["url"] = o.url
	}
	if o.html != "" {
		result["html"] = o.html
	}
	if o.preset != "" {
		result["preset"] = o.preset
	}

	// Viewport group
	viewport := map[string]interface{}{}
	if o.width != 0 {
		viewport["width"] = o.width
	}
	if o.height != 0 {
		viewport["height"] = o.height
	}
	if o.scale != 0 {
		viewport["scale"] = o.scale
	}
	if o.mobile != nil {
		viewport["mobile"] = *o.mobile
	}
	if o.device != "" {
		viewport["device"] = o.device
	}
	if len(viewport) > 0 {
		result["viewport"] = viewport
	}

	// Capture group
	capture := map[string]interface{}{}
	if o.fullPage != nil && *o.fullPage {
		capture["mode"] = "full_page"
	}
	if o.element != "" {
		capture["selector"] = o.element
	}
	if len(capture) > 0 {
		result["capture"] = capture
	}

	// Output group
	output := map[string]interface{}{}
	if o.format != "" {
		output["format"] = string(o.format)
	}
	if o.quality != 0 {
		output["quality"] = o.quality
	}
	if len(output) > 0 {
		result["output"] = output
	}

	// Wait group
	wait := map[string]interface{}{}
	if o.waitFor != "" {
		wait["until"] = string(o.waitFor)
	}
	if o.delay != 0 {
		wait["delay"] = o.delay
	}
	if o.waitForSelector != "" {
		wait["for_selector"] = o.waitForSelector
	}
	if o.waitForTimeout != 0 {
		wait["timeout"] = o.waitForTimeout
	}
	if len(wait) > 0 {
		result["wait"] = wait
	}

	// Block group
	block := map[string]interface{}{}
	if o.blockAds != nil {
		block["ads"] = *o.blockAds
	}
	if o.blockTrackers != nil {
		block["trackers"] = *o.blockTrackers
	}
	if o.blockCookieBanners != nil {
		block["cookie_banners"] = *o.blockCookieBanners
	}
	if o.blockChatWidgets != nil {
		block["chat_widgets"] = *o.blockChatWidgets
	}
	if len(o.blockURLs) > 0 {
		block["requests"] = o.blockURLs
	}
	if len(o.blockResources) > 0 {
		block["resources"] = o.blockResources
	}
	if len(block) > 0 {
		result["block"] = block
	}

	// Page group
	page := map[string]interface{}{}
	if o.injectScript != "" {
		page["scripts"] = []string{o.injectScript}
	}
	if o.injectStyle != "" {
		page["styles"] = []string{o.injectStyle}
	}
	if o.click != "" {
		page["click"] = o.click
	}
	if len(o.hide) > 0 {
		page["hide"] = o.hide
	}
	if len(o.remove) > 0 {
		page["remove"] = o.remove
	}
	if len(page) > 0 {
		result["page"] = page
	}

	// Browser group
	browser := map[string]interface{}{}
	if o.darkMode != nil {
		browser["dark_mode"] = *o.darkMode
	}
	if o.reducedMotion != nil {
		browser["reduced_motion"] = *o.reducedMotion
	}
	if o.mediaType != "" {
		browser["media"] = string(o.mediaType)
	}
	if o.userAgent != "" {
		browser["user_agent"] = o.userAgent
	}
	if o.timezone != "" {
		browser["timezone"] = o.timezone
	}
	if o.locale != "" {
		browser["locale"] = o.locale
	}
	if o.geolocation != nil {
		geo := map[string]interface{}{
			"latitude":  o.geolocation.Latitude,
			"longitude": o.geolocation.Longitude,
		}
		if o.geolocation.Accuracy != 0 {
			geo["accuracy"] = o.geolocation.Accuracy
		}
		browser["geolocation"] = geo
	}
	if len(browser) > 0 {
		result["browser"] = browser
	}

	// Network group
	network := map[string]interface{}{}
	if o.headers != nil {
		network["headers"] = o.headers
	}
	if len(o.cookies) > 0 {
		network["cookies"] = o.cookies
	}
	if o.bypassCSP != nil {
		network["bypass_csp"] = *o.bypassCSP
	}
	if o.authBasic != nil {
		network["auth"] = map[string]interface{}{
			"type":     "basic",
			"username": o.authBasic.username,
			"password": o.authBasic.password,
		}
	} else if o.authBearer != "" {
		network["auth"] = map[string]interface{}{
			"type":  "bearer",
			"token": o.authBearer,
		}
	}
	if len(network) > 0 {
		result["network"] = network
	}

	// Cache group
	cache := map[string]interface{}{}
	if o.cacheTTL != 0 {
		cache["ttl"] = o.cacheTTL
	}
	if o.cacheRefresh != nil {
		cache["refresh"] = *o.cacheRefresh
	}
	if len(cache) > 0 {
		result["cache"] = cache
	}

	// PDF group
	pdf := map[string]interface{}{}
	if o.pdfPaperSize != "" {
		pdf["paper"] = string(o.pdfPaperSize)
	}
	if o.pdfWidth != "" {
		pdf["width"] = o.pdfWidth
	}
	if o.pdfHeight != "" {
		pdf["height"] = o.pdfHeight
	}
	if o.pdfLandscape != nil {
		pdf["landscape"] = *o.pdfLandscape
	}
	if o.pdfScale != 0 {
		pdf["scale"] = o.pdfScale
	}
	if o.pdfPrintBackground != nil {
		pdf["background"] = *o.pdfPrintBackground
	}
	if o.pdfPageRanges != "" {
		pdf["page_ranges"] = o.pdfPageRanges
	}
	if o.pdfHeader != "" {
		pdf["header"] = o.pdfHeader
	}
	if o.pdfFooter != "" {
		pdf["footer"] = o.pdfFooter
	}
	if o.pdfFitOnePage != nil {
		pdf["fit_one_page"] = *o.pdfFitOnePage
	}
	if o.pdfPreferCSSSize != nil {
		pdf["prefer_css_page_size"] = *o.pdfPreferCSSSize
	}
	if o.pdfMargin != nil {
		if v := o.pdfMargin.toAPI(); v != nil {
			pdf["margin"] = v
		}
	}
	if len(pdf) > 0 {
		result["pdf"] = pdf
	}

	// Storage group
	storage := map[string]interface{}{}
	if o.storageEnabled != nil {
		storage["enabled"] = *o.storageEnabled
	}
	if o.storagePath != "" {
		storage["path"] = o.storagePath
	}
	if o.storageACL != "" {
		storage["acl"] = string(o.storageACL)
	}
	if len(storage) > 0 {
		result["storage"] = storage
	}

	return result
}

// ToQueryString converts the options to a flat query string for GET requests.
func (o *TakeOptions) ToQueryString() string {
	params := url.Values{}

	if o.url != "" {
		params.Set("url", o.url)
	}
	if o.html != "" {
		params.Set("html", o.html)
	}
	if o.preset != "" {
		params.Set("preset", o.preset)
	}
	if o.width != 0 {
		params.Set("width", fmt.Sprintf("%d", o.width))
	}
	if o.height != 0 {
		params.Set("height", fmt.Sprintf("%d", o.height))
	}
	if o.device != "" {
		params.Set("device", o.device)
	}
	if o.scale != 0 {
		params.Set("scale", formatFloat(o.scale))
	}
	if o.fullPage != nil && *o.fullPage {
		params.Set("full_page", "true")
	}
	if o.element != "" {
		params.Set("selector", o.element)
	}
	if o.format != "" {
		params.Set("format", string(o.format))
	}
	if o.quality != 0 {
		params.Set("quality", fmt.Sprintf("%d", o.quality))
	}
	if o.delay != 0 {
		params.Set("delay", fmt.Sprintf("%d", o.delay))
	}
	if o.waitForTimeout != 0 {
		params.Set("timeout", fmt.Sprintf("%d", o.waitForTimeout))
	}
	if o.blockAds != nil && *o.blockAds {
		params.Set("block_ads", "true")
	}
	if o.blockCookieBanners != nil && *o.blockCookieBanners {
		params.Set("block_cookies", "true")
	}
	if o.darkMode != nil && *o.darkMode {
		params.Set("dark_mode", "true")
	}
	if o.cacheTTL != 0 {
		params.Set("cache_ttl", fmt.Sprintf("%d", o.cacheTTL))
	}

	return params.Encode()
}

// toFlatMap converts the options to a flat key-value map for URL signing.
func (o *TakeOptions) toFlatMap() map[string]string {
	result := map[string]string{}

	if o.url != "" {
		result["url"] = o.url
	}
	if o.html != "" {
		result["html"] = o.html
	}
	if o.preset != "" {
		result["preset"] = o.preset
	}
	if o.width != 0 {
		result["width"] = fmt.Sprintf("%d", o.width)
	}
	if o.height != 0 {
		result["height"] = fmt.Sprintf("%d", o.height)
	}
	if o.device != "" {
		result["device"] = o.device
	}
	if o.scale != 0 {
		result["scale"] = formatFloat(o.scale)
	}
	if o.fullPage != nil && *o.fullPage {
		result["full_page"] = "true"
	}
	if o.element != "" {
		result["selector"] = o.element
	}
	if o.format != "" {
		result["format"] = string(o.format)
	}
	if o.quality != 0 {
		result["quality"] = fmt.Sprintf("%d", o.quality)
	}
	if o.mobile != nil && *o.mobile {
		result["mobile"] = "true"
	}
	if o.waitFor != "" {
		result["wait_for"] = string(o.waitFor)
	}
	if o.delay != 0 {
		result["delay"] = fmt.Sprintf("%d", o.delay)
	}
	if o.blockAds != nil && *o.blockAds {
		result["block_ads"] = "true"
	}
	if o.blockTrackers != nil && *o.blockTrackers {
		result["block_trackers"] = "true"
	}
	if o.darkMode != nil && *o.darkMode {
		result["dark_mode"] = "true"
	}
	if o.cacheTTL != 0 {
		result["cache_ttl"] = fmt.Sprintf("%d", o.cacheTTL)
	}

	return result
}

func formatFloat(f float64) string {
	s := fmt.Sprintf("%g", f)
	if !strings.Contains(s, ".") {
		s += ".0"
	}
	return s
}
