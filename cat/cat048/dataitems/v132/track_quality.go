// dataitems/cat048/track_quality.go
package v132

import (
	"bytes"
	"fmt"
)

// TrackQuality implements I048/210
// Track quality in the form of a vector of standard deviations.
type TrackQuality struct {
	SigmaX float64 // Standard deviation on X-axis (NM)
	SigmaY float64 // Standard deviation on Y-axis (NM)
	SigmaV float64 // Standard deviation on ground speed (NM/s)
	SigmaH float64 // Standard deviation on heading (degrees)
}

// Decode implements the DataItem interface
func (t *TrackQuality) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 4)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading track quality: %w", err)
	}
	if n != 4 {
		return n, fmt.Errorf("insufficient data for track quality: got %d bytes, want 4", n)
	}

	// Sigma X (1 byte), LSB = 1/128 NM = 1/128 * 1852m ≈ 14.5m
	t.SigmaX = float64(data[0]) / 128.0

	// Sigma Y (1 byte), LSB = 1/128 NM
	t.SigmaY = float64(data[1]) / 128.0

	// Sigma V (1 byte), LSB = 2^-14 NM/s ≈ 0.00006 NM/s ≈ 0.22 kt
	t.SigmaV = float64(data[2]) / 16384.0

	// Sigma H (1 byte), LSB = 360/2^12 degrees ≈ 0.088 degrees
	t.SigmaH = float64(data[3]) * (360.0 / 4096.0)

	return n, nil
}

// Encode implements the DataItem interface
func (t *TrackQuality) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	// Convert to raw values
	sigmaXRaw := uint8(t.SigmaX * 128.0)
	if t.SigmaX >= 2.0 { // Max value for 8 bits at this resolution
		sigmaXRaw = 255
	}

	sigmaYRaw := uint8(t.SigmaY * 128.0)
	if t.SigmaY >= 2.0 {
		sigmaYRaw = 255
	}

	sigmaVRaw := uint8(t.SigmaV * 16384.0)
	if t.SigmaV >= 0.015625 { // Max value for 8 bits at this resolution (≈ 56.25 kt)
		sigmaVRaw = 255
	}

	sigmaHRaw := uint8(t.SigmaH * (4096.0 / 360.0))
	if t.SigmaH >= 22.5 { // Max value for 8 bits at this resolution
		sigmaHRaw = 255
	}

	data := []byte{sigmaXRaw, sigmaYRaw, sigmaVRaw, sigmaHRaw}

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing track quality: %w", err)
	}
	return n, nil
}

// Validate implements the DataItem interface
func (t *TrackQuality) Validate() error {
	// Standard deviations are always positive
	if t.SigmaX < 0 {
		return fmt.Errorf("negative X standard deviation not allowed: %f", t.SigmaX)
	}
	if t.SigmaY < 0 {
		return fmt.Errorf("negative Y standard deviation not allowed: %f", t.SigmaY)
	}
	if t.SigmaV < 0 {
		return fmt.Errorf("negative speed standard deviation not allowed: %f", t.SigmaV)
	}
	if t.SigmaH < 0 {
		return fmt.Errorf("negative heading standard deviation not allowed: %f", t.SigmaH)
	}

	// Check maximum values (based on 8-bit representation)
	if t.SigmaX > 2.0 {
		return fmt.Errorf("the X standard deviation too large: %f NM (max 2.0 NM)", t.SigmaX)
	}
	if t.SigmaY > 2.0 {
		return fmt.Errorf("the Y standard deviation too large: %f NM (max 2.0 NM)", t.SigmaY)
	}
	if t.SigmaV > 0.015625 {
		return fmt.Errorf("speed standard deviation too large: %f NM/s (max 0.015625 NM/s)", t.SigmaV)
	}
	if t.SigmaH > 22.5 {
		return fmt.Errorf("heading standard deviation too large: %f degrees (max 22.5 degrees)", t.SigmaH)
	}

	return nil
}

// String returns a human-readable representation
func (t *TrackQuality) String() string {
	return fmt.Sprintf("σX: %.5f NM, σY: %.5f NM, σV: %.5f NM/s (%.1f kt), σH: %.3f°",
		t.SigmaX, t.SigmaY, t.SigmaV, t.SigmaV*3600.0, t.SigmaH)
}
