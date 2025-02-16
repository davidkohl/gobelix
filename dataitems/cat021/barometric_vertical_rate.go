// dataitems/cat021/barometric_vertical_rate.go
package cat021

import (
	"bytes"
	"fmt"
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
	// Convert back to raw value (divide by 6.25)
	rawRate := b.Rate / 6
	if b.Rate%6 >= 3 { // Round to nearest
		rawRate++
	}

	if !b.RE && (rawRate < -16384 || rawRate > 16383) {
		return 0, fmt.Errorf("rate out of range without RE flag: %d", b.Rate)
	}

	data := make([]byte, 2)
	if b.RE {
		data[0] |= 0x80
	}

	var rawVal uint16
	if rawRate < 0 {
		rawVal = uint16(0x4000 | (uint16(-rawRate) & 0x3FFF))
	} else {
		rawVal = uint16(rawRate & 0x3FFF)
	}

	data[0] |= byte(rawVal >> 8)
	data[1] = byte(rawVal)

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing barometric vertical rate: %w", err)
	}
	return n, nil
}

func (b *BarometricVerticalRate) Validate() error {
	return nil
}
