// dataitems/cat062/mode_of_movement.go
package v117

import (
	"bytes"
	"fmt"
)

// TransversalAcceleration represents the transversal acceleration states
type TransversalAcceleration uint8

const (
	TransConstantCourse TransversalAcceleration = iota
	TransRightTurn
	TransLeftTurn
	TransUndetermined
)

// LongitudinalAcceleration represents the longitudinal acceleration states
type LongitudinalAcceleration uint8

const (
	LongConstantGroundspeed LongitudinalAcceleration = iota
	LongIncreasingGroundspeed
	LongDecreasingGroundspeed
	LongUndetermined
)

// VerticalRate represents the vertical rate states
type VerticalRate uint8

const (
	VertLevel VerticalRate = iota
	VertClimb
	VertDescent
	VertUndetermined
)

// ModeOfMovement implements I062/200
// Calculated Mode of Movement of a target
type ModeOfMovement struct {
	Trans          TransversalAcceleration
	Long           LongitudinalAcceleration
	Vert           VerticalRate
	AltDiscrepancy bool // Altitude Discrepancy Flag
}

func (m *ModeOfMovement) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 1)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading mode of movement: %w", err)
	}
	if n != 1 {
		return n, fmt.Errorf("insufficient data for mode of movement: got %d bytes, want 1", n)
	}

	m.Trans = TransversalAcceleration((data[0] >> 6) & 0x03)
	m.Long = LongitudinalAcceleration((data[0] >> 4) & 0x03)
	m.Vert = VerticalRate((data[0] >> 2) & 0x03)
	m.AltDiscrepancy = (data[0] & 0x02) != 0

	return n, nil
}

func (m *ModeOfMovement) Encode(buf *bytes.Buffer) (int, error) {
	if err := m.Validate(); err != nil {
		return 0, err
	}

	val := byte(m.Trans)<<6 | byte(m.Long)<<4 | byte(m.Vert)<<2
	if m.AltDiscrepancy {
		val |= 0x02
	}
	// Bit 1 (LSB) is spare and set to 0

	err := buf.WriteByte(val)
	if err != nil {
		return 0, fmt.Errorf("writing mode of movement: %w", err)
	}
	return 1, nil
}

func (m *ModeOfMovement) Validate() error {
	if m.Trans > TransUndetermined {
		return fmt.Errorf("invalid transversal acceleration value: %d", m.Trans)
	}
	if m.Long > LongUndetermined {
		return fmt.Errorf("invalid longitudinal acceleration value: %d", m.Long)
	}
	if m.Vert > VertUndetermined {
		return fmt.Errorf("invalid vertical rate value: %d", m.Vert)
	}
	return nil
}

func (m *ModeOfMovement) String() string {
	trans := "Undetermined"
	switch m.Trans {
	case TransConstantCourse:
		trans = "Constant Course"
	case TransRightTurn:
		trans = "Right Turn"
	case TransLeftTurn:
		trans = "Left Turn"
	}

	long := "Undetermined"
	switch m.Long {
	case LongConstantGroundspeed:
		long = "Constant Groundspeed"
	case LongIncreasingGroundspeed:
		long = "Increasing Groundspeed"
	case LongDecreasingGroundspeed:
		long = "Decreasing Groundspeed"
	}

	vert := "Undetermined"
	switch m.Vert {
	case VertLevel:
		vert = "Level"
	case VertClimb:
		vert = "Climb"
	case VertDescent:
		vert = "Descent"
	}

	alt := ""
	if m.AltDiscrepancy {
		alt = ", Altitude Discrepancy"
	}

	return fmt.Sprintf("%s, %s, %s%s", trans, long, vert, alt)
}
