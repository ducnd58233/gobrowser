package ui

import (
	"image"
	"image/color"
	"strings"
	"sync"
	"unicode/utf8"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

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

type DrawCommand interface {
	Execute(gtx layout.Context, theme *material.Theme, scrollY float64)
}

type DisplayList interface {
	Paint(gtx layout.Context, theme *material.Theme, scrollY float64)
	GetHeight() float64
	AddCommand(cmd DrawCommand)
}

type LayoutEngine interface {
	Layout(document browser.Document, width, height float64) DisplayList
	GetScrollHeight() float64
}

type LayoutNode interface {
	Layout(width float64) float64
	Paint(displayList DisplayList)
	GetX() float64
	GetY() float64
	GetWidth() float64
	GetHeight() float64
	SetPosition(x, y float64)
	IsVisible(viewportY, viewportHeight float64) bool
}

// Dependencies for layout engine
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

type DrawText struct {
	X, Y           float64
	Text           string
	FontSize       float64
	Color          color.NRGBA
	cachedTextSize unit.Sp
}

func NewDrawText(x, y float64, text string, fontSize float64, textColor color.NRGBA) *DrawText {
	return &DrawText{
		X:              x,
		Y:              y,
		Text:           text,
		FontSize:       fontSize,
		Color:          textColor,
		cachedTextSize: unit.Sp(float32(fontSize)),
	}
}

func (dt *DrawText) Execute(gtx layout.Context, theme *material.Theme, scrollY float64) {
	adjustedY := dt.Y - scrollY

	if !dt.isVisible(adjustedY, float64(gtx.Constraints.Max.Y)) {
		return
	}

	label := material.Body1(theme, dt.Text)
	label.Color = dt.Color
	label.TextSize = dt.cachedTextSize

	stack := op.Offset(image.Pt(int(dt.X), int(adjustedY))).Push(gtx.Ops)
	defer stack.Pop()
	label.Layout(gtx)
}

func (dt *DrawText) isVisible(adjustedY, maxY float64) bool {
	textHeight := dt.FontSize
	if textHeight <= 0 {
		textHeight = browser.DefaultFontSize
	}
	return adjustedY >= -textHeight*2 && adjustedY <= maxY+textHeight
}

type DrawRect struct {
	X, Y          float64
	Width, Height float64
	Color         color.NRGBA
}

func (dr *DrawRect) Execute(gtx layout.Context, theme *material.Theme, scrollY float64) {
	adjustedY := dr.Y - scrollY

	if !dr.isVisible(adjustedY, float64(gtx.Constraints.Max.Y)) {
		return
	}

	rect := image.Rectangle{
		Min: image.Pt(int(dr.X), int(adjustedY)),
		Max: image.Pt(int(dr.X+dr.Width), int(adjustedY+dr.Height)),
	}

	defer clip.Rect(rect).Push(gtx.Ops).Pop()
	paint.ColorOp{Color: dr.Color}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
}

func (dr *DrawRect) isVisible(adjustedY, maxY float64) bool {
	return adjustedY <= maxY && adjustedY+dr.Height >= 0
}

type displayList struct {
	commands []DrawCommand
	height   float64
}

func NewDisplayList() DisplayList {
	return &displayList{
		commands: make([]DrawCommand, 0),
		height:   0,
	}
}

func (dl *displayList) Paint(gtx layout.Context, theme *material.Theme, scrollY float64) {
	for _, cmd := range dl.commands {
		cmd.Execute(gtx, theme, scrollY)
	}
}

func (dl *displayList) GetHeight() float64 {
	return dl.height
}

func (dl *displayList) AddCommand(cmd DrawCommand) {
	dl.commands = append(dl.commands, cmd)
}

// layoutEngine implementation
type layoutEngine struct {
	deps         LayoutEngineDependencies
	scrollHeight float64
}

func NewLayoutEngine(deps LayoutEngineDependencies) LayoutEngine {
	if deps.Cache == nil {
		deps.Cache = NewLayoutCache()
	}
	return &layoutEngine{
		deps: deps,
	}
}

func (le *layoutEngine) Layout(document browser.Document, width, height float64) DisplayList {
	root := le.buildLayoutTree(document.GetRoot(), document)
	if root == nil {
		return &displayList{
			commands: make([]DrawCommand, 0),
			height:   0,
		}
	}

	totalHeight := root.Layout(width)
	le.scrollHeight = totalHeight

	displayList := &displayList{
		commands: make([]DrawCommand, 0),
		height:   totalHeight,
	}

	root.Paint(displayList)

	return displayList
}

func (le *layoutEngine) GetScrollHeight() float64 {
	return le.scrollHeight
}

func (le *layoutEngine) buildLayoutTree(node browser.Node, document browser.Document) LayoutNode {
	if node == nil {
		return nil
	}

	switch node.GetType() {
	case browser.TextNodeType:
		return le.createTextLayout(node, document)
	case browser.ElementNodeType:
		return le.createElementLayout(node, document)
	}

	return nil
}

func (le *layoutEngine) createTextLayout(node browser.Node, document browser.Document) LayoutNode {
	text := node.GetText()
	if text == "" {
		return nil
	}

	parent := node.GetParent()
	whiteSpace := le.getWhiteSpaceValue(parent, document)
	processedText := le.processWhitespace(text, whiteSpace)

	if processedText == "" && !le.isPreformattedWhitespace(whiteSpace) {
		if parent != nil && le.isSignificantWhitespace(node, parent) {
			processedText = " "
		} else {
			return nil
		}
	}

	style := le.getNodeStyle(node, document)

	if le.isInsidePreTag(node) {
		return NewPreformattedInlineLayout(node, style, processedText, le.deps)
	}

	return NewInlineLayout(node, style, processedText, le.deps)
}

func (le *layoutEngine) createElementLayout(node browser.Node, document browser.Document) LayoutNode {
	if le.isInHead(node) {
		return nil
	}

	style := document.GetComputedStyle(node)
	display := style.GetProperty("display").Raw
	if display == "none" {
		return nil
	}

	var layoutNode LayoutNode
	tag := strings.ToLower(node.GetTag())

	switch tag {
	case "br":
		return le.createLineBreak(node, style)
	case "html":
		layoutNode = NewDocumentLayout(node)
	case "pre":
		layoutNode = NewPreformattedLayout(node, style, le.deps)
	default:
		layoutNode = NewBlockLayout(node, style, le.deps)
	}

	le.addChildrenToLayoutNode(layoutNode, node, document)
	return layoutNode
}

func (le *layoutEngine) createLineBreak(node browser.Node, style browser.Style) LayoutNode {
	return NewInlineLayout(node, style, "\n", le.deps)
}

func (le *layoutEngine) getWhiteSpaceValue(parent browser.Node, document browser.Document) string {
	if parent != nil && parent.GetType() == browser.ElementNodeType {
		parentStyle := document.GetComputedStyle(parent)
		if whiteSpace := parentStyle.GetProperty("white-space").Raw; whiteSpace != "" {
			return whiteSpace
		}
	}
	return "normal"
}

func (le *layoutEngine) getNodeStyle(node browser.Node, document browser.Document) browser.Style {
	if style := document.GetComputedStyle(node); style != nil {
		return style
	}
	return browser.NewStyle()
}

func (le *layoutEngine) isPreformattedWhitespace(whiteSpace string) bool {
	return whiteSpace == "pre" || whiteSpace == "pre-wrap"
}

func (le *layoutEngine) isInHead(node browser.Node) bool {
	current := node.GetParent()
	for current != nil {
		if current.GetTag() == "head" {
			return true
		}
		current = current.GetParent()
	}
	return false
}

func (le *layoutEngine) addChildrenToLayoutNode(layoutNode LayoutNode, node browser.Node, document browser.Document) {
	for _, child := range node.GetChildren() {
		childLayout := le.buildLayoutTree(child, document)
		if childLayout != nil {
			le.addChildToParent(layoutNode, childLayout)
		}
	}
}

func (le *layoutEngine) addChildToParent(parent LayoutNode, child LayoutNode) {
	switch p := parent.(type) {
	case *DocumentLayout:
		p.AddChild(child)
	case *BlockLayout:
		p.AddChild(child)
	case *PreformattedLayout:
		p.AddChild(child)
	}
}

type DocumentLayout struct {
	node     browser.Node
	children []LayoutNode
	x, y     float64
	width    float64
	height   float64
}

func NewDocumentLayout(node browser.Node) *DocumentLayout {
	return &DocumentLayout{
		node:     node,
		children: make([]LayoutNode, 0),
	}
}

func (dl *DocumentLayout) Layout(width float64) float64 {
	dl.width = width
	dl.height = 0

	currentY := 0.0
	for _, child := range dl.children {
		child.SetPosition(0, currentY)
		childHeight := child.Layout(width)
		currentY += childHeight
	}

	dl.height = currentY
	return dl.height
}

func (dl *DocumentLayout) Paint(displayList DisplayList) {
	for _, child := range dl.children {
		child.Paint(displayList)
	}
}

func (dl *DocumentLayout) GetX() float64      { return dl.x }
func (dl *DocumentLayout) GetY() float64      { return dl.y }
func (dl *DocumentLayout) GetWidth() float64  { return dl.width }
func (dl *DocumentLayout) GetHeight() float64 { return dl.height }
func (dl *DocumentLayout) SetPosition(x, y float64) {
	dl.x = x
	dl.y = y
}

func (dl *DocumentLayout) AddChild(child LayoutNode) {
	dl.children = append(dl.children, child)
}

func (dl *DocumentLayout) IsVisible(viewportY, viewportHeight float64) bool {
	return dl.y <= viewportY+viewportHeight && dl.y+dl.height >= viewportY
}

type BlockLayout struct {
	node     browser.Node
	children []LayoutNode
	x, y     float64
	width    float64
	height   float64
	style    browser.Style
	deps     LayoutEngineDependencies
	margin   BoxEdges
	padding  BoxEdges
}

func NewBlockLayout(node browser.Node, style browser.Style, deps LayoutEngineDependencies) *BlockLayout {
	bl := &BlockLayout{
		node:     node,
		children: make([]LayoutNode, 0),
		style:    style,
		deps:     deps,
	}

	bl.calculateBoxModel()
	return bl
}

func (bl *BlockLayout) calculateBoxModel() {
	bl.margin = BoxEdges{
		top:    bl.deps.Cache.GetUnit(bl.style.GetProperty(browser.PropMarginTop).Raw, bl.deps.UnitParser, browser.DefaultFontSize),
		right:  bl.deps.Cache.GetUnit(bl.style.GetProperty(browser.PropMarginRight).Raw, bl.deps.UnitParser, browser.DefaultFontSize),
		bottom: bl.deps.Cache.GetUnit(bl.style.GetProperty(browser.PropMarginBottom).Raw, bl.deps.UnitParser, browser.DefaultFontSize),
		left:   bl.deps.Cache.GetUnit(bl.style.GetProperty(browser.PropMarginLeft).Raw, bl.deps.UnitParser, browser.DefaultFontSize),
	}

	bl.padding = BoxEdges{
		top:    bl.deps.Cache.GetUnit(bl.style.GetProperty(browser.PropPaddingTop).Raw, bl.deps.UnitParser, browser.DefaultFontSize),
		right:  bl.deps.Cache.GetUnit(bl.style.GetProperty(browser.PropPaddingRight).Raw, bl.deps.UnitParser, browser.DefaultFontSize),
		bottom: bl.deps.Cache.GetUnit(bl.style.GetProperty(browser.PropPaddingBottom).Raw, bl.deps.UnitParser, browser.DefaultFontSize),
		left:   bl.deps.Cache.GetUnit(bl.style.GetProperty(browser.PropPaddingLeft).Raw, bl.deps.UnitParser, browser.DefaultFontSize),
	}
}

func (bl *BlockLayout) Layout(width float64) float64 {
	bl.width = width - bl.margin.left - bl.margin.right

	contentX := bl.x + bl.margin.left + bl.padding.left
	contentY := bl.y + bl.margin.top + bl.padding.top
	contentWidth := bl.width - bl.padding.left - bl.padding.right

	currentY := contentY

	for _, child := range bl.children {
		child.SetPosition(contentX, currentY)
		childHeight := child.Layout(contentWidth)
		currentY += childHeight
	}

	bl.height = (currentY - bl.y) + bl.padding.bottom + bl.margin.bottom
	return bl.height
}

func (bl *BlockLayout) Paint(displayList DisplayList) {
	bl.paintBackground(displayList)

	for _, child := range bl.children {
		child.Paint(displayList)
	}
}

func (bl *BlockLayout) paintBackground(displayList DisplayList) {
	bgColor := bl.style.GetProperty(browser.PropBackgroundColor)
	if bgColor.Raw == "transparent" || bgColor.Raw == "" {
		return
	}

	if color := bl.deps.Cache.GetColor(bgColor.Raw, bl.deps.ColorParser); color.A > 0 {
		displayList.AddCommand(&DrawRect{
			X:      bl.x + bl.margin.left,
			Y:      bl.y + bl.margin.top,
			Width:  bl.width,
			Height: bl.height - bl.margin.top - bl.margin.bottom,
			Color:  color,
		})
	}
}

func (bl *BlockLayout) GetX() float64      { return bl.x }
func (bl *BlockLayout) GetY() float64      { return bl.y }
func (bl *BlockLayout) GetWidth() float64  { return bl.width }
func (bl *BlockLayout) GetHeight() float64 { return bl.height }
func (bl *BlockLayout) SetPosition(x, y float64) {
	bl.x = x
	bl.y = y
}

func (bl *BlockLayout) AddChild(child LayoutNode) {
	bl.children = append(bl.children, child)
}

func (bl *BlockLayout) IsVisible(viewportY, viewportHeight float64) bool {
	return bl.y <= viewportY+viewportHeight && bl.y+bl.height >= viewportY
}

type InlineLayout struct {
	node      browser.Node
	text      string
	x, y      float64
	width     float64
	height    float64
	style     browser.Style
	metrics   TextMetrics
	deps      LayoutEngineDependencies
	clickable *widget.Clickable
	hovered   bool
}

func NewInlineLayout(node browser.Node, style browser.Style, text string, deps LayoutEngineDependencies) *InlineLayout {
	il := &InlineLayout{
		node:  node,
		text:  text,
		style: style,
		deps:  deps,
	}
	if node.GetTag() == "a" {
		il.clickable = &widget.Clickable{}
	}

	il.calculateTextMetrics()
	return il
}

func (il *InlineLayout) calculateTextMetrics() {
	fontSize := il.deps.Cache.GetFontSize(il.style.GetProperty(browser.PropFontSize).Raw, il.deps.UnitParser, browser.DefaultFontSize)

	il.metrics = TextMetrics{
		FontSize:       fontSize,
		CharacterWidth: fontSize * browser.CharWidthRatio,
		LineHeight:     fontSize * browser.LineHeightRatio,
	}
}

func (il *InlineLayout) Layout(width float64) float64 {
	il.width = width

	if il.text == "" {
		il.height = 0
		return 0
	}

	// Check if this is preformatted text (no wrapping)
	if il.isPreformattedText() {
		lines := strings.Split(il.text, "\n")
		il.height = float64(len(lines)) * il.metrics.LineHeight
		return il.height
	}

	lines, _ := il.metrics.calculateTextDimensions(il.text, width)
	il.height = float64(lines) * il.metrics.LineHeight
	return il.height
}

func (il *InlineLayout) isPreformattedText() bool {
	whiteSpace := il.style.GetProperty("white-space").Raw
	return whiteSpace == "pre" || whiteSpace == "pre-wrap"
}

func (il *InlineLayout) Paint(displayList DisplayList) {
	if il.text == "" {
		return
	}

	isLink := il.node.GetTag() == "a"
	textColor := il.deps.Cache.GetColor(il.style.GetProperty(browser.PropColor).Raw, il.deps.ColorParser)
	if isLink {
		if il.hovered {
			textColor = color.NRGBA{R: 0, G: 0, B: 180, A: 255}
		} else {
			textColor = color.NRGBA{R: 0, G: 0, B: 238, A: 255}
		}
	}

	if il.isPreformattedText() {
		il.paintPreformattedText(displayList, textColor)
	} else {
		il.paintWrappedText(displayList, textColor)
	}
}

func (il *InlineLayout) paintPreformattedText(displayList DisplayList, textColor color.NRGBA) {
	if il.text == "" {
		return
	}

	lines := strings.Split(il.text, "\n")
	currentY := il.y

	for _, line := range lines {
		if line != "" {
			drawCmd := NewDrawText(il.x, currentY, line, il.metrics.FontSize, textColor)
			displayList.AddCommand(drawCmd)
		}
		currentY += il.metrics.LineHeight
	}
}

func (il *InlineLayout) paintWrappedText(displayList DisplayList, textColor color.NRGBA) {
	if il.text == "" {
		return
	}

	lines := il.splitTextIntoLines()
	currentY := il.y

	for _, line := range lines {
		if line == "" {
			currentY += il.metrics.LineHeight
			continue
		}

		drawCmd := NewDrawText(il.x, currentY, line, il.metrics.FontSize, textColor)
		displayList.AddCommand(drawCmd)

		currentY += il.metrics.LineHeight
	}
}

func (il *InlineLayout) splitTextIntoLines() []string {
	lines := strings.Split(il.text, "\n")
	var wrappedLines []string

	charactersPerLine := il.calculateCharactersPerLine()

	for _, line := range lines {
		if line == "" {
			wrappedLines = append(wrappedLines, "")
			continue
		}
		wrappedLines = append(wrappedLines, il.wrapLineToWidth(line, charactersPerLine)...)
	}

	return wrappedLines
}

func (il *InlineLayout) calculateCharactersPerLine() int {
	charactersPerLine := int(il.width / il.metrics.CharacterWidth)
	if charactersPerLine < 1 {
		charactersPerLine = 1
	}
	return charactersPerLine
}

func (il *InlineLayout) wrapLineToWidth(line string, charactersPerLine int) []string {
	words := strings.Fields(line)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	currentLine := ""

	for _, word := range words {
		if il.canAddWordToLine(currentLine, word, charactersPerLine) {
			currentLine = il.addWordToLine(currentLine, word)
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

func (il *InlineLayout) canAddWordToLine(currentLine, word string, charactersPerLine int) bool {
	testLine := il.addWordToLine(currentLine, word)
	return utf8.RuneCountInString(testLine) <= charactersPerLine
}

func (il *InlineLayout) addWordToLine(currentLine, word string) string {
	if currentLine == "" {
		return word
	}
	return currentLine + " " + word
}

func (il *InlineLayout) GetX() float64      { return il.x }
func (il *InlineLayout) GetY() float64      { return il.y }
func (il *InlineLayout) GetWidth() float64  { return il.width }
func (il *InlineLayout) GetHeight() float64 { return il.height }
func (il *InlineLayout) SetPosition(x, y float64) {
	il.x = x
	il.y = y
}

func (il *InlineLayout) IsVisible(viewportY, viewportHeight float64) bool {
	return il.y <= viewportY+viewportHeight && il.y+il.height >= viewportY
}

func (le *layoutEngine) isSignificantWhitespace(textNode browser.Node, parent browser.Node) bool {
	siblings := parent.GetChildren()
	textNodeIndex := le.findNodeIndex(siblings, textNode)

	if textNodeIndex == -1 {
		return false
	}

	return le.hasElementBefore(siblings, textNodeIndex) && le.hasElementAfter(siblings, textNodeIndex)
}

func (le *layoutEngine) findNodeIndex(siblings []browser.Node, target browser.Node) int {
	for i, child := range siblings {
		if child == target {
			return i
		}
	}
	return -1
}

func (le *layoutEngine) hasElementBefore(siblings []browser.Node, index int) bool {
	for i := index - 1; i >= 0; i-- {
		if siblings[i].GetType() == browser.ElementNodeType {
			return true
		}
	}
	return false
}

func (le *layoutEngine) hasElementAfter(siblings []browser.Node, index int) bool {
	for i := index + 1; i < len(siblings); i++ {
		if siblings[i].GetType() == browser.ElementNodeType {
			return true
		}
	}
	return false
}

func (le *layoutEngine) processWhitespace(text, whiteSpace string) string {
	switch whiteSpace {
	case "pre", "pre-wrap":
		return text
	case "pre-line":
		return le.normalizeWhitespace(text, true)
	case "nowrap":
		return le.normalizeWhitespace(text, false)
	default:
		return le.normalizeWhitespace(text, false)
	}
}

func (le *layoutEngine) normalizeWhitespace(text string, preserveNewlines bool) string {
	if text == "" {
		return ""
	}

	var result strings.Builder
	var lastWasSpace bool
	var lastWasNewline bool

	for _, r := range text {
		switch r {
		case '\n', '\r':
			if preserveNewlines {
				if !lastWasNewline {
					result.WriteRune('\n')
					lastWasNewline = true
					lastWasSpace = false
				}
			} else {
				if !lastWasSpace {
					result.WriteRune(' ')
					lastWasSpace = true
					lastWasNewline = false
				}
			}
		case ' ', '\t':
			if !lastWasSpace && !lastWasNewline {
				result.WriteRune(' ')
				lastWasSpace = true
				lastWasNewline = false
			}
		default:
			result.WriteRune(r)
			lastWasSpace = false
			lastWasNewline = false
		}
	}

	if !preserveNewlines {
		return strings.TrimSpace(result.String())
	}

	return result.String()
}

type PreformattedLayout struct {
	node     browser.Node
	children []LayoutNode
	x, y     float64
	width    float64
	height   float64
	style    browser.Style
	deps     LayoutEngineDependencies
	margin   BoxEdges
	padding  BoxEdges
}

func NewPreformattedLayout(node browser.Node, style browser.Style, deps LayoutEngineDependencies) *PreformattedLayout {
	pl := &PreformattedLayout{
		node:     node,
		children: make([]LayoutNode, 0),
		style:    style,
		deps:     deps,
	}

	pl.calculateBoxModel()
	return pl
}

func (pl *PreformattedLayout) calculateBoxModel() {
	pl.margin = BoxEdges{
		top:    pl.deps.Cache.GetUnit(pl.style.GetProperty(browser.PropMarginTop).Raw, pl.deps.UnitParser, browser.DefaultFontSize),
		right:  pl.deps.Cache.GetUnit(pl.style.GetProperty(browser.PropMarginRight).Raw, pl.deps.UnitParser, browser.DefaultFontSize),
		bottom: pl.deps.Cache.GetUnit(pl.style.GetProperty(browser.PropMarginBottom).Raw, pl.deps.UnitParser, browser.DefaultFontSize),
		left:   pl.deps.Cache.GetUnit(pl.style.GetProperty(browser.PropMarginLeft).Raw, pl.deps.UnitParser, browser.DefaultFontSize),
	}

	pl.padding = BoxEdges{
		top:    pl.deps.Cache.GetUnit(pl.style.GetProperty(browser.PropPaddingTop).Raw, pl.deps.UnitParser, browser.DefaultFontSize),
		right:  pl.deps.Cache.GetUnit(pl.style.GetProperty(browser.PropPaddingRight).Raw, pl.deps.UnitParser, browser.DefaultFontSize),
		bottom: pl.deps.Cache.GetUnit(pl.style.GetProperty(browser.PropPaddingBottom).Raw, pl.deps.UnitParser, browser.DefaultFontSize),
		left:   pl.deps.Cache.GetUnit(pl.style.GetProperty(browser.PropPaddingLeft).Raw, pl.deps.UnitParser, browser.DefaultFontSize),
	}
}

func (pl *PreformattedLayout) Layout(width float64) float64 {
	pl.width = width - pl.margin.left - pl.margin.right

	contentX := pl.x + pl.margin.left + pl.padding.left
	contentY := pl.y + pl.margin.top + pl.padding.top
	contentWidth := pl.width - pl.padding.left - pl.padding.right

	currentY := contentY

	for _, child := range pl.children {
		child.SetPosition(contentX, currentY)
		childHeight := child.Layout(contentWidth)
		currentY += childHeight
	}

	pl.height = (currentY - pl.y) + pl.padding.bottom + pl.margin.bottom - (contentY - pl.y)
	return pl.height
}

func (pl *PreformattedLayout) Paint(displayList DisplayList) {
	pl.paintBackground(displayList)

	for _, child := range pl.children {
		child.Paint(displayList)
	}
}

func (pl *PreformattedLayout) paintBackground(displayList DisplayList) {
	bgColor := pl.style.GetProperty(browser.PropBackgroundColor)
	if bgColor.Raw == "transparent" || bgColor.Raw == "" {
		return
	}

	if color := pl.deps.Cache.GetColor(bgColor.Raw, pl.deps.ColorParser); color.A > 0 {
		displayList.AddCommand(&DrawRect{
			X:      pl.x + pl.margin.left,
			Y:      pl.y + pl.margin.top,
			Width:  pl.width,
			Height: pl.height - pl.margin.top - pl.margin.bottom,
			Color:  color,
		})
	}
}

func (pl *PreformattedLayout) GetX() float64      { return pl.x }
func (pl *PreformattedLayout) GetY() float64      { return pl.y }
func (pl *PreformattedLayout) GetWidth() float64  { return pl.width }
func (pl *PreformattedLayout) GetHeight() float64 { return pl.height }
func (pl *PreformattedLayout) SetPosition(x, y float64) {
	pl.x = x
	pl.y = y
}

func (pl *PreformattedLayout) AddChild(child LayoutNode) {
	pl.children = append(pl.children, child)
}

func (pl *PreformattedLayout) IsVisible(viewportY, viewportHeight float64) bool {
	return pl.y <= viewportY+viewportHeight && pl.y+pl.height >= viewportY
}

func (le *layoutEngine) isInsidePreTag(node browser.Node) bool {
	current := node.GetParent()
	for current != nil {
		if current.GetType() == browser.ElementNodeType && current.GetTag() == "pre" {
			return true
		}
		current = current.GetParent()
	}
	return false
}

type PreformattedInlineLayout struct {
	node    browser.Node
	text    string
	x, y    float64
	width   float64
	height  float64
	style   browser.Style
	metrics TextMetrics
	deps    LayoutEngineDependencies
}

func NewPreformattedInlineLayout(node browser.Node, style browser.Style, text string, deps LayoutEngineDependencies) *PreformattedInlineLayout {
	pil := &PreformattedInlineLayout{
		node:  node,
		text:  text,
		style: style,
		deps:  deps,
	}

	pil.calculateTextMetrics()
	return pil
}

func (pil *PreformattedInlineLayout) calculateTextMetrics() {
	fontSize := pil.deps.Cache.GetFontSize(pil.style.GetProperty(browser.PropFontSize).Raw, pil.deps.UnitParser, browser.DefaultFontSize)

	pil.metrics = TextMetrics{
		FontSize:       fontSize,
		CharacterWidth: fontSize * browser.CharWidthRatio,
		LineHeight:     fontSize * browser.LineHeightRatio,
	}
}

func (pil *PreformattedInlineLayout) Layout(width float64) float64 {
	pil.width = width

	if pil.text == "" {
		pil.height = 0
		return 0
	}

	// For preformatted text, split by newlines and calculate height
	lines := strings.Split(pil.text, "\n")
	pil.height = float64(len(lines)) * pil.metrics.LineHeight
	return pil.height
}

func (pil *PreformattedInlineLayout) Paint(displayList DisplayList) {
	if pil.text == "" {
		return
	}

	textColor := pil.deps.Cache.GetColor(pil.style.GetProperty(browser.PropColor).Raw, pil.deps.ColorParser)
	if textColor.A == 0 {
		textColor = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	}

	// Paint each line separately to preserve formatting
	lines := strings.Split(pil.text, "\n")
	currentY := pil.y

	for _, line := range lines {
		if line != "" {
			drawCmd := NewDrawText(pil.x, currentY, line, pil.metrics.FontSize, textColor)
			displayList.AddCommand(drawCmd)
		}
		currentY += pil.metrics.LineHeight
	}
}

func (pil *PreformattedInlineLayout) GetX() float64      { return pil.x }
func (pil *PreformattedInlineLayout) GetY() float64      { return pil.y }
func (pil *PreformattedInlineLayout) GetWidth() float64  { return pil.width }
func (pil *PreformattedInlineLayout) GetHeight() float64 { return pil.height }
func (pil *PreformattedInlineLayout) SetPosition(x, y float64) {
	pil.x = x
	pil.y = y
}

func (pil *PreformattedInlineLayout) IsVisible(viewportY, viewportHeight float64) bool {
	return pil.y <= viewportY+viewportHeight && pil.y+pil.height >= viewportY
}
