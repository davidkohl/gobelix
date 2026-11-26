// cat/cat062/fusion_bug_hunt_feb22_test.go
package cat062_test

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat062/uap"
)

// TestFusionMessageFeb22 uses a fusion message captured on 2026-02-22 09:22.
// Only the first 55 of its 89 bytes survived in the log, so it cannot be
// decoded; the decoder must report the truncation instead of misreading the
// partial record.
func TestFusionMessageFeb22(t *testing.T) {
	hexData := "3e0059bfffed0664010141f93f0094d7b700115ecef7e7f704b884fe070130fffe0200003534b7d77820c121010001024535" +
		"34b7d77820" // capture truncated in the log

	data, err := hex.DecodeString(hexData)
	if err != nil {
		t.Fatalf("Failed to decode hex: %v", err)
	}
	if length := int(data[1])<<8 | int(data[2]); length != 89 || len(data) != 55 {
		t.Fatalf("fixture changed: LENGTH=%d, have %d bytes", length, len(data))
	}

	uap062, err := uap.NewUAP120()
	if err != nil {
		t.Fatalf("Failed to create UAP: %v", err)
	}
	decoder := asterix.NewDecoder()
	decoder.RegisterUAP(uap062)

	_, err = decoder.DecodeAll(data)
	if err == nil {
		t.Fatal("DecodeAll() succeeded on a truncated message")
	}
	if !strings.Contains(err.Error(), "truncated") {
		t.Errorf("DecodeAll() error = %v, want a truncation error", err)
	}
}
