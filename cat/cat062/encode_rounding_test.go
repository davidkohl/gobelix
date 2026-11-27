// cat/cat062/encode_rounding_test.go
//
// The CAT 062 encoders round a physical value to the nearest LSB, half
// away from zero, instead of truncating it towards zero.
package cat062_test

import (
	"bytes"
	"math"
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	v117 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v117"
	v120 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v120"
)

func f64(v float64) *float64 { return &v }

// roundTrip encodes in, decodes into out and returns get(out).
func roundTrip(t *testing.T, in, out asterix.DataItem) {
	t.Helper()
	var b bytes.Buffer
	if _, err := in.Encode(&b); err != nil {
		t.Fatalf("%T: encode: %v", in, err)
	}
	if _, err := out.Decode(&b); err != nil {
		t.Fatalf("%T: decode: %v", out, err)
	}
}

// nearest is the value v rounded to a multiple of lsb, half away from
// zero.
func nearest(v, lsb float64) float64 { return math.Round(v/lsb) * lsb }

func TestEncodeRounds(t *testing.T) {
	type tc struct {
		name string
		lsb  float64
		v    float64
		got  func(v float64) float64
	}
	var cases []tc
	for _, v := range []float64{350.2, 350.1, 350.125, -10.125, -10.2, -0.1} {
		cases = append(cases,
			tc{"v120 I062/135", 0.25, v, func(v float64) float64 {
				var o v120.CalculatedTrackBarometricAltitude
				roundTrip(t, &v120.CalculatedTrackBarometricAltitude{Altitude: v}, &o)
				return o.Altitude
			}},
			tc{"v117 I062/135", 0.25, v, func(v float64) float64 {
				var o v117.CalculatedTrackBarometricAltitude
				roundTrip(t, &v117.CalculatedTrackBarometricAltitude{Altitude: v}, &o)
				return o.Altitude
			}},
			tc{"v120 I062/136", 0.25, v, func(v float64) float64 {
				var o v120.MeasuredFlightLevel
				roundTrip(t, &v120.MeasuredFlightLevel{FlightLevel: v}, &o)
				return o.FlightLevel
			}},
			tc{"v117 I062/136", 0.25, v, func(v float64) float64 {
				var o v117.MeasuredFlightLevel
				roundTrip(t, &v117.MeasuredFlightLevel{FlightLevel: v}, &o)
				return o.FlightLevel
			}},
			tc{"v117 I062/340 Mode C", 0.25, v, func(v float64) float64 {
				var o v117.MeasuredInformation
				roundTrip(t, &v117.MeasuredInformation{LastModeC: &v}, &o)
				return *o.LastModeC
			}},
		)
	}
	for _, v := range []float64{35012.6, 35012.4, 34999, -1012.6} {
		cases = append(cases,
			tc{"v120 I062/130", 6.25, v, func(v float64) float64 {
				var o v120.CalculatedTrackGeometricAltitude
				roundTrip(t, &v120.CalculatedTrackGeometricAltitude{Altitude: v}, &o)
				return o.Altitude
			}},
			tc{"v117 I062/130", 6.25, v, func(v float64) float64 {
				var o v117.CalculatedTrackGeometricAltitude
				roundTrip(t, &v117.CalculatedTrackGeometricAltitude{Altitude: v}, &o)
				return o.Altitude
			}},
			tc{"v120 I062/340 Mode C", 25, v, func(v float64) float64 {
				var o v120.MeasuredInformation
				roundTrip(t, &v120.MeasuredInformation{ModeCAltitude: &v}, &o)
				return *o.ModeCAltitude
			}},
			tc{"v117 I062/340 height", 25, v, func(v float64) float64 {
				var o v117.MeasuredInformation
				roundTrip(t, &v117.MeasuredInformation{Measured3DHeight: &v}, &o)
				return *o.Measured3DHeight
			}},
		)
	}
	for _, v := range []float64{40.3, 40.2, 0.25} {
		cases = append(cases,
			tc{"v117 I062/500 APC", 0.5, v, func(v float64) float64 {
				var o v117.EstimatedAccuracies
				roundTrip(t, &v117.EstimatedAccuracies{PositionAccuracyX: f64(v), PositionAccuracyY: f64(v)}, &o)
				return *o.PositionAccuracyY
			}},
			tc{"v117 I062/500 ABA", 0.25, v / 10, func(v float64) float64 {
				var o v117.EstimatedAccuracies
				roundTrip(t, &v117.EstimatedAccuracies{BarometricAltitudeAccuracy: f64(v)}, &o)
				return *o.BarometricAltitudeAccuracy
			}},
			tc{"v117 I062/500 ATV", 0.25, v / 10, func(v float64) float64 {
				var o v117.EstimatedAccuracies
				roundTrip(t, &v117.EstimatedAccuracies{VelocityAccuracyX: f64(v), VelocityAccuracyY: f64(v)}, &o)
				return *o.VelocityAccuracyX
			}},
			tc{"v117 I062/295 MFL", 0.25, v / 10, func(v float64) float64 {
				var o v117.TrackDataAges
				roundTrip(t, &v117.TrackDataAges{MFLAge: f64(v)}, &o)
				return *o.MFLAge
			}},
		)
	}
	for _, v := range []float64{-10.3, -10.2, 10.3, -10.25} {
		cases = append(cases, tc{"v117 I062/500 COV", 0.5, v, func(v float64) float64 {
			var o v117.EstimatedAccuracies
			roundTrip(t, &v117.EstimatedAccuracies{Covariance: f64(v)}, &o)
			return *o.Covariance
		}})
	}
	for _, c := range cases {
		want := nearest(c.v, c.lsb)
		if got := c.got(c.v); math.Abs(got-want) > 1e-9 {
			t.Errorf("%s: %v gives %v, want %v (nearest LSB %v)", c.name, c.v, got, want, c.lsb)
		}
	}
}
