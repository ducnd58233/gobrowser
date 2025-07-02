package layout

import (
	"image/color"
	"sync"

	"github.com/ducnd58233/gobrowser/internal/browser"
)

// Cache for parsed values to avoid repeated parsing
type LayoutCache struct {
	colorCache    map[string]color.NRGBA
	unitCache     map[string]float64
	fontSizeCache map[string]float64
	mu            sync.RWMutex
}

func NewLayoutCache() *LayoutCache {
	return &LayoutCache{
		colorCache:    make(map[string]color.NRGBA),
		unitCache:     make(map[string]float64),
		fontSizeCache: make(map[string]float64),
	}
}

func (lc *LayoutCache) GetColor(colorStr string, parser browser.ColorParser) color.NRGBA {
	if colorStr == "" || colorStr == "transparent" {
		return color.NRGBA{R: 0, G: 0, B: 0, A: 0}
	}

	lc.mu.RLock()
	if cached, exists := lc.colorCache[colorStr]; exists {
		lc.mu.RUnlock()
		return cached
	}
	lc.mu.RUnlock()

	r, g, b, a, err := parser.ParseColor(colorStr)
	result := color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	if err == nil {
		result = color.NRGBA{R: r, G: g, B: b, A: a}
	}

	lc.mu.Lock()
	lc.colorCache[colorStr] = result
	lc.mu.Unlock()

	return result
}

func (lc *LayoutCache) GetUnit(value string, parser browser.UnitParser, baseFontSize float64) float64 {
	if value == "" || value == "0" {
		return 0
	}

	lc.mu.RLock()
	if cached, exists := lc.unitCache[value]; exists {
		lc.mu.RUnlock()
		return cached
	}
	lc.mu.RUnlock()

	numValue, unit, ok := parser.ParseUnit(value)
	result := 0.0
	if ok {
		result = parser.ConvertToPixels(numValue, unit, baseFontSize)
	}

	lc.mu.Lock()
	lc.unitCache[value] = result
	lc.mu.Unlock()

	return result
}

func (lc *LayoutCache) GetFontSize(fontSizeStr string, parser browser.UnitParser, baseFontSize float64) float64 {
	if fontSizeStr == "" {
		return baseFontSize
	}

	lc.mu.RLock()
	if cached, exists := lc.fontSizeCache[fontSizeStr]; exists {
		lc.mu.RUnlock()
		return cached
	}
	lc.mu.RUnlock()

	numValue, unit, ok := parser.ParseUnit(fontSizeStr)
	result := baseFontSize
	if ok {
		result = parser.ConvertToPixels(numValue, unit, baseFontSize)
	}

	if result < browser.MinFontSize {
		result = browser.MinFontSize
	}
	if result > browser.MaxFontSize {
		result = browser.MaxFontSize
	}

	lc.mu.Lock()
	lc.fontSizeCache[fontSizeStr] = result
	lc.mu.Unlock()

	return result
}
