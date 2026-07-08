// dataitems/cat021/selected_altitude.go
package v26

import (
	"bytes"
	"fmt"
	"math"

	"github.com/davidkohl/gobelix/asterix"
)

// SelectedAltitude implements I021/146
// Selected Altitude (Mode Control Panel/Flight Control Unit) (2 octets)
type SelectedAltitude struct {
	SAS    bool  // Source information provided
	Source uint8 // 0=Unknown, 1=Aircraft, 2=FCU/MCP, 3=FMS
	Alt    float64 // Altitude in feet (13-bit raw, LSB 25 ft: range ±102,375 ft)
}

func (s *SelectedAltitude) Encode(buf *bytes.Buffer) (int, error) {
	if err := s.Validate(); err != nil {
		return 0, err
	}

	// Convert from feet to raw value: LSB = 25 ft (13-bit signed raw — the
	// feet value itself does not fit int16 above 32767 ft, hence float64)
	r := math.Round(s.Alt / 25.0)
	if r < -4096 || r > 4095 {
		return 0, fmt.Errorf("selected altitude %f ft out of 13-bit range", s.Alt)
	}
	raw := int16(r)

	var data [2]byte

	if s.SAS {
		data[0] |= 0x80 // Bit 16: SAS
	}

	// Bits 15-14: Source (2 bits)
	data[0] |= (s.Source & 0x03) << 5

	// Encode 13-bit signed value (bits 13-1)
	// Two's complement is automatically handled by uint16 cast
	rawUnsigned := uint16(raw) & 0x1FFF
	data[0] |= byte((rawUnsigned >> 8) & 0x1F) // Upper 5 bits
	data[1] = byte(rawUnsigned)                 // Lower 8 bits

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing selected altitude: %w", err)
	}
	return n, nil
}

func (s *SelectedAltitude) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading selected altitude: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("%w: need 2 bytes for selected altitude, have %d", asterix.ErrBufferTooShort, n)
	}

	s.SAS = (data[0] & 0x80) != 0
	s.Source = (data[0] >> 5) & 0x03

	// Extract 13-bit value
	rawVal := int16((uint16(data[0]&0x1F) << 8) | uint16(data[1]))

	// Sign extend from 13 bits to 16 bits
	if (rawVal & 0x1000) != 0 { // Check bit 12 (13th bit, 0-indexed)
		rawVal |= ^0x1FFF // Set upper bits to 1 for negative values
	}

	// Convert to feet
	s.Alt = float64(rawVal) * 25

	return n, s.Validate()
}

func (s *SelectedAltitude) Validate() error {
	if s.Source > 3 {
		return fmt.Errorf("invalid source: %d", s.Source)
	}
	return nil
}

func (s *SelectedAltitude) String() string {
	sourceStr := ""
	switch s.Source {
	case 0:
		sourceStr = "Unknown"
	case 1:
		sourceStr = "Aircraft"
	case 2:
		sourceStr = "FCU/MCP"
	case 3:
		sourceStr = "FMS"
	}

	if s.SAS {
		return fmt.Sprintf("%.0fft (%s)", s.Alt, sourceStr)
	}
	return fmt.Sprintf("%.0fft", s.Alt)
}
