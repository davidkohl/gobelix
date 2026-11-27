package v129

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// DataFilter represents I034/110 - Data Filter (CAT034 ed 1.29 §5.2.11).
// One octet, TYP:
//
//	0 invalid, 1 weather, 2 jamming strobe, 3 PSR, 4 SSR/Mode S,
//	5 SSR/Mode S + PSR, 6 enhanced surveillance, 7 PSR + enhanced
//	surveillance, 8 PSR + enhanced surveillance + SSR/Mode S not in area of
//	prime interest, 9 PSR + enhanced surveillance + all SSR/Mode S.
type DataFilter struct {
	TYP uint8
}

var dataFilterNames = map[uint8]string{
	0: "invalid",
	1: "weather",
	2: "jamming strobe",
	3: "PSR",
	4: "SSR/Mode S",
	5: "SSR/Mode S + PSR",
	6: "enhanced surveillance",
	7: "PSR + enhanced surveillance",
	8: "PSR + enhanced surveillance + SSR/Mode S not in area of prime interest",
	9: "PSR + enhanced surveillance + all SSR/Mode S",
}

func (d *DataFilter) Decode(buf *bytes.Buffer) (int, error) {
	b, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("%w: need 1 byte for I034/110", asterix.ErrBufferTooShort)
	}
	d.TYP = b
	return 1, nil
}

func (d *DataFilter) Encode(buf *bytes.Buffer) (int, error) {
	if err := buf.WriteByte(d.TYP); err != nil {
		return 0, fmt.Errorf("writing I034/110: %w", err)
	}
	return 1, nil
}

// Validate accepts every value: decoding is lenient, and a filter type
// added in a later edition must still round-trip.
func (d *DataFilter) Validate() error { return nil }

func (d *DataFilter) String() string {
	if name, ok := dataFilterNames[d.TYP]; ok {
		return "Filter: " + name
	}
	return fmt.Sprintf("Filter: unknown(%d)", d.TYP)
}
