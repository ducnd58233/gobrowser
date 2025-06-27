package browser

import "time"

const (
	DefaultUserAgent         = "GoBrowser/1.0"
	IDByteLength             = 8
	MaxConcurrentConnections = 10
	DefaultTimeout           = 30 * time.Second
	KeepAliveTimeout         = 30 * time.Second

	Important = "!important"
)

// Layout and Typography
const (
	DefaultFontSize      = 16.0
	DefaultLineHeight    = 20.0
	MinFontSize          = 8.0
	MaxFontSize          = 72.0
	BlockMargin          = 8.0
	InlineMargin         = 4.0
	DefaultTextFont      = "sans-serif"
	DefaultMonospaceFont = "monospace"
	DefaultSerifFont     = "serif"
	DefaultTextColor     = "#000000"
	DefaultBgColor       = "#FFFFFF"
	DefaultLinkColor     = "#0000EE"
	CharWidthRatio       = 0.6
	LineHeightRatio      = 1.2

	// UI Spacing Constants
	DefaultSpacing = 8
	CompactSpacing = 4
	LargeSpacing   = 16
	DefaultPadding = 8
	CompactPadding = 4
	LargePadding   = 16
	ButtonSpacing  = 4

	// Font Sizes
	FontSizeSmall   = 12
	FontSizeDefault = 14
	FontSizeLarge   = 16
	FontSizeHeading = 18
)

// CSS Properties
const (
	PropColor           = "color"
	PropBackgroundColor = "background-color"
	PropFontFamily      = "font-family"
	PropFontSize        = "font-size"
	PropFontWeight      = "font-weight"
	PropFontStyle       = "font-style"
	PropTextAlign       = "text-align"
	PropDisplay         = "display"
	PropMargin          = "margin"
	PropMarginTop       = "margin-top"
	PropMarginRight     = "margin-right"
	PropMarginBottom    = "margin-bottom"
	PropMarginLeft      = "margin-left"
	PropPadding         = "padding"
	PropPaddingTop      = "padding-top"
	PropPaddingRight    = "padding-right"
	PropPaddingBottom   = "padding-bottom"
	PropPaddingLeft     = "padding-left"
	PropBorder          = "border"
	PropBorderTop       = "border-top"
	PropBorderRight     = "border-right"
	PropBorderBottom    = "border-bottom"
	PropBorderLeft      = "border-left"
	PropWidth           = "width"
	PropHeight          = "height"
	PropPosition        = "position"
	PropTop             = "top"
	PropRight           = "right"
	PropBottom          = "bottom"
	PropLeft            = "left"
	PropZIndex          = "z-index"
	PropOpacity         = "opacity"
	PropVisibility      = "visibility"
	PropOverflow        = "overflow"
	PropTextDecoration  = "text-decoration"
	PropTextTransform   = "text-transform"
	PropLineHeight      = "line-height"
	PropLetterSpacing   = "letter-spacing"
	PropWordSpacing     = "word-spacing"
	PropWhiteSpace      = "white-space"
	PropCursor          = "cursor"
	PropListStyle       = "list-style"
)

// CSS Values
const (
	ValueNone        = "none"
	ValueBlock       = "block"
	ValueInline      = "inline"
	ValueInlineBlock = "inline-block"
	ValueFlex        = "flex"
	ValueGrid        = "grid"
	ValueHidden      = "hidden"
	ValueVisible     = "visible"
	ValueBold        = "bold"
	ValueNormal      = "normal"
	ValueItalic      = "italic"
	ValueUnderline   = "underline"
	ValueLineThrough = "line-through"
	ValueUppercase   = "uppercase"
	ValueLowercase   = "lowercase"
	ValueCapitalize  = "capitalize"
	ValueLeft        = "left"
	ValueRight       = "right"
	ValueCenter      = "center"
	ValueJustify     = "justify"
	ValueAuto        = "auto"
	ValueStatic      = "static"
	ValueRelative    = "relative"
	ValueAbsolute    = "absolute"
	ValueFixed       = "fixed"
	ValueSticky      = "sticky"
	ValueTransparent = "transparent"
)

// Color Parsing Constants
const (
	HexColorShortLength = 3 // #RGB
	HexColorFullLength  = 6 // #RRGGBB
	HexColorAlphaLength = 8 // #RRGGBBAA
	MaxColorValue       = 255
	MaxColorBits        = 8
	HexBase             = 16
	RGBShortMultiplier  = 17
	DefaultAlpha        = 255
)
