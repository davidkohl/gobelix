// dataitems/cat048/flight_level.go
package v132

import (
	"bytes"
	"fmt"
)

// FlightLevel implements I048/090
// Flight Level converted into binary representation.
type FlightLevel struct {
	V     bool    // Code validated
	G     bool    // Garbled code
	Level float64 // Flight Level
}

// Decode implements the DataItem interface
func (f *FlightLevel) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading flight level: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("insufficient data for flight level: got %d bytes, want 2", n)
	}

	f.V = (data[0] & 0x80) != 0 // bit 16
	f.G = (data[0] & 0x40) != 0 // bit 15

	// Extract flight level in two's complement format (bits 14-1)
	// Create a 16-bit value
	rawValue := int16(uint16(data[0]&0x3F)<<8 | uint16(data[1]))

	// Sign extension is handled properly by the int16 cast if bit 14 is set

	// Convert to flight level (1 FL = 100 ft), LSB = 1/4 FL
	f.Level = float64(rawValue) * 0.25

	return n, nil
}

// Encode implements the DataItem interface
func (f *FlightLevel) Encode(buf *bytes.Buffer) (int, error) {
	if err := f.Validate(); err != nil {
		return 0, err
	}

	// Convert flight level to raw value
	rawFL := int16(f.Level / 0.25)

	data := make([]byte, 2)

	// Set flag bits
	if f.V {
		data[0] |= 0x80 // bit 16
	}
	if f.G {
		data[0] |= 0x40 // bit 15
	}

	// Set flight level bits (bits 14-1)
	data[0] |= byte((rawFL >> 8) & 0x3F) // bits 14-9
	data[1] = byte(rawFL)                // bits 8-1

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing flight level: %w", err)
	}
	return n, nil
}

// Validate implements the DataItem interface
func (f *FlightLevel) Validate() error {
	// According to ICAO Annex 10, flight levels should be within a reasonable range
	// Standard range is -15 to 1500 FL (-1500 to 150000 ft)
	if f.Level < -15 || f.Level > 1500 {
		return fmt.Errorf("flight level out of valid range [-15,1500]: %f", f.Level)
	}
	return nil
}

// String returns a human-readable representation
func (f *FlightLevel) String() string {
	flags := ""
	if f.V {
		flags += "V,"
	}
	if f.G {
		flags += "G,"
	}

	if flags != "" {
		flags = flags[:len(flags)-1] + " " // Remove trailing comma
	}

	sign := ""
	if f.Level < 0 {
		sign = "-"
	}

	// Use absolute value and format as integer if whole number, otherwise with decimals
	absLevel := f.Level
	if absLevel < 0 {
		absLevel = -absLevel
	}

	// Format as integer or with decimals depending on whether it's a whole number
	levelStr := ""
	if absLevel == float64(int(absLevel)) {
		levelStr = fmt.Sprintf("%s%.0f", sign, absLevel)
	} else {
		levelStr = fmt.Sprintf("%s%.2f", sign, absLevel)
	}

	return fmt.Sprintf("%sFL%s (%.0f ft)", flags, levelStr, f.Level*100)
}
