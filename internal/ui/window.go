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
	"github.com/ducnd58233/gobrowser/internal/ui/components"
	blayout "github.com/ducnd58233/gobrowser/internal/ui/layout"
)

type MainWindow interface {
	Run()
}

type mainWindow struct {
	window *app.Window
	theme  *material.Theme
	engine browser.Engine

	tabView         components.TabView
	toolbar         components.Toolbar
	contentRenderer components.Content
}

func NewMainWindow(isDebugMode bool) MainWindow {
	window := createAppWindow()

	engine := browser.NewEngine()
	engine.SetDebugMode(isDebugMode)
	engine.AddTab()

	theme := createTheme()

	layoutEngineDeps := blayout.LayoutEngineDependencies{
		ColorParser: browser.NewColorParser(),
		UnitParser:  browser.NewUnitParser(),
		Cache:       blayout.NewLayoutCache(),
	}
	contentRendererDeps := components.ContentDependencies{
		Engine:       engine,
		LayoutEngine: blayout.NewLayoutEngine(layoutEngineDeps),
		DebugMode:    isDebugMode,
	}

	return &mainWindow{
		window:          window,
		theme:           theme,
		engine:          engine,
		tabView:         components.NewTabView(engine),
		toolbar:         components.NewToolbar(engine),
		contentRenderer: components.NewContentRenderer(contentRendererDeps),
	}
}

func createAppWindow() *app.Window {
	window := &app.Window{}
	window.Option(
		app.Title(components.AppName),
		app.Size(
			unit.Dp(components.WindowDefaultWidth),
			unit.Dp(components.WindowDefaultHeight),
		),
		app.MinSize(
			unit.Dp(components.WindowMinWidth),
			unit.Dp(components.WindowMinHeight),
		),
	)
	return window
}

func createTheme() *material.Theme {
	theme := material.NewTheme()
	theme.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	return theme
}

func (mw *mainWindow) Run() {
	go mw.runEventLoop()
	app.Main()
}

func (mw *mainWindow) runEventLoop() {
	var ops op.Ops

	for {
		event := mw.window.Event()
		switch e := event.(type) {
		case app.DestroyEvent:
			mw.handleDestroyEvent(e)
		case app.FrameEvent:
			mw.handleFrameEvent(&ops, e)
		}
	}
}

func (mw *mainWindow) handleDestroyEvent(e app.DestroyEvent) {
	if e.Err != nil {
		log.Fatalf("Window destroy error: %v", e.Err)
	}
	os.Exit(0)
}

func (mw *mainWindow) handleFrameEvent(ops *op.Ops, e app.FrameEvent) {
	gtx := app.NewContext(ops, e)
	mw.render(gtx)
	e.Frame(gtx.Ops)
	mw.window.Invalidate()
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
			return mw.contentRenderer.Render(gtx, mw.theme, mw.tabView.GetCurrentTabIndex())
		}),
	)
}
