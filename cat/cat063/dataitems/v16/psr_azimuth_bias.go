// cat/cat063/dataitems/v16/psr_azimuth_bias.go
package v16

import (
	"bytes"
	"fmt"
	"math"
)

// PSRAzimuthBias implements I063/091
// PSR Azimuth Bias, in two's complement form
type PSRAzimuthBias struct {
	Bias float64 // Azimuth bias in degrees (LSB = 360°/2^16 = 0.0055°)
}

// Use same constant as in SSRModeSAzimuthBias
const psrAzimuthLSB = 360.0 / 65536.0 // 360°/2^16 = 0.0055°

func (p *PSRAzimuthBias) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading PSR azimuth bias: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("insufficient data: got %d bytes, want 2", n)
	}

	// Decode as 16-bit two's complement
	biasRaw := int16(uint16(data[0])<<8 | uint16(data[1]))
	p.Bias = float64(biasRaw) * psrAzimuthLSB // LSB = 360°/2^16 = 0.0055°

	return n, p.Validate()
}

func (p *PSRAzimuthBias) Encode(buf *bytes.Buffer) (int, error) {
	if err := p.Validate(); err != nil {
		return 0, err
	}

	// Encode the bias
	biasRaw := int16(math.Round(p.Bias / psrAzimuthLSB))

	// Output as 2-byte value
	b := make([]byte, 2)
	b[0] = byte(uint16(biasRaw) >> 8)
	b[1] = byte(biasRaw)

	n, err := buf.Write(b)
	if err != nil {
		return n, fmt.Errorf("writing PSR azimuth bias: %w", err)
	}
	return n, nil
}

func (p *PSRAzimuthBias) Validate() error {
	// Check for values that would overflow int16 when converted to raw representation
	maxBias := float64(math.MaxInt16) * psrAzimuthLSB
	minBias := float64(math.MinInt16) * psrAzimuthLSB

	if p.Bias < minBias || p.Bias > maxBias {
		return fmt.Errorf("PSR azimuth bias out of valid range [%f,%f]: %f",
			minBias, maxBias, p.Bias)
	}

	return nil
}

func (p *PSRAzimuthBias) String() string {
	return fmt.Sprintf("%.4f°", p.Bias)
}
