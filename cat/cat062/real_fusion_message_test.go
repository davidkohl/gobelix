package cat062_test

import (
	"encoding/hex"
	"reflect"
	"sort"
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	v120 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v120"
	"github.com/davidkohl/gobelix/cat/cat062/uap"
)

// realFusionMessages are CAT062 v1.20 messages captured from ATLAS fusion
// (NATS opentrack.tracks.updates). Each was cross-checked against the
// EUROCONTROL CAT062 subfield lengths: the record ends exactly at LENGTH.
var realFusionMessages = []struct {
	name    string
	hexData string
	i380Len int // octets of I062/380 in the record
}{
	{
		name:    "I062/500 APW failure (I062/290 SSR+MDS)",
		hexData: "3e0058bfffed066401016da302008d3cfd001ff78d02ba90fb7453fc85014b00fe020000407539ccd460c12101004bb84d407539ccd4600000c34783010130b0000000000005f005f00001c4009d009e00a76e7284622728",
		i380Len: 15,
	},
	{
		name:    "ELY318 (I062/290 TRK+SSR+MDS)",
		hexData: "3E0056BFFFED066401016D32EE008271240023976605F3F7EEAD5603C8FC5114EA0A900014C673C78820C10101007380C714C673C788200EF603010130B0020202400006170617FFFDC401720171DCC4B7B784622A28",
		i380Len: 13,
	},
	{
		name:    "I062/220 EOF failure",
		hexData: "3E0056BFFFED066401015D1A2E0083AAA30012C7C5F803CEF036F5039FFDFBFDFF0A9600154237C31820C101010004012B154237C318201B5283010130B0000000000005C705C70002C400C800C6FF6F4D4B84640528",
		i380Len: 13,
	},
	{
		name:    "I062/500 ATV failure",
		hexData: "3E0056BFFFED066401015D1A2F008E9F4B001D3BC0009AE7FD1579FCDBFF830E000200005161B2D995A0C101010039DE475161B2D995A0CB2483010130B12000040400000005C805C8000084001400142A2484622528",
		i380Len: 13,
	},
	{
		// Captured with a trailing newline (0x0a) after the 88-byte block.
		name:    "LENGTH 88 with trailing capture newline",
		hexData: "3e0058bfffed066401014377090086cb860015d714faa6d2f3d617fd4bfd6b00010649005da6b5d92120c1210100471db95da6b5d921200000f5718b010130b008080a000005280528ffdcc4010d010cf9817c7c84620e18",
		i380Len: 15,
	},
}

// realFusionItems are the data items every realFusionMessages record carries
// (record FSPEC BF FF ED 06).
var realFusionItems = []string{
	"I062/010", "I062/015", "I062/040", "I062/060", "I062/070", "I062/080",
	"I062/100", "I062/105", "I062/135", "I062/136", "I062/185", "I062/200",
	"I062/210", "I062/220", "I062/245", "I062/290", "I062/295", "I062/340",
	"I062/380", "I062/500",
}

func decodeCat062v120(t *testing.T, data []byte) []*asterix.DataBlock {
	t.Helper()
	uap062, err := uap.NewUAP120()
	if err != nil {
		t.Fatalf("Failed to create UAP: %v", err)
	}
	decoder := asterix.NewDecoder()
	decoder.RegisterUAP(uap062)
	blocks, err := decoder.DecodeAll(data)
	if err != nil {
		t.Fatalf("DecodeAll() error = %v", err)
	}
	return blocks
}

// TestRealFusionMessage decodes real fusion messages that gobelix used to
// reject ("decoding APW: buffer too short for APW" and similar). The cause
// was I062/290 reading the 1-octet MDS age as 2 octets, which shifted every
// following item; I062/380 itself was consumed correctly.
func TestRealFusionMessage(t *testing.T) {
	for _, tt := range realFusionMessages {
		t.Run(tt.name, func(t *testing.T) {
			data, err := hex.DecodeString(tt.hexData)
			if err != nil {
				t.Fatalf("Failed to decode hex: %v", err)
			}
			if length := int(data[1])<<8 | int(data[2]); length != len(data) {
				t.Fatalf("fixture LENGTH=%d, have %d bytes", length, len(data))
			}

			blocks := decodeCat062v120(t, data)
			if len(blocks) != 1 {
				t.Fatalf("decoded %d data blocks, want 1", len(blocks))
			}
			records := blocks[0].Records()
			if len(records) != 1 {
				t.Fatalf("decoded %d records, want 1", len(records))
			}

			items := records[0].Items()
			var got []string
			for id := range items {
				got = append(got, id)
			}
			sort.Strings(got)
			if len(got) != len(realFusionItems) {
				t.Fatalf("decoded items %v, want %v", got, realFusionItems)
			}
			for i := range got {
				if got[i] != realFusionItems[i] {
					t.Fatalf("decoded items %v, want %v", got, realFusionItems)
				}
			}

			i380, ok := items["I062/380"].(*v120.AircraftDerivedData)
			if !ok {
				t.Fatalf("I062/380 has type %T", items["I062/380"])
			}
			if i380.TargetAddress == nil || i380.TargetIdentification == nil {
				t.Errorf("I062/380 missing ADR or ID: %s", i380)
			}
			if tt.i380Len == 15 && i380.CommunicationsCOM == nil {
				t.Errorf("I062/380 missing COM subfield: %s", i380)
			}
		})
	}
}

// TestRealFusionMessageReencode re-encodes each decoded real record and
// decodes the result again: the second decode must succeed with the same
// item values, which proves encoder and decoder agree on every item's
// length.
func TestRealFusionMessageReencode(t *testing.T) {
	for _, tt := range realFusionMessages {
		t.Run(tt.name, func(t *testing.T) {
			data, err := hex.DecodeString(tt.hexData)
			if err != nil {
				t.Fatalf("Failed to decode hex: %v", err)
			}
			items := decodeCat062v120(t, data)[0].Records()[0].Items()

			uap062, err := uap.NewUAP120()
			if err != nil {
				t.Fatalf("Failed to create UAP: %v", err)
			}
			db, err := asterix.NewDataBlock(asterix.Cat062, uap062)
			if err != nil {
				t.Fatalf("NewDataBlock() error = %v", err)
			}
			if err := db.EncodeRecord(items); err != nil {
				t.Fatalf("EncodeRecord() error = %v", err)
			}
			encoded, err := db.Encode()
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}

			again := decodeCat062v120(t, encoded)[0].Records()[0].Items()
			if len(again) != len(items) {
				t.Fatalf("re-decoded %d items, want %d", len(again), len(items))
			}
			for id, item := range items {
				if !reflect.DeepEqual(again[id], item) {
					t.Errorf("%s: re-decoded %v, want %v", id, again[id], item)
				}
			}
		})
	}
}
