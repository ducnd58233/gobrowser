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
	engine        browser.Engine
	urlEditor     *widget.Editor
	progress      float32
	goButton      *widget.Clickable
	backButton    *widget.Clickable
	forwardButton *widget.Clickable
	refreshButton *widget.Clickable
}

func NewToolbar(engine browser.Engine) Toolbar {
	return &toolbar{
		engine:        engine,
		urlEditor:     &widget.Editor{SingleLine: true, Submit: true},
		progress:      0.0,
		goButton:      &widget.Clickable{},
		backButton:    &widget.Clickable{},
		forwardButton: &widget.Clickable{},
		refreshButton: &widget.Clickable{},
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
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return t.renderNavigationButtons(gtx, theme, currTabIdx)
				}),
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
	// Update URL editor with current tab's URL
	tab := t.engine.GetTab(currTabIdx)
	if tab != nil && t.urlEditor.Text() == "" {
		currentURL := tab.GetURL()
		if currentURL != "" {
			t.urlEditor.SetText(currentURL)
		}
	}

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

func (t *toolbar) renderNavigationButtons(gtx layout.Context, theme *material.Theme, currTabIdx int) layout.Dimensions {
	tab := t.engine.GetTab(currTabIdx)
	canGoBack := tab != nil && tab.CanGoBack()
	canGoNext := tab != nil && tab.CanGoNext()

	return layout.Flex{
		Axis:    layout.Horizontal,
		Spacing: layout.SpaceAround,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.renderBackButton(gtx, theme, currTabIdx, canGoBack)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.renderForwardButton(gtx, theme, currTabIdx, canGoNext)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.renderRefreshButton(gtx, theme, currTabIdx)
		}),
	)
}

func (t *toolbar) renderBackButton(gtx layout.Context, theme *material.Theme, currTabIdx int, enabled bool) layout.Dimensions {
	if enabled && t.backButton.Clicked(gtx) {
		tab := t.engine.GetTab(currTabIdx)
		if tab != nil {
			tab.GoBack()
		}
	}

	btn := material.Button(theme, t.backButton, "←")
	if !enabled {
		btn.Color = color.NRGBA{R: 200, G: 200, B: 200, A: 255} // Gray for disabled
	}
	return btn.Layout(gtx)
}

func (t *toolbar) renderForwardButton(gtx layout.Context, theme *material.Theme, currTabIdx int, enabled bool) layout.Dimensions {
	if enabled && t.forwardButton.Clicked(gtx) {
		tab := t.engine.GetTab(currTabIdx)
		if tab != nil {
			tab.GoNext()
		}
	}

	btn := material.Button(theme, t.forwardButton, "→")
	if !enabled {
		btn.Color = color.NRGBA{R: 200, G: 200, B: 200, A: 255} // Gray for disabled
	}
	return btn.Layout(gtx)
}

func (t *toolbar) renderRefreshButton(gtx layout.Context, theme *material.Theme, currTabIdx int) layout.Dimensions {
	if t.refreshButton.Clicked(gtx) {
		t.engine.RefreshTab(currTabIdx)
		t.SetProgress(0.1)
		go func() {
			t.SetProgress(1.0)
		}()
	}

	btn := material.Button(theme, t.refreshButton, "⟳")
	return btn.Layout(gtx)
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
