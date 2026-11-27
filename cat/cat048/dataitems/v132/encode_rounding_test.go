package v132

import (
	"bytes"
	"math"
	"testing"
)

// TestFlightLevelRounds checks that I048/090 rounds to the nearest 1/4 FL,
// half away from zero, instead of truncating.
func TestFlightLevelRounds(t *testing.T) {
	for _, v := range []float64{350.2, 350.1, 350.125, -10.2, -10.125} {
		var b bytes.Buffer
		if _, err := (&FlightLevel{Level: v}).Encode(&b); err != nil {
			t.Fatal(err)
		}
		var o FlightLevel
		if _, err := o.Decode(&b); err != nil {
			t.Fatal(err)
		}
		if want := math.Round(v*4) / 4; o.Level != want {
			t.Errorf("%v gives %v, want %v", v, o.Level, want)
		}
	}
}
