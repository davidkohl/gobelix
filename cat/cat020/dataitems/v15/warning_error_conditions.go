// cat/cat020/dataitems/v15/warning_error_conditions.go
package v15

import (
	"bytes"
	"fmt"
)

// WarningErrorConditions implements I020/030 - Warning/Error Conditions
type WarningErrorConditions struct {
	WE uint8 // Warning/Error value (0-127 per spec)
}

func (w *WarningErrorConditions) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// First octet
	data, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading warning/error conditions: %w", err)
	}
	bytesRead++

	// Bits 8-2: Warning/Error value (7 bits)
	w.WE = (data >> 1) & 0x7F
	fx := (data & 0x01) != 0

	// Read remaining extension octets if any (spare in current spec)
	for fx {
		data, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading warning/error extension: %w", err)
		}
		bytesRead++
		fx = (data & 0x01) != 0
	}

	return bytesRead, nil
}

func (w *WarningErrorConditions) Encode(buf *bytes.Buffer) (int, error) {
	bytesWritten := 0

	// First octet
	octet1 := (w.WE & 0x7F) << 1
	// FX = 0 (no extensions in current spec)

	err := buf.WriteByte(octet1)
	if err != nil {
		return bytesWritten, fmt.Errorf("writing warning/error conditions: %w", err)
	}
	bytesWritten++

	return bytesWritten, nil
}

func (w *WarningErrorConditions) Validate() error {
	if w.WE > 127 {
		return fmt.Errorf("warning/error value out of range: %d", w.WE)
	}
	return nil
}

func (w *WarningErrorConditions) String() string {
	if w.WE == 0 {
		return "No warning/error"
	}
	return fmt.Sprintf("Warning/Error: %d", w.WE)
}
