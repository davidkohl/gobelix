package v129

import (
	"bytes"
	"fmt"
	"math"

	"github.com/davidkohl/gobelix/asterix"
)

// I034/100 quantisation (CAT034 ed 1.29 §5.2.10).
const (
	// PolarWindowRhoLSB is the RHO LSB: 1/256 NM (max 256 NM).
	PolarWindowRhoLSB = 1.0 / 256.0
	// PolarWindowThetaLSB is the THETA LSB: 360°/2^16.
	PolarWindowThetaLSB = 360.0 / 65536.0
)

// GenericPolarWindow represents I034/100 - Generic Polar Window (CAT034 ed
// 1.29 §5.2.10). Eight octets: RHO-START, RHO-END, THETA-START, THETA-END,
// each 16 bits unsigned.
type GenericPolarWindow struct {
	RhoStart   float64 // NM, LSB 1/256 NM, 0 .. <256
	RhoEnd     float64 // NM, LSB 1/256 NM, 0 .. <256
	ThetaStart float64 // degrees, LSB 360/2^16, 0 .. <360
	ThetaEnd   float64 // degrees, LSB 360/2^16, 0 .. <360
}

func (g *GenericPolarWindow) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 8 {
		return 0, fmt.Errorf("%w: need 8 bytes for I034/100, have %d", asterix.ErrBufferTooShort, buf.Len())
	}
	b := buf.Next(8)
	u := func(i int) float64 { return float64(uint16(b[i])<<8 | uint16(b[i+1])) }
	g.RhoStart = u(0) * PolarWindowRhoLSB
	g.RhoEnd = u(2) * PolarWindowRhoLSB
	g.ThetaStart = u(4) * PolarWindowThetaLSB
	g.ThetaEnd = u(6) * PolarWindowThetaLSB
	return 8, nil
}

func (g *GenericPolarWindow) raw() [4]float64 {
	return [4]float64{
		g.RhoStart / PolarWindowRhoLSB, g.RhoEnd / PolarWindowRhoLSB,
		g.ThetaStart / PolarWindowThetaLSB, g.ThetaEnd / PolarWindowThetaLSB,
	}
}

func (g *GenericPolarWindow) Encode(buf *bytes.Buffer) (int, error) {
	if err := g.Validate(); err != nil {
		return 0, err
	}
	var out [8]byte
	for i, v := range g.raw() {
		r := uint16(math.Round(v))
		out[2*i] = byte(r >> 8)
		out[2*i+1] = byte(r)
	}
	n, err := buf.Write(out[:])
	if err != nil {
		return n, fmt.Errorf("writing I034/100: %w", err)
	}
	return n, nil
}

func (g *GenericPolarWindow) Validate() error {
	names := [4]string{"RHO-START", "RHO-END", "THETA-START", "THETA-END"}
	for i, v := range g.raw() {
		r := math.Round(v)
		if math.IsNaN(r) || r < 0 || r > math.MaxUint16 {
			return fmt.Errorf("%w: I034/100 %s out of the unsigned 16-bit range", asterix.ErrInvalidMessage, names[i])
		}
	}
	return nil
}

func (g *GenericPolarWindow) String() string {
	return fmt.Sprintf("rho %.3f-%.3f NM, theta %.3f-%.3f°", g.RhoStart, g.RhoEnd, g.ThetaStart, g.ThetaEnd)
}
