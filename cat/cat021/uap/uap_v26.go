// dataitems/cat021/uap.go
package uap

import (
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
	v26 "github.com/davidkohl/gobelix/cat/cat021/dataitems/v26"
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

// UAP021 implements the User Application Profile for ASTERIX Category 021
type UAP26 struct {
	*asterix.BaseUAP
}

// NewUAP021 creates a new instance of the Category 021 UAP
func NewUAP26() (*UAP26, error) {
	base, err := asterix.NewBaseUAP(asterix.Cat021, "2.6", cat021Fields)
	if err != nil {
		return nil, err
	}

	return &UAP26{
		BaseUAP: base,
	}, nil
}

// CreateDataItem creates a new instance of a Cat021 data item
// This is performance-critical - keep it simple and fast
func (u *UAP26) CreateDataItem(id string) (asterix.DataItem, error) {
	switch id {
	case "I021/008":
		return &v26.AircraftOperationalStatus{}, nil
	case "I021/010":
		return &common.DataSourceIdentifier{}, nil
	case "I021/015":
		return &common.ServiceIdentification{}, nil
	case "I021/020":
		return &v26.EmitterCategory{}, nil
	case "I021/040":
		return &v26.TargetReportDescriptor{}, nil
	case "I021/080":
		return &v26.TargetAddress{}, nil
	case "I021/090":
		return &v26.QualityIndicators{}, nil
	case "I021/130":
		return &common.Position{}, nil
	case "I021/145":
		return &common.FlightLevel{}, nil
	case "I021/170":
		return &v26.TargetIdentification{}, nil
	case "I021/071":
		return &v26.TimeOfApplicabilityPosition{}, nil
	case "I021/072":
		return &v26.TimeOfApplicabilityVelocity{}, nil
	case "I021/073":
		return &v26.TimeOfMessageReceptionPosition{}, nil
	case "I021/074":
		return &v26.TimeOfMessageReceptionPositionHigh{}, nil
	case "I021/075":
		return &v26.TimeOfMessageReceptionVelocity{}, nil
	case "I021/076":
		return &v26.TimeOfMessageReceptionVelocityHigh{}, nil
	case "I021/077":
		return &v26.TimeOfReportTransmission{}, nil
	case "I021/200":
		return &v26.TargetStatus{}, nil
	case "I021/210":
		return &v26.MOPSVersion{}, nil
	case "I021/155":
		return &v26.BarometricVerticalRate{}, nil
	case "I021/150":
		return &v26.AirSpeed{}, nil
	default:
		return nil, fmt.Errorf("%w: %s", asterix.ErrUnknownDataItem, id)
	}
}

// Validate implements critical validations for Cat021
// Note: For high-frequency decode operations, consider if all validations are needed
func (u *UAP26) Validate(items map[string]asterix.DataItem) error {
	// First do base validation (mandatory fields)
	if err := u.BaseUAP.Validate(items); err != nil {
		return err
	}

	// Critical validations only
	if pos := items["I021/130"]; pos != nil {
		// Position requires quality indicators
		if _, exists := items["I021/090"]; !exists {
			return fmt.Errorf("%w: position without quality indicators", asterix.ErrInvalidMessage)
		}

		// Position requires time reference
		hasTimeRef := items["I021/071"] != nil || // Time of Applicability
			items["I021/073"] != nil // Time of Reception
		if !hasTimeRef {
			return fmt.Errorf("%w: position without time reference", asterix.ErrInvalidMessage)
		}
	}

	return nil
}

// cat021Fields defines the complete UAP for Category 021
var cat021Fields = []asterix.DataField{
	{
		FRN:         1,
		DataItem:    "I021/010",
		Description: "Data Source Identification",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   true,
	},
	{
		FRN:         2,
		DataItem:    "I021/040",
		Description: "Target Report Descriptor",
		Type:        asterix.Extended,
		Length:      1,
		Mandatory:   true,
	},
	{
		FRN:         4,
		DataItem:    "I021/015",
		Description: "Service Identification",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         5,
		DataItem:    "I021/071",
		Description: "Time of Applicability for Position",
		Type:        asterix.Fixed,
		Length:      3,
		Mandatory:   false,
	},
	{
		FRN:         6,
		DataItem:    "I021/130",
		Description: "Position in WGS-84 co-ordinates",
		Type:        asterix.Fixed,
		Length:      6,
		Mandatory:   false,
	},
	{
		FRN:         7,
		DataItem:    "I021/131",
		Description: "Position in WGS-84 co-ordinates, high resolution",
		Type:        asterix.Fixed,
		Length:      8,
		Mandatory:   false,
	},
	{
		FRN:         8,
		DataItem:    "I021/072",
		Description: "Time of Applicability for Velocity",
		Type:        asterix.Fixed,
		Length:      3,
		Mandatory:   false,
	},
	{
		FRN:         9,
		DataItem:    "I021/150",
		Description: "Air Speed",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         10,
		DataItem:    "I021/151",
		Description: "True Air Speed",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         11,
		DataItem:    "I021/080",
		Description: "Target Address",
		Type:        asterix.Fixed,
		Length:      3,
		Mandatory:   true,
	},
	{
		FRN:         12,
		DataItem:    "I021/073",
		Description: "Time of Message Reception of Position",
		Type:        asterix.Fixed,
		Length:      3,
		Mandatory:   false,
	},
	{
		FRN:         13,
		DataItem:    "I021/074",
		Description: "Time of Message Reception of Position-High Precision",
		Type:        asterix.Fixed,
		Length:      4,
		Mandatory:   false,
	},
	{
		FRN:         14,
		DataItem:    "I021/075",
		Description: "Time of Message Reception of Velocity",
		Type:        asterix.Fixed,
		Length:      3,
		Mandatory:   false,
	},
	{
		FRN:         15,
		DataItem:    "I021/076",
		Description: "Time of Message Reception of Velocity-High Precision",
		Type:        asterix.Fixed,
		Length:      4,
		Mandatory:   false,
	},
	{
		FRN:         16,
		DataItem:    "I021/140",
		Description: "Geometric Height",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         17,
		DataItem:    "I021/090",
		Description: "Quality Indicators",
		Type:        asterix.Extended,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         18,
		DataItem:    "I021/210",
		Description: "MOPS Version",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         19,
		DataItem:    "I021/070",
		Description: "Mode 3/A Code",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         20,
		DataItem:    "I021/230",
		Description: "Roll Angle",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         21,
		DataItem:    "I021/145",
		Description: "Flight Level",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         22,
		DataItem:    "I021/152",
		Description: "Magnetic Heading",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         23,
		DataItem:    "I021/200",
		Description: "Target Status",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         24,
		DataItem:    "I021/155",
		Description: "Barometric Vertical Rate",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         25,
		DataItem:    "I021/157",
		Description: "Geometric Vertical Rate",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         26,
		DataItem:    "I021/160",
		Description: "Airborne Ground Vector",
		Type:        asterix.Fixed,
		Length:      4,
		Mandatory:   false,
	},
	{
		FRN:         27,
		DataItem:    "I021/165",
		Description: "Track Angle Rate",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         28,
		DataItem:    "I021/077",
		Description: "Time of Report Transmission",
		Type:        asterix.Fixed,
		Length:      3,
		Mandatory:   false,
	},
	{
		FRN:         29,
		DataItem:    "I021/170",
		Description: "Target Identification",
		Type:        asterix.Fixed,
		Length:      6,
		Mandatory:   false,
	},
	{
		FRN:         30,
		DataItem:    "I021/020",
		Description: "Emitter Category",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         31,
		DataItem:    "I021/220",
		Description: "Met Information",
		Type:        asterix.Compound,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         32,
		DataItem:    "I021/146",
		Description: "Selected Altitude",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         33,
		DataItem:    "I021/148",
		Description: "Final State Selected Altitude",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         34,
		DataItem:    "I021/110",
		Description: "Trajectory Intent",
		Type:        asterix.Compound,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         35,
		DataItem:    "I021/016",
		Description: "Service Management",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         36,
		DataItem:    "I021/008",
		Description: "Aircraft Operational Status",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         37,
		DataItem:    "I021/271",
		Description: "Surface Capabilities and Characteristics",
		Type:        asterix.Extended,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         38,
		DataItem:    "I021/132",
		Description: "Message Amplitude",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         39,
		DataItem:    "I021/250",
		Description: "Mode S MB Data",
		Type:        asterix.Repetitive,
		Length:      8,
		Mandatory:   false,
	},
	{
		FRN:         40,
		DataItem:    "I021/260",
		Description: "ACAS Resolution Advisory Report",
		Type:        asterix.Fixed,
		Length:      7,
		Mandatory:   false,
	},
	{
		FRN:         41,
		DataItem:    "I021/400",
		Description: "Receiver ID",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         42,
		DataItem:    "I021/295",
		Description: "Data Ages",
		Type:        asterix.Compound,
		Length:      1,
		Mandatory:   false,
	},
}
