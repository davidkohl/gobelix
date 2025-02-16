// dataitems/cat021/aircraft_operational_status.go
package cat021

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// AircraftOperationalStatus implements I021/008
type AircraftOperationalStatus struct {
	Meta    asterix.DataField
	RA      bool  // TCAS Resolution Advisory active
	TC      uint8 // Target Trajectory Change Report Capability
	TS      bool  // Target State Report Capability
	ARV     bool  // Air-Referenced Velocity Report Capability
	CDTIA   bool  // Cockpit Display of Traffic Information airborne
	NotTCAS bool  // TCAS System Status
	SA      bool  // Single Antenna
}

func (a *AircraftOperationalStatus) Encode(buf *bytes.Buffer) (int, error) {
	if err := a.Validate(); err != nil {
		return 0, err
	}

	var b uint8
	if a.RA {
		b |= 0x80
	}
	b |= (a.TC & 0x03) << 5
	if a.TS {
		b |= 0x10
	}
	if a.ARV {
		b |= 0x08
	}
	if a.CDTIA {
		b |= 0x04
	}
	if a.NotTCAS {
		b |= 0x02
	}
	if a.SA {
		b |= 0x01
	}

	err := buf.WriteByte(b)
	if err != nil {
		return 0, fmt.Errorf("writing aircraft operational status: %w", err)
	}
	return 1, nil
}

func (a *AircraftOperationalStatus) Decode(buf *bytes.Buffer) (int, error) {
	b, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading aircraft operational status: %w", err)
	}

	a.RA = (b & 0x80) != 0
	a.TC = (b >> 5) & 0x03
	a.TS = (b & 0x10) != 0
	a.ARV = (b & 0x08) != 0
	a.CDTIA = (b & 0x04) != 0
	a.NotTCAS = (b & 0x02) != 0
	a.SA = (b & 0x01) != 0

	return 1, a.Validate()
}

func (a *AircraftOperationalStatus) Validate() error {
	if a.TC > 3 {
		return fmt.Errorf("invalid trajectory change capability: %d", a.TC)
	}
	return nil
}
