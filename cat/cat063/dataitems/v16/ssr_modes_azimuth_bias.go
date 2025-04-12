// cat/cat063/dataitems/v16/ssr_modes_azimuth_bias.go
package v16

import (
	"bytes"
	"fmt"
	"math"
)

// SSRModeSAzimuthBias implements I063/081
// SSR / Mode S Azimuth Bias, in two's complement form
type SSRModeSAzimuthBias struct {
	Bias float64 // Azimuth bias in degrees (LSB = 360°/2^16 = 0.0055°)
}

const azimuthLSB = 360.0 / 65536.0 // 360°/2^16 = 0.0055°

func (s *SSRModeSAzimuthBias) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading SSR/Mode S azimuth bias: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("insufficient data: got %d bytes, want 2", n)
	}

	// Decode as 16-bit two's complement
	biasRaw := int16(uint16(data[0])<<8 | uint16(data[1]))
	s.Bias = float64(biasRaw) * azimuthLSB // LSB = 360°/2^16 = 0.0055°

	return n, s.Validate()
}

func (s *SSRModeSAzimuthBias) Encode(buf *bytes.Buffer) (int, error) {
	if err := s.Validate(); err != nil {
		return 0, err
	}

	// Encode the bias
	biasRaw := int16(math.Round(s.Bias / azimuthLSB))

	// Output as 2-byte value
	b := make([]byte, 2)
	b[0] = byte(uint16(biasRaw) >> 8)
	b[1] = byte(biasRaw)

	n, err := buf.Write(b)
	if err != nil {
		return n, fmt.Errorf("writing SSR/Mode S azimuth bias: %w", err)
	}
	return n, nil
}

func (s *SSRModeSAzimuthBias) Validate() error {
	// Check for values that would overflow int16 when converted to raw representation
	maxBias := float64(math.MaxInt16) * azimuthLSB
	minBias := float64(math.MinInt16) * azimuthLSB

	if s.Bias < minBias || s.Bias > maxBias {
		return fmt.Errorf("SSR/Mode S azimuth bias out of valid range [%f,%f]: %f",
			minBias, maxBias, s.Bias)
	}

	return nil
}

func (s *SSRModeSAzimuthBias) String() string {
	return fmt.Sprintf("%.4f°", s.Bias)
}
