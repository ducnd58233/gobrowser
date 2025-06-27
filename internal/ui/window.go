package ui

import (
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/ducnd58233/gobrowser/internal/browser"
)

type MainWindow interface {
	Run()
}

type mainWindow struct {
	window *app.Window
	theme  *material.Theme
	engine browser.Engine

	tabView TabView
	toolbar Toolbar
}

func NewMainWindow() MainWindow {
	window := &app.Window{}
	window.Option(
		app.Title(AppName),
		app.Size(
			unit.Dp(WindowDefaultWidth),
			unit.Dp(WindowDefaultHeight),
		),
		app.MinSize(
			unit.Dp(WindowMinWidth),
			unit.Dp(WindowMinHeight),
		),
	)

	engine := browser.NewEngine()
	engine.AddTab()

	theme := material.NewTheme()
	theme.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	return &mainWindow{
		window:  window,
		theme:   theme,
		engine:  engine,
		tabView: NewTabView(engine),
		toolbar: NewToolbar(engine),
	}
}

func (mw *mainWindow) Run() {
	go func() {
		var ops op.Ops
		for {
			event := mw.window.Event()
			switch e := event.(type) {
			case app.DestroyEvent:
				if e.Err != nil {
					log.Fatalf("Window destroy error: %v", e.Err)
				}
				os.Exit(0)
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)
				mw.render(gtx)
				e.Frame(gtx.Ops)
			}
		}
	}()
	app.Main()
}

func (mw *mainWindow) render(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return mw.tabView.Render(gtx, mw.theme)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return mw.toolbar.Render(gtx, mw.theme, mw.tabView.GetCurrentTabIndex())
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return mw.renderContent()
		}),
	)
}

func (mw *mainWindow) renderContent() layout.Dimensions {
	currentTabIdx := mw.tabView.GetCurrentTabIndex()
	tab := mw.engine.GetTab(currentTabIdx)

	if tab == nil {
		return layout.Dimensions{}
	}

	document := tab.GetDocument()
	if document == nil {
		return layout.Dimensions{}
	}

	return layout.Dimensions{}
}
