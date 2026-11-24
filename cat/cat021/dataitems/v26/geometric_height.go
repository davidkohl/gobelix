// dataitems/cat021/geometric_height.go
package v26

import (
	"bytes"
	"fmt"
	"math"

	"github.com/davidkohl/gobelix/asterix"
)

// GeometricHeight implements I021/140
// Geometric Height above WGS-84 ellipsoid (2 octets, signed raw, LSB = 6.25 ft)
type GeometricHeight struct {
	Height float64 // Height in feet (spec range ±(2^15-1)*6.25 ≈ ±204793.75 ft)
}

func (g *GeometricHeight) Encode(buf *bytes.Buffer) (int, error) {
	// Convert from feet to raw value. The RAW value is int16; the feet value
	// itself exceeds int16 above 32767 ft (real geometric heights do), which
	// is why the field is float64.
	r := math.Round(g.Height / 6.25)
	if r < math.MinInt16 || r > math.MaxInt16 {
		return 0, fmt.Errorf("geometric height %f ft out of range", g.Height)
	}
	raw := int16(r)

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
	g.Height = float64(raw) * 6.25

	return n, nil
}

func (g *GeometricHeight) Validate() error {
	return nil
}

func (g *GeometricHeight) String() string {
	return fmt.Sprintf("%.2fft", g.Height)
}
