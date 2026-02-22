package v13

import (
	"bytes"
	"fmt"
)

// CalculatedAcceleration implements I011/210 - Calculated Acceleration
// Definition: Calculated Acceleration of the target, in two's complement form.
// Format: Two-Octet fixed length data item
// LSB = 0.25 m/s², max range = ±31 m/s²
type CalculatedAcceleration struct {
	Ax float64 // X-component acceleration in m/s²
	Ay float64 // Y-component acceleration in m/s²
}

const accelerationLSB = 0.25 // m/s²

func (a *CalculatedAcceleration) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading acceleration: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("acceleration: expected 2 bytes, got %d", n)
	}

	// Ax: bits 16-9 (first byte), two's complement
	axRaw := int8(data[0])
	a.Ax = float64(axRaw) * accelerationLSB

	// Ay: bits 8-1 (second byte), two's complement
	ayRaw := int8(data[1])
	a.Ay = float64(ayRaw) * accelerationLSB

	return 2, nil
}

func (a *CalculatedAcceleration) Encode(buf *bytes.Buffer) (int, error) {
	if err := a.Validate(); err != nil {
		return 0, err
	}

	axRaw := int8(a.Ax / accelerationLSB)
	ayRaw := int8(a.Ay / accelerationLSB)

	data := []byte{byte(axRaw), byte(ayRaw)}
	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing acceleration: %w", err)
	}
	return n, nil
}

func (a *CalculatedAcceleration) Validate() error {
	maxAcc := 31.75 // (127 * 0.25)
	if a.Ax < -maxAcc || a.Ax > maxAcc {
		return fmt.Errorf("Ax out of range: %f (expected ±31 m/s²)", a.Ax)
	}
	if a.Ay < -maxAcc || a.Ay > maxAcc {
		return fmt.Errorf("Ay out of range: %f (expected ±31 m/s²)", a.Ay)
	}
	return nil
}

func (a *CalculatedAcceleration) String() string {
	return fmt.Sprintf("Acceleration: Ax=%.2f m/s², Ay=%.2f m/s²", a.Ax, a.Ay)
}
