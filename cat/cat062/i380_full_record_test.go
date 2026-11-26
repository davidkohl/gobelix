// cat/cat062/i380_full_record_test.go
package cat062_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	v117 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v117"
	v120 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v120"
)

// i380AllSubfieldsHex is an I062/380 with all 28 subfields present
// (primary subfield FF FF FF FE), 97 octets, using the CAT062 subfield
// lengths: TID = REP 01 + 15 octets, MET = 8 octets, POS = 6 octets,
// MB = REP 01 + 8 octets.
const i380AllSubfieldsHex = "FFFFFFFE" +
	"123456" + "14C673C78820" + "4000" + "0123" + "01C2" + "E0C8" + "A0C8" + // ADR ID MHG IAS TAS SAL FSS
	"80" + "01" + "85012C123456FEDCBA5B0123450064" + "4CF5" + "5E05" + "11223344556677" + "FFC0" + "0040" + // TIS TID COM SAB ACS BVR GVR
	"FFF6" + "0010" + "4000" + "0100" + "05" + "F00014005AFFF002" + "0D" + // RAN TAR TAN GSP VUN MET EMC
	"123456FEDCBA" + "0640" + "07" + "01" + "1122334455667740" + "0123" + "0064" + "03F5" // POS GAL PUN MB IAR MAC BPS

// TestI380AllSubfieldsInFullRecord places an I062/380 carrying every
// subfield in a CAT062 record between other items and checks the whole data
// block decodes and re-encodes byte-exactly. I062/290 carries only the MDS
// age (one octet), the case the old decoder read as two octets.
func TestI380AllSubfieldsInFullRecord(t *testing.T) {
	record := "911F0104" + // FSPEC: 010 070 | 380 040 080 290 | - | 500
		"0664" + // I062/010
		"6D32EE" + // I062/070
		i380AllSubfieldsHex + // I062/380
		"0EF6" + // I062/040
		"03010130" + // I062/080
		"1014" + // I062/290: MDS age only
		"C401720171DCC4B7B7" // I062/500: APC COV ATV
	body, err := hex.DecodeString(record)
	if err != nil {
		t.Fatalf("bad hex: %v", err)
	}
	length := 3 + len(body)
	data := append([]byte{byte(asterix.Cat062), byte(length >> 8), byte(length)}, body...)

	blocks := decodeCat062v120(t, data)
	if len(blocks) != 1 || len(blocks[0].Records()) != 1 {
		t.Fatalf("decoded %d blocks, want 1 block with 1 record", len(blocks))
	}
	items := blocks[0].Records()[0].Items()
	if len(items) != 7 {
		t.Fatalf("decoded %d items, want 7", len(items))
	}

	i380, ok := items["I062/380"].(*v120.AircraftDerivedData)
	if !ok {
		t.Fatalf("I062/380 has type %T", items["I062/380"])
	}
	if i380.BarometricPressureSetting == nil || i380.PositionLongitude == nil || i380.MeteorologicalData == nil {
		t.Errorf("I062/380 subfields missing: %s", i380)
	}
	if tn, ok := items["I062/040"].(*v120.TrackNumber); !ok || tn.Value != 0x0EF6 {
		t.Errorf("I062/040 after I062/380 = %v, want track number 0x0EF6", items["I062/040"])
	}
	if _, ok := items["I062/500"].(*v120.EstimatedAccuracies); !ok {
		t.Errorf("I062/500 = %T", items["I062/500"])
	}

	encoded, err := blocks[0].Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	if !bytes.Equal(encoded, data) {
		t.Errorf("data block round trip not byte-exact:\n  in:  %x\n  out: %x", data, encoded)
	}
}

// TestI380AllSubfieldsV117 checks the edition 1.17 decoder consumes the same
// all-subfield item completely and re-encodes it byte-exactly, so the length
// handling stays aligned with v1.20.
func TestI380AllSubfieldsV117(t *testing.T) {
	input, err := hex.DecodeString(i380AllSubfieldsHex)
	if err != nil {
		t.Fatalf("bad hex: %v", err)
	}
	item := &v117.AircraftDerivedData{}
	n, err := item.Decode(bytes.NewBuffer(input))
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if n != len(input) {
		t.Fatalf("Decode() consumed %d bytes, want %d", n, len(input))
	}
	var out bytes.Buffer
	if _, err := item.Encode(&out); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	if !bytes.Equal(out.Bytes(), input) {
		t.Errorf("round trip not byte-exact:\n  in:  %x\n  out: %x", input, out.Bytes())
	}
}
