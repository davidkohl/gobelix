// dataitems/cat021/target_status.go
package v26

import (
	"bytes"
	"fmt"
	"strings"
)

// TargetStatus implements I021/200
type TargetStatus struct {
	ICF  bool  // Intent Change Flag
	LNAV bool  // LNAV Mode
	ME   bool  // Military Emergency
	PS   uint8 // Priority Status
	SS   uint8 // Surveillance Status
}

// Priority Status values
const (
	PSNoEmergency uint8 = iota
	PSGeneralEmergency
	PSLifeguard
	PSMinimumFuel
	PSNoCommunications
	PSUnlawfulInterference
	PSDownedAircraft
)

// Surveillance Status values
const (
	SSNoCondition uint8 = iota
	SSPermanentAlert
	SSTemporaryAlert
	SSSPI
)

func (t *TargetStatus) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 1)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading target status: %w", err)
	}
	if n != 1 {
		return n, fmt.Errorf("insufficient data for target status: got %d bytes, want 1", n)
	}

	t.ICF = (data[0] & 0x80) != 0  // bit 8
	t.LNAV = (data[0] & 0x40) != 0 // bit 7
	t.ME = (data[0] & 0x20) != 0   // bit 6
	t.PS = (data[0] >> 3) & 0x07   // bits 5-3
	t.SS = data[0] & 0x03          // bits 2-1

	return n, t.Validate()
}

func (t *TargetStatus) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	var b byte
	if t.ICF {
		b |= 0x80
	}
	if t.LNAV {
		b |= 0x40
	}
	if t.ME {
		b |= 0x20
	}
	b |= (t.PS & 0x07) << 3
	b |= t.SS & 0x03

	err := buf.WriteByte(b)
	if err != nil {
		return 0, fmt.Errorf("writing target status: %w", err)
	}
	return 1, nil
}

func (t *TargetStatus) Validate() error {
	if t.PS > PSDownedAircraft {
		return fmt.Errorf("invalid priority status: %d", t.PS)
	}
	if t.SS > SSSPI {
		return fmt.Errorf("invalid surveillance status: %d", t.SS)
	}
	return nil
}

func (t *TargetStatus) String() string {
	var parts []string

	if t.ICF {
		parts = append(parts, "ICF")
	}
	if t.LNAV {
		parts = append(parts, "LNAV")
	}
	if t.ME {
		parts = append(parts, "ME")
	}

	// Priority Status
	priority := "NoEmergency"
	switch t.PS {
	case PSGeneralEmergency:
		priority = "Emergency"
	case PSLifeguard:
		priority = "Lifeguard"
	case PSMinimumFuel:
		priority = "MinFuel"
	case PSNoCommunications:
		priority = "NoComm"
	case PSUnlawfulInterference:
		priority = "Unlawful"
	case PSDownedAircraft:
		priority = "Downed"
	}
	parts = append(parts, priority)

	// Surveillance Status
	switch t.SS {
	case SSPermanentAlert:
		parts = append(parts, "PermAlert")
	case SSTemporaryAlert:
		parts = append(parts, "TempAlert")
	case SSSPI:
		parts = append(parts, "SPI")
	}

	return strings.Join(parts, ", ")
}
