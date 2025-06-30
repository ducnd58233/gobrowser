package browser

import (
	"regexp"
	"strconv"
	"strings"
)

type CSSValueType int

const (
	CSSValueKeyword CSSValueType = iota
	CSSValueNumber
	CSSValueLength
	CSSValuePercentage
	CSSValueColor
	CSSValueURL
	CSSValueString
	CSSValueFunction
)

type Color struct {
	R, G, B, A uint8
}

type CSSValue struct {
	Raw       string
	ValueType CSSValueType
	Number    float64
	Unit      string
	Color     Color
	Keywords  []string
}

type Style interface {
	GetProperty(property string) CSSValue
	SetProperty(property string, value CSSValue)
	GetInheritedProperties() map[string]bool
	IsInherited(property string) bool
	Clone() Style
	Merge(other Style)
}

type style struct {
	properties map[string]CSSValue
	inherited  map[string]bool
}

func NewStyle() Style {
	return &style{
		properties: map[string]CSSValue{
			PropColor:             {Raw: DefaultTextColor, ValueType: CSSValueColor},
			PropBackgroundColor:   {Raw: DefaultBgColor, ValueType: CSSValueColor},
			PropFontFamily:        {Raw: DefaultTextFont, ValueType: CSSValueKeyword},
			PropFontSize:          {Raw: "16px", ValueType: CSSValueLength, Number: 16, Unit: "px"},
			PropFontWeight:        {Raw: ValueNormal, ValueType: CSSValueKeyword},
			PropFontStyle:         {Raw: ValueNormal, ValueType: CSSValueKeyword},
			PropTextAlign:         {Raw: ValueLeft, ValueType: CSSValueKeyword},
			PropTextDecoration:    {Raw: ValueNone, ValueType: CSSValueKeyword},
			PropDisplay:           {Raw: ValueBlock, ValueType: CSSValueKeyword},
			PropMargin:            {Raw: "0", ValueType: CSSValueLength, Number: 0, Unit: "px"},
			PropMarginTop:         {Raw: "0", ValueType: CSSValueLength, Number: 0, Unit: "px"},
			PropMarginRight:       {Raw: "0", ValueType: CSSValueLength, Number: 0, Unit: "px"},
			PropMarginBottom:      {Raw: "0", ValueType: CSSValueLength, Number: 0, Unit: "px"},
			PropMarginLeft:        {Raw: "0", ValueType: CSSValueLength, Number: 0, Unit: "px"},
			PropPadding:           {Raw: "0", ValueType: CSSValueLength, Number: 0, Unit: "px"},
			PropPaddingTop:        {Raw: "0", ValueType: CSSValueLength, Number: 0, Unit: "px"},
			PropPaddingRight:      {Raw: "0", ValueType: CSSValueLength, Number: 0, Unit: "px"},
			PropPaddingBottom:     {Raw: "0", ValueType: CSSValueLength, Number: 0, Unit: "px"},
			PropPaddingLeft:       {Raw: "0", ValueType: CSSValueLength, Number: 0, Unit: "px"},
			PropBorder:            {Raw: ValueNone, ValueType: CSSValueKeyword},
			PropWidth:             {Raw: ValueAuto, ValueType: CSSValueKeyword},
			PropHeight:            {Raw: ValueAuto, ValueType: CSSValueKeyword},
			PropMinWidth:          {Raw: "0", ValueType: CSSValueLength, Number: 0, Unit: "px"},
			PropMinHeight:         {Raw: "0", ValueType: CSSValueLength, Number: 0, Unit: "px"},
			PropMaxWidth:          {Raw: ValueNone, ValueType: CSSValueKeyword},
			PropMaxHeight:         {Raw: ValueNone, ValueType: CSSValueKeyword},
			PropPosition:          {Raw: ValueStatic, ValueType: CSSValueKeyword},
			PropTop:               {Raw: ValueAuto, ValueType: CSSValueKeyword},
			PropRight:             {Raw: ValueAuto, ValueType: CSSValueKeyword},
			PropBottom:            {Raw: ValueAuto, ValueType: CSSValueKeyword},
			PropLeft:              {Raw: ValueAuto, ValueType: CSSValueKeyword},
			PropZIndex:            {Raw: ValueAuto, ValueType: CSSValueKeyword},
			PropOpacity:           {Raw: "1", ValueType: CSSValueNumber, Number: 1},
			PropVisibility:        {Raw: ValueVisible, ValueType: CSSValueKeyword},
			PropOverflow:          {Raw: ValueVisible, ValueType: CSSValueKeyword},
			PropOverflowX:         {Raw: ValueVisible, ValueType: CSSValueKeyword},
			PropOverflowY:         {Raw: ValueVisible, ValueType: CSSValueKeyword},
			PropLineHeight:        {Raw: ValueNormal, ValueType: CSSValueKeyword},
			PropLetterSpacing:     {Raw: ValueNormal, ValueType: CSSValueKeyword},
			PropWordSpacing:       {Raw: ValueNormal, ValueType: CSSValueKeyword},
			PropTextTransform:     {Raw: ValueNone, ValueType: CSSValueKeyword},
			PropWhiteSpace:        {Raw: ValueNormal, ValueType: CSSValueKeyword},
			PropCursor:            {Raw: ValueAuto, ValueType: CSSValueKeyword},
			PropListStyle:         {Raw: ValueNone, ValueType: CSSValueKeyword},
			PropListStyleType:     {Raw: "disc", ValueType: CSSValueKeyword},
			PropListStylePosition: {Raw: "outside", ValueType: CSSValueKeyword},
		},
		inherited: map[string]bool{
			PropColor:             true,
			PropFontFamily:        true,
			PropFontSize:          true,
			PropFontWeight:        true,
			PropFontStyle:         true,
			PropTextAlign:         true,
			PropTextTransform:     true,
			PropLineHeight:        true,
			PropLetterSpacing:     true,
			PropWordSpacing:       true,
			PropWhiteSpace:        true,
			PropVisibility:        true,
			PropCursor:            true,
			PropListStyle:         true,
			PropListStyleType:     true,
			PropListStylePosition: true,
		},
	}
}

func (s *style) GetInheritedProperties() map[string]bool {
	return s.inherited
}

func (s *style) GetProperty(property string) CSSValue {
	if value, exists := s.properties[property]; exists {
		return value
	}
	return CSSValue{Raw: "", ValueType: CSSValueKeyword}
}

func (s *style) SetProperty(property string, value CSSValue) {
	s.properties[property] = value
}

func (s *style) IsInherited(property string) bool {
	if inherited, exists := s.inherited[property]; exists {
		return inherited
	}
	return false
}

func (s *style) Clone() Style {
	cloned := &style{
		properties: make(map[string]CSSValue),
		inherited:  make(map[string]bool),
	}

	for key, value := range s.properties {
		cloned.properties[key] = value
	}

	for key, value := range s.inherited {
		cloned.inherited[key] = value
	}

	return cloned
}

func (s *style) Merge(other Style) {
	if otherStyle, ok := other.(*style); ok {
		for key, value := range otherStyle.properties {
			s.properties[key] = value
		}
	}
}

type UnitParser interface {
	ParseUnit(value string) (float64, string, bool)
	ConvertToPixels(value float64, unit string, baseFontSize float64) float64
	ResolveRelativeUnits(value CSSValue, baseFontSize float64) CSSValue
}

type unitParser struct {
	unitConversions map[string]float64
}

func NewUnitParser() UnitParser {
	return &unitParser{
		unitConversions: map[string]float64{
			"px": 1.0,
			"pt": 1.333,  // 1pt = 1.333px at 96dpi
			"pc": 16.0,   // 1pc = 16px
			"in": 96.0,   // 1in = 96px at 96dpi
			"cm": 37.795, // 1cm = 37.795px at 96dpi
			"mm": 3.7795, // 1mm = 3.7795px at 96dpi
		},
	}
}

func (up *unitParser) ParseUnit(value string) (float64, string, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, "", false
	}

	// Handle unitless numbers
	if number, err := strconv.ParseFloat(value, 64); err == nil {
		return number, "", true
	}

	pattern := regexp.MustCompile(`^([+-]?(?:\d*\.)?\d+)(px|em|rem|%|pt|pc|in|cm|mm|ex|ch|vw|vh|vmin|vmax|deg|rad|turn|s|ms)?$`)
	matches := pattern.FindStringSubmatch(value)

	if len(matches) >= 2 {
		if number, err := strconv.ParseFloat(matches[1], 64); err == nil {
			unit := ""
			if len(matches) > 2 {
				unit = matches[2]
			}
			return number, unit, true
		}
	}

	return 0, "", false
}

func (up *unitParser) ConvertToPixels(value float64, unit string, baseFontSize float64) float64 {
	switch unit {
	case "px", "":
		return value
	case "em":
		return value * baseFontSize
	case "rem":
		return value * DefaultFontSize
	case "%":
		return value / 100 * baseFontSize
	case "ex":
		return value * baseFontSize * 0.5
	case "ch":
		return value * baseFontSize * 0.5
	case "vw":
		return value * 8
	case "vh":
		return value * 8
	case "vmin", "vmax":
		return value * 8
	default:
		if conversion, exists := up.unitConversions[unit]; exists {
			return value * conversion
		}
		return value
	}
}

func (up *unitParser) ResolveRelativeUnits(value CSSValue, baseFontSize float64) CSSValue {
	if value.ValueType != CSSValueLength && value.ValueType != CSSValuePercentage {
		return value
	}

	if value.Unit == "em" || value.Unit == "rem" || value.Unit == "%" {
		resolvedValue := up.ConvertToPixels(value.Number, value.Unit, baseFontSize)
		return CSSValue{
			Raw:       strconv.FormatFloat(resolvedValue, 'f', 2, 64) + "px",
			ValueType: CSSValueLength,
			Number:    resolvedValue,
			Unit:      "px",
		}
	}

	return value
}

type SelectorMatcher interface {
	MatchesSelector(node Node, selector string) bool
	CalculateSpecificity(selector string) int
}

type selectorMatcher struct {
	selectorCache map[string]int
}

func NewSelectorMatcher() SelectorMatcher {
	return &selectorMatcher{
		selectorCache: make(map[string]int),
	}
}

func (m *selectorMatcher) MatchesSelector(node Node, selector string) bool {
	if node == nil || selector == "" {
		return false
	}

	selector = strings.TrimSpace(selector)

	// Handle complex selectors (descendant, child, sibling)
	if strings.Contains(selector, " ") && !strings.Contains(selector, ":") {
		return m.matchesComplexSelector(node, selector)
	}

	if strings.Contains(selector, ">") {
		return m.matchesChildSelector(node, selector)
	}

	if strings.Contains(selector, "+") {
		return m.matchesAdjacentSiblingSelector(node, selector)
	}

	if strings.Contains(selector, "~") {
		return m.matchesGeneralSiblingSelector(node, selector)
	}

	return m.matchesSimpleSelector(node, selector)
}

func (m *selectorMatcher) CalculateSpecificity(selector string) int {
	if cached, exists := m.selectorCache[selector]; exists {
		return cached
	}

	specificity := m.calculateSelectorSpecificity(selector)
	m.selectorCache[selector] = specificity
	return specificity
}

func (m *selectorMatcher) calculateSelectorSpecificity(selector string) int {
	selector = strings.TrimSpace(selector)

	// CSS specificity: inline=1000, IDs=100, classes/attributes/pseudo=10, elements=1
	var idCount, classCount, elementCount int

	// Count IDs
	idMatches := regexp.MustCompile(`#[^\s#.\[:]+`).FindAllString(selector, -1)
	idCount = len(idMatches)

	// Count classes
	classMatches := regexp.MustCompile(`\.[^\s#.\[:]+`).FindAllString(selector, -1)
	classCount = len(classMatches)

	// Count attributes
	attributeMatches := regexp.MustCompile(`\[[^\]]*\]`).FindAllString(selector, -1)
	classCount += len(attributeMatches)

	// Count pseudo-classes (single colon, not double)
	pseudoClassCount := m.countPseudoClasses(selector)
	classCount += pseudoClassCount

	cleanSelector := selector
	cleanSelector = regexp.MustCompile(`#[^\s#.\[:]+`).ReplaceAllString(cleanSelector, "")
	cleanSelector = regexp.MustCompile(`\.[^\s#.\[:]+`).ReplaceAllString(cleanSelector, "")
	cleanSelector = regexp.MustCompile(`\[[^\]]*\]`).ReplaceAllString(cleanSelector, "")
	cleanSelector = regexp.MustCompile(`::[^:\s\[]+`).ReplaceAllString(cleanSelector, " ")
	cleanSelector = regexp.MustCompile(`:[^:\s\[]+`).ReplaceAllString(cleanSelector, "")

	elementMatches := regexp.MustCompile(`\b[a-zA-Z][a-zA-Z0-9]*\b`).FindAllString(cleanSelector, -1)
	for _, match := range elementMatches {
		if !m.isCSSKeyword(match) {
			elementCount++
		}
	}

	return idCount*100 + classCount*10 + elementCount
}

func (m *selectorMatcher) countPseudoClasses(selector string) int {
	count := 0
	for i := 0; i < len(selector); i++ {
		if i+1 < len(selector) && selector[i+1] == ':' {
			i++ // Skip pseudo-element
			continue
		}
		count++
	}
	return count
}

func (m *selectorMatcher) isCSSKeyword(word string) bool {
	keywords := map[string]bool{
		"and": true, "not": true, "only": true,
		"all": true, "screen": true, "print": true,
	}
	return keywords[strings.ToLower(word)]
}

// simple selectors (element, class, ID, attribute, pseudo)
func (m *selectorMatcher) matchesSimpleSelector(node Node, selector string) bool {
	if node.GetType() != ElementNodeType {
		return false
	}

	selector = strings.TrimSpace(selector)

	if selector == "*" {
		return true
	}

	parts := m.parseCompoundSelector(selector)
	for _, part := range parts {
		if !m.matchesSelectorPart(node, part) {
			return false
		}
	}

	return true
}

func (m *selectorMatcher) parseCompoundSelector(selector string) []string {
	var parts []string
	var currentPart strings.Builder

	for _, char := range selector {
		switch char {
		case '.', '#', '[', ':':
			if currentPart.Len() > 0 {
				parts = append(parts, currentPart.String())
				currentPart.Reset()
			}
			currentPart.WriteRune(char)
		case ']':
			currentPart.WriteRune(char)
			parts = append(parts, currentPart.String())
			currentPart.Reset()
		default:
			currentPart.WriteRune(char)
		}
	}

	if currentPart.Len() > 0 {
		parts = append(parts, currentPart.String())
	}

	return parts
}

func (m *selectorMatcher) matchesSelectorPart(node Node, part string) bool {
	if part == "" {
		return true
	}

	switch part[0] {
	case '.':
		// Class selector
		className := part[1:]
		return node.HasClass(className)
	case '#':
		// ID selector
		id := part[1:]
		return node.GetID() == id
	case '[':
		// Attribute selector
		return m.matchesAttributeSelector(node, part)
	case ':':
		// Pseudo-class selector
		pseudoClass := part[1:]
		return m.matchesPseudoSelector(node, pseudoClass)
	default:
		// Element selector
		return strings.EqualFold(node.GetTag(), part)
	}
}

// attribute selectors [attr], [attr=value], etc.
func (m *selectorMatcher) matchesAttributeSelector(node Node, selector string) bool {
	if !strings.HasPrefix(selector, "[") || !strings.HasSuffix(selector, "]") {
		return false
	}

	content := selector[1 : len(selector)-1]

	// [attr] - attribute exists
	if !strings.Contains(content, "=") {
		_, exists := node.GetAttribute(content)
		return exists
	}

	// [attr=value] or [attr*=value] etc.
	attrName, operator, value := m.parseAttributeSelector(content)
	if attrName == "" {
		return false
	}

	attrValue, exists := node.GetAttribute(strings.TrimSpace(attrName))
	if !exists {
		return false
	}

	return m.matchesAttributeOperator(attrValue, operator, value)
}

// extracts attribute name, operator, and value from selector content
func (m *selectorMatcher) parseAttributeSelector(content string) (string, string, string) {
	operators := []string{"*=", "^=", "$=", "~=", "|=", "="}

	for _, op := range operators {
		if strings.Contains(content, op) {
			parts := strings.SplitN(content, op, 2)
			if len(parts) == 2 {
				attrName := parts[0]
				value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
				return attrName, op, value
			}
		}
	}

	return "", "", ""
}

func (m *selectorMatcher) matchesAttributeOperator(attrValue, operator, value string) bool {
	switch operator {
	case "=":
		return attrValue == value
	case "*=":
		return strings.Contains(attrValue, value)
	case "^=":
		return strings.HasPrefix(attrValue, value)
	case "$=":
		return strings.HasSuffix(attrValue, value)
	case "~=":
		return m.matchesWordInList(attrValue, value)
	case "|=":
		return attrValue == value || strings.HasPrefix(attrValue, value+"-")
	default:
		return false
	}
}

func (m *selectorMatcher) matchesWordInList(attrValue, value string) bool {
	words := strings.Fields(attrValue)
	for _, word := range words {
		if word == value {
			return true
		}
	}
	return false
}

func (m *selectorMatcher) matchesPseudoSelector(node Node, pseudoClass string) bool {
	switch strings.ToLower(pseudoClass) {
	case "first-child":
		return m.isFirstChild(node)
	case "last-child":
		return m.isLastChild(node)
	case "first-of-type":
		return m.isFirstOfType(node)
	case "last-of-type":
		return m.isLastOfType(node)
	case "only-child":
		return m.isOnlyChild(node)
	case "only-of-type":
		return m.isOnlyOfType(node)
	case "empty":
		return len(node.GetChildren()) == 0
	case "root":
		return node.GetParent() == nil
	default:
		return false
	}
}

func (m *selectorMatcher) isFirstChild(node Node) bool {
	parent := node.GetParent()
	if parent == nil {
		return true
	}
	children := parent.GetChildren()
	for _, child := range children {
		if child.GetType() == ElementNodeType {
			return child == node
		}
	}
	return false
}

func (m *selectorMatcher) isLastChild(node Node) bool {
	parent := node.GetParent()
	if parent == nil {
		return true
	}
	children := parent.GetChildren()
	for i := len(children) - 1; i >= 0; i-- {
		if children[i].GetType() == ElementNodeType {
			return children[i] == node
		}
	}
	return false
}

func (m *selectorMatcher) isFirstOfType(node Node) bool {
	parent := node.GetParent()
	if parent == nil {
		return true
	}
	for _, child := range parent.GetChildren() {
		if child.GetType() == ElementNodeType && child.GetTag() == node.GetTag() {
			return child == node
		}
	}
	return false
}

func (m *selectorMatcher) isLastOfType(node Node) bool {
	parent := node.GetParent()
	if parent == nil {
		return true
	}
	children := parent.GetChildren()
	for i := len(children) - 1; i >= 0; i-- {
		if children[i].GetType() == ElementNodeType && children[i].GetTag() == node.GetTag() {
			return children[i] == node
		}
	}
	return false
}

func (m *selectorMatcher) isOnlyChild(node Node) bool {
	parent := node.GetParent()
	if parent == nil {
		return true
	}
	elementCount := 0
	for _, child := range parent.GetChildren() {
		if child.GetType() == ElementNodeType {
			elementCount++
		}
	}
	return elementCount == 1
}

func (m *selectorMatcher) isOnlyOfType(node Node) bool {
	parent := node.GetParent()
	if parent == nil {
		return true
	}
	typeCount := 0
	for _, child := range parent.GetChildren() {
		if child.GetType() == ElementNodeType && child.GetTag() == node.GetTag() {
			typeCount++
		}
	}
	return typeCount == 1
}

func (m *selectorMatcher) matchesComplexSelector(node Node, selector string) bool {
	parts := strings.Fields(selector)
	if len(parts) < 2 {
		return m.matchesSimpleSelector(node, selector)
	}

	targetSelector := parts[len(parts)-1]
	if !m.matchesSimpleSelector(node, targetSelector) {
		return false
	}

	return m.hasMatchingAncestor(node, parts[:len(parts)-1])
}

func (m *selectorMatcher) hasMatchingAncestor(node Node, ancestorSelectors []string) bool {
	current := node.GetParent()
	selectorIndex := len(ancestorSelectors) - 1

	for current != nil && selectorIndex >= 0 {
		if m.matchesSimpleSelector(current, ancestorSelectors[selectorIndex]) {
			selectorIndex--
			if selectorIndex < 0 {
				return true
			}
		}
		current = current.GetParent()
	}

	return selectorIndex < 0
}

func (m *selectorMatcher) matchesChildSelector(node Node, selector string) bool {
	parts := strings.Split(selector, ">")
	if len(parts) != 2 {
		return false
	}

	childSelector := strings.TrimSpace(parts[1])
	parentSelector := strings.TrimSpace(parts[0])

	if !m.matchesSimpleSelector(node, childSelector) {
		return false
	}

	parent := node.GetParent()
	return parent != nil && m.matchesSimpleSelector(parent, parentSelector)
}

func (m *selectorMatcher) matchesAdjacentSiblingSelector(node Node, selector string) bool {
	parts := strings.Split(selector, "+")
	if len(parts) != 2 {
		return false
	}

	targetSelector := strings.TrimSpace(parts[1])
	siblingSelector := strings.TrimSpace(parts[0])

	if !m.matchesSimpleSelector(node, targetSelector) {
		return false
	}

	return m.hasPreviousSiblingMatching(node, siblingSelector)
}

func (m *selectorMatcher) hasPreviousSiblingMatching(node Node, selector string) bool {
	parent := node.GetParent()
	if parent == nil {
		return false
	}

	siblings := parent.GetChildren()
	nodeIndex := m.findNodeIndexInSiblings(siblings, node)

	if nodeIndex <= 0 {
		return false
	}

	previousSibling := siblings[nodeIndex-1]
	return m.matchesSimpleSelector(previousSibling, selector)
}

func (m *selectorMatcher) findNodeIndexInSiblings(siblings []Node, target Node) int {
	for i, sibling := range siblings {
		if sibling == target {
			return i
		}
	}
	return -1
}

func (m *selectorMatcher) matchesGeneralSiblingSelector(node Node, selector string) bool {
	parts := strings.Split(selector, "~")
	if len(parts) != 2 {
		return false
	}

	targetSelector := strings.TrimSpace(parts[1])
	siblingSelector := strings.TrimSpace(parts[0])

	if !m.matchesSimpleSelector(node, targetSelector) {
		return false
	}

	return m.hasAnyPreviousSiblingMatching(node, siblingSelector)
}

func (m *selectorMatcher) hasAnyPreviousSiblingMatching(node Node, selector string) bool {
	parent := node.GetParent()
	if parent == nil {
		return false
	}

	siblings := parent.GetChildren()
	nodeIndex := m.findNodeIndexInSiblings(siblings, node)

	for i := 0; i < nodeIndex; i++ {
		if m.matchesSimpleSelector(siblings[i], selector) {
			return true
		}
	}

	return false
}
