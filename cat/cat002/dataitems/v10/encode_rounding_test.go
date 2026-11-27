package v10

import (
	"bytes"
	"testing"
)

// TestRotationPeriodRounds checks that I002/041 rounds to the nearest
// 1/128 s instead of truncating.
func TestRotationPeriodRounds(t *testing.T) {
	for _, c := range []struct{ v, want float64 }{{4.006, 4.0078125}, {4.002, 4}} {
		var b bytes.Buffer
		if _, err := (&AntennaRotationSpeed{RotationPeriod: c.v}).Encode(&b); err != nil {
			t.Fatal(err)
		}
		var o AntennaRotationSpeed
		if _, err := o.Decode(&b); err != nil {
			t.Fatal(err)
		}
		if o.RotationPeriod != c.want {
			t.Errorf("%v gives %v, want %v", c.v, o.RotationPeriod, c.want)
		}
	}
}
