package browser

import (
	"strconv"
	"strings"
)

type ComputedStyle struct {
	Properties map[string]string
}

type CSSApplicator interface {
	ApplyStyles(document Document, css *CSS) error
	ComputeStyle(node Node, css *CSS) *ComputedStyle
}

type cssApplicator struct {
	colorParser     ColorParser
	unitParser      UnitParser
	inheritedProps  map[string]bool
	defaultValues   map[string]string
	selectorMatcher SelectorMatcher
}

func NewCSSApplicator() CSSApplicator {
	return &cssApplicator{
		colorParser:     NewColorParser(),
		unitParser:      NewUnitParser(),
		inheritedProps:  createInheritedPropertiesMap(),
		defaultValues:   createDefaultValuesMap(),
		selectorMatcher: NewSelectorMatcher(),
	}
}

func createInheritedPropertiesMap() map[string]bool {
	return map[string]bool{
		"color": true, "font-family": true, "font-size": true, "font-style": true, "font-weight": true,
		"line-height": true, "text-align": true, "text-decoration": true, "white-space": true,
	}
}

func createDefaultValuesMap() map[string]string {
	return map[string]string{
		"display": "inline", "color": "#000000", "background-color": "transparent",
		"font-size": "16px", "font-weight": "normal", "font-style": "normal",
		"white-space": "normal", "line-height": "normal", "text-align": "left",
		"text-decoration": "none",
		"margin-top":      "0px", "margin-bottom": "0px", "margin-left": "0px", "margin-right": "0px",
		"padding-top": "0px", "padding-bottom": "0px", "padding-left": "0px", "padding-right": "0px",
	}
}

func (ca *cssApplicator) ApplyStyles(document Document, css *CSS) error {
	if css == nil {
		return nil
	}

	ca.applyStylesToNode(document.GetRoot(), css)
	return nil
}

func (ca *cssApplicator) ComputeStyle(node Node, css *CSS) *ComputedStyle {
	style := &ComputedStyle{
		Properties: make(map[string]string),
	}

	ca.setDefaultStyles(style, node)

	if node.GetParent() != nil {
		ca.applyInheritedStyles(style, node.GetParent(), css)
	}

	// Apply CSS rules
	if css != nil {
		ca.applyCSSRules(style, node, css)
	}

	// Apply inline styles (style attribute)
	ca.applyInlineStyles(style, node)

	// Compute final values (resolve percentages, etc.)
	ca.computeFinalValues(style, node, css)

	return style
}

func (ca *cssApplicator) applyStylesToNode(node Node, css *CSS) {
	if node == nil {
		return
	}

	for _, child := range node.GetChildren() {
		ca.applyStylesToNode(child, css)
	}
}

func (ca *cssApplicator) setDefaultStyles(style *ComputedStyle, node Node) {
	for prop, value := range ca.defaultValues {
		style.Properties[prop] = value
	}

	// Element-specific defaults
	ca.setElementSpecificDefaults(style, node)
}

func (ca *cssApplicator) setElementSpecificDefaults(style *ComputedStyle, node Node) {
	if node.GetType() != ElementNodeType {
		return
	}

	tag := strings.ToLower(node.GetTag())
	elementDefaults := ca.getElementDefaults(tag)

	for prop, value := range elementDefaults {
		style.Properties[prop] = value
	}
}

func (ca *cssApplicator) getElementDefaults(tag string) map[string]string {
	defaults := make(map[string]string)

	// Block display elements
	switch tag {
	case "html", "body", "div", "blockquote", "section", "article", "aside", "nav", "main", "header", "footer":
		defaults["display"] = "block"
	case "p":
		defaults["display"] = "block"
		defaults["margin-top"] = "1em"
		defaults["margin-bottom"] = "1em"
	case "h1":
		defaults["display"] = "block"
		defaults["font-size"] = "2em"
		defaults["font-weight"] = "bold"
		defaults["margin-top"] = "0.67em"
		defaults["margin-bottom"] = "0.67em"
	case "h2":
		defaults["display"] = "block"
		defaults["font-size"] = "1.5em"
		defaults["font-weight"] = "bold"
		defaults["margin-top"] = "0.75em"
		defaults["margin-bottom"] = "0.75em"
	case "h3":
		defaults["display"] = "block"
		defaults["font-size"] = "1.17em"
		defaults["font-weight"] = "bold"
		defaults["margin-top"] = "0.83em"
		defaults["margin-bottom"] = "0.83em"
	case "h4":
		defaults["display"] = "block"
		defaults["font-size"] = "1em"
		defaults["font-weight"] = "bold"
		defaults["margin-top"] = "1.12em"
		defaults["margin-bottom"] = "1.12em"
	case "h5":
		defaults["display"] = "block"
		defaults["font-size"] = "0.83em"
		defaults["font-weight"] = "bold"
		defaults["margin-top"] = "1.5em"
		defaults["margin-bottom"] = "1.5em"
	case "h6":
		defaults["display"] = "block"
		defaults["font-size"] = "0.75em"
		defaults["font-weight"] = "bold"
		defaults["margin-top"] = "1.67em"
		defaults["margin-bottom"] = "1.67em"
	case "pre":
		defaults["display"] = "block"
		defaults["white-space"] = "pre"
		defaults["font-family"] = "monospace"
		defaults["background-color"] = "#f5f5f5"
		defaults["margin-top"] = "1em"
		defaults["margin-bottom"] = "1em"
		defaults["padding"] = "0.5em"
	case "ul", "ol":
		defaults["display"] = "block"
		defaults["margin-top"] = "1em"
		defaults["margin-bottom"] = "1em"
		defaults["padding-left"] = "2em"
	case "li":
		defaults["display"] = "list-item"
	case "b", "strong":
		defaults["font-weight"] = "bold"
	case "i", "em":
		defaults["font-style"] = "italic"
	case "a":
		defaults["color"] = "#0000EE"
		defaults["text-decoration"] = "underline"
	case "button":
		defaults["display"] = "inline-block"
		defaults["padding"] = "0.25em 0.5em"
		defaults["border"] = "1px solid #ccc"
		defaults["background-color"] = "#f0f0f0"
	case "br":
		defaults["display"] = "inline"
	case "span", "code", "small", "big":
		defaults["display"] = "inline"
	}

	return defaults
}

func (ca *cssApplicator) applyInheritedStyles(style *ComputedStyle, parent Node, css *CSS) {
	if parent == nil {
		return
	}

	parentStyle := ca.ComputeStyle(parent, css)

	for prop := range ca.inheritedProps {
		if value, exists := parentStyle.Properties[prop]; exists {
			style.Properties[prop] = value
		}
	}
}

type ruleMatch struct {
	rule        *CSSRule
	specificity int
}

func (ca *cssApplicator) applyCSSRules(style *ComputedStyle, node Node, css *CSS) {
	matches := ca.collectMatchingRules(node, css)
	ca.sortMatchesBySpecificity(matches)
	ca.applyMatchedRules(style, matches)
}

func (ca *cssApplicator) collectMatchingRules(node Node, css *CSS) []ruleMatch {
	var matches []ruleMatch

	for i := range css.Rules {
		rule := &css.Rules[i]
		if match := ca.findRuleMatch(node, rule); match != nil {
			matches = append(matches, *match)
		}
	}

	return matches
}

func (ca *cssApplicator) findRuleMatch(node Node, rule *CSSRule) *ruleMatch {
	for _, selector := range rule.Selectors {
		if ca.selectorMatcher.MatchesSelector(node, selector) {
			specificity := ca.selectorMatcher.CalculateSpecificity(selector)
			return &ruleMatch{
				rule:        rule,
				specificity: specificity,
			}
		}
	}
	return nil
}

func (ca *cssApplicator) applyMatchedRules(style *ComputedStyle, matches []ruleMatch) {
	for _, match := range matches {
		for prop, declaration := range match.rule.Declarations {
			style.Properties[prop] = declaration.Value.Raw
		}
	}
}

func (ca *cssApplicator) sortMatchesBySpecificity(matches []ruleMatch) {
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[i].specificity > matches[j].specificity {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}
}

func (ca *cssApplicator) applyInlineStyles(style *ComputedStyle, node Node) {
	if node.GetType() != ElementNodeType {
		return
	}

	styleAttr, exists := node.GetAttribute("style")
	if !exists || styleAttr == "" {
		return
	}

	// Parse as CSS body (property: value; pairs)
	declarations := ca.parseInlineStyles(styleAttr)

	// Apply inline styles (highest specificity)
	for prop, value := range declarations {
		style.Properties[prop] = value
	}
}

func (ca *cssApplicator) parseInlineStyles(styleAttr string) map[string]string {
	declarations := make(map[string]string)

	// Split by semicolon
	parts := strings.Split(styleAttr, ";")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split by colon
		colonIndex := strings.Index(part, ":")
		if colonIndex == -1 {
			continue
		}

		prop := strings.TrimSpace(part[:colonIndex])
		value := strings.TrimSpace(part[colonIndex+1:])

		if prop != "" && value != "" {
			declarations[prop] = value
		}
	}

	return declarations
}

func (ca *cssApplicator) computeFinalValues(style *ComputedStyle, node Node, css *CSS) {
	// Resolve relative font sizes
	if fontSize, exists := style.Properties["font-size"]; exists {
		if strings.HasSuffix(fontSize, "em") {
			ca.resolveEmFontSize(style, node, fontSize, css)
		} else if strings.HasSuffix(fontSize, "%") {
			ca.resolvePercentageFontSize(style, node, fontSize, css)
		}
	}

	// Resolve other relative units
	ca.resolveRelativeUnits(style)
}

func (ca *cssApplicator) resolveEmFontSize(style *ComputedStyle, node Node, fontSize string, css *CSS) {
	// Parse em value
	emStr := strings.TrimSuffix(fontSize, "em")
	emValue, err := strconv.ParseFloat(emStr, 64)
	if err != nil {
		return
	}

	// Get parent font size
	parentFontSize := ca.getParentFontSize(node, css)

	// Calculate pixel value
	pixelValue := emValue * parentFontSize
	style.Properties["font-size"] = strconv.FormatFloat(pixelValue, 'f', 2, 64) + "px"
}

func (ca *cssApplicator) resolvePercentageFontSize(style *ComputedStyle, node Node, fontSize string, css *CSS) {
	// Parse percentage value
	percentStr := strings.TrimSuffix(fontSize, "%")
	percentValue, err := strconv.ParseFloat(percentStr, 64)
	if err != nil {
		return
	}

	// Get parent font size
	parentFontSize := ca.getParentFontSize(node, css)

	// Calculate pixel value
	pixelValue := (percentValue / 100.0) * parentFontSize
	style.Properties["font-size"] = strconv.FormatFloat(pixelValue, 'f', 2, 64) + "px"
}

func (ca *cssApplicator) getParentFontSize(node Node, css *CSS) float64 {
	parent := node.GetParent()
	if parent == nil {
		return DefaultFontSize
	}

	parentStyle := ca.ComputeStyle(parent, css)
	parentFontSizeStr, exists := parentStyle.Properties["font-size"]
	if !exists {
		return DefaultFontSize
	}

	if strings.HasSuffix(parentFontSizeStr, "px") {
		pixelStr := strings.TrimSuffix(parentFontSizeStr, "px")
		if val, err := strconv.ParseFloat(pixelStr, 64); err == nil {
			return val
		}
	}

	return DefaultFontSize
}

func (ca *cssApplicator) resolveRelativeUnits(style *ComputedStyle) {
	// Resolve em units in margins, padding, etc.
	properties := []string{
		"margin-top", "margin-bottom", "margin-left", "margin-right",
		"padding-top", "padding-bottom", "padding-left", "padding-right",
	}

	fontSize := DefaultFontSize
	if fs, exists := style.Properties["font-size"]; exists {
		if strings.HasSuffix(fs, "px") {
			if val, err := strconv.ParseFloat(strings.TrimSuffix(fs, "px"), 64); err == nil {
				fontSize = val
			}
		}
	}

	for _, prop := range properties {
		if value, exists := style.Properties[prop]; exists {
			if strings.HasSuffix(value, "em") {
				emStr := strings.TrimSuffix(value, "em")
				if emValue, err := strconv.ParseFloat(emStr, 64); err == nil {
					pixelValue := emValue * fontSize
					style.Properties[prop] = strconv.FormatFloat(pixelValue, 'f', 2, 64) + "px"
				}
			}
		}
	}
}
