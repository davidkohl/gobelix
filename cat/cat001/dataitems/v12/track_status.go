package v12

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// TrackStatus represents I001/170 - Track Status
// Variable length data item with FX extension mechanism
type TrackStatus struct {
	CON  bool // Confirmed track (0) / Track in initialization phase (1)
	RAD  bool // Primary track (0) / SSR/Combined track (1)
	MAN  bool // Aircraft manoeuv ring
	DOU  bool // Doubtful plot to track association
	RDPC bool // Radar Data Processing Chain (0=Chain 1, 1=Chain 2)
	GHO  bool // Ghost track
	TRE  bool // First extent: last report for a track (CAT001 ed 1.4 §5.2.18)
}

func (t *TrackStatus) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// First octet
	data, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("%w: need at least 1 byte for track status", asterix.ErrBufferTooShort)
	}
	bytesRead++

	t.CON = (data & 0x80) != 0  // bit 8
	t.RAD = (data & 0x40) != 0  // bit 7
	t.MAN = (data & 0x20) != 0  // bit 6
	t.DOU = (data & 0x10) != 0  // bit 5
	t.RDPC = (data & 0x08) != 0 // bit 4
	// bit 3 is spare
	t.GHO = (data & 0x02) != 0 // bit 2

	t.TRE = false

	// Check FX bit for extension
	hasFX := (data & 0x01) != 0

	// First extent: TRE (bit 8), spare (7-2), FX
	if hasFX {
		data, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("%w: incomplete track status extension", asterix.ErrBufferTooShort)
		}
		bytesRead++
		t.TRE = data&0x80 != 0
		hasFX = (data & 0x01) != 0
	}

	// Further extents are not defined in ed 1.4; skip them.
	for hasFX {
		data, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("%w: incomplete track status extension", asterix.ErrBufferTooShort)
		}
		bytesRead++
		hasFX = (data & 0x01) != 0
	}

	return bytesRead, nil
}

func (t *TrackStatus) Encode(buf *bytes.Buffer) (int, error) {
	// First octet
	octet := uint8(0)
	if t.CON {
		octet |= 0x80 // bit 8
	}
	if t.RAD {
		octet |= 0x40 // bit 7
	}
	if t.MAN {
		octet |= 0x20 // bit 6
	}
	if t.DOU {
		octet |= 0x10 // bit 5
	}
	if t.RDPC {
		octet |= 0x08 // bit 4
	}
	if t.GHO {
		octet |= 0x02 // bit 2
	}
	if t.TRE {
		octet |= 0x01 // FX: first extent follows
	}

	if err := buf.WriteByte(octet); err != nil {
		return 0, fmt.Errorf("writing track status: %w", err)
	}
	if !t.TRE {
		return 1, nil
	}
	if err := buf.WriteByte(0x80); err != nil {
		return 1, fmt.Errorf("writing track status extent: %w", err)
	}
	return 2, nil
}

func (t *TrackStatus) Validate() error {
	return nil
}

func (t *TrackStatus) String() string {
	result := "Track Status:"
	if t.CON {
		result += " INIT"
	} else {
		result += " CONFIRMED"
	}
	if t.RAD {
		result += " SSR/COMBINED"
	} else {
		result += " PRIMARY"
	}
	if t.MAN {
		result += " MANOEUVRING"
	}
	if t.DOU {
		result += " DOUBTFUL"
	}
	if t.RDPC {
		result += " CHAIN2"
	} else {
		result += " CHAIN1"
	}
	if t.GHO {
		result += " GHOST"
	}
	if t.TRE {
		result += " END"
	}
	return result
}
