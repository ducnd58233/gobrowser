package browser

import (
	"errors"
	"fmt"
)

// Error types for document building
var (
	ErrInvalidURL     = errors.New("invalid URL provided")
	ErrNetworkTimeout = errors.New("network request timed out")
	ErrHTTPError      = errors.New("HTTP error occurred")
	ErrInvalidInput   = errors.New("invalid input provided")
	ErrParsingFailed  = errors.New("parsing failed")
)

// BrowserError represents a browser-specific error with context
type BrowserError struct {
	Type    error
	Message string
	Context string
}

func (e *BrowserError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("%v: %s (context: %s)", e.Type, e.Message, e.Context)
	}
	return fmt.Sprintf("%v: %s", e.Type, e.Message)
}

// NewBrowserError creates a new browser error
func NewBrowserError(errType error, message string) error {
	return &BrowserError{
		Type:    errType,
		Message: message,
	}
}

// NewBrowserErrorWithContext creates a new browser error with context
func NewBrowserErrorWithContext(errType error, message, context string) error {
	return &BrowserError{
		Type:    errType,
		Message: message,
		Context: context,
	}
}
