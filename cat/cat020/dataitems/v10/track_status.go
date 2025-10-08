// cat/cat020/dataitems/v10/track_status.go
package v10

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/davidkohl/gobelix/asterix"
)

// TrackStatus represents I020/170 - Track Status
// Variable length: 1+ octets
// Status of track
type TrackStatus struct {
	// First octet
	CNF bool   // Confirmed track (false) / Track in initiation phase (true)
	TRE bool   // Last report for a track
	CST bool   // Extrapolated
	CDM uint8  // Climb/Descent Mode (0=Maintaining, 1=Climbing, 2=Descending, 3=Invalid)
	MAH bool   // Horizontal manoeuvre
	STH bool   // Smoothed position

	// First extent (if present)
	GHO bool // Ghost track
}

// NewTrackStatus creates a new Track Status data item
func NewTrackStatus() *TrackStatus {
	return &TrackStatus{}
}

// Decode decodes the Track Status from bytes
func (t *TrackStatus) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 1 {
		return 0, fmt.Errorf("%w: need at least 1 byte, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	bytesRead := 0
	data := buf.Next(1)
	bytesRead++

	// First octet
	t.CNF = (data[0] & 0x80) != 0
	t.TRE = (data[0] & 0x40) != 0
	t.CST = (data[0] & 0x20) != 0
	t.CDM = (data[0] >> 3) & 0x03
	t.MAH = (data[0] & 0x04) != 0
	t.STH = (data[0] & 0x02) != 0
	fx := (data[0] & 0x01) != 0

	// First extent (if FX bit is set)
	if fx {
		if buf.Len() < 1 {
			return bytesRead, fmt.Errorf("%w: need 1 more byte for extent", asterix.ErrBufferTooShort)
		}
		data = buf.Next(1)
		bytesRead++

		t.GHO = (data[0] & 0x80) != 0
		// Bits 7-2 are spare
		// fx = (data[0] & 0x01) != 0 // Could extend further
	}

	return bytesRead, nil
}

// Encode encodes the Track Status to bytes
func (t *TrackStatus) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	bytesWritten := 0

	// Determine if we need the first extent
	needsExtent := t.GHO

	// First octet
	var octet byte
	if t.CNF {
		octet |= 0x80
	}
	if t.TRE {
		octet |= 0x40
	}
	if t.CST {
		octet |= 0x20
	}
	octet |= (t.CDM & 0x03) << 3
	if t.MAH {
		octet |= 0x04
	}
	if t.STH {
		octet |= 0x02
	}
	if needsExtent {
		octet |= 0x01 // FX bit
	}

	if err := buf.WriteByte(octet); err != nil {
		return bytesWritten, fmt.Errorf("writing first octet: %w", err)
	}
	bytesWritten++

	// First extent (if needed)
	if needsExtent {
		octet = 0
		if t.GHO {
			octet |= 0x80
		}
		// Bits 7-2 are spare (0)
		// FX bit is 0 (no further extension)

		if err := buf.WriteByte(octet); err != nil {
			return bytesWritten, fmt.Errorf("writing first extent: %w", err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

// Validate validates the Track Status
func (t *TrackStatus) Validate() error {
	if t.CDM > 3 {
		return fmt.Errorf("%w: CDM must be 0-3, got %d", asterix.ErrInvalidMessage, t.CDM)
	}
	return nil
}

// String returns a string representation
func (t *TrackStatus) String() string {
	var parts []string

	if t.CNF {
		parts = append(parts, "CNF")
	} else {
		parts = append(parts, "Confirmed")
	}

	if t.TRE {
		parts = append(parts, "LastReport")
	}

	if t.CST {
		parts = append(parts, "Extrapolated")
	} else {
		parts = append(parts, "NotExtrapolated")
	}

	switch t.CDM {
	case 0:
		parts = append(parts, "Maintaining")
	case 1:
		parts = append(parts, "Climbing")
	case 2:
		parts = append(parts, "Descending")
	case 3:
		parts = append(parts, "CDM=Invalid")
	}

	if t.MAH {
		parts = append(parts, "HorizManoeuvre")
	}

	if t.STH {
		parts = append(parts, "Smoothed")
	} else {
		parts = append(parts, "Measured")
	}

	if t.GHO {
		parts = append(parts, "GHOST")
	}

	return strings.Join(parts, " ")
}
