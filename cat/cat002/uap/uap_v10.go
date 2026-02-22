// cat/cat002/uap/uap_v10.go
package uap

import (
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
	v10 "github.com/davidkohl/gobelix/cat/cat002/dataitems/v10"
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

// UAP10 implements the User Application Profile for ASTERIX Category 002 v1.0
type UAP10 struct {
	*asterix.BaseUAP
}

// NewUAP10 creates a new instance of the Category 002 v1.0 UAP
func NewUAP10() (*UAP10, error) {
	base, err := asterix.NewBaseUAP(asterix.Cat002, "1.0", cat002Fields)
	if err != nil {
		return nil, err
	}

	return &UAP10{
		BaseUAP: base,
	}, nil
}

// CreateDataItem creates a new instance of a Cat002 data item
func (u *UAP10) CreateDataItem(id string) (asterix.DataItem, error) {
	switch id {
	case "I002/010":
		return &common.DataSourceIdentifier{}, nil
	case "I002/000":
		return &v10.MessageType{}, nil
	case "I002/020":
		return &common.SectorNumber{}, nil
	case "I002/030":
		return &common.TimeOfDay{}, nil
	case "I002/041":
		return &v10.AntennaRotationSpeed{}, nil
	case "I002/050":
		return &v10.StationConfigurationStatus{}, nil
	case "I002/060":
		return &v10.StationProcessingMode{}, nil
	case "I002/070":
		return &v10.PlotCountValues{}, nil
	// TODO: Implement remaining data items
	// case "I002/080": Warning/Error Conditions
	// case "I002/090": Collimation Error
	// case "I002/100": Dynamic Window - Type 1
	default:
		return nil, fmt.Errorf("%w: %s", asterix.ErrUnknownDataItem, id)
	}
}

// Validate implements validations for Cat002
func (u *UAP10) Validate(items map[string]asterix.DataItem) error {
	// First do base validation (mandatory fields)
	if err := u.BaseUAP.Validate(items); err != nil {
		return err
	}

	// No additional validations needed for Cat002
	return nil
}

// cat002Fields defines the complete UAP for Category 002 v1.0
// Based on EUROCONTROL specification for Monoradar Service Messages (Edition 1.2)
var cat002Fields = []asterix.DataField{
	{
		FRN:         1,
		DataItem:    "I002/010",
		Description: "Data Source Identifier",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         2,
		DataItem:    "I002/000",
		Description: "Message Type",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         3,
		DataItem:    "I002/020",
		Description: "Sector Number",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         4,
		DataItem:    "I002/030",
		Description: "Time of Day",
		Type:        asterix.Fixed,
		Length:      3,
		Mandatory:   false,
	},
	{
		FRN:         5,
		DataItem:    "I002/041",
		Description: "Antenna Rotation Speed",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         6,
		DataItem:    "I002/050",
		Description: "Station Configuration Status",
		Type:        asterix.Extended,
		Mandatory:   false,
	},
	{
		FRN:         7,
		DataItem:    "I002/060",
		Description: "Station Processing Mode",
		Type:        asterix.Extended,
		Mandatory:   false,
	},
	// FX - Field Extension Indicator
	{
		FRN:         8,
		DataItem:    "I002/070",
		Description: "Plot Count Values",
		Type:        asterix.Repetitive,
		Mandatory:   false,
	},
	{
		FRN:         9,
		DataItem:    "I002/100",
		Description: "Dynamic Window - Type 1",
		Type:        asterix.Fixed,
		Length:      8,
		Mandatory:   false,
	},
	{
		FRN:         10,
		DataItem:    "I002/090",
		Description: "Collimation Error",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         11,
		DataItem:    "I002/080",
		Description: "Warning/Error Conditions",
		Type:        asterix.Extended,
		Mandatory:   false,
	},
}
