package browser

import "strings"

type Tab interface {
	Navigate(url string)
	CanGoBack() bool
	GoBack()
	CanGoForward() bool
	GoForward()
	GetTitle() string
	SetTitle(title string)
	GetContent() string
	SetContent(content string)
	GetURL() string
	SetURL(url string)
	GetID() string
}

type page struct {
	url  string
	prev *page
	next *page
}

type tab struct {
	id      string
	title   string
	content string
	url     string
	page    *page
	loading bool
}

func NewTab() Tab {
	return &tab{
		id:      generateID(),
		title:   DefaultTitle,
		content: "",
		url:     DefaultURL,
		page:    nil,
		loading: false,
	}
}

func (t *tab) Navigate(url string) {
	t.addTopage(url)
	t.url = url
	t.loading = true
}

func (t *tab) CanGoBack() bool {
	return t.page != nil && t.page.prev != nil
}

func (t *tab) GoBack() {
	if t.CanGoBack() {
		t.page = t.page.prev
		t.url = t.page.url
	}
}

func (t *tab) CanGoForward() bool {
	return t.page != nil && t.page.next != nil
}

func (t *tab) GoForward() {
	if t.CanGoForward() {
		t.page = t.page.next
		t.url = t.page.url
	}
}

func (t *tab) GetTitle() string {
	if t.loading {
		return "Loading..."
	}
	if strings.TrimSpace(t.title) == "" {
		return DefaultTitle
	}
	return t.title
}

func (t *tab) SetTitle(title string) {
	t.title = title
}

func (t *tab) GetContent() string {
	return t.content
}

func (t *tab) SetContent(content string) {
	t.content = content
	t.loading = false
}

func (t *tab) GetURL() string {
	if strings.TrimSpace(t.url) == "" {
		return DefaultURL
	}
	return t.url
}

func (t *tab) SetURL(url string) {
	t.url = url
}

func (t *tab) GetID() string {
	return t.id
}

func (t *tab) addTopage(url string) {
	newPage := &page{url: url}

	if t.page == nil {
		t.page = newPage
		return
	}

	t.page.next = nil
	newPage.prev = t.page
	t.page.next = newPage
	t.page = newPage
}
