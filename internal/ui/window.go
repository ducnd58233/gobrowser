package ui

import (
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/op"
	"gioui.org/unit"
)

type MainWindow interface {
	Run()
}

type mainWindow struct {
	window *app.Window
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

	return &mainWindow{
		window: window,
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
				e.Frame(gtx.Ops)
			}
		}
	}()
	app.Main()
}
