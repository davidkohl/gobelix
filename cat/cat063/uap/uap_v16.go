// cat/cat063/uap/uap_v16.go
package uap

import (
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
	v16 "github.com/davidkohl/gobelix/cat/cat063/dataitems/v16"
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

// UAP063 implements the User Application Profile for ASTERIX Category 063
type UAP063 struct {
	*asterix.BaseUAP
}

// NewUAP063 creates a new instance of the Category 063 UAP
func NewUAP063() (*UAP063, error) {
	base, err := asterix.NewBaseUAP(asterix.Cat063, "1.6", cat063Fields)
	if err != nil {
		return nil, err
	}

	return &UAP063{
		BaseUAP: base,
	}, nil
}

// CreateDataItem creates a new instance of a Cat063 data item
func (u *UAP063) CreateDataItem(id string) (asterix.DataItem, error) {
	switch id {
	case "I063/010":
		return &common.DataSourceIdentifier{}, nil
	case "I063/015":
		return &common.ServiceIdentification{}, nil
	case "I063/030":
		return &v16.TimeOfMessage{}, nil
	case "I063/050":
		return &v16.SensorIdentifier{}, nil
	case "I063/060":
		return &v16.SensorConfigurationAndStatus{}, nil
	case "I063/070":
		return &v16.TimeStampingBias{}, nil
	case "I063/080":
		return &v16.SSRModeSRangeGainAndBias{}, nil
	case "I063/081":
		return &v16.SSRModeSAzimuthBias{}, nil
	case "I063/090":
		return &v16.PSRRangeGainAndBias{}, nil
	case "I063/091":
		return &v16.PSRAzimuthBias{}, nil
	case "I063/092":
		return &v16.PSRElevationBias{}, nil
	case "RE063":
		return &v16.ReservedExpansion{}, nil
	case "SP063":
		return &v16.SpecialPurpose{}, nil
	default:
		return nil, fmt.Errorf("%w: %s", asterix.ErrUnknownDataItem, id)
	}
}

// Validate implements critical validations for Cat063
func (u *UAP063) Validate(items map[string]asterix.DataItem) error {
	// First do base validation (mandatory fields)
	if err := u.BaseUAP.Validate(items); err != nil {
		return err
	}

	// Check for the mandatory items according to the specification
	_, dataSourceExists := items["I063/010"]
	_, timeOfMessageExists := items["I063/030"]
	_, sensorIdentifierExists := items["I063/050"]

	if !dataSourceExists || !timeOfMessageExists || !sensorIdentifierExists {
		return fmt.Errorf("%w: missing mandatory field(s)", asterix.ErrMandatoryField)
	}

	return nil
}

// cat063Fields defines the complete UAP for Category 063
var cat063Fields = []asterix.DataField{
	{
		FRN:         1,
		DataItem:    "I063/010",
		Description: "Data Source Identifier",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   true,
	},
	{
		FRN:         2,
		DataItem:    "I063/015",
		Description: "Service Identification",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         3,
		DataItem:    "I063/030",
		Description: "Time of Message",
		Type:        asterix.Fixed,
		Length:      3,
		Mandatory:   true,
	},
	{
		FRN:         4,
		DataItem:    "I063/050",
		Description: "Sensor Identifier",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   true,
	},
	{
		FRN:         5,
		DataItem:    "I063/060",
		Description: "Sensor Configuration and Status",
		Type:        asterix.Extended,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         6,
		DataItem:    "I063/070",
		Description: "Time Stamping Bias",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         7,
		DataItem:    "I063/080",
		Description: "SSR/Mode S Range Gain and Bias",
		Type:        asterix.Fixed,
		Length:      4,
		Mandatory:   false,
	},
	{
		FRN:         8,
		DataItem:    "I063/081",
		Description: "SSR/Mode S Azimuth Bias",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         9,
		DataItem:    "I063/090",
		Description: "PSR Range Gain and Bias",
		Type:        asterix.Fixed,
		Length:      4,
		Mandatory:   false,
	},
	{
		FRN:         10,
		DataItem:    "I063/091",
		Description: "PSR Azimuth Bias",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         11,
		DataItem:    "I063/092",
		Description: "PSR Elevation Bias",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         12,
		DataItem:    "",
		Description: "Spare",
		Type:        asterix.Fixed,
		Length:      0,
		Mandatory:   false,
	},

	{
		FRN:         13,
		DataItem:    "RE063",
		Description: "Reserved Expansion Field",
		Type:        asterix.Repetitive,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         14,
		DataItem:    "SP063",
		Description: "Special Purpose Field",
		Type:        asterix.Repetitive,
		Length:      1,
		Mandatory:   false,
	},
}
