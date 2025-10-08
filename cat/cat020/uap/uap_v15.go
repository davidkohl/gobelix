// cat/cat020/uap/uap_v15.go
package uap

import (
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
	cat020 "github.com/davidkohl/gobelix/cat/cat020/dataitems/v15"
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

// UAP020 implements the User Application Profile for ASTERIX Category 020
type UAP020 struct {
	*asterix.BaseUAP
}

// NewUAP15 creates a new instance of the Category 020 UAP version 1.5
func NewUAP15() (*UAP020, error) {
	base, err := asterix.NewBaseUAP(asterix.Cat020, "1.5", cat020Fields)
	if err != nil {
		return nil, err
	}

	return &UAP020{
		BaseUAP: base,
	}, nil
}

// CreateDataItem creates a new instance of a Cat020 data item
func (u *UAP020) CreateDataItem(id string) (asterix.DataItem, error) {
	switch id {
	case "I020/010":
		return &common.DataSourceIdentifier{}, nil
	case "I020/020":
		return cat020.NewExtendedStub(id), nil
	case "I020/140":
		return &common.TimeOfDay{}, nil
	case "I020/041":
		return cat020.NewFixedStub(id, 8), nil
	case "I020/042":
		return cat020.NewFixedStub(id, 6), nil
	case "I020/161":
		return cat020.NewFixedStub(id, 2), nil
	case "I020/170":
		return cat020.NewExtendedStub(id), nil
	case "I020/070":
		return cat020.NewFixedStub(id, 2), nil
	case "I020/202":
		return cat020.NewFixedStub(id, 4), nil
	case "I020/090":
		return cat020.NewFixedStub(id, 2), nil
	case "I020/100":
		return cat020.NewFixedStub(id, 4), nil
	case "I020/220":
		return cat020.NewFixedStub(id, 3), nil
	case "I020/245":
		return cat020.NewFixedStub(id, 7), nil
	case "I020/110":
		return cat020.NewFixedStub(id, 2), nil
	case "I020/105":
		return cat020.NewFixedStub(id, 2), nil
	case "I020/210":
		return cat020.NewFixedStub(id, 2), nil
	case "I020/300":
		return cat020.NewFixedStub(id, 1), nil
	case "I020/310":
		return cat020.NewFixedStub(id, 1), nil
	case "I020/500":
		return &cat020.PositionAccuracy{}, nil
	case "I020/400":
		return &cat020.ContributingDevices{}, nil
	case "I020/250":
		return &cat020.ModeSMBData{}, nil
	case "I020/230":
		return cat020.NewFixedStub(id, 2), nil
	case "I020/260":
		return cat020.NewFixedStub(id, 7), nil
	case "I020/030":
		return cat020.NewExtendedStub(id), nil
	case "I020/055":
		return cat020.NewFixedStub(id, 1), nil
	case "I020/050":
		return cat020.NewFixedStub(id, 2), nil
	case "RE020":
		return cat020.NewExplicitStub(id), nil
	case "SP020":
		return cat020.NewExplicitStub(id), nil
	default:
		return nil, fmt.Errorf("%w: %s", asterix.ErrUnknownDataItem, id)
	}
}

// cat020Fields defines the UAP for Category 020 version 1.5
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
		Length:      1,
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
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         8,
		DataItem:    "I020/070",
		Description: "Mode 3/A Code in Octal Representation",
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
		Description: "Mode-C Code in Binary Representation",
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
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         20,
		DataItem:    "I020/400",
		Description: "Contributing Devices",
		Type:        asterix.Repetitive,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         21,
		DataItem:    "I020/250",
		Description: "Mode S MB Data",
		Type:        asterix.Repetitive,
		Length:      8,
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
		Length:      1,
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
	{
		FRN:         27,
		DataItem:    "RE020",
		Description: "Reserved Expansion Field",
		Type:        asterix.Explicit,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         28,
		DataItem:    "SP020",
		Description: "Special Purpose Field",
		Type:        asterix.Explicit,
		Length:      1,
		Mandatory:   false,
	},
}
