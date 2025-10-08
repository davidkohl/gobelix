// cat/cat020/dataitems/v15/position_accuracy.go
package v15

import (
	"bytes"
	"fmt"
)

// PositionAccuracy implements I020/500 - Position Accuracy
// This is a compound data item with optional subfields
type PositionAccuracy struct {
	RawData []byte // For now, just store raw bytes
}

// Decode reads the compound position accuracy data
func (p *PositionAccuracy) Decode(buf *bytes.Buffer) (int, error) {
	// Read primary subfield (1 byte indicating which subfields are present)
	primary, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading I020/500 primary subfield: %w", err)
	}
	bytesRead := 1
	p.RawData = []byte{primary}

	// Check which subfields are present
	// Bit 7: DOP of Position (6 octets)
	// Bit 6: Std Dev of Position (6 octets)
	// Bit 5: Std Dev of Geometric Height (2 octets)
	// Bit 4: Reserved
	// Bit 3: Reserved
	// Bit 2: Reserved
	// Bit 1: Reserved
	// Bit 0: FX (always 0 for now in v1.5)

	// Read DOP of Position (6 bytes) if present
	if primary&0x80 != 0 {
		data := make([]byte, 6)
		n, err := buf.Read(data)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading I020/500 DOP: %w", err)
		}
		if n != 6 {
			return bytesRead + n, fmt.Errorf("I020/500 DOP: expected 6 bytes, got %d", n)
		}
		p.RawData = append(p.RawData, data...)
		bytesRead += 6
	}

	// Read Std Dev of Position (6 bytes) if present
	if primary&0x40 != 0 {
		data := make([]byte, 6)
		n, err := buf.Read(data)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading I020/500 StdDev Position: %w", err)
		}
		if n != 6 {
			return bytesRead + n, fmt.Errorf("I020/500 StdDev Position: expected 6 bytes, got %d", n)
		}
		p.RawData = append(p.RawData, data...)
		bytesRead += 6
	}

	// Read Std Dev of Geometric Height (2 bytes) if present
	if primary&0x20 != 0 {
		data := make([]byte, 2)
		n, err := buf.Read(data)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading I020/500 StdDev Height: %w", err)
		}
		if n != 2 {
			return bytesRead + n, fmt.Errorf("I020/500 StdDev Height: expected 2 bytes, got %d", n)
		}
		p.RawData = append(p.RawData, data...)
		bytesRead += 2
	}

	return bytesRead, nil
}

// Encode writes the compound position accuracy data
func (p *PositionAccuracy) Encode(buf *bytes.Buffer) (int, error) {
	if len(p.RawData) == 0 {
		return 0, fmt.Errorf("I020/500: no data to encode")
	}

	n, err := buf.Write(p.RawData)
	if err != nil {
		return n, fmt.Errorf("writing I020/500: %w", err)
	}
	return n, nil
}

// Validate implements the DataItem interface
func (p *PositionAccuracy) Validate() error {
	if len(p.RawData) == 0 {
		return fmt.Errorf("I020/500: no data")
	}
	return nil
}

// String returns a string representation
func (p *PositionAccuracy) String() string {
	if len(p.RawData) == 0 {
		return "Position Accuracy: (empty)"
	}

	primary := p.RawData[0]
	parts := []string{}

	if primary&0x80 != 0 {
		parts = append(parts, "DOP")
	}
	if primary&0x40 != 0 {
		parts = append(parts, "StdDev(XY)")
	}
	if primary&0x20 != 0 {
		parts = append(parts, "StdDev(H)")
	}

	if len(parts) == 0 {
		return "Position Accuracy: (no subfields)"
	}

	return fmt.Sprintf("Position Accuracy: %v", parts)
}
