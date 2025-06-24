package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/ducnd58233/gobrowser/internal/browser"
)

type TabView interface {
	Container() fyne.CanvasObject
	Content() fyne.CanvasObject
	UpdateCurrentTab()
	RefreshTabs()
}

type tabView struct {
	engine     browser.Engine
	container  *fyne.Container
	content    *widget.RichText
	onSelected func(index int)
	tabButtons []*widget.Button
}

func NewTabView(engine browser.Engine, onSelected func(int)) TabView {
	tv := &tabView{
		engine:     engine,
		container:  container.NewHBox(),
		content:    widget.NewRichText(),
		onSelected: onSelected,
		tabButtons: make([]*widget.Button, 0),
	}

	tv.content.Wrapping = fyne.TextWrapWord
	return tv
}

func (tv *tabView) Container() fyne.CanvasObject {
	return tv.container
}

func (tv *tabView) Content() fyne.CanvasObject {
	return container.NewScroll(tv.content)
}

func (tv *tabView) UpdateCurrentTab() {
	fyne.Do(func() {
		tab := tv.engine.GetCurrentTab()
		if tab != nil {
			content := tab.GetContent()
			if content != "" {
				tv.content.ParseMarkdown(content)
			} else {
				tv.content.ParseMarkdown("# " + tab.GetTitle() + "\n\nLoading...")
			}
		}
	})
}

func (tv *tabView) RefreshTabs() {
	fyne.Do(func() {
		tv.tabButtons = make([]*widget.Button, 0)
		objects := make([]fyne.CanvasObject, 0)

		for i := 0; i < tv.engine.TabCount(); i++ {
			tab := tv.engine.GetTab(i)
			if tab != nil {
				button := tv.createTabButton(tab, i)
				tv.tabButtons = append(tv.tabButtons, button)
				objects = append(objects, button)
			}
		}

		objects = append(objects, tv.createAddButton())

		tv.container.Objects = objects
		tv.container.Refresh()
	})
}

func (tv *tabView) createTabButton(tab browser.Tab, index int) *widget.Button {
	button := widget.NewButton(tab.GetTitle(), func() {
		tv.selectTab(index)
	})
	button.Resize(fyne.NewSize(TabButtonWidth, TabButtonHeight))
	return button
}

func (tv *tabView) selectTab(index int) {
	tv.engine.SetCurrentTab(index)
	tv.UpdateCurrentTab()
	if tv.onSelected != nil {
		tv.onSelected(index)
	}
}

func (tv *tabView) createAddButton() *widget.Button {
	return widget.NewButtonWithIcon("", IconAdd, func() {
		tv.engine.AddTab()
		tv.RefreshTabs()
		tv.selectTab(tv.engine.TabCount() - 1)
	})
}
