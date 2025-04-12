// dataitems/cat048/communications_capability.go
package v132

import (
	"bytes"
	"fmt"
)

// CommunicationsCapability implements I048/230
// Communications capability of the transponder, capability of the on-board
// ACAS equipment and flight status.
type CommunicationsCapability struct {
	COM  uint8 // Communications capability (3 bits)
	STAT uint8 // Flight Status (3 bits)
	SI   bool  // SI/II Transponder Capability
	MSSC bool  // Mode-S Specific Service Capability
	ARC  bool  // Altitude reporting capability
	AIC  bool  // Aircraft identification capability
	B1A  bool  // BDS 1,0 bit 16
	B1B  uint8 // BDS 1,0 bits 37/40 (4 bits)
}

// COM (Communications Capability) values
const (
	COMNoCapability      uint8 = 0 // No communications capability (surveillance only)
	COMCommAB            uint8 = 1 // Comm. A and Comm. B capability
	COMCommAB_UplinkELM  uint8 = 2 // Comm. A, Comm. B and Uplink ELM
	COMCommAB_ELM        uint8 = 3 // Comm. A, Comm. B, Uplink ELM and Downlink ELM
	COMLevel5Transponder uint8 = 4 // Level 5 Transponder capability
)

// STAT (Flight Status) values
const (
	STATNoAlertAirborne   uint8 = 0 // No alert, no SPI, aircraft airborne
	STATNoAlertGround     uint8 = 1 // No alert, no SPI, aircraft on ground
	STATAlertAirborne     uint8 = 2 // Alert, no SPI, aircraft airborne
	STATAlertGround       uint8 = 3 // Alert, no SPI, aircraft on ground
	STATAlertSPI          uint8 = 4 // Alert, SPI, aircraft airborne or on ground
	STATNoAlertSPI        uint8 = 5 // No alert, SPI, aircraft airborne or on ground
	STATNotAssigned       uint8 = 6 // Not assigned
	STATUnknownFlightInfo uint8 = 7 // Unknown
)

// Decode implements the DataItem interface
func (c *CommunicationsCapability) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading communications capability: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("insufficient data for communications capability: got %d bytes, want 2", n)
	}

	c.COM = (data[0] >> 5) & 0x07  // bits 16-14
	c.STAT = (data[0] >> 2) & 0x07 // bits 13-11
	c.SI = (data[0] & 0x02) != 0   // bit 10
	// bit 9 is spare
	c.MSSC = (data[0] & 0x01) != 0 // bit 8
	c.ARC = (data[1] & 0x80) != 0  // bit 7
	c.AIC = (data[1] & 0x40) != 0  // bit 6
	c.B1A = (data[1] & 0x20) != 0  // bit 5
	c.B1B = data[1] & 0x0F         // bits 4-1

	return n, c.Validate()
}

// Encode implements the DataItem interface
func (c *CommunicationsCapability) Encode(buf *bytes.Buffer) (int, error) {
	if err := c.Validate(); err != nil {
		return 0, err
	}

	data := make([]byte, 2)

	// First byte
	data[0] |= (c.COM & 0x07) << 5  // bits 16-14
	data[0] |= (c.STAT & 0x07) << 2 // bits 13-11
	if c.SI {
		data[0] |= 0x02 // bit 10
	}
	// bit 9 is spare
	if c.MSSC {
		data[0] |= 0x01 // bit 8
	}

	// Second byte
	if c.ARC {
		data[1] |= 0x80 // bit 7
	}
	if c.AIC {
		data[1] |= 0x40 // bit 6
	}
	if c.B1A {
		data[1] |= 0x20 // bit 5
	}
	data[1] |= c.B1B & 0x0F // bits 4-1

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing communications capability: %w", err)
	}
	return n, nil
}

// Validate implements the DataItem interface
func (c *CommunicationsCapability) Validate() error {
	if c.COM > 7 {
		return fmt.Errorf("invalid COM value: %d", c.COM)
	}
	if c.STAT > 7 {
		return fmt.Errorf("invalid STAT value: %d", c.STAT)
	}
	if c.B1B > 15 {
		return fmt.Errorf("invalid B1B value: %d", c.B1B)
	}
	return nil
}

// String returns a human-readable representation
func (c *CommunicationsCapability) String() string {
	result := fmt.Sprintf("COM: %d", c.COM)

	// Add COM description
	switch c.COM {
	case COMNoCapability:
		result += " (Surveillance Only)"
	case COMCommAB:
		result += " (Comm-A/B)"
	case COMCommAB_UplinkELM:
		result += " (Comm-A/B + Uplink ELM)"
	case COMCommAB_ELM:
		result += " (Comm-A/B + ELM)"
	case COMLevel5Transponder:
		result += " (Level 5 Transponder)"
	}

	// Add flight status
	result += fmt.Sprintf(", Status: %d", c.STAT)

	// Add SI/II capability
	if c.SI {
		result += ", II-capable"
	} else {
		result += ", SI-capable"
	}

	// Add other capability bits
	if c.MSSC {
		result += ", MSSC"
	}
	if c.ARC {
		result += ", 25ft Resolution"
	} else {
		result += ", 100ft Resolution"
	}
	if c.AIC {
		result += ", AIC"
	}

	return result
}
