package ui

import (
	"gioui.org/layout"
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
}

func NewTabView(engine browser.Engine) TabView {
	return &tabView{
		engine:       engine,
		currentIdx:   0,
		tabButtons:   make([]*widget.Clickable, 0),
		closeButtons: make([]*widget.Clickable, 0),
		newTabButton: &widget.Clickable{},
	}
}

func (t *tabView) Render(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Horizontal,
	}.Layout(gtx, t.renderTabs(theme)...)
}

func (t *tabView) renderTabs(theme *material.Theme) []layout.FlexChild {
	var children []layout.FlexChild

	tabCount := t.engine.GetTabCount()

	for i := len(t.tabButtons); i < tabCount; i++ {
		t.tabButtons = append(t.tabButtons, &widget.Clickable{})
	}

	for i := len(t.closeButtons); i < tabCount; i++ {
		t.closeButtons = append(t.closeButtons, &widget.Clickable{})
	}

	for i := 0; i < tabCount; i++ {
		tab := t.engine.GetTab(i)
		if tab == nil {
			continue
		}

		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.renderTab(gtx, theme, tab, i)
		}))
	}

	children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return t.renderNewTabButton(gtx, theme)
	}))

	return children
}

func (t *tabView) renderTab(gtx layout.Context, theme *material.Theme, tab browser.Tab, index int) layout.Dimensions {
	if t.tabButtons[index].Clicked(gtx) {
		t.currentIdx = index
	}

	if t.closeButtons[index].Clicked(gtx) {
		t.engine.CloseTab(index)
		tabCount := t.engine.GetTabCount()
		if tabCount == 0 {
			t.currentIdx = 0
		} else if t.currentIdx >= tabCount {
			t.currentIdx = tabCount - 1
		}
		return layout.Dimensions{}
	}

	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.Button(theme, t.tabButtons[index], t.getTabTitle(tab)).Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.Button(theme, t.closeButtons[index], "×").Layout(gtx)
		}),
	)
}

func (t *tabView) renderNewTabButton(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	if t.newTabButton.Clicked(gtx) {
		t.engine.AddTab()
		t.currentIdx = t.engine.GetTabCount() - 1
	}

	return material.Button(theme, t.newTabButton, "+").Layout(gtx)
}

func (t *tabView) getTabTitle(tab browser.Tab) string {
	title := tab.GetTitle()

	if len(title) > MaxTabTitleLength {
		title = title[:MaxTabTitleLength-TruncationSuffixLength] + "..."
	}

	return title
}

func (t *tabView) GetCurrentTabIndex() int {
	return t.currentIdx
}
