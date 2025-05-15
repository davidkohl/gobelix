// asterix/errors.go
package asterix

import (
	"errors"
	"fmt"
)

// Core error types for better error classification
var (
	// Structural errors
	ErrInvalidMessage  = errors.New("invalid ASTERIX message")
	ErrInvalidLength   = errors.New("invalid length")
	ErrInvalidFSPEC    = errors.New("invalid FSPEC")
	ErrMandatoryField  = errors.New("mandatory field missing")
	ErrInvalidCategory = errors.New("invalid category")
	ErrUnknownDataItem = errors.New("unknown data item")
	ErrInvalidField    = errors.New("invalid field value")
	ErrUAPNotDefined   = errors.New("UAP not defined for category")
	ErrUnknownCategory = errors.New("unknown category")

	// Processing errors
	ErrBufferTooShort   = errors.New("buffer too short")
	ErrCorruptData      = errors.New("corrupt or malformed data")
	ErrDecodingFailure  = errors.New("failed to decode data")
	ErrTruncatedMessage = errors.New("message truncated")
)

// ValidationError provides detailed context for validation failures
type ValidationError struct {
	DataItem string
	Field    string
	Value    any
	Reason   string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error in %s.%s: %v - %s",
		e.DataItem, e.Field, e.Value, e.Reason)
}

// Unwrap returns the underlying error
func (e *ValidationError) Unwrap() error {
	return ErrInvalidField
}

// NewValidationError creates a new validation error
func NewValidationError(dataItem, field string, value any, reason string) *ValidationError {
	return &ValidationError{
		DataItem: dataItem,
		Field:    field,
		Value:    value,
		Reason:   reason,
	}
}

// DecodeError provides rich context about where a decoding error occurred
type DecodeError struct {
	Category   Category
	DataItem   string
	Position   int
	BufferSize int
	Message    string
	Cause      error
}

// Error implements the error interface
func (e *DecodeError) Error() string {
	s := fmt.Sprintf("decoding error in Category %d", e.Category)

	if e.DataItem != "" {
		s += fmt.Sprintf(", item %s", e.DataItem)
	}

	if e.Position > 0 {
		s += fmt.Sprintf(", at byte %d/%d", e.Position, e.BufferSize)
	}

	if e.Message != "" {
		s += ": " + e.Message
	}

	if e.Cause != nil {
		s += ": " + e.Cause.Error()
	}

	return s
}

// Unwrap returns the underlying error
func (e *DecodeError) Unwrap() error {
	return e.Cause
}

// NewDecodeError creates a new decode error
func NewDecodeError(category Category, dataItem string, message string, cause error) *DecodeError {
	return &DecodeError{
		Category: category,
		DataItem: dataItem,
		Message:  message,
		Cause:    cause,
	}
}

// WithPosition adds position information to a decode error
func (e *DecodeError) WithPosition(pos, bufSize int) *DecodeError {
	e.Position = pos
	e.BufferSize = bufSize
	return e
}

// EncodingError provides detailed information about encoding failures
type EncodingError struct {
	Category Category
	DataItem string
	Position int
	Message  string
	Cause    error
}

// Error implements the error interface
func (e *EncodingError) Error() string {
	s := fmt.Sprintf("encoding error in Category %d", e.Category)

	if e.DataItem != "" {
		s += fmt.Sprintf(", item %s", e.DataItem)
	}

	if e.Position > 0 {
		s += fmt.Sprintf(", at byte %d", e.Position)
	}

	if e.Message != "" {
		s += ": " + e.Message
	}

	if e.Cause != nil {
		s += ": " + e.Cause.Error()
	}

	return s
}

// Unwrap returns the underlying error
func (e *EncodingError) Unwrap() error {
	return e.Cause
}

// NewEncodingError creates a new encoding error
func NewEncodingError(category Category, dataItem string, message string, cause error) *EncodingError {
	return &EncodingError{
		Category: category,
		DataItem: dataItem,
		Message:  message,
		Cause:    cause,
	}
}

// WithPosition adds position information to an encoding error
func (e *EncodingError) WithPosition(pos int) *EncodingError {
	e.Position = pos
	return e
}

// IsDecodeError checks if an error is or wraps a DecodeError
func IsDecodeError(err error) bool {
	var decodeErr *DecodeError
	return errors.As(err, &decodeErr)
}

// IsEncodingError checks if an error is or wraps an EncodingError
func IsEncodingError(err error) bool {
	var encodeErr *EncodingError
	return errors.As(err, &encodeErr)
}

// IsValidationError checks if an error is or wraps a ValidationError
func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}

// IsBufferTooShort checks if an error indicates a buffer is too short
func IsBufferTooShort(err error) bool {
	return errors.Is(err, ErrBufferTooShort)
}

// IsMandatoryFieldMissing checks if an error indicates a mandatory field is missing
func IsMandatoryFieldMissing(err error) bool {
	return errors.Is(err, ErrMandatoryField)
}

// IsUnknownDataItem checks if an error indicates an unknown data item
func IsUnknownDataItem(err error) bool {
	return errors.Is(err, ErrUnknownDataItem)
}

// WrapError wraps an error with additional context
func WrapError(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(format+": %w", append(args, err)...)
}
