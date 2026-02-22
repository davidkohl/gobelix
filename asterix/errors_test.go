// asterix/errors_test.go
package asterix

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestValidationError(t *testing.T) {
	err := NewValidationError("I021/010", "SAC", 255, "value too high")

	// Test error message format
	errMsg := err.Error()
	if !strings.Contains(errMsg, "I021/010") || !strings.Contains(errMsg, "SAC") ||
		!strings.Contains(errMsg, "255") || !strings.Contains(errMsg, "value too high") {
		t.Errorf("Error message %q missing expected content", errMsg)
	}

	// Test unwrapping
	if !errors.Is(err, ErrInvalidField) {
		t.Errorf("ValidationError should unwrap to ErrInvalidField")
	}

	// Test error classification
	if !IsValidationError(err) {
		t.Errorf("IsValidationError should return true for ValidationError")
	}
}

func TestDecodeError(t *testing.T) {
	// Create a basic decode error
	err := NewDecodeError(Cat021, "I021/010", "failed to read bytes", ErrBufferTooShort)

	// Add position information
	err = err.WithPosition(42, 100)

	// Test error message format
	errMsg := err.Error()
	expectedParts := []string{"decoding error", "Category 21", "I021/010", "byte 42/100", "failed to read bytes", "buffer too short"}
	for _, part := range expectedParts {
		if !strings.Contains(errMsg, part) {
			t.Errorf("Error message %q missing expected part %q", errMsg, part)
		}
	}

	// Test unwrapping
	if !errors.Is(err, ErrBufferTooShort) {
		t.Errorf("DecodeError should unwrap to the cause error")
	}

	// Test error classification
	if !IsDecodeError(err) {
		t.Errorf("IsDecodeError should return true for DecodeError")
	}
}

func TestEncodingError(t *testing.T) {
	// Create a basic encoding error
	err := NewEncodingError(Cat021, "I021/010", "failed to encode value", ErrInvalidField)

	// Add position information
	err = err.WithPosition(42)

	// Test error message format
	errMsg := err.Error()
	expectedParts := []string{"encoding error", "Category 21", "I021/010", "byte 42", "failed to encode value", "invalid field"}
	for _, part := range expectedParts {
		if !strings.Contains(errMsg, part) {
			t.Errorf("Error message %q missing expected part %q", errMsg, part)
		}
	}

	// Test unwrapping
	if !errors.Is(err, ErrInvalidField) {
		t.Errorf("EncodingError should unwrap to the cause error")
	}

	// Test error classification
	if !IsEncodingError(err) {
		t.Errorf("IsEncodingError should return true for EncodingError")
	}
}

func TestErrorHelpers(t *testing.T) {
	testCases := []struct {
		name             string
		err              error
		isDecodeErr      bool
		isEncodeErr      bool
		isValidateErr    bool
		isBufferErr      bool
		isMandatoryErr   bool
		isUnknownItemErr bool
	}{
		{
			"DecodeError",
			NewDecodeError(Cat021, "I021/010", "test", nil),
			true, false, false, false, false, false,
		},
		{
			"EncodingError",
			NewEncodingError(Cat021, "I021/010", "test", nil),
			false, true, false, false, false, false,
		},
		{
			"ValidationError",
			NewValidationError("I021/010", "SAC", 255, "test"),
			false, false, true, false, false, false,
		},
		{
			"BufferTooShortError",
			ErrBufferTooShort,
			false, false, false, true, false, false,
		},
		{
			"MandatoryFieldError",
			ErrMandatoryField,
			false, false, false, false, true, false,
		},
		{
			"UnknownDataItemError",
			ErrUnknownDataItem,
			false, false, false, false, false, true,
		},
		{
			"WrappedDecodeError",
			fmt.Errorf("outer error: %w", NewDecodeError(Cat021, "I021/010", "test", nil)),
			true, false, false, false, false, false,
		},
		{
			"Nil",
			nil,
			false, false, false, false, false, false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if IsDecodeError(tc.err) != tc.isDecodeErr {
				t.Errorf("IsDecodeError = %v, want %v", IsDecodeError(tc.err), tc.isDecodeErr)
			}
			if IsEncodingError(tc.err) != tc.isEncodeErr {
				t.Errorf("IsEncodingError = %v, want %v", IsEncodingError(tc.err), tc.isEncodeErr)
			}
			if IsValidationError(tc.err) != tc.isValidateErr {
				t.Errorf("IsValidationError = %v, want %v", IsValidationError(tc.err), tc.isValidateErr)
			}
			if IsBufferTooShort(tc.err) != tc.isBufferErr {
				t.Errorf("IsBufferTooShort = %v, want %v", IsBufferTooShort(tc.err), tc.isBufferErr)
			}
			if IsMandatoryFieldMissing(tc.err) != tc.isMandatoryErr {
				t.Errorf("IsMandatoryFieldMissing = %v, want %v", IsMandatoryFieldMissing(tc.err), tc.isMandatoryErr)
			}
			if IsUnknownDataItem(tc.err) != tc.isUnknownItemErr {
				t.Errorf("IsUnknownDataItem = %v, want %v", IsUnknownDataItem(tc.err), tc.isUnknownItemErr)
			}
		})
	}
}

func TestWrapError(t *testing.T) {
	// Should return nil for nil error
	if WrapError(nil, "test") != nil {
		t.Errorf("WrapError(nil) should return nil")
	}

	// Should wrap error with message
	baseErr := errors.New("base error")
	wrappedErr := WrapError(baseErr, "context %d", 42)

	// Check message format
	errMsg := wrappedErr.Error()
	if !strings.Contains(errMsg, "context 42") || !strings.Contains(errMsg, "base error") {
		t.Errorf("Error message %q missing expected content", errMsg)
	}

	// Check unwrapping
	if !errors.Is(wrappedErr, baseErr) {
		t.Errorf("WrapError should preserve the original error for unwrapping")
	}
}
