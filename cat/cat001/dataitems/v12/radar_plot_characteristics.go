package v12

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// RadarPlotCharacteristics represents I001/130 - Radar Plot Characteristics
// (CAT001 ed 1.4 §5.2.13). Variable length: each octet carries a 7-bit
// indicator in bits 8-2 and FX in bit 1. The meaning of the indicators is
// application dependent, so the spec-level content is Indicators.
//
// SRL..APD are a convenience view of the first seven octets in the layout
// that CAT048 I048/130 later standardised. They are filled on Decode and
// used by Encode only when Indicators is empty.
type RadarPlotCharacteristics struct {
	SRL uint8 // Octet 1, bits 7-2: SSR plot runlength
	SRR uint8 // Octet 2, bits 7-2: number of received replies for M(SSR)
	SAM int8  // Octet 3, 7-bit signed: amplitude of M(SSR) reply
	PRL uint8 // Octet 4: primary plot runlength
	PAM int8  // Octet 5, 7-bit signed: amplitude of primary plot
	RPD int8  // Octet 6, 7-bit signed: difference in range between PSR and SSR
	APD int8  // Octet 7, 7-bit signed: difference in azimuth between PSR and SSR

	// Indicators holds the 7-bit indicator of every octet in order.
	Indicators []uint8
}

func signExtend7(v uint8) int8 {
	if v&0x40 != 0 {
		return int8(int16(v) - 128)
	}
	return int8(v)
}

func (r *RadarPlotCharacteristics) Decode(buf *bytes.Buffer) (int, error) {
	values, n, err := decodeWEValues(buf)
	if err != nil {
		return n, fmt.Errorf("%w: incomplete radar plot characteristics", asterix.ErrBufferTooShort)
	}
	*r = RadarPlotCharacteristics{Indicators: values}
	at := func(i int) uint8 {
		if i < len(values) {
			return values[i]
		}
		return 0
	}
	r.SRL = at(0) & 0x3F
	r.SRR = at(1) & 0x3F
	r.SAM = signExtend7(at(2))
	r.PRL = at(3)
	r.PAM = signExtend7(at(4))
	r.RPD = signExtend7(at(5))
	r.APD = signExtend7(at(6))
	return n, nil
}

func (r *RadarPlotCharacteristics) Encode(buf *bytes.Buffer) (int, error) {
	if err := r.Validate(); err != nil {
		return 0, err
	}
	if len(r.Indicators) > 0 {
		return encodeWEValues(buf, r.Indicators)
	}

	values := []uint8{
		r.SRL & 0x3F,
		r.SRR & 0x3F,
		uint8(r.SAM) & 0x7F,
		r.PRL & 0x7F,
		uint8(r.PAM) & 0x7F,
		uint8(r.RPD) & 0x7F,
		uint8(r.APD) & 0x7F,
	}
	// Drop trailing zero octets, but keep the first part.
	for len(values) > 1 && values[len(values)-1] == 0 {
		values = values[:len(values)-1]
	}
	return encodeWEValues(buf, values)
}

func (r *RadarPlotCharacteristics) String() string {
	return fmt.Sprintf("SRL=%d SRR=%d SAM=%d PRL=%d PAM=%d RPD=%d APD=%d",
		r.SRL, r.SRR, r.SAM, r.PRL, r.PAM, r.RPD, r.APD)
}

func (r *RadarPlotCharacteristics) Validate() error {
	for i, v := range r.Indicators {
		if v > 127 {
			return fmt.Errorf("%w: I001/130 indicator %d is 7 bits, got %d", asterix.ErrInvalidMessage, i, v)
		}
	}
	if r.SAM < -64 || r.PAM < -64 || r.RPD < -64 || r.APD < -64 ||
		r.SAM > 63 || r.PAM > 63 || r.RPD > 63 || r.APD > 63 {
		return fmt.Errorf("%w: I001/130 signed indicators are 7 bits (-64..63)", asterix.ErrInvalidMessage)
	}
	return nil
}
