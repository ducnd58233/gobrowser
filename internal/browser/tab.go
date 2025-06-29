package browser

import (
	"strings"
)

type page struct {
	url  string
	prev *page
	next *page
}

type Tab interface {
	GetID() string
	GetTitle() string
	SetTitle(title string)
	GetURL() string
	SetURL(url string)
	GetDocument() Document
	SetDocument(doc Document)
	Navigate(url string)
	CanGoBack() bool
	GoBack()
	CanGoNext() bool
	GoNext()
}

type tab struct {
	id       string
	title    string
	document Document
	history  *page
	loading  bool
}

func NewTab() Tab {
	return &tab{
		id:      NewIDGenerator().Generate(),
		title:   "New Tab",
		history: nil,
		loading: false,
	}
}

func (t *tab) GetID() string {
	return t.id
}

func (t *tab) GetTitle() string {
	if t.title != "" {
		return t.title
	}

	return "New Tab"
}

func (t *tab) SetTitle(title string) {
	t.title = title
}

func (t *tab) GetURL() string {
	if t.history == nil {
		return ""
	}
	return t.history.url
}

func (t *tab) SetURL(url string) {
	if t.history == nil {
		t.history = &page{url: url}
		return
	}
	t.history.url = url
}

func (t *tab) GetDocument() Document {
	return t.document
}

func (t *tab) SetDocument(doc Document) {
	t.document = doc
	if doc != nil && doc.GetTitle() != "" {
		t.title = doc.GetTitle()
	}
}

func (t *tab) Navigate(url string) {
	if strings.TrimSpace(url) == "" {
		return
	}
	t.addToHistory(url)
	t.loading = true
}

func (t *tab) CanGoBack() bool {
	return t.history != nil && t.history.prev != nil
}

func (t *tab) GoBack() {
	if !t.CanGoBack() {
		return
	}

	t.history = t.history.prev
}

func (t *tab) CanGoNext() bool {
	return t.history != nil && t.history.next != nil
}

func (t *tab) GoNext() {
	if !t.CanGoNext() {
		return
	}

	t.history = t.history.next
}

func (t *tab) addToHistory(url string) {
	newPage := &page{url: url}

	if t.history == nil {
		t.history = newPage
		return
	}

	t.history.next = nil
	newPage.prev = t.history
	t.history.next = newPage
	t.history = t.history.next
}
