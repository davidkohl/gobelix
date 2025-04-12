// cat/cat048/uap/uap_v132.go
package uap

import (
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
	cat048 "github.com/davidkohl/gobelix/cat/cat048/dataitems/v132"
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

// UAP048 implements the User Application Profile for ASTERIX Category 048
type UAP048 struct {
	*asterix.BaseUAP
}

// NewUAP132 creates a new instance of the Category 048 UAP version 1.32
func NewUAP132() (*UAP048, error) {
	base, err := asterix.NewBaseUAP(asterix.Cat048, "1.32", cat048Fields)
	if err != nil {
		return nil, err
	}

	return &UAP048{
		BaseUAP: base,
	}, nil
}

// CreateDataItem creates a new instance of a Cat048 data item
// This is performance-critical - keep it simple and fast
func (u *UAP048) CreateDataItem(id string) (asterix.DataItem, error) {
	switch id {
	case "I048/010":
		return &common.DataSourceIdentifier{}, nil
	case "I048/140":
		return &cat048.TimeOfDay{}, nil
	case "I048/020":
		return &cat048.TargetReportDescriptor{}, nil
	case "I048/040":
		return &cat048.MeasuredPosition{}, nil
	case "I048/070":
		return &cat048.Mode3ACode{}, nil
	case "I048/090":
		return &cat048.FlightLevel{}, nil
	case "I048/130":
		return &cat048.RadarPlotCharacteristics{}, nil
	case "I048/220":
		return &cat048.AircraftAddress{}, nil
	case "I048/240":
		return &cat048.AircraftIdentification{}, nil
	case "I048/250":
		return &cat048.BDSRegisterData{}, nil
	case "I048/161":
		return &cat048.TrackNumber{}, nil
	case "I048/042":
		return &cat048.CalculatedPosition{}, nil
	case "I048/200":
		return &cat048.CalculatedTrackVelocity{}, nil
	case "I048/170":
		return &cat048.TrackStatus{}, nil
	case "I048/210":
		return &cat048.TrackQuality{}, nil
	case "I048/030":
		return &cat048.WarningErrorCondition{}, nil
	case "I048/080":
		return &cat048.Mode3ACodeConfidence{}, nil
	case "I048/100":
		return &cat048.ModeCCodeAndConfidence{}, nil
	case "I048/110":
		return &cat048.Height3D{}, nil
	case "I048/120":
		return &cat048.RadialDopplerSpeed{}, nil
	case "I048/230":
		return &cat048.CommunicationsCapability{}, nil
	case "I048/260":
		return &cat048.ACASResolutionAdvisory{}, nil
	case "I048/055":
		return &cat048.Mode1Code{}, nil
	case "I048/050":
		return &cat048.Mode2Code{}, nil
	case "I048/065":
		return &cat048.Mode1CodeConfidence{}, nil
	case "I048/060":
		return &cat048.Mode2CodeConfidence{}, nil
	case "SP048":
		return &cat048.SpecialPurpose{}, nil
	case "RE048":
		return &cat048.ReservedExpansion{}, nil
	default:
		return nil, fmt.Errorf("%w: %s", asterix.ErrUnknownDataItem, id)
	}
}

// Validate implements critical validations for Cat048
func (u *UAP048) Validate(items map[string]asterix.DataItem) error {
	// First do base validation (mandatory fields)
	if err := u.BaseUAP.Validate(items); err != nil {
		return err
	}

	// Additional validations specific to CAT048
	// For data item I048/040 Measured Position, according to the specification
	// it shall be sent when there is a detection. We check this by looking at the TYP field
	// in I048/020, if it's not zero (no detection) then I048/040 must be present.
	if trdItem, exists := items["I048/020"]; exists {
		if trd, ok := trdItem.(*cat048.TargetReportDescriptor); ok {
			// Check if TYP is not 0 (meaning there is a detection)
			if trd.TYP != 0 {
				if _, exists := items["I048/040"]; !exists {
					return fmt.Errorf("%w: I048/040 required when detection exists", asterix.ErrMandatoryField)
				}
			}
		}
	}

	return nil
}

// cat048Fields defines the UAP for Category 048 version 1.32
var cat048Fields = []asterix.DataField{
	{
		FRN:         1,
		DataItem:    "I048/010",
		Description: "Data Source Identifier",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   true,
	},
	{
		FRN:         2,
		DataItem:    "I048/140",
		Description: "Time of Day",
		Type:        asterix.Fixed,
		Length:      3,
		Mandatory:   true,
	},
	{
		FRN:         3,
		DataItem:    "I048/020",
		Description: "Target Report Descriptor",
		Type:        asterix.Extended,
		Length:      1,
		Mandatory:   true,
	},
	{
		FRN:         4,
		DataItem:    "I048/040",
		Description: "Measured Position in Polar Co-ordinates",
		Type:        asterix.Fixed,
		Length:      4,
		Mandatory:   false,
	},
	{
		FRN:         5,
		DataItem:    "I048/070",
		Description: "Mode-3/A Code in Octal Representation",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         6,
		DataItem:    "I048/090",
		Description: "Flight Level in Binary Representation",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         7,
		DataItem:    "I048/130",
		Description: "Radar Plot Characteristics",
		Type:        asterix.Compound,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         8,
		DataItem:    "I048/220",
		Description: "Aircraft Address",
		Type:        asterix.Fixed,
		Length:      3,
		Mandatory:   false,
	},
	{
		FRN:         9,
		DataItem:    "I048/240",
		Description: "Aircraft Identification",
		Type:        asterix.Fixed,
		Length:      6,
		Mandatory:   false,
	},
	{
		FRN:         10,
		DataItem:    "I048/250",
		Description: "BDS Register Data",
		Type:        asterix.Repetitive,
		Length:      8,
		Mandatory:   false,
	},
	{
		FRN:         11,
		DataItem:    "I048/161",
		Description: "Track Number",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         12,
		DataItem:    "I048/042",
		Description: "Calculated Position in Cartesian Coordinates",
		Type:        asterix.Fixed,
		Length:      4,
		Mandatory:   false,
	},
	{
		FRN:         13,
		DataItem:    "I048/200",
		Description: "Calculated Track Velocity in Polar Representation",
		Type:        asterix.Fixed,
		Length:      4,
		Mandatory:   false,
	},
	{
		FRN:         14,
		DataItem:    "I048/170",
		Description: "Track Status",
		Type:        asterix.Extended,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         15,
		DataItem:    "I048/210",
		Description: "Track Quality",
		Type:        asterix.Fixed,
		Length:      4,
		Mandatory:   false,
	},
	{
		FRN:         16,
		DataItem:    "I048/030",
		Description: "Warning/Error Conditions and Target Classification",
		Type:        asterix.Extended,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         17,
		DataItem:    "I048/080",
		Description: "Mode-3/A Code Confidence Indicator",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         18,
		DataItem:    "I048/100",
		Description: "Mode-C Code and Code Confidence Indicator",
		Type:        asterix.Fixed,
		Length:      4,
		Mandatory:   false,
	},
	{
		FRN:         19,
		DataItem:    "I048/110",
		Description: "Height Measured by a 3D Radar",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         20,
		DataItem:    "I048/120",
		Description: "Radial Doppler Speed",
		Type:        asterix.Compound,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         21,
		DataItem:    "I048/230",
		Description: "Communications/ACAS Capability and Flight Status",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         22,
		DataItem:    "I048/260",
		Description: "ACAS Resolution Advisory Report",
		Type:        asterix.Fixed,
		Length:      7,
		Mandatory:   false,
	},
	{
		FRN:         23,
		DataItem:    "I048/055",
		Description: "Mode-1 Code in Octal Representation",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         24,
		DataItem:    "I048/050",
		Description: "Mode-2 Code in Octal Representation",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         25,
		DataItem:    "I048/065",
		Description: "Mode-1 Code Confidence Indicator",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         26,
		DataItem:    "I048/060",
		Description: "Mode-2 Code Confidence Indicator",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         27,
		DataItem:    "SP048",
		Description: "Special Purpose Field",
		Type:        asterix.Repetitive,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         28,
		DataItem:    "RE048",
		Description: "Reserved Expansion Field",
		Type:        asterix.Repetitive,
		Length:      1,
		Mandatory:   false,
	},
}
