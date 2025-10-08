// cat/cat020/uap/uap_v110.go
package uap

import (
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
	v110 "github.com/davidkohl/gobelix/cat/cat020/dataitems/v110"
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

// UAP110 implements the User Application Profile for ASTERIX Category 020 v1.10
type UAP110 struct {
	*asterix.BaseUAP
}

// NewUAP110 creates a new instance of the Category 020 v1.10 UAP
func NewUAP110() (*UAP110, error) {
	base, err := asterix.NewBaseUAP(asterix.Cat020, "1.10", cat020Fields)
	if err != nil {
		return nil, err
	}

	return &UAP110{
		BaseUAP: base,
	}, nil
}

// CreateDataItem creates a new instance of a Cat020 data item
func (u *UAP110) CreateDataItem(id string) (asterix.DataItem, error) {
	switch id {
	case "I020/010":
		return &common.DataSourceIdentifier{}, nil
	case "I020/020":
		return v110.NewTargetReportDescriptor(), nil
	case "I020/140":
		return &common.TimeOfDay{}, nil
	case "I020/041":
		return v110.NewPositionWGS84(), nil
	case "I020/042":
		return v110.NewPositionCartesian(), nil
	case "I020/161":
		return v110.NewTrackNumber(), nil
	case "I020/170":
		return v110.NewTrackStatus(), nil
	case "I020/070":
		return v110.NewMode3ACode(), nil
	case "I020/202":
		return v110.NewCalculatedTrackVelocity(), nil
	case "I020/090":
		return v110.NewFlightLevel(), nil
	case "I020/100":
		return v110.NewModeCCode(), nil
	case "I020/220":
		return v110.NewTargetAddress(), nil
	case "I020/245":
		return v110.NewTargetIdentification(), nil
	case "I020/110":
		return v110.NewMeasuredHeight(), nil
	case "I020/105":
		return v110.NewGeometricHeight(), nil
	case "I020/210":
		return v110.NewCalculatedAcceleration(), nil
	case "I020/300":
		return v110.NewVehicleFleetIdentification(), nil
	case "I020/310":
		return v110.NewPreprogrammedMessage(), nil
	case "I020/500":
		return v110.NewPositionAccuracy(), nil
	case "I020/400":
		return v110.NewContributingDevices(), nil
	case "I020/250":
		return v110.NewBDSRegisterData(), nil
	case "I020/230":
		return v110.NewCommunicationsACAS(), nil
	case "I020/260":
		return v110.NewACASResolutionAdvisory(), nil
	case "I020/030":
		return v110.NewWarningErrorConditions(), nil
	case "I020/055":
		return v110.NewMode1Code(), nil
	case "I020/050":
		return v110.NewMode2Code(), nil
	default:
		return nil, fmt.Errorf("%w: %s", asterix.ErrUnknownDataItem, id)
	}
}

// Validate implements validations for Cat020
func (u *UAP110) Validate(items map[string]asterix.DataItem) error {
	// First do base validation (mandatory fields)
	if err := u.BaseUAP.Validate(items); err != nil {
		return err
	}

	// No additional validations needed for Cat020
	return nil
}

// cat020Fields defines the complete UAP for Category 020 v1.10
var cat020Fields = []asterix.DataField{
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
		Mandatory:   false,
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
		Mandatory:   false,
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
		Description: "Calculated Track Velocity in Cartesian Coordinates",
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
		Description: "Mode C Code",
		Type:        asterix.Fixed,
		Length:      2,
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
		Description: "Measured Height",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         15,
		DataItem:    "I020/105",
		Description: "Geometric Height",
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
		Description: "Contributing Devices",
		Type:        asterix.Repetitive,
		Mandatory:   false,
	},
	{
		FRN:         21,
		DataItem:    "I020/250",
		Description: "BDS Register Data",
		Type:        asterix.Repetitive,
		Mandatory:   false,
	},
	{
		FRN:         22,
		DataItem:    "I020/230",
		Description: "Communications/ACAS Capability and Flight Status",
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
