// cat/cat020/uap/uap_v10.go
package uap

import (
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
	v10 "github.com/davidkohl/gobelix/cat/cat020/dataitems/v10"
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

// UAP10 implements the User Application Profile for ASTERIX Category 020 Edition 1.0 (November 2005)
type UAP10 struct {
	*asterix.BaseUAP
}

// NewUAP10 creates a new instance of the Category 020 Edition 1.0 UAP
func NewUAP10() (*UAP10, error) {
	base, err := asterix.NewBaseUAP(asterix.Cat020, "1.0", cat020FieldsV10)
	if err != nil {
		return nil, err
	}

	return &UAP10{
		BaseUAP: base,
	}, nil
}

// CreateDataItem creates a new instance of a Cat020 Edition 1.0 data item
func (u *UAP10) CreateDataItem(id string) (asterix.DataItem, error) {
	switch id {
	case "I020/010":
		return &common.DataSourceIdentifier{}, nil
	case "I020/020":
		return v10.NewTargetReportDescriptor(), nil
	case "I020/140":
		return &common.TimeOfDay{}, nil
	case "I020/041":
		return v10.NewPositionWGS84(), nil
	case "I020/042":
		return v10.NewPositionCartesian(), nil
	case "I020/161":
		return v10.NewTrackNumber(), nil
	case "I020/170":
		return v10.NewTrackStatus(), nil
	case "I020/070":
		return v10.NewMode3ACode(), nil
	case "I020/202":
		return v10.NewCalculatedTrackVelocity(), nil
	case "I020/090":
		return v10.NewFlightLevel(), nil
	case "I020/100":
		return v10.NewModeCCode(), nil
	case "I020/220":
		return v10.NewTargetAddress(), nil
	case "I020/245":
		return v10.NewTargetIdentification(), nil
	case "I020/110":
		return v10.NewMeasuredHeight(), nil
	case "I020/105":
		return v10.NewGeometricAltitude(), nil
	case "I020/210":
		return v10.NewCalculatedAcceleration(), nil
	case "I020/300":
		return v10.NewVehicleFleetIdentification(), nil
	case "I020/310":
		return v10.NewPreprogrammedMessage(), nil
	case "I020/500":
		return v10.NewPositionAccuracy(), nil
	case "I020/400":
		return v10.NewContributingReceivers(), nil
	case "I020/250":
		return v10.NewModeSMBData(), nil
	case "I020/230":
		return v10.NewCommunicationsACAS(), nil
	case "I020/260":
		return v10.NewACASResolutionAdvisory(), nil
	case "I020/030":
		return v10.NewWarningErrorConditions(), nil
	case "I020/055":
		return v10.NewMode1Code(), nil
	case "I020/050":
		return v10.NewMode2Code(), nil
	default:
		return nil, fmt.Errorf("%w: %s", asterix.ErrUnknownDataItem, id)
	}
}

// Validate implements validations for Cat020 Edition 1.0
func (u *UAP10) Validate(items map[string]asterix.DataItem) error {
	// First do base validation (mandatory fields)
	if err := u.BaseUAP.Validate(items); err != nil {
		return err
	}

	// Additional validation: Either I020/041 or I020/042 should be present
	// (I020/041 is mandatory for WAM, I020/042 optional for airport applications)
	_, has041 := items["I020/041"]
	_, has042 := items["I020/042"]

	if !has041 && !has042 {
		return fmt.Errorf("at least one position item (I020/041 or I020/042) must be present")
	}

	return nil
}

// cat020FieldsV10 defines the UAP for Category 020 Edition 1.0 (November 2005)
var cat020FieldsV10 = []asterix.DataField{
	// First FSPEC group (FRN 1-7)
	{
		FRN:         1,
		DataItem:    "I020/010",
		Description: "Data Source Identifier",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   true,
	},
	{
		FRN:         2,
		DataItem:    "I020/020",
		Description: "Target Report Descriptor",
		Type:        asterix.Extended,
		Mandatory:   true,
	},
	{
		FRN:         3,
		DataItem:    "I020/140",
		Description: "Time of Day",
		Type:        asterix.Fixed,
		Length:      3,
		Mandatory:   true,
	},
	{
		FRN:         4,
		DataItem:    "I020/041",
		Description: "Position in WGS-84 Coordinates",
		Type:        asterix.Fixed,
		Length:      8,
		Mandatory:   false, // Mandatory for WAM, optional for airport applications
	},
	{
		FRN:         5,
		DataItem:    "I020/042",
		Description: "Position in Cartesian Coordinates",
		Type:        asterix.Fixed,
		Length:      6,
		Mandatory:   false,
	},
	{
		FRN:         6,
		DataItem:    "I020/161",
		Description: "Track Number",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         7,
		DataItem:    "I020/170",
		Description: "Track Status",
		Type:        asterix.Extended,
		Mandatory:   false,
	},
	// Second FSPEC group (FRN 8-14)
	{
		FRN:         8,
		DataItem:    "I020/070",
		Description: "Mode-3/A Code in Octal Representation",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         9,
		DataItem:    "I020/202",
		Description: "Calculated Track Velocity in Cartesian Coord.",
		Type:        asterix.Fixed,
		Length:      4,
		Mandatory:   false,
	},
	{
		FRN:         10,
		DataItem:    "I020/090",
		Description: "Flight Level in Binary Representation",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         11,
		DataItem:    "I020/100",
		Description: "Mode-C Code",
		Type:        asterix.Fixed,
		Length:      4,
		Mandatory:   false,
	},
	{
		FRN:         12,
		DataItem:    "I020/220",
		Description: "Target Address",
		Type:        asterix.Fixed,
		Length:      3,
		Mandatory:   false,
	},
	{
		FRN:         13,
		DataItem:    "I020/245",
		Description: "Target Identification",
		Type:        asterix.Fixed,
		Length:      7,
		Mandatory:   false,
	},
	{
		FRN:         14,
		DataItem:    "I020/110",
		Description: "Measured Height (Cartesian Coordinates)",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	// Third FSPEC group (FRN 15-21)
	{
		FRN:         15,
		DataItem:    "I020/105",
		Description: "Geometric Altitude (WGS-84)",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         16,
		DataItem:    "I020/210",
		Description: "Calculated Acceleration",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         17,
		DataItem:    "I020/300",
		Description: "Vehicle Fleet Identification",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         18,
		DataItem:    "I020/310",
		Description: "Pre-programmed Message",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         19,
		DataItem:    "I020/500",
		Description: "Position Accuracy",
		Type:        asterix.Compound,
		Mandatory:   false,
	},
	{
		FRN:         20,
		DataItem:    "I020/400",
		Description: "Contributing Receivers",
		Type:        asterix.Repetitive,
		Mandatory:   false,
	},
	{
		FRN:         21,
		DataItem:    "I020/250",
		Description: "Mode S MB Data",
		Type:        asterix.Repetitive,
		Mandatory:   false,
	},
	// Fourth FSPEC group (FRN 22-26)
	{
		FRN:         22,
		DataItem:    "I020/230",
		Description: "Comms/ACAS Capability and Flight Status",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         23,
		DataItem:    "I020/260",
		Description: "ACAS Resolution Advisory Report",
		Type:        asterix.Fixed,
		Length:      7,
		Mandatory:   false,
	},
	{
		FRN:         24,
		DataItem:    "I020/030",
		Description: "Warning/Error Conditions",
		Type:        asterix.Extended,
		Mandatory:   false,
	},
	{
		FRN:         25,
		DataItem:    "I020/055",
		Description: "Mode-1 Code in Octal Representation",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         26,
		DataItem:    "I020/050",
		Description: "Mode-2 Code in Octal Representation",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
}
