package render

import (
	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/ducnd58233/gobrowser/internal/browser"
	"github.com/ducnd58233/gobrowser/internal/ui/types"
)

type DisplayList interface {
	Paint(gtx layout.Context, theme *material.Theme, scrollY float64)
	GetHeight() float64
	SetHeight(height float64)
	AddCommand(cmd DrawCommand)
	FindElementAt(x, y, scrollY float64) browser.Node
}

type displayList struct {
	commands []DrawCommand
	height   float64
}

func NewDisplayList() DisplayList {
	return &displayList{
		commands: make([]DrawCommand, 0),
		height:   0,
	}
}

func (dl *displayList) Paint(gtx layout.Context, theme *material.Theme, scrollY float64) {
	for _, cmd := range dl.commands {
		cmd.Execute(gtx, theme, scrollY)
	}
}

func (dl *displayList) GetHeight() float64 {
	return dl.height
}

func (dl *displayList) SetHeight(height float64) {
	dl.height = height
}

func (dl *displayList) AddCommand(cmd DrawCommand) {
	dl.commands = append(dl.commands, cmd)
}

func (dl *displayList) FindElementAt(x, y, scrollY float64) browser.Node {
	// find topmost element
	for i := len(dl.commands) - 1; i >= 0; i-- {
		cmd := dl.commands[i]
		bounds := cmd.GetBounds()

		adjY := bounds.Y - scrollY

		adjustedBounds := types.Bounds{
			X:      bounds.X,
			Y:      adjY,
			Width:  bounds.Width,
			Height: bounds.Height,
		}

		if adjustedBounds.Contains(x, y) {
			return cmd.GetNode()
		}
	}
	return nil
}
