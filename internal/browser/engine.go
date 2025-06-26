package browser

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
)

type Engine interface {
	GetTabCount() int
	GetTab(idx int) Tab
	AddTab() Tab
	CloseTab(idx int)
	RefreshTab(idx int)
	FetchContent(ctx context.Context, tabIdx int, rawURL string) error
}

type engine struct {
	client *http.Client
	tabs   []Tab
	mutex  sync.RWMutex

	urlNormalizer URLNormalizer
}

func NewEngine() Engine {
	return &engine{
		client: &http.Client{
			Timeout: DefaultTimeoutSec,
			Transport: &http.Transport{
				MaxIdleConns:        MaxConcurrentConnections,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     DefaultTimeoutSec,
				DisableCompression:  false,
			},
		},
		tabs:          make([]Tab, 0),
		urlNormalizer: NewURLNormalizer(),
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

func (e *engine) CloseTab(idx int) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if idx < 0 || idx >= len(e.tabs) {
		return
	}

	e.tabs = append(e.tabs[:idx], e.tabs[idx+1:]...)
}

func (e *engine) RefreshTab(idx int) {
	tab := e.GetTab(idx)
	if tab == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeoutSec)
	defer cancel()
	_ = e.FetchContent(ctx, idx, tab.GetURL())
}

func (e *engine) FetchContent(ctx context.Context, tabIdx int, rawURL string) error {
	tab := e.GetTab(tabIdx)
	if tab == nil {
		return errors.New("tab cannot be nil")
	}

	normalizedURL, err := e.urlNormalizer.Normalize(rawURL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	content, err := e.fetchHTTPContent(ctx, normalizedURL)
	if err != nil {
		return err
	}

	tab.SetURL(normalizedURL)
	tab.SetContent(content)

	return nil
}

func (e *engine) fetchHTTPContent(ctx context.Context, urlStr string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", DefaultUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := e.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("%s: %w", ErrNetworkTimeout, err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("%s: HTTP %d", ErrHTTPError, resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
