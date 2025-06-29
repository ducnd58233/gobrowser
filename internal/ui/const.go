package ui

// Application constants
const (
	AppName = "GoBrowser"
)

// Window dimensions
const (
	WindowDefaultWidth  = 1200
	WindowDefaultHeight = 800
	WindowMinWidth      = 800
	WindowMinHeight     = 600
)

// UI component heights (in pixels) - optimized for Qt
const (
	TabBarHeight   = 50 // Tab bar height
	ToolbarHeight  = 45 // Toolbar height
	URLInputHeight = 32 // URL input field height
	ButtonHeight   = 32 // Standard button height
)

// Tab styling constants - Chrome-like appearance
const (
	TabPadding      = 8
	TabSpacing      = 2
	TabMinWidth     = 180
	TabMaxWidth     = 280
	TabDefaultWidth = 200
	TabBorderRadius = 6
	CloseButtonSize = 16
)

// Tab text constants
const (
	CloseTabText           = "×"
	AddTabText             = "+"
	NewTabText             = "New Tab"
	MaxTabTitleLength      = 25
	TruncationSuffixLength = 3
)

// Progress bar constants
const (
	ProgressBarHeight = 2
	ProgressBarBg     = 0xE0E0E0 // Light gray background
	ProgressBarFill   = 0x4285F4 // Chrome blue fill
)

// Button and icon sizing
const (
	ButtonMinWidth = 32
	IconSize       = 16
	ButtonPadding  = 8
)

// Color constants (CSS-style colors for Qt)
const (
	// Active tab: Chrome blue
	TabColorActive   = "#4285f4"
	TabColorInactive = "#f5f5f5"
	TabColorHover    = "#ebebeb"
	TabBorderColor   = "#dadce0"

	// Text colors
	TabTextActive   = "#ffffff"
	TabTextInactive = "#5f6368"

	// Close button
	CloseButtonColor      = "#5f6368"
	CloseButtonHoverColor = "#ea4335"

	// URL input
	URLInputBorderColor = "#dadce0"
	URLInputBGColor     = "#ffffff"
	URLInputTextColor   = "#202124"

	// Toolbar buttons
	ButtonBGColor         = "#f8f9fa"
	ButtonBGColorDisabled = "#f1f3f4"
	ButtonTextColor       = "#5f6368"
	ButtonTextDisabled    = "#bdc1c6"

	// Go button
	GoButtonBGColor   = "#4285f4"
	GoButtonTextColor = "#ffffff"
)

// Text constants
const (
	URLPlaceholder = "Search Google or type a URL"
	LoadingText    = "Loading..."
	ErrorText      = "Error loading page"
	EmptyText      = "No content to display"
)
