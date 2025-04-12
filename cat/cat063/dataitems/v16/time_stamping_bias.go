// cat/cat063/dataitems/v16/time_stamping_bias.go
package v16

import (
	"bytes"
	"fmt"
)

// TimeStampingBias implements I063/070
// Plot Time stamping bias, in two's complement form
type TimeStampingBias struct {
	Bias int16 // Time bias in milliseconds, two's complement
}

func (t *TimeStampingBias) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading time stamping bias: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("insufficient data: got %d bytes, want 2", n)
	}

	// Decode value as signed 16-bit integer (two's complement)
	t.Bias = int16(uint16(data[0])<<8 | uint16(data[1]))

	return n, t.Validate()
}

func (t *TimeStampingBias) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	// Encode as two's complement 16-bit value
	b := make([]byte, 2)
	b[0] = byte(uint16(t.Bias) >> 8)
	b[1] = byte(t.Bias)

	n, err := buf.Write(b)
	if err != nil {
		return n, fmt.Errorf("writing time stamping bias: %w", err)
	}
	return n, nil
}

func (t *TimeStampingBias) Validate() error {
	// The specification doesn't indicate a specific range limitation
	// An int16 can hold values from -32768 to 32767 ms
	return nil
}

func (t *TimeStampingBias) String() string {
	return fmt.Sprintf("%d ms", t.Bias)
}
