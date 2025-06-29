package browser

import (
	"fmt"
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenTypeText TokenType = iota
	TokenTypeStartTag
	TokenTypeEndTag
	TokenTypeSelfClosingTag
	TokenTypeComment
	TokenTypeDoctype
	TokenTypeEOF
)

type Token struct {
	Type        TokenType
	Tag         string
	Attributes  map[string]string
	Text        string
	SelfClosing bool
}

type Tokenizer interface {
	NextToken() (*Token, error)
	HasMore() bool
	GetPosition() int
}

type attribute struct {
	name  string
	value string
}

type tokenizer struct {
	content string
	pos     int
	len     int
	current rune
}

func NewTokenizer(content string) Tokenizer {
	t := &tokenizer{
		content: content,
		pos:     0,
		len:     len(content),
	}
	if t.len > 0 {
		t.current = rune(content[0])
	}
	return t
}

func (t *tokenizer) HasMore() bool {
	return t.pos < t.len
}

func (t *tokenizer) GetPosition() int {
	return t.pos
}

func (t *tokenizer) NextToken() (*Token, error) {
	t.skipWhitespace()

	if !t.HasMore() {
		return &Token{Type: TokenTypeEOF}, nil
	}

	if t.current == '<' {
		return t.parseTag()
	}

	return t.parseText()
}

func (t *tokenizer) skipWhitespace() {
	for t.HasMore() && unicode.IsSpace(t.current) {
		t.advance()
	}
}

func (t *tokenizer) advance() {
	if t.pos < t.len-1 {
		t.pos++
		t.current = rune(t.content[t.pos])
	} else {
		t.pos = t.len
		t.current = 0
	}
}

func (t *tokenizer) parseTag() (*Token, error) {
	if !t.HasMore() || t.current != '<' {
		return nil, fmt.Errorf("expected '<' at position %d", t.pos)
	}

	t.advance()

	if t.current == '!' {
		return t.parseCommentOrDoctype()
	}

	isClosing := false
	if t.current == '/' {
		isClosing = true
		t.advance()
	}

	tagName := t.extractTagName()

	if tagName == "" {
		return &Token{Type: TokenTypeText, Text: t.content[:t.pos+1]}, nil
	}

	t.skipWhitespace()
	attributes := t.extractAttributes(isClosing)
	selfClosing := false
	if t.current == '/' {
		selfClosing = true
		t.advance()
	}

	if t.HasMore() && t.current != '>' {
		t.skipToChar('>')
	}
	if t.HasMore() {
		t.advance()
	}

	if isClosing {
		return &Token{
			Type: TokenTypeEndTag,
			Tag:  tagName,
		}, nil
	}

	if selfClosing {
		return &Token{
			Type:        TokenTypeSelfClosingTag,
			Tag:         tagName,
			Attributes:  attributes,
			SelfClosing: true,
		}, nil
	}

	return &Token{
		Type:       TokenTypeStartTag,
		Tag:        tagName,
		Attributes: attributes,
	}, nil
}

func (t *tokenizer) extractTagName() string {
	tagStart := t.pos
	for t.HasMore() && t.isValidTagChar() {
		t.advance()
	}
	return strings.ToLower(t.content[tagStart:t.pos])
}

func (t *tokenizer) isValidTagChar() bool {
	return !unicode.IsSpace(t.current) && t.current != '>' && t.current != '/'
}

func (t *tokenizer) extractAttributes(isClosing bool) map[string]string {
	if isClosing {
		return make(map[string]string)
	}

	attributes := make(map[string]string)
	for t.HasMore() && t.current != '>' && t.current != '/' {
		attr := t.parseSingleAttribute()
		if attr.name != "" {
			attributes[attr.name] = attr.value
		}
	}
	return attributes
}

func (t *tokenizer) parseSingleAttribute() attribute {
	t.skipWhitespace()
	if !t.HasMore() || t.current == '>' || t.current == '/' {
		return attribute{}
	}

	name := t.extractAttributeName()
	if name == "" {
		return attribute{}
	}

	t.skipWhitespace()
	value := t.extractAttributeValue()

	return attribute{name: strings.ToLower(name), value: value}
}

func (t *tokenizer) extractAttributeName() string {
	nameStart := t.pos
	for t.HasMore() && !unicode.IsSpace(t.current) && t.current != '=' && t.current != '>' && t.current != '/' {
		t.advance()
	}
	return t.content[nameStart:t.pos]
}

func (t *tokenizer) extractAttributeValue() string {
	if t.current != '=' {
		return ""
	}

	t.advance()
	t.skipWhitespace()

	if !t.HasMore() {
		return ""
	}

	if t.current == '"' || t.current == '\'' {
		return t.extractQuotedValue()
	}

	return t.extractUnquotedValue()
}

func (t *tokenizer) extractQuotedValue() string {
	quote := t.current
	t.advance()
	valueStart := t.pos

	for t.HasMore() && t.current != quote {
		t.advance()
	}

	value := t.content[valueStart:t.pos]
	if t.HasMore() {
		t.advance()
	}

	return value
}

func (t *tokenizer) extractUnquotedValue() string {
	valueStart := t.pos
	for t.HasMore() && !unicode.IsSpace(t.current) && t.current != '>' && t.current != '/' {
		t.advance()
	}
	return t.content[valueStart:t.pos]
}
func (t *tokenizer) parseCommentOrDoctype() (*Token, error) {
	if t.current != '!' {
		return nil, fmt.Errorf("expected '!' at position %d", t.pos)
	}

	t.advance()
	if !t.HasMore() {
		return &Token{Type: TokenTypeText, Text: "<!"}, nil
	}

	if t.pos+2 < t.len && t.content[t.pos:t.pos+2] == "--" {
		return t.parseComment()
	}

	if t.pos+7 < t.len && strings.ToUpper(t.content[t.pos:t.pos+7]) == "DOCTYPE" {
		return t.parseDoctype()
	}

	t.skipToChar('>')
	if t.HasMore() {
		t.advance()
	}
	return &Token{Type: TokenTypeComment, Text: ""}, nil
}

func (t *tokenizer) parseComment() (*Token, error) {
	t.pos += 2 // Skip "--"
	start := t.pos

	for t.pos+1 < t.len {
		if t.content[t.pos:t.pos+2] == "--" {
			content := t.content[start:t.pos]
			t.pos += 2
			t.skipToChar('>')
			if t.HasMore() {
				t.advance()
			}
			return &Token{Type: TokenTypeComment, Text: content}, nil
		}
		t.advance()
	}

	t.pos = t.len
	return &Token{Type: TokenTypeComment, Text: t.content[start:]}, nil
}

func (t *tokenizer) parseDoctype() (*Token, error) {
	start := t.pos
	t.skipToChar('>')
	content := t.content[start:t.pos]
	if t.HasMore() {
		t.advance()
	}
	return &Token{Type: TokenTypeDoctype, Text: content}, nil
}

func (t *tokenizer) parseText() (*Token, error) {
	start := t.pos
	for t.HasMore() && t.current != '<' {
		t.advance()
	}

	text := t.content[start:t.pos]
	if strings.TrimSpace(text) == "" && len(text) > 0 {
		return &Token{Type: TokenTypeText, Text: " "}, nil
	}

	return &Token{Type: TokenTypeText, Text: text}, nil
}

func (t *tokenizer) skipToChar(char rune) {
	for t.HasMore() && t.current != char {
		t.advance()
	}
}
