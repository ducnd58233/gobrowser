package browser

import (
	"regexp"
	"strconv"
	"strings"
)

type Viewport struct {
	Width  float64
	Height float64
}

type CSSApplicator interface {
	ApplyCSS(root Node, css *CSS, styles map[Node]Style)
}

type cssApplicator struct {
	selectorMatcher SelectorMatcher
	viewport        *Viewport
}

func NewCSSApplicator() CSSApplicator {
	return &cssApplicator{
		selectorMatcher: NewSelectorMatcher(),
	}
}

func (a *cssApplicator) ApplyCSS(root Node, css *CSS, styles map[Node]Style) {
	if root == nil || css == nil {
		return
	}

	for _, rule := range css.Rules {
		a.applyRuleToTree(root, &rule, styles)
	}

	if a.viewport != nil {
		for _, mediaRule := range css.MediaRules {
			if a.evaluateMediaQuery(mediaRule.MediaQuery, *a.viewport) {
				for _, rule := range mediaRule.Rules {
					a.applyRuleToTree(root, &rule, styles)
				}
			}
		}
	}

	a.applyInheritanceToTree(root, styles, nil)
}

func (a *cssApplicator) applyRuleToTree(node Node, rule *CSSRule, styles map[Node]Style) {
	if node == nil {
		return
	}

	for _, selector := range rule.Selectors {
		if a.selectorMatcher.MatchesSelector(node, selector) {
			a.applyRuleToNode(node, rule, a.getOrCreateStyle(node, styles))
		}
	}
}

func (a *cssApplicator) getOrCreateStyle(node Node, styles map[Node]Style) Style {
	style, exists := styles[node]
	if !exists {
		style = NewStyle()
		styles[node] = style
	}
	return style
}

func (a *cssApplicator) applyRuleToNode(node Node, rule *CSSRule, style Style) {
	if node == nil || rule == nil || style == nil {
		return
	}

	for property, declaration := range rule.Declarations {
		existingValue := style.GetProperty(property)

		if existingValue.Raw != "" || !declaration.Important {
			continue
		}

		style.SetProperty(property, declaration.Value)
	}
}

func (a *cssApplicator) applyInheritance(parent Style, child Style) {
	if parent == nil || child == nil {
		return
	}

	for property := range parent.GetInherited() {
		parentValue := parent.GetProperty(property)
		if parentValue.Raw != "" && child.GetProperty(property).Raw == "" {
			child.SetProperty(property, parentValue)
		}
	}
}

func (a *cssApplicator) applyInheritanceToTree(node Node, styles map[Node]Style, parentStyle Style) {
	if node == nil {
		return
	}

	currentStyle := a.getOrCreateStyle(node, styles)

	if parentStyle != nil {
		a.applyInheritance(parentStyle, currentStyle)
	}

	// Apply to children
	for _, child := range node.GetChildren() {
		a.applyInheritanceToTree(child, styles, currentStyle)
	}
}

func (a *cssApplicator) evaluateMediaQuery(mediaQuery string, viewport Viewport) bool {
	mediaQuery = strings.TrimSpace(strings.ToLower(mediaQuery))

	if mediaQuery == "all" || mediaQuery == "screen" {
		return true
	}
	if mediaQuery == "print" {
		return false
	}

	if strings.Contains(mediaQuery, "min-width") {
		pattern := regexp.MustCompile(`min-width:\s*(\d+)px`)
		matches := pattern.FindStringSubmatch(mediaQuery)
		if len(matches) > 1 {
			if minWidth, err := strconv.Atoi(matches[1]); err == nil {
				return float64(minWidth) <= viewport.Width
			}
		}
	}

	if strings.Contains(mediaQuery, "max-width") {
		pattern := regexp.MustCompile(`max-width:\s*(\d+)px`)
		matches := pattern.FindStringSubmatch(mediaQuery)
		if len(matches) > 1 {
			if maxWidth, err := strconv.Atoi(matches[1]); err == nil {
				return float64(maxWidth) >= viewport.Width
			}
		}
	}

	return true
}
