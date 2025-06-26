package browser

import (
	"net/http"
	"sync"
)

type Engine interface {
	GetTabCount() int
	GetTab(idx int) Tab
	AddTab() Tab
	CloseTab(index int)
}

type engine struct {
	client *http.Client
	tabs   []Tab
	mutex  sync.RWMutex
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
		tabs: make([]Tab, 0),
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

func (e *engine) CloseTab(index int) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if index < 0 || index >= len(e.tabs) {
		return
	}

	e.tabs = append(e.tabs[:index], e.tabs[index+1:]...)
}
