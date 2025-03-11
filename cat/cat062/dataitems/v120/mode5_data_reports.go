// dataitems/cat062/mode5_data_reports.go
package v120

import (
	"bytes"
	"fmt"
)

// Mode5DataReports implements I062/110
// Mode 5 Data reports & Extended Mode 1 Code
type Mode5DataReports struct {
	Data []byte
}

func (m *Mode5DataReports) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0
	m.Data = nil

	// Read primary subfield byte
	primaryByte := make([]byte, 1)
	n, err := buf.Read(primaryByte)
	if err != nil {
		return n, fmt.Errorf("reading mode 5 data reports primary subfield: %w", err)
	}
	bytesRead += n
	m.Data = append(m.Data, primaryByte[0])

	// Process each subfield based on bits in the primary subfield
	// SUM: bit-8 Subfield #1: Mode 5 Summary
	if (primaryByte[0] & 0x80) != 0 {
		subfieldData := make([]byte, 1)
		n, err := buf.Read(subfieldData)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading mode 5 summary subfield: %w", err)
		}
		bytesRead += n
		m.Data = append(m.Data, subfieldData...)
	}

	// PMN: bit-7 Subfield #2: Mode 5 PIN/ National Origin/Mission Code
	if (primaryByte[0] & 0x40) != 0 {
		subfieldData := make([]byte, 4)
		n, err := buf.Read(subfieldData)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading mode 5 PIN subfield: %w", err)
		}
		bytesRead += n
		m.Data = append(m.Data, subfieldData...)
	}

	// POS: bit-6 Subfield #3: Mode 5 Reported Position
	if (primaryByte[0] & 0x20) != 0 {
		subfieldData := make([]byte, 6)
		n, err := buf.Read(subfieldData)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading mode 5 reported position subfield: %w", err)
		}
		bytesRead += n
		m.Data = append(m.Data, subfieldData...)
	}

	// GA: bit-5 Subfield #4: Mode 5 GNSS-derived Altitude
	if (primaryByte[0] & 0x10) != 0 {
		subfieldData := make([]byte, 2)
		n, err := buf.Read(subfieldData)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading mode 5 GNSS-derived altitude subfield: %w", err)
		}
		bytesRead += n
		m.Data = append(m.Data, subfieldData...)
	}

	// EM1: bit-4 Subfield #5: Extended Mode 1 Code in Octal Representation
	if (primaryByte[0] & 0x08) != 0 {
		subfieldData := make([]byte, 2)
		n, err := buf.Read(subfieldData)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading extended mode 1 code subfield: %w", err)
		}
		bytesRead += n
		m.Data = append(m.Data, subfieldData...)
	}

	// TOS: bit-3 Subfield #6: Time Offset for POS and GA
	if (primaryByte[0] & 0x04) != 0 {
		subfieldData := make([]byte, 1)
		n, err := buf.Read(subfieldData)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading time offset subfield: %w", err)
		}
		bytesRead += n
		m.Data = append(m.Data, subfieldData...)
	}

	// XP: bit-2 Subfield #7: X Pulse Presence
	if (primaryByte[0] & 0x02) != 0 {
		subfieldData := make([]byte, 1)
		n, err := buf.Read(subfieldData)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading X pulse presence subfield: %w", err)
		}
		bytesRead += n
		m.Data = append(m.Data, subfieldData...)
	}

	// FX: bit-1 Extension Indicator - not used in this implementation but checked for completeness
	if (primaryByte[0] & 0x01) != 0 {
		// The specification doesn't define any further extensions
		// But we'll add a warning in the logs
		fmt.Println("Warning: Unexpected extension in Mode 5 Data Reports")
	}

	return bytesRead, nil
}

func (m *Mode5DataReports) Encode(buf *bytes.Buffer) (int, error) {
	if len(m.Data) == 0 {
		// If no data, encode a minimal valid value
		return buf.Write([]byte{0})
	}
	return buf.Write(m.Data)
}

func (m *Mode5DataReports) String() string {
	return fmt.Sprintf("Mode5DataReports[%d bytes]", len(m.Data))
}

// Mode5DataReports Validate method
func (m *Mode5DataReports) Validate() error {
	// Stub implementation
	return nil
}
