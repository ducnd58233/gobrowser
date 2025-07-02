package layout

import (
	"strings"

	"github.com/ducnd58233/gobrowser/internal/browser"
	"github.com/ducnd58233/gobrowser/internal/ui/render"
)

type LayoutEngine interface {
	Layout(document browser.Document, width, height float64) render.DisplayList
	GetScrollHeight() float64
}

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

func (le *layoutEngine) Layout(document browser.Document, width, height float64) render.DisplayList {
	root := le.buildLayoutTree(document.GetRoot(), document)
	displayList := render.NewDisplayList()

	if root == nil {
		return displayList
	}

	totalHeight := root.Layout(width)
	le.scrollHeight = totalHeight

	displayList.SetHeight(totalHeight)

	root.Paint(displayList)

	return displayList
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
			layoutNode.AddChild(childLayout)
		}
	}
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


func (le *layoutEngine) GetScrollHeight() float64 {
	return le.scrollHeight
}
