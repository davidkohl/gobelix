// cat/cat062/fusion_i500_decode_test.go
package cat062_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	v120 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v120"
)

// TestDecodeI500FromFusionMessage extracts and decodes I062/500 from real fusion message
func TestDecodeI500FromFusionMessage(t *testing.T) {
	// Real message from fusion
	hexData := "3E0056BFFFED066401016D32EE008271240023976605F3F7EEAD5603C8FC5114EA0A900014C673C78820C10101007380C714C673C788200EF603010130B0020202400006170617FFFDC401720171DCC4B7B784622A28"

	data, err := hex.DecodeString(hexData)
	if err != nil {
		t.Fatalf("Failed to decode hex: %v", err)
	}

	// I062/500 should start around byte 73
	// Let's look at bytes 70-85
	t.Log("Bytes around I062/500 location:")
	for i := 70; i < 86 && i < len(data); i++ {
		t.Logf("Byte %d: 0x%02X (%08b) '%c'", i, data[i], data[i], printable(data[i]))
	}

	// Based on error "at byte 73/81", I062/500 FSPEC is at offset 73
	// Let's try to decode I062/500 starting from byte 73
	i500Bytes := data[73:]
	t.Logf("\nI062/500 bytes (from offset 73): %s", hex.EncodeToString(i500Bytes))

	item := &v120.EstimatedAccuracies{}
	buf := bytes.NewBuffer(i500Bytes)

	bytesRead, err := item.Decode(buf)
	if err != nil {
		t.Logf("Decode error: %v", err)
		t.Logf("Bytes read before error: %d", bytesRead)

		// Show what was in the buffer
		t.Logf("\nFirst few bytes of I062/500:")
		for i := 0; i < 10 && i < len(i500Bytes); i++ {
			t.Logf("  Byte %d: 0x%02X (%08b)", i, i500Bytes[i], i500Bytes[i])
		}
	} else {
		t.Logf("Decode succeeded! Read %d bytes", bytesRead)
		t.Logf("APCX: %v", ptrVal(item.APCX))
		t.Logf("APCY: %v", ptrVal(item.APCY))
		t.Logf("COV: %v", ptrValInt16(item.XYCovariance))
		t.Logf("VelocityX: %v", ptrVal8(item.VelocityX))
		t.Logf("VelocityY: %v", ptrVal8(item.VelocityY))
		t.Logf("AccelX: %v", ptrVal8(item.AccelerationX))
		t.Logf("AccelY: %v", ptrVal8(item.AccelerationY))
	}
}

func printable(b byte) byte {
	if b >= 32 && b < 127 {
		return b
	}
	return '.'
}

func ptrVal(p *uint16) string {
	if p == nil {
		return "nil"
	}
	return fmt.Sprintf("%d", *p)
}

func ptrVal8(p *uint8) string {
	if p == nil {
		return "nil"
	}
	return fmt.Sprintf("%d", *p)
}

func ptrValInt16(p *int16) string {
	if p == nil {
		return "nil"
	}
	return fmt.Sprintf("%d", *p)
}
