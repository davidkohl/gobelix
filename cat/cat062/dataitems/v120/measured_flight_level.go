// dataitems/cat062/measured_flight_level.go
package v120

import (
	"bytes"
	"fmt"
)

// MeasuredFlightLevel implements I062/136
// Last valid and credible flight level used to update the track
type MeasuredFlightLevel struct {
	FlightLevel float64 // Flight level (1 FL = 100 ft)
}

func (m *MeasuredFlightLevel) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading measured flight level: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("insufficient data for measured flight level: got %d bytes, want 2", n)
	}

	// Flight level in two's complement form, LSB = 1/4 FL = 25 ft
	raw := int16(data[0])<<8 | int16(data[1])
	m.FlightLevel = float64(raw) * 0.25

	return n, nil
}

func (m *MeasuredFlightLevel) Encode(buf *bytes.Buffer) (int, error) {
	if err := m.Validate(); err != nil {
		return 0, err
	}

	// Convert to raw value
	raw := int16(m.FlightLevel / 0.25)

	data := []byte{
		byte(raw >> 8),
		byte(raw),
	}

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing measured flight level: %w", err)
	}
	return n, nil
}

func (m *MeasuredFlightLevel) Validate() error {
	// According to the spec, valid range is -15 FL to 1500 FL
	if m.FlightLevel < -15 || m.FlightLevel > 1500 {
		return fmt.Errorf("flight level out of range [-15,1500]: %f", m.FlightLevel)
	}
	return nil
}

func (m *MeasuredFlightLevel) String() string {
	return fmt.Sprintf("Measured Flight Level: FL %.2f", m.FlightLevel)
}
