// dataitems/cat021/quality_indicators.go
package v26

import (
	"bytes"
	"fmt"
	"strings"
)

// QualityIndicators implements I021/090
type QualityIndicators struct {
	// Primary Subfield
	NUCr_NACv uint8 // Navigation Uncertainty Category for velocity or NAC for Velocity
	NUCp_NIC  uint8 // Navigation Uncertainty Category for Position or NIC

	// First Extension
	NICbaro bool  // Navigation Integrity Category for Barometric Altitude
	SIL     uint8 // Surveillance/Source Integrity Level
	NACp    uint8 // Navigation Accuracy Category for Position

	// Second Extension
	SILS bool  // SIL Supplement
	SDA  uint8 // System Design Assurance Level
	GVA  uint8 // Geometric Vertical Accuracy

	// Third Extension
	PIC uint8 // Position Integrity Category

	hasExtensions uint8 // Tracks which extensions are present (0-3)
}

func (q *QualityIndicators) Encode(buf *bytes.Buffer) (int, error) {
	if err := q.Validate(); err != nil {
		return 0, err
	}

	bytesWritten := 0

	// Determine which extensions need to be included
	// Check if first extension needs to be included
	if q.NICbaro || q.SIL > 0 || q.NACp > 0 {
		q.hasExtensions = 1

		// Check if second extension needs to be included
		if q.SILS || q.SDA > 0 || q.GVA > 0 {
			q.hasExtensions = 2

			// Check if third extension needs to be included
			if q.PIC > 0 {
				q.hasExtensions = 3
			}
		}
	}

	// Primary Subfield
	b := (q.NUCr_NACv << 5) | (q.NUCp_NIC << 1)
	if q.hasExtensions > 0 {
		b |= 0x01
	}

	if err := buf.WriteByte(b); err != nil {
		return bytesWritten, fmt.Errorf("writing primary field: %w", err)
	}
	bytesWritten++

	// First Extension
	if q.hasExtensions > 0 {
		b = 0
		if q.NICbaro {
			b |= 0x80
		}
		b |= (q.SIL & 0x03) << 5
		b |= (q.NACp & 0x0F) << 1

		if q.hasExtensions > 1 {
			b |= 0x01
		}

		if err := buf.WriteByte(b); err != nil {
			return bytesWritten, fmt.Errorf("writing first extension: %w", err)
		}
		bytesWritten++
	}

	// Second Extension
	if q.hasExtensions > 1 {
		b = 0
		if q.SILS {
			b |= 0x20
		}
		b |= (q.SDA & 0x03) << 3
		b |= (q.GVA & 0x03) << 1

		if q.hasExtensions > 2 {
			b |= 0x01
		}

		if err := buf.WriteByte(b); err != nil {
			return bytesWritten, fmt.Errorf("writing second extension: %w", err)
		}
		bytesWritten++
	}

	// Third Extension
	if q.hasExtensions > 2 {
		b = (q.PIC & 0x0F) << 4

		if err := buf.WriteByte(b); err != nil {
			return bytesWritten, fmt.Errorf("writing third extension: %w", err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

func (q *QualityIndicators) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Primary Subfield
	b, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading primary field: %w", err)
	}
	bytesRead++

	q.NUCr_NACv = (b >> 5) & 0x07
	q.NUCp_NIC = (b >> 1) & 0x0F
	fx := (b & 0x01) != 0

	// First Extension
	if fx {
		q.hasExtensions = 1
		b, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading first extension: %w", err)
		}
		bytesRead++

		q.NICbaro = (b & 0x80) != 0
		q.SIL = (b >> 5) & 0x03
		q.NACp = (b >> 1) & 0x0F
		fx = (b & 0x01) != 0

		// Second Extension
		if fx {
			q.hasExtensions = 2
			b, err = buf.ReadByte()
			if err != nil {
				return bytesRead, fmt.Errorf("reading second extension: %w", err)
			}
			bytesRead++

			q.SILS = (b & 0x20) != 0
			q.SDA = (b >> 3) & 0x03
			q.GVA = (b >> 1) & 0x03
			fx = (b & 0x01) != 0

			// Third Extension
			if fx {
				q.hasExtensions = 3
				b, err = buf.ReadByte()
				if err != nil {
					return bytesRead, fmt.Errorf("reading third extension: %w", err)
				}
				bytesRead++

				q.PIC = (b >> 4) & 0x0F
			}
		}
	}

	return bytesRead, q.Validate()
}

func (q *QualityIndicators) Validate() error {
	if q.NUCr_NACv > 7 {
		return fmt.Errorf("invalid NUCr/NACv value: %d", q.NUCr_NACv)
	}
	if q.NUCp_NIC > 15 {
		return fmt.Errorf("invalid NUCp/NIC value: %d", q.NUCp_NIC)
	}
	if q.SIL > 3 {
		return fmt.Errorf("invalid SIL value: %d", q.SIL)
	}
	if q.NACp > 15 {
		return fmt.Errorf("invalid NACp value: %d", q.NACp)
	}
	if q.SDA > 3 {
		return fmt.Errorf("invalid SDA value: %d", q.SDA)
	}
	if q.GVA > 3 {
		return fmt.Errorf("invalid GVA value: %d", q.GVA)
	}
	if q.PIC > 15 {
		return fmt.Errorf("invalid PIC value: %d", q.PIC)
	}
	return nil
}

func (q *QualityIndicators) String() string {
	var details []string

	// NUCr/NACv (Navigation Accuracy)
	details = append(details, fmt.Sprintf("NAC-v: %d", q.NUCr_NACv))

	// NUCp/NIC (Navigation Integrity)
	details = append(details, fmt.Sprintf("NIC: %d", q.NUCp_NIC))

	// NICbaro (Barometric Altitude Integrity)
	if q.NICbaro {
		details = append(details, "NICbaro: Valid")
	}

	// SIL (Surveillance/Source Integrity Level)
	details = append(details, fmt.Sprintf("SIL: %d", q.SIL))

	// NACp (Position Accuracy)
	details = append(details, fmt.Sprintf("NACp: %d", q.NACp))

	// System Design Assurance
	if q.hasExtensions > 1 {
		details = append(details, fmt.Sprintf("SDA: %d", q.SDA))
	}

	// Geometric Vertical Accuracy
	if q.hasExtensions > 1 {
		details = append(details, fmt.Sprintf("GVA: %d", q.GVA))
	}

	// Position Integrity Category
	if q.hasExtensions > 2 {
		details = append(details, fmt.Sprintf("PIC: %d", q.PIC))
	}

	return strings.Join(details, ", ")
}
