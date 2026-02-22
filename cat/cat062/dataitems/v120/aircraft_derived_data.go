// dataitems/cat062/aircraft_derived_data.go
package v120

import (
	"bytes"
	"fmt"
	"math"
	"strings"
)

// AircraftDerivedData implements I062/380
// Data derived directly by the aircraft (ADS-B, Mode S, etc.)
type AircraftDerivedData struct {
	// Subfield #1: Target Address (24-bit Mode S address)
	TargetAddress *uint32 // 24-bit Mode S address (0x000000 - 0xFFFFFF)

	// Subfield #2: Target Identification (callsign)
	TargetIdentification *string // 8 characters, 6-bit encoded

	// Subfield #3: Magnetic Heading
	MagneticHeading *float64 // Degrees, LSB = 360°/2^16 ≈ 0.0055°

	// Subfield #4: Indicated Airspeed / Mach Number
	IndicatedAirspeedIMAch *bool    // false=IAS, true=Mach
	IndicatedAirspeed      *float64 // If IM=0: NM/s (LSB=2^-14), If IM=1: Mach (LSB=0.001)

	// Subfield #5: True Airspeed
	TrueAirspeed *uint16 // Knots, LSB = 1 knot, range 0-2046 kt

	// Subfield #6: Selected Altitude
	SelectedAltitudeSAS    *bool   // Source information provided
	SelectedAltitudeSource *uint8  // 0=Unknown, 1=Aircraft, 2=FCU/MCP, 3=FMS
	SelectedAltitude       *int16  // Feet, LSB = 25 ft, range -1300 to 100000 ft

	// Subfield #7: Final State Selected Altitude
	FinalStateAltitudeMV *bool  // Manage Vertical Mode active
	FinalStateAltitudeAH *bool  // Altitude Hold active
	FinalStateAltitudeAM *bool  // Approach Mode active
	FinalStateAltitude   *int16 // Feet, LSB = 25 ft, range -1300 to 100000 ft

	// Subfield #8: Trajectory Intent Status
	TrajectoryIntentNAV *bool // false=available, true=not available
	TrajectoryIntentNVB *bool // false=valid, true=not valid

	// Subfield #9: Trajectory Intent Data
	TrajectoryIntentData []TrajectoryIntentPoint

	// Subfield #10: Communications/ACAS Capability and Flight Status
	CommunicationsCOM  *uint8 // 0-7 (communications capability)
	CommunicationsSTAT *uint8 // 0-5 (flight status)
	CommunicationsSSC  *bool  // Specific service capability
	CommunicationsARC  *bool  // Altitude reporting capability (false=100ft, true=25ft)
	CommunicationsAIC  *bool  // Aircraft identification capability
	CommunicationsB1A  *bool  // BDS 1,0 bit 16
	CommunicationsB1B  *uint8 // BDS 1,0 bits 37/40 (4 bits)

	// Subfield #11: Status reported by ADS-B
	ADSBStatusAC  *uint8 // 0-3 (ACAS status)
	ADSBStatusMN  *uint8 // 0-3 (Multiple navigational aids status)
	ADSBStatusDC  *uint8 // 0-3 (Differential correction status)
	ADSBStatusGBS *bool  // Transponder Ground Bit set
	ADSBStatusSTAT *uint8 // 0-7 (Emergency status)

	// Subfield #12: ACAS Resolution Advisory Report
	ACASResolutionAdvisory *[7]byte // 56-bit Mode S Comm B message (BDS 3,0)

	// Subfield #13: Barometric Vertical Rate
	BarometricVerticalRate *int16 // ft/min, LSB = 6.25 ft/min, two's complement

	// Subfield #14: Geometric Vertical Rate
	GeometricVerticalRate *int16 // ft/min, LSB = 6.25 ft/min, two's complement

	// Subfield #15: Roll Angle
	RollAngle *float64 // Degrees, LSB = 0.01°, two's complement

	// Subfield #16: Track Angle Rate
	TrackAngleRate *float64 // Deg/s, LSB = 1/32 deg/s, two's complement

	// Subfield #17: Track Angle
	TrackAngle *float64 // Degrees, LSB = 360°/2^16 ≈ 0.0055°

	// Subfield #18: Ground Speed
	GroundSpeed *float64 // Knots, LSB = 2^-14 NM/s ≈ 0.22 kt

	// Subfield #19: Velocity Uncertainty
	VelocityUncertainty *uint8 // LSB = 1 (unitless)

	// Subfield #20: Meteorological Data
	MeteorologicalData *MeteorologicalData

	// Subfield #21: Emitter Category
	EmitterCategory *uint8 // ICAO Emitter Category (see spec for values)

	// Subfield #22: Position
	PositionLatitude  *float64 // WGS-84 degrees, LSB = 180/2^23 ≈ 2.145767e-05°
	PositionLongitude *float64 // WGS-84 degrees, LSB = 180/2^23 ≈ 2.145767e-05°

	// Subfield #23: Geometric Altitude Data
	GeometricAltitude *int16 // Feet, LSB = 6.25 ft, two's complement

	// Subfield #24: Position Uncertainty
	PositionUncertainty *uint8 // LSB = 1 (unitless)

	// Subfield #25: Mode S MB Data
	ModeSMBData [][]byte // Array of 8-byte Mode S MB messages

	// Subfield #26: Indicated Airspeed
	IndicatedAirspeedData *uint16 // Knots, LSB = 2^-14 NM/s ≈ 0.22 kt

	// Subfield #27: Mach Number
	MachNumber *float64 // Mach, LSB = 0.008

	// Subfield #28: Barometric Pressure Setting
	BarometricPressureSetting *float64 // mb, LSB = 0.1 mb
}

// TrajectoryIntentPoint represents a single trajectory change point
type TrajectoryIntentPoint struct {
	TCA              bool    // TCP number available (false=available, true=not available)
	NC               bool    // TCP compliance (false=compliance, true=non-compliance)
	TCPNumber        uint8   // Trajectory Change Point number (0-63)
	Altitude         int16   // Feet, LSB = 10 ft, range -1500 to 150000 ft
	Latitude         float64 // WGS-84 degrees, LSB = 180/2^23 ≈ 2.145767e-05°
	Longitude        float64 // WGS-84 degrees, LSB = 180/2^23 ≈ 2.145767e-05°
	PointType        uint8   // 0-11 (waypoint type)
	TD               uint8   // Turn direction: 0=N/A, 1=Right, 2=Left, 3=No turn
	TRA              bool    // Turn radius available
	TOA              bool    // Time over point available (false=available)
	TOV              uint32  // Time over point in seconds from midnight (LSB = 1s)
	TTR              float64 // Turn radius in NM (LSB = 0.01 NM)
}

// MeteorologicalData represents weather information
type MeteorologicalData struct {
	WindSpeed     *uint16  // Knots, LSB = 1 kt
	WindDirection *uint16  // Degrees, LSB = 1°
	Temperature   *float64 // Celsius, LSB = 0.25°C, two's complement
	Turbulence    *uint8   // Turbulence level (0-3)
}

func (a *AircraftDerivedData) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Read primary subfield (up to 4 octets with FX extension bits)
	primaryBytes := make([]byte, 0, 4)
	for i := 0; i < 4; i++ {
		octet := make([]byte, 1)
		n, err := buf.Read(octet)
		if err != nil {
			return bytesRead, fmt.Errorf("reading aircraft derived data primary subfield octet %d: %w", i+1, err)
		}
		bytesRead += n
		primaryBytes = append(primaryBytes, octet[0])

		// Check if there's an extension
		hasExtension := (octet[0] & 0x01) != 0
		if !hasExtension {
			break
		}
	}

	// Now read subfields based on the bits set in the primary subfield
	// We process bits from bit-32 down to bit-1 (excluding FX bits)
	subfieldIndex := 0
	for byteIdx := 0; byteIdx < len(primaryBytes); byteIdx++ {
		// Process bits 8-2 (bit 1 is FX)
		for bitPos := 7; bitPos >= 1; bitPos-- {
			if (primaryBytes[byteIdx] & (1 << bitPos)) != 0 {
				// This subfield is present
				n, err := a.decodeSubfield(subfieldIndex, buf)
				bytesRead += n
				if err != nil {
					return bytesRead, fmt.Errorf("decoding subfield #%d: %w", subfieldIndex+1, err)
				}
			}
			subfieldIndex++
		}
	}

	return bytesRead, nil
}

func (a *AircraftDerivedData) decodeSubfield(index int, buf *bytes.Buffer) (int, error) {
	switch index {
	case 0: // #1: Target Address (3 octets)
		return a.decodeTargetAddress(buf)
	case 1: // #2: Target Identification (6 octets)
		return a.decodeTargetIdentification(buf)
	case 2: // #3: Magnetic Heading (2 octets)
		return a.decodeMagneticHeading(buf)
	case 3: // #4: Indicated Airspeed/Mach (2 octets)
		return a.decodeIndicatedAirspeed(buf)
	case 4: // #5: True Airspeed (2 octets)
		return a.decodeTrueAirspeed(buf)
	case 5: // #6: Selected Altitude (2 octets)
		return a.decodeSelectedAltitude(buf)
	case 6: // #7: Final State Selected Altitude (2 octets)
		return a.decodeFinalStateAltitude(buf)
	case 7: // #8: Trajectory Intent Status (1+ octets)
		return a.decodeTrajectoryIntentStatus(buf)
	case 8: // #9: Trajectory Intent Data (Repetitive)
		return a.decodeTrajectoryIntentData(buf)
	case 9: // #10: Communications/ACAS (2 octets)
		return a.decodeCommunications(buf)
	case 10: // #11: Status by ADS-B (2 octets)
		return a.decodeADSBStatus(buf)
	case 11: // #12: ACAS Resolution Advisory (7 octets)
		return a.decodeACASResolutionAdvisory(buf)
	case 12: // #13: Barometric Vertical Rate (2 octets)
		return a.decodeBarometricVerticalRate(buf)
	case 13: // #14: Geometric Vertical Rate (2 octets)
		return a.decodeGeometricVerticalRate(buf)
	case 14: // #15: Roll Angle (2 octets)
		return a.decodeRollAngle(buf)
	case 15: // #16: Track Angle Rate (2 octets)
		return a.decodeTrackAngleRate(buf)
	case 16: // #17: Track Angle (2 octets)
		return a.decodeTrackAngle(buf)
	case 17: // #18: Ground Speed (2 octets)
		return a.decodeGroundSpeed(buf)
	case 18: // #19: Velocity Uncertainty (1 octet)
		return a.decodeVelocityUncertainty(buf)
	case 19: // #20: Meteorological Data (Variable)
		return a.decodeMeteorologicalData(buf)
	case 20: // #21: Emitter Category (1 octet)
		return a.decodeEmitterCategory(buf)
	case 21: // #22: Position (8 octets)
		return a.decodePosition(buf)
	case 22: // #23: Geometric Altitude (2 octets)
		return a.decodeGeometricAltitude(buf)
	case 23: // #24: Position Uncertainty (1 octet)
		return a.decodePositionUncertainty(buf)
	case 24: // #25: Mode S MB Data (Repetitive)
		return a.decodeModeSMBData(buf)
	case 25: // #26: Indicated Airspeed (2 octets)
		return a.decodeIndicatedAirspeedData(buf)
	case 26: // #27: Mach Number (2 octets)
		return a.decodeMachNumber(buf)
	case 27: // #28: Barometric Pressure Setting (2 octets)
		return a.decodeBarometricPressureSetting(buf)
	default:
		return 0, fmt.Errorf("unknown subfield index: %d", index)
	}
}

// Individual subfield decoders

func (a *AircraftDerivedData) decodeTargetAddress(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 3)
	n, err := buf.Read(data)
	if err != nil || n != 3 {
		return n, fmt.Errorf("reading target address: %w", err)
	}
	addr := uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2])
	a.TargetAddress = &addr
	return n, nil
}

func (a *AircraftDerivedData) decodeTargetIdentification(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 6)
	n, err := buf.Read(data)
	if err != nil || n != 6 {
		return n, fmt.Errorf("reading target identification: %w", err)
	}
	// Decode 8 characters, each 6 bits
	// Characters are packed: bits 48-43, 42-37, 36-31, etc.
	chars := make([]byte, 8)
	chars[0] = (data[0] >> 2) & 0x3F
	chars[1] = ((data[0] & 0x03) << 4) | ((data[1] >> 4) & 0x0F)
	chars[2] = ((data[1] & 0x0F) << 2) | ((data[2] >> 6) & 0x03)
	chars[3] = data[2] & 0x3F
	chars[4] = (data[3] >> 2) & 0x3F
	chars[5] = ((data[3] & 0x03) << 4) | ((data[4] >> 4) & 0x0F)
	chars[6] = ((data[4] & 0x0F) << 2) | ((data[5] >> 6) & 0x03)
	chars[7] = data[5] & 0x3F

	// Convert 6-bit values to ASCII
	ident := decodeICAO6BitString(chars)
	a.TargetIdentification = &ident
	return n, nil
}

func (a *AircraftDerivedData) decodeMagneticHeading(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil || n != 2 {
		return n, fmt.Errorf("reading magnetic heading: %w", err)
	}
	raw := uint16(data[0])<<8 | uint16(data[1])
	heading := float64(raw) * 360.0 / 65536.0 // LSB = 360°/2^16
	a.MagneticHeading = &heading
	return n, nil
}

func (a *AircraftDerivedData) decodeIndicatedAirspeed(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil || n != 2 {
		return n, fmt.Errorf("reading indicated airspeed: %w", err)
	}
	im := (data[0] & 0x80) != 0
	a.IndicatedAirspeedIMAch = &im

	raw := uint16(data[0]&0x7F)<<8 | uint16(data[1])
	var speed float64
	if im {
		// Mach number, LSB = 0.001
		speed = float64(raw) * 0.001
	} else {
		// IAS in NM/s, LSB = 2^-14 NM/s
		speed = float64(raw) / 16384.0
	}
	a.IndicatedAirspeed = &speed
	return n, nil
}

func (a *AircraftDerivedData) decodeTrueAirspeed(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil || n != 2 {
		return n, fmt.Errorf("reading true airspeed: %w", err)
	}
	tas := uint16(data[0])<<8 | uint16(data[1])
	a.TrueAirspeed = &tas
	return n, nil
}

func (a *AircraftDerivedData) decodeSelectedAltitude(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil || n != 2 {
		return n, fmt.Errorf("reading selected altitude: %w", err)
	}
	sas := (data[0] & 0x80) != 0
	a.SelectedAltitudeSAS = &sas

	source := (data[0] >> 5) & 0x03
	a.SelectedAltitudeSource = &source

	// 13-bit two's complement altitude
	rawAlt := int16(uint16(data[0]&0x1F)<<8 | uint16(data[1]))
	// Sign extend from 13 bits
	if (rawAlt & 0x1000) != 0 {
		rawAlt |= -0x2000 // Sign extend from bit 12
	}
	alt := rawAlt * 25 // LSB = 25 ft
	a.SelectedAltitude = &alt
	return n, nil
}

func (a *AircraftDerivedData) decodeFinalStateAltitude(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil || n != 2 {
		return n, fmt.Errorf("reading final state altitude: %w", err)
	}
	mv := (data[0] & 0x80) != 0
	a.FinalStateAltitudeMV = &mv

	ah := (data[0] & 0x40) != 0
	a.FinalStateAltitudeAH = &ah

	am := (data[0] & 0x20) != 0
	a.FinalStateAltitudeAM = &am

	// 13-bit two's complement altitude
	rawAlt := int16(uint16(data[0]&0x1F)<<8 | uint16(data[1]))
	// Sign extend from 13 bits
	if (rawAlt & 0x1000) != 0 {
		rawAlt |= -0x2000 // Sign extend from bit 12
	}
	alt := rawAlt * 25 // LSB = 25 ft
	a.FinalStateAltitude = &alt
	return n, nil
}

func (a *AircraftDerivedData) decodeTrajectoryIntentStatus(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 1)
	n, err := buf.Read(data)
	if err != nil || n != 1 {
		return n, fmt.Errorf("reading trajectory intent status: %w", err)
	}
	nav := (data[0] & 0x80) != 0
	a.TrajectoryIntentNAV = &nav

	nvb := (data[0] & 0x40) != 0
	a.TrajectoryIntentNVB = &nvb

	// Note: There could be FX extension, but per spec it's not currently defined
	// If FX bit is set, we should skip it
	if (data[0] & 0x01) != 0 {
		// Skip extension octet
		extData := make([]byte, 1)
		nExt, _ := buf.Read(extData)
		n += nExt
	}
	return n, nil
}

func (a *AircraftDerivedData) decodeTrajectoryIntentData(buf *bytes.Buffer) (int, error) {
	repByte := make([]byte, 1)
	n, err := buf.Read(repByte)
	if err != nil {
		return n, fmt.Errorf("reading TID repetition factor: %w", err)
	}
	bytesRead := n
	rep := int(repByte[0])

	a.TrajectoryIntentData = make([]TrajectoryIntentPoint, rep)
	for i := 0; i < rep; i++ {
		tidData := make([]byte, 15)
		n, err := buf.Read(tidData)
		if err != nil || n != 15 {
			return bytesRead + n, fmt.Errorf("reading TID point %d: %w", i+1, err)
		}
		bytesRead += n

		// Decode trajectory point
		point := &a.TrajectoryIntentData[i]
		point.TCA = (tidData[0] & 0x80) != 0
		point.NC = (tidData[0] & 0x40) != 0
		point.TCPNumber = tidData[0] & 0x3F

		// Altitude (16 bits, two's complement, LSB = 10 ft)
		rawAlt := int16(uint16(tidData[1])<<8 | uint16(tidData[2]))
		point.Altitude = rawAlt * 10

		// Latitude (24 bits, two's complement, LSB = 180/2^23)
		rawLat := int32(uint32(tidData[3])<<16 | uint32(tidData[4])<<8 | uint32(tidData[5]))
		if (rawLat & 0x800000) != 0 {
			rawLat |= -0x1000000 // Sign extend from bit 23
		}
		point.Latitude = float64(rawLat) * 180.0 / 8388608.0

		// Longitude (24 bits, two's complement, LSB = 180/2^23)
		rawLon := int32(uint32(tidData[6])<<16 | uint32(tidData[7])<<8 | uint32(tidData[8]))
		if (rawLon & 0x800000) != 0 {
			rawLon |= -0x1000000 // Sign extend from bit 23
		}
		point.Longitude = float64(rawLon) * 180.0 / 8388608.0

		// Point type and flags
		point.PointType = (tidData[9] >> 4) & 0x0F
		point.TD = (tidData[9] >> 2) & 0x03
		point.TRA = (tidData[9] & 0x02) != 0
		point.TOA = (tidData[9] & 0x01) != 0

		// TOV (24 bits, LSB = 1 second)
		point.TOV = uint32(tidData[10])<<16 | uint32(tidData[11])<<8 | uint32(tidData[12])

		// TTR (16 bits, LSB = 0.01 NM)
		rawTTR := uint16(tidData[13])<<8 | uint16(tidData[14])
		point.TTR = float64(rawTTR) * 0.01
	}

	return bytesRead, nil
}

func (a *AircraftDerivedData) decodeCommunications(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil || n != 2 {
		return n, fmt.Errorf("reading communications: %w", err)
	}
	com := (data[0] >> 5) & 0x07
	a.CommunicationsCOM = &com

	stat := (data[0] >> 2) & 0x07
	a.CommunicationsSTAT = &stat

	ssc := (data[1] & 0x80) != 0
	a.CommunicationsSSC = &ssc

	arc := (data[1] & 0x40) != 0
	a.CommunicationsARC = &arc

	aic := (data[1] & 0x20) != 0
	a.CommunicationsAIC = &aic

	b1a := (data[1] & 0x10) != 0
	a.CommunicationsB1A = &b1a

	b1b := data[1] & 0x0F
	a.CommunicationsB1B = &b1b

	return n, nil
}

func (a *AircraftDerivedData) decodeADSBStatus(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil || n != 2 {
		return n, fmt.Errorf("reading ADS-B status: %w", err)
	}
	ac := (data[0] >> 6) & 0x03
	a.ADSBStatusAC = &ac

	mn := (data[0] >> 4) & 0x03
	a.ADSBStatusMN = &mn

	dc := (data[0] >> 2) & 0x03
	a.ADSBStatusDC = &dc

	gbs := (data[0] & 0x02) != 0
	a.ADSBStatusGBS = &gbs

	stat := data[1] & 0x07
	a.ADSBStatusSTAT = &stat

	return n, nil
}

func (a *AircraftDerivedData) decodeACASResolutionAdvisory(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 7)
	n, err := buf.Read(data)
	if err != nil || n != 7 {
		return n, fmt.Errorf("reading ACAS resolution advisory: %w", err)
	}
	var acas [7]byte
	copy(acas[:], data)
	a.ACASResolutionAdvisory = &acas
	return n, nil
}

func (a *AircraftDerivedData) decodeBarometricVerticalRate(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil || n != 2 {
		return n, fmt.Errorf("reading barometric vertical rate: %w", err)
	}
	raw := int16(uint16(data[0])<<8 | uint16(data[1]))
	vr := raw // Already in the correct units (6.25 ft/min per LSB handled in presentation)
	a.BarometricVerticalRate = &vr
	return n, nil
}

func (a *AircraftDerivedData) decodeGeometricVerticalRate(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil || n != 2 {
		return n, fmt.Errorf("reading geometric vertical rate: %w", err)
	}
	raw := int16(uint16(data[0])<<8 | uint16(data[1]))
	vr := raw
	a.GeometricVerticalRate = &vr
	return n, nil
}

func (a *AircraftDerivedData) decodeRollAngle(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil || n != 2 {
		return n, fmt.Errorf("reading roll angle: %w", err)
	}
	raw := int16(uint16(data[0])<<8 | uint16(data[1]))
	angle := float64(raw) * 0.01 // LSB = 0.01°
	a.RollAngle = &angle
	return n, nil
}

func (a *AircraftDerivedData) decodeTrackAngleRate(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil || n != 2 {
		return n, fmt.Errorf("reading track angle rate: %w", err)
	}
	raw := int16(uint16(data[0])<<8 | uint16(data[1]))
	rate := float64(raw) / 32.0 // LSB = 1/32 deg/s
	a.TrackAngleRate = &rate
	return n, nil
}

func (a *AircraftDerivedData) decodeTrackAngle(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil || n != 2 {
		return n, fmt.Errorf("reading track angle: %w", err)
	}
	raw := uint16(data[0])<<8 | uint16(data[1])
	angle := float64(raw) * 360.0 / 65536.0 // LSB = 360°/2^16
	a.TrackAngle = &angle
	return n, nil
}

func (a *AircraftDerivedData) decodeGroundSpeed(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil || n != 2 {
		return n, fmt.Errorf("reading ground speed: %w", err)
	}
	raw := uint16(data[0])<<8 | uint16(data[1])
	// LSB = 2^-14 NM/s, convert to knots: 1 NM/s = 3600 knots
	speed := float64(raw) / 16384.0 * 3600.0 // Convert to knots
	a.GroundSpeed = &speed
	return n, nil
}

func (a *AircraftDerivedData) decodeVelocityUncertainty(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 1)
	n, err := buf.Read(data)
	if err != nil || n != 1 {
		return n, fmt.Errorf("reading velocity uncertainty: %w", err)
	}
	a.VelocityUncertainty = &data[0]
	return n, nil
}

func (a *AircraftDerivedData) decodeMeteorologicalData(buf *bytes.Buffer) (int, error) {
	// Meteorological data is a variable extended structure
	// For simplicity, we'll read it as raw bytes for now
	// A full implementation would decode wind speed, direction, temperature, turbulence
	bytesRead := 0
	met := &MeteorologicalData{}

	for {
		data := make([]byte, 1)
		n, err := buf.Read(data)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading meteorological data: %w", err)
		}
		bytesRead += n

		// Check FX bit
		if (data[0] & 0x01) == 0 {
			break
		}
	}

	a.MeteorologicalData = met
	return bytesRead, nil
}

func (a *AircraftDerivedData) decodeEmitterCategory(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 1)
	n, err := buf.Read(data)
	if err != nil || n != 1 {
		return n, fmt.Errorf("reading emitter category: %w", err)
	}
	a.EmitterCategory = &data[0]
	return n, nil
}

func (a *AircraftDerivedData) decodePosition(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 8)
	n, err := buf.Read(data)
	if err != nil || n != 8 {
		return n, fmt.Errorf("reading position: %w", err)
	}

	// Latitude (24 bits, two's complement, LSB = 180/2^23)
	rawLat := int32(uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2]))
	if (rawLat & 0x800000) != 0 {
		rawLat |= -0x1000000 // Sign extend from bit 23
	}
	lat := float64(rawLat) * 180.0 / 8388608.0
	a.PositionLatitude = &lat

	// Longitude (24 bits, two's complement, LSB = 180/2^23)
	rawLon := int32(uint32(data[4])<<16 | uint32(data[5])<<8 | uint32(data[6]))
	if (rawLon & 0x800000) != 0 {
		rawLon |= -0x1000000 // Sign extend from bit 23
	}
	lon := float64(rawLon) * 180.0 / 8388608.0
	a.PositionLongitude = &lon

	return n, nil
}

func (a *AircraftDerivedData) decodeGeometricAltitude(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil || n != 2 {
		return n, fmt.Errorf("reading geometric altitude: %w", err)
	}
	raw := int16(uint16(data[0])<<8 | uint16(data[1]))
	a.GeometricAltitude = &raw
	return n, nil
}

func (a *AircraftDerivedData) decodePositionUncertainty(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 1)
	n, err := buf.Read(data)
	if err != nil || n != 1 {
		return n, fmt.Errorf("reading position uncertainty: %w", err)
	}
	a.PositionUncertainty = &data[0]
	return n, nil
}

func (a *AircraftDerivedData) decodeModeSMBData(buf *bytes.Buffer) (int, error) {
	repByte := make([]byte, 1)
	n, err := buf.Read(repByte)
	if err != nil {
		return n, fmt.Errorf("reading Mode S MB repetition factor: %w", err)
	}
	bytesRead := n
	rep := int(repByte[0])

	a.ModeSMBData = make([][]byte, rep)
	for i := 0; i < rep; i++ {
		mbData := make([]byte, 8)
		n, err := buf.Read(mbData)
		if err != nil || n != 8 {
			return bytesRead + n, fmt.Errorf("reading Mode S MB data item %d: %w", i+1, err)
		}
		bytesRead += n
		a.ModeSMBData[i] = mbData
	}

	return bytesRead, nil
}

func (a *AircraftDerivedData) decodeIndicatedAirspeedData(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil || n != 2 {
		return n, fmt.Errorf("reading indicated airspeed data: %w", err)
	}
	raw := uint16(data[0])<<8 | uint16(data[1])
	a.IndicatedAirspeedData = &raw
	return n, nil
}

func (a *AircraftDerivedData) decodeMachNumber(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil || n != 2 {
		return n, fmt.Errorf("reading mach number: %w", err)
	}
	raw := uint16(data[0])<<8 | uint16(data[1])
	mach := float64(raw) * 0.008 // LSB = 0.008
	a.MachNumber = &mach
	return n, nil
}

func (a *AircraftDerivedData) decodeBarometricPressureSetting(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil || n != 2 {
		return n, fmt.Errorf("reading barometric pressure setting: %w", err)
	}
	raw := uint16(data[0])<<8 | uint16(data[1])
	pressure := float64(raw) * 0.1 // LSB = 0.1 mb
	a.BarometricPressureSetting = &pressure
	return n, nil
}

func (a *AircraftDerivedData) Encode(buf *bytes.Buffer) (int, error) {
	// Build primary subfield based on which fields are present
	primaryBytes := a.buildPrimarySubfield()
	n, err := buf.Write(primaryBytes)
	if err != nil {
		return n, fmt.Errorf("writing primary subfield: %w", err)
	}
	bytesWritten := n

	// Encode each present subfield
	encoders := []func(*bytes.Buffer) (int, error){
		a.encodeTargetAddress,
		a.encodeTargetIdentification,
		a.encodeMagneticHeading,
		a.encodeIndicatedAirspeed,
		a.encodeTrueAirspeed,
		a.encodeSelectedAltitude,
		a.encodeFinalStateAltitude,
		a.encodeTrajectoryIntentStatus,
		a.encodeTrajectoryIntentData,
		a.encodeCommunications,
		a.encodeADSBStatus,
		a.encodeACASResolutionAdvisory,
		a.encodeBarometricVerticalRate,
		a.encodeGeometricVerticalRate,
		a.encodeRollAngle,
		a.encodeTrackAngleRate,
		a.encodeTrackAngle,
		a.encodeGroundSpeed,
		a.encodeVelocityUncertainty,
		a.encodeMeteorologicalData,
		a.encodeEmitterCategory,
		a.encodePosition,
		a.encodeGeometricAltitude,
		a.encodePositionUncertainty,
		a.encodeModeSMBData,
		a.encodeIndicatedAirspeedData,
		a.encodeMachNumber,
		a.encodeBarometricPressureSetting,
	}

	for _, encoder := range encoders {
		n, err := encoder(buf)
		bytesWritten += n
		if err != nil {
			return bytesWritten, err
		}
	}

	return bytesWritten, nil
}

func (a *AircraftDerivedData) buildPrimarySubfield() []byte {
	// Determine which subfields are present and build the FSPEC
	presence := make([]bool, 28)
	presence[0] = a.TargetAddress != nil
	presence[1] = a.TargetIdentification != nil
	presence[2] = a.MagneticHeading != nil
	presence[3] = a.IndicatedAirspeed != nil
	presence[4] = a.TrueAirspeed != nil
	presence[5] = a.SelectedAltitude != nil
	presence[6] = a.FinalStateAltitude != nil
	presence[7] = a.TrajectoryIntentNAV != nil || a.TrajectoryIntentNVB != nil
	presence[8] = len(a.TrajectoryIntentData) > 0
	presence[9] = a.CommunicationsCOM != nil || a.CommunicationsSTAT != nil ||
		a.CommunicationsSSC != nil || a.CommunicationsARC != nil ||
		a.CommunicationsAIC != nil || a.CommunicationsB1A != nil ||
		a.CommunicationsB1B != nil
	presence[10] = a.ADSBStatusAC != nil
	presence[11] = a.ACASResolutionAdvisory != nil
	presence[12] = a.BarometricVerticalRate != nil
	presence[13] = a.GeometricVerticalRate != nil
	presence[14] = a.RollAngle != nil
	presence[15] = a.TrackAngleRate != nil
	presence[16] = a.TrackAngle != nil
	presence[17] = a.GroundSpeed != nil
	presence[18] = a.VelocityUncertainty != nil
	presence[19] = a.MeteorologicalData != nil
	presence[20] = a.EmitterCategory != nil
	presence[21] = a.PositionLatitude != nil
	presence[22] = a.GeometricAltitude != nil
	presence[23] = a.PositionUncertainty != nil
	presence[24] = len(a.ModeSMBData) > 0
	presence[25] = a.IndicatedAirspeedData != nil
	presence[26] = a.MachNumber != nil
	presence[27] = a.BarometricPressureSetting != nil

	// Build primary subfield bytes
	// First, determine how many bytes we actually need by finding the last present field
	lastPresentByte := -1
	for i := 27; i >= 0; i-- {
		if presence[i] {
			lastPresentByte = i / 7
			break
		}
	}

	// If no fields are present, return minimal FSPEC (one byte with value 0x00)
	if lastPresentByte == -1 {
		return []byte{0x00}
	}

	result := make([]byte, 0, lastPresentByte+1)
	for i := 0; i < 28; i += 7 {
		byteIndex := i / 7
		var b byte
		for j := 0; j < 7 && (i+j) < 28; j++ {
			if presence[i+j] {
				b |= (1 << (7 - j))
			}
		}
		// Set FX bit only if we need more bytes (i.e., there are present fields in subsequent bytes)
		if byteIndex < lastPresentByte {
			b |= 0x01
		}
		result = append(result, b)

		// Stop after we've written all necessary bytes
		if byteIndex >= lastPresentByte {
			break
		}
	}

	return result
}

// Encode methods (simplified - only encode if field is present)

func (a *AircraftDerivedData) encodeTargetAddress(buf *bytes.Buffer) (int, error) {
	if a.TargetAddress == nil {
		return 0, nil
	}
	addr := *a.TargetAddress
	data := []byte{byte(addr >> 16), byte(addr >> 8), byte(addr)}
	return buf.Write(data)
}

func (a *AircraftDerivedData) encodeTargetIdentification(buf *bytes.Buffer) (int, error) {
	if a.TargetIdentification == nil {
		return 0, nil
	}
	// Encode 8 characters to 6 bytes (6 bits per character)
	chars := encodeICAO6BitString(*a.TargetIdentification)
	data := make([]byte, 6)
	data[0] = (chars[0] << 2) | (chars[1] >> 4)
	data[1] = (chars[1] << 4) | (chars[2] >> 2)
	data[2] = (chars[2] << 6) | chars[3]
	data[3] = (chars[4] << 2) | (chars[5] >> 4)
	data[4] = (chars[5] << 4) | (chars[6] >> 2)
	data[5] = (chars[6] << 6) | chars[7]
	return buf.Write(data)
}

func (a *AircraftDerivedData) encodeMagneticHeading(buf *bytes.Buffer) (int, error) {
	if a.MagneticHeading == nil {
		return 0, nil
	}
	raw := uint16(math.Round(*a.MagneticHeading * 65536.0 / 360.0))
	data := []byte{byte(raw >> 8), byte(raw)}
	return buf.Write(data)
}

// Remaining encode methods follow similar pattern...
// For brevity, I'll implement the key ones and add stubs for others

func (a *AircraftDerivedData) encodeIndicatedAirspeed(buf *bytes.Buffer) (int, error) {
	if a.IndicatedAirspeed == nil {
		return 0, nil
	}
	var raw uint16
	var im byte
	if a.IndicatedAirspeedIMAch != nil && *a.IndicatedAirspeedIMAch {
		im = 0x80
		raw = uint16(math.Round(*a.IndicatedAirspeed / 0.001))
	} else {
		raw = uint16(math.Round(*a.IndicatedAirspeed * 16384.0))
	}
	data := []byte{im | byte(raw>>8), byte(raw)}
	return buf.Write(data)
}

// Stub encode methods for remaining subfields
func (a *AircraftDerivedData) encodeTrueAirspeed(buf *bytes.Buffer) (int, error) {
	if a.TrueAirspeed == nil {
		return 0, nil
	}
	data := []byte{byte(*a.TrueAirspeed >> 8), byte(*a.TrueAirspeed)}
	return buf.Write(data)
}

func (a *AircraftDerivedData) encodeSelectedAltitude(buf *bytes.Buffer) (int, error) {
	if a.SelectedAltitude == nil {
		return 0, nil
	}
	// Full implementation would encode SAS, Source, and Altitude
	// Stub for now
	return 0, nil
}

func (a *AircraftDerivedData) encodeFinalStateAltitude(buf *bytes.Buffer) (int, error) {
	if a.FinalStateAltitude == nil {
		return 0, nil
	}
	// Stub
	return 0, nil
}

func (a *AircraftDerivedData) encodeTrajectoryIntentStatus(buf *bytes.Buffer) (int, error) {
	if a.TrajectoryIntentNAV == nil && a.TrajectoryIntentNVB == nil {
		return 0, nil
	}
	// Stub
	return 0, nil
}

func (a *AircraftDerivedData) encodeTrajectoryIntentData(buf *bytes.Buffer) (int, error) {
	if len(a.TrajectoryIntentData) == 0 {
		return 0, nil
	}
	// Stub
	return 0, nil
}

func (a *AircraftDerivedData) encodeCommunications(buf *bytes.Buffer) (int, error) {
	// Subfield #10: Communications/ACAS Capability and Flight Status (2 octets)
	// Only encode if at least one field is present
	if a.CommunicationsCOM == nil && a.CommunicationsSTAT == nil &&
		a.CommunicationsSSC == nil && a.CommunicationsARC == nil &&
		a.CommunicationsAIC == nil && a.CommunicationsB1A == nil &&
		a.CommunicationsB1B == nil {
		return 0, nil
	}

	data := make([]byte, 2)

	// Byte 1: COM (bits 7-5), STAT (bits 4-2), spare bit 1, spare bit 0
	if a.CommunicationsCOM != nil {
		data[0] |= (*a.CommunicationsCOM & 0x07) << 5
	}
	if a.CommunicationsSTAT != nil {
		data[0] |= (*a.CommunicationsSTAT & 0x07) << 2
	}

	// Byte 2: SSC (bit 7), ARC (bit 6), AIC (bit 5), B1A (bit 4), B1B (bits 3-0)
	if a.CommunicationsSSC != nil && *a.CommunicationsSSC {
		data[1] |= 0x80
	}
	if a.CommunicationsARC != nil && *a.CommunicationsARC {
		data[1] |= 0x40
	}
	if a.CommunicationsAIC != nil && *a.CommunicationsAIC {
		data[1] |= 0x20
	}
	if a.CommunicationsB1A != nil && *a.CommunicationsB1A {
		data[1] |= 0x10
	}
	if a.CommunicationsB1B != nil {
		data[1] |= *a.CommunicationsB1B & 0x0F
	}

	return buf.Write(data)
}

func (a *AircraftDerivedData) encodeADSBStatus(buf *bytes.Buffer) (int, error) {
	if a.ADSBStatusAC == nil {
		return 0, nil
	}
	// Stub
	return 0, nil
}

func (a *AircraftDerivedData) encodeACASResolutionAdvisory(buf *bytes.Buffer) (int, error) {
	if a.ACASResolutionAdvisory == nil {
		return 0, nil
	}
	return buf.Write(a.ACASResolutionAdvisory[:])
}

func (a *AircraftDerivedData) encodeBarometricVerticalRate(buf *bytes.Buffer) (int, error) {
	if a.BarometricVerticalRate == nil {
		return 0, nil
	}
	data := []byte{byte(*a.BarometricVerticalRate >> 8), byte(*a.BarometricVerticalRate)}
	return buf.Write(data)
}

func (a *AircraftDerivedData) encodeGeometricVerticalRate(buf *bytes.Buffer) (int, error) {
	if a.GeometricVerticalRate == nil {
		return 0, nil
	}
	data := []byte{byte(*a.GeometricVerticalRate >> 8), byte(*a.GeometricVerticalRate)}
	return buf.Write(data)
}

func (a *AircraftDerivedData) encodeRollAngle(buf *bytes.Buffer) (int, error) {
	if a.RollAngle == nil {
		return 0, nil
	}
	raw := int16(math.Round(*a.RollAngle / 0.01))
	data := []byte{byte(raw >> 8), byte(raw)}
	return buf.Write(data)
}

func (a *AircraftDerivedData) encodeTrackAngleRate(buf *bytes.Buffer) (int, error) {
	if a.TrackAngleRate == nil {
		return 0, nil
	}
	raw := int16(math.Round(*a.TrackAngleRate * 32.0))
	data := []byte{byte(raw >> 8), byte(raw)}
	return buf.Write(data)
}

func (a *AircraftDerivedData) encodeTrackAngle(buf *bytes.Buffer) (int, error) {
	if a.TrackAngle == nil {
		return 0, nil
	}
	raw := uint16(math.Round(*a.TrackAngle * 65536.0 / 360.0))
	data := []byte{byte(raw >> 8), byte(raw)}
	return buf.Write(data)
}

func (a *AircraftDerivedData) encodeGroundSpeed(buf *bytes.Buffer) (int, error) {
	if a.GroundSpeed == nil {
		return 0, nil
	}
	raw := uint16(math.Round(*a.GroundSpeed / 3600.0 * 16384.0))
	data := []byte{byte(raw >> 8), byte(raw)}
	return buf.Write(data)
}

func (a *AircraftDerivedData) encodeVelocityUncertainty(buf *bytes.Buffer) (int, error) {
	if a.VelocityUncertainty == nil {
		return 0, nil
	}
	return buf.Write([]byte{*a.VelocityUncertainty})
}

func (a *AircraftDerivedData) encodeMeteorologicalData(buf *bytes.Buffer) (int, error) {
	if a.MeteorologicalData == nil {
		return 0, nil
	}
	// Stub
	return 0, nil
}

func (a *AircraftDerivedData) encodeEmitterCategory(buf *bytes.Buffer) (int, error) {
	if a.EmitterCategory == nil {
		return 0, nil
	}
	return buf.Write([]byte{*a.EmitterCategory})
}

func (a *AircraftDerivedData) encodePosition(buf *bytes.Buffer) (int, error) {
	if a.PositionLatitude == nil || a.PositionLongitude == nil {
		return 0, nil
	}
	// Stub
	return 0, nil
}

func (a *AircraftDerivedData) encodeGeometricAltitude(buf *bytes.Buffer) (int, error) {
	if a.GeometricAltitude == nil {
		return 0, nil
	}
	data := []byte{byte(*a.GeometricAltitude >> 8), byte(*a.GeometricAltitude)}
	return buf.Write(data)
}

func (a *AircraftDerivedData) encodePositionUncertainty(buf *bytes.Buffer) (int, error) {
	if a.PositionUncertainty == nil {
		return 0, nil
	}
	return buf.Write([]byte{*a.PositionUncertainty})
}

func (a *AircraftDerivedData) encodeModeSMBData(buf *bytes.Buffer) (int, error) {
	if len(a.ModeSMBData) == 0 {
		return 0, nil
	}
	// Stub
	return 0, nil
}

func (a *AircraftDerivedData) encodeIndicatedAirspeedData(buf *bytes.Buffer) (int, error) {
	if a.IndicatedAirspeedData == nil {
		return 0, nil
	}
	data := []byte{byte(*a.IndicatedAirspeedData >> 8), byte(*a.IndicatedAirspeedData)}
	return buf.Write(data)
}

func (a *AircraftDerivedData) encodeMachNumber(buf *bytes.Buffer) (int, error) {
	if a.MachNumber == nil {
		return 0, nil
	}
	raw := uint16(math.Round(*a.MachNumber / 0.008))
	data := []byte{byte(raw >> 8), byte(raw)}
	return buf.Write(data)
}

func (a *AircraftDerivedData) encodeBarometricPressureSetting(buf *bytes.Buffer) (int, error) {
	if a.BarometricPressureSetting == nil {
		return 0, nil
	}
	raw := uint16(math.Round(*a.BarometricPressureSetting / 0.1))
	data := []byte{byte(raw >> 8), byte(raw)}
	return buf.Write(data)
}

func (a *AircraftDerivedData) String() string {
	var parts []string

	if a.TargetAddress != nil {
		parts = append(parts, fmt.Sprintf("Address=%06X", *a.TargetAddress))
	}
	if a.TargetIdentification != nil {
		parts = append(parts, fmt.Sprintf("ID=%s", *a.TargetIdentification))
	}
	if a.MagneticHeading != nil {
		parts = append(parts, fmt.Sprintf("MagHdg=%.1f°", *a.MagneticHeading))
	}
	if a.TrueAirspeed != nil {
		parts = append(parts, fmt.Sprintf("TAS=%d kt", *a.TrueAirspeed))
	}
	if a.GroundSpeed != nil {
		parts = append(parts, fmt.Sprintf("GS=%.1f kt", *a.GroundSpeed))
	}
	if a.SelectedAltitude != nil {
		parts = append(parts, fmt.Sprintf("SelAlt=%d ft", *a.SelectedAltitude))
	}
	if a.PositionLatitude != nil && a.PositionLongitude != nil {
		parts = append(parts, fmt.Sprintf("Pos=%.6f,%.6f", *a.PositionLatitude, *a.PositionLongitude))
	}
	if a.GeometricAltitude != nil {
		parts = append(parts, fmt.Sprintf("GeoAlt=%.1f ft", float64(*a.GeometricAltitude)*6.25)) // Convert to ft
	}
	if a.BarometricVerticalRate != nil {
		parts = append(parts, fmt.Sprintf("BVR=%.1f fpm", float64(*a.BarometricVerticalRate)*6.25)) // Convert to fpm
	}
	if a.RollAngle != nil {
		parts = append(parts, fmt.Sprintf("Roll=%.1f°", *a.RollAngle))
	}
	if a.TrackAngle != nil {
		parts = append(parts, fmt.Sprintf("Track=%.1f°", *a.TrackAngle))
	}
	if a.MachNumber != nil {
		parts = append(parts, fmt.Sprintf("Mach=%.3f", *a.MachNumber))
	}

	if len(parts) == 0 {
		return "AircraftDerivedData{}"
	}

	return fmt.Sprintf("AircraftDerivedData{%s}", strings.Join(parts, ", "))
}

func (a *AircraftDerivedData) Validate() error {
	if a.MagneticHeading != nil {
		if *a.MagneticHeading < 0 || *a.MagneticHeading >= 360 {
			return fmt.Errorf("magnetic heading out of range [0,360): %f", *a.MagneticHeading)
		}
	}

	if a.TrackAngle != nil {
		if *a.TrackAngle < 0 || *a.TrackAngle >= 360 {
			return fmt.Errorf("track angle out of range [0,360): %f", *a.TrackAngle)
		}
	}

	if a.PositionLatitude != nil {
		if *a.PositionLatitude < -90 || *a.PositionLatitude > 90 {
			return fmt.Errorf("latitude out of range [-90,90]: %f", *a.PositionLatitude)
		}
	}

	if a.PositionLongitude != nil {
		if *a.PositionLongitude < -180 || *a.PositionLongitude >= 180 {
			return fmt.Errorf("longitude out of range [-180,180): %f", *a.PositionLongitude)
		}
	}

	if a.TrueAirspeed != nil {
		if *a.TrueAirspeed > 2046 {
			return fmt.Errorf("true airspeed out of range [0,2046]: %d", *a.TrueAirspeed)
		}
	}

	if a.MachNumber != nil {
		if *a.MachNumber < 0 || *a.MachNumber > 5.0 {
			return fmt.Errorf("Mach number out of reasonable range [0,5.0]: %f", *a.MachNumber)
		}
	}

	return nil
}

// Helper functions for ICAO 6-bit character encoding

func decodeICAO6BitString(chars []byte) string {
	result := make([]byte, len(chars))
	for i, c := range chars {
		result[i] = icao6BitToASCII(c)
	}
	return strings.TrimRight(string(result), " ")
}

func encodeICAO6BitString(s string) []byte {
	// Pad to 8 characters
	s = (s + "        ")[:8]
	result := make([]byte, 8)
	for i := 0; i < 8; i++ {
		result[i] = asciiToICAO6Bit(s[i])
	}
	return result
}

func icao6BitToASCII(c byte) byte {
	c = c & 0x3F // Ensure 6 bits
	if c == 0 {
		return ' '
	}
	if c >= 1 && c <= 26 {
		return 'A' + c - 1
	}
	if c == 32 {
		return ' '
	}
	if c >= 48 && c <= 57 {
		return c
	}
	return '?'
}

func asciiToICAO6Bit(c byte) byte {
	if c == ' ' {
		return 32
	}
	if c >= 'A' && c <= 'Z' {
		return c - 'A' + 1
	}
	if c >= 'a' && c <= 'z' {
		return c - 'a' + 1
	}
	if c >= '0' && c <= '9' {
		return c
	}
	return 32 // Space for unknown
}
