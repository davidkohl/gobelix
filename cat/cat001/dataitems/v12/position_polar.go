// cat/cat001/dataitems/v12/position_polar.go
package v12

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// PositionPolar represents I001/042 - Calculated Position in Polar Coordinates
// Fixed length: 4 bytes
type PositionPolar struct {
	RHO   float64 // Range in NM (LSB = 1/128 NM)
	THETA float64 // Azimuth in degrees (LSB = 360/2^16 degrees)
}

// Decode decodes Position in Polar Coordinates from bytes
func (p *PositionPolar) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 4 {
		return 0, fmt.Errorf("%w: need 4 bytes for position polar, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	data := buf.Next(4)

	// RHO: 2 bytes, LSB = 1/128 NM
	rho := uint16(data[0])<<8 | uint16(data[1])
	p.RHO = float64(rho) / 128.0

	// THETA: 2 bytes, LSB = 360/2^16 degrees
	theta := uint16(data[2])<<8 | uint16(data[3])
	p.THETA = float64(theta) * 360.0 / 65536.0

	return 4, nil
}

// Encode encodes Position in Polar Coordinates to bytes
func (p *PositionPolar) Encode(buf *bytes.Buffer) (int, error) {
	if err := p.Validate(); err != nil {
		return 0, err
	}

	// RHO: convert NM to 1/128 NM units
	rho := uint16(p.RHO * 128.0)

	// THETA: convert degrees to (360/2^16) units
	theta := uint16(p.THETA * 65536.0 / 360.0)

	data := []byte{
		byte(rho >> 8),
		byte(rho),
		byte(theta >> 8),
		byte(theta),
	}

	n, err := buf.Write(data)
	if err != nil {
		return 0, fmt.Errorf("writing position polar: %w", err)
	}

	return n, nil
}

// Validate validates the Position in Polar Coordinates
func (p *PositionPolar) Validate() error {
	if p.RHO < 0 || p.RHO > 512 {
		return fmt.Errorf("%w: RHO out of range (0-512 NM): %.3f", asterix.ErrInvalidMessage, p.RHO)
	}
	if p.THETA < 0 || p.THETA >= 360 {
		return fmt.Errorf("%w: THETA out of range (0-360°): %.3f", asterix.ErrInvalidMessage, p.THETA)
	}
	return nil
}

// String returns a string representation
func (p *PositionPolar) String() string {
	return fmt.Sprintf("RHO: %.3f NM, THETA: %.3f°", p.RHO, p.THETA)
}
