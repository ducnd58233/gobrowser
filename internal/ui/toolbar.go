package ui

import (
	"context"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/ducnd58233/gobrowser/internal/browser"
)

type Toolbar interface {
	Container() fyne.CanvasObject
	UpdateFromCurrentTab()
	SetLoading(bool)
}

type toolbar struct {
	app         fyne.App
	engine      browser.Engine
	container   *fyne.Container
	urlEntry    *widget.Entry
	backBtn     *widget.Button
	forwardBtn  *widget.Button
	refreshBtn  *widget.Button
	loadingIcon *widget.ProgressBarInfinite
	onNavigate  func()
}

func NewToolbar(app fyne.App, engine browser.Engine, onNavigate func()) Toolbar {
	t := &toolbar{
		app:        app,
		engine:     engine,
		onNavigate: onNavigate,
	}
	t.setupWidgets()
	t.createContainer()
	return t
}

func (t *toolbar) Container() fyne.CanvasObject {
	return t.container
}

func (t *toolbar) setupWidgets() {
	t.urlEntry = widget.NewEntry()
	t.urlEntry.SetPlaceHolder(URLPlaceholder)
	t.urlEntry.OnSubmitted = t.handleURLSubmitted

	t.backBtn = widget.NewButtonWithIcon("", IconBack, t.handleBack)
	t.forwardBtn = widget.NewButtonWithIcon("", IconForward, t.handleForward)
	t.refreshBtn = widget.NewButtonWithIcon("", IconRefresh, t.handleRefresh)

	t.loadingIcon = widget.NewProgressBarInfinite()
	t.loadingIcon.Hide()

	t.UpdateFromCurrentTab()
}

func (t *toolbar) createContainer() {
	navButtons := container.NewHBox(t.backBtn, t.forwardBtn, t.refreshBtn)

	urlContainer := container.NewBorder(nil, nil, nil, t.loadingIcon, t.urlEntry)
	t.container = container.NewBorder(nil, nil, navButtons, nil, urlContainer)
}

func (t *toolbar) SetLoading(loading bool) {
	fyne.Do(func() {
		if loading {
			t.loadingIcon.Show()
			t.loadingIcon.Start()
			t.refreshBtn.SetIcon(IconStop)
		} else {
			t.loadingIcon.Stop()
			t.loadingIcon.Hide()
			t.refreshBtn.SetIcon(IconRefresh)
		}
	})
}

func (t *toolbar) UpdateFromCurrentTab() {
	fyne.Do(func() {
		tab := t.engine.GetCurrentTab()
		if tab != nil {
			t.urlEntry.SetText(tab.GetURL())
			t.backBtn.Enable()
			t.forwardBtn.Enable()
			t.refreshBtn.Enable()

			if !tab.CanGoBack() {
				t.backBtn.Disable()
			}
			if !tab.CanGoForward() {
				t.forwardBtn.Disable()
			}
		} else {
			t.urlEntry.SetText("")
			t.backBtn.Disable()
			t.forwardBtn.Disable()
			t.refreshBtn.Disable()
		}
	})
}

func (t *toolbar) handleURLSubmitted(input string) {
	if strings.TrimSpace(input) == "" {
		return
	}

	tab := t.engine.GetCurrentTab()
	if tab == nil {
		tab = t.engine.AddTab()
	}

	go t.navigateAsync(tab, input)
}

func (t *toolbar) navigateAsync(tab browser.Tab, input string) {
	t.SetLoading(true)
	defer t.SetLoading(false)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tab.Navigate(input)

	if err := t.engine.FetchContent(ctx, tab, input); err != nil {
		fyne.Do(func() {
			if t.app != nil {
				t.app.SendNotification(&fyne.Notification{
					Title:   "Navigation Error",
					Content: err.Error(),
				})
			}
		})
		return
	}

	if t.onNavigate != nil {
		t.onNavigate()
	}
}

func (t *toolbar) handleBack() {
	tab := t.engine.GetCurrentTab()
	if tab != nil && tab.CanGoBack() {
		tab.GoBack()
		go t.navigateAsync(tab, tab.GetURL())
	}
}

func (t *toolbar) handleForward() {
	tab := t.engine.GetCurrentTab()
	if tab != nil && tab.CanGoForward() {
		tab.GoForward()
		go t.navigateAsync(tab, tab.GetURL())
	}
}

func (t *toolbar) handleRefresh() {
	tab := t.engine.GetCurrentTab()
	if tab != nil {
		go t.navigateAsync(tab, tab.GetURL())
	}
}
