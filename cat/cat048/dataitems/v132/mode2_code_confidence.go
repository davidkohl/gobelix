// dataitems/cat048/mode2_code_confidence.go
package v132

import (
	"bytes"
	"fmt"
)

// Mode2CodeConfidence implements I048/060
// Confidence level for each bit of a Mode-2 reply as provided by a monopulse SSR station.
type Mode2CodeConfidence struct {
	QA4 bool // Quality pulse A4
	QA2 bool // Quality pulse A2
	QA1 bool // Quality pulse A1
	QB4 bool // Quality pulse B4
	QB2 bool // Quality pulse B2
	QB1 bool // Quality pulse B1
	QC4 bool // Quality pulse C4
	QC2 bool // Quality pulse C2
	QC1 bool // Quality pulse C1
	QD4 bool // Quality pulse D4
	QD2 bool // Quality pulse D2
	QD1 bool // Quality pulse D1
}

// Decode implements the DataItem interface
func (m *Mode2CodeConfidence) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading Mode-2 code confidence: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("insufficient data for Mode-2 code confidence: got %d bytes, want 2", n)
	}

	// bits 16-13 are spare
	m.QA4 = (data[0] & 0x08) != 0 // bit 12
	m.QA2 = (data[0] & 0x04) != 0 // bit 11
	m.QA1 = (data[0] & 0x02) != 0 // bit 10
	m.QB4 = (data[0] & 0x01) != 0 // bit 9
	m.QB2 = (data[1] & 0x80) != 0 // bit 8
	m.QB1 = (data[1] & 0x40) != 0 // bit 7
	m.QC4 = (data[1] & 0x20) != 0 // bit 6
	m.QC2 = (data[1] & 0x10) != 0 // bit 5
	m.QC1 = (data[1] & 0x08) != 0 // bit 4
	m.QD4 = (data[1] & 0x04) != 0 // bit 3
	m.QD2 = (data[1] & 0x02) != 0 // bit 2
	m.QD1 = (data[1] & 0x01) != 0 // bit 1

	return n, nil
}

// Encode implements the DataItem interface
func (m *Mode2CodeConfidence) Encode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)

	// First byte (bits 16-9)
	// bits 16-13 are spare
	if m.QA4 {
		data[0] |= 0x08 // bit 12
	}
	if m.QA2 {
		data[0] |= 0x04 // bit 11
	}
	if m.QA1 {
		data[0] |= 0x02 // bit 10
	}
	if m.QB4 {
		data[0] |= 0x01 // bit 9
	}

	// Second byte (bits 8-1)
	if m.QB2 {
		data[1] |= 0x80 // bit 8
	}
	if m.QB1 {
		data[1] |= 0x40 // bit 7
	}
	if m.QC4 {
		data[1] |= 0x20 // bit 6
	}
	if m.QC2 {
		data[1] |= 0x10 // bit 5
	}
	if m.QC1 {
		data[1] |= 0x08 // bit 4
	}
	if m.QD4 {
		data[1] |= 0x04 // bit 3
	}
	if m.QD2 {
		data[1] |= 0x02 // bit 2
	}
	if m.QD1 {
		data[1] |= 0x01 // bit 1
	}

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing Mode-2 code confidence: %w", err)
	}
	return n, nil
}

// Validate implements the DataItem interface
func (m *Mode2CodeConfidence) Validate() error {
	// No validation needed for bit flags
	return nil
}

// String returns a human-readable representation
func (m *Mode2CodeConfidence) String() string {
	// Return low quality pulses only, since that's what matters
	var lowQualityPulses []string

	if m.QA4 {
		lowQualityPulses = append(lowQualityPulses, "A4")
	}
	if m.QA2 {
		lowQualityPulses = append(lowQualityPulses, "A2")
	}
	if m.QA1 {
		lowQualityPulses = append(lowQualityPulses, "A1")
	}
	if m.QB4 {
		lowQualityPulses = append(lowQualityPulses, "B4")
	}
	if m.QB2 {
		lowQualityPulses = append(lowQualityPulses, "B2")
	}
	if m.QB1 {
		lowQualityPulses = append(lowQualityPulses, "B1")
	}
	if m.QC4 {
		lowQualityPulses = append(lowQualityPulses, "C4")
	}
	if m.QC2 {
		lowQualityPulses = append(lowQualityPulses, "C2")
	}
	if m.QC1 {
		lowQualityPulses = append(lowQualityPulses, "C1")
	}
	if m.QD4 {
		lowQualityPulses = append(lowQualityPulses, "D4")
	}
	if m.QD2 {
		lowQualityPulses = append(lowQualityPulses, "D2")
	}
	if m.QD1 {
		lowQualityPulses = append(lowQualityPulses, "D1")
	}

	if len(lowQualityPulses) == 0 {
		return "All pulses high quality"
	}

	return fmt.Sprintf("Low quality pulses: %v", lowQualityPulses)
}

// HasLowQualityPulses returns true if at least one pulse is of low quality
func (m *Mode2CodeConfidence) HasLowQualityPulses() bool {
	return m.QA4 || m.QA2 || m.QA1 || m.QB4 || m.QB2 || m.QB1 ||
		m.QC4 || m.QC2 || m.QC1 || m.QD4 || m.QD2 || m.QD1
}
