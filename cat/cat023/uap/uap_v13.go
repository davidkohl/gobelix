// cat/cat023/uap/uap_v13.go
package uap

import (
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
	v13 "github.com/davidkohl/gobelix/cat/cat023/dataitems/v13"
)

// UAP13 implements the User Application Profile for ASTERIX Category 023 v1.3
type UAP13 struct {
	*asterix.BaseUAP
}

// NewUAP13 creates a new instance of the Category 023 v1.3 UAP
func NewUAP13() (*UAP13, error) {
	base, err := asterix.NewBaseUAP(asterix.Cat023, "1.3", cat023Fields)
	if err != nil {
		return nil, err
	}

	return &UAP13{
		BaseUAP: base,
	}, nil
}

// CreateDataItem creates a new instance of a Cat023 data item
func (u *UAP13) CreateDataItem(id string) (asterix.DataItem, error) {
	switch id {
	case "I023/000":
		return &v13.ReportType{}, nil
	case "I023/010":
		return &v13.DataSourceIdentifier{}, nil
	case "I023/015":
		return &v13.ServiceTypeIdentification{}, nil
	case "I023/070":
		return &v13.TimeOfDay{}, nil
	case "I023/100":
		return &v13.GroundStationStatus{}, nil
	case "I023/101":
		return &v13.ServiceConfiguration{}, nil
	case "I023/110":
		return &v13.ServiceStatus{}, nil
	case "I023/120":
		return &v13.ServiceStatistics{}, nil
	case "I023/200":
		return &v13.OperationalRange{}, nil
	case "RE023":
		return &v13.ReservedExpansion{}, nil
	case "SP023":
		return &v13.SpecialPurpose{}, nil
	default:
		return nil, fmt.Errorf("%w: %s", asterix.ErrUnknownDataItem, id)
	}
}

// Validate implements validations for Cat023
func (u *UAP13) Validate(items map[string]asterix.DataItem) error {
	// First do base validation (mandatory fields)
	if err := u.BaseUAP.Validate(items); err != nil {
		return err
	}

	// Additional validation: if report type is Ground Station Status (1),
	// I023/100 should be present
	if rtItem, exists := items["I023/000"]; exists {
		if rt, ok := rtItem.(*v13.ReportType); ok {
			if rt.ReportType == 1 {
				if _, exists := items["I023/100"]; !exists {
					return fmt.Errorf("%w: I023/100 required for Ground Station Status reports", asterix.ErrMandatoryField)
				}
			}
		}
	}

	return nil
}

// cat023Fields defines the complete UAP for Category 023 v1.3
var cat023Fields = []asterix.DataField{
	{
		FRN:         1,
		DataItem:    "I023/010",
		Description: "Data Source Identifier",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         2,
		DataItem:    "I023/000",
		Description: "Report Type",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         3,
		DataItem:    "I023/015",
		Description: "Service Type and Identification",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         4,
		DataItem:    "I023/070",
		Description: "Time of Day",
		Type:        asterix.Fixed,
		Length:      3,
		Mandatory:   false,
	},
	{
		FRN:         5,
		DataItem:    "I023/100",
		Description: "Ground Station Status",
		Type:        asterix.Extended,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         6,
		DataItem:    "I023/101",
		Description: "Service Configuration",
		Type:        asterix.Extended,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         7,
		DataItem:    "I023/200",
		Description: "Operational Range",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	// FRN 8-13 spare
	{
		FRN:         14,
		DataItem:    "I023/110",
		Description: "Service Status",
		Type:        asterix.Compound,
		Mandatory:   false,
	},
	{
		FRN:         15,
		DataItem:    "I023/120",
		Description: "Service Statistics",
		Type:        asterix.Compound,
		Mandatory:   false,
	},
	// FRN 16-20 spare
	{
		FRN:         21,
		DataItem:    "RE023",
		Description: "Reserved Expansion Field",
		Type:        asterix.Explicit,
		Mandatory:   false,
	},
	{
		FRN:         22,
		DataItem:    "SP023",
		Description: "Special Purpose Field",
		Type:        asterix.Explicit,
		Mandatory:   false,
	},
}
