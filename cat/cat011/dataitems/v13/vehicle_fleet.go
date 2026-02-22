package v13

import (
	"bytes"
	"fmt"
)

// VehicleFleetIdentification implements I011/300 - Vehicle Fleet Identification
// Definition: Vehicle fleet identification number.
// Format: One octet fixed length Data Item.
type VehicleFleetIdentification struct {
	VFI uint8 // Vehicle Fleet Identification (0-16)
}

// Vehicle fleet constants
const (
	VFIFlyco             uint8 = 0  // Flyco (follow me)
	VFIATCMaintenance    uint8 = 1  // ATC equipment maintenance
	VFIAirportMaint      uint8 = 2  // Airport maintenance
	VFIFire              uint8 = 3  // Fire
	VFIBirdScarer        uint8 = 4  // Bird scarer
	VFISnowPlough        uint8 = 5  // Snow plough
	VFIRunwaySweeper     uint8 = 6  // Runway sweeper
	VFIEmergency         uint8 = 7  // Emergency
	VFIPolice            uint8 = 8  // Police
	VFIBus               uint8 = 9  // Bus
	VFITug               uint8 = 10 // Tug (push/tow)
	VFIGrassCutter       uint8 = 11 // Grass cutter
	VFIFuel              uint8 = 12 // Fuel
	VFIBaggage           uint8 = 13 // Baggage
	VFICatering          uint8 = 14 // Catering
	VFIAircraftMaint     uint8 = 15 // Aircraft maintenance
	VFIUnknown           uint8 = 16 // Unknown
)

func (v *VehicleFleetIdentification) Decode(buf *bytes.Buffer) (int, error) {
	data, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading vehicle fleet ID: %w", err)
	}

	v.VFI = data
	return 1, nil
}

func (v *VehicleFleetIdentification) Encode(buf *bytes.Buffer) (int, error) {
	if err := buf.WriteByte(v.VFI); err != nil {
		return 0, fmt.Errorf("writing vehicle fleet ID: %w", err)
	}
	return 1, nil
}

func (v *VehicleFleetIdentification) Validate() error {
	if v.VFI > 16 {
		return fmt.Errorf("VFI out of range: %d (expected 0-16)", v.VFI)
	}
	return nil
}

func (v *VehicleFleetIdentification) String() string {
	fleetType := ""
	switch v.VFI {
	case VFIFlyco:
		fleetType = "Flyco (follow me)"
	case VFIATCMaintenance:
		fleetType = "ATC equipment maintenance"
	case VFIAirportMaint:
		fleetType = "Airport maintenance"
	case VFIFire:
		fleetType = "Fire"
	case VFIBirdScarer:
		fleetType = "Bird scarer"
	case VFISnowPlough:
		fleetType = "Snow plough"
	case VFIRunwaySweeper:
		fleetType = "Runway sweeper"
	case VFIEmergency:
		fleetType = "Emergency"
	case VFIPolice:
		fleetType = "Police"
	case VFIBus:
		fleetType = "Bus"
	case VFITug:
		fleetType = "Tug (push/tow)"
	case VFIGrassCutter:
		fleetType = "Grass cutter"
	case VFIFuel:
		fleetType = "Fuel"
	case VFIBaggage:
		fleetType = "Baggage"
	case VFICatering:
		fleetType = "Catering"
	case VFIAircraftMaint:
		fleetType = "Aircraft maintenance"
	case VFIUnknown:
		fleetType = "Unknown"
	default:
		fleetType = "Reserved"
	}
	return fmt.Sprintf("Vehicle Fleet: %s (%d)", fleetType, v.VFI)
}

// PreProgrammedMessage implements I011/310 - Pre-programmed Message
// Definition: Number related to a pre-programmed message that can be transmitted by a vehicle.
// Format: One octet fixed length Data Item.
type PreProgrammedMessage struct {
	TRB bool  // bit 8: 0=Default, 1=In Trouble
	MSG uint8 // bits 7-1: Message code
}

// Pre-programmed message constants
const (
	MSGTowing       uint8 = 1 // Towing aircraft
	MSGFollowMe     uint8 = 2 // "Follow me" operation
	MSGRunwayCheck  uint8 = 3 // Runway check
	MSGEmergencyOp  uint8 = 4 // Emergency operation (fire, medical...)
	MSGWorkProgress uint8 = 5 // Work in progress (maintenance, bird scarer, sweepers...)
)

func (p *PreProgrammedMessage) Decode(buf *bytes.Buffer) (int, error) {
	data, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading pre-programmed message: %w", err)
	}

	p.TRB = (data & 0x80) != 0
	p.MSG = data & 0x7F

	return 1, nil
}

func (p *PreProgrammedMessage) Encode(buf *bytes.Buffer) (int, error) {
	var data uint8
	if p.TRB {
		data = 0x80
	}
	data |= (p.MSG & 0x7F)

	if err := buf.WriteByte(data); err != nil {
		return 0, fmt.Errorf("writing pre-programmed message: %w", err)
	}
	return 1, nil
}

func (p *PreProgrammedMessage) Validate() error {
	if p.MSG > 127 {
		return fmt.Errorf("MSG out of range: %d (max 127)", p.MSG)
	}
	return nil
}

func (p *PreProgrammedMessage) String() string {
	trouble := ""
	if p.TRB {
		trouble = " [TROUBLE]"
	}

	msgStr := ""
	switch p.MSG {
	case MSGTowing:
		msgStr = "Towing aircraft"
	case MSGFollowMe:
		msgStr = "Follow me operation"
	case MSGRunwayCheck:
		msgStr = "Runway check"
	case MSGEmergencyOp:
		msgStr = "Emergency operation"
	case MSGWorkProgress:
		msgStr = "Work in progress"
	default:
		msgStr = fmt.Sprintf("Message %d", p.MSG)
	}

	return fmt.Sprintf("Pre-programmed: %s%s", msgStr, trouble)
}

// PhaseOfFlight implements I011/430 - Phase of Flight
// Definition: Current phase of the flight.
// Format: One-octet fixed length Data Item.
type PhaseOfFlight struct {
	FLS uint8 // Flight status (0-9)
}

// Phase of flight constants
const (
	FLSUnknown          uint8 = 0 // Unknown
	FLSOnStand          uint8 = 1 // On stand
	FLSTaxiingDeparture uint8 = 2 // Taxiing for departure
	FLSTaxiingArrival   uint8 = 3 // Taxiing for arrival
	FLSRunwayDeparture  uint8 = 4 // Runway for departure
	FLSRunwayArrival    uint8 = 5 // Runway for arrival
	FLSHoldDeparture    uint8 = 6 // Hold for departure
	FLSHoldArrival      uint8 = 7 // Hold for arrival
	FLSPushBack         uint8 = 8 // Push back
	FLSOnFinals         uint8 = 9 // On finals
)

func (p *PhaseOfFlight) Decode(buf *bytes.Buffer) (int, error) {
	data, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading phase of flight: %w", err)
	}

	p.FLS = data
	return 1, nil
}

func (p *PhaseOfFlight) Encode(buf *bytes.Buffer) (int, error) {
	if err := buf.WriteByte(p.FLS); err != nil {
		return 0, fmt.Errorf("writing phase of flight: %w", err)
	}
	return 1, nil
}

func (p *PhaseOfFlight) Validate() error {
	if p.FLS > 9 {
		return fmt.Errorf("FLS out of range: %d (expected 0-9)", p.FLS)
	}
	return nil
}

func (p *PhaseOfFlight) String() string {
	phaseStr := ""
	switch p.FLS {
	case FLSUnknown:
		phaseStr = "Unknown"
	case FLSOnStand:
		phaseStr = "On stand"
	case FLSTaxiingDeparture:
		phaseStr = "Taxiing for departure"
	case FLSTaxiingArrival:
		phaseStr = "Taxiing for arrival"
	case FLSRunwayDeparture:
		phaseStr = "Runway for departure"
	case FLSRunwayArrival:
		phaseStr = "Runway for arrival"
	case FLSHoldDeparture:
		phaseStr = "Hold for departure"
	case FLSHoldArrival:
		phaseStr = "Hold for arrival"
	case FLSPushBack:
		phaseStr = "Push back"
	case FLSOnFinals:
		phaseStr = "On finals"
	default:
		phaseStr = "Reserved"
	}
	return fmt.Sprintf("Phase of Flight: %s (%d)", phaseStr, p.FLS)
}
