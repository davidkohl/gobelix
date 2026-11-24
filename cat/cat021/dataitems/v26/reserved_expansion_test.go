// dataitems/cat021/v26/reserved_expansion_test.go
package v26_test

import (
	"bytes"
	"testing"

	v26 "github.com/davidkohl/gobelix/cat/cat021/dataitems/v26"
)

func TestReservedExpansion_RawPassthrough(t *testing.T) {
	original := v26.ReservedExpansion{Data: []byte{3, 0x00, 0xEF}}

	buf := new(bytes.Buffer)
	if _, err := original.Encode(buf); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	if !bytes.Equal(buf.Bytes(), original.Data) {
		t.Errorf("Encode() = % X, want % X", buf.Bytes(), original.Data)
	}

	decoded := &v26.ReservedExpansion{}
	if _, err := decoded.Decode(bytes.NewBuffer(buf.Bytes())); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if !bytes.Equal(decoded.Data, original.Data) {
		t.Errorf("Decode() Data = % X, want % X", decoded.Data, original.Data)
	}
}

func TestReservedExpansion_EmptyEncodesZeroLength(t *testing.T) {
	buf := new(bytes.Buffer)
	if _, err := (&v26.ReservedExpansion{}).Encode(buf); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	if !bytes.Equal(buf.Bytes(), []byte{0}) {
		t.Errorf("Encode() = % X, want 00", buf.Bytes())
	}
}

func TestReservedExpansion_StructuredRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input v26.ReservedExpansion
	}{
		{
			name: "BPS",
			input: v26.ReservedExpansion{
				HasBPS: true,
				BPSHpa: 1013.2,
			},
		},
		{
			name: "SelH magnetic valid",
			input: v26.ReservedExpansion{
				HasSelH:      true,
				SelHDeg:      87.890625, // exact multiple of the 0.703125 LSB
				SelHMagnetic: true,
				SelHValid:    true,
			},
		},
		{
			name: "SelH true unvalidated",
			input: v26.ReservedExpansion{
				HasSelH:      true,
				SelHDeg:      0,
				SelHMagnetic: false,
				SelHValid:    false,
			},
		},
		{
			name: "NAV all flags",
			input: v26.ReservedExpansion{
				HasNAV:          true,
				NAVAutopilot:    true,
				NAVVNAV:         true,
				NAVAltitudeHold: true,
				NAVApproach:     true,
				NAVMCPPopulated: true,
			},
		},
		{
			name: "NAV MCP not populated",
			input: v26.ReservedExpansion{
				HasNAV: true,
			},
		},
		{
			name: "GAO right",
			input: v26.ReservedExpansion{
				HasGAO:           true,
				GAORight:         true,
				GAOLateralM:      6,
				GAOLongitudinalM: 30,
			},
		},
		{
			name: "GAO left",
			input: v26.ReservedExpansion{
				HasGAO:           true,
				GAORight:         false,
				GAOLateralM:      0,
				GAOLongitudinalM: 2,
			},
		},
		{
			name: "TNH",
			input: v26.ReservedExpansion{
				HasTNH: true,
				TNHDeg: 180.0,
			},
		},
		{
			name: "SGV raw",
			input: v26.ReservedExpansion{
				HasSGV: true,
				RawSGV: []byte{0xC0, 0x01, 0x44},
			},
		},
		{
			name: "STA raw",
			input: v26.ReservedExpansion{
				HasSTA: true,
				RawSTA: []byte{0x00},
			},
		},
		{
			name: "MES raw single octet",
			input: v26.ReservedExpansion{
				HasMES: true,
				RawMES: []byte{0x80}, // FX=0: no extension
			},
		},
		{
			name: "MES raw FX-chained",
			input: v26.ReservedExpansion{
				HasMES: true,
				RawMES: []byte{0x81, 0x00}, // FX=1 then extension octet with FX=0
			},
		},
		{
			name: "BPS + SelH + GAO combined",
			input: v26.ReservedExpansion{
				HasBPS:           true,
				BPSHpa:           900.0,
				HasSelH:          true,
				SelHDeg:          45.0,
				SelHMagnetic:     true,
				SelHValid:        true,
				HasGAO:           true,
				GAORight:         true,
				GAOLateralM:      2,
				GAOLongitudinalM: 4,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			if _, err := tt.input.Encode(buf); err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			encoded := buf.Bytes()
			if int(encoded[0]) != len(encoded) {
				t.Fatalf("LEN octet %d does not match total encoded length %d", encoded[0], len(encoded))
			}

			decoded := &v26.ReservedExpansion{}
			n, err := decoded.Decode(bytes.NewBuffer(encoded))
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if n != len(encoded) {
				t.Errorf("Decode() consumed %d bytes, want %d", n, len(encoded))
			}

			if decoded.HasBPS != tt.input.HasBPS || decoded.BPSHpa != tt.input.BPSHpa {
				t.Errorf("BPS: got Has=%v Hpa=%v, want Has=%v Hpa=%v", decoded.HasBPS, decoded.BPSHpa, tt.input.HasBPS, tt.input.BPSHpa)
			}
			if decoded.HasSelH != tt.input.HasSelH || decoded.SelHDeg != tt.input.SelHDeg ||
				decoded.SelHMagnetic != tt.input.SelHMagnetic || decoded.SelHValid != tt.input.SelHValid {
				t.Errorf("SelH: got %+v, want deg=%v mag=%v valid=%v",
					decoded, tt.input.SelHDeg, tt.input.SelHMagnetic, tt.input.SelHValid)
			}
			if decoded.HasNAV != tt.input.HasNAV || decoded.NAVAutopilot != tt.input.NAVAutopilot ||
				decoded.NAVVNAV != tt.input.NAVVNAV || decoded.NAVAltitudeHold != tt.input.NAVAltitudeHold ||
				decoded.NAVApproach != tt.input.NAVApproach || decoded.NAVMCPPopulated != tt.input.NAVMCPPopulated {
				t.Errorf("NAV: got %+v, want %+v", decoded, tt.input)
			}
			if decoded.HasGAO != tt.input.HasGAO || decoded.GAORight != tt.input.GAORight ||
				decoded.GAOLateralM != tt.input.GAOLateralM || decoded.GAOLongitudinalM != tt.input.GAOLongitudinalM {
				t.Errorf("GAO: got %+v, want %+v", decoded, tt.input)
			}
			if decoded.HasTNH != tt.input.HasTNH || decoded.TNHDeg != tt.input.TNHDeg {
				t.Errorf("TNH: got Has=%v Deg=%v, want Has=%v Deg=%v", decoded.HasTNH, decoded.TNHDeg, tt.input.HasTNH, tt.input.TNHDeg)
			}
			if decoded.HasSGV != tt.input.HasSGV || !bytes.Equal(decoded.RawSGV, tt.input.RawSGV) {
				t.Errorf("SGV: got Has=%v Raw=% X, want Has=%v Raw=% X", decoded.HasSGV, decoded.RawSGV, tt.input.HasSGV, tt.input.RawSGV)
			}
			if decoded.HasSTA != tt.input.HasSTA || !bytes.Equal(decoded.RawSTA, tt.input.RawSTA) {
				t.Errorf("STA: got Has=%v Raw=% X, want Has=%v Raw=% X", decoded.HasSTA, decoded.RawSTA, tt.input.HasSTA, tt.input.RawSTA)
			}
			if decoded.HasMES != tt.input.HasMES || !bytes.Equal(decoded.RawMES, tt.input.RawMES) {
				t.Errorf("MES: got Has=%v Raw=% X, want Has=%v Raw=% X", decoded.HasMES, decoded.RawMES, tt.input.HasMES, tt.input.RawMES)
			}
		})
	}
}

func TestReservedExpansion_String(t *testing.T) {
	raw := v26.ReservedExpansion{Data: []byte{3, 0xAB, 0xCD}}
	if got, want := raw.String(), "REF{raw 03 AB CD}"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}

	structured := v26.ReservedExpansion{HasTNH: true, TNHDeg: 180.0}
	if got := structured.String(); got == "" || got == "REF{}" {
		t.Errorf("String() for structured REF looks empty: %q", got)
	}
}

func TestReservedExpansion_DecodeInvalidLength(t *testing.T) {
	buf := bytes.NewBuffer([]byte{0x00})
	if _, err := (&v26.ReservedExpansion{}).Decode(buf); err == nil {
		t.Error("Decode() with LEN=0 should error")
	}
}

func TestReservedExpansion_DecodeTruncated(t *testing.T) {
	// LEN=5 claims 4 more bytes but only 2 are available.
	buf := bytes.NewBuffer([]byte{0x05, 0x80, 0x00})
	if _, err := (&v26.ReservedExpansion{}).Decode(buf); err == nil {
		t.Error("Decode() with truncated buffer should error")
	}
}

// SGV (REF §2.7) from the byte patterns of real surface reports.
func TestReservedExpansion_SGVDecode(t *testing.T) {
	tests := []struct {
		name                               string
		sgv                                []byte
		stopped, valid, track, mag, hasHdg bool
		speedKt, hdgDeg                    float64
	}{
		{"taxiing, true heading", []byte{0x41, 0x31, 0xB2}, false, true, false, false, true, 19, 250.3125},
		{"taxiing faster", []byte{0x41, 0x51, 0xB2}, false, true, false, false, true, 21, 250.3125},
		{"stopped, heading invalid", []byte{0xA0, 0x01, 0x00}, true, false, true, false, true, 0, 0},
		{"magnetic ground track", []byte{0x70, 0x11, 0x44}, false, true, true, true, true, 1, 95.625},
		{"no extension", []byte{0x41, 0x30}, false, true, false, false, false, 19, 0},
		{"max speed", []byte{0x0F, 0xFE}, false, false, false, false, false, 255.875, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := append([]byte{byte(2 + len(tt.sgv)), 0x08}, tt.sgv...)
			var r v26.ReservedExpansion
			if _, err := r.Decode(bytes.NewBuffer(raw)); err != nil {
				t.Fatal(err)
			}
			if !r.HasSGV || !bytes.Equal(r.RawSGV, tt.sgv) {
				t.Fatalf("HasSGV=%v Raw=% X, want raw % X", r.HasSGV, r.RawSGV, tt.sgv)
			}
			if r.SGVStopped != tt.stopped || r.SGVValid != tt.valid || r.SGVTrack != tt.track ||
				r.SGVMagnetic != tt.mag || r.SGVHasHeading != tt.hasHdg {
				t.Errorf("flags STP=%v HTS=%v HTT=%v HRD=%v hdg=%v, want %v %v %v %v %v",
					r.SGVStopped, r.SGVValid, r.SGVTrack, r.SGVMagnetic, r.SGVHasHeading,
					tt.stopped, tt.valid, tt.track, tt.mag, tt.hasHdg)
			}
			if r.SGVSpeedKt != tt.speedKt || r.SGVHeadingDeg != tt.hdgDeg {
				t.Errorf("GS %v kt hdg %v°, want %v kt %v°", r.SGVSpeedKt, r.SGVHeadingDeg, tt.speedKt, tt.hdgDeg)
			}

			// Built from the fields alone, the item encodes to the same bytes.
			built := r
			built.RawSGV, built.Data = nil, nil
			var buf bytes.Buffer
			if _, err := built.Encode(&buf); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(buf.Bytes(), raw) {
				t.Errorf("encoded from fields % X, want % X", buf.Bytes(), raw)
			}
		})
	}
}

func TestReservedExpansion_SGVEncodeClamps(t *testing.T) {
	r := v26.ReservedExpansion{HasSGV: true, SGVValid: true, SGVSpeedKt: 300, SGVHasHeading: true, SGVHeadingDeg: 359.9}
	var buf bytes.Buffer
	if _, err := r.Encode(&buf); err != nil {
		t.Fatal(err)
	}
	// GSS clamps to 0x7FF; 359.9° rounds to 360° and wraps to HGT 0.
	want := []byte{0x05, 0x08, 0x4F, 0xFF, 0x00}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Errorf("got % X, want % X", buf.Bytes(), want)
	}
}

func TestReservedExpansion_SGVTruncatedExtension(t *testing.T) {
	// FX announces the HGT octet, but the REF ends.
	var r v26.ReservedExpansion
	if _, err := r.Decode(bytes.NewBuffer([]byte{0x04, 0x08, 0x41, 0x31})); err == nil {
		t.Fatal("want an error for an SGV whose extension is missing")
	}
}
