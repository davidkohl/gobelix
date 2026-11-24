// asterix/cat021_realdata_test.go
//
// Decodes a CAT021 message captured from a live sensor feed and verifies
// the record content, including the Reserved Expansion Field whose length
// indicator counts the total field length including the indicator itself
// (EUROCONTROL-SPEC-0149 Part 1, 5.2.5.5.1).
package asterix_test

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/davidkohl/gobelix/asterix"

	uap021 "github.com/davidkohl/gobelix/cat/cat021/uap"
	v26cat021 "github.com/davidkohl/gobelix/cat/cat021/dataitems/v26"
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

// Captured CAT021 data block: CAT=21, LEN=69, one record with a 7-octet
// FSPEC declaring FRNs 1,2,3,4,7,11,12,13,17,18,23,28,29,30,35,37,38,42,48(RE).
const cat021RealMessageB64 = "FQBF8x0xQ8NjBAEPEUgLLgAO7u7vB3d3d0sZTzatUShoxyhRcxGwEgA2reNhlrHLOCADAgFgwxEBEUB2EXYHHIPAAUQA"

func TestCAT021DecodeRealMessage(t *testing.T) {
	raw, err := base64.StdEncoding.DecodeString(cat021RealMessageB64)
	if err != nil {
		t.Fatalf("decoding base64 fixture: %v", err)
	}

	uap, err := uap021.NewUAP26()
	if err != nil {
		t.Fatalf("creating CAT021 UAP: %v", err)
	}

	dec := asterix.NewDecoder()
	dec.RegisterUAP(uap)
	blocks, err := dec.DecodeAll(raw)
	if err != nil {
		t.Fatalf("DecodeAll failed: %v\nraw bytes: % X", err, raw)
	}
	if len(blocks) != 1 {
		t.Fatalf("expected 1 data block, got %d", len(blocks))
	}
	recs := blocks[0].Records()
	if len(recs) != 1 {
		t.Fatalf("expected 1 record, got %d", len(recs))
	}
	rec := recs[0]

	assertItem(t, rec, "I021/010", func(item asterix.DataItem) {
		dsi, ok := item.(*common.DataSourceIdentifier)
		if !ok {
			t.Fatalf("I021/010: unexpected type %T", item)
		}
		assertEq(t, "SAC", uint8(1), dsi.SAC)
		assertEq(t, "SIC", uint8(15), dsi.SIC)
	})

	assertItem(t, rec, "I021/161", func(item asterix.DataItem) {
		tn, ok := item.(*common.TrackNumber)
		if !ok {
			t.Fatalf("I021/161: unexpected type %T", item)
		}
		assertEq(t, "Value", uint16(2862), tn.Value)
	})

	assertItem(t, rec, "I021/080", func(item asterix.DataItem) {
		ta, ok := item.(*common.TargetAddress)
		if !ok {
			t.Fatalf("I021/080: unexpected type %T", item)
		}
		assertEq(t, "Address", uint32(0x4B194F), ta.Address)
	})

	assertItem(t, rec, "I021/170", func(item asterix.DataItem) {
		ti, ok := item.(*v26cat021.TargetIdentification)
		if !ok {
			t.Fatalf("I021/170: unexpected type %T", item)
		}
		assertEq(t, "Ident", "XYZ123", ti.Ident)
	})

	// The RE field is the last 7 octets of the record: 07 1C 83 C0 01 44 00.
	// LEN=0x07 covers the whole field including itself; Items Indicator
	// 0x1C = GAO+SGV+STA present. GAO=0x83, SGV=C0 01 44 (FX-extended
	// primary + one extension octet), STA=00 (single primary octet, FX=0).
	assertItem(t, rec, "RE", func(item asterix.DataItem) {
		re, ok := item.(*v26cat021.ReservedExpansion)
		if !ok {
			t.Fatalf("RE: unexpected type %T", item)
		}
		want := []byte{0x07, 0x1C, 0x83, 0xC0, 0x01, 0x44, 0x00}
		if !bytes.Equal(re.Data, want) {
			t.Errorf("RE data: want % X, got % X", want, re.Data)
		}
		if len(re.Data) > 0 && int(re.Data[0]) != len(re.Data) {
			t.Errorf("RE length indicator %d does not cover total field length %d (must include itself)",
				re.Data[0], len(re.Data))
		}
		if !re.HasGAO || re.HasBPS || re.HasSelH || re.HasNAV || !re.HasSGV || !re.HasSTA || re.HasTNH || re.HasMES {
			t.Fatalf("RE items indicator mismatch: got GAO=%v SGV=%v STA=%v (want only these three)",
				re.HasGAO, re.HasSGV, re.HasSTA)
		}
		if !re.GAORight || re.GAOLateralM != 0 || re.GAOLongitudinalM != 6 {
			t.Errorf("RE.GAO: want right=true lat=0m lon=6m, got right=%v lat=%dm lon=%dm",
				re.GAORight, re.GAOLateralM, re.GAOLongitudinalM)
		}
		if !bytes.Equal(re.RawSGV, []byte{0xC0, 0x01, 0x44}) {
			t.Errorf("RE.RawSGV: want C0 01 44, got % X", re.RawSGV)
		}
		if !bytes.Equal(re.RawSTA, []byte{0x00}) {
			t.Errorf("RE.RawSTA: want 00, got % X", re.RawSTA)
		}
		t.Logf("RE.String(): %s", re.String())
	})

	// Re-encoding the decoded block must reproduce the captured bytes exactly.
	encoded, err := blocks[0].Encode()
	if err != nil {
		t.Fatalf("re-encoding decoded block: %v", err)
	}
	if !bytes.Equal(encoded, raw) {
		t.Errorf("re-encoded block differs from captured message\nwant % X\ngot  % X", raw, encoded)
	}
}
