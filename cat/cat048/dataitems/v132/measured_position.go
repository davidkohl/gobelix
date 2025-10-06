// dataitems/cat048/measured_position.go
package v132

import (
	"bytes"
	"fmt"
	"math"
)

// MeasuredPosition implements I048/040
// Measured position of an aircraft in local polar co-ordinates.
// According to ASTERIX spec: Measured distance (range) expressed in
// slant range and angle (azimuth) in the local polar coordinate system
// of the radar station.
type MeasuredPosition struct {
	RHO   float64 // Measured range (slant range) in NM, LSB = 1/256 NM, max ~256 NM
	THETA float64 // Measured azimuth in degrees, LSB = 360°/2^16 ≈ 0.0055°
}

// Decode implements the DataItem interface
func (m *MeasuredPosition) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 4)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading measured position: %w", err)
	}
	if n != 4 {
		return n, fmt.Errorf("insufficient data for measured position: got %d bytes, want 4", n)
	}

	// RHO (16 bits): Range, LSB = 1/256 NM
	rhoRaw := uint16(data[0])<<8 | uint16(data[1])
	m.RHO = float64(rhoRaw) / 256.0

	// THETA (16 bits): Azimuth, LSB = 360/2^16 degrees ≈ 0.0055 degrees
	thetaRaw := uint16(data[2])<<8 | uint16(data[3])
	m.THETA = float64(thetaRaw) * (360.0 / 65536.0)

	return n, nil
}

// Encode implements the DataItem interface
func (m *MeasuredPosition) Encode(buf *bytes.Buffer) (int, error) {
	if err := m.Validate(); err != nil {
		return 0, err
	}

	// Convert RHO to raw value, ensuring it stays within valid range
	rhoRaw := uint16(math.Round(m.RHO * 256.0))
	if m.RHO >= 256.0 {
		rhoRaw = 0xFFFF // Cap at maximum value
	}

	// Convert THETA to raw value, normalize to [0, 360) degrees
	theta := math.Mod(m.THETA, 360.0)
	if theta < 0 {
		theta += 360.0
	}
	thetaRaw := uint16(math.Round(theta * (65536.0 / 360.0)))
	if theta >= 360.0-1e-10 { // Handle potential rounding for 360 degrees
		thetaRaw = 0
	}

	data := make([]byte, 4)
	data[0] = byte(rhoRaw >> 8)
	data[1] = byte(rhoRaw)
	data[2] = byte(thetaRaw >> 8)
	data[3] = byte(thetaRaw)

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing measured position: %w", err)
	}
	return n, nil
}

// Validate implements the DataItem interface
func (m *MeasuredPosition) Validate() error {
	if m.RHO < 0 {
		return fmt.Errorf("negative range value not allowed: %f", m.RHO)
	}
	if m.RHO >= 256.0 {
		return fmt.Errorf("range value too large (>= 256 NM): %f", m.RHO)
	}
	return nil
}

// String returns a human-readable representation
func (m *MeasuredPosition) String() string {
	return fmt.Sprintf("RHO: %.3f NM, THETA: %.3f°", m.RHO, m.THETA)
}
