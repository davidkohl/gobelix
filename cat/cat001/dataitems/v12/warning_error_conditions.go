package v12

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// WarningErrorConditions represents I001/030 - Warning/Error Conditions
// Extended item with various warning/error flags
type WarningErrorConditions struct {
	W uint8 // Warning/error value
}

func (w *WarningErrorConditions) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	if buf.Len() < 1 {
		return 0, fmt.Errorf("%w: need at least 1 byte for warning/error conditions", asterix.ErrBufferTooShort)
	}

	data := buf.Next(1)
	bytesRead++

	// First octet: bits 8-2 = warning/error value, bit 1 = FX
	w.W = (data[0] >> 1) & 0x7F

	// Check FX bit for extension
	hasFX := (data[0] & 0x01) != 0

	// Handle extensions if present
	for hasFX {
		if buf.Len() < 1 {
			return bytesRead, fmt.Errorf("%w: incomplete warning/error conditions extension", asterix.ErrBufferTooShort)
		}
		data = buf.Next(1)
		bytesRead++
		hasFX = (data[0] & 0x01) != 0
	}

	return bytesRead, nil
}

func (w *WarningErrorConditions) Encode(buf *bytes.Buffer) (int, error) {
	// First octet: warning value in bits 8-2, no FX
	octet := (w.W & 0x7F) << 1
	buf.WriteByte(octet)
	return 1, nil
}

func (w *WarningErrorConditions) String() string {
	if w.W == 0 {
		return "No warning/error"
	}
	return fmt.Sprintf("Warning/Error: %d", w.W)
}

func (w *WarningErrorConditions) Validate() error {
	return nil
}
