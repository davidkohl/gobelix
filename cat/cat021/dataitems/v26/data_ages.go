// cat/cat021/dataitems/v26/data_ages.go

package v26

import (
	"bytes"
	"fmt"
	"strings"
)

// DataAges implements I021/295
type DataAges struct {
	// Primary subfield (first octet)
	AOS bool // Aircraft Operational Status age
	TRD bool // Target Report Descriptor age
	M3A bool // Mode 3/A Code age
	QI  bool // Quality Indicators age
	TI  bool // Trajectory Intent age
	MAM bool // Message Amplitude age
	GH  bool // Geometric Height age

	// Second octet (first extension)
	FL  bool // Flight Level age
	ISA bool // Intermediate State Selected Altitude age
	FSA bool // Final State Selected Altitude age
	AS  bool // Air Speed age
	TAS bool // True Air Speed age
	MH  bool // Magnetic Heading age
	BVR bool // Barometric Vertical Rate age

	// Third octet (second extension)
	GVR bool // Geometric Vertical Rate age
	GV  bool // Ground Vector age
	TAR bool // Track Angle Rate age
	TID bool // Target Identification age
	TS  bool // Target Status age
	MET bool // Met Information age
	ROA bool // Roll Angle age

	// Fourth octet (third extension)
	ARA bool // ACAS Resolution Advisory age
	SCC bool // Surface Capabilities and Characteristics age

	// Age values in tenths of a second (0.1s)
	AOSAge uint8 // Aircraft Operational Status age
	TRDAge uint8 // Target Report Descriptor age
	M3AAge uint8 // Mode 3/A Code age
	QIAge  uint8 // Quality Indicators age
	TIAge  uint8 // Trajectory Intent age
	MAMAge uint8 // Message Amplitude age
	GHAge  uint8 // Geometric Height age

	FLAge  uint8 // Flight Level age
	ISAAge uint8 // Intermediate State Selected Altitude age
	FSAAge uint8 // Final State Selected Altitude age
	ASAge  uint8 // Air Speed age
	TASAge uint8 // True Air Speed age
	MHAge  uint8 // Magnetic Heading age
	BVRAge uint8 // Barometric Vertical Rate age

	GVRAge uint8 // Geometric Vertical Rate age
	GVAge  uint8 // Ground Vector age
	TARAge uint8 // Track Angle Rate age
	TIDAge uint8 // Target Identification age
	TSAge  uint8 // Target Status age
	METAge uint8 // Met Information age
	ROAAge uint8 // Roll Angle age

	ARAAge uint8 // ACAS Resolution Advisory age
	SCCAge uint8 // Surface Capabilities and Characteristics age
}

// HasAnyAge checks if there's at least one age value set
func (d *DataAges) HasAnyAge() bool {
	return d.AOS || d.TRD || d.M3A || d.QI || d.TI || d.MAM || d.GH ||
		d.FL || d.ISA || d.FSA || d.AS || d.TAS || d.MH || d.BVR ||
		d.GVR || d.GV || d.TAR || d.TID || d.TS || d.MET || d.ROA ||
		d.ARA || d.SCC
}

// HasFirstExtension checks if any field in the first extension is set
func (d *DataAges) HasFirstExtension() bool {
	return d.FL || d.ISA || d.FSA || d.AS || d.TAS || d.MH || d.BVR
}

// HasSecondExtension checks if any field in the second extension is set
func (d *DataAges) HasSecondExtension() bool {
	return d.GVR || d.GV || d.TAR || d.TID || d.TS || d.MET || d.ROA
}

// HasThirdExtension checks if any field in the third extension is set
func (d *DataAges) HasThirdExtension() bool {
	return d.ARA || d.SCC
}

func (d *DataAges) Encode(buf *bytes.Buffer) (int, error) {
	if !d.HasAnyAge() {
		return 0, fmt.Errorf("no data ages to encode")
	}

	bytesWritten := 0

	// Create primary subfield
	var octet1 uint8 = 0

	if d.AOS {
		octet1 |= 0x80 // bit 32
	}
	if d.TRD {
		octet1 |= 0x40 // bit 31
	}
	if d.M3A {
		octet1 |= 0x20 // bit 30
	}
	if d.QI {
		octet1 |= 0x10 // bit 29
	}
	if d.TI {
		octet1 |= 0x08 // bit 28
	}
	if d.MAM {
		octet1 |= 0x04 // bit 27
	}
	if d.GH {
		octet1 |= 0x02 // bit 26
	}

	// Set extension bit if needed
	if d.HasFirstExtension() || d.HasSecondExtension() || d.HasThirdExtension() {
		octet1 |= 0x01 // bit 25 (FX)
	}

	// Write first octet
	if err := buf.WriteByte(octet1); err != nil {
		return bytesWritten, fmt.Errorf("writing primary field: %w", err)
	}
	bytesWritten++

	// If no extension, write the actual ages for primary subfield
	if octet1&0x01 == 0 {
		// Write ages for primary subfield
		if d.AOS {
			if err := buf.WriteByte(d.AOSAge); err != nil {
				return bytesWritten, fmt.Errorf("writing AOS age: %w", err)
			}
			bytesWritten++
		}
		if d.TRD {
			if err := buf.WriteByte(d.TRDAge); err != nil {
				return bytesWritten, fmt.Errorf("writing TRD age: %w", err)
			}
			bytesWritten++
		}
		if d.M3A {
			if err := buf.WriteByte(d.M3AAge); err != nil {
				return bytesWritten, fmt.Errorf("writing M3A age: %w", err)
			}
			bytesWritten++
		}
		if d.QI {
			if err := buf.WriteByte(d.QIAge); err != nil {
				return bytesWritten, fmt.Errorf("writing QI age: %w", err)
			}
			bytesWritten++
		}
		if d.TI {
			if err := buf.WriteByte(d.TIAge); err != nil {
				return bytesWritten, fmt.Errorf("writing TI age: %w", err)
			}
			bytesWritten++
		}
		if d.MAM {
			if err := buf.WriteByte(d.MAMAge); err != nil {
				return bytesWritten, fmt.Errorf("writing MAM age: %w", err)
			}
			bytesWritten++
		}
		if d.GH {
			if err := buf.WriteByte(d.GHAge); err != nil {
				return bytesWritten, fmt.Errorf("writing GH age: %w", err)
			}
			bytesWritten++
		}
		return bytesWritten, nil
	}

	// First extension
	var octet2 uint8 = 0
	if d.FL {
		octet2 |= 0x80 // bit 24
	}
	if d.ISA {
		octet2 |= 0x40 // bit 23
	}
	if d.FSA {
		octet2 |= 0x20 // bit 22
	}
	if d.AS {
		octet2 |= 0x10 // bit 21
	}
	if d.TAS {
		octet2 |= 0x08 // bit 20
	}
	if d.MH {
		octet2 |= 0x04 // bit 19
	}
	if d.BVR {
		octet2 |= 0x02 // bit 18
	}

	// Set extension bit if needed
	if d.HasSecondExtension() || d.HasThirdExtension() {
		octet2 |= 0x01 // bit 17 (FX)
	}

	// Write second octet
	if err := buf.WriteByte(octet2); err != nil {
		return bytesWritten, fmt.Errorf("writing first extension: %w", err)
	}
	bytesWritten++

	// Write ages for primary subfield
	if d.AOS {
		if err := buf.WriteByte(d.AOSAge); err != nil {
			return bytesWritten, fmt.Errorf("writing AOS age: %w", err)
		}
		bytesWritten++
	}
	if d.TRD {
		if err := buf.WriteByte(d.TRDAge); err != nil {
			return bytesWritten, fmt.Errorf("writing TRD age: %w", err)
		}
		bytesWritten++
	}
	if d.M3A {
		if err := buf.WriteByte(d.M3AAge); err != nil {
			return bytesWritten, fmt.Errorf("writing M3A age: %w", err)
		}
		bytesWritten++
	}
	if d.QI {
		if err := buf.WriteByte(d.QIAge); err != nil {
			return bytesWritten, fmt.Errorf("writing QI age: %w", err)
		}
		bytesWritten++
	}
	if d.TI {
		if err := buf.WriteByte(d.TIAge); err != nil {
			return bytesWritten, fmt.Errorf("writing TI age: %w", err)
		}
		bytesWritten++
	}
	if d.MAM {
		if err := buf.WriteByte(d.MAMAge); err != nil {
			return bytesWritten, fmt.Errorf("writing MAM age: %w", err)
		}
		bytesWritten++
	}
	if d.GH {
		if err := buf.WriteByte(d.GHAge); err != nil {
			return bytesWritten, fmt.Errorf("writing GH age: %w", err)
		}
		bytesWritten++
	}

	// If no second extension, write the ages for first extension and return
	if octet2&0x01 == 0 {
		// Write ages for first extension
		if d.FL {
			if err := buf.WriteByte(d.FLAge); err != nil {
				return bytesWritten, fmt.Errorf("writing FL age: %w", err)
			}
			bytesWritten++
		}
		if d.ISA {
			if err := buf.WriteByte(d.ISAAge); err != nil {
				return bytesWritten, fmt.Errorf("writing ISA age: %w", err)
			}
			bytesWritten++
		}
		if d.FSA {
			if err := buf.WriteByte(d.FSAAge); err != nil {
				return bytesWritten, fmt.Errorf("writing FSA age: %w", err)
			}
			bytesWritten++
		}
		if d.AS {
			if err := buf.WriteByte(d.ASAge); err != nil {
				return bytesWritten, fmt.Errorf("writing AS age: %w", err)
			}
			bytesWritten++
		}
		if d.TAS {
			if err := buf.WriteByte(d.TASAge); err != nil {
				return bytesWritten, fmt.Errorf("writing TAS age: %w", err)
			}
			bytesWritten++
		}
		if d.MH {
			if err := buf.WriteByte(d.MHAge); err != nil {
				return bytesWritten, fmt.Errorf("writing MH age: %w", err)
			}
			bytesWritten++
		}
		if d.BVR {
			if err := buf.WriteByte(d.BVRAge); err != nil {
				return bytesWritten, fmt.Errorf("writing BVR age: %w", err)
			}
			bytesWritten++
		}
		return bytesWritten, nil
	}

	// Second extension
	var octet3 uint8 = 0
	if d.GVR {
		octet3 |= 0x80 // bit 16
	}
	if d.GV {
		octet3 |= 0x40 // bit 15
	}
	if d.TAR {
		octet3 |= 0x20 // bit 14
	}
	if d.TID {
		octet3 |= 0x10 // bit 13
	}
	if d.TS {
		octet3 |= 0x08 // bit 12
	}
	if d.MET {
		octet3 |= 0x04 // bit 11
	}
	if d.ROA {
		octet3 |= 0x02 // bit 10
	}

	// Set extension bit if needed
	if d.HasThirdExtension() {
		octet3 |= 0x01 // bit 9 (FX)
	}

	// Write third octet
	if err := buf.WriteByte(octet3); err != nil {
		return bytesWritten, fmt.Errorf("writing second extension: %w", err)
	}
	bytesWritten++

	// Write ages for first extension
	if d.FL {
		if err := buf.WriteByte(d.FLAge); err != nil {
			return bytesWritten, fmt.Errorf("writing FL age: %w", err)
		}
		bytesWritten++
	}
	if d.ISA {
		if err := buf.WriteByte(d.ISAAge); err != nil {
			return bytesWritten, fmt.Errorf("writing ISA age: %w", err)
		}
		bytesWritten++
	}
	if d.FSA {
		if err := buf.WriteByte(d.FSAAge); err != nil {
			return bytesWritten, fmt.Errorf("writing FSA age: %w", err)
		}
		bytesWritten++
	}
	if d.AS {
		if err := buf.WriteByte(d.ASAge); err != nil {
			return bytesWritten, fmt.Errorf("writing AS age: %w", err)
		}
		bytesWritten++
	}
	if d.TAS {
		if err := buf.WriteByte(d.TASAge); err != nil {
			return bytesWritten, fmt.Errorf("writing TAS age: %w", err)
		}
		bytesWritten++
	}
	if d.MH {
		if err := buf.WriteByte(d.MHAge); err != nil {
			return bytesWritten, fmt.Errorf("writing MH age: %w", err)
		}
		bytesWritten++
	}
	if d.BVR {
		if err := buf.WriteByte(d.BVRAge); err != nil {
			return bytesWritten, fmt.Errorf("writing BVR age: %w", err)
		}
		bytesWritten++
	}

	// If no third extension, write the ages for second extension and return
	if octet3&0x01 == 0 {
		// Write ages for second extension
		if d.GVR {
			if err := buf.WriteByte(d.GVRAge); err != nil {
				return bytesWritten, fmt.Errorf("writing GVR age: %w", err)
			}
			bytesWritten++
		}
		if d.GV {
			if err := buf.WriteByte(d.GVAge); err != nil {
				return bytesWritten, fmt.Errorf("writing GV age: %w", err)
			}
			bytesWritten++
		}
		if d.TAR {
			if err := buf.WriteByte(d.TARAge); err != nil {
				return bytesWritten, fmt.Errorf("writing TAR age: %w", err)
			}
			bytesWritten++
		}
		if d.TID {
			if err := buf.WriteByte(d.TIDAge); err != nil {
				return bytesWritten, fmt.Errorf("writing TID age: %w", err)
			}
			bytesWritten++
		}
		if d.TS {
			if err := buf.WriteByte(d.TSAge); err != nil {
				return bytesWritten, fmt.Errorf("writing TS age: %w", err)
			}
			bytesWritten++
		}
		if d.MET {
			if err := buf.WriteByte(d.METAge); err != nil {
				return bytesWritten, fmt.Errorf("writing MET age: %w", err)
			}
			bytesWritten++
		}
		if d.ROA {
			if err := buf.WriteByte(d.ROAAge); err != nil {
				return bytesWritten, fmt.Errorf("writing ROA age: %w", err)
			}
			bytesWritten++
		}
		return bytesWritten, nil
	}

	// Third extension
	var octet4 uint8 = 0
	if d.ARA {
		octet4 |= 0x80 // bit 8
	}
	if d.SCC {
		octet4 |= 0x40 // bit 7
	}
	// Bits 6-2 are spare, set to 0
	// Bit 1 is extension (FX), set to 0 for no further extension

	// Write fourth octet
	if err := buf.WriteByte(octet4); err != nil {
		return bytesWritten, fmt.Errorf("writing third extension: %w", err)
	}
	bytesWritten++

	// Write ages for second extension
	if d.GVR {
		if err := buf.WriteByte(d.GVRAge); err != nil {
			return bytesWritten, fmt.Errorf("writing GVR age: %w", err)
		}
		bytesWritten++
	}
	if d.GV {
		if err := buf.WriteByte(d.GVAge); err != nil {
			return bytesWritten, fmt.Errorf("writing GV age: %w", err)
		}
		bytesWritten++
	}
	if d.TAR {
		if err := buf.WriteByte(d.TARAge); err != nil {
			return bytesWritten, fmt.Errorf("writing TAR age: %w", err)
		}
		bytesWritten++
	}
	if d.TID {
		if err := buf.WriteByte(d.TIDAge); err != nil {
			return bytesWritten, fmt.Errorf("writing TID age: %w", err)
		}
		bytesWritten++
	}
	if d.TS {
		if err := buf.WriteByte(d.TSAge); err != nil {
			return bytesWritten, fmt.Errorf("writing TS age: %w", err)
		}
		bytesWritten++
	}
	if d.MET {
		if err := buf.WriteByte(d.METAge); err != nil {
			return bytesWritten, fmt.Errorf("writing MET age: %w", err)
		}
		bytesWritten++
	}
	if d.ROA {
		if err := buf.WriteByte(d.ROAAge); err != nil {
			return bytesWritten, fmt.Errorf("writing ROA age: %w", err)
		}
		bytesWritten++
	}

	// Write ages for third extension
	if d.ARA {
		if err := buf.WriteByte(d.ARAAge); err != nil {
			return bytesWritten, fmt.Errorf("writing ARA age: %w", err)
		}
		bytesWritten++
	}
	if d.SCC {
		if err := buf.WriteByte(d.SCCAge); err != nil {
			return bytesWritten, fmt.Errorf("writing SCC age: %w", err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

func (d *DataAges) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Read primary subfield
	octet1, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading primary field: %w", err)
	}
	bytesRead++

	// Extract flags from primary subfield
	d.AOS = (octet1 & 0x80) != 0
	d.TRD = (octet1 & 0x40) != 0
	d.M3A = (octet1 & 0x20) != 0
	d.QI = (octet1 & 0x10) != 0
	d.TI = (octet1 & 0x08) != 0
	d.MAM = (octet1 & 0x04) != 0
	d.GH = (octet1 & 0x02) != 0

	// Check for extension
	hasFirstExtension := (octet1 & 0x01) != 0

	// Read ages for primary subfield
	if d.AOS {
		d.AOSAge, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading AOS age: %w", err)
		}
		bytesRead++
	}
	if d.TRD {
		d.TRDAge, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading TRD age: %w", err)
		}
		bytesRead++
	}
	if d.M3A {
		d.M3AAge, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading M3A age: %w", err)
		}
		bytesRead++
	}
	if d.QI {
		d.QIAge, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading QI age: %w", err)
		}
		bytesRead++
	}
	if d.TI {
		d.TIAge, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading TI age: %w", err)
		}
		bytesRead++
	}
	if d.MAM {
		d.MAMAge, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading MAM age: %w", err)
		}
		bytesRead++
	}
	if d.GH {
		d.GHAge, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading GH age: %w", err)
		}
		bytesRead++
	}

	// If no extension, return
	if !hasFirstExtension {
		return bytesRead, nil
	}

	// Read first extension
	octet2, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading first extension: %w", err)
	}
	bytesRead++

	// Extract flags from first extension
	d.FL = (octet2 & 0x80) != 0
	d.ISA = (octet2 & 0x40) != 0
	d.FSA = (octet2 & 0x20) != 0
	d.AS = (octet2 & 0x10) != 0
	d.TAS = (octet2 & 0x08) != 0
	d.MH = (octet2 & 0x04) != 0
	d.BVR = (octet2 & 0x02) != 0

	// Check for extension
	hasSecondExtension := (octet2 & 0x01) != 0

	// Read ages for first extension
	if d.FL {
		d.FLAge, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading FL age: %w", err)
		}
		bytesRead++
	}
	if d.ISA {
		d.ISAAge, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading ISA age: %w", err)
		}
		bytesRead++
	}
	if d.FSA {
		d.FSAAge, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading FSA age: %w", err)
		}
		bytesRead++
	}
	if d.AS {
		d.ASAge, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading AS age: %w", err)
		}
		bytesRead++
	}
	if d.TAS {
		d.TASAge, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading TAS age: %w", err)
		}
		bytesRead++
	}
	if d.MH {
		d.MHAge, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading MH age: %w", err)
		}
		bytesRead++
	}
	if d.BVR {
		d.BVRAge, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading BVR age: %w", err)
		}
		bytesRead++
	}

	// If no second extension, return
	if !hasSecondExtension {
		return bytesRead, nil
	}

	// Read second extension
	octet3, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading second extension: %w", err)
	}
	bytesRead++

	// Extract flags from second extension
	d.GVR = (octet3 & 0x80) != 0
	d.GV = (octet3 & 0x40) != 0
	d.TAR = (octet3 & 0x20) != 0
	d.TID = (octet3 & 0x10) != 0
	d.TS = (octet3 & 0x08) != 0
	d.MET = (octet3 & 0x04) != 0
	d.ROA = (octet3 & 0x02) != 0

	// Check for extension
	hasThirdExtension := (octet3 & 0x01) != 0

	// Read ages for second extension
	if d.GVR {
		d.GVRAge, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading GVR age: %w", err)
		}
		bytesRead++
	}
	if d.GV {
		d.GVAge, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading GV age: %w", err)
		}
		bytesRead++
	}
	if d.TAR {
		d.TARAge, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading TAR age: %w", err)
		}
		bytesRead++
	}
	if d.TID {
		d.TIDAge, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading TID age: %w", err)
		}
		bytesRead++
	}
	if d.TS {
		d.TSAge, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading TS age: %w", err)
		}
		bytesRead++
	}
	if d.MET {
		d.METAge, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading MET age: %w", err)
		}
		bytesRead++
	}
	if d.ROA {
		d.ROAAge, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading ROA age: %w", err)
		}
		bytesRead++
	}

	// If no third extension, return
	if !hasThirdExtension {
		return bytesRead, nil
	}

	// Read third extension
	octet4, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading third extension: %w", err)
	}
	bytesRead++

	// Extract flags from third extension
	d.ARA = (octet4 & 0x80) != 0
	d.SCC = (octet4 & 0x40) != 0
	// Bits 6-2 are spare, Bit 1 is FX

	// Read ages for third extension
	if d.ARA {
		d.ARAAge, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading ARA age: %w", err)
		}
		bytesRead++
	}
	if d.SCC {
		d.SCCAge, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading SCC age: %w", err)
		}
		bytesRead++
	}

	return bytesRead, nil
}

func (d *DataAges) Validate() error {
	// No specific validation needed for ages
	return nil
}

// String returns a human-readable representation of the data ages
func (d *DataAges) String() string {
	var parts []string

	// Primary subfield
	if d.AOS {
		parts = append(parts, fmt.Sprintf("Aircraft Operational Status: %.1fs", float64(d.AOSAge)/10))
	}
	if d.TRD {
		parts = append(parts, fmt.Sprintf("Target Report Descriptor: %.1fs", float64(d.TRDAge)/10))
	}
	if d.M3A {
		parts = append(parts, fmt.Sprintf("Mode 3/A Code: %.1fs", float64(d.M3AAge)/10))
	}
	if d.QI {
		parts = append(parts, fmt.Sprintf("Quality Indicators: %.1fs", float64(d.QIAge)/10))
	}
	if d.TI {
		parts = append(parts, fmt.Sprintf("Trajectory Intent: %.1fs", float64(d.TIAge)/10))
	}
	if d.MAM {
		parts = append(parts, fmt.Sprintf("Message Amplitude: %.1fs", float64(d.MAMAge)/10))
	}
	if d.GH {
		parts = append(parts, fmt.Sprintf("Geometric Height: %.1fs", float64(d.GHAge)/10))
	}

	// First extension
	if d.FL {
		parts = append(parts, fmt.Sprintf("Flight Level: %.1fs", float64(d.FLAge)/10))
	}
	if d.ISA {
		parts = append(parts, fmt.Sprintf("Intermediate Selected Altitude: %.1fs", float64(d.ISAAge)/10))
	}
	if d.FSA {
		parts = append(parts, fmt.Sprintf("Final Selected Altitude: %.1fs", float64(d.FSAAge)/10))
	}
	if d.AS {
		parts = append(parts, fmt.Sprintf("Air Speed: %.1fs", float64(d.ASAge)/10))
	}
	if d.TAS {
		parts = append(parts, fmt.Sprintf("True Air Speed: %.1fs", float64(d.TASAge)/10))
	}
	if d.MH {
		parts = append(parts, fmt.Sprintf("Magnetic Heading: %.1fs", float64(d.MHAge)/10))
	}
	if d.BVR {
		parts = append(parts, fmt.Sprintf("Barometric Vertical Rate: %.1fs", float64(d.BVRAge)/10))
	}

	// Second extension
	if d.GVR {
		parts = append(parts, fmt.Sprintf("Geometric Vertical Rate: %.1fs", float64(d.GVRAge)/10))
	}
	if d.GV {
		parts = append(parts, fmt.Sprintf("Ground Vector: %.1fs", float64(d.GVAge)/10))
	}
	if d.TAR {
		parts = append(parts, fmt.Sprintf("Track Angle Rate: %.1fs", float64(d.TARAge)/10))
	}
	if d.TID {
		parts = append(parts, fmt.Sprintf("Target Identification: %.1fs", float64(d.TIDAge)/10))
	}
	if d.TS {
		parts = append(parts, fmt.Sprintf("Target Status: %.1fs", float64(d.TSAge)/10))
	}
	if d.MET {
		parts = append(parts, fmt.Sprintf("Met Information: %.1fs", float64(d.METAge)/10))
	}
	if d.ROA {
		parts = append(parts, fmt.Sprintf("Roll Angle: %.1fs", float64(d.ROAAge)/10))
	}

	// Third extension
	if d.ARA {
		parts = append(parts, fmt.Sprintf("ACAS Resolution Advisory: %.1fs", float64(d.ARAAge)/10))
	}
	if d.SCC {
		parts = append(parts, fmt.Sprintf("Surface Capabilities and Characteristics: %.1fs", float64(d.SCCAge)/10))
	}

	// Return empty string if no ages are available
	if len(parts) == 0 {
		return "No data ages available"
	}

	return strings.Join(parts, ", ")
}
