package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"github.com/ducnd58233/gobrowser/internal/browser"
)

type MainWindow interface {
	Run()
}

type mainWindow struct {
	app     fyne.App
	window  fyne.Window
	engine  browser.Engine
	tabView TabView
	toolbar Toolbar
}

func NewMainWindow() MainWindow {
	a := app.NewWithID("gobrowser.app")
	
	w := a.NewWindow(AppName)
	w.Resize(fyne.NewSize(browser.WindowMinWidth, browser.WindowMinHeight))
	w.SetMaster()

	return &mainWindow{
		app:    a,
		window: w,
		engine: browser.NewEngine(),
	}
}

func (m *mainWindow) Run() {
	m.setupUI()
	m.createInitialTab()
	m.window.ShowAndRun()
}

func (m *mainWindow) setupUI() {
	m.tabView = NewTabView(m.engine, m.onTabSelected)
	m.toolbar = NewToolbar(m.app, m.engine, m.onNavigate)

	content := container.NewBorder(
		m.tabView.Container(),
		nil,
		nil,
		nil,
		m.createMainContent(),
	)

	m.window.SetContent(content)
}

func (m *mainWindow) createMainContent() fyne.CanvasObject {
	return container.NewBorder(
		m.toolbar.Container(),
		nil,
		nil,
		nil,
		m.tabView.Content(),
	)
}

func (m *mainWindow) createInitialTab() {
	m.engine.AddTab()
	m.tabView.RefreshTabs()
	m.toolbar.UpdateFromCurrentTab()
}

func (m *mainWindow) onNavigate() {
	fyne.Do(func() {
		m.tabView.UpdateCurrentTab()
		m.toolbar.UpdateFromCurrentTab()
		m.tabView.RefreshTabs()
	})
}

func (m *mainWindow) onTabSelected(index int) {
	fyne.Do(func() {
		m.toolbar.UpdateFromCurrentTab()
		m.tabView.UpdateCurrentTab()
	})
}
