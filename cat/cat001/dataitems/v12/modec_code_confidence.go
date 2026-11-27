package v12

import (
	"bytes"
	"fmt"
	"math"
	"sync"

	"github.com/davidkohl/gobelix/asterix"
)

// ModeCCodeConfidence represents I001/100 - Mode-C Code and Code Confidence
// Indicator (CAT001 ed 1.4 §5.2.11). Four octets:
//
//	V(32) G(31) spare(30-29) C1 A1 C2 A2 C4 A4 B1 D1 B2 D2 B4 D4 (28-17)
//	spare(16-13) QC1 QA1 QC2 QA2 QC4 QA4 QB1 QD1 QB2 QD2 QB4 QD4 (12-1)
//
// Bits 28-17 are the Mode-C reply in Gray (Gillham) notation, not a binary
// flight level.
type ModeCCodeConfidence struct {
	V bool // Code not validated
	G bool // Garbled code

	// GrayCode is the 12-bit Mode-C reply exactly as in bits 28-17
	// (bit 12 = C1 ... bit 1 = D4). When it is zero, Encode derives it from
	// FlightLevel.
	GrayCode uint16
	// FlightLevel is the altitude decoded from GrayCode in FL (100 ft).
	// It is only meaningful when AltitudeValid is true.
	FlightLevel float64
	// AltitudeValid reports whether GrayCode is a valid Gillham code.
	AltitudeValid bool

	QC1 bool // Quality C1
	QA1 bool // Quality A1
	QC2 bool // Quality C2
	QA2 bool // Quality A2
	QC4 bool // Quality C4
	QA4 bool // Quality A4
	QB1 bool // Quality B1
	QD1 bool // Quality D1
	QB2 bool // Quality B2
	QD2 bool // Quality D2
	QB4 bool // Quality B4
	QD4 bool // Quality D4
}

// Bit positions of the Mode-C pulses in the 12-bit I001/100 code field.
const (
	grayC1 = 1 << 11
	grayA1 = 1 << 10
	grayC2 = 1 << 9
	grayA2 = 1 << 8
	grayC4 = 1 << 7
	grayA4 = 1 << 6
	grayB1 = 1 << 5
	grayD1 = 1 << 4
	grayB2 = 1 << 3
	grayD2 = 1 << 2
	grayB4 = 1 << 1
	grayD4 = 1 << 0
)

// GillhamToAltitude decodes a Mode-C reply in the I001/100 bit order
// (C1 A1 C2 A2 C4 A4 B1 D1 B2 D2 B4 D4, MSB first) to an altitude in feet.
// It returns false for codes that are not valid Gillham altitudes.
//
// The 500 ft increments are a reflected Gray code over D2 D4 A1 A2 A4 B1 B2
// B4; the 100 ft increments are the five-state code over C1 C2 C4, reversed
// when the 500 ft count is odd (ICAO Annex 10 Vol IV, Mode C altitude
// encoding). D1 is never used for altitude.
func GillhamToAltitude(code uint16) (int, bool) {
	if code&grayD1 != 0 {
		return 0, false
	}
	var hundreds uint
	if code&grayC1 != 0 {
		hundreds ^= 7
	}
	if code&grayC2 != 0 {
		hundreds ^= 3
	}
	if code&grayC4 != 0 {
		hundreds ^= 1
	}
	if hundreds&5 == 5 {
		hundreds ^= 2
	}
	if hundreds == 0 || hundreds > 5 {
		return 0, false
	}

	var fiveHundreds uint
	for _, p := range []struct {
		bit  uint16
		mask uint
	}{
		{grayD2, 0xFF}, {grayD4, 0x7F}, {grayA1, 0x3F}, {grayA2, 0x1F},
		{grayA4, 0x0F}, {grayB1, 0x07}, {grayB2, 0x03}, {grayB4, 0x01},
	} {
		if code&p.bit != 0 {
			fiveHundreds ^= p.mask
		}
	}
	if fiveHundreds&1 != 0 {
		hundreds = 6 - hundreds
	}
	return (int(fiveHundreds)*5 + int(hundreds) - 13) * 100, true
}

var (
	gillhamOnce  sync.Once
	gillhamTable map[int]uint16
)

// AltitudeToGillham encodes an altitude in feet (a multiple of 100 ft) to a
// Mode-C reply in the I001/100 bit order. It returns false when the
// altitude cannot be represented.
func AltitudeToGillham(altitudeFt int) (uint16, bool) {
	gillhamOnce.Do(func() {
		gillhamTable = make(map[int]uint16)
		for code := uint16(0); code < 1<<12; code++ {
			if alt, ok := GillhamToAltitude(code); ok {
				if _, seen := gillhamTable[alt]; !seen {
					gillhamTable[alt] = code
				}
			}
		}
	})
	code, ok := gillhamTable[altitudeFt]
	return code, ok
}

func (m *ModeCCodeConfidence) Decode(buf *bytes.Buffer) (int, error) {
	var data [4]byte
	n, err := buf.Read(data[:])
	if err != nil || n != 4 {
		return n, fmt.Errorf("%w: need 4 bytes for Mode-C code confidence, have %d", asterix.ErrBufferTooShort, n)
	}

	m.V = data[0]&0x80 != 0
	m.G = data[0]&0x40 != 0
	m.GrayCode = uint16(data[0]&0x0F)<<8 | uint16(data[1])
	alt, ok := GillhamToAltitude(m.GrayCode)
	m.AltitudeValid = ok
	m.FlightLevel = 0
	if ok {
		m.FlightLevel = float64(alt) / 100.0
	}

	m.QC1 = data[2]&0x08 != 0
	m.QA1 = data[2]&0x04 != 0
	m.QC2 = data[2]&0x02 != 0
	m.QA2 = data[2]&0x01 != 0
	m.QC4 = data[3]&0x80 != 0
	m.QA4 = data[3]&0x40 != 0
	m.QB1 = data[3]&0x20 != 0
	m.QD1 = data[3]&0x10 != 0
	m.QB2 = data[3]&0x08 != 0
	m.QD2 = data[3]&0x04 != 0
	m.QB4 = data[3]&0x02 != 0
	m.QD4 = data[3]&0x01 != 0

	return 4, nil
}

// code returns the Gray code to transmit.
func (m *ModeCCodeConfidence) code() (uint16, error) {
	if m.GrayCode != 0 {
		return m.GrayCode, nil
	}
	if m.FlightLevel == 0 && !m.AltitudeValid {
		return 0, nil
	}
	ft := m.FlightLevel * 100
	if math.Abs(ft-math.Round(ft/100)*100) > 1e-6 {
		return 0, fmt.Errorf("%w: Mode-C altitude %.2f FL is not a multiple of 100 ft", asterix.ErrInvalidMessage, m.FlightLevel)
	}
	code, ok := AltitudeToGillham(int(math.Round(ft)))
	if !ok {
		return 0, fmt.Errorf("%w: Mode-C altitude %.0f ft cannot be Gillham encoded", asterix.ErrInvalidMessage, ft)
	}
	return code, nil
}

func (m *ModeCCodeConfidence) Encode(buf *bytes.Buffer) (int, error) {
	if err := m.Validate(); err != nil {
		return 0, err
	}
	code, err := m.code()
	if err != nil {
		return 0, err
	}

	octet1 := uint8(code>>8) & 0x0F
	if m.V {
		octet1 |= 0x80
	}
	if m.G {
		octet1 |= 0x40
	}
	octet2 := uint8(code)

	var octet3, octet4 uint8
	for _, q := range []struct {
		set bool
		oct *uint8
		bit uint8
	}{
		{m.QC1, &octet3, 0x08}, {m.QA1, &octet3, 0x04}, {m.QC2, &octet3, 0x02}, {m.QA2, &octet3, 0x01},
		{m.QC4, &octet4, 0x80}, {m.QA4, &octet4, 0x40}, {m.QB1, &octet4, 0x20}, {m.QD1, &octet4, 0x10},
		{m.QB2, &octet4, 0x08}, {m.QD2, &octet4, 0x04}, {m.QB4, &octet4, 0x02}, {m.QD4, &octet4, 0x01},
	} {
		if q.set {
			*q.oct |= q.bit
		}
	}

	n, err := buf.Write([]byte{octet1, octet2, octet3, octet4})
	if err != nil {
		return n, fmt.Errorf("writing Mode-C code confidence: %w", err)
	}
	return n, nil
}

func (m *ModeCCodeConfidence) String() string {
	flags := ""
	if m.V {
		flags += " NOT_VALIDATED"
	}
	if m.G {
		flags += " GARBLED"
	}
	if !m.AltitudeValid && m.GrayCode != 0 {
		return fmt.Sprintf("Gray=%04o (invalid)%s", m.GrayCode, flags)
	}
	return fmt.Sprintf("FL%.0f%s", m.FlightLevel, flags)
}

func (m *ModeCCodeConfidence) Validate() error {
	if m.GrayCode > 0x0FFF {
		return fmt.Errorf("%w: I001/100 Mode-C code is 12 bits, got %#x", asterix.ErrInvalidMessage, m.GrayCode)
	}
	return nil
}
