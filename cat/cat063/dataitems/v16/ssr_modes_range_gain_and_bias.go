// cat/cat063/dataitems/v16/ssr_modes_range_gain_and_bias.go
package v16

import (
	"bytes"
	"fmt"
	"math"
)

// SSRModeSRangeGainAndBias implements I063/080
// SSR / Mode S Range Gain and Range Bias, in two's complement form
type SSRModeSRangeGainAndBias struct {
	Gain float64 // Range gain (LSB = 10^-5)
	Bias float64 // Range bias in NM (LSB = 1/128 NM)
}

func (s *SSRModeSRangeGainAndBias) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 4)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading SSR/Mode S range gain and bias: %w", err)
	}
	if n != 4 {
		return n, fmt.Errorf("insufficient data: got %d bytes, want 4", n)
	}

	// Decode gain (first 2 bytes) as 16-bit two's complement
	gainRaw := int16(uint16(data[0])<<8 | uint16(data[1]))
	s.Gain = float64(gainRaw) * 1e-5 // LSB = 10^-5

	// Decode bias (last 2 bytes) as 16-bit two's complement
	biasRaw := int16(uint16(data[2])<<8 | uint16(data[3]))
	s.Bias = float64(biasRaw) / 128.0 // LSB = 1/128 NM

	return n, s.Validate()
}

func (s *SSRModeSRangeGainAndBias) Encode(buf *bytes.Buffer) (int, error) {
	if err := s.Validate(); err != nil {
		return 0, err
	}

	// Encode the gain
	gainRaw := int16(math.Round(s.Gain / 1e-5))

	// Encode the bias
	biasRaw := int16(math.Round(s.Bias * 128.0))

	// Combine into 4-byte output
	b := make([]byte, 4)
	b[0] = byte(uint16(gainRaw) >> 8)
	b[1] = byte(gainRaw)
	b[2] = byte(uint16(biasRaw) >> 8)
	b[3] = byte(biasRaw)

	n, err := buf.Write(b)
	if err != nil {
		return n, fmt.Errorf("writing SSR/Mode S range gain and bias: %w", err)
	}
	return n, nil
}

func (s *SSRModeSRangeGainAndBias) Validate() error {
	// Check for values that would overflow int16 when converted to raw representation
	if s.Gain < -0.32768 || s.Gain > 0.32767 {
		return fmt.Errorf("SSR/Mode S range gain out of valid range: %f", s.Gain)
	}

	if s.Bias < -256.0 || s.Bias > 255.99 {
		return fmt.Errorf("SSR/Mode S range bias out of valid range: %f", s.Bias)
	}

	return nil
}

func (s *SSRModeSRangeGainAndBias) String() string {
	return fmt.Sprintf("Gain: %.5f, Bias: %.3f NM", s.Gain, s.Bias)
}
