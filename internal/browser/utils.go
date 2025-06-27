package browser

import (
	"crypto/rand"
	"fmt"
	"net/url"
	"strconv"
	"strings"
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

type textProcessor struct {}

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

