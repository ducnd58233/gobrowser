package browser

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Engine interface {
	GetTab(index int) Tab
	GetAllTabs() []Tab
	GetCurrentTab() Tab
	SetCurrentTab(index int)
	RefreshCurrentTab()
	AddTab() Tab
	CloseTab(index int)
	FetchContent(ctx context.Context, tab Tab, rawURL string) error
	TabCount() int
}

type engine struct {
	client  *http.Client
	tabs    []Tab
	current int
}

func NewEngine() Engine {
	return &engine{
		client: &http.Client{
			Timeout: DefaultTimeoutSec * time.Second,
		},
		tabs:    make([]Tab, 0),
		current: -1,
	}
}

func (e *engine) GetTab(index int) Tab {
	if index < 0 || index >= len(e.tabs) {
		return nil
	}
	return e.tabs[index]
}

func (e *engine) GetAllTabs() []Tab {
	return e.tabs
}

func (e *engine) GetCurrentTab() Tab {
	if e.current < 0 || e.current >= len(e.tabs) {
		return nil
	}
	return e.tabs[e.current]
}

func (e *engine) RefreshCurrentTab() {
	tab := e.GetCurrentTab()
	if tab != nil {
		e.FetchContent(context.Background(), tab, tab.GetURL())
	}
}

func (e *engine) SetCurrentTab(index int) {
	if index < 0 || index >= len(e.tabs) {
		return
	}
	e.current = index
}

func (e *engine) AddTab() Tab {
	tab := NewTab()
	e.tabs = append(e.tabs, tab)
	e.current = len(e.tabs) - 1
	return tab
}

func (e *engine) CloseTab(index int) {
	if index < 0 || index >= len(e.tabs) {
		return
	}
	e.tabs = append(e.tabs[:index], e.tabs[index+1:]...)
	if len(e.tabs) == 0 {
		e.current = -1
	} else if e.current >= index && e.current > 0 {
		e.current--
	} else if e.current >= len(e.tabs) {
		e.current = len(e.tabs) - 1
	}
}

func (e *engine) FetchContent(ctx context.Context, tab Tab, rawURL string) error {
	parsedURL, err := normalizeURL(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", parsedURL, nil)
	if err != nil {
		return err
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() error {
		if resp.Body != nil {
			if err := resp.Body.Close(); err != nil {
				return err
			}
		}
		return nil
	}()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	tab.SetContent(string(content))
	tab.SetTitle(e.extractTitle(string(content)))
	tab.SetURL(parsedURL)
	return nil
}

func (e *engine) TabCount() int {
	return len(e.tabs)
}

func (e *engine) extractTitle(content string) string {
	re := regexp.MustCompile(`<title[^>]*>([^<]+)</title>`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return DefaultTitle
}
