package browser

import (
	"compress/gzip"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type IDGenerator interface {
	Generate() string
}

type idGenerator struct{}

func NewIDGenerator() IDGenerator {
	return &idGenerator{}
}

func (g *idGenerator) Generate() string {
	bytes := make([]byte, IDByteLength)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("id_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x", bytes)
}

type URLNormalizer interface {
	Normalize(rawURL string) (string, error)
	IsValidScheme(scheme string) bool
}

type urlNormalizer struct {
	supportedSchemes map[string]bool
}

func NewURLNormalizer() URLNormalizer {
	return &urlNormalizer{
		supportedSchemes: map[string]bool{
			"http":  true,
			"https": true,
		},
	}
}

func (n *urlNormalizer) Normalize(rawURL string) (string, error) {
	if rawURL == "" {
		return "", ErrInvalidURL
	}

	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("%s: %w", ErrInvalidURL, err)
	}

	return parsed.String(), nil
}

func (n *urlNormalizer) IsValidScheme(scheme string) bool {
	return n.supportedSchemes[scheme]
}

type ColorParser interface {
	ParseColor(color string) (r, g, b, a uint8, err error)
	IsValidColor(color string) bool
	NormalizeColor(color string) string
}

type colorParser struct {
	namedColors map[string]string
}

func NewColorParser() ColorParser {
	return &colorParser{
		namedColors: map[string]string{
			"black":   "#000000",
			"white":   "#FFFFFF",
			"red":     "#FF0000",
			"green":   "#008000",
			"blue":    "#0000FF",
			"yellow":  "#FFFF00",
			"cyan":    "#00FFFF",
			"magenta": "#FF00FF",
			"silver":  "#C0C0C0",
			"gray":    "#808080",
			"maroon":  "#800000",
			"olive":   "#808000",
			"lime":    "#00FF00",
			"aqua":    "#00FFFF",
			"teal":    "#008080",
			"navy":    "#000080",
			"fuchsia": "#FF00FF",
			"purple":  "#800080",
		},
	}
}

func (cp *colorParser) ParseColor(color string) (r, g, b, a uint8, err error) {
	color = strings.TrimSpace(strings.ToLower(color))

	if hex, ok := cp.namedColors[color]; ok {
		color = hex
	}

	if strings.HasPrefix(color, "#") {
		return cp.parseHexColor(color)
	}

	if strings.HasPrefix(color, "rgb") {
		return cp.parseRGBColor(color)
	}

	return 0, 0, 0, 0, fmt.Errorf("unsupported color format: %s", color)
}

func (cp *colorParser) parseHexColor(color string) (r, g, b, a uint8, err error) {
	hex := color[1:]

	rgb, err := strconv.ParseUint(hex, HexBase, 32)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	switch len(hex) {
	case HexColorShortLength: // #RGB
		return cp.parseShortHex(rgb)
	case HexColorFullLength: // #RRGGBB
		return cp.parseFullHex(rgb)
	case HexColorAlphaLength: // #RRGGBBAA
		return cp.parseAlphaHex(rgb)
	default:
		return 0, 0, 0, 0, fmt.Errorf("invalid hex color format")
	}
}

func (cp *colorParser) parseShortHex(rgb uint64) (r, g, b, a uint8, err error) {
	r = uint8((rgb >> MaxColorBits) & 0xF * RGBShortMultiplier)
	g = uint8((rgb >> 4) & 0xF * RGBShortMultiplier)
	b = uint8(rgb & 0xF * RGBShortMultiplier)
	a = DefaultAlpha
	return r, g, b, a, nil
}

func (cp *colorParser) parseFullHex(rgb uint64) (r, g, b, a uint8, err error) {
	r = uint8((rgb >> 16) & 0xFF)
	g = uint8((rgb >> MaxColorBits) & 0xFF)
	b = uint8(rgb & 0xFF)
	a = DefaultAlpha
	return r, g, b, a, nil
}

func (cp *colorParser) parseAlphaHex(rgb uint64) (r, g, b, a uint8, err error) {
	r = uint8((rgb >> 24) & 0xFF)
	g = uint8((rgb >> 16) & 0xFF)
	b = uint8((rgb >> MaxColorBits) & 0xFF)
	a = uint8(rgb & 0xFF)
	return r, g, b, a, nil
}

func (cp *colorParser) parseRGBColor(color string) (r, g, b, a uint8, err error) {
	start := strings.Index(color, "(")
	end := strings.LastIndex(color, ")")
	if start == -1 || end == -1 {
		return 0, 0, 0, 0, fmt.Errorf("invalid rgb format")
	}

	values := strings.Split(color[start+1:end], ",")
	if len(values) < 3 {
		return 0, 0, 0, 0, fmt.Errorf("insufficient rgb values")
	}

	rVal, err := cp.parseColorValue(values[0])
	if err != nil {
		return 0, 0, 0, 0, err
	}

	gVal, err := cp.parseColorValue(values[1])
	if err != nil {
		return 0, 0, 0, 0, err
	}

	bVal, err := cp.parseColorValue(values[2])
	if err != nil {
		return 0, 0, 0, 0, err
	}

	aVal := DefaultAlpha
	if len(values) > 3 {
		if alphaFloat, alphaErr := strconv.ParseFloat(strings.TrimSpace(values[3]), 64); alphaErr == nil {
			aVal = int(alphaFloat * MaxColorValue)
		}
	}

	return uint8(rVal), uint8(gVal), uint8(bVal), uint8(aVal), nil
}

func (cp *colorParser) parseColorValue(value string) (int, error) {
	val, err := strconv.ParseInt(strings.TrimSpace(value), 10, MaxColorBits)
	if err != nil {
		return 0, err
	}
	if val < 0 {
		val = 0
	}
	if val > MaxColorValue {
		val = MaxColorValue
	}
	return int(val), nil
}

func (cp *colorParser) IsValidColor(color string) bool {
	_, _, _, _, err := cp.ParseColor(color)
	return err == nil
}

func (cp *colorParser) NormalizeColor(color string) string {
	r, g, b, a, err := cp.ParseColor(color)
	if err != nil {
		return DefaultTextColor
	}

	if a == 255 {
		return fmt.Sprintf("#%02X%02X%02X", r, g, b)
	}
	return fmt.Sprintf("#%02X%02X%02X%02X", r, g, b, a)
}

type TextProcessor interface {
	IsNameChar(c byte) bool
	IsWhitespace(c byte) bool
	TrimAndNormalize(text string) string
}

type textProcessor struct{}

func NewTextProcessor() TextProcessor {
	return &textProcessor{}
}

func (tp *textProcessor) IsNameChar(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '-' || c == '_'
}

func (tp *textProcessor) IsWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f'
}

func (tp *textProcessor) TrimAndNormalize(text string) string {
	normalized := strings.Fields(text)
	return strings.Join(normalized, " ")
}

type APIHandler interface {
	FetchContent(ctx context.Context, url string) (string, error)
}

type apiHandler struct {
	client         *http.Client
	urlNormalizer  URLNormalizer
	activeRequests map[string]context.CancelFunc
	requestMutex   sync.Mutex
	fetchPool      chan struct{}
}

func NewAPIHandler() APIHandler {
	transport := &http.Transport{
		MaxIdleConns:        MaxConcurrentConnections,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     DefaultTimeout,
		DisableCompression:  false,
		ForceAttemptHTTP2:   true,
	}

	client := &http.Client{
		Timeout:   DefaultTimeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	return &apiHandler{
		client:         client,
		urlNormalizer:  NewURLNormalizer(),
		fetchPool:      make(chan struct{}, MaxConcurrentConnections),
		activeRequests: make(map[string]context.CancelFunc),
	}
}

func (ah *apiHandler) FetchContent(ctx context.Context, url string) (string, error) {
	normalizedURL, err := ah.urlNormalizer.Normalize(url)
	if err != nil {
		return "", fmt.Errorf("%s: %w", ErrInvalidURL, err)
	}
	if err := ah.acquireFetchSlot(ctx); err != nil {
		return "", err
	}
	defer ah.releaseFetchSlot()

	cancelCtx := ah.registerRequest(ctx, normalizedURL)
	defer ah.unregisterRequest(url)

	content, err := ah.performHTTPRequest(cancelCtx, normalizedURL)
	if err != nil {
		return "", err
	}

	return content, nil
}

func (ah *apiHandler) performHTTPRequest(ctx context.Context, urlStr string) (string, error) {
	req, err := ah.createHTTPRequest(ctx, urlStr)
	if err != nil {
		return "", err
	}

	resp, err := ah.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return ah.readResponseContent(resp)
}

func (ah *apiHandler) createHTTPRequest(ctx context.Context, urlStr string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, NewBrowserError(ErrInvalidURL, "failed to create request: "+err.Error())
	}

	ah.setRequestHeaders(req)
	return req, nil
}

func (ah *apiHandler) readResponseContent(resp *http.Response) (string, error) {
	reader := ah.createResponseReader(resp)
	defer ah.closeReader(reader, resp)

	content, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func (ah *apiHandler) createResponseReader(resp *http.Response) io.Reader {
	if !strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		return resp.Body
	}

	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return resp.Body
	}
	return gzReader
}

func (ah *apiHandler) closeReader(reader io.Reader, resp *http.Response) {
	if gzReader, ok := reader.(*gzip.Reader); ok && gzReader != resp.Body {
		gzReader.Close()
	}
}

func (ah *apiHandler) setRequestHeaders(req *http.Request) {
	req.Header.Set("User-Agent", DefaultUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
}

func (ah *apiHandler) acquireFetchSlot(ctx context.Context) error {
	select {
	case ah.fetchPool <- struct{}{}:
		return nil
	case <-ctx.Done():
		return NewBrowserError(ErrNetworkTimeout, "request cancelled")
	}
}

func (ah *apiHandler) releaseFetchSlot() {
	<-ah.fetchPool
}

func (ah *apiHandler) registerRequest(ctx context.Context, urlStr string) context.Context {
	ah.requestMutex.Lock()
	defer ah.requestMutex.Unlock()

	cancelCtx, cancel := context.WithCancel(ctx)
	ah.activeRequests[urlStr] = cancel
	return cancelCtx
}

func (ah *apiHandler) unregisterRequest(urlStr string) {
	ah.requestMutex.Lock()
	defer ah.requestMutex.Unlock()
	delete(ah.activeRequests, urlStr)
}

type URLResolver interface {
	Resolve(baseURL, relativeURL string) (string, error)
}

type urlResolver struct{}

func NewURLResolver() URLResolver {
	return &urlResolver{}
}

func (r *urlResolver) Resolve(baseURL, relativeURL string) (string, error) {
	if relativeURL == "" {
		return "", ErrInvalidURL
	}
	if strings.Contains(relativeURL, "://") {
		return relativeURL, nil
	}
	parsedBase, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(relativeURL, "//") {
		return parsedBase.Scheme + ":" + relativeURL, nil
	}
	if strings.HasPrefix(relativeURL, "/") {
		return parsedBase.Scheme + "://" + parsedBase.Host + relativeURL, nil
	}
	// Handle .. in relative path
	basePath := parsedBase.Path
	if !strings.HasSuffix(basePath, "/") {
		basePath = basePath[:strings.LastIndex(basePath, "/")+1]
	}
	for strings.HasPrefix(relativeURL, "../") {
		relativeURL = relativeURL[3:]
		if basePath != "/" {
			basePath = basePath[:strings.LastIndex(basePath[:len(basePath)-1], "/")+1]
		}
	}
	absPath := basePath + relativeURL
	return parsedBase.Scheme + "://" + parsedBase.Host + absPath, nil
}
