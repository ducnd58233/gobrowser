package browser

import "errors"

var (
	ErrInvalidURL     = errors.New("invalid URL provided")
	ErrNetworkTimeout = errors.New("network request timed out")
	ErrHTTPError      = errors.New("HTTP error occurred")
)
