package components

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/ducnd58233/gobrowser/internal/browser"
	blayout "github.com/ducnd58233/gobrowser/internal/ui/layout"
)

type Content interface {
	Render(gtx layout.Context, theme *material.Theme, tabIndex int) layout.Dimensions
}

type ContentDependencies struct {
	Engine       browser.Engine
	LayoutEngine blayout.LayoutEngine
	DebugMode    bool
}

type contentRenderer struct {
	deps ContentDependencies
	list widget.List
}

func NewContentRenderer(deps ContentDependencies) Content {
	return &contentRenderer{
		deps: deps,
		list: widget.List{List: layout.List{Axis: layout.Vertical}},
	}
}

func (cr *contentRenderer) Render(gtx layout.Context, theme *material.Theme, tabIndex int) layout.Dimensions {
	tab := cr.deps.Engine.GetTab(tabIndex)
	if tab == nil {
		return cr.renderEmptyState(gtx, theme, "No tab available")
	}

	document := tab.GetDocument()
	if document == nil {
		return cr.renderEmptyState(gtx, theme, "Loading content...")
	}

	return cr.renderDocumentContent(gtx, theme, document)
}

func (cr *contentRenderer) renderDocumentContent(gtx layout.Context, theme *material.Theme, document browser.Document) layout.Dimensions {
	viewportWidth := float64(gtx.Constraints.Max.X)
	viewportHeight := float64(gtx.Constraints.Max.Y)

	displayList := cr.deps.LayoutEngine.Layout(document, viewportWidth, viewportHeight)
	contentHeight := displayList.GetHeight()

	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return cr.list.Layout(gtx, 1, func(gtx layout.Context, index int) layout.Dimensions {
				if index == 0 {
					// Create a clipping area for the content
					contentArea := clip.Rect{Max: gtx.Constraints.Max}
					defer contentArea.Push(gtx.Ops).Pop()
					scrollY := float64(cr.list.Position.Offset)
					displayList.Paint(gtx, theme, scrollY)

					return layout.Dimensions{
						Size: image.Pt(gtx.Constraints.Max.X, int(contentHeight)),
					}
				}
				return layout.Dimensions{}
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if contentHeight > viewportHeight && contentHeight > 0 {
				viewportStart := float32(cr.list.Position.Offset) / float32(contentHeight)
				viewportEnd := float32(cr.list.Position.Offset+int(viewportHeight)) / float32(contentHeight)
				if viewportEnd > 1 {
					viewportEnd = 1
				}
				return material.Scrollbar(theme, &cr.list.Scrollbar).Layout(gtx, layout.Vertical, viewportStart, viewportEnd)
			}
			return layout.Dimensions{}
		}),
	)
}

func (cr *contentRenderer) renderEmptyState(gtx layout.Context, theme *material.Theme, message string) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		label := material.Body1(theme, message)
		return label.Layout(gtx)
	})
}
