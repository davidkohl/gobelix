// dataitems/common/flightlevel.go
package common

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type FlightLevel struct {
	Value float64 // In flight levels
}

const (
	FlightLevelResolution = 0.25  // 1/4 FL
	MinFlightLevel        = -15.0 // As per ASTERIX spec
	MaxFlightLevel        = 1500.0
)

func (f *FlightLevel) Encode(buf *bytes.Buffer) (int, error) {
	if err := f.Validate(); err != nil {
		return 0, err
	}

	// Convert to 1/4 FL units and to int16
	flUnits := int16(f.Value / FlightLevelResolution)
	err := binary.Write(buf, binary.BigEndian, flUnits)
	if err != nil {
		return 0, fmt.Errorf("writing flight level: %w", err)
	}

	return 2, nil
}

func (f *FlightLevel) Decode(buf *bytes.Buffer) (int, error) {
	var flUnits int16
	err := binary.Read(buf, binary.BigEndian, &flUnits)
	if err != nil {
		return 0, fmt.Errorf("reading flight level: %w", err)
	}

	f.Value = float64(flUnits) * FlightLevelResolution
	return 2, nil
}

func (f *FlightLevel) Validate() error {
	if f.Value < MinFlightLevel || f.Value > MaxFlightLevel {
		return fmt.Errorf("flight level %f outside valid range [%f,%f]",
			f.Value, MinFlightLevel, MaxFlightLevel)
	}
	// Check if value is a multiple of 1/4
	quarterFLs := f.Value / FlightLevelResolution
	if quarterFLs != float64(int(quarterFLs)) {
		return fmt.Errorf("flight level %f not a multiple of %f",
			f.Value, FlightLevelResolution)
	}
	return nil
}

func (f *FlightLevel) String() string {
	if f.Value < 0 {
		return fmt.Sprintf("FL-%.0f", float64(-f.Value))
	}
	return fmt.Sprintf("FL%.0f (%v ft)", float64(f.Value), float64(f.Value)*100)
}
