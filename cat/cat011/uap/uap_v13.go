// cat/cat011/uap/uap_v13.go
package uap

import (
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
	cat011 "github.com/davidkohl/gobelix/cat/cat011/dataitems/v13"
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

// UAP011 implements the User Application Profile for ASTERIX Category 011
type UAP011 struct {
	*asterix.BaseUAP
}

// NewUAP13 creates a new instance of the Category 011 UAP version 1.3
func NewUAP13() (*UAP011, error) {
	base, err := asterix.NewBaseUAP(asterix.Cat011, "1.3", cat011Fields)
	if err != nil {
		return nil, err
	}

	return &UAP011{
		BaseUAP: base,
	}, nil
}

// CreateDataItem creates a new instance of a Cat011 data item
func (u *UAP011) CreateDataItem(id string) (asterix.DataItem, error) {
	switch id {
	case "I011/010":
		return &common.DataSourceIdentifier{}, nil
	case "I011/000":
		return &cat011.MessageType{}, nil
	case "I011/015":
		return &cat011.ServiceIdentification{}, nil
	case "I011/140":
		return &common.TimeOfDay{}, nil
	case "I011/041":
		return &cat011.PositionWGS84{}, nil
	case "I011/042":
		return &cat011.PositionCartesian{}, nil
	case "I011/202":
		return &cat011.TrackVelocityCartesian{}, nil
	case "I011/210":
		return &cat011.CalculatedAcceleration{}, nil
	case "I011/060":
		return &cat011.Mode3ACode{}, nil
	case "I011/245":
		return &cat011.TargetIdentification{}, nil
	case "I011/380":
		return &cat011.ModeSRelatedData{}, nil
	case "I011/161":
		return &cat011.TrackNumber{}, nil
	case "I011/170":
		return &cat011.TrackStatus{}, nil
	case "I011/290":
		return &cat011.SystemTrackUpdateAges{}, nil
	case "I011/430":
		return &cat011.PhaseOfFlight{}, nil
	case "I011/090":
		return &cat011.MeasuredFlightLevel{}, nil
	case "I011/093":
		return &cat011.BarometricAltitude{}, nil
	case "I011/092":
		return &cat011.GeometricAltitude{}, nil
	case "I011/215":
		return &cat011.RateOfClimbDescent{}, nil
	case "I011/270":
		return &cat011.TargetSizeOrientation{}, nil
	case "I011/390":
		return &cat011.FlightPlanRelatedData{}, nil
	case "I011/300":
		return &cat011.VehicleFleetIdentification{}, nil
	case "I011/310":
		return &cat011.PreProgrammedMessage{}, nil
	case "I011/500":
		return &cat011.EstimatedAccuracies{}, nil
	case "I011/600":
		return &cat011.AlertMessages{}, nil
	case "I011/605":
		return &cat011.TracksInAlert{}, nil
	case "I011/610":
		return &cat011.HoldbarStatus{}, nil
	case "SP011":
		return cat011.NewExplicitStub("SP011"), nil
	case "RE011":
		return cat011.NewExplicitStub("RE011"), nil
	default:
		return nil, fmt.Errorf("%w: %s", asterix.ErrUnknownDataItem, id)
	}
}

// cat011Fields defines the UAP for Category 011 version 1.3
// Based on Table 3 in EUROCONTROL-SPEC-0149-8
var cat011Fields = []asterix.DataField{
	{
		FRN:         1,
		DataItem:    "I011/010",
		Description: "Data Source Identifier",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   true,
	},
	{
		FRN:         2,
		DataItem:    "I011/000",
		Description: "Message Type",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   true,
	},
	{
		FRN:         3,
		DataItem:    "I011/015",
		Description: "Service Identification",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         4,
		DataItem:    "I011/140",
		Description: "Time of Track Information",
		Type:        asterix.Fixed,
		Length:      3,
		Mandatory:   false,
	},
	{
		FRN:         5,
		DataItem:    "I011/041",
		Description: "Position in WGS-84 Co-ordinates",
		Type:        asterix.Fixed,
		Length:      8,
		Mandatory:   false,
	},
	{
		FRN:         6,
		DataItem:    "I011/042",
		Description: "Calculated Position in Cartesian Co-ordinates",
		Type:        asterix.Fixed,
		Length:      4,
		Mandatory:   false,
	},
	{
		FRN:         7,
		DataItem:    "I011/202",
		Description: "Calculated Track Velocity in Cartesian Co-ordinates",
		Type:        asterix.Fixed,
		Length:      4,
		Mandatory:   false,
	},
	{
		FRN:         8,
		DataItem:    "I011/210",
		Description: "Calculated Acceleration",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         9,
		DataItem:    "I011/060",
		Description: "Mode-3/A Code in Octal Representation",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         10,
		DataItem:    "I011/245",
		Description: "Target Identification",
		Type:        asterix.Fixed,
		Length:      7,
		Mandatory:   false,
	},
	{
		FRN:         11,
		DataItem:    "I011/380",
		Description: "Mode-S / ADS-B Related Data",
		Type:        asterix.Compound,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         12,
		DataItem:    "I011/161",
		Description: "Track Number",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         13,
		DataItem:    "I011/170",
		Description: "Track Status",
		Type:        asterix.Extended,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         14,
		DataItem:    "I011/290",
		Description: "System Track Update Ages",
		Type:        asterix.Compound,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         15,
		DataItem:    "I011/430",
		Description: "Phase of Flight",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         16,
		DataItem:    "I011/090",
		Description: "Measured Flight Level",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         17,
		DataItem:    "I011/093",
		Description: "Calculated Track Barometric Altitude",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         18,
		DataItem:    "I011/092",
		Description: "Calculated Track Geometric Altitude",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         19,
		DataItem:    "I011/215",
		Description: "Calculated Rate of Climb/Descent",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         20,
		DataItem:    "I011/270",
		Description: "Target Size & Orientation",
		Type:        asterix.Extended,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         21,
		DataItem:    "I011/390",
		Description: "Flight Plan Related Data",
		Type:        asterix.Compound,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         22,
		DataItem:    "I011/300",
		Description: "Vehicle Fleet Identification",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         23,
		DataItem:    "I011/310",
		Description: "Pre-programmed Message",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         24,
		DataItem:    "I011/500",
		Description: "Estimated Accuracies",
		Type:        asterix.Compound,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         25,
		DataItem:    "I011/600",
		Description: "Alert Messages",
		Type:        asterix.Fixed,
		Length:      3,
		Mandatory:   false,
	},
	{
		FRN:         26,
		DataItem:    "I011/605",
		Description: "Tracks in Alert",
		Type:        asterix.Repetitive,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         27,
		DataItem:    "I011/610",
		Description: "Holdbar Status",
		Type:        asterix.Repetitive,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         28,
		DataItem:    "SP011",
		Description: "Special Purpose Field",
		Type:        asterix.Explicit,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         29,
		DataItem:    "RE011",
		Description: "Reserved Expansion Field",
		Type:        asterix.Explicit,
		Length:      1,
		Mandatory:   false,
	},
}
