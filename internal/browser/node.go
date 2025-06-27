package browser

import (
	"encoding/json"
	"fmt"
	"strings"
)

type NodeType int

const (
	TextNodeType NodeType = iota
	ElementNodeType
	CommentNodeType
)

type NodeSearcher interface {
	FindElementsByTag(tag string) []Node
	FindElementsByClass(className string) []Node
	FindElementByID(id string) Node
	GetTextContent() string
}

type Node interface {
	GetID() string
	GetType() NodeType
	GetTag() string
	GetText() string
	GetAttributes() map[string]string
	GetAttribute(key string) (string, bool)
	SetAttribute(key, value string)
	GetParent() Node
	SetParent(Node)
	HasClass(className string) bool

	GetChildren() []Node
	AddChild(Node)

	String() string
}

type elementNode struct {
	tag        string
	attributes map[string]string
	children   []Node
	parent     Node
}

func NewElementNode(tag string, attributes map[string]string) Node {
	if attributes == nil {
		attributes = make(map[string]string)
	}
	return &elementNode{
		tag:        strings.ToLower(tag),
		attributes: attributes,
		children:   make([]Node, 0),
	}
}

func (e *elementNode) GetID() string                    { return e.attributes["id"] }
func (e *elementNode) GetType() NodeType                { return ElementNodeType }
func (e *elementNode) GetTag() string                   { return e.tag }
func (e *elementNode) GetText() string                  { return "" }
func (e *elementNode) GetAttributes() map[string]string { return e.attributes }
func (e *elementNode) GetAttribute(key string) (string, bool) {
	value, exists := e.attributes[strings.ToLower(key)]
	return value, exists
}
func (e *elementNode) SetAttribute(key, value string) {
	e.attributes[strings.ToLower(key)] = value
}
func (e *elementNode) GetParent() Node       { return e.parent }
func (e *elementNode) SetParent(parent Node) { e.parent = parent }
func (e *elementNode) HasClass(className string) bool {
	class, exists := e.attributes["class"]
	if !exists {
		return false
	}

	classes := strings.Fields(class)
	for _, c := range classes {
		if c == className {
			return true
		}
	}
	return false
}
func (e *elementNode) GetChildren() []Node { return e.children }
func (e *elementNode) AddChild(child Node) {
	if child == nil {
		return
	}
	e.children = append(e.children, child)
	child.SetParent(e)
}
func (e *elementNode) String() string {
	attrs, _ := json.Marshal(e.attributes)
	return fmt.Sprintf("ElementNode(tag=%s, attributes=%v)\n", e.tag, string(attrs))
}

func (e *elementNode) FindElementByTag(tag string) []Node {
	var elements []Node
	if e.tag == tag {
		elements = append(elements, e)
	}
	for _, child := range e.children {
		if searcher, ok := child.(NodeSearcher); ok {
			elements = append(elements, searcher.FindElementsByTag(tag)...)
		}
	}
	return elements
}
func (e *elementNode) FindElementsByClass(className string) []Node {
	var elements []Node
	if e.HasClass(className) {
		elements = append(elements, e)
	}
	for _, child := range e.children {
		if searcher, ok := child.(NodeSearcher); ok {
			elements = append(elements, searcher.FindElementsByClass(className)...)
		}
	}
	return elements
}
func (e *elementNode) FindElementByID(id string) Node {
	if e.GetID() == id {
		return e
	}
	for _, child := range e.children {
		if searcher, ok := child.(NodeSearcher); ok {
			if result := searcher.FindElementByID(id); result != nil {
				return result
			}
		}
	}
	return nil
}
func (e *elementNode) GetTextContent() string {
	var text strings.Builder
	for _, child := range e.children {
		if traverser, ok := child.(NodeSearcher); ok {
			text.WriteString(traverser.GetTextContent())
		}
	}
	return text.String()
}

type textNode struct {
	content string
	parent  Node
}

func NewTextNode(content string) Node {
	return &textNode{content: content}
}

func (t *textNode) GetID() string                          { return "" }
func (t *textNode) GetType() NodeType                      { return TextNodeType }
func (t *textNode) GetTag() string                         { return "" }
func (t *textNode) GetText() string                        { return t.content }
func (t *textNode) GetAttributes() map[string]string       { return nil }
func (t *textNode) GetAttribute(key string) (string, bool) { return "", false }
func (t *textNode) SetAttribute(key, value string) {
	// No action
}
func (t *textNode) GetParent() Node                { return t.parent }
func (t *textNode) SetParent(parent Node)          { t.parent = parent }
func (t *textNode) HasClass(className string) bool { return false }
func (t *textNode) GetChildren() []Node            { return nil }
func (t *textNode) AddChild(child Node) {
	// No action
}
func (t *textNode) String() string {
	return fmt.Sprintf("TextNode(content=\"%s\")\n", t.content)
}

func (t *textNode) FindElementsByTag(tag string) []Node         { return nil }
func (t *textNode) FindElementsByClass(className string) []Node { return nil }
func (t *textNode) FindElementByID(id string) Node              { return nil }
func (t *textNode) GetTextContent() string                      { return t.content }

type commentNode struct {
	content string
	parent  Node
}

func NewCommentNode(content string) Node {
	return &commentNode{content: content}
}

func (c *commentNode) GetID() string                          { return "" }
func (c *commentNode) GetType() NodeType                      { return CommentNodeType }
func (c *commentNode) GetTag() string                         { return "" }
func (c *commentNode) GetText() string                        { return c.content }
func (c *commentNode) GetAttributes() map[string]string       { return nil }
func (c *commentNode) GetAttribute(key string) (string, bool) { return "", false }
func (c *commentNode) SetAttribute(key, value string) {
	// No action
}
func (c *commentNode) GetParent() Node                { return c.parent }
func (c *commentNode) SetParent(parent Node)          { c.parent = parent }
func (c *commentNode) HasClass(className string) bool { return false }
func (c *commentNode) GetChildren() []Node            { return nil }
func (c *commentNode) AddChild(child Node) {
	// No action
}
func (c *commentNode) String() string {
	return fmt.Sprintf("CommentNode(content=\"%s\")\n", c.content)
}

func (c *commentNode) FindElementsByTag(tag string) []Node         { return nil }
func (c *commentNode) FindElementsByClass(className string) []Node { return nil }
func (c *commentNode) FindElementByID(id string) Node              { return nil }
func (c *commentNode) GetTextContent() string                      { return c.content }
