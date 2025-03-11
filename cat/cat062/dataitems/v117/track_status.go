// dataitems/cat062/track_status.go
package v117

import (
	"bytes"
	"fmt"
	"strings"
)

// TrackStatus implements I062/080
// Status of a track
type TrackStatus struct {
	// Main fields (first octet)
	MON bool  // Monosensor track (true) or multisensor (false)
	SPI bool  // SPI present in the last report
	MRH bool  // Most Reliable Height (true = Geometric, false = Barometric)
	SRC uint8 // Source of calculated track altitude (0-7)
	CNF bool  // Confirmed track (false) or tentative track (true)

	// First extension
	SIM bool // Simulated track
	TSE bool // Last message transmitted to the user for the track
	TSB bool // First message transmitted to the user for the track
	FPC bool // Flight plan correlated
	AFF bool // ADS-B data inconsistent with other surveillance information
	STP bool // Slave Track Promotion
	KOS bool // Background service used (true) or Complementary service (false)

	// Second extension
	AMA bool  // Track from amalgamation process
	MD4 uint8 // Mode 4 interrogation (0-3)
	ME  bool  // Military Emergency present
	MI  bool  // Military Identification present
	MD5 uint8 // Mode 5 interrogation (0-3)

	// Third extension
	CST bool // Age of the last track update is higher than threshold (coasting)
	PSR bool // Age of the last PSR track update is higher than threshold
	SSR bool // Age of the last SSR track update is higher than threshold
	MDS bool // Age of the last Mode S track update is higher than threshold
	ADS bool // Age of the last ADS-B track update is higher than threshold
	SUC bool // Special Used Code
	AAC bool // Assigned Mode A Code Conflict

	// Fourth extension
	SDS  uint8 // Surveillance Data Status (0-3)
	EMS  uint8 // Emergency Status (0-7)
	PFT  bool  // Potential False Track Indication
	FPLT bool  // Track created/updated with FPL data

	// Fifth extension
	DUPT bool // Duplicate Mode 3/A Code
	DUPF bool // Duplicate Flight Plan
	DUPM bool // Duplicate Flight Plan due to manual correlation
	SFC  bool // Surface target
	IDD  bool // Duplicate Flight-ID
	IEC  bool // Inconsistent Emergency Code
	MLAT bool // Age of the last MLAT track update is higher than threshold

	// Counts which extensions are present (0-5)
	hasExtensions uint8
}

func (t *TrackStatus) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Read first octet (mandatory)
	b, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading track status: %w", err)
	}
	bytesRead++

	// Parse first octet
	t.MON = (b & 0x80) != 0 // bit 8
	t.SPI = (b & 0x40) != 0 // bit 7
	t.MRH = (b & 0x20) != 0 // bit 6
	t.SRC = (b & 0x38) >> 3 // bits 5-3
	t.CNF = (b & 0x02) != 0 // bit 2
	fx := (b & 0x01) != 0   // bit 1 (FX)

	// Read first extension if present
	if fx {
		t.hasExtensions = 1
		b, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading track status first extension: %w", err)
		}
		bytesRead++

		// Parse first extension
		t.SIM = (b & 0x80) != 0 // bit 8
		t.TSE = (b & 0x40) != 0 // bit 7
		t.TSB = (b & 0x20) != 0 // bit 6
		t.FPC = (b & 0x10) != 0 // bit 5
		t.AFF = (b & 0x08) != 0 // bit 4
		t.STP = (b & 0x04) != 0 // bit 3
		t.KOS = (b & 0x02) != 0 // bit 2
		fx = (b & 0x01) != 0    // bit 1 (FX)

		// Read second extension if present
		if fx {
			t.hasExtensions = 2
			b, err = buf.ReadByte()
			if err != nil {
				return bytesRead, fmt.Errorf("reading track status second extension: %w", err)
			}
			bytesRead++

			// Parse second extension
			t.AMA = (b & 0x80) != 0 // bit 8
			t.MD4 = (b & 0x60) >> 5 // bits 7-6
			t.ME = (b & 0x10) != 0  // bit 5
			t.MI = (b & 0x08) != 0  // bit 4
			t.MD5 = (b & 0x06) >> 1 // bits 3-2
			fx = (b & 0x01) != 0    // bit 1 (FX)

			// Read third extension if present
			if fx {
				t.hasExtensions = 3
				b, err = buf.ReadByte()
				if err != nil {
					return bytesRead, fmt.Errorf("reading track status third extension: %w", err)
				}
				bytesRead++

				// Parse third extension
				t.CST = (b & 0x80) != 0 // bit 8
				t.PSR = (b & 0x40) != 0 // bit 7
				t.SSR = (b & 0x20) != 0 // bit 6
				t.MDS = (b & 0x10) != 0 // bit 5
				t.ADS = (b & 0x08) != 0 // bit 4
				t.SUC = (b & 0x04) != 0 // bit 3
				t.AAC = (b & 0x02) != 0 // bit 2
				fx = (b & 0x01) != 0    // bit 1 (FX)

				// Read fourth extension if present
				if fx {
					t.hasExtensions = 4
					b, err = buf.ReadByte()
					if err != nil {
						return bytesRead, fmt.Errorf("reading track status fourth extension: %w", err)
					}
					bytesRead++

					// Parse fourth extension
					t.SDS = (b & 0xC0) >> 6  // bits 8-7
					t.EMS = (b & 0x38) >> 3  // bits 6-4
					t.PFT = (b & 0x04) != 0  // bit 3
					t.FPLT = (b & 0x02) != 0 // bit 2
					fx = (b & 0x01) != 0     // bit 1 (FX)

					// Read fifth extension if present
					if fx {
						t.hasExtensions = 5
						b, err = buf.ReadByte()
						if err != nil {
							return bytesRead, fmt.Errorf("reading track status fifth extension: %w", err)
						}
						bytesRead++

						// Parse fifth extension
						t.DUPT = (b & 0x80) != 0 // bit 8
						t.DUPF = (b & 0x40) != 0 // bit 7
						t.DUPM = (b & 0x20) != 0 // bit 6
						t.SFC = (b & 0x10) != 0  // bit 5
						t.IDD = (b & 0x08) != 0  // bit 4
						t.IEC = (b & 0x04) != 0  // bit 3
						t.MLAT = (b & 0x02) != 0 // bit 2
						// Bit 1 is FX but we don't check it as there are no more extensions
					}
				}
			}
		}
	}

	return bytesRead, nil
}

func (t *TrackStatus) Encode(buf *bytes.Buffer) (int, error) {
	bytesWritten := 0

	// First octet (mandatory)
	b := byte(0)
	if t.MON {
		b |= 0x80 // bit 8
	}
	if t.SPI {
		b |= 0x40 // bit 7
	}
	if t.MRH {
		b |= 0x20 // bit 6
	}
	b |= (t.SRC & 0x07) << 3 // bits 5-3
	if t.CNF {
		b |= 0x02 // bit 2
	}
	if t.hasExtensions > 0 {
		b |= 0x01 // bit 1 (FX)
	}

	err := buf.WriteByte(b)
	if err != nil {
		return bytesWritten, fmt.Errorf("writing track status: %w", err)
	}
	bytesWritten++

	// First extension if needed
	if t.hasExtensions > 0 {
		b = byte(0)
		if t.SIM {
			b |= 0x80 // bit 8
		}
		if t.TSE {
			b |= 0x40 // bit 7
		}
		if t.TSB {
			b |= 0x20 // bit 6
		}
		if t.FPC {
			b |= 0x10 // bit 5
		}
		if t.AFF {
			b |= 0x08 // bit 4
		}
		if t.STP {
			b |= 0x04 // bit 3
		}
		if t.KOS {
			b |= 0x02 // bit 2
		}
		if t.hasExtensions > 1 {
			b |= 0x01 // bit 1 (FX)
		}

		err = buf.WriteByte(b)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing track status first extension: %w", err)
		}
		bytesWritten++

		// Second extension if needed
		if t.hasExtensions > 1 {
			b = byte(0)
			if t.AMA {
				b |= 0x80 // bit 8
			}
			b |= (t.MD4 & 0x03) << 5 // bits 7-6
			if t.ME {
				b |= 0x10 // bit 5
			}
			if t.MI {
				b |= 0x08 // bit 4
			}
			b |= (t.MD5 & 0x03) << 1 // bits 3-2
			if t.hasExtensions > 2 {
				b |= 0x01 // bit 1 (FX)
			}

			err = buf.WriteByte(b)
			if err != nil {
				return bytesWritten, fmt.Errorf("writing track status second extension: %w", err)
			}
			bytesWritten++

			// Third extension if needed
			if t.hasExtensions > 2 {
				b = byte(0)
				if t.CST {
					b |= 0x80 // bit 8
				}
				if t.PSR {
					b |= 0x40 // bit 7
				}
				if t.SSR {
					b |= 0x20 // bit 6
				}
				if t.MDS {
					b |= 0x10 // bit 5
				}
				if t.ADS {
					b |= 0x08 // bit 4
				}
				if t.SUC {
					b |= 0x04 // bit 3
				}
				if t.AAC {
					b |= 0x02 // bit 2
				}
				if t.hasExtensions > 3 {
					b |= 0x01 // bit 1 (FX)
				}

				err = buf.WriteByte(b)
				if err != nil {
					return bytesWritten, fmt.Errorf("writing track status third extension: %w", err)
				}
				bytesWritten++

				// Fourth extension if needed
				if t.hasExtensions > 3 {
					b = byte(0)
					b |= (t.SDS & 0x03) << 6 // bits 8-7
					b |= (t.EMS & 0x07) << 3 // bits 6-4
					if t.PFT {
						b |= 0x04 // bit 3
					}
					if t.FPLT {
						b |= 0x02 // bit 2
					}
					if t.hasExtensions > 4 {
						b |= 0x01 // bit 1 (FX)
					}

					err = buf.WriteByte(b)
					if err != nil {
						return bytesWritten, fmt.Errorf("writing track status fourth extension: %w", err)
					}
					bytesWritten++

					// Fifth extension if needed
					if t.hasExtensions > 4 {
						b = byte(0)
						if t.DUPT {
							b |= 0x80 // bit 8
						}
						if t.DUPF {
							b |= 0x40 // bit 7
						}
						if t.DUPM {
							b |= 0x20 // bit 6
						}
						if t.SFC {
							b |= 0x10 // bit 5
						}
						if t.IDD {
							b |= 0x08 // bit 4
						}
						if t.IEC {
							b |= 0x04 // bit 3
						}
						if t.MLAT {
							b |= 0x02 // bit 2
						}
						// Bit 1 is FX but we don't set it as there are no more extensions

						err = buf.WriteByte(b)
						if err != nil {
							return bytesWritten, fmt.Errorf("writing track status fifth extension: %w", err)
						}
						bytesWritten++
					}
				}
			}
		}
	}

	return bytesWritten, nil
}

func (t *TrackStatus) Validate() error {
	if t.SRC > 7 {
		return fmt.Errorf("invalid SRC value: %d", t.SRC)
	}
	if t.MD4 > 3 {
		return fmt.Errorf("invalid MD4 value: %d", t.MD4)
	}
	if t.MD5 > 3 {
		return fmt.Errorf("invalid MD5 value: %d", t.MD5)
	}
	if t.SDS > 3 {
		return fmt.Errorf("invalid SDS value: %d", t.SDS)
	}
	if t.EMS > 7 {
		return fmt.Errorf("invalid EMS value: %d", t.EMS)
	}
	return nil
}

func (t *TrackStatus) String() string {
	var details []string

	// Main fields
	if t.MON {
		details = append(details, "Monosensor")
	} else {
		details = append(details, "Multisensor")
	}

	if t.SPI {
		details = append(details, "SPI")
	}

	if t.MRH {
		details = append(details, "Geo Alt")
	} else {
		details = append(details, "Baro Alt")
	}

	// Source of calculated track altitude
	srcMap := map[uint8]string{
		0: "No Source",
		1: "GNSS",
		2: "3D Radar",
		3: "Triangulation",
		4: "Height Coverage",
		5: "Speed Lookup",
		6: "Default Height",
		7: "Multilateration",
	}
	details = append(details, fmt.Sprintf("SRC: %s", srcMap[t.SRC]))

	if t.CNF {
		details = append(details, "Tentative")
	} else {
		details = append(details, "Confirmed")
	}

	// First extension
	if t.hasExtensions > 0 {
		if t.SIM {
			details = append(details, "Simulated")
		}
		if t.TSE {
			details = append(details, "Last Message")
		}
		if t.TSB {
			details = append(details, "First Message")
		}
		if t.FPC {
			details = append(details, "Flight Plan Correlated")
		}
		if t.AFF {
			details = append(details, "ADS-B Inconsistent")
		}
		if t.STP {
			details = append(details, "Slave Track Promotion")
		}
		if t.KOS {
			details = append(details, "Background Service")
		}
	}

	// Second extension
	if t.hasExtensions > 1 {
		if t.AMA {
			details = append(details, "Amalgamated")
		}

		md4Map := map[uint8]string{
			0: "No Mode 4",
			1: "Mode 4 Friendly",
			2: "Mode 4 Unknown",
			3: "Mode 4 No Reply",
		}
		details = append(details, fmt.Sprintf("MD4: %s", md4Map[t.MD4]))

		if t.ME {
			details = append(details, "Military Emergency")
		}
		if t.MI {
			details = append(details, "Military ID")
		}

		md5Map := map[uint8]string{
			0: "No Mode 5",
			1: "Mode 5 Friendly",
			2: "Mode 5 Unknown",
			3: "Mode 5 No Reply",
		}
		details = append(details, fmt.Sprintf("MD5: %s", md5Map[t.MD5]))
	}

	// Third extension
	if t.hasExtensions > 2 {
		if t.CST {
			details = append(details, "Coasting")
		}
		if t.PSR {
			details = append(details, "PSR Coast")
		}
		if t.SSR {
			details = append(details, "SSR Coast")
		}
		if t.MDS {
			details = append(details, "Mode S Coast")
		}
		if t.ADS {
			details = append(details, "ADS-B Coast")
		}
		if t.SUC {
			details = append(details, "Special Used Code")
		}
		if t.AAC {
			details = append(details, "Mode A Conflict")
		}
	}

	// Fourth extension
	if t.hasExtensions > 3 {
		sdsMap := map[uint8]string{
			0: "Combined",
			1: "Co-operative Only",
			2: "Non-Cooperative Only",
			3: "Not Defined",
		}
		details = append(details, fmt.Sprintf("SDS: %s", sdsMap[t.SDS]))

		emsMap := map[uint8]string{
			0: "No Emergency",
			1: "General Emergency",
			2: "Medical Emergency",
			3: "Minimum Fuel",
			4: "No Communications",
			5: "Unlawful Interference",
			6: "Downed Aircraft",
			7: "Undefined",
		}
		details = append(details, fmt.Sprintf("EMS: %s", emsMap[t.EMS]))

		if t.PFT {
			details = append(details, "Potential False Track")
		}
		if t.FPLT {
			details = append(details, "FPL Track")
		}
	}

	// Fifth extension
	if t.hasExtensions > 4 {
		if t.DUPT {
			details = append(details, "Duplicate Mode 3/A")
		}
		if t.DUPF {
			details = append(details, "Duplicate Flight Plan")
		}
		if t.DUPM {
			details = append(details, "Duplicate Manual")
		}
		if t.SFC {
			details = append(details, "Surface")
		}
		if t.IDD {
			details = append(details, "Duplicate Flight-ID")
		}
		if t.IEC {
			details = append(details, "Inconsistent Emergency")
		}
		if t.MLAT {
			details = append(details, "MLAT Coast")
		}
	}

	return strings.Join(details, ", ")
}

// SetHasExtension sets the appropriate hasExtension value based on which fields are used
func (t *TrackStatus) SetHasExtension() {
	// Check if any field from fifth extension is set
	if t.DUPT || t.DUPF || t.DUPM || t.SFC || t.IDD || t.IEC || t.MLAT {
		t.hasExtensions = 5
		return
	}

	// Check if any field from fourth extension is set
	if t.SDS > 0 || t.EMS > 0 || t.PFT || t.FPLT {
		t.hasExtensions = 4
		return
	}

	// Check if any field from third extension is set
	if t.CST || t.PSR || t.SSR || t.MDS || t.ADS || t.SUC || t.AAC {
		t.hasExtensions = 3
		return
	}

	// Check if any field from second extension is set
	if t.AMA || t.MD4 > 0 || t.ME || t.MI || t.MD5 > 0 {
		t.hasExtensions = 2
		return
	}

	// Check if any field from first extension is set
	if t.SIM || t.TSE || t.TSB || t.FPC || t.AFF || t.STP || t.KOS {
		t.hasExtensions = 1
		return
	}

	t.hasExtensions = 0
}
