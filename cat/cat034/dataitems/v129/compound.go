package v129

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// Primary subfield bits shared by I034/050 and I034/060 (ed 1.29 §5.2.6,
// §5.2.7): COM(8) spare(7) spare(6) PSR(5) SSR(4) MDS(3) spare(2) FX(1).
const (
	primaryCOM   = 0x80
	primaryPSR   = 0x10
	primarySSR   = 0x08
	primaryMDS   = 0x04
	primarySpare = 0x62
)

// readPrimary reads the extensible primary subfield of I034/050 or
// I034/060. A spare subfield bit, or any subfield bit in an extension
// octet, announces a subfield whose length ed 1.29 does not define, so the
// rest of the record could not be decoded: that is an error.
func readPrimary(buf *bytes.Buffer, item string) (byte, int, error) {
	first, err := buf.ReadByte()
	if err != nil {
		return 0, 0, fmt.Errorf("%w: reading %s primary subfield", asterix.ErrBufferTooShort, item)
	}
	n := 1
	if first&primarySpare != 0 {
		return 0, n, fmt.Errorf("%w: %s primary subfield %#02x announces a spare subfield", asterix.ErrInvalidField, item, first)
	}
	for b := first; b&0x01 != 0; {
		b, err = buf.ReadByte()
		if err != nil {
			return 0, n, fmt.Errorf("%w: reading %s primary subfield extension", asterix.ErrBufferTooShort, item)
		}
		n++
		if b&0xFE != 0 {
			return 0, n, fmt.Errorf("%w: %s primary subfield extension %#02x announces an undefined subfield", asterix.ErrInvalidField, item, b)
		}
	}
	return first, n, nil
}

func readOctet(buf *bytes.Buffer, item, sub string) (byte, error) {
	b, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("%w: reading %s %s subfield", asterix.ErrBufferTooShort, item, sub)
	}
	return b, nil
}
