// cat/cat020/dataitems/v15/flight_level.go
package v15

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// FlightLevel implements I020/090 - Flight Level in Binary Representation
type FlightLevel struct {
	V           bool    // Validated
	G           bool    // Garbled
	FlightLevel float64 // Flight level in 1/4 FL increments
}

func (f *FlightLevel) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading flight level: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("%w: need 2 bytes for flight level, have %d", asterix.ErrBufferTooShort, n)
	}

	// Bit 16: V (validated)
	// Bit 15: G (garbled)
	// Bits 14-1: Flight level (14-bit signed, LSB = 1/4 FL)
	f.V = (data[0] & 0x80) != 0
	f.G = (data[0] & 0x40) != 0

	// Extract 14-bit value
	rawFL := int16((uint16(data[0]&0x3F) << 8) | uint16(data[1]))

	// Sign extend from 14 bits to 16 bits
	if (rawFL & 0x2000) != 0 { // Check bit 13
		rawFL |= ^0x3FFF // Set upper bits for negative
	}

	f.FlightLevel = float64(rawFL) / 4.0

	return n, nil
}

func (f *FlightLevel) Encode(buf *bytes.Buffer) (int, error) {
	// Convert to 1/4 FL units
	rawFL := int16(f.FlightLevel * 4.0)

	// Mask to 14 bits
	rawUnsigned := uint16(rawFL) & 0x3FFF

	var data [2]byte
	if f.V {
		data[0] |= 0x80 // Bit 16: V
	}
	if f.G {
		data[0] |= 0x40 // Bit 15: G
	}
	data[0] |= byte((rawUnsigned >> 8) & 0x3F)
	data[1] = byte(rawUnsigned & 0xFF)

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing flight level: %w", err)
	}
	return n, nil
}

func (f *FlightLevel) Validate() error {
	return nil
}

func (f *FlightLevel) String() string {
	flags := ""
	if !f.V {
		flags += " NOT_VALIDATED"
	}
	if f.G {
		flags += " GARBLED"
	}
	return fmt.Sprintf("FL%.2f%s", f.FlightLevel, flags)
}
