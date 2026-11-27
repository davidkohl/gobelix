// asterix/record_uap_value_test.go
package asterix

import (
	"bytes"
	"testing"
)

// valueUAP is a UAP used by value whose dynamic type is not comparable
// (it holds a slice). Comparing two interface values holding it panics.
type valueUAP struct {
	*MockUAP
	tags []string
}

func TestRecordEncodeNonComparableUAP(t *testing.T) {
	_, mock, err := setupTestRecord()
	if err != nil {
		t.Fatal(err)
	}
	uap := valueUAP{MockUAP: mock, tags: []string{"x"}}
	rec, err := NewRecord(Cat021, uap)
	if err != nil {
		t.Fatal(err)
	}
	_ = rec.SetDataItem("I021/010", &MockDataItem{data: []byte{1, 2}, fixedLen: 2})
	_ = rec.SetDataItem("I021/040", &MockDataItem{data: []byte{3}, fixedLen: 1})
	var buf bytes.Buffer
	if _, err := rec.Encode(&buf); err != nil {
		t.Fatal(err)
	}
	if want := []byte{0xC0, 1, 2, 3}; !bytes.Equal(buf.Bytes(), want) {
		t.Errorf("got % X, want % X", buf.Bytes(), want)
	}
}

// AsterixMessage.String sorts a copy of the fields, not the UAP's slice.
func TestAsterixMessageStringKeepsUAPFields(t *testing.T) {
	mock := &MockUAP{category: Cat021, fields: []DataField{
		{FRN: 2, DataItem: "I021/040", Type: Fixed, Length: 1},
		{FRN: 1, DataItem: "I021/010", Type: Fixed, Length: 2},
	}}
	m := NewAsterixMessage(Cat021)
	m.uap = mock
	m.AddRecord(map[string]DataItem{"I021/010": &MockDataItem{data: []byte{1, 2}}})
	_ = m.String()
	if mock.fields[0].FRN != 2 {
		t.Error("String reordered the UAP's field slice")
	}
}
