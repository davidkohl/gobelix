// cat/cat020/dataitems/v10/flight_level.go
package v10

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// FlightLevel represents I020/090 - Flight Level in Binary Representation
// Fixed length: 2 bytes
// Flight Level (Mode S Altitude) converted into binary two's complement representation
type FlightLevel struct {
	V           bool    // Code validated (false) / Code not validated (true)
	G           bool    // Default (false) / Garbled code (true)
	FlightLevel float64 // Flight Level in FL
}

// NewFlightLevel creates a new Flight Level data item
func NewFlightLevel() *FlightLevel {
	return &FlightLevel{}
}

// Decode decodes the Flight Level from bytes
func (f *FlightLevel) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 2 {
		return 0, fmt.Errorf("%w: need 2 bytes, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	data := buf.Next(2)
	value := binary.BigEndian.Uint16(data)

	f.V = (value & 0x8000) != 0
	f.G = (value & 0x4000) != 0

	// Bits 14-1 contain flight level in two's complement, LSB = 1/4 FL
	flRaw := int16(value & 0x3FFF)
	// Sign extend from 14 bits
	if flRaw&0x2000 != 0 {
		flRaw |= ^int16(0x3FFF)
	}
	f.FlightLevel = float64(flRaw) * 0.25

	return 2, nil
}

// Encode encodes the Flight Level to bytes
func (f *FlightLevel) Encode(buf *bytes.Buffer) (int, error) {
	if err := f.Validate(); err != nil {
		return 0, err
	}

	// Convert flight level to raw value (LSB = 1/4 FL)
	flRaw := int16(f.FlightLevel / 0.25)

	var value uint16
	if f.V {
		value |= 0x8000
	}
	if f.G {
		value |= 0x4000
	}
	value |= uint16(flRaw) & 0x3FFF

	if err := binary.Write(buf, binary.BigEndian, value); err != nil {
		return 0, fmt.Errorf("writing flight level: %w", err)
	}

	return 2, nil
}

// Validate validates the Flight Level
func (f *FlightLevel) Validate() error {
	// Range check based on 14-bit two's complement with LSB = 1/4 FL
	// Range: -2048 to +2047.75 FL
	if f.FlightLevel < -2048.0 || f.FlightLevel > 2047.75 {
		return fmt.Errorf("%w: flight level must be in range [-2048, 2047.75], got %.2f", asterix.ErrInvalidMessage, f.FlightLevel)
	}
	return nil
}

// String returns a string representation
func (f *FlightLevel) String() string {
	status := ""
	if f.V {
		status += "V"
	}
	if f.G {
		status += "G"
	}
	if status == "" {
		status = "OK"
	}
	return fmt.Sprintf("FL%.2f (%s)", f.FlightLevel, status)
}
