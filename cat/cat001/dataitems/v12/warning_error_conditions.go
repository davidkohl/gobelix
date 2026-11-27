package v12

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// WarningErrorConditions represents I001/030 - Warning/Error Conditions
// (CAT001 ed 1.4 §5.2.3). Variable length: each octet carries one 7-bit W/E
// value in bits 8-2 and FX in bit 1.
type WarningErrorConditions struct {
	// W is the first W/E value. Kept for compatibility; it equals Values[0]
	// after Decode. Encode uses it only when Values is empty. When both are
	// set (W non-zero and Values non-empty) they must agree, W == Values[0],
	// or Validate and Encode return an error.
	W uint8
	// Values holds every W/E value in transmission order.
	Values []uint8
}

func (w *WarningErrorConditions) Decode(buf *bytes.Buffer) (int, error) {
	values, n, err := decodeWEValues(buf)
	if err != nil {
		return n, err
	}
	w.Values = values
	w.W = values[0]
	return n, nil
}

func (w *WarningErrorConditions) Encode(buf *bytes.Buffer) (int, error) {
	if err := w.Validate(); err != nil {
		return 0, err
	}
	values := w.Values
	if len(values) == 0 {
		values = []uint8{w.W}
	}
	return encodeWEValues(buf, values)
}

func (w *WarningErrorConditions) String() string {
	values := w.Values
	if len(values) == 0 {
		values = []uint8{w.W}
	}
	if len(values) == 1 && values[0] == 0 {
		return "No warning/error"
	}
	return fmt.Sprintf("Warning/Error: %v", values)
}

func (w *WarningErrorConditions) Validate() error {
	if w.W > 127 {
		return fmt.Errorf("%w: W/E value is 7 bits, got %d", asterix.ErrInvalidMessage, w.W)
	}
	for i, v := range w.Values {
		if v > 127 {
			return fmt.Errorf("%w: W/E value %d is 7 bits, got %d", asterix.ErrInvalidMessage, i, v)
		}
	}
	if len(w.Values) > 0 && w.W != 0 && w.W != w.Values[0] {
		return fmt.Errorf("%w: I001/030 W %d disagrees with Values[0] %d", asterix.ErrInvalidMessage, w.W, w.Values[0])
	}
	return nil
}

// decodeWEValues reads a chain of FX-extended octets, each carrying a 7-bit
// value in bits 8-2. It always returns at least one value on success.
func decodeWEValues(buf *bytes.Buffer) ([]uint8, int, error) {
	var values []uint8
	n := 0
	for {
		b, err := buf.ReadByte()
		if err != nil {
			return nil, n, fmt.Errorf("%w: incomplete warning/error conditions", asterix.ErrBufferTooShort)
		}
		n++
		values = append(values, b>>1)
		if b&0x01 == 0 {
			return values, n, nil
		}
	}
}

// encodeWEValues writes values as a chain of FX-extended octets.
func encodeWEValues(buf *bytes.Buffer, values []uint8) (int, error) {
	for i, v := range values {
		b := (v & 0x7F) << 1
		if i < len(values)-1 {
			b |= 0x01
		}
		if err := buf.WriteByte(b); err != nil {
			return i, fmt.Errorf("writing warning/error conditions: %w", err)
		}
	}
	return len(values), nil
}
