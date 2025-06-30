package browser

import (
	"html"
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
	GetStylesheetLinks() []string
	PrintTree() string
}

type htmlParser struct {
	root           Node
	tokenizer      Tokenizer
	specialTags    HTMLSpecialTags
	styleTags      strings.Builder
	metadata       map[string]string
	scripts        []ScriptInfo
	stylesheetURLs []string
	stack          []Node
}

func NewHTMLParser(html string) HTMLParser {
	return &htmlParser{
		tokenizer:      NewTokenizer(html),
		specialTags:    NewHTMLSpecialTags(),
		metadata:       make(map[string]string),
		scripts:        make([]ScriptInfo, 0),
		stylesheetURLs: make([]string, 0),
		stack:          make([]Node, 0),
	}
}

func (p *htmlParser) Parse() (Node, error) {
	root := NewElementNode("html", make(map[string]string))
	head := NewElementNode("head", make(map[string]string))
	body := NewElementNode("body", make(map[string]string))
	root.AddChild(head)
	root.AddChild(body)
	p.root = root
	p.stack = []Node{root, body}

	for p.tokenizer.HasMore() {
		token, err := p.tokenizer.NextToken()
		if err != nil {
			continue
		}
		if token.Type == TokenTypeEOF {
			break
		}
		p.processToken(token)
	}
	return root, nil
}

func (p *htmlParser) processToken(token *Token) {
	switch token.Type {
	case TokenTypeStartTag:
		p.handleStartTag(token)
	case TokenTypeSelfClosingTag:
		p.handleSelfClosingTag(token)
	case TokenTypeEndTag:
		p.handleEndTag(token)
	case TokenTypeText:
		p.handleTextToken(token)
	case TokenTypeComment:
		p.handleCommentToken(token)
	}
}

func (p *htmlParser) currentParent() Node {
	if len(p.stack) == 0 {
		return nil
	}
	return p.stack[len(p.stack)-1]
}

func (p *htmlParser) handleStartTag(token *Token) {
	node := NewElementNode(token.Tag, token.Attributes)
	parent := p.currentParent()
	if parent != nil {
		parent.AddChild(node)
	}
	p.stack = append(p.stack, node)
	if token.Tag == "style" {
		content := p.extractTagContent("style")
		p.styleTags.WriteString(content)
		p.styleTags.WriteString("\n")
	}
	if token.Tag == "link" {
		p.processLinkTag(token)
	}
	if token.Tag == "script" {
		p.extractScriptFromToken(token)
	}
	if token.Tag == "title" {
		content := p.extractTagContent("title")
		p.metadata["title"] = content
	}
	if token.Tag == "meta" {
		p.extractMetadata(node)
	}
}

func (p *htmlParser) handleEndTag(token *Token) {
	for len(p.stack) > 1 {
		top := p.stack[len(p.stack)-1]
		if top.GetTag() == token.Tag {
			p.stack = p.stack[:len(p.stack)-1]
			return
		}
		p.stack = p.stack[:len(p.stack)-1]
	}
}

func (p *htmlParser) handleSelfClosingTag(token *Token) {
	node := NewElementNode(token.Tag, token.Attributes)
	parent := p.currentParent()
	if parent != nil {
		parent.AddChild(node)
	}
	if token.Tag == "meta" {
		p.extractMetadata(node)
	}
}

func (p *htmlParser) handleTextToken(token *Token) {
	if token.Text == "" {
		return
	}
	parent := p.currentParent()
	if parent == nil {
		return
	}
	decodedText := html.UnescapeString(token.Text)
	if p.isInPre() {
		decodedText = p.preservePreformattedText(decodedText)
	}
	textNode := NewTextNode(decodedText)
	parent.AddChild(textNode)
}

func (p *htmlParser) isInPre() bool {
	for i := len(p.stack) - 1; i >= 0; i-- {
		n := p.stack[i]
		if n.GetTag() == "pre" {
			return true
		}
	}
	return false
}

func (p *htmlParser) preservePreformattedText(text string) string {
	return strings.ReplaceAll(text, "\t", "    ")
}

func (p *htmlParser) handleCommentToken(token *Token) {
	parent := p.currentParent()
	if parent == nil {
		return
	}
	commentNode := NewCommentNode(token.Text)
	parent.AddChild(commentNode)
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

func (p *htmlParser) GetStylesheetLinks() []string {
	return p.stylesheetURLs
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

func (p *htmlParser) processLinkTag(token *Token) {
	attrs := token.Attributes
	rel, hasRel := attrs["rel"]
	href, hasHref := attrs["href"]

	if hasRel && hasHref && rel == "stylesheet" {
		p.stylesheetURLs = append(p.stylesheetURLs, href)
	}
}
