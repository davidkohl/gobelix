package v12

import (
	"bytes"
	"math"
	"testing"
)

// TestEncodeRounds checks that I001/040, I001/090 and I001/141 round to
// the nearest LSB, half away from zero, instead of truncating; azimuth
// and truncated time wrap.
func TestEncodeRounds(t *testing.T) {
	enc := func(in interface {
		Encode(*bytes.Buffer) (int, error)
	}, out interface {
		Decode(*bytes.Buffer) (int, error)
	}) {
		t.Helper()
		var b bytes.Buffer
		if _, err := in.Encode(&b); err != nil {
			t.Fatal(err)
		}
		if _, err := out.Decode(&b); err != nil {
			t.Fatal(err)
		}
	}
	for _, c := range []struct{ rho, theta, wantRho, wantTheta float64 }{
		{10.006, 90.004, 10.0078125, 90.0054931640625},
		{10.002, 359.999, 10, 0},
	} {
		var o PositionPolar
		enc(&PositionPolar{RHO: c.rho, THETA: c.theta}, &o)
		if o.RHO != c.wantRho || o.THETA != c.wantTheta {
			t.Errorf("polar %v %v gives %v %v, want %v %v", c.rho, c.theta, o.RHO, o.THETA, c.wantRho, c.wantTheta)
		}
	}
	for _, v := range []float64{350.2, 350.125, -0.15, -10.2} {
		var o ModeCCode
		enc(&ModeCCode{V: true, FlightLevel: v}, &o)
		if want := math.Round(v*4) / 4; o.FlightLevel != want {
			t.Errorf("Mode C %v gives %v, want %v", v, o.FlightLevel, want)
		}
	}
	for _, v := range []float64{100.006, 100.002} {
		var o TruncatedTimeOfDay
		enc(&TruncatedTimeOfDay{Time: v}, &o)
		if want := math.Round(v*128) / 128; o.Time != want {
			t.Errorf("truncated time %v gives %v, want %v", v, o.Time, want)
		}
	}
}
