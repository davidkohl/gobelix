// dataitems/cat048/aircraft_address.go
package v132

import (
	"bytes"
	"fmt"
)

// AircraftAddress implements I048/220
// Aircraft address (24-bits Mode S address) assigned uniquely to each aircraft.
type AircraftAddress struct {
	Address uint32 // 24-bit Mode S address
}

// Decode implements the DataItem interface
func (a *AircraftAddress) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 3)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading aircraft address: %w", err)
	}
	if n != 3 {
		return n, fmt.Errorf("insufficient data for aircraft address: got %d bytes, want 3", n)
	}

	// 24-bit address as 3 bytes
	a.Address = uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2])

	return n, a.Validate()
}

// Encode implements the DataItem interface
func (a *AircraftAddress) Encode(buf *bytes.Buffer) (int, error) {
	if err := a.Validate(); err != nil {
		return 0, err
	}

	data := make([]byte, 3)
	data[0] = byte(a.Address >> 16) // Most significant 8 bits
	data[1] = byte(a.Address >> 8)  // Middle 8 bits
	data[2] = byte(a.Address)       // Least significant 8 bits

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing aircraft address: %w", err)
	}
	return n, nil
}

// Validate implements the DataItem interface
func (a *AircraftAddress) Validate() error {
	if a.Address > 0xFFFFFF { // 2^24 - 1
		return fmt.Errorf("aircraft address exceeds 24 bits: %X", a.Address)
	}
	return nil
}

// String returns a human-readable representation
func (a *AircraftAddress) String() string {
	return fmt.Sprintf("%06X", a.Address)
}

// FromString sets the aircraft address from a hexadecimal string
func (a *AircraftAddress) FromString(s string) error {
	_, err := fmt.Sscanf(s, "%x", &a.Address)
	if err != nil {
		return fmt.Errorf("invalid aircraft address format: %w", err)
	}
	return a.Validate()
}
