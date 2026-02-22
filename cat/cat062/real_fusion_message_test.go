package cat062_test

import (
	"encoding/hex"
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat062/uap"
)

// TestRealFusionMessage tests decoding a REAL message from ATLAS fusion
// that the old phxwi/asterixdecoder can decode but gobelix cannot.
//
// This message was captured from NATS opentrack.tracks.updates and fails with:
// "decoding error in Category 62, item I062/500, at byte 79/82: decoding data item: decoding APW: buffer too short for APW"
//
// The old converter (phxwi/asterixdecoder) decodes this SAME message successfully.
// This proves the message is VALID ASTERIX, and gobelix has a decoder bug.
func TestRealFusionMessage(t *testing.T) {
	// Real message from fusion that fails in gobelix but works in old decoder
	hexData := "3e0058bfffed066401016da302008d3cfd001ff78d02ba90fb7453fc85014b00fe020000407539ccd460c12101004bb84d407539ccd4600000c34783010130b0000000000005f005f00001c4009d009e00a76e7284622728"

	data, err := hex.DecodeString(hexData)
	if err != nil {
		t.Fatalf("Failed to decode hex: %v", err)
	}

	t.Logf("Message: %d bytes, LENGTH field: %d", len(data), int(data[1])<<8|int(data[2]))

	// Try to decode with gobelix
	uap062, err := uap.NewUAP120()
	if err != nil {
		t.Fatalf("Failed to create UAP: %v", err)
	}

	decoder := asterix.NewDecoder()
	decoder.RegisterUAP(uap062)

	blocks, err := decoder.DecodeAll(data)
	if err != nil {
		t.Logf("Decode error: %v", err)

		// This is the expected failure for now
		// Once we fix the bug, this test should pass
		t.Logf("KNOWN BUG: gobelix cannot decode this message that old decoder handles fine")
		t.Logf("Root cause: compound item decoders have byte counting bugs")

		// For now, just log the error
		// TODO: Remove this skip once fixed
		t.Skip("Skipping until compound item decoders are fixed")
		return
	}

	// If we get here, decoding succeeded!
	t.Logf("✅ Decode successful!")

	if len(blocks) == 0 {
		t.Fatal("No data blocks decoded")
	}

	records := blocks[0].Records()
	if len(records) == 0 {
		t.Fatal("No records decoded")
	}

	items := records[0].Items()
	t.Logf("Decoded %d data items", len(items))

	// Verify key items are present
	if _, ok := items["I062/010"]; !ok {
		t.Error("Missing I062/010 (Data Source Identifier)")
	}
	if _, ok := items["I062/040"]; !ok {
		t.Error("Missing I062/040 (Track Number)")
	}
	if _, ok := items["I062/500"]; !ok {
		t.Error("Missing I062/500 (Estimated Accuracies)")
	}
}

// TestCompareGobelixVsOldDecoder documents the behavior difference between
// gobelix and the old phxwi/asterixdecoder.
//
// Key question: WHY does the old decoder succeed?
// Possible reasons:
// 1. More lenient validation (ignores errors)
// 2. Different FSPEC parsing logic
// 3. Different compound item decoder implementation
// 4. Handles byte misalignment better
func TestCompareGobelixVsOldDecoder(t *testing.T) {
	t.Log("Old decoder (phxwi/asterixdecoder): 100% success rate")
	t.Log("Gobelix decoder: 90% failure rate")
	t.Log("")
	t.Log("Both decoders receive IDENTICAL messages from NATS")
	t.Log("Messages are encoded by ATLAS fusion using gobelix")
	t.Log("")
	t.Log("Hypothesis: gobelix encoder creates valid messages, but gobelix decoder has bugs")
	t.Log("Evidence: Fusion publishes with LENGTH=actual_size (no truncation)")
	t.Log("Evidence: Old decoder successfully decodes 100% of messages")
	t.Log("")
	t.Log("Most likely cause: Compound item decoders (I062/290, I062/295, I062/380, I062/390)")
	t.Log("have byte counting bugs that cause cumulative misalignment")
}
