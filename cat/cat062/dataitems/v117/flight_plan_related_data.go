// dataitems/cat062/flight_plan_related_data.go
package v117

import (
	"bytes"
	"fmt"
	"strings"
)

// FlightPlanRelatedData implements I062/390
// Contains all flight plan related information provided by ground-based systems
type FlightPlanRelatedData struct {
	// Subfield #1: FPPS Identification Tag
	FPPSSAC *uint8 // System Area Code of FPPS
	FPPSSIC *uint8 // System Identity Code of FPPS

	// Subfield #2: Callsign
	Callsign *string // Aircraft callsign (up to 7 chars)

	// Subfield #3: IFPS_FLIGHT_ID
	IFPSFlightIDType *uint8  // Type of flight plan number
	IFPSFlightIDNum  *uint32 // Flight plan number (0-99999999)

	// Subfield #4: Flight Category
	FlightCategory     *uint8 // GAT/OAT (bits 8-7)
	FlightRules        *uint8 // IFR/VFR (bits 6-5)
	RVSM               *uint8 // RVSM status (bits 4-3)
	HighPriorityFlight bool   // Whether this is a high priority flight

	// Subfield #5: Type of Aircraft
	TypeOfAircraft *string // ICAO aircraft type designator (up to 4 chars)

	// Subfield #6: Wake Turbulence Category
	WakeTurbulenceCategory *byte // 'L' (Light), 'M' (Medium), 'H' (Heavy), 'J' (Super)

	// Subfield #7: Departure Airport
	DepartureAirport *string // ICAO airport code (4 chars)

	// Subfield #8: Destination Airport
	DestinationAirport *string // ICAO airport code (4 chars)

	// Subfield #9: Runway Designation
	RunwayNumber1 *uint8 // First number (ASCII)
	RunwayNumber2 *uint8 // Second number (ASCII)
	RunwayLetter  *uint8 // Letter (ASCII)

	// Subfield #10: Current Cleared Flight Level
	ClearedFlightLevel *float64 // In flight levels (FL)

	// Subfield #11: Current Control Position
	ControlCentre   *uint8 // Center identification code
	ControlPosition *uint8 // Position identification code

	// Subfield #12: Time of Departure / Arrival
	TimeTypeList    []uint8 // Type of time (see spec)
	DayList         []uint8 // 0=today, 1=yesterday, 2=tomorrow, 3=invalid
	HourList        []uint8 // Hour (0-23)
	MinuteList      []uint8 // Minute (0-59)
	SecondList      []uint8 // Second (0-59) if available
	SecondAvailList []bool  // Whether seconds are available

	// Subfield #13: Aircraft Stand
	AircraftStand *string // Aircraft stand identification (up to 6 chars)

	// Subfield #14: Stand Status
	StandEmpty     *uint8 // Empty/Occupied/Unknown (bits 8-7)
	StandAvailable *uint8 // Available/Not available/Unknown (bits 6-5)

	// Subfield #15: Standard Instrument Departure
	SID *string // SID identifier (up to 7 chars)

	// Subfield #16: Standard Instrument Arrival
	STAR *string // STAR identifier (up to 7 chars)

	// Subfield #17: Pre-Emergency Mode 3/A
	PreEmergencyMode3A      *uint16 // Mode 3/A code before emergency
	PreEmergencyMode3AValid bool    // Whether the code is valid

	// Subfield #18: Pre-Emergency Callsign
	PreEmergencyCallsign *string // Callsign before emergency (up to 7 chars)
}

// Decode parses an ASTERIX Category 062 I390 data item from the buffer
func (f *FlightPlanRelatedData) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Read the primary subfield (FSPEC)
	if buf.Len() < 1 {
		return 0, fmt.Errorf("buffer too short for flight plan FSPEC")
	}

	fspec1, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading flight plan FSPEC: %w", err)
	}
	bytesRead++

	// Check for the first FX extension bit
	hasSecondFSPEC := (fspec1 & 0x01) != 0
	var fspec2 byte
	if hasSecondFSPEC {
		if buf.Len() < 1 {
			return bytesRead, fmt.Errorf("buffer too short for second FSPEC byte")
		}
		fspec2, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading second FSPEC byte: %w", err)
		}
		bytesRead++
	}

	// Check for the second FX extension bit
	hasThirdFSPEC := hasSecondFSPEC && (fspec2&0x01) != 0
	var fspec3 byte
	if hasThirdFSPEC {
		if buf.Len() < 1 {
			return bytesRead, fmt.Errorf("buffer too short for third FSPEC byte")
		}
		fspec3, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading third FSPEC byte: %w", err)
		}
		bytesRead++
	}

	// Process first FSPEC byte
	// Subfield #1: FPPS Identification Tag
	if (fspec1 & 0x80) != 0 {
		if buf.Len() < 2 {
			return bytesRead, fmt.Errorf("buffer too short for FPPS Identification Tag")
		}
		data := make([]byte, 2)
		n, err := buf.Read(data)
		if err != nil || n != 2 {
			return bytesRead + n, fmt.Errorf("reading FPPS Identification Tag: %w", err)
		}
		bytesRead += n

		sac := data[0]
		sic := data[1]
		f.FPPSSAC = &sac
		f.FPPSSIC = &sic
	}

	// Subfield #2: Callsign
	if (fspec1 & 0x40) != 0 {
		if buf.Len() < 7 {
			return bytesRead, fmt.Errorf("buffer too short for callsign")
		}
		data := make([]byte, 7)
		n, err := buf.Read(data)
		if err != nil || n != 7 {
			return bytesRead + n, fmt.Errorf("reading callsign: %w", err)
		}
		bytesRead += n

		// Each byte is an ASCII character
		// Trim trailing spaces
		callsign := strings.TrimRight(string(data), " ")
		f.Callsign = &callsign
	}

	// Subfield #3: IFPS_FLIGHT_ID
	if (fspec1 & 0x20) != 0 {
		if buf.Len() < 4 {
			return bytesRead, fmt.Errorf("buffer too short for IFPS flight ID")
		}
		data := make([]byte, 4)
		n, err := buf.Read(data)
		if err != nil || n != 4 {
			return bytesRead + n, fmt.Errorf("reading IFPS flight ID: %w", err)
		}
		bytesRead += n

		// Extract type (bits 32-31)
		typ := (data[0] >> 6) & 0x03
		f.IFPSFlightIDType = &typ

		// Extract flight ID number (bits 27-1)
		num := uint32(data[0]&0x0F)<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])
		f.IFPSFlightIDNum = &num
	}

	// Subfield #4: Flight Category
	if (fspec1 & 0x10) != 0 {
		if buf.Len() < 1 {
			return bytesRead, fmt.Errorf("buffer too short for flight category")
		}
		data, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading flight category: %w", err)
		}
		bytesRead++

		// Extract GAT/OAT (bits 8-7)
		gat := (data >> 6) & 0x03
		f.FlightCategory = &gat

		// Extract IFR/VFR (bits 6-5)
		ifr := (data >> 4) & 0x03
		f.FlightRules = &ifr

		// Extract RVSM (bits 4-3)
		rvsm := (data >> 2) & 0x03
		f.RVSM = &rvsm

		// Extract high priority (bit 2)
		f.HighPriorityFlight = (data & 0x02) != 0
		// Bit 1 is spare
	}

	// Subfield #5: Type of Aircraft
	if (fspec1 & 0x08) != 0 {
		if buf.Len() < 4 {
			return bytesRead, fmt.Errorf("buffer too short for type of aircraft")
		}
		data := make([]byte, 4)
		n, err := buf.Read(data)
		if err != nil || n != 4 {
			return bytesRead + n, fmt.Errorf("reading type of aircraft: %w", err)
		}
		bytesRead += n

		// Each byte is an ASCII character
		// Trim trailing spaces
		typeAircraft := strings.TrimRight(string(data), " ")
		f.TypeOfAircraft = &typeAircraft
	}

	// Subfield #6: Wake Turbulence Category
	if (fspec1 & 0x04) != 0 {
		if buf.Len() < 1 {
			return bytesRead, fmt.Errorf("buffer too short for wake turbulence category")
		}
		data, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading wake turbulence category: %w", err)
		}
		bytesRead++

		f.WakeTurbulenceCategory = &data
	}

	// Subfield #7: Departure Airport
	if (fspec1 & 0x02) != 0 {
		if buf.Len() < 4 {
			return bytesRead, fmt.Errorf("buffer too short for departure airport")
		}
		data := make([]byte, 4)
		n, err := buf.Read(data)
		if err != nil || n != 4 {
			return bytesRead + n, fmt.Errorf("reading departure airport: %w", err)
		}
		bytesRead += n

		// Each byte is an ASCII character
		depAirport := string(data)
		f.DepartureAirport = &depAirport
	}

	// Process second FSPEC byte if present
	if hasSecondFSPEC {
		// Subfield #8: Destination Airport
		if (fspec2 & 0x80) != 0 {
			if buf.Len() < 4 {
				return bytesRead, fmt.Errorf("buffer too short for destination airport")
			}
			data := make([]byte, 4)
			n, err := buf.Read(data)
			if err != nil || n != 4 {
				return bytesRead + n, fmt.Errorf("reading destination airport: %w", err)
			}
			bytesRead += n

			// Each byte is an ASCII character
			destAirport := string(data)
			f.DestinationAirport = &destAirport
		}

		// Subfield #9: Runway Designation
		if (fspec2 & 0x40) != 0 {
			if buf.Len() < 3 {
				return bytesRead, fmt.Errorf("buffer too short for runway designation")
			}
			data := make([]byte, 3)
			n, err := buf.Read(data)
			if err != nil || n != 3 {
				return bytesRead + n, fmt.Errorf("reading runway designation: %w", err)
			}
			bytesRead += n

			num1 := data[0]
			num2 := data[1]
			letter := data[2]
			f.RunwayNumber1 = &num1
			f.RunwayNumber2 = &num2
			f.RunwayLetter = &letter
		}

		// Subfield #10: Current Cleared Flight Level
		if (fspec2 & 0x20) != 0 {
			if buf.Len() < 2 {
				return bytesRead, fmt.Errorf("buffer too short for current cleared flight level")
			}
			data := make([]byte, 2)
			n, err := buf.Read(data)
			if err != nil || n != 2 {
				return bytesRead + n, fmt.Errorf("reading current cleared flight level: %w", err)
			}
			bytesRead += n

			flBits := uint16(data[0])<<8 | uint16(data[1])
			fl := float64(flBits) * 0.25 // LSB = 1/4 FL
			f.ClearedFlightLevel = &fl
		}

		// Subfield #11: Current Control Position
		if (fspec2 & 0x10) != 0 {
			if buf.Len() < 2 {
				return bytesRead, fmt.Errorf("buffer too short for current control position")
			}
			data := make([]byte, 2)
			n, err := buf.Read(data)
			if err != nil || n != 2 {
				return bytesRead + n, fmt.Errorf("reading current control position: %w", err)
			}
			bytesRead += n

			centre := data[0]
			position := data[1]
			f.ControlCentre = &centre
			f.ControlPosition = &position
		}

		// Subfield #12: Time of Departure / Arrival
		if (fspec2 & 0x08) != 0 {
			if buf.Len() < 1 {
				return bytesRead, fmt.Errorf("buffer too short for time of departure/arrival REP")
			}
			repByte, err := buf.ReadByte()
			if err != nil {
				return bytesRead, fmt.Errorf("reading time of departure/arrival REP: %w", err)
			}
			bytesRead++

			rep := int(repByte)
			if buf.Len() < rep*4 {
				return bytesRead, fmt.Errorf("buffer too short for time of departure/arrival data")
			}

			// Initialize our slice fields
			f.TimeTypeList = make([]uint8, rep)
			f.DayList = make([]uint8, rep)
			f.HourList = make([]uint8, rep)
			f.MinuteList = make([]uint8, rep)
			f.SecondList = make([]uint8, rep)
			f.SecondAvailList = make([]bool, rep)

			for i := 0; i < rep; i++ {
				// First byte: type and day
				data1, err := buf.ReadByte()
				if err != nil {
					return bytesRead, fmt.Errorf("reading time type and day: %w", err)
				}
				bytesRead++

				f.TimeTypeList[i] = (data1 >> 3) & 0x1F
				f.DayList[i] = (data1 >> 1) & 0x03
				// Bit 1 is spare

				// Second byte: hours
				data2, err := buf.ReadByte()
				if err != nil {
					return bytesRead, fmt.Errorf("reading hours: %w", err)
				}
				bytesRead++

				f.HourList[i] = (data2 >> 2) & 0x1F
				// Bits 2-1 are spare

				// Third byte: minutes
				data3, err := buf.ReadByte()
				if err != nil {
					return bytesRead, fmt.Errorf("reading minutes: %w", err)
				}
				bytesRead++

				f.MinuteList[i] = (data3 >> 1) & 0x3F
				// Bit 1 is spare

				// Fourth byte: seconds (if available)
				data4, err := buf.ReadByte()
				if err != nil {
					return bytesRead, fmt.Errorf("reading seconds: %w", err)
				}
				bytesRead++

				f.SecondAvailList[i] = (data4 & 0x80) == 0 // AVS bit (inverted)
				if f.SecondAvailList[i] {
					f.SecondList[i] = (data4 >> 1) & 0x3F
				} else {
					f.SecondList[i] = 0
				}
				// Bit 1 is spare
			}
		}

		// Subfield #13: Aircraft Stand
		if (fspec2 & 0x04) != 0 {
			if buf.Len() < 6 {
				return bytesRead, fmt.Errorf("buffer too short for aircraft stand")
			}
			data := make([]byte, 6)
			n, err := buf.Read(data)
			if err != nil || n != 6 {
				return bytesRead + n, fmt.Errorf("reading aircraft stand: %w", err)
			}
			bytesRead += n

			// Each byte is an ASCII character
			// Trim trailing spaces
			stand := strings.TrimRight(string(data), " ")
			f.AircraftStand = &stand
		}

		// Subfield #14: Stand Status
		if (fspec2 & 0x02) != 0 {
			if buf.Len() < 1 {
				return bytesRead, fmt.Errorf("buffer too short for stand status")
			}
			data, err := buf.ReadByte()
			if err != nil {
				return bytesRead, fmt.Errorf("reading stand status: %w", err)
			}
			bytesRead++

			// Extract Empty/Occupied (bits 8-7)
			empty := (data >> 6) & 0x03
			f.StandEmpty = &empty

			// Extract Available/Not available (bits 6-5)
			avail := (data >> 4) & 0x03
			f.StandAvailable = &avail
			// Bits 4-1 are spare
		}
	}

	// Process third FSPEC byte if present
	if hasThirdFSPEC {
		// Subfield #15: Standard Instrument Departure
		if (fspec3 & 0x80) != 0 {
			if buf.Len() < 7 {
				return bytesRead, fmt.Errorf("buffer too short for SID")
			}
			data := make([]byte, 7)
			n, err := buf.Read(data)
			if err != nil || n != 7 {
				return bytesRead + n, fmt.Errorf("reading SID: %w", err)
			}
			bytesRead += n

			// Each byte is an ASCII character
			// Trim trailing spaces
			sid := strings.TrimRight(string(data), " ")
			f.SID = &sid
		}

		// Subfield #16: Standard Instrument Arrival
		if (fspec3 & 0x40) != 0 {
			if buf.Len() < 7 {
				return bytesRead, fmt.Errorf("buffer too short for STAR")
			}
			data := make([]byte, 7)
			n, err := buf.Read(data)
			if err != nil || n != 7 {
				return bytesRead + n, fmt.Errorf("reading STAR: %w", err)
			}
			bytesRead += n

			// Each byte is an ASCII character
			// Trim trailing spaces
			star := strings.TrimRight(string(data), " ")
			f.STAR = &star
		}

		// Subfield #17: Pre-Emergency Mode 3/A
		if (fspec3 & 0x20) != 0 {
			if buf.Len() < 2 {
				return bytesRead, fmt.Errorf("buffer too short for pre-emergency Mode 3/A")
			}
			data := make([]byte, 2)
			n, err := buf.Read(data)
			if err != nil || n != 2 {
				return bytesRead + n, fmt.Errorf("reading pre-emergency Mode 3/A: %w", err)
			}
			bytesRead += n

			// Check validity flag (bit 13)
			f.PreEmergencyMode3AValid = (data[0] & 0x08) != 0

			if f.PreEmergencyMode3AValid {
				// Extract Mode 3/A code (12 bits)
				mode3A := uint16(data[0]&0x07)<<8 | uint16(data[1])
				f.PreEmergencyMode3A = &mode3A
			}
		}

		// Subfield #18: Pre-Emergency Callsign
		if (fspec3 & 0x10) != 0 {
			if buf.Len() < 7 {
				return bytesRead, fmt.Errorf("buffer too short for pre-emergency callsign")
			}
			data := make([]byte, 7)
			n, err := buf.Read(data)
			if err != nil || n != 7 {
				return bytesRead + n, fmt.Errorf("reading pre-emergency callsign: %w", err)
			}
			bytesRead += n

			// Each byte is an ASCII character
			// Trim trailing spaces
			callsign := strings.TrimRight(string(data), " ")
			f.PreEmergencyCallsign = &callsign
		}
	}

	return bytesRead, nil
}

// Encode serializes the flight plan related data into the buffer
func (f *FlightPlanRelatedData) Encode(buf *bytes.Buffer) (int, error) {
	bytesWritten := 0

	// Determine which subfields are present
	hasFPPS := f.FPPSSAC != nil && f.FPPSSIC != nil
	hasCallsign := f.Callsign != nil
	hasIFPS := f.IFPSFlightIDType != nil && f.IFPSFlightIDNum != nil
	hasFlightCategory := f.FlightCategory != nil && f.FlightRules != nil && f.RVSM != nil
	hasTypeAircraft := f.TypeOfAircraft != nil
	hasWTC := f.WakeTurbulenceCategory != nil
	hasDeparture := f.DepartureAirport != nil

	hasDestination := f.DestinationAirport != nil
	hasRunway := f.RunwayNumber1 != nil && f.RunwayNumber2 != nil && f.RunwayLetter != nil
	hasCFL := f.ClearedFlightLevel != nil
	hasControl := f.ControlCentre != nil && f.ControlPosition != nil
	hasTime := f.TimeTypeList != nil && len(f.TimeTypeList) > 0
	hasStand := f.AircraftStand != nil
	hasStandStatus := f.StandEmpty != nil && f.StandAvailable != nil

	hasSID := f.SID != nil
	hasSTAR := f.STAR != nil
	hasPreMode3A := f.PreEmergencyMode3A != nil
	hasPreCallsign := f.PreEmergencyCallsign != nil

	// Need second FSPEC byte?
	needSecondByte := hasDestination || hasRunway || hasCFL || hasControl || hasTime || hasStand || hasStandStatus

	// Need third FSPEC byte?
	needThirdByte := hasSID || hasSTAR || hasPreMode3A || hasPreCallsign

	// First FSPEC byte
	fspec1 := byte(0)
	if hasFPPS {
		fspec1 |= 0x80 // Bit 8: FPPS Identification Tag
	}
	if hasCallsign {
		fspec1 |= 0x40 // Bit 7: Callsign
	}
	if hasIFPS {
		fspec1 |= 0x20 // Bit 6: IFPS_FLIGHT_ID
	}
	if hasFlightCategory {
		fspec1 |= 0x10 // Bit 5: Flight Category
	}
	if hasTypeAircraft {
		fspec1 |= 0x08 // Bit 4: Type of Aircraft
	}
	if hasWTC {
		fspec1 |= 0x04 // Bit 3: Wake Turbulence Category
	}
	if hasDeparture {
		fspec1 |= 0x02 // Bit 2: Departure Airport
	}
	if needSecondByte {
		fspec1 |= 0x01 // Bit 1: FX
	}

	// Write first FSPEC byte
	err := buf.WriteByte(fspec1)
	if err != nil {
		return 0, fmt.Errorf("writing first FSPEC byte: %w", err)
	}
	bytesWritten++

	// Second FSPEC byte if needed
	if needSecondByte {
		fspec2 := byte(0)
		if hasDestination {
			fspec2 |= 0x80 // Bit 8: Destination Airport
		}
		if hasRunway {
			fspec2 |= 0x40 // Bit 7: Runway Designation
		}
		if hasCFL {
			fspec2 |= 0x20 // Bit 6: Current Cleared Flight Level
		}
		if hasControl {
			fspec2 |= 0x10 // Bit 5: Current Control Position
		}
		if hasTime {
			fspec2 |= 0x08 // Bit 4: Time of Departure / Arrival
		}
		if hasStand {
			fspec2 |= 0x04 // Bit 3: Aircraft Stand
		}
		if hasStandStatus {
			fspec2 |= 0x02 // Bit 2: Stand Status
		}
		if needThirdByte {
			fspec2 |= 0x01 // Bit 1: FX
		}

		// Write second FSPEC byte
		err := buf.WriteByte(fspec2)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing second FSPEC byte: %w", err)
		}
		bytesWritten++
	}

	// Third FSPEC byte if needed
	if needThirdByte {
		fspec3 := byte(0)
		if hasSID {
			fspec3 |= 0x80 // Bit 8: Standard Instrument Departure
		}
		if hasSTAR {
			fspec3 |= 0x40 // Bit 7: Standard Instrument Arrival
		}
		if hasPreMode3A {
			fspec3 |= 0x20 // Bit 6: Pre-Emergency Mode 3/A
		}
		if hasPreCallsign {
			fspec3 |= 0x10 // Bit 5: Pre-Emergency Callsign
		}
		// Bits 4-2 are spare
		// Bit 1 (FX) = 0, no extension

		// Write third FSPEC byte
		err := buf.WriteByte(fspec3)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing third FSPEC byte: %w", err)
		}
		bytesWritten++
	}

	// Write Subfield #1: FPPS Identification Tag
	if hasFPPS {
		data := []byte{*f.FPPSSAC, *f.FPPSSIC}
		n, err := buf.Write(data)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing FPPS identification tag: %w", err)
		}
		bytesWritten += n
	}

	// Write Subfield #2: Callsign
	if hasCallsign {
		// Pad with spaces if needed
		callsign := *f.Callsign
		if len(callsign) > 7 {
			callsign = callsign[:7]
		} else if len(callsign) < 7 {
			callsign = callsign + strings.Repeat(" ", 7-len(callsign))
		}

		n, err := buf.WriteString(callsign)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing callsign: %w", err)
		}
		bytesWritten += n
	}

	// Write Subfield #3: IFPS_FLIGHT_ID
	if hasIFPS {
		// Prepare first byte with type and high bits of number
		firstByte := ((*f.IFPSFlightIDType & 0x03) << 6) | byte((*f.IFPSFlightIDNum>>24)&0x0F)

		// Prepare remaining bytes
		data := []byte{
			firstByte,
			byte((*f.IFPSFlightIDNum >> 16) & 0xFF),
			byte((*f.IFPSFlightIDNum >> 8) & 0xFF),
			byte(*f.IFPSFlightIDNum & 0xFF),
		}
		n, err := buf.Write(data)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing IFPS flight ID: %w", err)
		}
		bytesWritten += n
	}

	// Write Subfield #4: Flight Category
	if hasFlightCategory {
		// Combine all bits into a single byte
		data := byte(0)
		data |= (*f.FlightCategory & 0x03) << 6 // GAT/OAT
		data |= (*f.FlightRules & 0x03) << 4    // IFR/VFR
		data |= (*f.RVSM & 0x03) << 2           // RVSM
		if f.HighPriorityFlight {
			data |= 0x02 // HPR
		}
		// Bit 1 is spare

		err := buf.WriteByte(data)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing flight category: %w", err)
		}
		bytesWritten++
	}

	// Write Subfield #5: Type of Aircraft
	if hasTypeAircraft {
		// Pad with spaces if needed
		typeAircraft := *f.TypeOfAircraft
		if len(typeAircraft) > 4 {
			typeAircraft = typeAircraft[:4]
		} else if len(typeAircraft) < 4 {
			typeAircraft = typeAircraft + strings.Repeat(" ", 4-len(typeAircraft))
		}

		n, err := buf.WriteString(typeAircraft)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing type of aircraft: %w", err)
		}
		bytesWritten += n
	}

	// Write Subfield #6: Wake Turbulence Category
	if hasWTC {
		err := buf.WriteByte(*f.WakeTurbulenceCategory)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing wake turbulence category: %w", err)
		}
		bytesWritten++
	}

	// Write Subfield #7: Departure Airport
	if hasDeparture {
		// Ensure exactly 4 characters
		depAirport := *f.DepartureAirport
		if len(depAirport) > 4 {
			depAirport = depAirport[:4]
		} else if len(depAirport) < 4 {
			depAirport = depAirport + strings.Repeat(" ", 4-len(depAirport))
		}

		n, err := buf.WriteString(depAirport)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing departure airport: %w", err)
		}
		bytesWritten += n
	}

	// Write Subfield #8: Destination Airport
	if hasDestination {
		// Ensure exactly 4 characters
		destAirport := *f.DestinationAirport
		if len(destAirport) > 4 {
			destAirport = destAirport[:4]
		} else if len(destAirport) < 4 {
			destAirport = destAirport + strings.Repeat(" ", 4-len(destAirport))
		}

		n, err := buf.WriteString(destAirport)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing destination airport: %w", err)
		}
		bytesWritten += n
	}

	// Write Subfield #9: Runway Designation
	if hasRunway {
		data := []byte{*f.RunwayNumber1, *f.RunwayNumber2, *f.RunwayLetter}
		n, err := buf.Write(data)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing runway designation: %w", err)
		}
		bytesWritten += n
	}

	// Write Subfield #10: Current Cleared Flight Level
	if hasCFL {
		// Convert to binary (1/4 FL resolution)
		flBits := uint16(*f.ClearedFlightLevel * 4)
		data := []byte{byte(flBits >> 8), byte(flBits)}
		n, err := buf.Write(data)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing current cleared flight level: %w", err)
		}
		bytesWritten += n
	}

	// Write Subfield #11: Current Control Position
	if hasControl {
		data := []byte{*f.ControlCentre, *f.ControlPosition}
		n, err := buf.Write(data)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing current control position: %w", err)
		}
		bytesWritten += n
	}

	// Write Subfield #12: Time of Departure / Arrival
	if hasTime {
		// Write repetition factor
		rep := uint8(len(f.TimeTypeList))
		err := buf.WriteByte(rep)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing time repetition factor: %w", err)
		}
		bytesWritten++

		// Write each time entry
		for i := 0; i < int(rep); i++ {
			// First byte: type and day
			data1 := (f.TimeTypeList[i] & 0x1F) << 3
			data1 |= (f.DayList[i] & 0x03) << 1
			// Bit 1 is spare

			// Second byte: hours
			data2 := (f.HourList[i] & 0x1F) << 2
			// Bits 2-1 are spare

			// Third byte: minutes
			data3 := (f.MinuteList[i] & 0x3F) << 1
			// Bit 1 is spare

			// Fourth byte: seconds (if available)
			data4 := byte(0)
			if f.SecondAvailList[i] {
				// Bit 8 (AVS) = 0 if seconds available
				data4 |= (f.SecondList[i] & 0x3F) << 1
			} else {
				// Bit 8 (AVS) = 1 if seconds not available
				data4 |= 0x80
			}
			// Bit 1 is spare

			data := []byte{data1, data2, data3, data4}
			n, err := buf.Write(data)
			if err != nil {
				return bytesWritten, fmt.Errorf("writing time data: %w", err)
			}
			bytesWritten += n
		}
	}

	// Write Subfield #13: Aircraft Stand
	if hasStand {
		// Pad with spaces if needed
		stand := *f.AircraftStand
		if len(stand) > 6 {
			stand = stand[:6]
		} else if len(stand) < 6 {
			stand = stand + strings.Repeat(" ", 6-len(stand))
		}

		n, err := buf.WriteString(stand)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing aircraft stand: %w", err)
		}
		bytesWritten += n
	}

	// Write Subfield #14: Stand Status
	if hasStandStatus {
		// Combine bits into a single byte
		data := byte(0)
		data |= (*f.StandEmpty & 0x03) << 6     // Empty/Occupied
		data |= (*f.StandAvailable & 0x03) << 4 // Available/Not available
		// Bits 4-1 are spare

		err := buf.WriteByte(data)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing stand status: %w", err)
		}
		bytesWritten++
	}

	// Write Subfield #15: Standard Instrument Departure
	if hasSID {
		// Pad with spaces if needed
		sid := *f.SID
		if len(sid) > 7 {
			sid = sid[:7]
		} else if len(sid) < 7 {
			sid = sid + strings.Repeat(" ", 7-len(sid))
		}

		n, err := buf.WriteString(sid)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing SID: %w", err)
		}
		bytesWritten += n
	}

	// Write Subfield #16: Standard Instrument Arrival
	if hasSTAR {
		// Pad with spaces if needed
		star := *f.STAR
		if len(star) > 7 {
			star = star[:7]
		} else if len(star) < 7 {
			star = star + strings.Repeat(" ", 7-len(star))
		}

		n, err := buf.WriteString(star)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing STAR: %w", err)
		}
		bytesWritten += n
	}

	// Write Subfield #17: Pre-Emergency Mode 3/A
	if hasPreMode3A {
		// First byte contains validity flag (bit 4) and high bits of code
		data1 := byte(0)
		if f.PreEmergencyMode3AValid {
			data1 |= 0x08 // Validity bit (VA)
		}
		data1 |= byte((*f.PreEmergencyMode3A >> 8) & 0x07) // High bits of code

		// Second byte contains low bits of code
		data2 := byte(*f.PreEmergencyMode3A & 0xFF)

		data := []byte{data1, data2}
		n, err := buf.Write(data)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing pre-emergency Mode 3/A: %w", err)
		}
		bytesWritten += n
	}

	// Write Subfield #18: Pre-Emergency Callsign
	if hasPreCallsign {
		// Pad with spaces if needed
		callsign := *f.PreEmergencyCallsign
		if len(callsign) > 7 {
			callsign = callsign[:7]
		} else if len(callsign) < 7 {
			callsign = callsign + strings.Repeat(" ", 7-len(callsign))
		}

		n, err := buf.WriteString(callsign)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing pre-emergency callsign: %w", err)
		}
		bytesWritten += n
	}

	return bytesWritten, nil
}

// String returns a human-readable representation of the flight plan related data
func (f *FlightPlanRelatedData) String() string {
	parts := []string{}

	if f.Callsign != nil {
		parts = append(parts, fmt.Sprintf("C/S: %s", *f.Callsign))
	}

	if f.TypeOfAircraft != nil {
		parts = append(parts, fmt.Sprintf("Type: %s", *f.TypeOfAircraft))
	}

	if f.WakeTurbulenceCategory != nil {
		wtc := string(*f.WakeTurbulenceCategory)
		parts = append(parts, fmt.Sprintf("WTC: %s", wtc))
	}

	if f.DepartureAirport != nil {
		parts = append(parts, fmt.Sprintf("Dep: %s", *f.DepartureAirport))
	}

	if f.DestinationAirport != nil {
		parts = append(parts, fmt.Sprintf("Dest: %s", *f.DestinationAirport))
	}

	if f.ClearedFlightLevel != nil {
		parts = append(parts, fmt.Sprintf("CFL: FL%.0f", *f.ClearedFlightLevel))
	}

	if f.SID != nil {
		parts = append(parts, fmt.Sprintf("SID: %s", *f.SID))
	}

	if f.STAR != nil {
		parts = append(parts, fmt.Sprintf("STAR: %s", *f.STAR))
	}

	// Add flight category if present
	if f.FlightCategory != nil && f.FlightRules != nil && f.RVSM != nil {
		// Map GAT/OAT values
		gatOat := []string{"Unknown", "GAT", "OAT", "N/A"}
		gatOatStr := "Unknown"
		if int(*f.FlightCategory) < len(gatOat) {
			gatOatStr = gatOat[*f.FlightCategory]
		}

		// Map IFR/VFR values
		ifrVfr := []string{"IFR", "VFR", "N/A", "CVFR"}
		ifrVfrStr := "Unknown"
		if int(*f.FlightRules) < len(ifrVfr) {
			ifrVfrStr = ifrVfr[*f.FlightRules]
		}

		// Map RVSM values
		rvsm := []string{"Unknown", "Approved", "Exempt", "Not Approved"}
		rvsmStr := "Unknown"
		if int(*f.RVSM) < len(rvsm) {
			rvsmStr = rvsm[*f.RVSM]
		}

		flightCat := fmt.Sprintf("%s/%s/RVSM-%s", gatOatStr, ifrVfrStr, rvsmStr)
		if f.HighPriorityFlight {
			flightCat += "/Priority"
		}
		parts = append(parts, flightCat)
	}

	// Include times if available
	if f.TimeTypeList != nil && len(f.TimeTypeList) > 0 {
		for i := 0; i < len(f.TimeTypeList); i++ {
			// Map time type to string
			timeTypes := []string{
				"SOBT", "EOBT", "ETOT", "AOBT", "PTRY", "ATRY", "ALUT", "ATOT",
				"ETA", "PLT", "ALDT", "AORT", "PTG", "AOBT",
			}

			typeStr := "Unknown"
			if int(f.TimeTypeList[i]) < len(timeTypes) {
				typeStr = timeTypes[f.TimeTypeList[i]]
			}

			// Format time
			var timeStr string
			if f.SecondAvailList[i] {
				timeStr = fmt.Sprintf("%02d:%02d:%02d", f.HourList[i], f.MinuteList[i], f.SecondList[i])
			} else {
				timeStr = fmt.Sprintf("%02d:%02d", f.HourList[i], f.MinuteList[i])
			}

			// Add day indication if not today
			switch f.DayList[i] {
			case 1:
				timeStr += " (-1)" // yesterday
			case 2:
				timeStr += " (+1)" // tomorrow
			}

			parts = append(parts, fmt.Sprintf("%s: %s", typeStr, timeStr))
		}
	}

	if len(parts) == 0 {
		return "FPD[empty]"
	}

	return fmt.Sprintf("FPD[%s]", strings.Join(parts, ", "))
}

// Validate performs validation on the flight plan related data
func (f *FlightPlanRelatedData) Validate() error {
	// Check callsign length
	if f.Callsign != nil && len(*f.Callsign) > 7 {
		return fmt.Errorf("callsign too long (max 7 chars): %s", *f.Callsign)
	}

	// Check wake turbulence category
	if f.WakeTurbulenceCategory != nil {
		wtc := *f.WakeTurbulenceCategory
		if wtc != 'L' && wtc != 'M' && wtc != 'H' && wtc != 'J' {
			return fmt.Errorf("invalid wake turbulence category (should be L, M, H, or J): %c", wtc)
		}
	}

	// Check airport codes (should be 4 characters)
	if f.DepartureAirport != nil && len(*f.DepartureAirport) != 4 {
		return fmt.Errorf("departure airport code should be exactly 4 characters: %s", *f.DepartureAirport)
	}

	if f.DestinationAirport != nil && len(*f.DestinationAirport) != 4 {
		return fmt.Errorf("destination airport code should be exactly 4 characters: %s", *f.DestinationAirport)
	}

	// Check runway numbers (should be ASCII digits)
	if f.RunwayNumber1 != nil && (*f.RunwayNumber1 < '0' || *f.RunwayNumber1 > '9') {
		return fmt.Errorf("runway number 1 should be an ASCII digit: %c", *f.RunwayNumber1)
	}

	if f.RunwayNumber2 != nil && (*f.RunwayNumber2 < '0' || *f.RunwayNumber2 > '9') {
		return fmt.Errorf("runway number 2 should be an ASCII digit: %c", *f.RunwayNumber2)
	}

	// Check time list lengths match
	if f.TimeTypeList != nil {
		timeLen := len(f.TimeTypeList)

		if f.DayList != nil && len(f.DayList) != timeLen {
			return fmt.Errorf("day list length (%d) doesn't match time type list length (%d)",
				len(f.DayList), timeLen)
		}

		if f.HourList != nil && len(f.HourList) != timeLen {
			return fmt.Errorf("hour list length (%d) doesn't match time type list length (%d)",
				len(f.HourList), timeLen)
		}

		if f.MinuteList != nil && len(f.MinuteList) != timeLen {
			return fmt.Errorf("minute list length (%d) doesn't match time type list length (%d)",
				len(f.MinuteList), timeLen)
		}

		if f.SecondList != nil && len(f.SecondList) != timeLen {
			return fmt.Errorf("second list length (%d) doesn't match time type list length (%d)",
				len(f.SecondList), timeLen)
		}

		if f.SecondAvailList != nil && len(f.SecondAvailList) != timeLen {
			return fmt.Errorf("second availability list length (%d) doesn't match time type list length (%d)",
				len(f.SecondAvailList), timeLen)
		}

		// Validate time values
		for i := 0; i < timeLen; i++ {
			if f.HourList[i] > 23 {
				return fmt.Errorf("invalid hour value at index %d: %d (max 23)", i, f.HourList[i])
			}

			if f.MinuteList[i] > 59 {
				return fmt.Errorf("invalid minute value at index %d: %d (max 59)", i, f.MinuteList[i])
			}

			if f.SecondAvailList[i] && f.SecondList[i] > 59 {
				return fmt.Errorf("invalid second value at index %d: %d (max 59)", i, f.SecondList[i])
			}
		}
	}

	return nil
}

// GetFormattedTimeType returns a human-readable string for a time type
func GetFormattedTimeType(timeType uint8) string {
	timeTypes := map[uint8]string{
		0:  "Scheduled Off-Block Time",
		1:  "Estimated Off-Block Time",
		2:  "Estimated Take-Off Time",
		3:  "Actual Off-Block Time",
		4:  "Predicted Time at Runway Hold",
		5:  "Actual Time at Runway Hold",
		6:  "Actual Line-Up Time",
		7:  "Actual Take-Off Time",
		8:  "Estimated Time of Arrival",
		9:  "Predicted Landing Time",
		10: "Actual Landing Time",
		11: "Actual Time Off Runway",
		12: "Predicted Time to Gate",
		13: "Actual On-Block Time",
	}

	if name, ok := timeTypes[timeType]; ok {
		return name
	}
	return fmt.Sprintf("Unknown Time Type (%d)", timeType)
}
