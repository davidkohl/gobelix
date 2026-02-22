package v13

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// MeasuredFlightLevel implements I011/090 - Measured Flight Level
// Definition: Last valid and credible flight level used to update the track,
// in two's complement representation.
// Format: Two-octet fixed length Data Item
// LSB = 1/4 FL, range -12 FL to 1500 FL
type MeasuredFlightLevel struct {
	FlightLevel float64 // Flight level in FL units
}

const flightLevelLSB = 0.25 // 1/4 FL

func (f *MeasuredFlightLevel) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading measured flight level: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("measured flight level: expected 2 bytes, got %d", n)
	}

	raw := int16(binary.BigEndian.Uint16(data[:]))
	f.FlightLevel = float64(raw) * flightLevelLSB

	return 2, nil
}

func (f *MeasuredFlightLevel) Encode(buf *bytes.Buffer) (int, error) {
	if err := f.Validate(); err != nil {
		return 0, err
	}

	var data [2]byte
	raw := int16(f.FlightLevel / flightLevelLSB)
	binary.BigEndian.PutUint16(data[:], uint16(raw))

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing measured flight level: %w", err)
	}
	return n, nil
}

func (f *MeasuredFlightLevel) Validate() error {
	if f.FlightLevel < -12 || f.FlightLevel > 1500 {
		return fmt.Errorf("flight level out of range: %f (expected -12 to 1500)", f.FlightLevel)
	}
	return nil
}

func (f *MeasuredFlightLevel) String() string {
	return fmt.Sprintf("Flight Level: FL%.2f", f.FlightLevel)
}

// GeometricAltitude implements I011/092 - Calculated Track Geometric Altitude
// Definition: Calculated geometric vertical distance above mean sea level,
// not related to barometric pressure.
// Format: Two-Octet fixed length data item
// LSB = 6.25 ft, range -1500 ft to 150000 ft
type GeometricAltitude struct {
	Altitude float64 // Altitude in feet
}

const geometricAltitudeLSB = 6.25 // feet

func (g *GeometricAltitude) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading geometric altitude: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("geometric altitude: expected 2 bytes, got %d", n)
	}

	raw := int16(binary.BigEndian.Uint16(data[:]))
	g.Altitude = float64(raw) * geometricAltitudeLSB

	return 2, nil
}

func (g *GeometricAltitude) Encode(buf *bytes.Buffer) (int, error) {
	if err := g.Validate(); err != nil {
		return 0, err
	}

	var data [2]byte
	raw := int16(g.Altitude / geometricAltitudeLSB)
	binary.BigEndian.PutUint16(data[:], uint16(raw))

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing geometric altitude: %w", err)
	}
	return n, nil
}

func (g *GeometricAltitude) Validate() error {
	if g.Altitude < -1500 || g.Altitude > 150000 {
		return fmt.Errorf("geometric altitude out of range: %f (expected -1500 to 150000 ft)", g.Altitude)
	}
	return nil
}

func (g *GeometricAltitude) String() string {
	return fmt.Sprintf("Geometric Altitude: %.1f ft", g.Altitude)
}

// BarometricAltitude implements I011/093 - Calculated Track Barometric Altitude
// Definition: Calculated Barometric Altitude of the track.
// Format: Two-Octet fixed length data item
// Bit 16: QNH correction applied flag
// Bits 15-1: Altitude, LSB = 1/4 FL = 25 ft, range -15 FL to 1500 FL
type BarometricAltitude struct {
	QNH         bool    // true if QNH correction applied
	FlightLevel float64 // Flight level
}

func (b *BarometricAltitude) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading barometric altitude: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("barometric altitude: expected 2 bytes, got %d", n)
	}

	raw := binary.BigEndian.Uint16(data[:])
	b.QNH = (raw & 0x8000) != 0

	// Bits 15-1 contain the altitude (signed 15-bit value)
	altRaw := int16(raw & 0x7FFF)
	if altRaw&0x4000 != 0 {
		// Sign extend from 15 bits
		altRaw |= -0x4000
	}
	b.FlightLevel = float64(altRaw) * flightLevelLSB

	return 2, nil
}

func (b *BarometricAltitude) Encode(buf *bytes.Buffer) (int, error) {
	if err := b.Validate(); err != nil {
		return 0, err
	}

	var data [2]byte
	raw := int16(b.FlightLevel / flightLevelLSB)
	encoded := uint16(raw) & 0x7FFF
	if b.QNH {
		encoded |= 0x8000
	}
	binary.BigEndian.PutUint16(data[:], encoded)

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing barometric altitude: %w", err)
	}
	return n, nil
}

func (b *BarometricAltitude) Validate() error {
	if b.FlightLevel < -15 || b.FlightLevel > 1500 {
		return fmt.Errorf("barometric altitude out of range: %f (expected -15 to 1500 FL)", b.FlightLevel)
	}
	return nil
}

func (b *BarometricAltitude) String() string {
	qnh := ""
	if b.QNH {
		qnh = " (QNH corrected)"
	}
	return fmt.Sprintf("Barometric Altitude: FL%.2f%s", b.FlightLevel, qnh)
}

// RateOfClimbDescent implements I011/215 - Calculated Rate Of Climb/Descent
// Definition: Calculated rate of Climb/Descent of an aircraft, in two's complement form.
// Format: Two-Octet fixed length data item
// LSB = 6.25 feet/minute, max range = ±204800 feet/minute
type RateOfClimbDescent struct {
	Rate float64 // Rate in feet/minute (positive = climbing, negative = descending)
}

const rateOfClimbLSB = 6.25 // feet/minute

func (r *RateOfClimbDescent) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading rate of climb/descent: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("rate of climb/descent: expected 2 bytes, got %d", n)
	}

	raw := int16(binary.BigEndian.Uint16(data[:]))
	r.Rate = float64(raw) * rateOfClimbLSB

	return 2, nil
}

func (r *RateOfClimbDescent) Encode(buf *bytes.Buffer) (int, error) {
	if err := r.Validate(); err != nil {
		return 0, err
	}

	var data [2]byte
	raw := int16(r.Rate / rateOfClimbLSB)
	binary.BigEndian.PutUint16(data[:], uint16(raw))

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing rate of climb/descent: %w", err)
	}
	return n, nil
}

func (r *RateOfClimbDescent) Validate() error {
	maxRate := 204800.0
	if r.Rate < -maxRate || r.Rate > maxRate {
		return fmt.Errorf("rate of climb/descent out of range: %f", r.Rate)
	}
	return nil
}

func (r *RateOfClimbDescent) String() string {
	direction := "Level"
	if r.Rate > 0 {
		direction = "Climbing"
	} else if r.Rate < 0 {
		direction = "Descending"
	}
	return fmt.Sprintf("Rate: %.1f ft/min (%s)", r.Rate, direction)
}
