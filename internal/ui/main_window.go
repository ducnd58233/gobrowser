package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

const (
	minWindowWidth  = 800
	minWindowHeight = 600
)

type MainWindow interface {
	Run()
}

type mainWindow struct {
	app    fyne.App
	window fyne.Window
}

func NewMainWindow() MainWindow {
	a := app.New()
	w := a.NewWindow("GoBrowser")
	w.Resize(fyne.NewSize(minWindowWidth, minWindowHeight))
	return &mainWindow{
		app:    a,
		window: w,
	}
}

func (m *mainWindow) Run() {
	m.window.ShowAndRun()
}
