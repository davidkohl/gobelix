package v10

import (
	"bytes"
	"fmt"
)

// WarningErrorConditions represents I002/080 - Warning/Error Conditions
// (CAT002 ed 1.2 §5.2.9). Variable length: each octet carries one 7-bit W/E
// value (bits 8-2) and FX (bit 1). Values 1-63 are reserved for common
// standard use, 64-127 are application dependent.
type WarningErrorConditions struct {
	Values []uint8 // W/E values in transmission order (at least one)
}

func (w *WarningErrorConditions) Decode(buf *bytes.Buffer) (int, error) {
	values, n, err := decodeFXChain(buf, "warning/error conditions")
	if err != nil {
		return n, err
	}
	w.Values = values
	return n, nil
}

func (w *WarningErrorConditions) Encode(buf *bytes.Buffer) (int, error) {
	if err := w.Validate(); err != nil {
		return 0, err
	}
	return encodeFXChain(buf, w.Values, "warning/error conditions")
}

func (w *WarningErrorConditions) Validate() error {
	if len(w.Values) == 0 {
		return fmt.Errorf("I002/080 needs at least one W/E value: %w", errEmpty)
	}
	return check7Bit(w.Values, "warning/error conditions")
}

func (w *WarningErrorConditions) String() string {
	return fmt.Sprintf("W/E: %v", w.Values)
}
