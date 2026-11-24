package asterix_test

import (
	"encoding/base64"
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	uap021 "github.com/davidkohl/gobelix/cat/cat021/uap"
)

// TestCAT021TripleRecordOneBlock packs the SAME real-world record three times
// into ONE data block and proves all three decode with byte-exact re-encode
// and no data lost at the block tail. This guards the REF LEN semantics (LEN
// includes the LEN octet): the pre-fix decoder over-read one byte per RE,
// which on a multi-record block would desynchronise every following record —
// a single-record message could not catch that.
func TestCAT021TripleRecordOneBlock(t *testing.T) {
	raw, err := base64.StdEncoding.DecodeString(
		"FQBF8x0xQ8NjBAEPEUgLLgAO7u7vB3d3d0sZTzatUShoxyhRcxGwEgA2reNhlrHLOCADAgFgwxEBEUB2EXYHHIPAAUQA")
	if err != nil {
		t.Fatal(err)
	}
	// Strip the 3-byte block header (CAT + LEN); the rest is one record.
	record := raw[3:]
	// Rebuild a block with the record three times: CAT, LEN, rec, rec, rec.
	total := 3 + 3*len(record)
	block := make([]byte, 0, total)
	block = append(block, raw[0], byte(total>>8), byte(total))
	for i := 0; i < 3; i++ {
		block = append(block, record...)
	}

	uap, err := uap021.NewUAP26()
	if err != nil {
		t.Fatal(err)
	}
	dec := asterix.NewDecoder()
	dec.RegisterUAP(uap)
	blocks, err := dec.DecodeAll(block)
	if err != nil {
		t.Fatalf("DecodeAll failed on triple-record block: %v", err)
	}
	if len(blocks) != 1 {
		t.Fatalf("want 1 data block, got %d", len(blocks))
	}
	recs := blocks[0].Records()
	if len(recs) != 3 {
		t.Fatalf("want 3 records decoded, got %d (record desync — check RE LEN handling)", len(recs))
	}
	// All three records must decode identically (same source bytes) and
	// re-encode byte-exactly, proving no per-record byte slippage.
	for i, r := range recs {
		if r.ItemCount() != recs[0].ItemCount() {
			t.Fatalf("record %d item count %d != record 0 count %d", i, r.ItemCount(), recs[0].ItemCount())
		}
	}
	out, err := blocks[0].Encode()
	if err != nil {
		t.Fatalf("re-encode: %v", err)
	}
	if len(out) != total {
		t.Fatalf("re-encoded length %d != original %d (tail data lost)", len(out), total)
	}
	for i := range out {
		if out[i] != block[i] {
			t.Fatalf("re-encode differs at byte %d: %02x != %02x", i, out[i], block[i])
		}
	}
}
