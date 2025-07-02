package layout

import (
	"image/color"
	"strings"
	"unicode/utf8"

	"gioui.org/widget"
	"github.com/ducnd58233/gobrowser/internal/browser"
	"github.com/ducnd58233/gobrowser/internal/ui/render"
	"github.com/ducnd58233/gobrowser/internal/ui/types"
)

type NodeType int

const (
	NodeTypeDocument NodeType = iota
	NodeTypeBlock
	NodeTypeInline
	NodeTypePreformatted
	NodeTypePreformattedInline
)

type LayoutNode interface {
	GetType() NodeType
	Layout(width float64) float64
	Paint(displayList render.DisplayList)
	GetBounds() types.Bounds
	SetPosition(x, y float64)
	IsVisible(viewportY, viewportHeight float64) bool
	GetNode() browser.Node
	AddChild(child LayoutNode)
}

type documentLayout struct {
	node     browser.Node
	children []LayoutNode
	bounds   types.Bounds
}

func NewDocumentLayout(node browser.Node) LayoutNode {
	return &documentLayout{
		node:     node,
		children: make([]LayoutNode, 0),
		bounds: types.Bounds{
			X:      0,
			Y:      0,
			Width:  0,
			Height: 0,
		},
	}
}

func (dl *documentLayout) GetType() NodeType {
	return NodeTypeDocument
}

func (dl *documentLayout) Layout(width float64) float64 {
	dl.bounds.Width = width

	currY := 0.0
	for _, child := range dl.children {
		child.SetPosition(0, currY)
		childHeight := child.Layout(width)
		currY += childHeight
	}

	dl.bounds.Height = currY
	return currY
}

func (dl *documentLayout) Paint(displayList render.DisplayList) {
	for _, child := range dl.children {
		child.Paint(displayList)
	}
}

func (dl *documentLayout) GetBounds() types.Bounds {
	return dl.bounds
}

func (dl *documentLayout) SetPosition(x, y float64) {
	dl.bounds.X = x
	dl.bounds.Y = y
}

func (dl *documentLayout) AddChild(child LayoutNode) {
	dl.children = append(dl.children, child)
}

func (dl *documentLayout) IsVisible(viewportY, viewportHeight float64) bool {
	viewportBounds := types.Bounds{
		X:      0,
		Y:      viewportY,
		Width:  dl.bounds.Width,
		Height: viewportHeight,
	}

	return dl.bounds.Intersects(&viewportBounds)
}

func (dl *documentLayout) GetNode() browser.Node {
	return dl.node
}

type blockLayout struct {
	node     browser.Node
	children []LayoutNode
	bounds   types.Bounds
	style    browser.Style
	margin   BoxEdges
	padding  BoxEdges
	deps     LayoutEngineDependencies
}

func NewBlockLayout(node browser.Node, style browser.Style, deps LayoutEngineDependencies) LayoutNode {
	bl := &blockLayout{
		node:  node,
		style: style,
		deps:  deps,
		bounds: types.Bounds{
			X:      0,
			Y:      0,
			Width:  0,
			Height: 0,
		},
	}
	bl.calculateBoxModel()
	return bl
}

func (bl *blockLayout) GetType() NodeType {
	return NodeTypeBlock
}

func (bl *blockLayout) calculateBoxModel() {
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

func (bl *blockLayout) Layout(width float64) float64 {
	bl.bounds.Width = width - bl.margin.left - bl.margin.right

	contentX := bl.bounds.X + bl.margin.left + bl.padding.left
	contentY := bl.bounds.Y + bl.margin.top + bl.padding.top
	contentWidth := bl.bounds.Width - bl.padding.left - bl.padding.right

	currentY := contentY

	for _, child := range bl.children {
		child.SetPosition(contentX, currentY)
		childHeight := child.Layout(contentWidth)
		currentY += childHeight
	}

	bl.bounds.Height = (currentY - bl.bounds.Y) + bl.padding.bottom + bl.margin.bottom
	return bl.bounds.Height
}

func (bl *blockLayout) Paint(displayList render.DisplayList) {
	bl.paintBackground(displayList)

	for _, child := range bl.children {
		child.Paint(displayList)
	}
}

func (bl *blockLayout) paintBackground(displayList render.DisplayList) {
	bgColor := bl.style.GetProperty(browser.PropBackgroundColor)
	if bgColor.Raw == "transparent" || bgColor.Raw == "" {
		return
	}

	if color := bl.deps.Cache.GetColor(bgColor.Raw, bl.deps.ColorParser); color.A > 0 {
		displayList.AddCommand(render.NewDrawRect(
			types.Bounds{
				X:      bl.bounds.X + bl.margin.left,
				Y:      bl.bounds.Y + bl.margin.top,
				Width:  bl.bounds.Width,
				Height: bl.bounds.Height - bl.margin.top - bl.margin.bottom,
			},
			color,
			bl.node,
		))
	}
}

func (bl *blockLayout) GetBounds() types.Bounds {
	return bl.bounds
}

func (bl *blockLayout) SetPosition(x, y float64) {
	bl.bounds.X = x
	bl.bounds.Y = y
}

func (bl *blockLayout) AddChild(child LayoutNode) {
	bl.children = append(bl.children, child)
}

func (bl *blockLayout) GetNode() browser.Node {
	return bl.node
}

func (bl *blockLayout) IsVisible(viewportY, viewportHeight float64) bool {
	viewportBounds := types.Bounds{
		X:      0,
		Y:      viewportY,
		Width:  bl.bounds.Width,
		Height: viewportHeight,
	}

	return bl.bounds.Intersects(&viewportBounds)
}

type inlineLayout struct {
	node      browser.Node
	text      string
	bounds    types.Bounds
	style     browser.Style
	metrics   TextMetrics
	deps      LayoutEngineDependencies
	clickable *widget.Clickable
	hovered   bool
}

func NewInlineLayout(
	node browser.Node,
	style browser.Style,
	text string,
	deps LayoutEngineDependencies,
) LayoutNode {
	il := &inlineLayout{
		node:  node,
		text:  text,
		style: style,
		deps:  deps,
		bounds: types.Bounds{
			X:      0,
			Y:      0,
			Width:  0,
			Height: 0,
		},
	}

	if node.GetTag() == "a" {
		il.clickable = &widget.Clickable{}
	}

	il.calculateTextMetrics()
	return il
}

func (il *inlineLayout) GetType() NodeType {
	return NodeTypeInline
}

func (il *inlineLayout) calculateTextMetrics() {
	fontSize := il.deps.Cache.GetFontSize(il.style.GetProperty(browser.PropFontSize).Raw, il.deps.UnitParser, browser.DefaultFontSize)

	il.metrics = TextMetrics{
		FontSize:       fontSize,
		CharacterWidth: fontSize * browser.CharWidthRatio,
		LineHeight:     fontSize * browser.LineHeightRatio,
	}
}

func (il *inlineLayout) Layout(width float64) float64 {
	il.bounds.Width = width

	if il.text == "" {
		il.bounds.Height = 0
		return 0
	}

	if il.isPreformattedText() {
		lines := strings.Split(il.text, "\n")
		il.bounds.Height = float64(len(lines)) * il.metrics.LineHeight
		return il.bounds.Height
	}

	lines, _ := il.metrics.calculateTextDimensions(il.text, width)
	il.bounds.Height = float64(lines) * il.metrics.LineHeight
	return il.bounds.Height
}

func (il *inlineLayout) isPreformattedText() bool {
	whiteSpace := il.style.GetProperty("white-space").Raw
	return whiteSpace == "pre" || whiteSpace == "pre-wrap" || whiteSpace == "pre-line"
}

func (il *inlineLayout) Paint(displayList render.DisplayList) {
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

func (il *inlineLayout) paintPreformattedText(displayList render.DisplayList, textColor color.NRGBA) {
	if il.text == "" {
		return
	}

	lines := strings.Split(il.text, "\n")
	currentY := il.bounds.Y

	for _, line := range lines {
		if line != "" {
			drawCmd := render.NewDrawText(il.bounds.X, currentY, line, il.metrics.FontSize, textColor, il.node)
			displayList.AddCommand(drawCmd)
		}
		currentY += il.metrics.LineHeight
	}
}

func (il *inlineLayout) paintWrappedText(displayList render.DisplayList, textColor color.NRGBA) {
	if il.text == "" {
		return
	}

	lines := il.splitTextIntoLines()
	currentY := il.bounds.Y

	for _, line := range lines {
		if line != "" {
			drawCmd := render.NewDrawText(il.bounds.X, currentY, line, il.metrics.FontSize, textColor, il.node)
			displayList.AddCommand(drawCmd)
		}
		currentY += il.metrics.LineHeight
	}
}

func (il *inlineLayout) splitTextIntoLines() []string {
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

func (il *inlineLayout) calculateCharactersPerLine() int {
	charactersPerLine := int(il.bounds.Width / il.metrics.CharacterWidth)
	if charactersPerLine < 1 {
		charactersPerLine = 1
	}
	return charactersPerLine
}

func (il *inlineLayout) wrapLineToWidth(line string, charactersPerLine int) []string {
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

func (il *inlineLayout) canAddWordToLine(currentLine, word string, charactersPerLine int) bool {
	testLine := il.addWordToLine(currentLine, word)
	return utf8.RuneCountInString(testLine) <= charactersPerLine
}

func (il *inlineLayout) addWordToLine(currentLine, word string) string {
	if currentLine == "" {
		return word
	}
	return currentLine + " " + word
}

func (il *inlineLayout) GetBounds() types.Bounds {
	return il.bounds
}

func (il *inlineLayout) SetPosition(x, y float64) {
	il.bounds.X = x
	il.bounds.Y = y
}

func (il *inlineLayout) AddChild(child LayoutNode) {
	// Inline layout doesn't have children
}

func (il *inlineLayout) GetNode() browser.Node {
	return il.node
}

func (il *inlineLayout) IsVisible(viewportY, viewportHeight float64) bool {
	viewportBounds := types.Bounds{
		X:      0,
		Y:      viewportY,
		Width:  il.bounds.Width,
		Height: viewportHeight,
	}

	return il.bounds.Intersects(&viewportBounds)
}

type preformattedLayout struct {
	node     browser.Node
	children []LayoutNode
	bounds   types.Bounds
	style    browser.Style
	deps     LayoutEngineDependencies
	margin   BoxEdges
	padding  BoxEdges
}

func NewPreformattedLayout(node browser.Node, style browser.Style, deps LayoutEngineDependencies) LayoutNode {
	pl := &preformattedLayout{
		node:     node,
		children: make([]LayoutNode, 0),
		style:    style,
		deps:     deps,
		bounds: types.Bounds{
			X:      0,
			Y:      0,
			Width:  0,
			Height: 0,
		},
	}

	pl.calculateBoxModel()
	return pl
}

func (pl *preformattedLayout) calculateBoxModel() {
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

func (pl *preformattedLayout) GetType() NodeType {
	return NodeTypePreformatted
}

func (pl *preformattedLayout) Layout(width float64) float64 {
	pl.bounds.Width = width - pl.margin.left - pl.margin.right

	contentX := pl.bounds.X + pl.margin.left + pl.padding.left
	contentY := pl.bounds.Y + pl.margin.top + pl.padding.top
	contentWidth := pl.bounds.Width - pl.padding.left - pl.padding.right

	currentY := contentY

	for _, child := range pl.children {
		child.SetPosition(contentX, currentY)
		childHeight := child.Layout(contentWidth)
		currentY += childHeight
	}

	pl.bounds.Height = (currentY - pl.bounds.Y) + pl.padding.bottom + pl.margin.bottom - (contentY - pl.bounds.Y)
	return pl.bounds.Height
}

func (pl *preformattedLayout) Paint(displayList render.DisplayList) {
	pl.paintBackground(displayList)

	for _, child := range pl.children {
		child.Paint(displayList)
	}
}

func (pl *preformattedLayout) paintBackground(displayList render.DisplayList) {
	bgColor := pl.style.GetProperty(browser.PropBackgroundColor)
	if bgColor.Raw == "transparent" || bgColor.Raw == "" {
		return
	}

	if color := pl.deps.Cache.GetColor(bgColor.Raw, pl.deps.ColorParser); color.A > 0 {
		displayList.AddCommand(render.NewDrawRect(
			types.Bounds{
				X:      pl.bounds.X + pl.margin.left,
				Y:      pl.bounds.Y + pl.margin.top,
				Width:  pl.bounds.Width,
				Height: pl.bounds.Height - pl.margin.top - pl.margin.bottom,
			},
			color,
			pl.node,
		))
	}
}

func (pl *preformattedLayout) GetBounds() types.Bounds {
	return pl.bounds
}

func (pl *preformattedLayout) SetPosition(x, y float64) {
	pl.bounds.X = x
	pl.bounds.Y = y
}

func (pl *preformattedLayout) AddChild(child LayoutNode) {
	pl.children = append(pl.children, child)
}

func (pl *preformattedLayout) GetNode() browser.Node {
	return pl.node
}

func (pl *preformattedLayout) IsVisible(viewportY, viewportHeight float64) bool {
	viewportBounds := types.Bounds{
		X:      0,
		Y:      viewportY,
		Width:  pl.bounds.Width,
		Height: viewportHeight,
	}

	return pl.bounds.Intersects(&viewportBounds)
}

type preformattedInlineLayout struct {
	node    browser.Node
	text    string
	bounds  types.Bounds
	style   browser.Style
	metrics TextMetrics
	deps    LayoutEngineDependencies
}

func NewPreformattedInlineLayout(node browser.Node, style browser.Style, text string, deps LayoutEngineDependencies) LayoutNode {
	pil := &preformattedInlineLayout{
		node:  node,
		text:  text,
		style: style,
		deps:  deps,
	}

	pil.calculateTextMetrics()
	return pil
}

func (pil *preformattedInlineLayout) calculateTextMetrics() {
	fontSize := pil.deps.Cache.GetFontSize(pil.style.GetProperty(browser.PropFontSize).Raw, pil.deps.UnitParser, browser.DefaultFontSize)

	pil.metrics = TextMetrics{
		FontSize:       fontSize,
		CharacterWidth: fontSize * browser.CharWidthRatio,
		LineHeight:     fontSize * browser.LineHeightRatio,
	}
}

func (pil *preformattedInlineLayout) GetType() NodeType {
	return NodeTypePreformattedInline
}

func (pil *preformattedInlineLayout) Layout(width float64) float64 {
	pil.bounds.Width = width

	if pil.text == "" {
		pil.bounds.Height = 0
		return 0
	}

	// For preformatted text, split by newlines and calculate height
	lines := strings.Split(pil.text, "\n")
	pil.bounds.Height = float64(len(lines)) * pil.metrics.LineHeight
	return pil.bounds.Height
}

func (pil *preformattedInlineLayout) Paint(displayList render.DisplayList) {
	if pil.text == "" {
		return
	}

	textColor := pil.deps.Cache.GetColor(pil.style.GetProperty(browser.PropColor).Raw, pil.deps.ColorParser)
	if textColor.A == 0 {
		textColor = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	}

	// Paint each line separately to preserve formatting
	lines := strings.Split(pil.text, "\n")
	currentY := pil.bounds.Y

	for _, line := range lines {
		if line != "" {
			drawCmd := render.NewDrawText(pil.bounds.X, currentY, line, pil.metrics.FontSize, textColor, pil.node)
			displayList.AddCommand(drawCmd)
		}
		currentY += pil.metrics.LineHeight
	}
}

func (pil *preformattedInlineLayout) GetBounds() types.Bounds {
	return pil.bounds
}

func (pil *preformattedInlineLayout) SetPosition(x, y float64) {
	pil.bounds.X = x
	pil.bounds.Y = y
}

func (pil *preformattedInlineLayout) AddChild(child LayoutNode) {
	// Inline layout doesn't have children
}

func (pil *preformattedInlineLayout) GetNode() browser.Node {
	return pil.node
}

func (pil *preformattedInlineLayout) IsVisible(viewportY, viewportHeight float64) bool {
	viewportBounds := types.Bounds{
		X:      0,
		Y:      viewportY,
		Width:  pil.bounds.Width,
		Height: viewportHeight,
	}

	return pil.bounds.Intersects(&viewportBounds)
}
