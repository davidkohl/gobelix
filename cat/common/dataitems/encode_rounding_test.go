package common

import (
	"bytes"
	"math"
	"testing"
)

// TestEncodeRounds checks that the time of day and the sector number
// round to the nearest LSB instead of truncating; the sector wraps at 360°.
// FlightLevel accepts only multiples of 1/4 FL, so it has nothing to round.
func TestEncodeRounds(t *testing.T) {
	near := func(v, lsb float64) float64 { return math.Round(v/lsb) * lsb }
	for _, v := range []float64{36000.006, 36000.002, 0.004} {
		var b bytes.Buffer
		if _, err := (&TimeOfDay{Time: v}).Encode(&b); err != nil {
			t.Fatal(err)
		}
		var o TimeOfDay
		if _, err := o.Decode(&b); err != nil {
			t.Fatal(err)
		}
		if want := near(v, 1.0/128); o.Time != want {
			t.Errorf("TimeOfDay %v gives %v, want %v", v, o.Time, want)
		}
	}
	for _, c := range []struct{ v, want float64 }{{1.0, 1.40625}, {0.5, 0}, {359.9, 0}, {90.6, 90}, {90.8, 91.40625}} {
		var b bytes.Buffer
		if _, err := (&SectorNumber{SectorNumber: c.v}).Encode(&b); err != nil {
			t.Fatal(err)
		}
		var o SectorNumber
		if _, err := o.Decode(&b); err != nil {
			t.Fatal(err)
		}
		if o.SectorNumber != c.want {
			t.Errorf("SectorNumber %v gives %v, want %v", c.v, o.SectorNumber, c.want)
		}
	}
}
