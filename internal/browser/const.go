package browser

import "time"

const (
	DefaultUserAgent         = "GoBrowser/1.0"
	IDByteLength             = 8
	MaxConcurrentConnections = 10
	DefaultTimeoutSec        = 30 * time.Second
	KeepAliveTimeoutSec      = 30 * time.Second
)

const (
	DefaultSpacing = 8
	DefaultPadding = 8

	// Font Sizes
	FontSizeSmall   = 12
	FontSizeDefault = 14
	FontSizeLarge   = 16
	FontSizeHeading = 18
)
