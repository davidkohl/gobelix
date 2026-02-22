// dataitems/cat021/geometric_height.go
package v26

import (
	"bytes"
	"fmt"
	"math"

	"github.com/davidkohl/gobelix/asterix"
)

// GeometricHeight implements I021/140
// Geometric Height above WGS-84 ellipsoid (2 octets)
type GeometricHeight struct {
	Height int16 // Height in feet, LSB = 6.25 ft
}

func (g *GeometricHeight) Encode(buf *bytes.Buffer) (int, error) {
	// Convert from feet to raw value
	raw := int16(math.Round(float64(g.Height) / 6.25))

	var data [2]byte
	data[0] = byte(raw >> 8)
	data[1] = byte(raw)

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing geometric height: %w", err)
	}
	return n, nil
}

func (g *GeometricHeight) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading geometric height: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("%w: need 2 bytes for geometric height, have %d", asterix.ErrBufferTooShort, n)
	}

	// Combine into signed int16
	raw := int16(uint16(data[0])<<8 | uint16(data[1]))

	// Convert to feet
	g.Height = int16(math.Round(float64(raw) * 6.25))

	return n, nil
}

func (g *GeometricHeight) Validate() error {
	return nil
}

func (g *GeometricHeight) String() string {
	return fmt.Sprintf("%dft", g.Height)
}
