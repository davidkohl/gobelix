// dataitems/cat021/v26/mode_s_mb_data_test.go
package v26_test

import (
	"bytes"
	"testing"

	v26 "github.com/davidkohl/gobelix/cat/cat021/dataitems/v26"
)

func TestModeSMBData_RoundTrip(t *testing.T) {
	in := v26.ModeSMBData{Records: []v26.ModeSMBDataRecord{
		{MBData: [8]byte{1, 2, 3, 4, 5, 6, 7}, BDS1: 4, BDS2: 0},
		{MBData: [8]byte{0xA0, 0, 0, 0, 0, 0, 0xFF}, BDS1: 6, BDS2: 0},
	}}
	var buf bytes.Buffer
	n, err := in.Encode(&buf)
	if err != nil || n != 1+2*8 {
		t.Fatalf("Encode: n=%d err=%v, want n=17", n, err)
	}
	var out v26.ModeSMBData
	if _, err := out.Decode(&buf); err != nil {
		t.Fatal(err)
	}
	if len(out.Records) != 2 || out.Records[0] != in.Records[0] || out.Records[1] != in.Records[1] {
		t.Errorf("round trip: got %+v, want %+v", out.Records, in.Records)
	}
}

func TestModeSMBData_TooManyRecords(t *testing.T) {
	// 256 records must be refused, not wrapped to a repetition factor of 0.
	m := v26.ModeSMBData{Records: make([]v26.ModeSMBDataRecord, 256)}
	var buf bytes.Buffer
	if _, err := m.Encode(&buf); err == nil {
		t.Fatal("want an error for 256 records")
	}
	if buf.Len() != 0 {
		t.Errorf("wrote %d bytes before failing", buf.Len())
	}

	m.Records = m.Records[:255]
	if n, err := m.Encode(&buf); err != nil || n != 1+255*8 || buf.Bytes()[0] != 255 {
		t.Errorf("255 records: n=%d err=%v rep=%d", n, err, buf.Bytes()[0])
	}
}
