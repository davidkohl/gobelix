// cat/cat063/dataitems/v16/psr_range_gain_and_bias.go
package v16

import (
	"bytes"
	"fmt"
	"math"
)

// PSRRangeGainAndBias implements I063/090
// PSR Range Gain and PSR Range Bias, in two's complement form
type PSRRangeGainAndBias struct {
	Gain float64 // Range gain (LSB = 10^-5)
	Bias float64 // Range bias in NM (LSB = 1/128 NM)
}

func (p *PSRRangeGainAndBias) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 4)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading PSR range gain and bias: %w", err)
	}
	if n != 4 {
		return n, fmt.Errorf("insufficient data: got %d bytes, want 4", n)
	}

	// Decode gain (first 2 bytes) as 16-bit two's complement
	gainRaw := int16(uint16(data[0])<<8 | uint16(data[1]))
	p.Gain = float64(gainRaw) * 1e-5 // LSB = 10^-5

	// Decode bias (last 2 bytes) as 16-bit two's complement
	biasRaw := int16(uint16(data[2])<<8 | uint16(data[3]))
	p.Bias = float64(biasRaw) / 128.0 // LSB = 1/128 NM

	return n, p.Validate()
}

func (p *PSRRangeGainAndBias) Encode(buf *bytes.Buffer) (int, error) {
	if err := p.Validate(); err != nil {
		return 0, err
	}

	// Encode the gain
	gainRaw := int16(math.Round(p.Gain / 1e-5))

	// Encode the bias
	biasRaw := int16(math.Round(p.Bias * 128.0))

	// Combine into 4-byte output
	b := make([]byte, 4)
	b[0] = byte(uint16(gainRaw) >> 8)
	b[1] = byte(gainRaw)
	b[2] = byte(uint16(biasRaw) >> 8)
	b[3] = byte(biasRaw)

	n, err := buf.Write(b)
	if err != nil {
		return n, fmt.Errorf("writing PSR range gain and bias: %w", err)
	}
	return n, nil
}

func (p *PSRRangeGainAndBias) Validate() error {
	// Check for values that would overflow int16 when converted to raw representation
	if p.Gain < -0.32768 || p.Gain > 0.32767 {
		return fmt.Errorf("PSR range gain out of valid range: %f", p.Gain)
	}

	if p.Bias < -256.0 || p.Bias > 255.99 {
		return fmt.Errorf("PSR range bias out of valid range: %f", p.Bias)
	}

	return nil
}

func (p *PSRRangeGainAndBias) String() string {
	return fmt.Sprintf("Gain: %.5f, Bias: %.3f NM", p.Gain, p.Bias)
}
