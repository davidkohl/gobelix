// dataitems/cat062/uap.go
package uap

import (
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
	cat062 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v120"
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

// UAP062 implements the User Application Profile for ASTERIX Category 062
type UAP062 struct {
	*asterix.BaseUAP
}

// NewUAP062 creates a new instance of the Category 062 UAP
func NewUAP120() (*UAP062, error) {
	base, err := asterix.NewBaseUAP(asterix.Cat062, "1.20", cat062Fields)
	if err != nil {
		return nil, err
	}

	return &UAP062{
		BaseUAP: base,
	}, nil
}

// CreateDataItem creates a new instance of a Cat062 data item
// This is performance-critical - keep it simple and fast
func (u *UAP062) CreateDataItem(id string) (asterix.DataItem, error) {
	switch id {
	case "I062/010":
		return &common.DataSourceIdentifier{}, nil
	case "I062/015":
		return &common.ServiceIdentification{}, nil
	case "I062/040":
		return &cat062.TrackNumber{}, nil
	case "I062/060":
		return &cat062.TrackMode3ACode{}, nil
	case "I062/070":
		return &cat062.TimeOfTrackInformation{}, nil
	case "I062/080":
		return &cat062.TrackStatus{}, nil
	case "I062/100":
		return &cat062.CalculatedTrackPositionCartesian{}, nil
	case "I062/105":
		return &cat062.CalculatedPositionWGS84{}, nil
	case "I062/110":
		return &cat062.Mode5DataReports{}, nil
	case "I062/120":
		return &cat062.TrackMode2Code{}, nil
	case "I062/130":
		return &cat062.CalculatedTrackGeometricAltitude{}, nil
	case "I062/135":
		return &cat062.CalculatedTrackBarometricAltitude{}, nil
	case "I062/136":
		return &cat062.MeasuredFlightLevel{}, nil
	case "I062/185":
		return &cat062.CalculatedTrackVelocity{}, nil
	case "I062/200":
		return &cat062.ModeOfMovement{}, nil
	case "I062/210":
		return &cat062.CalculatedAcceleration{}, nil
	case "I062/220":
		return &cat062.CalculatedRateOfClimbDescent{}, nil
	case "I062/245":
		return &cat062.TargetIdentification{}, nil
	case "I062/270":
		return &cat062.TargetSizeOrientation{}, nil
	case "I062/290":
		return &cat062.SystemTrackUpdateAges{}, nil
	case "I062/295":
		return &cat062.TrackDataAges{}, nil
	case "I062/300":
		return &cat062.VehicleFleetIdentification{}, nil
	case "I062/340":
		return &cat062.MeasuredInformation{}, nil
	case "I062/380":
		return &cat062.AircraftDerivedData{}, nil
	case "I062/390":
		return &cat062.FlightPlanRelatedData{}, nil
	case "I062/500":
		return &cat062.EstimatedAccuracies{}, nil
	case "I062/510":
		return &cat062.ComposedTrackNumber{}, nil
	case "RE062":
		return &cat062.ReservedExpansion{}, nil
	case "SP062":
		return &cat062.SpecialPurpose{}, nil
	default:
		return nil, fmt.Errorf("%w: %s", asterix.ErrUnknownDataItem, id)
	}
}

// Validate implements critical validations for Cat062
func (u *UAP062) Validate(items map[string]asterix.DataItem) error {
	// First do base validation (mandatory fields)
	if err := u.BaseUAP.Validate(items); err != nil {
		return err
	}

	// Critical validations only
	// According to the specification, mandatory items are:
	// - I062/010 Data Source Identifier
	// - I062/040 Track Number
	// - I062/070 Time Of Track Information
	// - I062/080 Track Status

	// These are already checked by the BaseUAP validate, so no additional checks needed

	return nil
}

// cat062Fields defines the complete UAP for Category 062
var cat062Fields = []asterix.DataField{
	{
		FRN:         1,
		DataItem:    "I062/010",
		Description: "Data Source Identifier",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   true,
	},
	{
		FRN:         2,
		DataItem:    "",
		Description: "Spare",
		Type:        asterix.Fixed,
		Length:      0,
		Mandatory:   false,
	},
	{
		FRN:         3,
		DataItem:    "I062/015",
		Description: "Service Identification",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         4,
		DataItem:    "I062/070",
		Description: "Time Of Track Information",
		Type:        asterix.Fixed,
		Length:      3,
		Mandatory:   true,
	},
	{
		FRN:         5,
		DataItem:    "I062/105",
		Description: "Calculated Position in WGS-84 Co-ordinates",
		Type:        asterix.Fixed,
		Length:      8,
		Mandatory:   false,
	},
	{
		FRN:         6,
		DataItem:    "I062/100",
		Description: "Calculated Track Position (Cartesian)",
		Type:        asterix.Fixed,
		Length:      6,
		Mandatory:   false,
	},
	{
		FRN:         7,
		DataItem:    "I062/185",
		Description: "Calculated Track Velocity (Cartesian)",
		Type:        asterix.Fixed,
		Length:      4,
		Mandatory:   false,
	},
	{
		FRN:         8,
		DataItem:    "I062/210",
		Description: "Calculated Acceleration (Cartesian)",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         9,
		DataItem:    "I062/060",
		Description: "Track Mode 3/A Code",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         10,
		DataItem:    "I062/245",
		Description: "Target Identification",
		Type:        asterix.Fixed,
		Length:      7,
		Mandatory:   false,
	},
	{
		FRN:         11,
		DataItem:    "I062/380",
		Description: "Aircraft Derived Data",
		Type:        asterix.Compound,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         12,
		DataItem:    "I062/040",
		Description: "Track Number",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   true,
	},
	{
		FRN:         13,
		DataItem:    "I062/080",
		Description: "Track Status",
		Type:        asterix.Extended,
		Length:      1,
		Mandatory:   true,
	},
	{
		FRN:         14,
		DataItem:    "I062/290",
		Description: "System Track Update Ages",
		Type:        asterix.Compound,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         15,
		DataItem:    "I062/200",
		Description: "Mode of Movement",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         16,
		DataItem:    "I062/295",
		Description: "Track Data Ages",
		Type:        asterix.Compound,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         17,
		DataItem:    "I062/136",
		Description: "Measured Flight Level",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         18,
		DataItem:    "I062/130",
		Description: "Calculated Track Geometric Altitude",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         19,
		DataItem:    "I062/135",
		Description: "Calculated Track Barometric Altitude",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         20,
		DataItem:    "I062/220",
		Description: "Calculated Rate Of Climb/Descent",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         21,
		DataItem:    "I062/390",
		Description: "Flight Plan Related Data",
		Type:        asterix.Compound,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         22,
		DataItem:    "I062/270",
		Description: "Target Size & Orientation",
		Type:        asterix.Extended,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         23,
		DataItem:    "I062/300",
		Description: "Vehicle Fleet Identification",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         24,
		DataItem:    "I062/110",
		Description: "Mode 5 Data reports & Extended Mode 1 Code",
		Type:        asterix.Compound,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         25,
		DataItem:    "I062/120",
		Description: "Track Mode 2 Code",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         26,
		DataItem:    "I062/510",
		Description: "Composed Track Number",
		Type:        asterix.Extended,
		Length:      3,
		Mandatory:   false,
	},
	{
		FRN:         27,
		DataItem:    "I062/500",
		Description: "Estimated Accuracies",
		Type:        asterix.Compound,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         28,
		DataItem:    "I062/340",
		Description: "Measured Information",
		Type:        asterix.Compound,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         29,
		DataItem:    "",
		Description: "Spare",
		Type:        asterix.Fixed,
		Length:      0,
		Mandatory:   false,
	},
	{
		FRN:         30,
		DataItem:    "",
		Description: "Spare",
		Type:        asterix.Fixed,
		Length:      0,
		Mandatory:   false,
	},
	{
		FRN:         31,
		DataItem:    "",
		Description: "Spare",
		Type:        asterix.Fixed,
		Length:      0,
		Mandatory:   false,
	},
	{
		FRN:         32,
		DataItem:    "",
		Description: "Spare",
		Type:        asterix.Fixed,
		Length:      0,
		Mandatory:   false,
	},
	{
		FRN:         33,
		DataItem:    "",
		Description: "Spare",
		Type:        asterix.Fixed,
		Length:      0,
		Mandatory:   false,
	},
	{
		FRN:         34,
		DataItem:    "RE062",
		Description: "Reserved Expansion Field",
		Type:        asterix.Repetitive,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         35,
		DataItem:    "SP062",
		Description: "Special Purpose Field",
		Type:        asterix.Repetitive,
		Length:      1,
		Mandatory:   false,
	},
}
