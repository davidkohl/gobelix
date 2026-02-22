package v13

import (
	"bytes"
	"fmt"
)

// TargetSizeOrientation implements I011/270 - Target Size & Orientation
// Definition: Target size defined as length and width of the detected target, and orientation.
// Format: Variable length Data Item comprising a first part of one octet,
// followed by one-octet extents as necessary.
type TargetSizeOrientation struct {
	Length      uint8   // Length in meters (7 bits, LSB = 1m)
	Orientation float64 // Orientation in degrees (7 bits, LSB = 360°/128 ≈ 2.81°)
	Width       uint8   // Width in meters (7 bits, LSB = 1m)
	HasOrient   bool    // True if orientation is present
	HasWidth    bool    // True if width is present
}

const orientationLSB = 360.0 / 128.0 // ≈ 2.8125 degrees

func (t *TargetSizeOrientation) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// First octet: LENGTH (bits 8-2), FX (bit 1)
	data, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading target size: %w", err)
	}
	bytesRead++

	t.Length = (data >> 1) & 0x7F
	fx := (data & 0x01) != 0

	if !fx {
		return bytesRead, nil
	}

	// First extent: ORIENTATION (bits 8-2), FX (bit 1)
	data, err = buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading target orientation: %w", err)
	}
	bytesRead++

	t.HasOrient = true
	t.Orientation = float64((data>>1)&0x7F) * orientationLSB
	fx = (data & 0x01) != 0

	if !fx {
		return bytesRead, nil
	}

	// Second extent: WIDTH (bits 8-2), FX (bit 1)
	data, err = buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading target width: %w", err)
	}
	bytesRead++

	t.HasWidth = true
	t.Width = (data >> 1) & 0x7F
	fx = (data & 0x01) != 0

	// Skip any additional extensions
	for fx {
		data, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading target size extension: %w", err)
		}
		bytesRead++
		fx = (data & 0x01) != 0
	}

	return bytesRead, nil
}

func (t *TargetSizeOrientation) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	bytesWritten := 0

	// First octet: LENGTH (bits 8-2), FX (bit 1)
	octet := (t.Length & 0x7F) << 1
	if t.HasOrient {
		octet |= 0x01 // FX
	}

	if err := buf.WriteByte(octet); err != nil {
		return bytesWritten, fmt.Errorf("writing target length: %w", err)
	}
	bytesWritten++

	if !t.HasOrient {
		return bytesWritten, nil
	}

	// First extent: ORIENTATION (bits 8-2), FX (bit 1)
	orientRaw := uint8(t.Orientation / orientationLSB)
	octet = (orientRaw & 0x7F) << 1
	if t.HasWidth {
		octet |= 0x01 // FX
	}

	if err := buf.WriteByte(octet); err != nil {
		return bytesWritten, fmt.Errorf("writing target orientation: %w", err)
	}
	bytesWritten++

	if !t.HasWidth {
		return bytesWritten, nil
	}

	// Second extent: WIDTH (bits 8-2), FX (bit 1) = 0
	octet = (t.Width & 0x7F) << 1

	if err := buf.WriteByte(octet); err != nil {
		return bytesWritten, fmt.Errorf("writing target width: %w", err)
	}
	bytesWritten++

	return bytesWritten, nil
}

func (t *TargetSizeOrientation) Validate() error {
	if t.Length > 127 {
		return fmt.Errorf("length out of range: %d (max 127)", t.Length)
	}
	if t.Orientation < 0 || t.Orientation >= 360 {
		return fmt.Errorf("orientation out of range: %f (expected 0-359)", t.Orientation)
	}
	if t.Width > 127 {
		return fmt.Errorf("width out of range: %d (max 127)", t.Width)
	}
	return nil
}

func (t *TargetSizeOrientation) String() string {
	s := fmt.Sprintf("Target Size: Length=%dm", t.Length)
	if t.HasOrient {
		s += fmt.Sprintf(", Orientation=%.1f°", t.Orientation)
	}
	if t.HasWidth {
		s += fmt.Sprintf(", Width=%dm", t.Width)
	}
	return s
}
