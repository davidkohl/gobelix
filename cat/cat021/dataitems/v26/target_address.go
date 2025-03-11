// dataitems/cat021/target_address.go
package v26

import (
	"bytes"
	"fmt"
)

// TargetAddress implements I021/080
// Contains the ICAO 24-bit aircraft address
type TargetAddress struct {
	Address uint32 // 24-bit ICAO address
}

func (t *TargetAddress) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	// Write 24-bit address as 3 bytes
	b := make([]byte, 3)
	b[0] = byte(t.Address >> 16)
	b[1] = byte(t.Address >> 8)
	b[2] = byte(t.Address)

	n, err := buf.Write(b)
	if err != nil {
		return n, fmt.Errorf("writing target address: %w", err)
	}
	return n, nil
}

func (t *TargetAddress) Decode(buf *bytes.Buffer) (int, error) {
	b := make([]byte, 3)
	n, err := buf.Read(b)
	if err != nil {
		return n, fmt.Errorf("reading target address: %w", err)
	}
	if n != 3 {
		return n, fmt.Errorf("incomplete target address data: got %d bytes, want 3", n)
	}

	t.Address = uint32(b[0])<<16 | uint32(b[1])<<8 | uint32(b[2])
	return n, t.Validate()
}

func (t *TargetAddress) Validate() error {
	// Check that address fits in 24 bits
	if t.Address > 0xFFFFFF {
		return fmt.Errorf("invalid target address: exceeds 24 bits")
	}
	return nil
}

// String returns the ICAO address in hex format
func (t *TargetAddress) String() string {
	return fmt.Sprintf("%06X", t.Address)
}

// FromString sets the address from a hex string
func (t *TargetAddress) FromString(s string) error {
	var addr uint32
	_, err := fmt.Sscanf(s, "%x", &addr)
	if err != nil {
		return fmt.Errorf("parsing target address: %w", err)
	}
	t.Address = addr
	return t.Validate()
}
