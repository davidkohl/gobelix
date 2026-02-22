// dataitems/cat062/flight_plan_related_data.go
package v120

import (
	"bytes"
	"fmt"
	"strings"
)

// FlightPlanRelatedData implements I062/390
// All flight plan related information, provided by ground-based systems
type FlightPlanRelatedData struct {
	// Subfield #1: FPPS Identification Tag
	FPPSIdentificationTag *uint16 // SAC/SIC of FPPS

	// Subfield #2: Callsign
	Callsign *string // 7 characters, IA5 encoding

	// Subfield #3: IFPS_FLIGHT_ID
	IFPSFlightID *[4]byte // 4 characters

	// Subfield #4: Flight Category
	FlightCategory *uint8 // 2 bits: 0=Unknown, 1=GAT, 2=OAT, 3=Military

	// Subfield #5: Type of Aircraft
	TypeOfAircraft *string // 4 characters, ICAO designator

	// Subfield #6: Wake Turbulence Category
	WakeTurbulenceCategory *uint8 // 'L', 'M', 'H', 'J' (ASCII)

	// Subfield #7: Departure Airport
	DepartureAirport *string // 4 characters, ICAO code

	// Subfield #8: Destination Airport
	DestinationAirport *string // 4 characters, ICAO code

	// Subfield #9: Runway Designation
	RunwayDesignation *string // 3 characters

	// Subfield #10: Current Cleared Flight Level
	CurrentClearedFlightLevel *int16 // Flight level in hundreds of feet, LSB = 0.25 FL

	// Subfield #11: Current Control Position
	CurrentControlPositionCentre *uint8  // 8-bit centre
	CurrentControlPositionPos    *uint8  // 8-bit position

	// Subfield #12: Time of Departure/Arrival (Repetitive)
	TimeOfDepartureArrival []TimeOfDepartureArrival

	// Subfield #13: Aircraft Stand
	AircraftStand *string // 6 characters

	// Subfield #14: Stand Status
	StandStatusEMP *bool // Empty
	StandStatusAVL *bool // Available
	StandStatusOCC *bool // Occupied

	// Subfield #15: Standard Instrument Departure
	StandardInstrumentDeparture *string // 7 characters

	// Subfield #16: Standard Instrument Arrival
	StandardInstrumentArrival *string // 7 characters

	// Subfield #17: Pre-emergency Mode 3/A
	PreEmergencyMode3A *uint16 // 12-bit Mode 3/A code

	// Subfield #18: Pre-emergency Callsign
	PreEmergencyCallsign *string // 7 characters
}

// TimeOfDepartureArrival represents a single time entry
type TimeOfDepartureArrival struct {
	TYP uint8   // 5-bit type indicator
	DAY uint8   // 2-bit day (today, yesterday, tomorrow, +1)
	HOR uint8   // 4-bit hour correction (-8 to +7 from DAY)
	MIN uint8   // 6-bit minutes
	AVS bool    // Available for departure/arrival
	SEC uint16  // 16-bit seconds from midnight
}

func (f *FlightPlanRelatedData) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Read primary subfield (up to 3 octets with FX extension bits)
	var primaryBytes [3]byte
	primaryLen := 0
	for i := 0; i < 3; i++ {
		octet, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading flight plan related data primary subfield octet %d: %w", i+1, err)
		}
		bytesRead++
		primaryBytes[i] = octet
		primaryLen++

		// Check if there's an extension
		hasExtension := (octet & 0x01) != 0
		if !hasExtension {
			break
		}
	}

	// Now read subfields based on the bits set in the primary subfield
	// We process bits from bit-24 down to bit-1 (excluding FX bits)
	subfieldIndex := 0
	for byteIdx := 0; byteIdx < primaryLen; byteIdx++ {
		// Process bits 8-2 (bit 1 is FX)
		for bitPos := 7; bitPos >= 1; bitPos-- {
			if (primaryBytes[byteIdx] & (1 << bitPos)) != 0 {
				// This subfield is present
				n, err := f.decodeSubfield(subfieldIndex, buf)
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

func (f *FlightPlanRelatedData) decodeSubfield(index int, buf *bytes.Buffer) (int, error) {
	switch index {
	case 0: // #1: FPPS Identification Tag (2 octets)
		return f.decodeFPPSIdentificationTag(buf)
	case 1: // #2: Callsign (7 octets)
		return f.decodeCallsign(buf)
	case 2: // #3: IFPS_FLIGHT_ID (4 octets)
		return f.decodeIFPSFlightID(buf)
	case 3: // #4: Flight Category (1 octet)
		return f.decodeFlightCategory(buf)
	case 4: // #5: Type of Aircraft (4 octets)
		return f.decodeTypeOfAircraft(buf)
	case 5: // #6: Wake Turbulence Category (1 octet)
		return f.decodeWakeTurbulenceCategory(buf)
	case 6: // #7: Departure Airport (4 octets)
		return f.decodeDepartureAirport(buf)
	case 7: // #8: Destination Airport (4 octets)
		return f.decodeDestinationAirport(buf)
	case 8: // #9: Runway Designation (3 octets)
		return f.decodeRunwayDesignation(buf)
	case 9: // #10: Current Cleared Flight Level (2 octets)
		return f.decodeCurrentClearedFlightLevel(buf)
	case 10: // #11: Current Control Position (2 octets)
		return f.decodeCurrentControlPosition(buf)
	case 11: // #12: Time of Departure/Arrival (Repetitive)
		return f.decodeTimeOfDepartureArrival(buf)
	case 12: // #13: Aircraft Stand (6 octets)
		return f.decodeAircraftStand(buf)
	case 13: // #14: Stand Status (1 octet)
		return f.decodeStandStatus(buf)
	case 14: // #15: Standard Instrument Departure (7 octets)
		return f.decodeStandardInstrumentDeparture(buf)
	case 15: // #16: Standard Instrument Arrival (7 octets)
		return f.decodeStandardInstrumentArrival(buf)
	case 16: // #17: Pre-emergency Mode 3/A (2 octets)
		return f.decodePreEmergencyMode3A(buf)
	case 17: // #18: Pre-emergency Callsign (7 octets)
		return f.decodePreEmergencyCallsign(buf)
	default:
		return 0, fmt.Errorf("unknown subfield index: %d", index)
	}
}

// Individual subfield decoders

func (f *FlightPlanRelatedData) decodeFPPSIdentificationTag(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil || n != 2 {
		return n, fmt.Errorf("reading FPPS identification tag: %w", err)
	}
	tag := uint16(data[0])<<8 | uint16(data[1])
	f.FPPSIdentificationTag = &tag
	return n, nil
}

func (f *FlightPlanRelatedData) decodeCallsign(buf *bytes.Buffer) (int, error) {
	var data [7]byte
	n, err := buf.Read(data[:])
	if err != nil || n != 7 {
		return n, fmt.Errorf("reading callsign: %w", err)
	}
	// IA5 encoding - just convert to string and trim spaces
	callsign := strings.TrimRight(string(data[:]), " ")
	f.Callsign = &callsign
	return n, nil
}

func (f *FlightPlanRelatedData) decodeIFPSFlightID(buf *bytes.Buffer) (int, error) {
	var data [4]byte
	n, err := buf.Read(data[:])
	if err != nil || n != 4 {
		return n, fmt.Errorf("reading IFPS flight ID: %w", err)
	}
	f.IFPSFlightID = &data
	return n, nil
}

func (f *FlightPlanRelatedData) decodeFlightCategory(buf *bytes.Buffer) (int, error) {
	data, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading flight category: %w", err)
	}
	cat := (data >> 6) & 0x03 // Bits 8-7
	f.FlightCategory = &cat
	return 1, nil
}

func (f *FlightPlanRelatedData) decodeTypeOfAircraft(buf *bytes.Buffer) (int, error) {
	var data [4]byte
	n, err := buf.Read(data[:])
	if err != nil || n != 4 {
		return n, fmt.Errorf("reading type of aircraft: %w", err)
	}
	// IA5 encoding - ICAO aircraft designator
	acType := strings.TrimRight(string(data[:]), " ")
	f.TypeOfAircraft = &acType
	return n, nil
}

func (f *FlightPlanRelatedData) decodeWakeTurbulenceCategory(buf *bytes.Buffer) (int, error) {
	data, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading wake turbulence category: %w", err)
	}
	f.WakeTurbulenceCategory = &data
	return 1, nil
}

func (f *FlightPlanRelatedData) decodeDepartureAirport(buf *bytes.Buffer) (int, error) {
	var data [4]byte
	n, err := buf.Read(data[:])
	if err != nil || n != 4 {
		return n, fmt.Errorf("reading departure airport: %w", err)
	}
	// IA5 encoding - ICAO airport code
	airport := strings.TrimRight(string(data[:]), " ")
	f.DepartureAirport = &airport
	return n, nil
}

func (f *FlightPlanRelatedData) decodeDestinationAirport(buf *bytes.Buffer) (int, error) {
	var data [4]byte
	n, err := buf.Read(data[:])
	if err != nil || n != 4 {
		return n, fmt.Errorf("reading destination airport: %w", err)
	}
	// IA5 encoding - ICAO airport code
	airport := strings.TrimRight(string(data[:]), " ")
	f.DestinationAirport = &airport
	return n, nil
}

func (f *FlightPlanRelatedData) decodeRunwayDesignation(buf *bytes.Buffer) (int, error) {
	var data [3]byte
	n, err := buf.Read(data[:])
	if err != nil || n != 3 {
		return n, fmt.Errorf("reading runway designation: %w", err)
	}
	// IA5 encoding
	runway := strings.TrimRight(string(data[:]), " ")
	f.RunwayDesignation = &runway
	return n, nil
}

func (f *FlightPlanRelatedData) decodeCurrentClearedFlightLevel(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil || n != 2 {
		return n, fmt.Errorf("reading current cleared flight level: %w", err)
	}
	// 16-bit two's complement, LSB = 0.25 FL (stored as raw value)
	fl := int16(uint16(data[0])<<8 | uint16(data[1]))
	f.CurrentClearedFlightLevel = &fl
	return n, nil
}

func (f *FlightPlanRelatedData) decodeCurrentControlPosition(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil || n != 2 {
		return n, fmt.Errorf("reading current control position: %w", err)
	}
	centre := data[0]
	pos := data[1]
	f.CurrentControlPositionCentre = &centre
	f.CurrentControlPositionPos = &pos
	return n, nil
}

func (f *FlightPlanRelatedData) decodeTimeOfDepartureArrival(buf *bytes.Buffer) (int, error) {
	repByte, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading time of departure/arrival repetition factor: %w", err)
	}
	bytesRead := 1
	rep := int(repByte)

	f.TimeOfDepartureArrival = make([]TimeOfDepartureArrival, rep)
	for i := 0; i < rep; i++ {
		var todData [4]byte
		n, err := buf.Read(todData[:])
		if err != nil || n != 4 {
			return bytesRead + n, fmt.Errorf("reading time of departure/arrival item %d: %w", i+1, err)
		}
		bytesRead += n

		// Decode time entry
		entry := &f.TimeOfDepartureArrival[i]
		entry.TYP = (todData[0] >> 3) & 0x1F // Bits 8-4
		entry.DAY = (todData[0] >> 1) & 0x03 // Bits 3-2
		entry.HOR = ((todData[0] & 0x01) << 3) | ((todData[1] >> 5) & 0x07) // Bit 1 + bits 8-6
		entry.MIN = (todData[1] >> 2) & 0x3F // Bits 5-1 (6 bits spans across bytes, actually bits 13-8)
		entry.AVS = (todData[1] & 0x02) != 0 // Bit 2 of second byte
		entry.SEC = uint16(todData[2])<<8 | uint16(todData[3])
	}

	return bytesRead, nil
}

func (f *FlightPlanRelatedData) decodeAircraftStand(buf *bytes.Buffer) (int, error) {
	var data [6]byte
	n, err := buf.Read(data[:])
	if err != nil || n != 6 {
		return n, fmt.Errorf("reading aircraft stand: %w", err)
	}
	// IA5 encoding
	stand := strings.TrimRight(string(data[:]), " ")
	f.AircraftStand = &stand
	return n, nil
}

func (f *FlightPlanRelatedData) decodeStandStatus(buf *bytes.Buffer) (int, error) {
	data, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading stand status: %w", err)
	}
	emp := (data & 0x80) != 0
	avl := (data & 0x40) != 0
	occ := (data & 0x20) != 0
	f.StandStatusEMP = &emp
	f.StandStatusAVL = &avl
	f.StandStatusOCC = &occ
	return 1, nil
}

func (f *FlightPlanRelatedData) decodeStandardInstrumentDeparture(buf *bytes.Buffer) (int, error) {
	var data [7]byte
	n, err := buf.Read(data[:])
	if err != nil || n != 7 {
		return n, fmt.Errorf("reading standard instrument departure: %w", err)
	}
	// IA5 encoding
	sid := strings.TrimRight(string(data[:]), " ")
	f.StandardInstrumentDeparture = &sid
	return n, nil
}

func (f *FlightPlanRelatedData) decodeStandardInstrumentArrival(buf *bytes.Buffer) (int, error) {
	var data [7]byte
	n, err := buf.Read(data[:])
	if err != nil || n != 7 {
		return n, fmt.Errorf("reading standard instrument arrival: %w", err)
	}
	// IA5 encoding
	star := strings.TrimRight(string(data[:]), " ")
	f.StandardInstrumentArrival = &star
	return n, nil
}

func (f *FlightPlanRelatedData) decodePreEmergencyMode3A(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil || n != 2 {
		return n, fmt.Errorf("reading pre-emergency Mode 3/A: %w", err)
	}
	// 12-bit Mode 3/A code in bits 12-1
	code := uint16(data[0]&0x0F)<<8 | uint16(data[1])
	f.PreEmergencyMode3A = &code
	return n, nil
}

func (f *FlightPlanRelatedData) decodePreEmergencyCallsign(buf *bytes.Buffer) (int, error) {
	var data [7]byte
	n, err := buf.Read(data[:])
	if err != nil || n != 7 {
		return n, fmt.Errorf("reading pre-emergency callsign: %w", err)
	}
	// IA5 encoding
	callsign := strings.TrimRight(string(data[:]), " ")
	f.PreEmergencyCallsign = &callsign
	return n, nil
}

func (f *FlightPlanRelatedData) Encode(buf *bytes.Buffer) (int, error) {
	// Build primary subfield based on which fields are present
	primaryBytes := f.buildPrimarySubfield()
	n, err := buf.Write(primaryBytes)
	if err != nil {
		return n, fmt.Errorf("writing primary subfield: %w", err)
	}
	bytesWritten := n

	// Encode each present subfield
	encoders := []func(*bytes.Buffer) (int, error){
		f.encodeFPPSIdentificationTag,
		f.encodeCallsign,
		f.encodeIFPSFlightID,
		f.encodeFlightCategory,
		f.encodeTypeOfAircraft,
		f.encodeWakeTurbulenceCategory,
		f.encodeDepartureAirport,
		f.encodeDestinationAirport,
		f.encodeRunwayDesignation,
		f.encodeCurrentClearedFlightLevel,
		f.encodeCurrentControlPosition,
		f.encodeTimeOfDepartureArrival,
		f.encodeAircraftStand,
		f.encodeStandStatus,
		f.encodeStandardInstrumentDeparture,
		f.encodeStandardInstrumentArrival,
		f.encodePreEmergencyMode3A,
		f.encodePreEmergencyCallsign,
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

func (f *FlightPlanRelatedData) buildPrimarySubfield() []byte {
	// Determine which subfields are present and build the FSPEC
	presence := make([]bool, 18)
	presence[0] = f.FPPSIdentificationTag != nil
	presence[1] = f.Callsign != nil
	presence[2] = f.IFPSFlightID != nil
	presence[3] = f.FlightCategory != nil
	presence[4] = f.TypeOfAircraft != nil
	presence[5] = f.WakeTurbulenceCategory != nil
	presence[6] = f.DepartureAirport != nil
	presence[7] = f.DestinationAirport != nil
	presence[8] = f.RunwayDesignation != nil
	presence[9] = f.CurrentClearedFlightLevel != nil
	presence[10] = f.CurrentControlPositionCentre != nil || f.CurrentControlPositionPos != nil
	presence[11] = len(f.TimeOfDepartureArrival) > 0
	presence[12] = f.AircraftStand != nil
	presence[13] = f.StandStatusEMP != nil || f.StandStatusAVL != nil || f.StandStatusOCC != nil
	presence[14] = f.StandardInstrumentDeparture != nil
	presence[15] = f.StandardInstrumentArrival != nil
	presence[16] = f.PreEmergencyMode3A != nil
	presence[17] = f.PreEmergencyCallsign != nil

	// Find the last byte that has any present fields
	lastPresentByte := -1
	for i := 17; i >= 0; i-- {
		if presence[i] {
			lastPresentByte = i / 7
			break
		}
	}

	// If no fields are present, return minimal FSPEC (one byte with value 0x00)
	if lastPresentByte == -1 {
		return []byte{0x00}
	}

	// Build primary subfield bytes (up to 3 bytes for 18 subfields)
	result := make([]byte, 0, lastPresentByte+1)
	for i := 0; i < 18; i += 7 {
		byteIndex := i / 7
		var b byte
		for j := 0; j < 7 && (i+j) < 18; j++ {
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

// Encode methods

func (f *FlightPlanRelatedData) encodeFPPSIdentificationTag(buf *bytes.Buffer) (int, error) {
	if f.FPPSIdentificationTag == nil {
		return 0, nil
	}
	tag := *f.FPPSIdentificationTag
	data := []byte{byte(tag >> 8), byte(tag)}
	return buf.Write(data)
}

func (f *FlightPlanRelatedData) encodeCallsign(buf *bytes.Buffer) (int, error) {
	if f.Callsign == nil {
		return 0, nil
	}
	// Pad to 7 characters
	cs := (*f.Callsign + "       ")[:7]
	return buf.Write([]byte(cs))
}

func (f *FlightPlanRelatedData) encodeIFPSFlightID(buf *bytes.Buffer) (int, error) {
	if f.IFPSFlightID == nil {
		return 0, nil
	}
	return buf.Write(f.IFPSFlightID[:])
}

func (f *FlightPlanRelatedData) encodeFlightCategory(buf *bytes.Buffer) (int, error) {
	if f.FlightCategory == nil {
		return 0, nil
	}
	data := []byte{(*f.FlightCategory & 0x03) << 6}
	return buf.Write(data)
}

func (f *FlightPlanRelatedData) encodeTypeOfAircraft(buf *bytes.Buffer) (int, error) {
	if f.TypeOfAircraft == nil {
		return 0, nil
	}
	// Pad to 4 characters
	ac := (*f.TypeOfAircraft + "    ")[:4]
	return buf.Write([]byte(ac))
}

func (f *FlightPlanRelatedData) encodeWakeTurbulenceCategory(buf *bytes.Buffer) (int, error) {
	if f.WakeTurbulenceCategory == nil {
		return 0, nil
	}
	return buf.Write([]byte{*f.WakeTurbulenceCategory})
}

func (f *FlightPlanRelatedData) encodeDepartureAirport(buf *bytes.Buffer) (int, error) {
	if f.DepartureAirport == nil {
		return 0, nil
	}
	// Pad to 4 characters
	ap := (*f.DepartureAirport + "    ")[:4]
	return buf.Write([]byte(ap))
}

func (f *FlightPlanRelatedData) encodeDestinationAirport(buf *bytes.Buffer) (int, error) {
	if f.DestinationAirport == nil {
		return 0, nil
	}
	// Pad to 4 characters
	ap := (*f.DestinationAirport + "    ")[:4]
	return buf.Write([]byte(ap))
}

func (f *FlightPlanRelatedData) encodeRunwayDesignation(buf *bytes.Buffer) (int, error) {
	if f.RunwayDesignation == nil {
		return 0, nil
	}
	// Pad to 3 characters
	rwy := (*f.RunwayDesignation + "   ")[:3]
	return buf.Write([]byte(rwy))
}

func (f *FlightPlanRelatedData) encodeCurrentClearedFlightLevel(buf *bytes.Buffer) (int, error) {
	if f.CurrentClearedFlightLevel == nil {
		return 0, nil
	}
	fl := *f.CurrentClearedFlightLevel
	data := []byte{byte(fl >> 8), byte(fl)}
	return buf.Write(data)
}

func (f *FlightPlanRelatedData) encodeCurrentControlPosition(buf *bytes.Buffer) (int, error) {
	if f.CurrentControlPositionCentre == nil && f.CurrentControlPositionPos == nil {
		return 0, nil
	}
	centre := byte(0)
	pos := byte(0)
	if f.CurrentControlPositionCentre != nil {
		centre = *f.CurrentControlPositionCentre
	}
	if f.CurrentControlPositionPos != nil {
		pos = *f.CurrentControlPositionPos
	}
	return buf.Write([]byte{centre, pos})
}

func (f *FlightPlanRelatedData) encodeTimeOfDepartureArrival(buf *bytes.Buffer) (int, error) {
	if len(f.TimeOfDepartureArrival) == 0 {
		return 0, nil
	}
	// Write repetition factor
	n, err := buf.Write([]byte{byte(len(f.TimeOfDepartureArrival))})
	if err != nil {
		return n, err
	}
	bytesWritten := n

	// Write each entry
	for i := range f.TimeOfDepartureArrival {
		entry := &f.TimeOfDepartureArrival[i]
		var data [4]byte
		data[0] = (entry.TYP << 3) | (entry.DAY << 1) | ((entry.HOR >> 3) & 0x01)
		data[1] = (entry.HOR << 5) | (entry.MIN << 2)
		if entry.AVS {
			data[1] |= 0x02
		}
		data[2] = byte(entry.SEC >> 8)
		data[3] = byte(entry.SEC)

		n, err := buf.Write(data[:])
		bytesWritten += n
		if err != nil {
			return bytesWritten, err
		}
	}

	return bytesWritten, nil
}

func (f *FlightPlanRelatedData) encodeAircraftStand(buf *bytes.Buffer) (int, error) {
	if f.AircraftStand == nil {
		return 0, nil
	}
	// Pad to 6 characters
	stand := (*f.AircraftStand + "      ")[:6]
	return buf.Write([]byte(stand))
}

func (f *FlightPlanRelatedData) encodeStandStatus(buf *bytes.Buffer) (int, error) {
	if f.StandStatusEMP == nil && f.StandStatusAVL == nil && f.StandStatusOCC == nil {
		return 0, nil
	}
	var b byte
	if f.StandStatusEMP != nil && *f.StandStatusEMP {
		b |= 0x80
	}
	if f.StandStatusAVL != nil && *f.StandStatusAVL {
		b |= 0x40
	}
	if f.StandStatusOCC != nil && *f.StandStatusOCC {
		b |= 0x20
	}
	return buf.Write([]byte{b})
}

func (f *FlightPlanRelatedData) encodeStandardInstrumentDeparture(buf *bytes.Buffer) (int, error) {
	if f.StandardInstrumentDeparture == nil {
		return 0, nil
	}
	// Pad to 7 characters
	sid := (*f.StandardInstrumentDeparture + "       ")[:7]
	return buf.Write([]byte(sid))
}

func (f *FlightPlanRelatedData) encodeStandardInstrumentArrival(buf *bytes.Buffer) (int, error) {
	if f.StandardInstrumentArrival == nil {
		return 0, nil
	}
	// Pad to 7 characters
	star := (*f.StandardInstrumentArrival + "       ")[:7]
	return buf.Write([]byte(star))
}

func (f *FlightPlanRelatedData) encodePreEmergencyMode3A(buf *bytes.Buffer) (int, error) {
	if f.PreEmergencyMode3A == nil {
		return 0, nil
	}
	code := *f.PreEmergencyMode3A
	data := []byte{byte(code >> 8), byte(code)}
	return buf.Write(data)
}

func (f *FlightPlanRelatedData) encodePreEmergencyCallsign(buf *bytes.Buffer) (int, error) {
	if f.PreEmergencyCallsign == nil {
		return 0, nil
	}
	// Pad to 7 characters
	cs := (*f.PreEmergencyCallsign + "       ")[:7]
	return buf.Write([]byte(cs))
}

func (f *FlightPlanRelatedData) String() string {
	var parts []string

	if f.Callsign != nil {
		parts = append(parts, fmt.Sprintf("Callsign=%s", *f.Callsign))
	}
	if f.TypeOfAircraft != nil {
		parts = append(parts, fmt.Sprintf("Type=%s", *f.TypeOfAircraft))
	}
	if f.DepartureAirport != nil {
		parts = append(parts, fmt.Sprintf("Dep=%s", *f.DepartureAirport))
	}
	if f.DestinationAirport != nil {
		parts = append(parts, fmt.Sprintf("Dest=%s", *f.DestinationAirport))
	}
	if f.CurrentClearedFlightLevel != nil {
		parts = append(parts, fmt.Sprintf("FL=%.2f", float64(*f.CurrentClearedFlightLevel)*0.25))
	}
	if f.RunwayDesignation != nil {
		parts = append(parts, fmt.Sprintf("Rwy=%s", *f.RunwayDesignation))
	}
	if f.AircraftStand != nil {
		parts = append(parts, fmt.Sprintf("Stand=%s", *f.AircraftStand))
	}
	if f.StandardInstrumentDeparture != nil {
		parts = append(parts, fmt.Sprintf("SID=%s", *f.StandardInstrumentDeparture))
	}
	if f.StandardInstrumentArrival != nil {
		parts = append(parts, fmt.Sprintf("STAR=%s", *f.StandardInstrumentArrival))
	}

	if len(parts) == 0 {
		return "FlightPlanRelatedData{}"
	}

	return fmt.Sprintf("FlightPlanRelatedData{%s}", strings.Join(parts, ", "))
}

func (f *FlightPlanRelatedData) Validate() error {
	if f.FlightCategory != nil {
		if *f.FlightCategory > 3 {
			return fmt.Errorf("flight category out of range [0,3]: %d", *f.FlightCategory)
		}
	}

	if f.WakeTurbulenceCategory != nil {
		// Should be 'L', 'M', 'H', or 'J'
		wtc := *f.WakeTurbulenceCategory
		if wtc != 'L' && wtc != 'M' && wtc != 'H' && wtc != 'J' {
			return fmt.Errorf("invalid wake turbulence category: %c", wtc)
		}
	}

	return nil
}
