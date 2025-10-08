// cat/cat020/dataitems/v15/mode_s_mb_data.go
package v15

import (
	"bytes"
	"fmt"
)

// ModeSMBData implements I020/250 - Mode S MB Data
// This is a repetitive data item where each repetition is 8 octets (BDS register)
type ModeSMBData struct {
	MBData [][]byte // Each entry is 8 bytes
}

// Decode reads the repetitive BDS register data
func (m *ModeSMBData) Decode(buf *bytes.Buffer) (int, error) {
	// Read REP byte
	rep, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading I020/250 REP: %w", err)
	}
	bytesRead := 1

	// Read repetitions (each is 8 bytes)
	m.MBData = make([][]byte, rep)
	for i := 0; i < int(rep); i++ {
		data := make([]byte, 8)
		n, err := buf.Read(data)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading I020/250 BDS register %d: %w", i, err)
		}
		if n != 8 {
			return bytesRead + n, fmt.Errorf("I020/250 BDS register %d: expected 8 bytes, got %d", i, n)
		}
		m.MBData[i] = data
		bytesRead += 8
	}

	return bytesRead, nil
}

// Encode writes the repetitive BDS register data
func (m *ModeSMBData) Encode(buf *bytes.Buffer) (int, error) {
	if len(m.MBData) > 255 {
		return 0, fmt.Errorf("I020/250: too many BDS registers (%d > 255)", len(m.MBData))
	}

	// Write REP byte
	err := buf.WriteByte(byte(len(m.MBData)))
	if err != nil {
		return 0, fmt.Errorf("writing I020/250 REP: %w", err)
	}
	bytesWritten := 1

	// Write each BDS register
	for i, data := range m.MBData {
		if len(data) != 8 {
			return bytesWritten, fmt.Errorf("I020/250 BDS register %d: invalid length %d (must be 8)", i, len(data))
		}
		n, err := buf.Write(data)
		if err != nil {
			return bytesWritten + n, fmt.Errorf("writing I020/250 BDS register %d: %w", i, err)
		}
		bytesWritten += n
	}

	return bytesWritten, nil
}

// Validate implements the DataItem interface
func (m *ModeSMBData) Validate() error {
	for i, data := range m.MBData {
		if len(data) != 8 {
			return fmt.Errorf("I020/250 BDS register %d: invalid length %d (must be 8)", i, len(data))
		}
	}
	return nil
}

// String returns a string representation
func (m *ModeSMBData) String() string {
	if len(m.MBData) == 0 {
		return "Mode S MB Data: (none)"
	}
	return fmt.Sprintf("Mode S MB Data: %d BDS registers", len(m.MBData))
}
