package layout

import (
	"unicode/utf8"

	"github.com/ducnd58233/gobrowser/internal/browser"
)

type LayoutEngineDependencies struct {
	ColorParser browser.ColorParser
	UnitParser  browser.UnitParser
	Cache       *LayoutCache
}

type BoxEdges struct {
	top, right, bottom, left float64
}

type TextMetrics struct {
	CharacterWidth float64
	LineHeight     float64
	FontSize       float64
}

func (tm *TextMetrics) calculateTextDimensions(text string, availableWidth float64) (int, float64) {
	if text == "" {
		return 1, 0
	}

	characterCount := utf8.RuneCountInString(text)
	textWidth := float64(characterCount) * tm.CharacterWidth

	if textWidth <= availableWidth {
		return 1, textWidth
	}

	charactersPerLine := int(availableWidth / tm.CharacterWidth)
	if charactersPerLine < 1 {
		charactersPerLine = 1
	}

	// just divide by characters per line
	lines := (characterCount + charactersPerLine - 1) / charactersPerLine
	return lines, float64(charactersPerLine) * tm.CharacterWidth
}
