package components

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/ducnd58233/gobrowser/internal/browser"
)

type TabView interface {
	Render(gtx layout.Context, theme *material.Theme) layout.Dimensions
	GetCurrentTabIndex() int
}

type tabView struct {
	engine       browser.Engine
	currentIdx   int
	tabButtons   []*widget.Clickable
	closeButtons []*widget.Clickable
	newTabButton *widget.Clickable
	colorParser  browser.ColorParser

	tabHoverStates   []bool
	closeHoverStates []bool
}

func NewTabView(engine browser.Engine) TabView {
	return &tabView{
		engine:           engine,
		currentIdx:       0,
		tabButtons:       make([]*widget.Clickable, 0),
		closeButtons:     make([]*widget.Clickable, 0),
		newTabButton:     &widget.Clickable{},
		colorParser:      browser.NewColorParser(),
		tabHoverStates:   make([]bool, 0),
		closeHoverStates: make([]bool, 0),
	}
}

func (t *tabView) parseColor(hexStr string) color.NRGBA {
	if hexStr == "" {
		return color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	}

	r, g, b, a, err := t.colorParser.ParseColor(hexStr)
	if err != nil {
		return color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	}

	return color.NRGBA{R: r, G: g, B: b, A: a}
}

func (t *tabView) Render(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Horizontal,
	}.Layout(gtx, t.renderTabs(theme)...)
}

func (t *tabView) renderTabs(theme *material.Theme) []layout.FlexChild {
	var children []layout.FlexChild

	tabCount := t.engine.GetTabCount()
	t.ensureTabButtonsCapacity(tabCount)

	for i := 0; i < tabCount; i++ {
		tab := t.engine.GetTab(i)
		if tab == nil {
			continue
		}

		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if t.tabButtons[i].Clicked(gtx) {
				t.currentIdx = i
			}

			if t.closeButtons[i].Clicked(gtx) {
				t.engine.CloseTab(i)
				tabCount = t.engine.GetTabCount()
				if t.currentIdx >= tabCount {
					t.currentIdx = tabCount - 1
				}
			}

			return t.renderTab(gtx, theme, tab, i)
		}))
	}

	children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return t.renderNewTabButton(gtx, theme)
	}))

	return children
}

func (t *tabView) renderTab(gtx layout.Context, theme *material.Theme, tab browser.Tab, index int) layout.Dimensions {
	isActive := index == t.currentIdx
	tabTitle := t.getTabTitle(tab)

	minSize := image.Pt(gtx.Dp(unit.Dp(TabMinWidth)), gtx.Dp(unit.Dp(TabBarHeight)))
	maxSize := image.Pt(gtx.Dp(unit.Dp(TabMaxWidth)), gtx.Dp(unit.Dp(TabBarHeight)))

	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min = minSize
			gtx.Constraints.Max = maxSize

			return t.renderTabBackground(gtx, isActive, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					// Tab title
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return t.renderTabButton(gtx, theme, tabTitle, index, isActive)
					}),
					// Close button
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return t.renderCloseButton(gtx, theme, index, isActive)
					}),
				)
			})
		}),
	)
}

func (t *tabView) renderTabBackground(gtx layout.Context, isActive bool, w layout.Widget) layout.Dimensions {
	var bgColor color.NRGBA
	if isActive {
		bgColor = t.parseColor(TabColorActive)
	} else {
		bgColor = t.parseColor(TabColorInactive)
	}

	// Draw background
	defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: bgColor}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)

	// Add border if not active
	if !isActive {
		borderColor := t.parseColor(TabBorderColor)
		t.drawTabBorder(gtx, borderColor)
	}

	return w(gtx)
}

func (t *tabView) renderTabButton(gtx layout.Context, theme *material.Theme, title string, index int, isActive bool) layout.Dimensions {
	var textColor color.NRGBA
	if isActive {
		textColor = t.parseColor(TabTextActive)
	} else {
		textColor = t.parseColor(TabTextInactive)
	}

	return material.Clickable(gtx, t.tabButtons[index], func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(TabPadding)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			label := material.Body1(theme, title)
			label.Color = textColor
			label.MaxLines = 1
			return label.Layout(gtx)
		})
	})
}

func (t *tabView) renderCloseButton(gtx layout.Context, theme *material.Theme, index int, isActive bool) layout.Dimensions {
	buttonColor := t.parseColor(CloseButtonColor)
	if isActive {
		buttonColor = t.parseColor(TabTextActive)
	}

	return material.Clickable(gtx, t.closeButtons[index], func(gtx layout.Context) layout.Dimensions {
		size := gtx.Dp(unit.Dp(CloseButtonSize))
		gtx.Constraints = layout.Exact(image.Pt(size, size))

		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			label := material.Body1(theme, CloseTabText)
			label.Color = buttonColor
			label.Alignment = text.Middle
			return label.Layout(gtx)
		})
	})
}

func (t *tabView) renderNewTabButton(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	if t.newTabButton.Clicked(gtx) {
		newTab := t.engine.AddTab()
		if newTab != nil {
			t.currentIdx = t.engine.GetTabCount() - 1
		}
	}

	buttonSize := gtx.Dp(unit.Dp(TabBarHeight))

	return material.Clickable(gtx, t.newTabButton, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints = layout.Exact(image.Pt(buttonSize, buttonSize))

		bgColor := t.parseColor(TabColorInactive)
		defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
		paint.ColorOp{Color: bgColor}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)

		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			label := material.Body1(theme, AddTabText)
			label.Color = t.parseColor(TabTextInactive)
			return label.Layout(gtx)
		})
	})
}

func (t *tabView) drawTabBorder(gtx layout.Context, borderColor color.NRGBA) {
	borderWidth := gtx.Dp(unit.Dp(1))
	size := gtx.Constraints.Max

	// Top border
	defer clip.Rect{Max: image.Pt(size.X, borderWidth)}.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: borderColor}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)

	// Right border
	stack := op.Offset(image.Pt(size.X-borderWidth, 0)).Push(gtx.Ops)
	defer stack.Pop()
	defer clip.Rect{Max: image.Pt(borderWidth, size.Y)}.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: borderColor}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
}

func (t *tabView) ensureTabButtonsCapacity(tabCount int) {
	for len(t.tabButtons) < tabCount {
		t.tabButtons = append(t.tabButtons, &widget.Clickable{})
		t.closeButtons = append(t.closeButtons, &widget.Clickable{})
		t.tabHoverStates = append(t.tabHoverStates, false)
		t.closeHoverStates = append(t.closeHoverStates, false)
	}
}

func (t *tabView) getTabTitle(tab browser.Tab) string {
	title := tab.GetTitle()
	if title == "" {
		title = NewTabText
	}

	if len(title) > MaxTabTitleLength {
		title = title[:MaxTabTitleLength-TruncationSuffixLength] + "..."
	}

	return title
}

func (t *tabView) GetCurrentTabIndex() int {
	return t.currentIdx
}
