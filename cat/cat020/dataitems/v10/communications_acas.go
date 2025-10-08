// cat/cat020/dataitems/v10/communications_acas.go
package v10

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/davidkohl/gobelix/asterix"
)

// CommunicationsACAS represents I020/230 - Communications/ACAS Capability and Flight Status
// Fixed length: 2 bytes
// Communications capability of the transponder, capability of the on-board ACAS equipment
// and flight status
type CommunicationsACAS struct {
	COM  uint8 // Communications capability (3 bits)
	STAT uint8 // Flight Status (3 bits)
	MSSC bool  // Mode-S Specific Service Capability
	ARC  bool  // Altitude reporting capability (false=100ft, true=25ft)
	AIC  bool  // Aircraft identification capability
	B1A  bool  // BDS 1,0 bit 16
	B1B  uint8 // BDS 1,0 bits 37-40 (4 bits)
}

// NewCommunicationsACAS creates a new Communications/ACAS data item
func NewCommunicationsACAS() *CommunicationsACAS {
	return &CommunicationsACAS{}
}

// Decode decodes the Communications/ACAS from bytes
func (c *CommunicationsACAS) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 2 {
		return 0, fmt.Errorf("%w: need 2 bytes, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	data := buf.Next(2)
	value := binary.BigEndian.Uint16(data)

	c.COM = uint8((value >> 13) & 0x07)
	c.STAT = uint8((value >> 10) & 0x07)
	// Bits 10-9 are spare
	c.MSSC = (value & 0x0080) != 0
	c.ARC = (value & 0x0040) != 0
	c.AIC = (value & 0x0020) != 0
	c.B1A = (value & 0x0010) != 0
	c.B1B = uint8(value & 0x000F)

	return 2, nil
}

// Encode encodes the Communications/ACAS to bytes
func (c *CommunicationsACAS) Encode(buf *bytes.Buffer) (int, error) {
	if err := c.Validate(); err != nil {
		return 0, err
	}

	var value uint16
	value |= uint16(c.COM&0x07) << 13
	value |= uint16(c.STAT&0x07) << 10
	// Bits 10-9 are spare (0)
	if c.MSSC {
		value |= 0x0080
	}
	if c.ARC {
		value |= 0x0040
	}
	if c.AIC {
		value |= 0x0020
	}
	if c.B1A {
		value |= 0x0010
	}
	value |= uint16(c.B1B & 0x0F)

	if err := binary.Write(buf, binary.BigEndian, value); err != nil {
		return 0, fmt.Errorf("writing communications/ACAS: %w", err)
	}

	return 2, nil
}

// Validate validates the Communications/ACAS
func (c *CommunicationsACAS) Validate() error {
	if c.COM > 7 {
		return fmt.Errorf("%w: COM must be 0-7, got %d", asterix.ErrInvalidMessage, c.COM)
	}
	if c.STAT > 7 {
		return fmt.Errorf("%w: STAT must be 0-7, got %d", asterix.ErrInvalidMessage, c.STAT)
	}
	if c.B1B > 15 {
		return fmt.Errorf("%w: B1B must be 0-15, got %d", asterix.ErrInvalidMessage, c.B1B)
	}
	return nil
}

// String returns a string representation
func (c *CommunicationsACAS) String() string {
	var parts []string

	// COM capability
	comStr := ""
	switch c.COM {
	case 0:
		comStr = "No comm"
	case 1:
		comStr = "Comm A/B"
	case 2:
		comStr = "Comm A/B/Uplink ELM"
	case 3:
		comStr = "Comm A/B/Uplink+Downlink ELM"
	case 4:
		comStr = "Level 5"
	default:
		comStr = fmt.Sprintf("COM=%d", c.COM)
	}
	parts = append(parts, comStr)

	// STAT flight status
	statStr := ""
	switch c.STAT {
	case 0:
		statStr = "Airborne"
	case 1:
		statStr = "OnGround"
	case 2:
		statStr = "Alert,Airborne"
	case 3:
		statStr = "Alert,OnGround"
	case 4:
		statStr = "Alert,SPI"
	case 5:
		statStr = "SPI"
	default:
		statStr = fmt.Sprintf("STAT=%d", c.STAT)
	}
	parts = append(parts, statStr)

	// Additional capabilities
	if c.MSSC {
		parts = append(parts, "MSSC")
	}
	if c.ARC {
		parts = append(parts, "ARC=25ft")
	} else {
		parts = append(parts, "ARC=100ft")
	}
	if c.AIC {
		parts = append(parts, "AIC")
	}

	return strings.Join(parts, " ")
}
