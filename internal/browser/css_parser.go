package browser

import (
	"regexp"
	"strconv"
	"strings"
)

type CSS struct {
	Rules      []CSSRule
	MediaRules []MediaRule
	Imports    []string
	Charset    string
}

// PrintTree returns a formatted string representation of the CSS structure for debugging
func (css *CSS) PrintTree() string {
	var builder strings.Builder

	builder.WriteString("CSS Document\n")
	builder.WriteString("===========\n\n")

	if css.Charset != "" {
		builder.WriteString("Charset: " + css.Charset + "\n\n")
	}

	if len(css.Imports) > 0 {
		builder.WriteString("Imports:\n")
		for _, imp := range css.Imports {
			builder.WriteString("  @import \"" + imp + "\"\n")
		}
		builder.WriteString("\n")
	}

	if len(css.Rules) > 0 {
		builder.WriteString("CSS Rules (" + strconv.Itoa(len(css.Rules)) + "):\n")
		for i, rule := range css.Rules {
			builder.WriteString(css.formatRule(i+1, &rule))
		}
		builder.WriteString("\n")
	}

	if len(css.MediaRules) > 0 {
		builder.WriteString("Media Rules (" + strconv.Itoa(len(css.MediaRules)) + "):\n")
		for i, mediaRule := range css.MediaRules {
			builder.WriteString(css.formatMediaRule(i+1, &mediaRule))
		}
	}

	return builder.String()
}

func (css *CSS) formatRule(index int, rule *CSSRule) string {
	var builder strings.Builder

	builder.WriteString("  Rule " + strconv.Itoa(index) + ":\n")

	// Selectors
	builder.WriteString("    Selectors: ")
	for i, selector := range rule.Selectors {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString("\"" + selector + "\"")
		if specificity, ok := rule.Specificity[selector]; ok {
			builder.WriteString(" (specificity: " + strconv.Itoa(specificity) + ")")
		}
	}
	builder.WriteString("\n")

	// Declarations
	if len(rule.Declarations) > 0 {
		builder.WriteString("    Declarations:\n")
		for property, declaration := range rule.Declarations {
			builder.WriteString("      " + property + ": " + declaration.Value.Raw)
			if declaration.Important {
				builder.WriteString(" !important")
			}
			builder.WriteString(" (type: " + css.getValueTypeName(declaration.Value.ValueType) + ")\n")
		}
	}

	builder.WriteString("\n")
	return builder.String()
}

func (css *CSS) formatMediaRule(index int, mediaRule *MediaRule) string {
	var builder strings.Builder

	builder.WriteString("  Media Rule " + strconv.Itoa(index) + ":\n")
	builder.WriteString("    Query: \"" + mediaRule.MediaQuery + "\"\n")
	builder.WriteString("    Rules (" + strconv.Itoa(len(mediaRule.Rules)) + "):\n")

	for i, rule := range mediaRule.Rules {
		lines := strings.Split(css.formatRule(i+1, &rule), "\n")
		for _, line := range lines {
			if line != "" {
				builder.WriteString("    " + line + "\n")
			}
		}
	}

	return builder.String()
}

func (css *CSS) getValueTypeName(valueType CSSValueType) string {
	switch valueType {
	case CSSValueKeyword:
		return "keyword"
	case CSSValueNumber:
		return "number"
	case CSSValueLength:
		return "length"
	case CSSValuePercentage:
		return "percentage"
	case CSSValueColor:
		return "color"
	case CSSValueURL:
		return "url"
	case CSSValueString:
		return "string"
	case CSSValueFunction:
		return "function"
	default:
		return "unknown"
	}
}

type MediaRule struct {
	MediaQuery string
	Rules      []CSSRule
}

type CSSDeclaration struct {
	Property  string
	Value     CSSValue
	Important bool
}

type CSSRule struct {
	Selectors    []string
	Declarations map[string]CSSDeclaration
	Specificity  map[string]int
}

type CSSParser interface {
	Parse() *CSS
	ParseRule(rule string) *CSSRule
	ParseDeclaration(declaration string) CSSDeclaration
	ParseValue(value string) CSSValue
	ParseMediaQuery(mediaRule string) *MediaRule
}

type cssParser struct {
	content       string
	textProcessor TextProcessor
	colorParser   ColorParser
	unitParser    UnitParser
}

func NewCSSParser(content string) CSSParser {
	return &cssParser{
		content:       content,
		textProcessor: NewTextProcessor(),
		colorParser:   NewColorParser(),
		unitParser:    NewUnitParser(),
	}
}

func (p *cssParser) Parse() *CSS {
	css := &CSS{
		Rules:      make([]CSSRule, 0),
		MediaRules: make([]MediaRule, 0),
		Imports:    make([]string, 0),
	}

	if p.content == "" {
		return css
	}

	commentRegex := regexp.MustCompile(`/\*.*?\*/`)
	content := commentRegex.ReplaceAllString(p.content, "")

	// Parse @-rules first
	css.Charset, content = p.extractCharset(content)
	css.Imports, content = p.extractImports(content)

	// Parse media queries
	mediaRules, content := p.extractMediaRules(content)
	css.MediaRules = mediaRules

	// Parse regular rules
	rules := p.splitRules(content)
	for _, ruleText := range rules {
		if rule := p.ParseRule(ruleText); rule != nil {
			css.Rules = append(css.Rules, *rule)
		}
	}

	return css
}

func (p *cssParser) ParseMediaQuery(mediaRule string) *MediaRule {
	mediaPattern := regexp.MustCompile(`@media\s+([^{]+)\s*\{((?:[^{}]*\{[^{}]*\})*[^{}]*)\}`)
	matches := mediaPattern.FindStringSubmatch(mediaRule)

	if len(matches) < 3 {
		return nil
	}

	mediaQuery := strings.TrimSpace(matches[1])
	rulesContent := matches[2]

	rules := make([]CSSRule, 0)
	ruleTexts := p.splitRules(rulesContent)
	for _, ruleText := range ruleTexts {
		if rule := p.ParseRule(ruleText); rule != nil {
			rules = append(rules, *rule)
		}
	}

	return &MediaRule{
		MediaQuery: mediaQuery,
		Rules:      rules,
	}
}

func (p *cssParser) ParseRule(rule string) *CSSRule {
	rule = strings.TrimSpace(rule)
	if rule == "" || !strings.Contains(rule, "{") {
		return nil
	}

	braceIndex := strings.Index(rule, "{")
	selectorsPart := strings.TrimSpace(rule[:braceIndex])
	declarationsPart := strings.TrimSpace(rule[braceIndex+1:])

	declarationsPart = strings.TrimSuffix(declarationsPart, "}")

	selectors := p.parseSelectors(selectorsPart)
	declarations := p.parseDeclarations(declarationsPart)

	// Calculate specificity for each selector
	specificity := make(map[string]int)
	for _, selector := range selectors {
		specificity[selector] = p.calculateSpecificity(selector)
	}

	return &CSSRule{
		Selectors:    selectors,
		Declarations: declarations,
		Specificity:  specificity,
	}
}

// ParseDeclaration parses a single CSS declaration with !important support
func (p *cssParser) ParseDeclaration(declaration string) CSSDeclaration {
	declaration = strings.TrimSpace(declaration)
	if declaration == "" || !strings.Contains(declaration, ":") {
		return CSSDeclaration{}
	}

	colonIndex := strings.Index(declaration, ":")
	property := strings.TrimSpace(declaration[:colonIndex])
	valueText := strings.TrimSpace(declaration[colonIndex+1:])

	important := false
	if strings.HasSuffix(valueText, "!important") {
		important = true
		valueText = strings.TrimSpace(valueText[:len(valueText)-10])
	}

	value := p.ParseValue(valueText)

	return CSSDeclaration{
		Property:  property,
		Value:     value,
		Important: important,
	}
}

func (p *cssParser) ParseValue(valueText string) CSSValue {
	valueText = strings.TrimSpace(valueText)
	if valueText == "" {
		return CSSValue{Raw: valueText, ValueType: CSSValueKeyword}
	}

	// URL values
	if strings.HasPrefix(valueText, "url(") && strings.HasSuffix(valueText, ")") {
		return CSSValue{
			Raw:       valueText,
			ValueType: CSSValueURL,
		}
	}

	// String values
	if (strings.HasPrefix(valueText, "\"") && strings.HasSuffix(valueText, "\"")) ||
		(strings.HasPrefix(valueText, "'") && strings.HasSuffix(valueText, "'")) {
		return CSSValue{
			Raw:       valueText,
			ValueType: CSSValueString,
		}
	}

	// Function values
	if strings.Contains(valueText, "(") && strings.HasSuffix(valueText, ")") {
		return CSSValue{
			Raw:       valueText,
			ValueType: CSSValueFunction,
		}
	}

	// Color values
	if r, g, b, a, err := p.colorParser.ParseColor(valueText); err == nil {
		return CSSValue{
			Raw:       valueText,
			ValueType: CSSValueColor,
			Color:     Color{R: r, G: g, B: b, A: a},
		}
	}

	// Number with unit (length, percentage)
	if number, unit, ok := p.unitParser.ParseUnit(valueText); ok {
		valueType := CSSValueNumber
		if unit == "%" {
			valueType = CSSValuePercentage
		} else if unit != "" {
			valueType = CSSValueLength
		}

		return CSSValue{
			Raw:       valueText,
			ValueType: valueType,
			Number:    number,
			Unit:      unit,
		}
	}

	// Multiple keywords (e.g., "Arial, sans-serif")
	if strings.Contains(valueText, ",") {
		keywords := strings.Split(valueText, ",")
		for i, keyword := range keywords {
			keywords[i] = strings.TrimSpace(keyword)
		}
		return CSSValue{
			Raw:       valueText,
			ValueType: CSSValueKeyword,
			Keywords:  keywords,
		}
	}

	// Single keyword
	return CSSValue{
		Raw:       valueText,
		ValueType: CSSValueKeyword,
	}
}

func (p *cssParser) extractCharset(content string) (string, string) {
	charsetPattern := regexp.MustCompile(`@charset\s+["']([^"']+)["']\s*;`)
	matches := charsetPattern.FindStringSubmatch(content)
	if len(matches) > 1 {
		remaining := charsetPattern.ReplaceAllString(content, "")
		return matches[1], remaining
	}
	return "", content
}

func (p *cssParser) extractImports(content string) ([]string, string) {
	importPattern := regexp.MustCompile(`@import\s+(?:url\()?["']?([^"')]+)["']?\)?[^;]*;`)
	matches := importPattern.FindAllStringSubmatch(content, -1)

	imports := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			imports = append(imports, match[1])
		}
	}

	remaining := importPattern.ReplaceAllString(content, "")
	return imports, remaining
}

func (p *cssParser) extractMediaRules(content string) ([]MediaRule, string) {
	mediaPattern := regexp.MustCompile(`@media\s+([^{]+)\s*\{((?:[^{}]*\{[^{}]*\})*[^{}]*)\}`)
	matches := mediaPattern.FindAllStringSubmatch(content, -1)

	mediaRules := make([]MediaRule, 0, len(matches))
	for _, match := range matches {
		if len(match) > 2 {
			mediaRule := p.ParseMediaQuery(match[0])
			if mediaRule != nil {
				mediaRules = append(mediaRules, *mediaRule)
			}
		}
	}

	remaining := mediaPattern.ReplaceAllString(content, "")
	return mediaRules, remaining
}

type ruleParseState struct {
	braceCount int
	inString   bool
	escapeNext bool
}

func (p *cssParser) splitRules(content string) []string {
	rules := make([]string, 0)
	var currentRule strings.Builder

	state := &ruleParseState{
		braceCount: 0,
		inString:   false,
		escapeNext: false,
	}

	for _, char := range content {
		if p.shouldSkipChar(char, state, &currentRule) {
			continue
		}

		p.updateParseState(char, state)
		currentRule.WriteRune(char)

		if p.shouldFinishRule(char, state) {
			if rule := p.extractRule(&currentRule); rule != "" {
				rules = append(rules, rule)
			}
		}
	}

	if rule := p.extractRule(&currentRule); rule != "" {
		rules = append(rules, rule)
	}

	return rules
}

func (p *cssParser) shouldSkipChar(char rune, state *ruleParseState, currentRule *strings.Builder) bool {
	if state.escapeNext {
		currentRule.WriteRune(char)
		state.escapeNext = false
		return true
	}

	if char == '\\' {
		state.escapeNext = true
		currentRule.WriteRune(char)
		return true
	}

	return false
}

func (p *cssParser) updateParseState(char rune, state *ruleParseState) {
	if char == '"' || char == '\'' {
		state.inString = !state.inString
	}

	if !state.inString {
		switch char {
		case '{':
			state.braceCount++
		case '}':
			state.braceCount--
		}
	}
}

func (p *cssParser) shouldFinishRule(char rune, state *ruleParseState) bool {
	return !state.inString && state.braceCount == 0 && char == '}'
}

// extractRule extracts and resets the current rule
func (p *cssParser) extractRule(currentRule *strings.Builder) string {
	rule := strings.TrimSpace(currentRule.String())
	currentRule.Reset()
	return rule
}

// parseSelectors parses comma-separated selectors
func (p *cssParser) parseSelectors(selectorsPart string) []string {
	selectors := strings.Split(selectorsPart, ",")
	result := make([]string, 0, len(selectors))

	for _, selector := range selectors {
		selector = strings.TrimSpace(selector)
		if selector != "" {
			result = append(result, selector)
		}
	}

	return result
}

// parseDeclarations parses semicolon-separated declarations
func (p *cssParser) parseDeclarations(declarationsPart string) map[string]CSSDeclaration {
	declarations := make(map[string]CSSDeclaration)

	parts := strings.Split(declarationsPart, ";")
	for _, part := range parts {
		declaration := p.ParseDeclaration(part)
		if declaration.Property != "" {
			declarations[declaration.Property] = declaration
		}
	}

	return declarations
}

func (p *cssParser) calculateSpecificity(selector string) int {
	selector = strings.TrimSpace(selector)

	// Specificity calculation: inline=1000, IDs=100, classes/attributes/pseudo=10, elements/pseudo-elements=1
	specificity := 0

	// Count IDs
	idCount := strings.Count(selector, "#")
	specificity += idCount * 100

	// Count classes, attributes, and pseudo-classes
	classCount := strings.Count(selector, ".")
	attributeCount := strings.Count(selector, "[")

	// Count pseudo-classes (single colon, not double colon)
	pseudoClassCount := p.countPseudoClasses(selector)
	specificity += (classCount + attributeCount + pseudoClassCount) * 10

	// Count elements and pseudo-elements
	// Remove IDs, classes, attributes, and pseudo-classes first
	elementSelector := selector
	elementSelector = regexp.MustCompile(`#[^\s#.\[:]+`).ReplaceAllString(elementSelector, "")
	elementSelector = regexp.MustCompile(`\.[^\s#.\[:]+`).ReplaceAllString(elementSelector, "")
	elementSelector = regexp.MustCompile(`\[[^\]]*\]`).ReplaceAllString(elementSelector, "")
	elementSelector = regexp.MustCompile(`::[^:\s\[]+`).ReplaceAllString(elementSelector, " ")
	elementSelector = regexp.MustCompile(`:[^:\s\[]+`).ReplaceAllString(elementSelector, "")

	// Count remaining element names
	elementPattern := regexp.MustCompile(`\b[a-zA-Z][a-zA-Z0-9]*\b`)
	elementMatches := elementPattern.FindAllString(elementSelector, -1)
	elementCount := 0
	for _, match := range elementMatches {
		if !p.isCSSKeyword(match) {
			elementCount++
		}
	}
	specificity += elementCount

	return specificity
}

func (p *cssParser) isCSSKeyword(word string) bool {
	keywords := map[string]bool{
		"and": true, "not": true, "only": true,
		"all": true, "screen": true, "print": true,
	}
	return keywords[strings.ToLower(word)]
}

func (p *cssParser) countPseudoClasses(selector string) int {
	count := 0
	for i := 0; i < len(selector); i++ {
		if selector[i] == ':' {
			// Check if it's a pseudo-element (double colon)
			if i+1 < len(selector) && selector[i+1] == ':' {
				i++ // Skip the second colon
				continue
			}
			// It's a pseudo-class (single colon)
			count++
		}
	}
	return count
}
