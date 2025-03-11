// dataitems/cat062/calculated_track_position_cartesian.go
package v117

import (
	"bytes"
	"fmt"
	"math"
)

// CalculatedTrackPositionCartesian implements I062/100
// Calculated position in Cartesian co-ordinates with a resolution of 0.5m
type CalculatedTrackPositionCartesian struct {
	X float64 // Meters, positive = east
	Y float64 // Meters, positive = north
}

func (p *CalculatedTrackPositionCartesian) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 6)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading cartesian position: %w", err)
	}
	if n != 6 {
		return n, fmt.Errorf("insufficient data for cartesian position: got %d bytes, want 6", n)
	}

	// Extract X (24 bits, two's complement)
	rawX := int32(uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2]))
	// Sign extension for 24-bit two's complement
	if (rawX & 0x800000) != 0 {
		rawX = rawX | int32(-16777216) // -16777216 is -2^24
	}
	p.X = float64(rawX) * 0.5 // LSB = 0.5 meters

	// Extract Y (24 bits, two's complement)
	rawY := int32(uint32(data[3])<<16 | uint32(data[4])<<8 | uint32(data[5]))
	// Sign extension for 24-bit two's complement
	if (rawY & 0x800000) != 0 {
		rawY = rawY | int32(-16777216) // -16777216 is -2^24
	}
	p.Y = float64(rawY) * 0.5 // LSB = 0.5 meters

	return n, p.Validate()
}

func (p *CalculatedTrackPositionCartesian) Encode(buf *bytes.Buffer) (int, error) {
	if err := p.Validate(); err != nil {
		return 0, err
	}

	// Convert X to raw value
	rawX := int32(math.Round(p.X / 0.5))

	// Convert Y to raw value
	rawY := int32(math.Round(p.Y / 0.5))

	data := make([]byte, 6)
	// Encode X (24 bits only)
	data[0] = byte((rawX >> 16) & 0xFF)
	data[1] = byte((rawX >> 8) & 0xFF)
	data[2] = byte(rawX & 0xFF)

	// Encode Y (24 bits only)
	data[3] = byte((rawY >> 16) & 0xFF)
	data[4] = byte((rawY >> 8) & 0xFF)
	data[5] = byte(rawY & 0xFF)

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing cartesian position: %w", err)
	}
	return n, nil
}

func (p *CalculatedTrackPositionCartesian) Validate() error {
	// Check range: The max value for a 24-bit two's complement number is 2^23-1,
	// which translates to (2^23-1)*0.5 meters
	maxValue := (1<<23 - 1) * 0.5
	minValue := -(1 << 23) * 0.5

	if p.X < minValue || p.X > maxValue {
		return fmt.Errorf("the X coordinate out of range [%f,%f]: %f", minValue, maxValue, p.X)
	}
	if p.Y < minValue || p.Y > maxValue {
		return fmt.Errorf("the Y coordinate out of range [%f,%f]: %f", minValue, maxValue, p.Y)
	}
	return nil
}

func (p *CalculatedTrackPositionCartesian) String() string {
	return fmt.Sprintf("X: %.1fm, Y: %.1fm", p.X, p.Y)
}
