// cat/cat020/dataitems/v15/vehicle_fleet.go
package v15

import (
	"bytes"
	"fmt"
)

// VehicleFleetIdentification implements I020/300 - Vehicle Fleet Identification
type VehicleFleetIdentification struct {
	VFI uint8 // Vehicle Fleet Identification (0-15)
}

func (v *VehicleFleetIdentification) Decode(buf *bytes.Buffer) (int, error) {
	data, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading vehicle fleet ID: %w", err)
	}

	// Bits 8-5: spare
	// Bits 4-1: VFI
	v.VFI = data & 0x0F

	return 1, nil
}

func (v *VehicleFleetIdentification) Encode(buf *bytes.Buffer) (int, error) {
	// Bits 8-5: spare (0)
	// Bits 4-1: VFI
	data := v.VFI & 0x0F

	err := buf.WriteByte(data)
	if err != nil {
		return 0, fmt.Errorf("writing vehicle fleet ID: %w", err)
	}
	return 1, nil
}

func (v *VehicleFleetIdentification) Validate() error {
	if v.VFI > 15 {
		return fmt.Errorf("VFI out of range: %d", v.VFI)
	}
	return nil
}

func (v *VehicleFleetIdentification) String() string {
	fleetType := ""
	switch v.VFI {
	case 0:
		fleetType = "Unknown"
	case 1:
		fleetType = "ATC equipment maintenance"
	case 2:
		fleetType = "Airport maintenance"
	case 3:
		fleetType = "Fire"
	case 4:
		fleetType = "Bird scarer"
	case 5:
		fleetType = "Snow plough"
	case 6:
		fleetType = "Runway sweeper"
	case 7:
		fleetType = "Emergency"
	case 8:
		fleetType = "Police"
	case 9:
		fleetType = "Bus"
	case 10:
		fleetType = "Tug (push/tow)"
	case 11:
		fleetType = "Grass cutter"
	case 12:
		fleetType = "Fuel"
	case 13:
		fleetType = "Baggage"
	case 14:
		fleetType = "Catering"
	case 15:
		fleetType = "Aircraft maintenance"
	}
	return fmt.Sprintf("VFI: %s (%d)", fleetType, v.VFI)
}

// PreProgrammedMessage implements I020/310 - Pre-programmed Message
type PreProgrammedMessage struct {
	TRB uint8 // Trouble indicator (0-1)
	MSG uint8 // Pre-programmed message (0-127)
}

func (p *PreProgrammedMessage) Decode(buf *bytes.Buffer) (int, error) {
	data, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading pre-programmed message: %w", err)
	}

	// Bit 8: TRB
	// Bits 7-1: MSG
	p.TRB = (data >> 7) & 0x01
	p.MSG = data & 0x7F

	return 1, nil
}

func (p *PreProgrammedMessage) Encode(buf *bytes.Buffer) (int, error) {
	data := ((p.TRB & 0x01) << 7) | (p.MSG & 0x7F)

	err := buf.WriteByte(data)
	if err != nil {
		return 0, fmt.Errorf("writing pre-programmed message: %w", err)
	}
	return 1, nil
}

func (p *PreProgrammedMessage) Validate() error {
	if p.TRB > 1 {
		return fmt.Errorf("TRB out of range: %d", p.TRB)
	}
	if p.MSG > 127 {
		return fmt.Errorf("MSG out of range: %d", p.MSG)
	}
	return nil
}

func (p *PreProgrammedMessage) String() string {
	trouble := ""
	if p.TRB == 1 {
		trouble = " [TROUBLE]"
	}
	return fmt.Sprintf("MSG: %d%s", p.MSG, trouble)
}
