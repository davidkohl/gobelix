package cat062_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	cat062 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v120"
)

// TestI500ExactScenario tests the EXACT I062/500 configuration from the failing message
// FSPEC: 0x7c = 0b01111100
// COV + APW + AGA + ABA + ATV
func TestI500ExactScenario(t *testing.T) {
	i500 := &cat062.EstimatedAccuracies{}
	
	// NO APC (bit 7 = 0)
	
	// COV (bit 6 = 1): XY Covariance
	cov := int16(123)
	i500.XYCovariance = &cov
	
	// APW (bit 5 = 1): WGS-84 Position Accuracy
	apwLat := uint16(100)
	apwLon := uint16(200)
	i500.APWLatitude = &apwLat
	i500.APWLongitude = &apwLon
	
	// AGA (bit 4 = 1): Geometric Altitude Accuracy
	aga := uint8(50)
	i500.GeometricAltitudeAccuracy = &aga
	
	// ABA (bit 3 = 1): Barometric Altitude Accuracy
	aba := uint8(60)
	i500.BarometricAltitudeAccuracy = &aba
	
	// ATV (bit 2 = 1): Velocity Accuracy
	vx := uint8(20)
	vy := uint8(24)
	i500.VelocityX = &vx
	i500.VelocityY = &vy
	
	// NO AA (bit 1 = 0)
	// NO FX (bit 0 = 0)
	
	// Encode
	var buf bytes.Buffer
	n, err := i500.Encode(&buf)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	
	encoded := buf.Bytes()
	t.Logf("Encoded %d bytes (returned n=%d)", len(encoded), n)
	t.Logf("Hex: %s", hex.EncodeToString(encoded))
	
	// Expected:
	// 1 byte FSPEC (0x7c)
	// 2 bytes COV
	// 4 bytes APW  
	// 1 byte AGA
	// 1 byte ABA
	// 2 bytes ATV
	// Total: 11 bytes
	
	expected := 11
	if len(encoded) != expected {
		t.Errorf("Expected %d bytes, got %d", expected, len(encoded))
	}
	
	if n != len(encoded) {
		t.Errorf("❌ CRITICAL BUG: Encode() returned n=%d but wrote %d bytes (diff=%+d)",
			n, len(encoded), len(encoded)-n)
	}
	
	// Verify FSPEC
	if len(encoded) > 0 {
		fspec := encoded[0]
		t.Logf("FSPEC: 0x%02x", fspec)
		if fspec != 0x7c {
			t.Errorf("Expected FSPEC 0x7c, got 0x%02x", fspec)
		}
	}
	
	// Decode it back
	decBuf := bytes.NewBuffer(encoded)
	var decoded cat062.EstimatedAccuracies
	nDec, err := decoded.Decode(decBuf)
	if err != nil {
		t.Fatalf("Decode failed after %d bytes: %v", nDec, err)
	}
	
	t.Logf("Decode succeeded: read %d bytes", nDec)
	
	if nDec != n {
		t.Errorf("Encode returned %d but Decode read %d bytes", n, nDec)
	}
}
