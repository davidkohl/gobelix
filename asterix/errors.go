// asterix/errors.go
package asterix

import (
	"fmt"
	"strings"
)

// Error type constants for better error classification
var (
	// Core errors
	ErrInvalidMessage  = fmt.Errorf("invalid ASTERIX message")
	ErrInvalidLength   = fmt.Errorf("invalid length")
	ErrInvalidFSPEC    = fmt.Errorf("invalid FSPEC")
	ErrMandatoryField  = fmt.Errorf("mandatory field missing")
	ErrInvalidCategory = fmt.Errorf("invalid category")
	ErrUnknownDataItem = fmt.Errorf("unknown data item")
	ErrInvalidField    = fmt.Errorf("invalid field value")
	ErrUAPNotDefined   = fmt.Errorf("UAP not defined for category")
	ErrUnknownCategory = fmt.Errorf("unknown category")

	// Additional error types for better error handling
	ErrBufferTooShort  = fmt.Errorf("buffer too short")
	ErrCorruptData     = fmt.Errorf("corrupt or malformed data")
	ErrDecodingFailure = fmt.Errorf("failed to decode data")
)

// ValidationError provides detailed context for validation failures
type ValidationError struct {
	DataItem string
	Field    string
	Value    any
	Reason   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error in %s.%s: %v - %s",
		e.DataItem, e.Field, e.Value, e.Reason)
}

func (e *ValidationError) Unwrap() error {
	return ErrInvalidField
}

// DecodeError provides rich context about where a decoding error occurred
type DecodeError struct {
	Category   Category
	Message    string
	DataItem   string
	Position   int
	BufferSize int
	Cause      error
}

func (e *DecodeError) Error() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("decoding error in Category %d", e.Category))

	if e.DataItem != "" {
		builder.WriteString(fmt.Sprintf(", item %s", e.DataItem))
	}

	if e.Position > 0 {
		builder.WriteString(fmt.Sprintf(", at byte %d/%d", e.Position, e.BufferSize))
	}

	if e.Message != "" {
		builder.WriteString(": " + e.Message)
	}

	if e.Cause != nil {
		builder.WriteString(": " + e.Cause.Error())
	}

	return builder.String()
}

func (e *DecodeError) Unwrap() error {
	return e.Cause
}

// IsDecodeError checks if an error is or wraps a DecodeError
func IsDecodeError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "decoding") ||
		strings.Contains(err.Error(), "decode"))
}

// NewDecodeError creates a new DecodeError with the given parameters
func NewDecodeError(category Category, message string, cause error) *DecodeError {
	return &DecodeError{
		Category: category,
		Message:  message,
		Cause:    cause,
	}
}

// WithDataItem adds data item context to a DecodeError
func (e *DecodeError) WithDataItem(dataItem string) *DecodeError {
	e.DataItem = dataItem
	return e
}

// WithPosition adds position context to a DecodeError
func (e *DecodeError) WithPosition(pos, bufSize int) *DecodeError {
	e.Position = pos
	e.BufferSize = bufSize
	return e
}
