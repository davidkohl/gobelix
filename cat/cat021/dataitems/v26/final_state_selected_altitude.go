// dataitems/cat021/final_state_selected_altitude.go
package v26

import (
	"bytes"
	"fmt"
	"math"

	"github.com/davidkohl/gobelix/asterix"
)

// FinalStateSelectedAltitude implements I021/148
// Final State Selected Altitude (2 octets)
type FinalStateSelectedAltitude struct {
	MV  bool  // Manage Vertical Mode active
	AH  bool  // Altitude Hold Mode active
	AM  bool  // Approach Mode active
	Alt float64 // Altitude in feet (13-bit raw, LSB 25 ft: range ±102,375 ft)
}

func (f *FinalStateSelectedAltitude) Encode(buf *bytes.Buffer) (int, error) {
	// Convert from feet to raw value: LSB = 25 ft (13-bit signed raw — the
	// feet value itself does not fit int16 above 32767 ft, hence float64)
	r := math.Round(f.Alt / 25.0)
	if r < -4096 || r > 4095 {
		return 0, fmt.Errorf("selected altitude %f ft out of 13-bit range", f.Alt)
	}
	raw := int16(r)

	var data [2]byte

	if f.MV {
		data[0] |= 0x80 // Bit 16: MV
	}
	if f.AH {
		data[0] |= 0x40 // Bit 15: AH
	}
	if f.AM {
		data[0] |= 0x20 // Bit 14: AM
	}

	// Encode 13-bit signed value (bits 13-1)
	// Two's complement is automatically handled by uint16 cast
	rawUnsigned := uint16(raw) & 0x1FFF
	data[0] |= byte((rawUnsigned >> 8) & 0x1F) // Upper 5 bits
	data[1] = byte(rawUnsigned)                 // Lower 8 bits

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing final state selected altitude: %w", err)
	}
	return n, nil
}

func (f *FinalStateSelectedAltitude) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading final state selected altitude: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("%w: need 2 bytes for final state selected altitude, have %d", asterix.ErrBufferTooShort, n)
	}

	f.MV = (data[0] & 0x80) != 0
	f.AH = (data[0] & 0x40) != 0
	f.AM = (data[0] & 0x20) != 0

	// Extract 13-bit value
	rawVal := int16((uint16(data[0]&0x1F) << 8) | uint16(data[1]))

	// Sign extend from 13 bits to 16 bits
	if (rawVal & 0x1000) != 0 { // Check bit 12 (13th bit, 0-indexed)
		rawVal |= ^0x1FFF // Set upper bits to 1 for negative values
	}

	// Convert to feet
	f.Alt = float64(rawVal) * 25

	return n, nil
}

func (f *FinalStateSelectedAltitude) Validate() error {
	return nil
}

func (f *FinalStateSelectedAltitude) String() string {
	modes := []string{}
	if f.MV {
		modes = append(modes, "MV")
	}
	if f.AH {
		modes = append(modes, "AH")
	}
	if f.AM {
		modes = append(modes, "AM")
	}

	if len(modes) > 0 {
		return fmt.Sprintf("%.0fft [%s]", f.Alt, fmt.Sprint(modes))
	}
	return fmt.Sprintf("%.0fft", f.Alt)
}
