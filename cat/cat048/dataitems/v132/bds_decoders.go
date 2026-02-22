// dataitems/cat048/bds_decoders.go
package v132

import (
	"fmt"
	"math"
)

// BDS40Data represents decoded BDS 4,0 - Selected vertical intention
type BDS40Data struct {
	MCPFCUSelectedAltitude    *int     // feet, bits 1-12
	FMSSelectedAltitude       *int     // feet, bits 14-25
	BarometricPressureSetting *float64 // mb, bits 27-38
}

// BDS50Data represents decoded BDS 5,0 - Track and turn report
type BDS50Data struct {
	RollAngle      *float64 // degrees, bits 1-10
	TrueTrackAngle *float64 // degrees, bits 12-21
	GroundSpeed    *float64 // knots, bits 23-32
	TrackAngleRate *float64 // degrees/second, bits 34-42
	TrueAirspeed   *float64 // knots, bits 44-53
}

// BDS60Data represents decoded BDS 6,0 - Heading and speed report
type BDS60Data struct {
	MagneticHeading          *float64 // degrees, bits 1-11
	IndicatedAirspeed        *float64 // knots, bits 13-22
	MachNumber               *float64 // Mach, bits 24-33
	BarometricAltitudeRate   *int     // feet/minute, bits 35-44
	InertialVerticalVelocity *int     // feet/minute, bits 46-55
}

// DecodeBDS40 decodes BDS 4,0 register (Selected vertical intention)
func DecodeBDS40(data []byte) (*BDS40Data, error) {
	if len(data) != 7 {
		return nil, fmt.Errorf("BDS 4,0 data must be 7 bytes")
	}

	result := &BDS40Data{}

	// MCP/FCU Selected Altitude (bits 1-12)
	if (data[0] & 0x80) != 0 { // Status bit 1
		alt := int(data[0]&0x7F)<<5 | int(data[1]&0xF8)>>3
		// 12-bit signed value, LSB = 16 ft
		if alt&0x800 != 0 { // Sign extend
			alt = alt - 0x1000
		}
		altitude := alt * 16
		// Validate: reasonable altitude range -1000 to 65000 ft
		if altitude >= -1000 && altitude <= 65000 {
			result.MCPFCUSelectedAltitude = &altitude
		}
	}

	// FMS Selected Altitude (bits 14-25)
	if (data[1] & 0x04) != 0 { // Status bit 14
		alt := int(data[1]&0x03)<<10 | int(data[2])<<2 | int(data[3]&0xC0)>>6
		// 12-bit signed value, LSB = 16 ft
		if alt&0x800 != 0 { // Sign extend
			alt = alt - 0x1000
		}
		altitude := alt * 16
		// Validate: reasonable altitude range -1000 to 65000 ft
		if altitude >= -1000 && altitude <= 65000 {
			result.FMSSelectedAltitude = &altitude
		}
	}

	// Barometric Pressure Setting (bits 27-38)
	if (data[3] & 0x20) != 0 { // Status bit 27
		press := int(data[3]&0x1F)<<7 | int(data[4]&0xFE)>>1
		// 12-bit value, LSB = 0.1 mb, offset = 800 mb
		pressure := float64(press)*0.1 + 800.0
		// Validate: reasonable pressure range 800-1100 mb
		if pressure >= 800.0 && pressure <= 1100.0 {
			result.BarometricPressureSetting = &pressure
		}
	}

	return result, nil
}

// DecodeBDS50 decodes BDS 5,0 register (Track and turn report)
func DecodeBDS50(data []byte) (*BDS50Data, error) {
	if len(data) != 7 {
		return nil, fmt.Errorf("BDS 5,0 data must be 7 bytes")
	}

	result := &BDS50Data{}

	// Roll Angle (bits 1-10)
	if (data[0] & 0x80) != 0 { // Status bit 1
		roll := int(data[0]&0x7F)<<3 | int(data[1]&0xE0)>>5
		// 10-bit signed value, LSB = 45/256 degrees
		if roll&0x200 != 0 { // Sign extend
			roll = roll - 0x400
		}
		rollAngle := float64(roll) * 45.0 / 256.0
		// Validate: reasonable roll angle -50 to +50 degrees
		if rollAngle >= -50.0 && rollAngle <= 50.0 {
			result.RollAngle = &rollAngle
		}
	}

	// True Track Angle (bits 12-21)
	if (data[1] & 0x10) != 0 { // Status bit 12
		track := int(data[1]&0x0F)<<6 | int(data[2]&0xFC)>>2
		// 10-bit signed value, LSB = 90/512 degrees
		trackAngle := float64(track) * 90.0 / 512.0
		// Validate: 0-360 degrees
		if trackAngle >= 0.0 && trackAngle <= 360.0 {
			result.TrueTrackAngle = &trackAngle
		}
	}

	// Ground Speed (bits 23-32)
	if (data[2] & 0x02) != 0 { // Status bit 23
		gs := int(data[2]&0x01)<<9 | int(data[3])<<1 | int(data[4]&0x80)>>7
		// 10-bit value, LSB = 2 knots
		groundSpeed := float64(gs) * 2.0
		// Validate: reasonable ground speed 0-750 knots
		if groundSpeed >= 0.0 && groundSpeed <= 750.0 {
			result.GroundSpeed = &groundSpeed
		}
	}

	// Track Angle Rate (bits 34-42)
	if (data[4] & 0x40) != 0 { // Status bit 34
		tar := int(data[4]&0x3F)<<3 | int(data[5]&0xE0)>>5
		// 9-bit signed value, LSB = 8/256 degrees/second
		if tar&0x100 != 0 { // Sign extend
			tar = tar - 0x200
		}
		trackAngleRate := float64(tar) * 8.0 / 256.0
		// Validate: reasonable turn rate -16 to +16 degrees/second
		if trackAngleRate >= -16.0 && trackAngleRate <= 16.0 {
			result.TrackAngleRate = &trackAngleRate
		}
	}

	// True Airspeed (bits 44-53)
	if (data[5] & 0x10) != 0 { // Status bit 44
		tas := int(data[5]&0x0F)<<6 | int(data[6]&0xFC)>>2
		// 10-bit value, LSB = 2 knots
		trueAirspeed := float64(tas) * 2.0
		// Validate: reasonable TAS 0-750 knots
		if trueAirspeed >= 0.0 && trueAirspeed <= 750.0 {
			result.TrueAirspeed = &trueAirspeed
		}
	}

	return result, nil
}

// DecodeBDS60 decodes BDS 6,0 register (Heading and speed report)
func DecodeBDS60(data []byte) (*BDS60Data, error) {
	if len(data) != 7 {
		return nil, fmt.Errorf("BDS 6,0 data must be 7 bytes")
	}

	result := &BDS60Data{}

	// Magnetic Heading (bits 1-11)
	if (data[0] & 0x80) != 0 { // Status bit 1
		hdg := int(data[0]&0x7F)<<4 | int(data[1]&0xF0)>>4
		// 11-bit signed value, LSB = 90/512 degrees
		heading := float64(hdg) * 90.0 / 512.0
		// Validate: 0-360 degrees
		if heading >= 0.0 && heading <= 360.0 {
			result.MagneticHeading = &heading
		}
	}

	// Indicated Airspeed (bits 13-22)
	if (data[1] & 0x08) != 0 { // Status bit 13
		ias := int(data[1]&0x07)<<7 | int(data[2]&0xFE)>>1
		// 10-bit value, LSB = 1 knot
		airspeed := float64(ias)
		// Validate: reasonable IAS 0-700 knots
		if airspeed >= 0.0 && airspeed <= 700.0 {
			result.IndicatedAirspeed = &airspeed
		}
	}

	// Mach Number (bits 24-33)
	if (data[2] & 0x01) != 0 { // Status bit 24
		mach := int(data[3])<<2 | int(data[4]&0xC0)>>6
		// 10-bit value, LSB = 0.008 Mach
		machNumber := float64(mach) * 0.008
		// Validate: reasonable Mach 0-1.0
		if machNumber >= 0.0 && machNumber <= 1.0 {
			result.MachNumber = &machNumber
		}
	}

	// Barometric Altitude Rate (bits 35-44)
	if (data[4] & 0x20) != 0 { // Status bit 35
		baro := int(data[4]&0x1F)<<5 | int(data[5]&0xF8)>>3
		// 10-bit signed value, LSB = 32 ft/min
		if baro&0x200 != 0 { // Sign extend
			baro = baro - 0x400
		}
		baroRate := baro * 32
		// Validate: reasonable climb/descent rate -10000 to +10000 ft/min
		if baroRate >= -10000 && baroRate <= 10000 {
			result.BarometricAltitudeRate = &baroRate
		}
	}

	// Inertial Vertical Velocity (bits 46-55)
	if (data[5] & 0x04) != 0 { // Status bit 46
		ivv := int(data[5]&0x03)<<8 | int(data[6])
		// 10-bit signed value, LSB = 32 ft/min
		if ivv&0x200 != 0 { // Sign extend
			ivv = ivv - 0x400
		}
		inertialVV := ivv * 32
		// Validate: reasonable vertical velocity -10000 to +10000 ft/min
		if inertialVV >= -10000 && inertialVV <= 10000 {
			result.InertialVerticalVelocity = &inertialVV
		}
	}

	return result, nil
}

// String methods for formatted output

func (b *BDS40Data) String() string {
	result := "BDS 4,0 (Selected Vertical Intention):"
	hasData := false

	if b.MCPFCUSelectedAltitude != nil {
		result += fmt.Sprintf("\n    MCP/FCU Selected Altitude: %d ft", *b.MCPFCUSelectedAltitude)
		hasData = true
	}
	if b.FMSSelectedAltitude != nil {
		result += fmt.Sprintf("\n    FMS Selected Altitude: %d ft", *b.FMSSelectedAltitude)
		hasData = true
	}
	if b.BarometricPressureSetting != nil {
		result += fmt.Sprintf("\n    Barometric Pressure: %.1f mb", *b.BarometricPressureSetting)
		hasData = true
	}

	if !hasData {
		result += " (no valid data)"
	}

	return result
}

func (b *BDS50Data) String() string {
	result := "BDS 5,0 (Track and Turn Report):"
	hasData := false

	if b.RollAngle != nil {
		result += fmt.Sprintf("\n    Roll Angle: %.2f°", *b.RollAngle)
		hasData = true
	}
	if b.TrueTrackAngle != nil {
		result += fmt.Sprintf("\n    True Track Angle: %.2f°", *b.TrueTrackAngle)
		hasData = true
	}
	if b.GroundSpeed != nil {
		result += fmt.Sprintf("\n    Ground Speed: %.0f kt", *b.GroundSpeed)
		hasData = true
	}
	if b.TrackAngleRate != nil {
		result += fmt.Sprintf("\n    Track Angle Rate: %.3f°/s", *b.TrackAngleRate)
		hasData = true
	}
	if b.TrueAirspeed != nil {
		result += fmt.Sprintf("\n    True Airspeed: %.0f kt", *b.TrueAirspeed)
		hasData = true
	}

	if !hasData {
		result += " (no valid data)"
	}

	return result
}

func (b *BDS60Data) String() string {
	result := "BDS 6,0 (Heading and Speed Report):"
	hasData := false

	if b.MagneticHeading != nil {
		result += fmt.Sprintf("\n    Magnetic Heading: %.2f°", *b.MagneticHeading)
		hasData = true
	}
	if b.IndicatedAirspeed != nil {
		result += fmt.Sprintf("\n    Indicated Airspeed: %.0f kt", *b.IndicatedAirspeed)
		hasData = true
	}
	if b.MachNumber != nil {
		result += fmt.Sprintf("\n    Mach Number: %.3f", *b.MachNumber)
		hasData = true
	}
	if b.BarometricAltitudeRate != nil {
		sign := ""
		if *b.BarometricAltitudeRate > 0 {
			sign = "+"
		}
		result += fmt.Sprintf("\n    Barometric Altitude Rate: %s%d ft/min", sign, *b.BarometricAltitudeRate)
		hasData = true
	}
	if b.InertialVerticalVelocity != nil {
		sign := ""
		if *b.InertialVerticalVelocity > 0 {
			sign = "+"
		}
		result += fmt.Sprintf("\n    Inertial Vertical Velocity: %s%d ft/min", sign, *b.InertialVerticalVelocity)
		hasData = true
	}

	if !hasData {
		result += " (no valid data)"
	}

	return result
}

// Helper to check if value is reasonable (for validation)
func isReasonable(value float64, min, max float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= min && value <= max
}
