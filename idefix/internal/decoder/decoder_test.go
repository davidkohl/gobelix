// internal/decoder/decoder_test.go
package decoder

import (
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat020"
)

func TestCreateDecoder_Cat020Version(t *testing.T) {
	config := Config{
		DumpCat020: true,
	}

	decoder, err := CreateDecoder(config)
	if err != nil {
		t.Fatalf("CreateDecoder failed: %v", err)
	}

	// Get the registered UAP for CAT020
	uap := decoder.GetUAP(asterix.Cat020)
	if uap == nil {
		t.Fatal("CAT020 UAP not registered")
	}

	// Verify it's version 1.0
	if uap.Version() != "1.0" {
		t.Errorf("Expected CAT020 version 1.0, got %s", uap.Version())
	}

	// Also verify using the constant
	expectedVersion := cat020.Version10
	if uap.Version() != expectedVersion {
		t.Errorf("Expected CAT020 version %s, got %s", expectedVersion, uap.Version())
	}

	t.Logf("CAT020 UAP version: %s ✓", uap.Version())
}

func TestCreateDecoder_AllCategories(t *testing.T) {
	config := Config{
		DumpAll: true,
	}

	decoder, err := CreateDecoder(config)
	if err != nil {
		t.Fatalf("CreateDecoder failed: %v", err)
	}

	// Verify CAT020 is using Edition 1.0
	uap020 := decoder.GetUAP(asterix.Cat020)
	if uap020 == nil {
		t.Fatal("CAT020 UAP not registered with DumpAll")
	}

	if uap020.Version() != "1.0" {
		t.Errorf("With DumpAll, expected CAT020 version 1.0, got %s", uap020.Version())
	}

	t.Log("All categories registered successfully with CAT020 Edition 1.0 ✓")
}
