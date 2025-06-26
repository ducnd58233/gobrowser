package ui

import (
	"context"
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/ducnd58233/gobrowser/internal/browser"
)

type Toolbar interface {
	Render(gtx layout.Context, theme *material.Theme, currTabIdx int) layout.Dimensions
	GetProgress() float32
	SetProgress(progress float32)
}

type toolbar struct {
	engine    browser.Engine
	urlEditor *widget.Editor
	progress  float32
	goButton  *widget.Clickable
}

func NewToolbar(engine browser.Engine) Toolbar {
	return &toolbar{
		engine:    engine,
		urlEditor: &widget.Editor{SingleLine: true, Submit: true},
		progress:  0.0,
		goButton:  &widget.Clickable{},
	}
}

func (t *toolbar) SetProgress(progress float32) {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	t.progress = progress
}

func (t *toolbar) GetProgress() float32 {
	return t.progress
}

func (t *toolbar) Render(gtx layout.Context, theme *material.Theme, currTabIdx int) layout.Dimensions {
	return layout.Flex{}.Layout(gtx, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(browser.DefaultPadding),
			Bottom: unit.Dp(browser.DefaultPadding),
			Left:   unit.Dp(browser.DefaultPadding),
			Right:  unit.Dp(browser.DefaultPadding),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
				Spacing:   layout.SpaceEvenly,
			}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return t.renderURLBarWithProgress(gtx, theme, currTabIdx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return t.renderActionButtons(gtx, theme, currTabIdx)
				}),
			)
		})
	}))
}

func (t *toolbar) renderURLBarWithProgress(gtx layout.Context, theme *material.Theme, currTabIdx int) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.renderURLBar(gtx, theme, currTabIdx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if t.progress > 0 && t.progress < 1 {
				return t.renderProgressBar(gtx)
			}
			return layout.Dimensions{}
		}),
	)
}

func (t *toolbar) renderURLBar(gtx layout.Context, theme *material.Theme, currTabIdx int) layout.Dimensions {
	if event, hasEvent := t.urlEditor.Update(gtx); hasEvent {
		if _, isSubmit := event.(widget.SubmitEvent); isSubmit {
			go t.handleNavigate(currTabIdx)
		}
	}

	field := material.Editor(theme, t.urlEditor, URLPlaceholder)
	field.HintColor = theme.Fg

	return layout.Inset{
		Left:  unit.Dp(browser.DefaultSpacing),
		Right: unit.Dp(browser.DefaultSpacing),
	}.Layout(gtx, field.Layout)
}

func (t *toolbar) renderProgressBar(gtx layout.Context) layout.Dimensions {
	progressBarHeight := gtx.Dp(unit.Dp(ProgressBarHeight))
	totalWidth := gtx.Constraints.Max.X
	progressWidth := int(float32(totalWidth) * t.progress)

	bgColor := color.NRGBA{
		R: uint8((ProgressBarBg >> 16) & 0xFF),
		G: uint8((ProgressBarBg >> 8) & 0xFF),
		B: uint8(ProgressBarBg & 0xFF),
		A: 0xFF,
	}

	fillColor := color.NRGBA{
		R: uint8((ProgressBarFill >> 16) & 0xFF),
		G: uint8((ProgressBarFill >> 8) & 0xFF),
		B: uint8(ProgressBarFill & 0xFF),
		A: 0xFF,
	}

	paint.FillShape(gtx.Ops, bgColor,
		clip.Rect{Max: image.Pt(totalWidth, progressBarHeight)}.Op())

	if progressWidth > 0 {
		paint.FillShape(gtx.Ops, fillColor,
			clip.Rect{Max: image.Pt(progressWidth, progressBarHeight)}.Op())
	}

	return layout.Dimensions{Size: image.Pt(totalWidth, progressBarHeight)}
}

func (t *toolbar) renderActionButtons(gtx layout.Context, theme *material.Theme, currTabIdx int) layout.Dimensions {
	if t.goButton.Clicked(gtx) {
		go t.handleNavigate(currTabIdx)
	}

	btn := material.Button(theme, t.goButton, "Go")
	return btn.Layout(gtx)
}

func (t *toolbar) handleNavigate(currTabIdx int) {
	url := t.urlEditor.Text()
	if url == "" {
		return
	}

	tab := t.engine.GetTab(currTabIdx)
	if tab == nil {
		tab = t.engine.AddTab()
	}

	t.SetProgress(0.1)
	tab.Navigate(url)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), browser.DefaultTimeout)
		defer cancel()

		t.SetProgress(0.3)
		if err := t.engine.FetchContent(ctx, currTabIdx, url); err == nil {
			t.SetProgress(1.0)
		} else {
			t.SetProgress(0.0)
		}

		t.engine.RefreshTab(currTabIdx)
	}()
}
