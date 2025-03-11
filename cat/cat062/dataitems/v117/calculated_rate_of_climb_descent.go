// dataitems/cat062/calculated_rate_of_climb_descent.go
package v117

import (
	"bytes"
	"fmt"
	"math"
)

// CalculatedRateOfClimbDescent implements I062/220
// Calculated rate of Climb/Descent of an aircraft
type CalculatedRateOfClimbDescent struct {
	Rate float64 // Rate in feet/minute, positive for climb, negative for descent
}

func (c *CalculatedRateOfClimbDescent) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading calculated rate of climb/descent: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("insufficient data for calculated rate of climb/descent: got %d bytes, want 2", n)
	}

	// Rate in two's complement form, LSB = 6.25 feet/minute
	raw := int16(data[0])<<8 | int16(data[1])
	c.Rate = float64(raw) * 6.25

	return n, nil
}

func (c *CalculatedRateOfClimbDescent) Encode(buf *bytes.Buffer) (int, error) {
	if err := c.Validate(); err != nil {
		return 0, err
	}

	// Convert to raw value
	raw := int16(math.Round(c.Rate / 6.25))

	data := []byte{
		byte(raw >> 8),
		byte(raw),
	}

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing calculated rate of climb/descent: %w", err)
	}
	return n, nil
}

func (c *CalculatedRateOfClimbDescent) Validate() error {
	// int16 range with LSB of 6.25 gives range of approximately ±32000 feet/minute
	// but there's no specific range mentioned in the spec, so we use a reasonable limit
	if c.Rate < -32000 || c.Rate > 32000 {
		return fmt.Errorf("rate of climb/descent out of range: %f", c.Rate)
	}
	return nil
}

func (c *CalculatedRateOfClimbDescent) String() string {
	if c.Rate > 0 {
		return fmt.Sprintf("Rate of Climb: %.0f ft/min", c.Rate)
	} else if c.Rate < 0 {
		return fmt.Sprintf("Rate of Descent: %.0f ft/min", -c.Rate)
	}
	return "Level Flight"
}
