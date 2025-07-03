package browser

import (
	"context"
	"sync"
)

type Engine interface {
	GetTabCount() int
	GetTab(idx int) Tab
	AddTab() Tab
	CloseTab(idx int) error
	RefreshTab(idx int) error
	Navigate(ctx context.Context, tabIdx int, url string) error
	GetURLHandler() URLHandler
	SetDebugMode(enabled bool)
	GetDebugMode() bool
}

type engine struct {
	tabs  []Tab
	mutex sync.RWMutex

	apiHandler      APIHandler
	documentBuilder DocumentBuilder
	urlHandler      URLHandler

	debugMode      bool
	isShuttingDown bool
}

func NewEngine() Engine {
	return &engine{
		tabs:            make([]Tab, 0),
		apiHandler:      NewAPIHandler(),
		documentBuilder: NewDocumentBuilder(),
		urlHandler:      NewURLHandler(),
		debugMode:       false,
		isShuttingDown:  false,
	}
}
func (e *engine) GetTabCount() int {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	return len(e.tabs)
}

func (e *engine) GetTab(idx int) Tab {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	if idx < 0 || idx >= len(e.tabs) {
		return nil
	}

	return e.tabs[idx]
}

func (e *engine) AddTab() Tab {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	tab := NewTab()
	e.tabs = append(e.tabs, tab)

	return tab
}

func (e *engine) CloseTab(idx int) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if idx < 0 || idx >= len(e.tabs) {
		return NewBrowserError(ErrInvalidInput, "invalid tab index")
	}

	e.tabs = append(e.tabs[:idx], e.tabs[idx+1:]...)
	return nil
}

func (e *engine) RefreshTab(idx int) error {
	tab := e.GetTab(idx)
	if tab == nil {
		return NewBrowserError(ErrInvalidInput, "invalid tab index")
	}

	return e.fetchContentForTab(context.Background(), idx, tab.GetURL())
}

func (e *engine) Navigate(ctx context.Context, tabIdx int, rawURL string) error {
	tab := e.GetTab(tabIdx)
	if tab == nil {
		return NewBrowserError(ErrInvalidInput, "invalid tab index")
	}

	normalizedURL, err := e.urlHandler.Normalize(rawURL)
	if err != nil {
		return err
	}

	tab.Navigate(normalizedURL)
	return e.fetchContentForTab(ctx, tabIdx, normalizedURL)
}

func (e *engine) fetchContentForTab(ctx context.Context, tabIdx int, normalizedURL string) error {
	tab := e.GetTab(tabIdx)
	if tab == nil {
		return NewBrowserError(ErrInvalidInput, "invalid tab index")
	}

	content, err := e.apiHandler.FetchContent(ctx, normalizedURL)
	if err != nil {
		return NewBrowserError(ErrNetworkTimeout, err.Error())
	}

	e.documentBuilder.SetBaseURL(normalizedURL)

	doc, err := e.documentBuilder.Build(content)
	if err != nil {
		return NewBrowserError(ErrParsingFailed, "failed to build document: "+err.Error())
	}

	tab.SetURL(normalizedURL)
	tab.SetDocument(doc)

	return nil
}

func (e *engine) GetURLHandler() URLHandler {
	return e.urlHandler
}

func (e *engine) SetDebugMode(enabled bool) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.debugMode = enabled
	e.documentBuilder.SetDebugMode(enabled)
}

func (e *engine) GetDebugMode() bool {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	return e.debugMode
}
