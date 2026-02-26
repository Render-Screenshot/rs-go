package renderscreenshot

import (
	"encoding/json"
	"testing"
)

func TestURLConstructor(t *testing.T) {
	opts := URL("https://example.com")
	params := opts.ToParams()
	if params["url"] != "https://example.com" {
		t.Errorf("expected url to be https://example.com, got %v", params["url"])
	}
}

func TestHTMLConstructor(t *testing.T) {
	opts := HTML("<h1>Hello</h1>")
	params := opts.ToParams()
	if params["html"] != "<h1>Hello</h1>" {
		t.Errorf("expected html content, got %v", params["html"])
	}
}

func TestFluentChaining(t *testing.T) {
	opts := URL("https://example.com").
		Width(1200).
		Height(630).
		Format(FormatPNG).
		Quality(90).
		BlockAds().
		DarkMode()

	params := opts.ToParams()

	if params["url"] != "https://example.com" {
		t.Errorf("expected url, got %v", params["url"])
	}

	viewport := params["viewport"].(map[string]interface{})
	if viewport["width"] != 1200 {
		t.Errorf("expected width 1200, got %v", viewport["width"])
	}
	if viewport["height"] != 630 {
		t.Errorf("expected height 630, got %v", viewport["height"])
	}

	output := params["output"].(map[string]interface{})
	if output["format"] != "png" {
		t.Errorf("expected format png, got %v", output["format"])
	}
	if output["quality"] != 90 {
		t.Errorf("expected quality 90, got %v", output["quality"])
	}

	block := params["block"].(map[string]interface{})
	if block["ads"] != true {
		t.Errorf("expected ads blocking true, got %v", block["ads"])
	}

	browser := params["browser"].(map[string]interface{})
	if browser["dark_mode"] != true {
		t.Errorf("expected dark_mode true, got %v", browser["dark_mode"])
	}
}

func TestToParamsViewport(t *testing.T) {
	opts := URL("https://example.com").Width(1920).Height(1080).Scale(2.0).Mobile()
	params := opts.ToParams()

	viewport := params["viewport"].(map[string]interface{})
	if viewport["width"] != 1920 {
		t.Errorf("width = %v, want 1920", viewport["width"])
	}
	if viewport["height"] != 1080 {
		t.Errorf("height = %v, want 1080", viewport["height"])
	}
	if viewport["scale"] != 2.0 {
		t.Errorf("scale = %v, want 2.0", viewport["scale"])
	}
	if viewport["mobile"] != true {
		t.Errorf("mobile = %v, want true", viewport["mobile"])
	}
}

func TestToParamsCapture(t *testing.T) {
	opts := URL("https://example.com").FullPage().Element("#main")
	params := opts.ToParams()

	capture := params["capture"].(map[string]interface{})
	if capture["mode"] != "full_page" {
		t.Errorf("mode = %v, want full_page", capture["mode"])
	}
	if capture["selector"] != "#main" {
		t.Errorf("selector = %v, want #main", capture["selector"])
	}
}

func TestToParamsWait(t *testing.T) {
	opts := URL("https://example.com").
		WaitFor(WaitNetworkIdle).
		Delay(500).
		WaitForSelector(".loaded").
		WaitForTimeout(10000)
	params := opts.ToParams()

	wait := params["wait"].(map[string]interface{})
	if wait["until"] != "networkidle" {
		t.Errorf("until = %v, want networkidle", wait["until"])
	}
	if wait["delay"] != 500 {
		t.Errorf("delay = %v, want 500", wait["delay"])
	}
	if wait["for_selector"] != ".loaded" {
		t.Errorf("for_selector = %v, want .loaded", wait["for_selector"])
	}
	if wait["timeout"] != 10000 {
		t.Errorf("timeout = %v, want 10000", wait["timeout"])
	}
}

func TestToParamsBlocking(t *testing.T) {
	opts := URL("https://example.com").
		BlockAds().
		BlockTrackers().
		BlockCookieBanners().
		BlockChatWidgets().
		BlockURLs([]string{"*.ads.com"}).
		BlockResources([]string{"font", "media"})
	params := opts.ToParams()

	block := params["block"].(map[string]interface{})
	if block["ads"] != true {
		t.Errorf("ads = %v, want true", block["ads"])
	}
	if block["trackers"] != true {
		t.Errorf("trackers = %v, want true", block["trackers"])
	}
	if block["cookie_banners"] != true {
		t.Errorf("cookie_banners = %v, want true", block["cookie_banners"])
	}
	if block["chat_widgets"] != true {
		t.Errorf("chat_widgets = %v, want true", block["chat_widgets"])
	}
	requests := block["requests"].([]string)
	if len(requests) != 1 || requests[0] != "*.ads.com" {
		t.Errorf("requests = %v, want [*.ads.com]", requests)
	}
	resources := block["resources"].([]string)
	if len(resources) != 2 {
		t.Errorf("resources = %v, want [font, media]", resources)
	}
}

func TestToParamsPage(t *testing.T) {
	opts := URL("https://example.com").
		InjectScript("console.log('hi')").
		InjectStyle("body { color: red }").
		Click("#btn").
		Hide([]string{".ad", ".popup"}).
		Remove([]string{".banner"})
	params := opts.ToParams()

	page := params["page"].(map[string]interface{})
	scripts := page["scripts"].([]string)
	if len(scripts) != 1 || scripts[0] != "console.log('hi')" {
		t.Errorf("scripts = %v", scripts)
	}
	styles := page["styles"].([]string)
	if len(styles) != 1 || styles[0] != "body { color: red }" {
		t.Errorf("styles = %v", styles)
	}
	if page["click"] != "#btn" {
		t.Errorf("click = %v, want #btn", page["click"])
	}
	hide := page["hide"].([]string)
	if len(hide) != 2 {
		t.Errorf("hide = %v, want [.ad, .popup]", hide)
	}
	remove := page["remove"].([]string)
	if len(remove) != 1 || remove[0] != ".banner" {
		t.Errorf("remove = %v, want [.banner]", remove)
	}
}

func TestToParamsBrowser(t *testing.T) {
	opts := URL("https://example.com").
		DarkMode().
		ReducedMotion().
		SetMediaType(MediaPrint).
		UserAgent("CustomBot/1.0").
		Timezone("America/New_York").
		Locale("en-US").
		SetGeolocation(40.7128, -74.0060, 100)
	params := opts.ToParams()

	browser := params["browser"].(map[string]interface{})
	if browser["dark_mode"] != true {
		t.Errorf("dark_mode = %v, want true", browser["dark_mode"])
	}
	if browser["reduced_motion"] != true {
		t.Errorf("reduced_motion = %v, want true", browser["reduced_motion"])
	}
	if browser["media"] != "print" {
		t.Errorf("media = %v, want print", browser["media"])
	}
	if browser["user_agent"] != "CustomBot/1.0" {
		t.Errorf("user_agent = %v, want CustomBot/1.0", browser["user_agent"])
	}
	if browser["timezone"] != "America/New_York" {
		t.Errorf("timezone = %v, want America/New_York", browser["timezone"])
	}
	if browser["locale"] != "en-US" {
		t.Errorf("locale = %v, want en-US", browser["locale"])
	}

	geo := browser["geolocation"].(map[string]interface{})
	if geo["latitude"] != 40.7128 {
		t.Errorf("latitude = %v, want 40.7128", geo["latitude"])
	}
	if geo["longitude"] != -74.006 {
		t.Errorf("longitude = %v, want -74.006", geo["longitude"])
	}
	if geo["accuracy"] != 100 {
		t.Errorf("accuracy = %v, want 100", geo["accuracy"])
	}
}

func TestToParamsNetwork(t *testing.T) {
	opts := URL("https://example.com").
		Headers(map[string]string{"X-Custom": "value"}).
		AuthBasic("user", "pass").
		BypassCSP()
	params := opts.ToParams()

	network := params["network"].(map[string]interface{})
	headers := network["headers"].(map[string]string)
	if headers["X-Custom"] != "value" {
		t.Errorf("headers = %v", headers)
	}
	if network["bypass_csp"] != true {
		t.Errorf("bypass_csp = %v, want true", network["bypass_csp"])
	}

	auth := network["auth"].(map[string]interface{})
	if auth["type"] != "basic" {
		t.Errorf("auth type = %v, want basic", auth["type"])
	}
	if auth["username"] != "user" {
		t.Errorf("auth username = %v, want user", auth["username"])
	}
}

func TestToParamsNetworkBearer(t *testing.T) {
	opts := URL("https://example.com").AuthBearer("token123")
	params := opts.ToParams()

	network := params["network"].(map[string]interface{})
	auth := network["auth"].(map[string]interface{})
	if auth["type"] != "bearer" {
		t.Errorf("auth type = %v, want bearer", auth["type"])
	}
	if auth["token"] != "token123" {
		t.Errorf("auth token = %v, want token123", auth["token"])
	}
}

func TestToParamsCache(t *testing.T) {
	opts := URL("https://example.com").CacheTTL(3600).CacheRefresh()
	params := opts.ToParams()

	cache := params["cache"].(map[string]interface{})
	if cache["ttl"] != 3600 {
		t.Errorf("ttl = %v, want 3600", cache["ttl"])
	}
	if cache["refresh"] != true {
		t.Errorf("refresh = %v, want true", cache["refresh"])
	}
}

func TestToParamsPDF(t *testing.T) {
	opts := URL("https://example.com").
		Format(FormatPDF).
		PDFPaperSize(PaperA4).
		PDFLandscape().
		PDFPrintBackground().
		PDFMarginUniform("2cm").
		PDFScale(1.5).
		PDFPageRanges("1-5").
		PDFHeader("<h1>Header</h1>").
		PDFFooter("<p>Footer</p>").
		PDFFitOnePage().
		PDFPreferCSSPageSize()
	params := opts.ToParams()

	pdf := params["pdf"].(map[string]interface{})
	if pdf["paper"] != "a4" {
		t.Errorf("paper = %v, want a4", pdf["paper"])
	}
	if pdf["landscape"] != true {
		t.Errorf("landscape = %v, want true", pdf["landscape"])
	}
	if pdf["background"] != true {
		t.Errorf("background = %v, want true", pdf["background"])
	}
	if pdf["margin"] != "2cm" {
		t.Errorf("margin = %v, want 2cm", pdf["margin"])
	}
	if pdf["scale"] != 1.5 {
		t.Errorf("scale = %v, want 1.5", pdf["scale"])
	}
	if pdf["page_ranges"] != "1-5" {
		t.Errorf("page_ranges = %v, want 1-5", pdf["page_ranges"])
	}
	if pdf["header"] != "<h1>Header</h1>" {
		t.Errorf("header = %v", pdf["header"])
	}
	if pdf["footer"] != "<p>Footer</p>" {
		t.Errorf("footer = %v", pdf["footer"])
	}
	if pdf["fit_one_page"] != true {
		t.Errorf("fit_one_page = %v, want true", pdf["fit_one_page"])
	}
	if pdf["prefer_css_page_size"] != true {
		t.Errorf("prefer_css_page_size = %v, want true", pdf["prefer_css_page_size"])
	}
}

func TestToParamsPDFSidesMargin(t *testing.T) {
	opts := URL("https://example.com").
		Format(FormatPDF).
		PDFMarginSides("1cm", "2cm", "3cm", "4cm")
	params := opts.ToParams()

	pdf := params["pdf"].(map[string]interface{})
	margin := pdf["margin"].(map[string]string)
	if margin["top"] != "1cm" {
		t.Errorf("margin top = %v, want 1cm", margin["top"])
	}
	if margin["right"] != "2cm" {
		t.Errorf("margin right = %v, want 2cm", margin["right"])
	}
	if margin["bottom"] != "3cm" {
		t.Errorf("margin bottom = %v, want 3cm", margin["bottom"])
	}
	if margin["left"] != "4cm" {
		t.Errorf("margin left = %v, want 4cm", margin["left"])
	}
}

func TestToParamsStorage(t *testing.T) {
	opts := URL("https://example.com").
		StorageEnabled().
		StoragePath("screenshots/{date}/{hash}.{ext}").
		StorageACL(ACLPublicRead)
	params := opts.ToParams()

	storage := params["storage"].(map[string]interface{})
	if storage["enabled"] != true {
		t.Errorf("enabled = %v, want true", storage["enabled"])
	}
	if storage["path"] != "screenshots/{date}/{hash}.{ext}" {
		t.Errorf("path = %v", storage["path"])
	}
	if storage["acl"] != "public-read" {
		t.Errorf("acl = %v, want public-read", storage["acl"])
	}
}

func TestToParamsEmpty(t *testing.T) {
	opts := &TakeOptions{}
	params := opts.ToParams()
	if len(params) != 0 {
		t.Errorf("expected empty params, got %v", params)
	}
}

func TestToQueryString(t *testing.T) {
	opts := URL("https://example.com").
		Width(1200).
		Format(FormatPNG).
		BlockAds().
		DarkMode()
	qs := opts.ToQueryString()

	if qs == "" {
		t.Fatal("expected non-empty query string")
	}

	// Parse the query string to verify
	values, err := parseQueryString(qs)
	if err != nil {
		t.Fatalf("failed to parse query string: %v", err)
	}

	// url.Values.Encode() URL-encodes the value
	if values["url"] != "https%3A%2F%2Fexample.com" && values["url"] != "https://example.com" {
		t.Errorf("url = %v", values["url"])
	}
	if values["width"] != "1200" {
		t.Errorf("width = %v", values["width"])
	}
	if values["format"] != "png" {
		t.Errorf("format = %v", values["format"])
	}
	if values["block_ads"] != "true" {
		t.Errorf("block_ads = %v", values["block_ads"])
	}
	if values["dark_mode"] != "true" {
		t.Errorf("dark_mode = %v", values["dark_mode"])
	}
}

func TestToParamsJSONSerializable(t *testing.T) {
	opts := URL("https://example.com").
		Width(1200).
		Height(630).
		Format(FormatPNG).
		BlockAds().
		DarkMode().
		Preset("og_card")

	params := opts.ToParams()
	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("failed to marshal params: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty JSON")
	}
}

func TestFromConfig(t *testing.T) {
	config := map[string]interface{}{
		"url":    "https://example.com",
		"preset": "og_card",
		"width":  1200,
		"format": "png",
	}
	opts := FromConfig(config)
	params := opts.ToParams()

	if params["url"] != "https://example.com" {
		t.Errorf("url = %v", params["url"])
	}
	if params["preset"] != "og_card" {
		t.Errorf("preset = %v", params["preset"])
	}
}

func TestBooleanMethodsDefaultTrue(t *testing.T) {
	opts := URL("https://example.com").
		Mobile().
		FullPage().
		BlockAds().
		BlockTrackers().
		BlockCookieBanners().
		BlockChatWidgets().
		DarkMode().
		ReducedMotion().
		BypassCSP().
		CacheRefresh().
		PDFLandscape().
		PDFPrintBackground().
		PDFFitOnePage().
		PDFPreferCSSPageSize().
		StorageEnabled()

	params := opts.ToParams()

	// Verify all boolean fields defaulted to true
	viewport := params["viewport"].(map[string]interface{})
	if viewport["mobile"] != true {
		t.Error("mobile should default to true")
	}

	capture := params["capture"].(map[string]interface{})
	if capture["mode"] != "full_page" {
		t.Error("full_page should enable full_page mode")
	}
}

func TestBooleanMethodsExplicitFalse(t *testing.T) {
	opts := URL("https://example.com").
		Mobile(false).
		BlockAds(false)

	params := opts.ToParams()

	viewport := params["viewport"].(map[string]interface{})
	if viewport["mobile"] != false {
		t.Error("mobile should be false when explicitly set")
	}

	block := params["block"].(map[string]interface{})
	if block["ads"] != false {
		t.Error("ads should be false when explicitly set")
	}
}

func TestMutableBuilder(t *testing.T) {
	opts := URL("https://example.com")
	opts.Width(1200)
	opts.Height(630)

	params := opts.ToParams()
	viewport := params["viewport"].(map[string]interface{})
	if viewport["width"] != 1200 {
		t.Error("mutable builder should retain width")
	}
	if viewport["height"] != 630 {
		t.Error("mutable builder should retain height")
	}
}

// parseQueryString is a test helper to parse query string into a map
func parseQueryString(qs string) (map[string]string, error) {
	result := map[string]string{}
	if qs == "" {
		return result, nil
	}
	for _, part := range splitQueryString(qs) {
		kv := splitKeyValue(part)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}
	return result, nil
}

func splitQueryString(qs string) []string {
	var result []string
	current := ""
	for _, c := range qs {
		if c == '&' {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func splitKeyValue(kv string) []string {
	for i, c := range kv {
		if c == '=' {
			return []string{kv[:i], kv[i+1:]}
		}
	}
	return []string{kv}
}
