// dataitems/cat048/track_status.go
package v132

import (
	"bytes"
	"fmt"
	"strings"
)

// TrackStatus implements I048/170
// Status of monoradar track (PSR and/or SSR updated).
type TrackStatus struct {
	// First Part
	CNF bool  // Confirmed vs. Tentative Track
	RAD uint8 // Type of Sensor(s) maintaining Track
	DOU bool  // Confidence in plot to track association
	MAH bool  // Manoeuvre detection in Horizontal Sense
	CDM uint8 // Climbing/Descending Mode

	// First Extension (if present)
	TRE bool // End of Track lifetime
	GHO bool // Ghost vs. true target
	SUP bool // Track maintained with network track info
	TCC bool // Type of plot coordinate transformation

	// Track which extensions are present
	extension bool
}

// Decode implements the DataItem interface
func (t *TrackStatus) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// First Part
	b, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading track status: %w", err)
	}
	bytesRead++

	t.CNF = (b & 0x80) != 0 // bit 8
	t.RAD = (b >> 6) & 0x03 // bits 7-6
	t.DOU = (b & 0x20) != 0 // bit 5
	t.MAH = (b & 0x10) != 0 // bit 4
	t.CDM = (b >> 2) & 0x03 // bits 3-2
	fx := (b & 0x01) != 0   // bit 1 (FX)

	// First Extension
	if fx {
		t.extension = true
		b, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading track status extension: %w", err)
		}
		bytesRead++

		t.TRE = (b & 0x80) != 0 // bit 8
		t.GHO = (b & 0x40) != 0 // bit 7
		t.SUP = (b & 0x20) != 0 // bit 6
		t.TCC = (b & 0x10) != 0 // bit 5
		// bits 4-2 are spare
		// bit 1 (FX) is not used in this specification
	}

	return bytesRead, t.Validate()
}

// Encode implements the DataItem interface
func (t *TrackStatus) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	bytesWritten := 0

	// First Part
	b := byte(0)
	if t.CNF {
		b |= 0x80 // bit 8
	}
	b |= (t.RAD & 0x03) << 6 // bits 7-6
	if t.DOU {
		b |= 0x20 // bit 5
	}
	if t.MAH {
		b |= 0x10 // bit 4
	}
	b |= (t.CDM & 0x03) << 2 // bits 3-2
	if t.extension {
		b |= 0x01 // bit 1 (FX)
	}

	err := buf.WriteByte(b)
	if err != nil {
		return bytesWritten, fmt.Errorf("writing track status: %w", err)
	}
	bytesWritten++

	// First Extension
	if t.extension {
		b = 0
		if t.TRE {
			b |= 0x80 // bit 8
		}
		if t.GHO {
			b |= 0x40 // bit 7
		}
		if t.SUP {
			b |= 0x20 // bit 6
		}
		if t.TCC {
			b |= 0x10 // bit 5
		}
		// bits 4-2 are spare
		// bit 1 (FX) is not used in this specification

		err := buf.WriteByte(b)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing track status extension: %w", err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

// Validate implements the DataItem interface
func (t *TrackStatus) Validate() error {
	if t.RAD > 3 {
		return fmt.Errorf("invalid RAD value: %d", t.RAD)
	}
	if t.CDM > 3 {
		return fmt.Errorf("invalid CDM value: %d", t.CDM)
	}
	return nil
}

// String returns a human-readable representation
func (t *TrackStatus) String() string {
	var parts []string

	// Confirmed/Tentative status
	if t.CNF {
		parts = append(parts, "Tentative")
	} else {
		parts = append(parts, "Confirmed")
	}

	// Track type
	switch t.RAD {
	case 0:
		parts = append(parts, "Combined")
	case 1:
		parts = append(parts, "PSR")
	case 2:
		parts = append(parts, "SSR/Mode S")
	case 3:
		parts = append(parts, "Invalid")
	}

	// Association confidence
	if t.DOU {
		parts = append(parts, "Low Confidence")
	}

	// Horizontal maneuver
	if t.MAH {
		parts = append(parts, "Horizontal Maneuver")
	}

	// Climbing/Descending
	switch t.CDM {
	case 0:
		parts = append(parts, "Maintaining")
	case 1:
		parts = append(parts, "Climbing")
	case 2:
		parts = append(parts, "Descending")
	case 3:
		parts = append(parts, "Unknown Vertical")
	}

	// Extension fields
	if t.extension {
		if t.TRE {
			parts = append(parts, "End of Track")
		}

		if t.GHO {
			parts = append(parts, "Ghost")
		}

		if t.SUP {
			parts = append(parts, "Network Assisted")
		}

		if t.TCC {
			parts = append(parts, "Slant Corrected")
		} else {
			parts = append(parts, "Radar Plane")
		}
	}

	return strings.Join(parts, ", ")
}

// SetExtension sets the extension flag based on whether any extension fields are used
func (t *TrackStatus) SetExtension() {
	t.extension = t.TRE || t.GHO || t.SUP || t.TCC
}
