// cat/cat020/dataitems/v15/communications_capability.go
package v15

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// CommunicationsCapability implements I020/230 - Communications/ACAS Capability and Flight Status
// Per CAT020 spec v1.11 pages 28-29:
// Octet 1: COM(16-14) STAT(13-11) CASEVN(10-9)
// Octet 2: MSSC(8) ARC(7) AIC(6) B1A(5) B1B(4-1)
type CommunicationsCapability struct {
	COM    uint8 // bits 16-14: Communications capability (0-7)
	STAT   uint8 // bits 13-11: Flight status (0-7)
	CASEVN uint8 // bits 10-9: CAS Extended Version Number (0-3)
	MSSC   bool  // bit 8: Mode-S Specific Service Capability
	ARC    bool  // bit 7: Altitude Reporting Capability
	AIC    bool  // bit 6: Aircraft Identification Capability
	B1A    bool  // bit 5: BDS 1,0 bit 16
	B1B    uint8 // bits 4-1: BDS 1,0 bits 37-40 (4 bits)
}

func (c *CommunicationsCapability) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading communications capability: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("%w: need 2 bytes for communications capability, have %d", asterix.ErrBufferTooShort, n)
	}

	// Octet 1 (data[0]): COM(bits 7-5) STAT(bits 4-2) CASEVN(bits 1-0)
	c.COM = (data[0] >> 5) & 0x07    // bits 16-14
	c.STAT = (data[0] >> 2) & 0x07   // bits 13-11
	c.CASEVN = data[0] & 0x03        // bits 10-9

	// Octet 2 (data[1]): MSSC(bit 7) ARC(bit 6) AIC(bit 5) B1A(bit 4) B1B(bits 3-0)
	c.MSSC = (data[1] & 0x80) != 0   // bit 8
	c.ARC = (data[1] & 0x40) != 0    // bit 7
	c.AIC = (data[1] & 0x20) != 0    // bit 6
	c.B1A = (data[1] & 0x10) != 0    // bit 5
	c.B1B = data[1] & 0x0F           // bits 4-1

	return n, nil
}

func (c *CommunicationsCapability) Encode(buf *bytes.Buffer) (int, error) {
	if err := c.Validate(); err != nil {
		return 0, err
	}

	var data [2]byte

	// Octet 1: COM(bits 7-5) STAT(bits 4-2) CASEVN(bits 1-0)
	data[0] = ((c.COM & 0x07) << 5) | ((c.STAT & 0x07) << 2) | (c.CASEVN & 0x03)

	// Octet 2: MSSC(bit 7) ARC(bit 6) AIC(bit 5) B1A(bit 4) B1B(bits 3-0)
	if c.MSSC {
		data[1] |= 0x80 // bit 8
	}
	if c.ARC {
		data[1] |= 0x40 // bit 7
	}
	if c.AIC {
		data[1] |= 0x20 // bit 6
	}
	if c.B1A {
		data[1] |= 0x10 // bit 5
	}
	data[1] |= c.B1B & 0x0F // bits 4-1

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing communications capability: %w", err)
	}
	return n, nil
}

func (c *CommunicationsCapability) Validate() error {
	if c.COM > 7 {
		return fmt.Errorf("COM out of range: %d", c.COM)
	}
	if c.STAT > 7 {
		return fmt.Errorf("STAT out of range: %d", c.STAT)
	}
	if c.CASEVN > 3 {
		return fmt.Errorf("CASEVN out of range: %d", c.CASEVN)
	}
	if c.B1B > 15 {
		return fmt.Errorf("B1B out of range: %d", c.B1B)
	}
	return nil
}

func (c *CommunicationsCapability) String() string {
	comStr := ""
	switch c.COM {
	case 0:
		comStr = "No comms"
	case 1:
		comStr = "Comm A/B"
	case 2:
		comStr = "Comm A/B + Uplink ELM"
	case 3:
		comStr = "Comm A/B + Uplink/Downlink ELM"
	case 4:
		comStr = "Level 5 Transponder"
	default:
		comStr = fmt.Sprintf("Reserved (%d)", c.COM)
	}

	statStr := ""
	switch c.STAT {
	case 0:
		statStr = "No alert, no SPI, airborne"
	case 1:
		statStr = "No alert, no SPI, on ground"
	case 2:
		statStr = "Alert, no SPI, airborne"
	case 3:
		statStr = "Alert, no SPI, on ground"
	case 4:
		statStr = "Alert, SPI"
	case 5:
		statStr = "No alert, SPI"
	case 6:
		statStr = "Not assigned"
	case 7:
		statStr = "Not yet extracted"
	}

	return fmt.Sprintf("COM: %s, STAT: %s", comStr, statStr)
}
