package browser

import (
	"context"
	"log"
	"strings"
)

type Document interface {
	GetRoot() Node
	GetTitle() string
	GetCharset() string
	GetLanguage() string
	GetMetadata() map[string]string
	GetStyleSheet() *CSS
	GetScripts() []ScriptInfo
	GetComputedStyle(node Node) Style
	SetComputedStyle(node Node, style Style)
}

type document struct {
	root       Node
	title      string
	charset    string
	language   string
	metadata   map[string]string
	stylesheet *CSS
	scripts    []ScriptInfo
	styles     map[Node]Style
}

func (d *document) GetRoot() Node                  { return d.root }
func (d *document) GetTitle() string               { return d.title }
func (d *document) GetCharset() string             { return d.charset }
func (d *document) GetLanguage() string            { return d.language }
func (d *document) GetMetadata() map[string]string { return d.metadata }
func (d *document) GetStyleSheet() *CSS            { return d.stylesheet }
func (d *document) GetScripts() []ScriptInfo       { return d.scripts }

func (d *document) GetComputedStyle(node Node) Style {
	if style, ok := d.styles[node]; ok {
		return style
	}
	return NewStyle()
}

func (d *document) SetComputedStyle(node Node, style Style) {
	d.styles[node] = style
}

type DocumentBuilder interface {
	Build(content string) (Document, error)
	SetDebugMode(enabled bool)
}

type documentBuilder struct {
	apiHandler    APIHandler
	htmlParser    HTMLParser
	cssParser     CSSParser
	cssApplicator CSSApplicator
	debugMode     bool
}

func NewDocumentBuilder() DocumentBuilder {
	return &documentBuilder{
		cssApplicator: NewCSSApplicator(),
		debugMode:     false,
		apiHandler:    NewAPIHandler(),
	}
}

func (db *documentBuilder) SetDebugMode(enabled bool) {
	db.debugMode = enabled
}

func (db *documentBuilder) Build(content string) (Document, error) {
	if content == "" {
		return nil, NewBrowserError(ErrInvalidInput, "content cannot be empty")
	}

	doc := &document{
		metadata: make(map[string]string),
		styles:   make(map[Node]Style),
	}

	if err := db.parseHTML(content, doc); err != nil {
		return nil, err
	}

	if err := db.parseCSS(doc); err != nil {
		return nil, err
	}

	if err := db.applyStyles(doc); err != nil {
		return nil, err
	}

	return doc, nil
}

func (db *documentBuilder) parseHTML(content string, doc *document) error {
	db.htmlParser = NewHTMLParser(content)

	root, err := db.htmlParser.Parse()
	if err != nil {
		return NewBrowserError(ErrParsingFailed, "failed to parse HTML: "+err.Error())
	}

	doc.root = root
	doc.metadata = db.htmlParser.GetMetadata()
	doc.scripts = db.htmlParser.GetScripts()

	if title, ok := doc.metadata["title"]; ok {
		doc.title = title
	}
	if charset, ok := doc.metadata["charset"]; ok {
		doc.charset = charset
	}
	if lang, ok := doc.metadata["lang"]; ok {
		doc.language = lang
	}

	if db.debugMode {
		log.Println("HTML Parser Output:")
		log.Println(db.htmlParser.PrintTree())
	}

	return nil
}

// parseCSS extracts and parses CSS from style tags and inline styles
func (db *documentBuilder) parseCSS(doc *document) error {
	styleContent := db.htmlParser.GetStyleTags()
	stylesheetURLs := db.htmlParser.GetStylesheetLinks()

	externalCSS := db.fetchExternalStylesheets(stylesheetURLs)

	defaultCSS := db.getDefaultCSS()
	fullCSS := defaultCSS + "\n" + styleContent + "\n" + externalCSS

	db.cssParser = NewCSSParser(fullCSS)
	css := db.cssParser.Parse()

	if db.debugMode {
		log.Println("CSS Parser Output:")
		log.Println(css.PrintTree())
	}

	doc.stylesheet = css
	return nil
}

func (db *documentBuilder) fetchExternalStylesheets(urls []string) string {
	if len(urls) == 0 {
		return ""
	}

	var combinedCSS strings.Builder
	cssChannel := make(chan string, len(urls))

	for _, url := range urls {
		go db.fetchStylesheetAsync(url, cssChannel)
	}

	for i := 0; i < len(urls); i++ {
		css := <-cssChannel
		if css != "" {
			combinedCSS.WriteString(css)
			combinedCSS.WriteString("\n")
		}
	}

	return combinedCSS.String()
}

func (db *documentBuilder) fetchStylesheetAsync(url string, result chan<- string) {
	defer func() {
		if r := recover(); r != nil {
			result <- ""
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	content, err := db.apiHandler.FetchContent(ctx, url)
	if err != nil {
		if db.debugMode {
			log.Printf("Failed to fetch stylesheet %s: %v", url, err)
		}
		result <- ""
		return
	}

	result <- content
}

func (db *documentBuilder) applyStyles(doc *document) error {
	if doc.stylesheet == nil || doc.root == nil {
		return nil
	}

	db.applyStylesToTree(doc, doc.root, doc.stylesheet)
	return nil
}

func (db *documentBuilder) applyStylesToTree(doc *document, node Node, css *CSS) {
	if node == nil {
		return
	}

	computedStyle := db.cssApplicator.ComputeStyle(node, css)

	style := db.convertComputedStyleToStyle(computedStyle)
	doc.SetComputedStyle(node, style)

	for _, child := range node.GetChildren() {
		db.applyStylesToTree(doc, child, css)
	}
}

func (db *documentBuilder) convertComputedStyleToStyle(computedStyle *ComputedStyle) Style {
	style := NewStyle()

	for prop, value := range computedStyle.Properties {
		style.SetProperty(prop, CSSValue{
			Raw:       value,
			ValueType: CSSValueKeyword,
		})
	}

	return style
}

func (db *documentBuilder) getDefaultCSS() string {
	return `
/* Default browser styles */
html, body {
	margin: 0;
	padding: 0;
	font-family: sans-serif;
	font-size: 16px;
	line-height: 1.2;
	color: #000000;
	background-color: #ffffff;
	display: block;
}

h1, h2, h3, h4, h5, h6 {
	font-weight: bold;
	margin: 0.5em 0;
	display: block;
}

h1 { 
	font-size: 2em; 
	margin: 0.67em 0;
}
h2 { 
	font-size: 1.5em; 
	margin: 0.75em 0;
}
h3 { 
	font-size: 1.17em; 
	margin: 0.83em 0;
}
h4 { 
	font-size: 1em; 
	margin: 1.12em 0;
}
h5 { 
	font-size: 0.83em; 
	margin: 1.5em 0;
}
h6 { 
	font-size: 0.75em; 
	margin: 1.67em 0;
}

p {
	margin: 1em 0;
	display: block;
}

div, section, article, aside, nav, main, header, footer, blockquote {
	display: block;
	margin: 0;
	padding: 0;
}

pre {
	display: block;
	white-space: pre;
	font-family: monospace;
	margin: 1em 0;
	background-color: #f5f5f5;
	padding: 0.5em;
}

a {
	color: #0000EE;
	text-decoration: underline;
}

a:visited {
	color: #551A8B;
}

strong, b {
	font-weight: bold;
}

em, i {
	font-style: italic;
}

ul, ol {
	margin: 1em 0;
	padding-left: 2em;
	display: block;
}

li {
	margin: 0;
	padding: 0;
	display: list-item;
}

br {
	display: inline;
}

span, code, small, big {
	display: inline;
}

/* Button and form elements */
button {
	display: inline-block;
	padding: 0.25em 0.5em;
	margin: 0;
	border: 1px solid #ccc;
	background-color: #f0f0f0;
	cursor: pointer;
}

input {
	display: inline-block;
	padding: 0.25em;
	margin: 0;
	border: 1px solid #ccc;
}

/* Table elements */
table {
	display: table;
	border-collapse: separate;
	border-spacing: 2px;
}

tr {
	display: table-row;
}

td, th {
	display: table-cell;
	padding: 2px;
}

th {
	font-weight: bold;
	text-align: center;
}
`
}
