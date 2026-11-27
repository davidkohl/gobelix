package v10

import (
	"bytes"
	"fmt"
	"math"

	"github.com/davidkohl/gobelix/asterix"
)

// I002/100 quantisation (CAT002 ed 1.2 §5.2.11).
const (
	// DynamicWindowRhoLSB is the RHO LSB: 1/128 NM (max 512 NM).
	DynamicWindowRhoLSB = 1.0 / 128.0
	// DynamicWindowThetaLSB is the THETA LSB: 360°/2^16.
	DynamicWindowThetaLSB = 360.0 / 65536.0
)

// DynamicWindow represents I002/100 - Dynamic Window - Type 1 (CAT002 ed 1.2
// §5.2.11). Eight octets: RHO-START, RHO-END, THETA-START, THETA-END, each
// 16 bits unsigned. The meaning of the window is given by I002/000.
type DynamicWindow struct {
	RhoStart   float64 // NM, LSB 1/128 NM, 0 .. <512
	RhoEnd     float64 // NM, LSB 1/128 NM, 0 .. <512
	ThetaStart float64 // degrees, LSB 360/2^16, 0 .. <360
	ThetaEnd   float64 // degrees, LSB 360/2^16, 0 .. <360
}

func (d *DynamicWindow) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 8 {
		return 0, fmt.Errorf("%w: need 8 bytes for dynamic window, have %d", asterix.ErrBufferTooShort, buf.Len())
	}
	b := buf.Next(8)
	u := func(i int) float64 { return float64(uint16(b[i])<<8 | uint16(b[i+1])) }
	d.RhoStart = u(0) * DynamicWindowRhoLSB
	d.RhoEnd = u(2) * DynamicWindowRhoLSB
	d.ThetaStart = u(4) * DynamicWindowThetaLSB
	d.ThetaEnd = u(6) * DynamicWindowThetaLSB
	return 8, nil
}

func (d *DynamicWindow) Encode(buf *bytes.Buffer) (int, error) {
	if err := d.Validate(); err != nil {
		return 0, err
	}
	var out [8]byte
	for i, v := range []float64{
		d.RhoStart / DynamicWindowRhoLSB, d.RhoEnd / DynamicWindowRhoLSB,
		d.ThetaStart / DynamicWindowThetaLSB, d.ThetaEnd / DynamicWindowThetaLSB,
	} {
		r := uint16(math.Round(v))
		out[2*i] = byte(r >> 8)
		out[2*i+1] = byte(r)
	}
	n, err := buf.Write(out[:])
	if err != nil {
		return n, fmt.Errorf("writing dynamic window: %w", err)
	}
	return n, nil
}

func (d *DynamicWindow) Validate() error {
	for _, f := range []struct {
		name string
		raw  float64
	}{
		{"RHO-START", d.RhoStart / DynamicWindowRhoLSB},
		{"RHO-END", d.RhoEnd / DynamicWindowRhoLSB},
		{"THETA-START", d.ThetaStart / DynamicWindowThetaLSB},
		{"THETA-END", d.ThetaEnd / DynamicWindowThetaLSB},
	} {
		r := math.Round(f.raw)
		if math.IsNaN(r) || r < 0 || r > math.MaxUint16 {
			return fmt.Errorf("%w: I002/100 %s out of the unsigned 16-bit range", asterix.ErrInvalidMessage, f.name)
		}
	}
	return nil
}

func (d *DynamicWindow) String() string {
	return fmt.Sprintf("rho %.3f-%.3f NM, theta %.3f-%.3f°", d.RhoStart, d.RhoEnd, d.ThetaStart, d.ThetaEnd)
}
