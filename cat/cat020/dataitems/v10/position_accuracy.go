// cat/cat020/dataitems/v10/position_accuracy.go
package v10

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/davidkohl/gobelix/asterix"
)

// PositionAccuracy represents I020/500 - Position Accuracy
// Compound data item: 1+n octets
// Standard Deviation of Position
type PositionAccuracy struct {
	// Subfield #1: DOP of Position
	DOPPresent bool
	DOPx       float64 // DOP along x axis
	DOPy       float64 // DOP along y axis
	DOPxy      float64 // Correlation DOP

	// Subfield #2: Standard Deviation of Position
	SDPPresent bool
	SDPx       float64 // Standard deviation of X component (m)
	SDPy       float64 // Standard deviation of Y component (m)
	SDPxy      float64 // Correlation coefficient

	// Subfield #3: Standard Deviation of Geometric Altitude
	SDAPresent bool
	SDGA       float64 // Standard deviation of geometric altitude (m)
}

// NewPositionAccuracy creates a new Position Accuracy data item
func NewPositionAccuracy() *PositionAccuracy {
	return &PositionAccuracy{}
}

// Decode decodes the Position Accuracy from bytes
func (p *PositionAccuracy) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 1 {
		return 0, fmt.Errorf("%w: need at least 1 byte, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	bytesRead := 0

	// Read primary subfield
	primary := buf.Next(1)
	bytesRead++

	p.DOPPresent = (primary[0] & 0x80) != 0
	p.SDPPresent = (primary[0] & 0x40) != 0
	p.SDAPresent = (primary[0] & 0x20) != 0
	// Bits 5-1 are spare

	// Subfield #1: DOP of Position (6 bytes)
	if p.DOPPresent {
		if buf.Len() < 6 {
			return bytesRead, fmt.Errorf("%w: need 6 bytes for DOP subfield", asterix.ErrBufferTooShort)
		}
		data := buf.Next(6)
		bytesRead += 6

		// DOPx: 2 bytes, LSB = 0.25
		dopxRaw := binary.BigEndian.Uint16(data[0:2])
		p.DOPx = float64(dopxRaw) * 0.25

		// DOPy: 2 bytes, LSB = 0.25
		dopyRaw := binary.BigEndian.Uint16(data[2:4])
		p.DOPy = float64(dopyRaw) * 0.25

		// DOPxy: 2 bytes, LSB = 0.25
		dopxyRaw := binary.BigEndian.Uint16(data[4:6])
		p.DOPxy = float64(dopxyRaw) * 0.25
	}

	// Subfield #2: Standard Deviation of Position (6 bytes)
	if p.SDPPresent {
		if buf.Len() < 6 {
			return bytesRead, fmt.Errorf("%w: need 6 bytes for SDP subfield", asterix.ErrBufferTooShort)
		}
		data := buf.Next(6)
		bytesRead += 6

		// SDPx: 2 bytes, LSB = 0.25 m
		sdpxRaw := binary.BigEndian.Uint16(data[0:2])
		p.SDPx = float64(sdpxRaw) * 0.25

		// SDPy: 2 bytes, LSB = 0.25 m
		sdpyRaw := binary.BigEndian.Uint16(data[2:4])
		p.SDPy = float64(sdpyRaw) * 0.25

		// SDPxy: 2 bytes, two's complement correlation coefficient, LSB = 0.25
		sdpxyRaw := int16(binary.BigEndian.Uint16(data[4:6]))
		p.SDPxy = float64(sdpxyRaw) * 0.25
	}

	// Subfield #3: Standard Deviation of Geometric Altitude (2 bytes)
	if p.SDAPresent {
		if buf.Len() < 2 {
			return bytesRead, fmt.Errorf("%w: need 2 bytes for SDA subfield", asterix.ErrBufferTooShort)
		}
		data := buf.Next(2)
		bytesRead += 2

		// SDGA: 2 bytes, LSB = 0.5 m
		sdgaRaw := binary.BigEndian.Uint16(data)
		p.SDGA = float64(sdgaRaw) * 0.5
	}

	return bytesRead, nil
}

// Encode encodes the Position Accuracy to bytes
func (p *PositionAccuracy) Encode(buf *bytes.Buffer) (int, error) {
	if err := p.Validate(); err != nil {
		return 0, err
	}

	bytesWritten := 0

	// Primary subfield
	var primary byte
	if p.DOPPresent {
		primary |= 0x80
	}
	if p.SDPPresent {
		primary |= 0x40
	}
	if p.SDAPresent {
		primary |= 0x20
	}
	// Bits 5-1 are spare (0)

	if err := buf.WriteByte(primary); err != nil {
		return bytesWritten, fmt.Errorf("writing primary subfield: %w", err)
	}
	bytesWritten++

	// Subfield #1: DOP of Position
	if p.DOPPresent {
		dopxRaw := uint16(p.DOPx / 0.25)
		dopyRaw := uint16(p.DOPy / 0.25)
		dopxyRaw := uint16(p.DOPxy / 0.25)

		if err := binary.Write(buf, binary.BigEndian, dopxRaw); err != nil {
			return bytesWritten, fmt.Errorf("writing DOPx: %w", err)
		}
		bytesWritten += 2

		if err := binary.Write(buf, binary.BigEndian, dopyRaw); err != nil {
			return bytesWritten, fmt.Errorf("writing DOPy: %w", err)
		}
		bytesWritten += 2

		if err := binary.Write(buf, binary.BigEndian, dopxyRaw); err != nil {
			return bytesWritten, fmt.Errorf("writing DOPxy: %w", err)
		}
		bytesWritten += 2
	}

	// Subfield #2: Standard Deviation of Position
	if p.SDPPresent {
		sdpxRaw := uint16(p.SDPx / 0.25)
		sdpyRaw := uint16(p.SDPy / 0.25)
		sdpxyRaw := int16(p.SDPxy / 0.25)

		if err := binary.Write(buf, binary.BigEndian, sdpxRaw); err != nil {
			return bytesWritten, fmt.Errorf("writing SDPx: %w", err)
		}
		bytesWritten += 2

		if err := binary.Write(buf, binary.BigEndian, sdpyRaw); err != nil {
			return bytesWritten, fmt.Errorf("writing SDPy: %w", err)
		}
		bytesWritten += 2

		if err := binary.Write(buf, binary.BigEndian, sdpxyRaw); err != nil {
			return bytesWritten, fmt.Errorf("writing SDPxy: %w", err)
		}
		bytesWritten += 2
	}

	// Subfield #3: Standard Deviation of Geometric Altitude
	if p.SDAPresent {
		sdgaRaw := uint16(p.SDGA / 0.5)

		if err := binary.Write(buf, binary.BigEndian, sdgaRaw); err != nil {
			return bytesWritten, fmt.Errorf("writing SDGA: %w", err)
		}
		bytesWritten += 2
	}

	return bytesWritten, nil
}

// Validate validates the Position Accuracy
func (p *PositionAccuracy) Validate() error {
	// Check correlation coefficient range
	if p.SDPPresent {
		if math.Abs(p.SDPxy) > 1.0 {
			return fmt.Errorf("%w: SDPxy correlation coefficient must be in range [-1, 1], got %.2f", asterix.ErrInvalidMessage, p.SDPxy)
		}
	}
	return nil
}

// String returns a string representation
func (p *PositionAccuracy) String() string {
	result := ""

	if p.DOPPresent {
		result += fmt.Sprintf("DOP(x=%.2f, y=%.2f, xy=%.2f)", p.DOPx, p.DOPy, p.DOPxy)
	}

	if p.SDPPresent {
		if result != "" {
			result += ", "
		}
		result += fmt.Sprintf("SDP(σx=%.2fm, σy=%.2fm, ρ=%.2f)", p.SDPx, p.SDPy, p.SDPxy)
	}

	if p.SDAPresent {
		if result != "" {
			result += ", "
		}
		result += fmt.Sprintf("SDA(σGA=%.2fm)", p.SDGA)
	}

	if result == "" {
		result = "No accuracy data"
	}

	return result
}
