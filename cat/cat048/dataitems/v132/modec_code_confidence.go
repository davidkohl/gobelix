// dataitems/cat048/modec_code_confidence.go
package v132

import (
	"bytes"
	"fmt"
)

// ModeCCodeAndConfidence implements I048/100
// Mode-C height in Gray notation as received from the transponder together with
// the confidence level for each reply bit as provided by a MSSR/Mode S station.
type ModeCCodeAndConfidence struct {
	V    bool   // Code validated
	G    bool   // Garbled code
	Code uint16 // Raw Mode-C code in Gray code format

	// Quality bits for each pulse
	QC1 bool // Quality of pulse C1
	QA1 bool // Quality of pulse A1
	QC2 bool // Quality of pulse C2
	QA2 bool // Quality of pulse A2
	QC4 bool // Quality of pulse C4
	QA4 bool // Quality of pulse A4
	QB1 bool // Quality of pulse B1
	QD1 bool // Quality of pulse D1
	QB2 bool // Quality of pulse B2
	QD2 bool // Quality of pulse D2
	QB4 bool // Quality of pulse B4
	QD4 bool // Quality of pulse D4
}

// Decode implements the DataItem interface
func (m *ModeCCodeAndConfidence) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 4)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading Mode-C code and confidence: %w", err)
	}
	if n != 4 {
		return n, fmt.Errorf("insufficient data for Mode-C code and confidence: got %d bytes, want 4", n)
	}

	// First byte contains flags
	m.V = (data[0] & 0x80) != 0 // bit 32
	m.G = (data[0] & 0x40) != 0 // bit 31
	// bits 30-29 are spare

	// Next 12 bits contain the Mode-C code in a specific order
	c1 := (data[0] & 0x08) != 0 // bit 28
	a1 := (data[0] & 0x04) != 0 // bit 27
	c2 := (data[0] & 0x02) != 0 // bit 26
	a2 := (data[0] & 0x01) != 0 // bit 25
	c4 := (data[1] & 0x80) != 0 // bit 24
	a4 := (data[1] & 0x40) != 0 // bit 23
	b1 := (data[1] & 0x20) != 0 // bit 22
	d1 := (data[1] & 0x10) != 0 // bit 21
	b2 := (data[1] & 0x08) != 0 // bit 20
	d2 := (data[1] & 0x04) != 0 // bit 19
	b4 := (data[1] & 0x02) != 0 // bit 18
	d4 := (data[1] & 0x01) != 0 // bit 17

	// Assemble the code into a 12-bit value
	m.Code = 0
	if c1 {
		m.Code |= 0x800 // bit 12
	}
	if a1 {
		m.Code |= 0x400 // bit 11
	}
	if c2 {
		m.Code |= 0x200 // bit 10
	}
	if a2 {
		m.Code |= 0x100 // bit 9
	}
	if c4 {
		m.Code |= 0x080 // bit 8
	}
	if a4 {
		m.Code |= 0x040 // bit 7
	}
	if b1 {
		m.Code |= 0x020 // bit 6
	}
	if d1 {
		m.Code |= 0x010 // bit 5
	}
	if b2 {
		m.Code |= 0x008 // bit 4
	}
	if d2 {
		m.Code |= 0x004 // bit 3
	}
	if b4 {
		m.Code |= 0x002 // bit 2
	}
	if d4 {
		m.Code |= 0x001 // bit 1
	}

	// Third and fourth bytes contain quality bits
	// bits 16-13 are spare
	m.QC1 = (data[2] & 0x08) != 0 // bit 12
	m.QA1 = (data[2] & 0x04) != 0 // bit 11
	m.QC2 = (data[2] & 0x02) != 0 // bit 10
	m.QA2 = (data[2] & 0x01) != 0 // bit 9
	m.QC4 = (data[3] & 0x80) != 0 // bit 8
	m.QA4 = (data[3] & 0x40) != 0 // bit 7
	m.QB1 = (data[3] & 0x20) != 0 // bit 6
	m.QD1 = (data[3] & 0x10) != 0 // bit 5
	m.QB2 = (data[3] & 0x08) != 0 // bit 4
	m.QD2 = (data[3] & 0x04) != 0 // bit 3
	m.QB4 = (data[3] & 0x02) != 0 // bit 2
	m.QD4 = (data[3] & 0x01) != 0 // bit 1

	return n, nil
}

// Encode implements the DataItem interface
func (m *ModeCCodeAndConfidence) Encode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 4)

	// First byte - flags and first part of Mode-C code
	if m.V {
		data[0] |= 0x80 // bit 32
	}
	if m.G {
		data[0] |= 0x40 // bit 31
	}
	// bits 30-29 are spare
	if (m.Code & 0x800) != 0 { // C1 (bit 12)
		data[0] |= 0x08 // bit 28
	}
	if (m.Code & 0x400) != 0 { // A1 (bit 11)
		data[0] |= 0x04 // bit 27
	}
	if (m.Code & 0x200) != 0 { // C2 (bit 10)
		data[0] |= 0x02 // bit 26
	}
	if (m.Code & 0x100) != 0 { // A2 (bit 9)
		data[0] |= 0x01 // bit 25
	}

	// Second byte - rest of Mode-C code
	if (m.Code & 0x080) != 0 { // C4 (bit 8)
		data[1] |= 0x80 // bit 24
	}
	if (m.Code & 0x040) != 0 { // A4 (bit 7)
		data[1] |= 0x40 // bit 23
	}
	if (m.Code & 0x020) != 0 { // B1 (bit 6)
		data[1] |= 0x20 // bit 22
	}
	if (m.Code & 0x010) != 0 { // D1 (bit 5)
		data[1] |= 0x10 // bit 21
	}
	if (m.Code & 0x008) != 0 { // B2 (bit 4)
		data[1] |= 0x08 // bit 20
	}
	if (m.Code & 0x004) != 0 { // D2 (bit 3)
		data[1] |= 0x04 // bit 19
	}
	if (m.Code & 0x002) != 0 { // B4 (bit 2)
		data[1] |= 0x02 // bit 18
	}
	if (m.Code & 0x001) != 0 { // D4 (bit 1)
		data[1] |= 0x01 // bit 17
	}

	// Third byte - first part of quality bits
	// bits 16-13 are spare
	if m.QC1 {
		data[2] |= 0x08 // bit 12
	}
	if m.QA1 {
		data[2] |= 0x04 // bit 11
	}
	if m.QC2 {
		data[2] |= 0x02 // bit 10
	}
	if m.QA2 {
		data[2] |= 0x01 // bit 9
	}

	// Fourth byte - rest of quality bits
	if m.QC4 {
		data[3] |= 0x80 // bit 8
	}
	if m.QA4 {
		data[3] |= 0x40 // bit 7
	}
	if m.QB1 {
		data[3] |= 0x20 // bit 6
	}
	if m.QD1 {
		data[3] |= 0x10 // bit 5
	}
	if m.QB2 {
		data[3] |= 0x08 // bit 4
	}
	if m.QD2 {
		data[3] |= 0x04 // bit 3
	}
	if m.QB4 {
		data[3] |= 0x02 // bit 2
	}
	if m.QD4 {
		data[3] |= 0x01 // bit 1
	}

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing Mode-C code and confidence: %w", err)
	}
	return n, nil
}

// Validate implements the DataItem interface
func (m *ModeCCodeAndConfidence) Validate() error {
	if m.Code > 0xFFF { // Ensure it fits in 12 bits
		return fmt.Errorf("Mode-C code too large (exceeds 12 bits): %X", m.Code)
	}
	return nil
}

// String returns a human-readable representation
func (m *ModeCCodeAndConfidence) String() string {
	flags := ""
	if m.V {
		flags += "V,"
	}
	if m.G {
		flags += "G,"
	}

	if flags != "" {
		flags = flags[:len(flags)-1] + " " // Remove trailing comma
	}

	// Get low quality pulses
	var lowQualityPulses []string
	if m.QC1 {
		lowQualityPulses = append(lowQualityPulses, "C1")
	}
	if m.QA1 {
		lowQualityPulses = append(lowQualityPulses, "A1")
	}
	if m.QC2 {
		lowQualityPulses = append(lowQualityPulses, "C2")
	}
	if m.QA2 {
		lowQualityPulses = append(lowQualityPulses, "A2")
	}
	if m.QC4 {
		lowQualityPulses = append(lowQualityPulses, "C4")
	}
	if m.QA4 {
		lowQualityPulses = append(lowQualityPulses, "A4")
	}
	if m.QB1 {
		lowQualityPulses = append(lowQualityPulses, "B1")
	}
	if m.QD1 {
		lowQualityPulses = append(lowQualityPulses, "D1")
	}
	if m.QB2 {
		lowQualityPulses = append(lowQualityPulses, "B2")
	}
	if m.QD2 {
		lowQualityPulses = append(lowQualityPulses, "D2")
	}
	if m.QB4 {
		lowQualityPulses = append(lowQualityPulses, "B4")
	}
	if m.QD4 {
		lowQualityPulses = append(lowQualityPulses, "D4")
	}

	qualityStr := "All pulses high quality"
	if len(lowQualityPulses) > 0 {
		qualityStr = fmt.Sprintf("Low quality: %v", lowQualityPulses)
	}

	return fmt.Sprintf("%sGray Code: %03X, %s", flags, m.Code, qualityStr)
}

// HasLowQualityPulses returns true if at least one pulse is of low quality
func (m *ModeCCodeAndConfidence) HasLowQualityPulses() bool {
	return m.QC1 || m.QA1 || m.QC2 || m.QA2 || m.QC4 || m.QA4 ||
		m.QB1 || m.QD1 || m.QB2 || m.QD2 || m.QB4 || m.QD4
}
