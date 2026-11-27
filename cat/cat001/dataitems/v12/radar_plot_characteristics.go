package v12

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/davidkohl/gobelix/asterix"
)

// RadarPlotCharacteristics represents I001/130 - Radar Plot Characteristics
// (CAT001 ed 1.4 §5.2.13). Variable length: each octet carries a 7-bit
// indicator in bits 8-2 and FX in bit 1. The meaning of the indicators is
// application dependent ("its contents being defined between the users
// concerned", §5.3.2.6), so the item exposes only the raw indicators.
//
// The earlier SRL..APD fields, a guess at the layout that CAT048 I048/130
// later standardised, were removed: they are not part of CAT001.
type RadarPlotCharacteristics struct {
	// Indicators holds the 7-bit indicator of every octet in order
	// (at least one).
	Indicators []uint8
}

func (r *RadarPlotCharacteristics) Decode(buf *bytes.Buffer) (int, error) {
	values, n, err := decodeWEValues(buf)
	if err != nil {
		return n, fmt.Errorf("%w: incomplete radar plot characteristics", asterix.ErrBufferTooShort)
	}
	r.Indicators = values
	return n, nil
}

func (r *RadarPlotCharacteristics) Encode(buf *bytes.Buffer) (int, error) {
	if err := r.Validate(); err != nil {
		return 0, err
	}
	return encodeWEValues(buf, r.Indicators)
}

func (r *RadarPlotCharacteristics) String() string {
	parts := make([]string, len(r.Indicators))
	for i, v := range r.Indicators {
		parts[i] = fmt.Sprintf("%02X", v)
	}
	return "Plot characteristics: " + strings.Join(parts, " ")
}

func (r *RadarPlotCharacteristics) Validate() error {
	if len(r.Indicators) == 0 {
		return fmt.Errorf("%w: I001/130 needs at least one indicator octet", asterix.ErrInvalidMessage)
	}
	for i, v := range r.Indicators {
		if v > 127 {
			return fmt.Errorf("%w: I001/130 indicator %d is 7 bits, got %d", asterix.ErrInvalidMessage, i, v)
		}
	}
	return nil
}
