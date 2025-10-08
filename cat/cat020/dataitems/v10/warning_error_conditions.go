// cat/cat020/dataitems/v10/warning_error_conditions.go
package v10

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/davidkohl/gobelix/asterix"
)

// WarningErrorConditions represents I020/030 - Warning/Error Conditions
// Variable length: 1+ octets
// Warning/error conditions detected by a system for the target report involved
type WarningErrorConditions struct {
	Conditions []uint8 // Warning/error condition values (7 bits each)
}

// NewWarningErrorConditions creates a new Warning/Error Conditions data item
func NewWarningErrorConditions() *WarningErrorConditions {
	return &WarningErrorConditions{}
}

// Decode decodes the Warning/Error Conditions from bytes
func (w *WarningErrorConditions) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 1 {
		return 0, fmt.Errorf("%w: need at least 1 byte, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	bytesRead := 0
	w.Conditions = []uint8{}

	for {
		if buf.Len() < 1 {
			return bytesRead, fmt.Errorf("%w: incomplete warning/error data", asterix.ErrBufferTooShort)
		}

		data := buf.Next(1)
		bytesRead++

		// Extract 7-bit value
		value := (data[0] >> 1) & 0x7F
		w.Conditions = append(w.Conditions, value)

		// Check FX bit
		fx := (data[0] & 0x01) != 0
		if !fx {
			break
		}
	}

	return bytesRead, nil
}

// Encode encodes the Warning/Error Conditions to bytes
func (w *WarningErrorConditions) Encode(buf *bytes.Buffer) (int, error) {
	if err := w.Validate(); err != nil {
		return 0, err
	}

	if len(w.Conditions) == 0 {
		return 0, fmt.Errorf("%w: at least one condition required", asterix.ErrInvalidMessage)
	}

	bytesWritten := 0

	for i, cond := range w.Conditions {
		var octet byte = (cond & 0x7F) << 1

		// Set FX bit if not the last condition
		if i < len(w.Conditions)-1 {
			octet |= 0x01
		}

		if err := buf.WriteByte(octet); err != nil {
			return bytesWritten, fmt.Errorf("writing condition %d: %w", i, err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

// Validate validates the Warning/Error Conditions
func (w *WarningErrorConditions) Validate() error {
	for i, cond := range w.Conditions {
		if cond > 127 {
			return fmt.Errorf("%w: condition %d must be 0-127, got %d", asterix.ErrInvalidMessage, i, cond)
		}
	}
	return nil
}

// String returns a string representation
func (w *WarningErrorConditions) String() string {
	if len(w.Conditions) == 0 {
		return "No conditions"
	}

	var parts []string
	for _, cond := range w.Conditions {
		condStr := getWarningErrorDescription(cond)
		parts = append(parts, condStr)
	}

	return strings.Join(parts, "; ")
}

// getWarningErrorDescription returns a description for a warning/error code
func getWarningErrorDescription(code uint8) string {
	switch code {
	case 0:
		return "Not defined"
	case 1:
		return "Multipath Reply (Reflection)"
	case 3:
		return "Split plot"
	case 10:
		return "Phantom SSR plot"
	case 11:
		return "Non-Matching Mode-3/A Code"
	case 12:
		return "Mode C code/Mode S altitude code abnormal value"
	case 15:
		return "Transponder anomaly detected"
	case 16:
		return "Duplicated or Illegal Mode S Aircraft Address"
	case 17:
		return "Mode S error correction applied"
	case 18:
		return "Undecodable Mode C code/Mode S altitude code"
	default:
		return fmt.Sprintf("Code %d", code)
	}
}
