package v13

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

// FlightPlanRelatedData implements I011/390 - Flight Plan Related Data
// Definition: All flight plan related information.
// Format: Compound Data Item, comprising a primary subfield of two octets,
// followed by up to fourteen subfields.
type FlightPlanRelatedData struct {
	// Presence flags
	HasTAG bool // FPPS Identification Tag
	HasCSN bool // Callsign
	HasIFI bool // IFPS_FLIGHT_ID
	HasFCT bool // Flight Category
	HasTAC bool // Type of Aircraft
	HasWTC bool // Wake Turbulence Category
	HasDEP bool // Departure Airport
	HasDST bool // Destination Airport
	HasRDS bool // Runway Designation
	HasCFL bool // Current Cleared Flight Level
	HasCTL bool // Current Control Position
	HasTOD bool // Time of Departure
	HasAST bool // Aircraft Stand
	HasSTS bool // Stand Status

	// Subfield values
	TAG FPPSTag       // SAC/SIC of FPPS
	CSN string        // Callsign (7 chars)
	IFI IFPSFlightID  // IFPS Flight ID
	FCT FlightCategory
	TAC string        // Type of Aircraft (4 chars)
	WTC byte          // Wake Turbulence Category (L/M/H/J)
	DEP string        // Departure Airport (4 chars)
	DST string        // Destination Airport (4 chars)
	RDS RunwayDesig   // Runway Designation
	CFL float64       // Current Cleared FL (LSB = 1/4 FL)
	CTL ControlPos    // Current Control Position
	TOD []TimeOfDep   // Time of Departure (repetitive)
	AST string        // Aircraft Stand (6 chars)
	STS StandStatus
}

type FPPSTag struct {
	SAC uint8
	SIC uint8
}

type IFPSFlightID struct {
	TYP uint8  // 0=Plan, 1-3=Unit internal
	NBR uint32 // Number 0-99999999
}

type FlightCategory struct {
	GATOAT uint8 // 0=Unknown, 1=GAT, 2=OAT, 3=N/A
	FR     uint8 // 0=IFR, 1=VFR, 2=N/A, 3=CVFR
	RVSM   uint8 // 0=Unknown, 1=Approved, 2=Exempt, 3=Not Approved
	HPR    bool  // High Priority Flight
}

type RunwayDesig struct {
	NU1 byte // First number
	NU2 byte // Second number
	LTR byte // Letter
}

type ControlPos struct {
	Centre   uint8
	Position uint8
}

type TimeOfDep struct {
	TYP     uint8 // Time type (0-13)
	DAY     uint8 // 0=Today, 1=Yesterday, 2=Tomorrow
	Hours   uint8 // 0-23
	Minutes uint8 // 0-59
	AVS     bool  // Seconds available
	Seconds uint8 // 0-59
}

type StandStatus struct {
	EMP uint8 // 0=Empty, 1=Occupied, 2=Unknown
	AVL uint8 // 0=Available, 1=Not available, 2=Unknown
}

func (f *FlightPlanRelatedData) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Primary subfield
	octet1, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading I011/390 primary: %w", err)
	}
	bytesRead++

	f.HasTAG = (octet1 & 0x80) != 0
	f.HasCSN = (octet1 & 0x40) != 0
	f.HasIFI = (octet1 & 0x20) != 0
	f.HasFCT = (octet1 & 0x10) != 0
	f.HasTAC = (octet1 & 0x08) != 0
	f.HasWTC = (octet1 & 0x04) != 0
	f.HasDEP = (octet1 & 0x02) != 0
	fx := (octet1 & 0x01) != 0

	if fx {
		octet2, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading I011/390 ext: %w", err)
		}
		bytesRead++

		f.HasDST = (octet2 & 0x80) != 0
		f.HasRDS = (octet2 & 0x40) != 0
		f.HasCFL = (octet2 & 0x20) != 0
		f.HasCTL = (octet2 & 0x10) != 0
		f.HasTOD = (octet2 & 0x08) != 0
		f.HasAST = (octet2 & 0x04) != 0
		f.HasSTS = (octet2 & 0x02) != 0
		fx = (octet2 & 0x01) != 0

		for fx {
			data, err := buf.ReadByte()
			if err != nil {
				return bytesRead, fmt.Errorf("reading I011/390 ext: %w", err)
			}
			bytesRead++
			fx = (data & 0x01) != 0
		}
	}

	// Read subfields
	if f.HasTAG {
		var data [2]byte
		n, err := buf.Read(data[:])
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("reading TAG: %w", err)
		}
		f.TAG.SAC = data[0]
		f.TAG.SIC = data[1]
	}

	if f.HasCSN {
		var data [7]byte
		n, err := buf.Read(data[:])
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("reading CSN: %w", err)
		}
		f.CSN = strings.TrimRight(string(data[:]), " \x00")
	}

	if f.HasIFI {
		var data [4]byte
		n, err := buf.Read(data[:])
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("reading IFI: %w", err)
		}
		f.IFI.TYP = (data[0] >> 6) & 0x03
		raw := binary.BigEndian.Uint32(data[:])
		f.IFI.NBR = raw & 0x07FFFFFF
	}

	if f.HasFCT {
		data, err := buf.ReadByte()
		bytesRead++
		if err != nil {
			return bytesRead, fmt.Errorf("reading FCT: %w", err)
		}
		f.FCT.GATOAT = (data >> 6) & 0x03
		f.FCT.FR = (data >> 4) & 0x03
		f.FCT.RVSM = (data >> 2) & 0x03
		f.FCT.HPR = (data & 0x02) != 0
	}

	if f.HasTAC {
		var data [4]byte
		n, err := buf.Read(data[:])
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("reading TAC: %w", err)
		}
		f.TAC = strings.TrimRight(string(data[:]), " \x00")
	}

	if f.HasWTC {
		data, err := buf.ReadByte()
		bytesRead++
		if err != nil {
			return bytesRead, fmt.Errorf("reading WTC: %w", err)
		}
		f.WTC = data
	}

	if f.HasDEP {
		var data [4]byte
		n, err := buf.Read(data[:])
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("reading DEP: %w", err)
		}
		f.DEP = strings.TrimRight(string(data[:]), " \x00")
	}

	if f.HasDST {
		var data [4]byte
		n, err := buf.Read(data[:])
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("reading DST: %w", err)
		}
		f.DST = strings.TrimRight(string(data[:]), " \x00")
	}

	if f.HasRDS {
		var data [3]byte
		n, err := buf.Read(data[:])
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("reading RDS: %w", err)
		}
		f.RDS.NU1 = data[0]
		f.RDS.NU2 = data[1]
		f.RDS.LTR = data[2]
	}

	if f.HasCFL {
		var data [2]byte
		n, err := buf.Read(data[:])
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("reading CFL: %w", err)
		}
		raw := binary.BigEndian.Uint16(data[:])
		f.CFL = float64(raw) * 0.25
	}

	if f.HasCTL {
		var data [2]byte
		n, err := buf.Read(data[:])
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("reading CTL: %w", err)
		}
		f.CTL.Centre = data[0]
		f.CTL.Position = data[1]
	}

	if f.HasTOD {
		rep, err := buf.ReadByte()
		bytesRead++
		if err != nil {
			return bytesRead, fmt.Errorf("reading TOD REP: %w", err)
		}

		f.TOD = make([]TimeOfDep, rep)
		for i := 0; i < int(rep); i++ {
			var data [4]byte
			n, err := buf.Read(data[:])
			bytesRead += n
			if err != nil {
				return bytesRead, fmt.Errorf("reading TOD: %w", err)
			}

			f.TOD[i].TYP = (data[0] >> 3) & 0x1F
			f.TOD[i].DAY = (data[0] >> 1) & 0x03
			f.TOD[i].Hours = (data[1] >> 3) & 0x1F
			f.TOD[i].Minutes = ((data[1] & 0x01) << 5) | ((data[2] >> 3) & 0x1F)
			f.TOD[i].AVS = (data[3] & 0x80) != 0
			f.TOD[i].Seconds = data[3] & 0x3F
		}
	}

	if f.HasAST {
		var data [6]byte
		n, err := buf.Read(data[:])
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("reading AST: %w", err)
		}
		f.AST = strings.TrimRight(string(data[:]), " \x00")
	}

	if f.HasSTS {
		data, err := buf.ReadByte()
		bytesRead++
		if err != nil {
			return bytesRead, fmt.Errorf("reading STS: %w", err)
		}
		f.STS.EMP = (data >> 6) & 0x03
		f.STS.AVL = (data >> 4) & 0x03
	}

	return bytesRead, nil
}

func (f *FlightPlanRelatedData) Encode(buf *bytes.Buffer) (int, error) {
	bytesWritten := 0

	// Build primary subfield
	needSecondOctet := f.HasDST || f.HasRDS || f.HasCFL || f.HasCTL || f.HasTOD || f.HasAST || f.HasSTS

	var octet1 uint8
	if f.HasTAG {
		octet1 |= 0x80
	}
	if f.HasCSN {
		octet1 |= 0x40
	}
	if f.HasIFI {
		octet1 |= 0x20
	}
	if f.HasFCT {
		octet1 |= 0x10
	}
	if f.HasTAC {
		octet1 |= 0x08
	}
	if f.HasWTC {
		octet1 |= 0x04
	}
	if f.HasDEP {
		octet1 |= 0x02
	}
	if needSecondOctet {
		octet1 |= 0x01
	}

	if err := buf.WriteByte(octet1); err != nil {
		return bytesWritten, fmt.Errorf("writing I011/390 primary: %w", err)
	}
	bytesWritten++

	if needSecondOctet {
		var octet2 uint8
		if f.HasDST {
			octet2 |= 0x80
		}
		if f.HasRDS {
			octet2 |= 0x40
		}
		if f.HasCFL {
			octet2 |= 0x20
		}
		if f.HasCTL {
			octet2 |= 0x10
		}
		if f.HasTOD {
			octet2 |= 0x08
		}
		if f.HasAST {
			octet2 |= 0x04
		}
		if f.HasSTS {
			octet2 |= 0x02
		}

		if err := buf.WriteByte(octet2); err != nil {
			return bytesWritten, fmt.Errorf("writing I011/390 ext: %w", err)
		}
		bytesWritten++
	}

	// Write subfields
	if f.HasTAG {
		if _, err := buf.Write([]byte{f.TAG.SAC, f.TAG.SIC}); err != nil {
			return bytesWritten, fmt.Errorf("writing TAG: %w", err)
		}
		bytesWritten += 2
	}

	if f.HasCSN {
		csn := f.CSN
		for len(csn) < 7 {
			csn += " "
		}
		if _, err := buf.Write([]byte(csn[:7])); err != nil {
			return bytesWritten, fmt.Errorf("writing CSN: %w", err)
		}
		bytesWritten += 7
	}

	if f.HasIFI {
		var data [4]byte
		val := (uint32(f.IFI.TYP) << 30) | (f.IFI.NBR & 0x07FFFFFF)
		binary.BigEndian.PutUint32(data[:], val)
		if _, err := buf.Write(data[:]); err != nil {
			return bytesWritten, fmt.Errorf("writing IFI: %w", err)
		}
		bytesWritten += 4
	}

	if f.HasFCT {
		data := (f.FCT.GATOAT << 6) | (f.FCT.FR << 4) | (f.FCT.RVSM << 2)
		if f.FCT.HPR {
			data |= 0x02
		}
		if err := buf.WriteByte(data); err != nil {
			return bytesWritten, fmt.Errorf("writing FCT: %w", err)
		}
		bytesWritten++
	}

	if f.HasTAC {
		tac := f.TAC
		for len(tac) < 4 {
			tac += " "
		}
		if _, err := buf.Write([]byte(tac[:4])); err != nil {
			return bytesWritten, fmt.Errorf("writing TAC: %w", err)
		}
		bytesWritten += 4
	}

	if f.HasWTC {
		if err := buf.WriteByte(f.WTC); err != nil {
			return bytesWritten, fmt.Errorf("writing WTC: %w", err)
		}
		bytesWritten++
	}

	if f.HasDEP {
		dep := f.DEP
		for len(dep) < 4 {
			dep += " "
		}
		if _, err := buf.Write([]byte(dep[:4])); err != nil {
			return bytesWritten, fmt.Errorf("writing DEP: %w", err)
		}
		bytesWritten += 4
	}

	if f.HasDST {
		dst := f.DST
		for len(dst) < 4 {
			dst += " "
		}
		if _, err := buf.Write([]byte(dst[:4])); err != nil {
			return bytesWritten, fmt.Errorf("writing DST: %w", err)
		}
		bytesWritten += 4
	}

	if f.HasRDS {
		if _, err := buf.Write([]byte{f.RDS.NU1, f.RDS.NU2, f.RDS.LTR}); err != nil {
			return bytesWritten, fmt.Errorf("writing RDS: %w", err)
		}
		bytesWritten += 3
	}

	if f.HasCFL {
		var data [2]byte
		binary.BigEndian.PutUint16(data[:], uint16(f.CFL/0.25))
		if _, err := buf.Write(data[:]); err != nil {
			return bytesWritten, fmt.Errorf("writing CFL: %w", err)
		}
		bytesWritten += 2
	}

	if f.HasCTL {
		if _, err := buf.Write([]byte{f.CTL.Centre, f.CTL.Position}); err != nil {
			return bytesWritten, fmt.Errorf("writing CTL: %w", err)
		}
		bytesWritten += 2
	}

	if f.HasTOD {
		if err := buf.WriteByte(uint8(len(f.TOD))); err != nil {
			return bytesWritten, fmt.Errorf("writing TOD REP: %w", err)
		}
		bytesWritten++

		for _, tod := range f.TOD {
			var data [4]byte
			data[0] = (tod.TYP << 3) | (tod.DAY << 1)
			data[1] = (tod.Hours << 3) | ((tod.Minutes >> 5) & 0x01)
			data[2] = (tod.Minutes & 0x1F) << 3
			data[3] = tod.Seconds & 0x3F
			if tod.AVS {
				data[3] |= 0x80
			}

			if _, err := buf.Write(data[:]); err != nil {
				return bytesWritten, fmt.Errorf("writing TOD: %w", err)
			}
			bytesWritten += 4
		}
	}

	if f.HasAST {
		ast := f.AST
		for len(ast) < 6 {
			ast += " "
		}
		if _, err := buf.Write([]byte(ast[:6])); err != nil {
			return bytesWritten, fmt.Errorf("writing AST: %w", err)
		}
		bytesWritten += 6
	}

	if f.HasSTS {
		data := (f.STS.EMP << 6) | (f.STS.AVL << 4)
		if err := buf.WriteByte(data); err != nil {
			return bytesWritten, fmt.Errorf("writing STS: %w", err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

func (f *FlightPlanRelatedData) Validate() error {
	return nil
}

func (f *FlightPlanRelatedData) String() string {
	parts := []string{}
	if f.HasCSN {
		parts = append(parts, fmt.Sprintf("Callsign=%s", f.CSN))
	}
	if f.HasTAC {
		parts = append(parts, fmt.Sprintf("Type=%s", f.TAC))
	}
	if f.HasWTC {
		parts = append(parts, fmt.Sprintf("WTC=%c", f.WTC))
	}
	if f.HasDEP {
		parts = append(parts, fmt.Sprintf("DEP=%s", f.DEP))
	}
	if f.HasDST {
		parts = append(parts, fmt.Sprintf("DST=%s", f.DST))
	}
	if f.HasCFL {
		parts = append(parts, fmt.Sprintf("CFL=FL%.0f", f.CFL))
	}
	return fmt.Sprintf("Flight Plan: %v", parts)
}
