// dataitems/cat021/barometric_vertical_rate.go
package v26

import (
	"bytes"
	"fmt"
	"math"
)

// BarometricVerticalRate implements I021/155
type BarometricVerticalRate struct {
	RE   bool  // Range Exceeded Indicator
	Rate int16 // Rate in feet/minute
}

func (b *BarometricVerticalRate) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading barometric vertical rate: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("insufficient data for barometric vertical rate: got %d bytes, want 2", n)
	}

	b.RE = (data[0] & 0x80) != 0 // bit 16

	// Combine bytes into a raw value, masking the RE bit
	rawVal := uint16(data[0]&0x7F)<<8 | uint16(data[1])

	// Convert to signed int16
	var raw int16
	if (rawVal & 0x4000) != 0 {
		// Negative number
		raw = -int16(0x4000 - (rawVal & 0x3FFF))
	} else {
		// Positive number
		raw = int16(rawVal)
	}

	// Convert to feet/minute
	b.Rate = raw * 6        // Using 6 instead of 6.25 to avoid floating point truncation
	b.Rate += (raw * 1) / 4 // Add the 0.25 component carefully
	return n, nil
}

func (b *BarometricVerticalRate) Encode(buf *bytes.Buffer) (int, error) {
	rawRateFloat := float64(b.Rate) / 6.25
	rawRate := int16(math.Round(rawRateFloat))

	// Check if the value is within range (-16384 to +16383)
	if !b.RE && (rawRate < -16384 || rawRate > 16383) {
		return 0, fmt.Errorf("rate out of range without RE flag: %d", b.Rate)
	}

	// Create 2-byte buffer for the data
	data := make([]byte, 2)

	rawVal := uint16(rawRate & 0x7FFF)

	if b.RE {
		data[0] = 0x80
	} else {
		data[0] = 0x00
	}

	if rawRate < 0 {

		data[0] |= 0x40 | byte((rawVal>>8)&0x3F)
	} else {
		data[0] |= byte((rawVal >> 8) & 0x3F)
	}

	data[1] = byte(rawVal & 0xFF)

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing barometric vertical rate: %w", err)
	}
	return n, nil
}

func (b *BarometricVerticalRate) Validate() error {
	return nil
}

func (b *BarometricVerticalRate) String() string {
	if b.RE {
		return fmt.Sprintf(">%dft/min", b.Rate)
	}
	return fmt.Sprintf("%dft/min", b.Rate)
}
