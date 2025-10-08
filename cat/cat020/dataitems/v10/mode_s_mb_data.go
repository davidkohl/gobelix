// cat/cat020/dataitems/v10/mode_s_mb_data.go
package v10

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// ModeSMBData represents I020/250 - Mode S MB Data
// Repetitive data item: 1+8n octets
// Mode S Comm B data as extracted from the aircraft transponder
type ModeSMBData struct {
	Reports []ModeSMBReport // Mode S MB reports
}

// ModeSMBReport represents a single Mode S MB report (8 bytes)
type ModeSMBReport struct {
	MBData [7]byte // 56-bit message conveying Mode S Comm B message data
	BDS1   uint8   // Comm B Data Buffer Store 1 Address (4 bits)
	BDS2   uint8   // Comm B Data Buffer Store 2 Address (4 bits)
}

// NewModeSMBData creates a new Mode S MB Data item
func NewModeSMBData() *ModeSMBData {
	return &ModeSMBData{}
}

// Decode decodes the Mode S MB Data from bytes
func (m *ModeSMBData) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 1 {
		return 0, fmt.Errorf("%w: need at least 1 byte, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	bytesRead := 0

	// Read repetition factor
	rep := buf.Next(1)
	bytesRead++
	repCount := int(rep[0])

	if repCount == 0 {
		return bytesRead, nil
	}

	// Each report is 8 bytes
	if buf.Len() < repCount*8 {
		return bytesRead, fmt.Errorf("%w: need %d bytes for MB reports, have %d", asterix.ErrBufferTooShort, repCount*8, buf.Len())
	}

	m.Reports = make([]ModeSMBReport, repCount)

	for i := 0; i < repCount; i++ {
		data := buf.Next(8)
		bytesRead += 8

		// Copy 7 bytes of MB data
		copy(m.Reports[i].MBData[:], data[0:7])

		// Last byte contains BDS1 and BDS2
		m.Reports[i].BDS1 = (data[7] >> 4) & 0x0F
		m.Reports[i].BDS2 = data[7] & 0x0F
	}

	return bytesRead, nil
}

// Encode encodes the Mode S MB Data to bytes
func (m *ModeSMBData) Encode(buf *bytes.Buffer) (int, error) {
	if err := m.Validate(); err != nil {
		return 0, err
	}

	bytesWritten := 0

	// Write repetition factor
	repCount := byte(len(m.Reports))
	if err := buf.WriteByte(repCount); err != nil {
		return bytesWritten, fmt.Errorf("writing repetition factor: %w", err)
	}
	bytesWritten++

	if repCount == 0 {
		return bytesWritten, nil
	}

	// Write reports
	for i := range m.Reports {
		// Write 7 bytes of MB data
		n, err := buf.Write(m.Reports[i].MBData[:])
		bytesWritten += n
		if err != nil {
			return bytesWritten, fmt.Errorf("writing MB data: %w", err)
		}

		// Write BDS codes
		bdsByte := ((m.Reports[i].BDS1 & 0x0F) << 4) | (m.Reports[i].BDS2 & 0x0F)
		if err := buf.WriteByte(bdsByte); err != nil {
			return bytesWritten, fmt.Errorf("writing BDS codes: %w", err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

// Validate validates the Mode S MB Data
func (m *ModeSMBData) Validate() error {
	if len(m.Reports) > 255 {
		return fmt.Errorf("%w: too many reports, max 255, got %d", asterix.ErrInvalidMessage, len(m.Reports))
	}

	for i := range m.Reports {
		if m.Reports[i].BDS1 > 15 {
			return fmt.Errorf("%w: BDS1 must be 0-15, got %d", asterix.ErrInvalidMessage, m.Reports[i].BDS1)
		}
		if m.Reports[i].BDS2 > 15 {
			return fmt.Errorf("%w: BDS2 must be 0-15, got %d", asterix.ErrInvalidMessage, m.Reports[i].BDS2)
		}
	}

	return nil
}

// String returns a string representation
func (m *ModeSMBData) String() string {
	if len(m.Reports) == 0 {
		return "No MB reports"
	}

	return fmt.Sprintf("%d MB report(s)", len(m.Reports))
}
