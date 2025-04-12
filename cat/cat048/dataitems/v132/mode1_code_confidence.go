// dataitems/cat048/mode1_code_confidence.go
package v132

import (
	"bytes"
	"fmt"
)

// Mode1CodeConfidence implements I048/065
// Confidence level for each bit of a Mode-1 reply as provided by a monopulse SSR station.
type Mode1CodeConfidence struct {
	QA4 bool // Quality pulse A4
	QA2 bool // Quality pulse A2
	QA1 bool // Quality pulse A1
	QB2 bool // Quality pulse B2
	QB1 bool // Quality pulse B1
}

// Decode implements the DataItem interface
func (m *Mode1CodeConfidence) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 1)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading Mode-1 code confidence: %w", err)
	}
	if n != 1 {
		return n, fmt.Errorf("insufficient data for Mode-1 code confidence: got %d bytes, want 1", n)
	}

	// bits 8-6 are spare
	m.QA4 = (data[0] & 0x10) != 0 // bit 5
	m.QA2 = (data[0] & 0x08) != 0 // bit 4
	m.QA1 = (data[0] & 0x04) != 0 // bit 3
	m.QB2 = (data[0] & 0x02) != 0 // bit 2
	m.QB1 = (data[0] & 0x01) != 0 // bit 1

	return n, nil
}

// Encode implements the DataItem interface
func (m *Mode1CodeConfidence) Encode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 1)

	// bits 8-6 are spare
	if m.QA4 {
		data[0] |= 0x10 // bit 5
	}
	if m.QA2 {
		data[0] |= 0x08 // bit 4
	}
	if m.QA1 {
		data[0] |= 0x04 // bit 3
	}
	if m.QB2 {
		data[0] |= 0x02 // bit 2
	}
	if m.QB1 {
		data[0] |= 0x01 // bit 1
	}

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing Mode-1 code confidence: %w", err)
	}
	return n, nil
}

// Validate implements the DataItem interface
func (m *Mode1CodeConfidence) Validate() error {
	// No validation needed for bit flags
	return nil
}

// String returns a human-readable representation
func (m *Mode1CodeConfidence) String() string {
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
	if m.QB2 {
		lowQualityPulses = append(lowQualityPulses, "B2")
	}
	if m.QB1 {
		lowQualityPulses = append(lowQualityPulses, "B1")
	}

	if len(lowQualityPulses) == 0 {
		return "All pulses high quality"
	}

	return fmt.Sprintf("Low quality pulses: %v", lowQualityPulses)
}

// HasLowQualityPulses returns true if at least one pulse is of low quality
func (m *Mode1CodeConfidence) HasLowQualityPulses() bool {
	return m.QA4 || m.QA2 || m.QA1 || m.QB2 || m.QB1
}
