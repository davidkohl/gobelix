// dataitems/cat021/uap.go
package cat021

import (
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/dataitems/common"
)

type UAP021 struct {
	fields []asterix.DataField
}

func NewUAP021() *UAP021 {
	return &UAP021{
		fields: cat021Fields,
	}
}

func (u *UAP021) Category() asterix.Category {
	return asterix.Cat021
}

func (u *UAP021) Version() string {
	return "2.6"
}

func (u *UAP021) Fields() []asterix.DataField {
	return u.fields
}

func (u *UAP021) FRNByID(id string) uint8 {
	for _, field := range u.fields {
		if field.DataItem == id {
			return field.FRN
		}
	}
	return 0
}

func (u *UAP021) DataFieldByID(id string) *asterix.DataField {
	for _, field := range u.fields {
		if field.DataItem == id {
			return &field
		}
	}
	return nil
}

func (u *UAP021) CreateDataItem(id string) (asterix.DataItem, error) {
	switch id {
	case "I021/008":
		return &AircraftOperationalStatus{}, nil
	case "I021/010":
		return &common.DataSourceIdentifier{}, nil
	case "I021/015":
		return &ServiceID{}, nil
	case "I021/040":
		return &TargetReportDescriptor{}, nil
	case "I021/080":
		return &TargetAddress{}, nil
	case "I021/090":
		return &QualityIndicators{}, nil
	case "I021/130":
		return &common.Position{}, nil
	case "I021/145":
		return &common.FlightLevel{}, nil
	case "I021/170":
		return &TargetIdentification{}, nil
	case "I021/071":
		return &TimeOfApplicabilityPosition{}, nil
	case "I021/072":
		return &TimeOfApplicabilityVelocity{}, nil
	case "I021/073":
		return &TimeOfMessageReceptionPosition{}, nil
	case "I021/074":
		return &TimeOfMessageReceptionPositionHigh{}, nil
	case "I021/075":
		return &TimeOfMessageReceptionVelocity{}, nil
	case "I021/076":
		return &TimeOfMessageReceptionVelocityHigh{}, nil
	case "I021/077":
		return &TimeOfReportTransmission{}, nil
	case "I021/200":
		return &TargetStatus{}, nil
	case "I021/210":
		return &MOPSVersion{}, nil
	case "I021/155":
		return &BarometricVerticalRate{}, nil
	case "I021/150":
		return &AirSpeed{}, nil

	default:
		return nil, fmt.Errorf("%w: %s", asterix.ErrUnknownDataItem, id)
	}
}

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
		FRN:         3,
		DataItem:    "I021/161",
		Description: "Track Number",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
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
		Mandatory:   false,
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
		Length:      1,
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
	{
		FRN:         43,
		DataItem:    "RE",
		Description: "Reserved Expansion Field",
		Type:        asterix.Compound,
		Mandatory:   false,
	},
	{
		FRN:         44,
		DataItem:    "SP",
		Description: "Special Purpose Field",
		Type:        asterix.Compound,
		Mandatory:   false,
	},
}
