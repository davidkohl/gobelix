// cat/common/dataitems/target_address.go
package common

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// TargetAddress implements the 24-bit ICAO aircraft address
// Used in multiple categories: I020/220, I021/080, I048/220, I062/080
// Fixed length: 3 bytes
type TargetAddress struct {
	Address uint32 // 24-bit ICAO address (Mode S transponder address)
}

// Decode decodes the Target Address from bytes
func (t *TargetAddress) Decode(buf *bytes.Buffer) (int, error) {
	var data [3]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading target address: %w", err)
	}
	if n != 3 {
		return n, fmt.Errorf("%w: need 3 bytes for target address, have %d", asterix.ErrBufferTooShort, n)
	}

	t.Address = uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2])

	return n, nil
}

// Encode encodes the Target Address to bytes
func (t *TargetAddress) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	data := []byte{
		byte((t.Address >> 16) & 0xFF),
		byte((t.Address >> 8) & 0xFF),
		byte(t.Address & 0xFF),
	}

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing target address: %w", err)
	}
	return n, nil
}

// Validate validates the Target Address
func (t *TargetAddress) Validate() error {
	if t.Address > 0xFFFFFF {
		return fmt.Errorf("%w: target address exceeds 24 bits: 0x%X", asterix.ErrInvalidMessage, t.Address)
	}
	return nil
}

// String returns the ICAO address in hex format
func (t *TargetAddress) String() string {
	return fmt.Sprintf("%06X", t.Address)
}

// FromHex sets the address from a hex string (e.g., "4CA123")
func (t *TargetAddress) FromHex(s string) error {
	var addr uint32
	_, err := fmt.Sscanf(s, "%x", &addr)
	if err != nil {
		return fmt.Errorf("parsing target address: %w", err)
	}
	t.Address = addr
	return t.Validate()
}

// FromString sets the address from a hex string.
// This is an alias for FromHex for backwards compatibility.
func (t *TargetAddress) FromString(s string) error {
	return t.FromHex(s)
}
