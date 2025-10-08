// cat/cat020/dataitems/v10/vehicle_fleet_identification.go
package v10

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// VehicleFleetIdentification represents I020/300 - Vehicle Fleet Identification
// Fixed length: 1 byte
// Vehicle fleet identification number
type VehicleFleetIdentification struct {
	VFI uint8 // Vehicle Fleet Identification
}

// NewVehicleFleetIdentification creates a new Vehicle Fleet Identification data item
func NewVehicleFleetIdentification() *VehicleFleetIdentification {
	return &VehicleFleetIdentification{}
}

// Decode decodes the Vehicle Fleet Identification from bytes
func (v *VehicleFleetIdentification) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 1 {
		return 0, fmt.Errorf("%w: need 1 byte, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	data := buf.Next(1)
	v.VFI = data[0]

	return 1, nil
}

// Encode encodes the Vehicle Fleet Identification to bytes
func (v *VehicleFleetIdentification) Encode(buf *bytes.Buffer) (int, error) {
	if err := v.Validate(); err != nil {
		return 0, err
	}

	if err := buf.WriteByte(v.VFI); err != nil {
		return 0, fmt.Errorf("writing VFI: %w", err)
	}

	return 1, nil
}

// Validate validates the Vehicle Fleet Identification
func (v *VehicleFleetIdentification) Validate() error {
	// All values 0-255 are valid
	return nil
}

// String returns a string representation
func (v *VehicleFleetIdentification) String() string {
	vfiStr := ""
	switch v.VFI {
	case 0:
		vfiStr = "Unknown"
	case 1:
		vfiStr = "ATC equipment maintenance"
	case 2:
		vfiStr = "Airport maintenance"
	case 3:
		vfiStr = "Fire"
	case 4:
		vfiStr = "Bird scarer"
	case 5:
		vfiStr = "Snow plough"
	case 6:
		vfiStr = "Runway sweeper"
	case 7:
		vfiStr = "Emergency"
	case 8:
		vfiStr = "Police"
	case 9:
		vfiStr = "Bus"
	case 10:
		vfiStr = "Tug (push/tow)"
	case 11:
		vfiStr = "Grass cutter"
	case 12:
		vfiStr = "Fuel"
	case 13:
		vfiStr = "Baggage"
	case 14:
		vfiStr = "Catering"
	case 15:
		vfiStr = "Aircraft maintenance"
	case 16:
		vfiStr = "Flyco (follow me)"
	default:
		vfiStr = fmt.Sprintf("Reserved(%d)", v.VFI)
	}
	return vfiStr
}
