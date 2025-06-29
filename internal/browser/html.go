package browser

import (
	"slices"
	"strings"
)

type HTMLSpecialTags interface {
	IsSelfClosing(tag string) bool
	IsBlockElement(tag string) bool
}

type htmlSpecialTags struct {
	selfClosing   map[string]bool
	blockElements map[string]bool
}

func NewHTMLSpecialTags() HTMLSpecialTags {
	return &htmlSpecialTags{
		selfClosing: map[string]bool{
			"area": true, "base": true, "br": true, "col": true,
			"embed": true, "hr": true, "img": true, "input": true,
			"link": true, "meta": true, "param": true, "source": true,
			"track": true, "wbr": true,
		},
		blockElements: map[string]bool{
			"html": true, "body": true, "article": true, "section": true,
			"nav": true, "aside": true, "h1": true, "h2": true, "h3": true,
			"h4": true, "h5": true, "h6": true, "hgroup": true, "header": true,
			"footer": true, "address": true, "p": true, "hr": true, "pre": true,
			"blockquote": true, "ol": true, "ul": true, "li": true, "dl": true,
			"dt": true, "dd": true, "figure": true, "figcaption": true,
			"main": true, "div": true, "table": true, "form": true,
			"fieldset": true, "legend": true, "details": true, "summary": true,
			"code": true, "button": true,
		},
	}
}

func (h *htmlSpecialTags) IsSelfClosing(tag string) bool {
	return h.selfClosing[tag]
}

func (h *htmlSpecialTags) IsBlockElement(tag string) bool {
	return h.blockElements[tag]
}

type ScriptInfo struct {
	Type    string
	Src     string
	Content string
	Async   bool
	Defer   bool
}

type HTMLParser interface {
	Parse() (Node, error)
	GetMetadata() map[string]string
	GetStyleTags() string
	GetScripts() []ScriptInfo
	PrintTree() string
}

type htmlParser struct {
	root        Node
	tokenizer   Tokenizer
	specialTags HTMLSpecialTags
	styleTags   strings.Builder
	metadata    map[string]string
	scripts     []ScriptInfo
}

func NewHTMLParser(html string) HTMLParser {
	return &htmlParser{
		tokenizer:   NewTokenizer(html),
		specialTags: NewHTMLSpecialTags(),
		metadata:    make(map[string]string),
		scripts:     make([]ScriptInfo, 0),
	}
}

type parserState struct {
	root          Node
	head          Node
	body          Node
	currentParent Node
	stack         []Node
	inHead        bool
}

func (p *htmlParser) Parse() (Node, error) {
	state := p.initializeParserState()

	for p.tokenizer.HasMore() {
		token, err := p.tokenizer.NextToken()
		if err != nil {
			continue
		}

		if token.Type == TokenTypeEOF {
			break
		}

		p.processToken(token, state)
	}

	return state.root, nil
}

func (p *htmlParser) initializeParserState() *parserState {
	root := NewElementNode("html", make(map[string]string))
	head := NewElementNode("head", make(map[string]string))
	body := NewElementNode("body", make(map[string]string))

	root.AddChild(head)
	root.AddChild(body)
	p.root = root

	return &parserState{
		root:          root,
		head:          head,
		body:          body,
		currentParent: body,
		stack:         []Node{root, body},
		inHead:        false,
	}
}

func (p *htmlParser) processToken(token *Token, state *parserState) {
	switch token.Type {
	case TokenTypeStartTag:
		p.handleStartTag(token, state)
	case TokenTypeSelfClosingTag:
		p.handleSelfClosingTag(token, state)
	case TokenTypeEndTag:
		p.handleEndTag(token, state)
	case TokenTypeText:
		p.handleTextToken(token, state)
	case TokenTypeComment:
		p.handleCommentToken(token, state)
	}
}

func (p *htmlParser) handleStartTag(token *Token, state *parserState) {
	node := NewElementNode(token.Tag, token.Attributes)

	if p.isSpecialStructuralTag(token.Tag) {
		p.handleStructuralTag(token, node, state)
		return
	}

	p.processSemanticTag(token.Tag, node, token.Attributes)

	if state.inHead {
		p.extractHeadData(token)
	}

	if p.shouldAddToDOM(token.Tag) {
		p.addNodeToDOM(node, token.Tag, state)
	}
}

func (p *htmlParser) isSpecialStructuralTag(tag string) bool {
	return tag == "head" || tag == "body"
}

func (p *htmlParser) handleStructuralTag(token *Token, node Node, state *parserState) {
	switch token.Tag {
	case "head":
		state.head = node
		state.root.GetChildren()[0] = state.head
		state.currentParent = state.head
		state.stack = []Node{state.root, state.head}
		state.inHead = true
	case "body":
		state.body = node
		p.updateBodyInDOM(node, state)
		state.currentParent = state.body
		state.stack = []Node{state.root, state.body}
		state.inHead = false
	}
}

func (p *htmlParser) updateBodyInDOM(node Node, state *parserState) {
	if len(state.root.GetChildren()) > 1 {
		state.root.GetChildren()[1] = node
	} else {
		state.root.AddChild(node)
	}
}

func (p *htmlParser) shouldAddToDOM(tag string) bool {
	return !slices.Contains([]string{"script", "style"}, tag)
}

func (p *htmlParser) addNodeToDOM(node Node, tag string, state *parserState) {
	if p.specialTags.IsSelfClosing(tag) {
		state.currentParent.AddChild(node)
	} else {
		state.currentParent.AddChild(node)
		state.stack = append(state.stack, node)
		state.currentParent = node
	}
}

func (p *htmlParser) handleSelfClosingTag(token *Token, state *parserState) {
	node := NewElementNode(token.Tag, token.Attributes)
	if state.inHead && token.Tag == "meta" {
		p.extractMetadata(node)
	}
	state.currentParent.AddChild(node)
}

func (p *htmlParser) handleEndTag(token *Token, state *parserState) {
	for i := len(state.stack) - 1; i >= 0; i-- {
		if p.isMatchingElement(state.stack[i], token.Tag) {
			state.stack = state.stack[:i+1]
			p.updateCurrentParent(state)
			p.handleSpecialEndTag(token.Tag, state)
			break
		}
	}
}

func (p *htmlParser) isMatchingElement(node Node, tag string) bool {
	return node.GetType() == ElementNodeType && node.GetTag() == tag
}

func (p *htmlParser) updateCurrentParent(state *parserState) {
	if len(state.stack) > 1 {
		state.currentParent = state.stack[len(state.stack)-2]
	}
}

func (p *htmlParser) handleSpecialEndTag(tag string, state *parserState) {
	if tag == "head" {
		state.inHead = false
		state.currentParent = state.body
		state.stack = []Node{state.root, state.body}
	}
}

func (p *htmlParser) handleTextToken(token *Token, state *parserState) {
	if strings.TrimSpace(token.Text) != "" {
		textNode := NewTextNode(token.Text)
		state.currentParent.AddChild(textNode)
	}
}

func (p *htmlParser) handleCommentToken(token *Token, state *parserState) {
	commentNode := NewCommentNode(token.Text)
	state.currentParent.AddChild(commentNode)
}

func (p *htmlParser) extractHeadData(token *Token) {
	switch token.Tag {
	case "style":
		content := p.extractTagContent("style")
		p.styleTags.WriteString(content)
		p.styleTags.WriteString("\n")
	case "script":
		p.extractScriptFromToken(token)
	case "title":
		content := p.extractTagContent("title")
		p.metadata["title"] = content
	case "meta":
		node := NewElementNode(token.Tag, token.Attributes)
		p.extractMetadata(node)
	}
}

func (p *htmlParser) extractMetadata(node Node) {
	attrs := node.GetAttributes()

	if charset, ok := attrs["charset"]; ok {
		p.metadata["charset"] = charset
	}

	if name, ok := attrs["name"]; ok {
		if content, ok := attrs["content"]; ok {
			p.metadata[name] = content
		}
	}

	if httpEquiv, ok := attrs["http-equiv"]; ok {
		if content, ok := attrs["content"]; ok {
			p.metadata["http-equiv-"+httpEquiv] = content
		}
	}

	if property, ok := attrs["property"]; ok {
		if content, ok := attrs["content"]; ok {
			p.metadata[property] = content
		}
	}
}

func (p *htmlParser) extractScriptFromToken(token *Token) {
	attrs := token.Attributes
	script := ScriptInfo{
		Type:  attrs["type"],
		Src:   attrs["src"],
		Async: attrs["async"] != "",
		Defer: attrs["defer"] != "",
	}

	if script.Src == "" {
		script.Content = p.extractTagContent("script")
	}

	p.scripts = append(p.scripts, script)
}

func (p *htmlParser) extractTagContent(tag string) string {
	content := strings.Builder{}
	depth := 1

	for p.tokenizer.HasMore() && depth > 0 {
		token, err := p.tokenizer.NextToken()
		if err != nil {
			break
		}

		if token.Tag == tag {
			if token.Type == TokenTypeStartTag {
				depth++
			}

			if token.Type == TokenTypeEndTag {
				depth--
			}
		}

		if depth > 0 && token.Type == TokenTypeText {
			content.WriteString(token.Text)
		}
	}

	return strings.TrimSpace(content.String())
}

func (p *htmlParser) GetMetadata() map[string]string {
	return p.metadata
}

func (p *htmlParser) GetStyleTags() string {
	return p.styleTags.String()
}

func (p *htmlParser) GetScripts() []ScriptInfo {
	return p.scripts
}

func (p *htmlParser) PrintTree() string {
	if p.root == nil {
		return "No DOM tree available\n"
	}
	return p.printNodeRecursive(p.root, 0)
}

func (p *htmlParser) printNodeRecursive(node Node, depth int) string {
	if node == nil {
		return ""
	}

	indent := strings.Repeat("-", depth)
	result := indent + node.String()

	for _, child := range node.GetChildren() {
		result += p.printNodeRecursive(child, depth+1)
	}

	return result
}

func (p *htmlParser) processSemanticTag(tag string, node Node, attributes map[string]string) {
	switch tag {
	case "a":
		if href, exists := attributes["href"]; exists {
			node.SetAttribute("href", href)
			node.SetAttribute("role", "link")
		}
	case "nav":
		node.SetAttribute("role", "navigation")
	case "button":
		node.SetAttribute("role", "button")
		if buttonType, exists := attributes["type"]; !exists {
			node.SetAttribute("type", "button")
		} else {
			node.SetAttribute("type", buttonType)
		}
	case "pre", "code":
		node.SetAttribute("whitespace", "pre")
		if lang, exists := attributes["class"]; exists && strings.HasPrefix(lang, "language-") {
			node.SetAttribute("syntax-highlight", strings.TrimPrefix(lang, "language-"))
		}
	case "ul", "ol":
		node.SetAttribute("role", "list")
	case "li":
		node.SetAttribute("role", "listitem")
	}
}
