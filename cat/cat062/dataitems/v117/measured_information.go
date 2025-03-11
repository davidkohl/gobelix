// dataitems/cat062/measured_information.go
package v117

import (
	"bytes"
	"fmt"
)

// MeasuredInformation implements I062/340
// All measured data related to the last report used to update the track
type MeasuredInformation struct {
	Data []byte
}

func (m *MeasuredInformation) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0
	m.Data = nil

	// Primary subfield
	primaryByte := make([]byte, 1)
	n, err := buf.Read(primaryByte)
	if err != nil {
		return n, fmt.Errorf("reading measured information primary subfield: %w", err)
	}
	bytesRead += n
	m.Data = append(m.Data, primaryByte[0])

	// SID: bit-8 Subfield #1: Sensor Identification
	if (primaryByte[0] & 0x80) != 0 {
		subfieldData := make([]byte, 2)
		n, err := buf.Read(subfieldData)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading sensor identification subfield: %w", err)
		}
		bytesRead += n
		m.Data = append(m.Data, subfieldData...)
	}

	// POS: bit-7 Subfield #2: Measured Position
	if (primaryByte[0] & 0x40) != 0 {
		subfieldData := make([]byte, 4)
		n, err := buf.Read(subfieldData)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading measured position subfield: %w", err)
		}
		bytesRead += n
		m.Data = append(m.Data, subfieldData...)
	}

	// HEI: bit-6 Subfield #3: Measured 3-D Height
	if (primaryByte[0] & 0x20) != 0 {
		subfieldData := make([]byte, 2)
		n, err := buf.Read(subfieldData)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading measured 3-D height subfield: %w", err)
		}
		bytesRead += n
		m.Data = append(m.Data, subfieldData...)
	}

	// MDC: bit-5 Subfield #4: Last Measured Mode C code
	if (primaryByte[0] & 0x10) != 0 {
		subfieldData := make([]byte, 2)
		n, err := buf.Read(subfieldData)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading last measured Mode C code subfield: %w", err)
		}
		bytesRead += n
		m.Data = append(m.Data, subfieldData...)
	}

	// MDA: bit-4 Subfield #5: Last Measured Mode 3/A code
	if (primaryByte[0] & 0x08) != 0 {
		subfieldData := make([]byte, 2)
		n, err := buf.Read(subfieldData)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading last measured Mode 3/A code subfield: %w", err)
		}
		bytesRead += n
		m.Data = append(m.Data, subfieldData...)
	}

	// TYP: bit-3 Subfield #6: Report Type
	if (primaryByte[0] & 0x04) != 0 {
		subfieldData := make([]byte, 1)
		n, err := buf.Read(subfieldData)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading report type subfield: %w", err)
		}
		bytesRead += n
		m.Data = append(m.Data, subfieldData...)
	}

	// FX: bit-1 Extension Indicator - not used in this implementation but checked for completeness
	if (primaryByte[0] & 0x01) != 0 {
		// The specification doesn't define any further extensions
		// But we'll add a warning in the logs
		fmt.Println("Warning: Unexpected extension in Measured Information")
	}

	return bytesRead, nil
}

func (m *MeasuredInformation) Encode(buf *bytes.Buffer) (int, error) {
	if len(m.Data) == 0 {
		// If no data, encode a minimal valid value
		return buf.Write([]byte{0})
	}
	return buf.Write(m.Data)
}

func (m *MeasuredInformation) String() string {
	return fmt.Sprintf("MeasuredInformation[%d bytes]", len(m.Data))
}

func (m *MeasuredInformation) Validate() error {
	return nil
}
