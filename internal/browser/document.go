package browser

import (
	"log"
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

type DocumentBuilder interface {
	Build(content string) (Document, error)
	SetDebugMode(enabled bool)
}

type documentBuilder struct {
	htmlParser    HTMLParser
	cssParser     CSSParser
	cssApplicator CSSApplicator
	debugMode     bool
}

// NewDocumentBuilder creates a new document builder instance
func NewDocumentBuilder() DocumentBuilder {
	return &documentBuilder{
		cssApplicator: NewCSSApplicator(),
		debugMode:     false,
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

	defaultCSS := db.getDefaultCSS()
	fullCSS := defaultCSS + "\n" + styleContent

	db.cssParser = NewCSSParser(fullCSS)
	css := db.cssParser.Parse()

	if db.debugMode {
		log.Println("CSS Parser Output:")
		log.Println(css.PrintTree())
	}

	doc.stylesheet = css
	return nil
}

func (db *documentBuilder) applyStyles(doc *document) error {
	if doc.stylesheet == nil || doc.root == nil {
		return nil
	}

	db.cssApplicator.ApplyCSS(doc.root, doc.stylesheet, doc.styles)
	return nil
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
}

h1, h2, h3, h4, h5, h6 {
	font-weight: bold;
	margin: 0.5em 0;
}

h1 { font-size: 2em; }
h2 { font-size: 1.5em; }
h3 { font-size: 1.17em; }
h4 { font-size: 1em; }
h5 { font-size: 0.83em; }
h6 { font-size: 0.75em; }

p {
	margin: 1em 0;
}

div, section, article, aside, nav, main, header, footer {
	display: block;
}

a {
	color: #0000EE;
	text-decoration: underline;
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
}

li {
	margin: 0.5em 0;
}

table {
	border-collapse: collapse;
}

th, td {
	padding: 4px;
	border: 1px solid #ccc;
}

img {
	max-width: 100%;
	height: auto;
}
`
}
