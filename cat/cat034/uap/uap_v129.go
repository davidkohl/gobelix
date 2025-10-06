// cat/cat034/uap/uap_v129.go
package uap

import (
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
	v129 "github.com/davidkohl/gobelix/cat/cat034/dataitems/v129"
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

// UAP129 implements the User Application Profile for ASTERIX Category 034 v1.29
type UAP129 struct {
	*asterix.BaseUAP
}

// NewUAP129 creates a new instance of the Category 034 v1.29 UAP
func NewUAP129() (*UAP129, error) {
	base, err := asterix.NewBaseUAP(asterix.Cat034, "1.29", cat034Fields)
	if err != nil {
		return nil, err
	}

	return &UAP129{
		BaseUAP: base,
	}, nil
}

// CreateDataItem creates a new instance of a Cat034 data item
func (u *UAP129) CreateDataItem(id string) (asterix.DataItem, error) {
	switch id {
	case "I034/000":
		return v129.NewMessageType(), nil
	case "I034/010":
		return &common.DataSourceIdentifier{}, nil
	case "I034/020":
		return v129.NewSectorNumber(), nil
	case "I034/030":
		return &common.TimeOfDay{}, nil
	case "I034/041":
		return v129.NewAntennaRotationPeriod(), nil
	case "I034/050":
		return v129.NewSystemConfigurationStatus(), nil
	case "I034/060":
		return v129.NewSystemProcessingMode(), nil
	default:
		return nil, fmt.Errorf("%w: %s", asterix.ErrUnknownDataItem, id)
	}
}

// Validate implements validations for Cat034
func (u *UAP129) Validate(items map[string]asterix.DataItem) error {
	// First do base validation (mandatory fields)
	if err := u.BaseUAP.Validate(items); err != nil {
		return err
	}

	// No additional validations needed for Cat034
	return nil
}

// cat034Fields defines the complete UAP for Category 034 v1.29
var cat034Fields = []asterix.DataField{
	{
		FRN:         1,
		DataItem:    "I034/010",
		Description: "Data Source Identifier",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         2,
		DataItem:    "I034/000",
		Description: "Message Type",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         3,
		DataItem:    "I034/030",
		Description: "Time of Day",
		Type:        asterix.Fixed,
		Length:      3,
		Mandatory:   false,
	},
	{
		FRN:         4,
		DataItem:    "I034/020",
		Description: "Sector Number",
		Type:        asterix.Fixed,
		Length:      1,
		Mandatory:   false,
	},
	{
		FRN:         5,
		DataItem:    "I034/041",
		Description: "Antenna Rotation Period",
		Type:        asterix.Fixed,
		Length:      2,
		Mandatory:   false,
	},
	{
		FRN:         6,
		DataItem:    "I034/050",
		Description: "System Configuration and Status",
		Type:        asterix.Compound,
		Mandatory:   false,
	},
	{
		FRN:         7,
		DataItem:    "I034/060",
		Description: "System Processing Mode",
		Type:        asterix.Compound,
		Mandatory:   false,
	},
}
