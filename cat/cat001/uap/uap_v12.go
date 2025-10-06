// cat/cat001/uap/uap_v12.go
package uap

import (
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
	v12 "github.com/davidkohl/gobelix/cat/cat001/dataitems/v12"
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

// UAP12 implements the User Application Profile for ASTERIX Category 001 v1.2
type UAP12 struct {
	*asterix.BaseUAP
}

// NewUAP12 creates a new instance of the Category 001 v1.2 UAP
func NewUAP12() (*UAP12, error) {
	base, err := asterix.NewBaseUAP(asterix.Cat001, "1.2", cat001Fields)
	if err != nil {
		return nil, err
	}

	return &UAP12{
		BaseUAP: base,
	}, nil
}

// CreateDataItem creates a new instance of a Cat001 data item
func (u *UAP12) CreateDataItem(id string) (asterix.DataItem, error) {
	switch id {
	case "I001/010":
		return &common.DataSourceIdentifier{}, nil
	case "I001/020":
		return &v12.TargetReportDescriptor{}, nil
	case "I001/030":
		return &v12.WarningErrorConditions{}, nil
	case "I001/040":
		return &v12.PositionPolar{}, nil
	case "I001/050":
		return &v12.Mode2Code{}, nil
	case "I001/060":
		return &v12.Mode2CodeConfidence{}, nil
	case "I001/070":
		return &v12.Mode3ACode{}, nil
	case "I001/080":
		return &v12.Mode3ACodeConfidence{}, nil
	case "I001/090":
		return &v12.ModeCCode{}, nil
	case "I001/100":
		return &v12.ModeCCodeConfidence{}, nil
	case "I001/120":
		return &v12.MeasuredRadialDopplerSpeed{}, nil
	case "I001/130":
		return &v12.RadarPlotCharacteristics{}, nil
	case "I001/131":
		return &v12.ReceivedPower{}, nil
	case "I001/141":
		return &v12.TruncatedTimeOfDay{}, nil
	default:
		return nil, fmt.Errorf("%w: %s", asterix.ErrUnknownDataItem, id)
	}
}

// Validate implements validations for Cat001
func (u *UAP12) Validate(items map[string]asterix.DataItem) error {
	// First do base validation (mandatory fields)
	if err := u.BaseUAP.Validate(items); err != nil {
		return err
	}

	// No additional validations needed for Cat001
	return nil
}

// cat001Fields defines the complete UAP for Category 001 v1.2
// Based on EUROCONTROL specification Part 2a Edition 1.4
var cat001Fields = []asterix.DataField{
	{
		FRN:         1,
		DataItem:    "I001/010",
		Description: "Data Source Identifier",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         2,
		DataItem:    "I001/020",
		Description: "Target Report Descriptor",
		Type:        asterix.Extended,
		Mandatory:   false,
	},
	{
		FRN:         3,
		DataItem:    "I001/040",
		Description: "Measured Position in Polar Coordinates",
		Type:        asterix.Fixed,
		Length:      4,
		Mandatory:   false,
	},
	{
		FRN:         4,
		DataItem:    "I001/070",
		Description: "Mode-3/A Code in Octal Representation",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         5,
		DataItem:    "I001/090",
		Description: "Mode-C Code in Binary Representation",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         6,
		DataItem:    "I001/130",
		Description: "Radar Plot Characteristics",
		Type:        asterix.Extended,
		Mandatory:   false,
	},
	{
		FRN:         7,
		DataItem:    "I001/141",
		Description: "Truncated Time of Day",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         8,
		DataItem:    "I001/050",
		Description: "Mode-2 Code in Octal Representation",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         9,
		DataItem:    "I001/120",
		Description: "Measured Radial Doppler Speed",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         10,
		DataItem:    "I001/131",
		Description: "Received Power",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         11,
		DataItem:    "I001/080",
		Description: "Mode-3/A Code Confidence Indicator",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         12,
		DataItem:    "I001/100",
		Description: "Mode-C Code and Code Confidence Indicator",
		Type:        asterix.Fixed,
		Length:      4,
		Mandatory:   false,
	},
	{
		FRN:         13,
		DataItem:    "I001/060",
		Description: "Mode-2 Code Confidence Indicator",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         14,
		DataItem:    "I001/030",
		Description: "Warning/Error Conditions",
		Type:        asterix.Extended,
		Mandatory:   false,
	},
}
