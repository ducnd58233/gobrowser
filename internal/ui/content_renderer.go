package ui

import (
	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/ducnd58233/gobrowser/internal/browser"
)

type ContentRenderer interface {
	Render(gtx layout.Context, theme *material.Theme, tabIndex int) layout.Dimensions
}

type contentRenderer struct {
	engine         browser.Engine
	debugMode      bool
}

func NewContentRenderer(
	engine         browser.Engine,
	debugMode      bool,
) ContentRenderer {
	return &contentRenderer{
		engine:         engine,
		debugMode:      debugMode,
	}
}

func (cr *contentRenderer) Render(gtx layout.Context, theme *material.Theme, tabIndex int) layout.Dimensions {
	tab := cr.engine.GetTab(tabIndex)
	if tab == nil {
		return cr.renderEmptyState(gtx, theme, "No tab available")
	}

	document := tab.GetDocument()
	if document == nil {
		return cr.renderEmptyState(gtx, theme, "Loading content...")
	}

	return layout.Dimensions{}
}
func (cr *contentRenderer) renderEmptyState(gtx layout.Context, theme *material.Theme, message string) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		label := material.Body1(theme, message)
		return label.Layout(gtx)
	})
}
