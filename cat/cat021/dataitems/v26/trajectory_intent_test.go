// dataitems/cat021/v26/trajectory_intent_test.go
package v26_test

import (
	"bytes"
	"testing"

	v26 "github.com/davidkohl/gobelix/cat/cat021/dataitems/v26"
)

func TestTrajectoryIntent_TooManyPoints(t *testing.T) {
	// 256 points must be refused before anything is written, not wrapped to
	// a repetition factor of 0 after the primary and TIS octets.
	nav := true
	ti := v26.TrajectoryIntent{NAV: &nav, Points: make([]v26.TrajectoryIntentPoint, 256)}
	var buf bytes.Buffer
	if _, err := ti.Encode(&buf); err == nil {
		t.Fatal("want an error for 256 points")
	}
	if buf.Len() != 0 {
		t.Errorf("wrote %d bytes before failing", buf.Len())
	}

	ti.Points = ti.Points[:255]
	n, err := ti.Encode(&buf)
	if err != nil {
		t.Fatalf("255 points: %v", err)
	}
	if b := buf.Bytes(); n != buf.Len() || b[0] != 0xC0 || b[2] != 255 {
		t.Errorf("255 points: n=%d len=%d primary=%#x rep=%d", n, buf.Len(), b[0], b[2])
	}
}
