// dataitems/cat021/geometric_vertical_rate.go
package v26

import (
	"bytes"
	"fmt"
	"math"

	"github.com/davidkohl/gobelix/asterix"
)

// GeometricVerticalRate implements I021/157
// Geometric Vertical Rate (2 octets)
type GeometricVerticalRate struct {
	RE   bool  // Range Exceeded Indicator
	Rate int16 // Rate in feet/minute
}

func (g *GeometricVerticalRate) Encode(buf *bytes.Buffer) (int, error) {
	// Convert from feet/minute to raw value: LSB = 6.25 ft/min
	raw := int16(math.Round(float64(g.Rate) / 6.25))

	// Check range
	if !g.RE && (raw < -16384 || raw > 16383) {
		return 0, fmt.Errorf("rate out of range without RE flag: %d", g.Rate)
	}

	var data [2]byte

	// Convert signed to unsigned representation
	rawVal := uint16(raw & 0x7FFF)

	if g.RE {
		data[0] = 0x80
	}

	if raw < 0 {
		data[0] |= 0x40 | byte((rawVal>>8)&0x3F)
	} else {
		data[0] |= byte((rawVal >> 8) & 0x3F)
	}

	data[1] = byte(rawVal)

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing geometric vertical rate: %w", err)
	}
	return n, nil
}

func (g *GeometricVerticalRate) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading geometric vertical rate: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("%w: need 2 bytes for geometric vertical rate, have %d", asterix.ErrBufferTooShort, n)
	}

	g.RE = (data[0] & 0x80) != 0

	// Extract the 15-bit value
	rawVal := uint16(data[0]&0x7F)<<8 | uint16(data[1])

	// Convert to signed int16
	var raw int16
	if (rawVal & 0x4000) != 0 {
		// Negative number (two's complement)
		raw = -int16(0x4000 - (rawVal & 0x3FFF))
	} else {
		// Positive number
		raw = int16(rawVal)
	}

	// Convert to feet/minute
	g.Rate = int16(math.Round(float64(raw) * 6.25))

	return n, nil
}

func (g *GeometricVerticalRate) Validate() error {
	return nil
}

func (g *GeometricVerticalRate) String() string {
	if g.RE {
		return fmt.Sprintf(">%dft/min", g.Rate)
	}
	return fmt.Sprintf("%dft/min", g.Rate)
}
