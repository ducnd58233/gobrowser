package ui

import "fyne.io/fyne/v2/theme"

const (
	AppName        = "GoBrowser"

	// Layout constants
	TabButtonWidth  = 150
	TabButtonHeight = 32
	ToolbarHeight   = 50
	Padding         = 8
	
	// Font and icon sizes
	FontSize = 12
	IconSize = 16
)

var (
	IconAdd     = theme.ContentAddIcon()
	IconClose   = theme.CancelIcon()
	IconBack    = theme.NavigateBackIcon()
	IconForward = theme.NavigateNextIcon()
	IconRefresh = theme.ViewRefreshIcon()
	IconStop    = theme.MediaStopIcon()
	IconSearch  = theme.SearchIcon()

	URLPlaceholder = "Enter URL or search terms..."
)
