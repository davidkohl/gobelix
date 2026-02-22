// dataitems/cat021/trajectory_intent.go
package v26

import (
	"bytes"
	"fmt"
	"math"

	"github.com/davidkohl/gobelix/asterix"
)

// TrajectoryIntentPoint represents a single trajectory intent waypoint
type TrajectoryIntentPoint struct {
	NAV bool    // Trajectory Change Point (TCP) number available
	NVB bool    // TCP number valid
	TCP uint8   // Trajectory Change Point number (0-63)
	Latitude  float64 // Latitude in WGS-84, LSB = 180/2^23
	Longitude float64 // Longitude in WGS-84, LSB = 180/2^23

	// Optional fields based on subfield presence
	Altitude      *uint16 // Altitude in feet, LSB = 10 ft
	TimeToGo      *uint32 // Time to TCP in seconds, LSB = 1 s
	Distance      *uint32 // Distance to TCP in NM, LSB = 0.01 NM
	TurnRadius    *uint8  // Turn radius, LSB = 0.1 NM
	TurnDirection *uint8  // Turn direction: 0=left, 1=right
}

// TrajectoryIntent implements I021/110
// Trajectory Intent (Compound with repetitive data)
type TrajectoryIntent struct {
	// Subfield #1: Trajectory Intent Status
	NAV *bool // Trajectory Intent Data is available for this aircraft
	NVB *bool // Trajectory Intent Data is valid

	// Subfield #2: Trajectory Intent Data (Repetitive)
	Points []TrajectoryIntentPoint
}

func (t *TrajectoryIntent) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	bytesWritten := 0

	// Primary subfield indicator
	var primaryField uint8
	if t.NAV != nil {
		primaryField |= 0x80 // bit 8: TIS (Trajectory Intent Status) present
	}
	if len(t.Points) > 0 {
		primaryField |= 0x40 // bit 7: TID (Trajectory Intent Data) present
	}
	// Bits 6-2 are spare
	// Bit 1 is FX (no extension)

	if err := buf.WriteByte(primaryField); err != nil {
		return bytesWritten, fmt.Errorf("writing primary field: %w", err)
	}
	bytesWritten++

	// Write Trajectory Intent Status if present (1 octet)
	if t.NAV != nil {
		var tisField uint8
		if *t.NAV {
			tisField |= 0x80 // NAV bit
		}
		if t.NVB != nil && *t.NVB {
			tisField |= 0x40 // NVB bit
		}
		// Bits 6-1 are spare

		if err := buf.WriteByte(tisField); err != nil {
			return bytesWritten, fmt.Errorf("writing TIS: %w", err)
		}
		bytesWritten++
	}

	// Write Trajectory Intent Data if present (Repetitive)
	if len(t.Points) > 0 {
		// Write repetition factor
		rep := uint8(len(t.Points))
		if rep > 255 {
			return bytesWritten, fmt.Errorf("too many trajectory points: %d (max 255)", len(t.Points))
		}

		if err := buf.WriteByte(rep); err != nil {
			return bytesWritten, fmt.Errorf("writing repetition factor: %w", err)
		}
		bytesWritten++

		// Write each point
		for i, point := range t.Points {
			// Subfield presence indicator
			var subfield uint8
			subfield |= (point.TCP & 0x3F) << 2
			if point.NAV {
				subfield |= 0x02
			}
			if point.NVB {
				subfield |= 0x01
			}

			if err := buf.WriteByte(subfield); err != nil {
				return bytesWritten, fmt.Errorf("writing point %d subfield: %w", i, err)
			}
			bytesWritten++

			// Latitude (3 octets)
			latRaw := int32(math.Round(point.Latitude * math.Pow(2, 23) / 180.0))
			var latBytes [3]byte
			latBytes[0] = byte(latRaw >> 16)
			latBytes[1] = byte(latRaw >> 8)
			latBytes[2] = byte(latRaw)
			n, err := buf.Write(latBytes[:])
			if err != nil {
				return bytesWritten, fmt.Errorf("writing point %d latitude: %w", i, err)
			}
			bytesWritten += n

			// Longitude (3 octets)
			lonRaw := int32(math.Round(point.Longitude * math.Pow(2, 23) / 180.0))
			var lonBytes [3]byte
			lonBytes[0] = byte(lonRaw >> 16)
			lonBytes[1] = byte(lonRaw >> 8)
			lonBytes[2] = byte(lonRaw)
			n, err = buf.Write(lonBytes[:])
			if err != nil {
				return bytesWritten, fmt.Errorf("writing point %d longitude: %w", i, err)
			}
			bytesWritten += n

			// Note: Optional altitude and other fields require compound subfield indicators
			// For now, only lat/lon are encoded per basic ASTERIX CAT021 spec
			// TODO: Implement full compound structure with subfield presence indicators
		}
	}

	return bytesWritten, nil
}

func (t *TrajectoryIntent) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Read primary field
	primaryField, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading primary field: %w", err)
	}
	bytesRead++

	hasTIS := (primaryField & 0x80) != 0
	hasTID := (primaryField & 0x40) != 0

	// Read Trajectory Intent Status if present
	if hasTIS {
		tisField, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading TIS: %w", err)
		}
		bytesRead++

		nav := (tisField & 0x80) != 0
		t.NAV = &nav

		if (tisField & 0x40) != 0 {
			nvb := true
			t.NVB = &nvb
		} else {
			nvb := false
			t.NVB = &nvb
		}
	}

	// Read Trajectory Intent Data if present
	if hasTID {
		// Read repetition factor
		rep, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading repetition factor: %w", err)
		}
		bytesRead++

		t.Points = make([]TrajectoryIntentPoint, rep)
		for i := 0; i < int(rep); i++ {
			// Read subfield indicator
			subfield, err := buf.ReadByte()
			if err != nil {
				return bytesRead, fmt.Errorf("reading point %d subfield: %w", i, err)
			}
			bytesRead++

			t.Points[i].TCP = (subfield >> 2) & 0x3F
			t.Points[i].NAV = (subfield & 0x02) != 0
			t.Points[i].NVB = (subfield & 0x01) != 0

			// Read Latitude (3 octets)
			var latBytes [3]byte
			n, err := buf.Read(latBytes[:])
			if err != nil {
				return bytesRead, fmt.Errorf("reading point %d latitude: %w", i, err)
			}
			if n != 3 {
				return bytesRead, fmt.Errorf("%w: need 3 bytes for point %d latitude, have %d", asterix.ErrBufferTooShort, i, n)
			}
			latRaw := int32(uint32(latBytes[0])<<16 | uint32(latBytes[1])<<8 | uint32(latBytes[2]))
			// Sign extend from 24 bits to 32 bits
			if latRaw&0x800000 != 0 {
				latRaw = latRaw - 0x01000000
			}
			t.Points[i].Latitude = float64(latRaw) * 180.0 / math.Pow(2, 23)
			bytesRead += n

			// Read Longitude (3 octets)
			var lonBytes [3]byte
			n, err = buf.Read(lonBytes[:])
			if err != nil {
				return bytesRead, fmt.Errorf("reading point %d longitude: %w", i, err)
			}
			if n != 3 {
				return bytesRead, fmt.Errorf("%w: need 3 bytes for point %d longitude, have %d", asterix.ErrBufferTooShort, i, n)
			}
			lonRaw := int32(uint32(lonBytes[0])<<16 | uint32(lonBytes[1])<<8 | uint32(lonBytes[2]))
			// Sign extend from 24 bits to 32 bits
			if lonRaw&0x800000 != 0 {
				lonRaw = lonRaw - 0x01000000
			}
			t.Points[i].Longitude = float64(lonRaw) * 180.0 / math.Pow(2, 23)
			bytesRead += n
		}
	}

	return bytesRead, t.Validate()
}

func (t *TrajectoryIntent) Validate() error {
	for i, point := range t.Points {
		if point.TCP > 63 {
			return fmt.Errorf("invalid TCP number for point %d: %d", i, point.TCP)
		}
		if point.Latitude < -90 || point.Latitude > 90 {
			return fmt.Errorf("latitude out of range for point %d: %.6f", i, point.Latitude)
		}
		if point.Longitude < -180 || point.Longitude > 180 {
			return fmt.Errorf("longitude out of range for point %d: %.6f", i, point.Longitude)
		}
	}
	return nil
}

func (t *TrajectoryIntent) String() string {
	status := "Unknown"
	if t.NAV != nil {
		if *t.NAV {
			status = "Available"
		} else {
			status = "Not Available"
		}
	}

	if t.NVB != nil && *t.NVB {
		status += ", Valid"
	}

	return fmt.Sprintf("Trajectory: %s, %d waypoint(s)", status, len(t.Points))
}
