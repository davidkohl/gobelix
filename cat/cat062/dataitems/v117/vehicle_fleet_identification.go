// dataitems/cat062/vehicle_fleet_identification.go
package v117

import (
	"bytes"
	"fmt"
)

// VehicleFleetType represents the type of vehicle fleet
type VehicleFleetType uint8

const (
	UnknownVehicle VehicleFleetType = iota
	ATCEquipmentMaintenance
	AirportMaintenance
	Fire
	BirdScarer
	SnowPlough
	RunwaySweeper
	Emergency
	Police
	Bus
	Tug
	GrassCutter
	Fuel
	Baggage
	Catering
	AircraftMaintenance
	Flyco
)

// VehicleFleetIdentification implements I062/300
// Vehicle fleet identification number
type VehicleFleetIdentification struct {
	VehicleType VehicleFleetType
}

func (v *VehicleFleetIdentification) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 1)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading vehicle fleet identification: %w", err)
	}
	if n != 1 {
		return n, fmt.Errorf("insufficient data for vehicle fleet identification: got %d bytes, want 1", n)
	}

	v.VehicleType = VehicleFleetType(data[0])

	return n, nil
}

func (v *VehicleFleetIdentification) Encode(buf *bytes.Buffer) (int, error) {
	if err := v.Validate(); err != nil {
		return 0, err
	}

	err := buf.WriteByte(byte(v.VehicleType))
	if err != nil {
		return 0, fmt.Errorf("writing vehicle fleet identification: %w", err)
	}
	return 1, nil
}

func (v *VehicleFleetIdentification) Validate() error {
	if v.VehicleType > Flyco {
		return fmt.Errorf("invalid vehicle fleet type: %d", v.VehicleType)
	}
	return nil
}

func (v *VehicleFleetIdentification) String() string {
	typeStr := "Unknown"

	switch v.VehicleType {
	case ATCEquipmentMaintenance:
		typeStr = "ATC Equipment Maintenance"
	case AirportMaintenance:
		typeStr = "Airport Maintenance"
	case Fire:
		typeStr = "Fire"
	case BirdScarer:
		typeStr = "Bird Scarer"
	case SnowPlough:
		typeStr = "Snow Plough"
	case RunwaySweeper:
		typeStr = "Runway Sweeper"
	case Emergency:
		typeStr = "Emergency"
	case Police:
		typeStr = "Police"
	case Bus:
		typeStr = "Bus"
	case Tug:
		typeStr = "Tug (Push/Tow)"
	case GrassCutter:
		typeStr = "Grass Cutter"
	case Fuel:
		typeStr = "Fuel"
	case Baggage:
		typeStr = "Baggage"
	case Catering:
		typeStr = "Catering"
	case AircraftMaintenance:
		typeStr = "Aircraft Maintenance"
	case Flyco:
		typeStr = "Flyco (Follow Me)"
	}

	return fmt.Sprintf("Vehicle: %s", typeStr)
}
