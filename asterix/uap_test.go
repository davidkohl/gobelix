// asterix/uap_test.go
package asterix

import (
	"bytes"
	"fmt"
	"testing"
)

// MockUAP implements UAP for testing
type MockUAP struct {
	category       Category
	version        string
	fields         []DataField
	createItemFunc func(id string) (DataItem, error)
}

func (m *MockUAP) Category() Category {
	return m.category
}

func (m *MockUAP) Version() string {
	return m.version
}

func (m *MockUAP) Fields() []DataField {
	return m.fields
}

func (m *MockUAP) CreateDataItem(id string) (DataItem, error) {
	if m.createItemFunc != nil {
		return m.createItemFunc(id)
	}

	// Default implementation
	switch id {
	case "I021/010":
		return &MockDataItem{id: id, data: []byte{0x01, 0x02}, fixedLen: 2}, nil
	case "I021/040":
		return &MockDataItem{id: id, data: []byte{0xFF}, fixedLen: 1}, nil
	case "I021/030":
		return &MockDataItem{id: id, data: []byte{0x03, 0x04, 0x05}, fixedLen: 3}, nil
	case "UnknownItem":
		return nil, fmt.Errorf("unknown item: %s", id)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownDataItem, id)
	}
}

func (m *MockUAP) Validate(items map[string]DataItem) error {
	// Check mandatory fields
	for _, field := range m.fields {
		if field.Mandatory {
			if _, exists := items[field.DataItem]; !exists {
				return fmt.Errorf("%w: %s", ErrMandatoryField, field.DataItem)
			}
		}
	}
	return nil
}

// MockDataItem implements DataItem for testing
type MockDataItem struct {
	id          string
	data        []byte
	fixedLen    int
	decodeErr   error
	encodeErr   error
	validateErr error
}

func (m *MockDataItem) Encode(buf *bytes.Buffer) (int, error) {
	if m.encodeErr != nil {
		return 0, m.encodeErr
	}
	return buf.Write(m.data)
}

func (m *MockDataItem) Decode(buf *bytes.Buffer) (int, error) {
	if m.decodeErr != nil {
		return 0, m.decodeErr
	}
	if buf.Len() < m.fixedLen {
		return 0, fmt.Errorf("buffer too short: need %d bytes, have %d", m.fixedLen, buf.Len())
	}
	m.data = make([]byte, m.fixedLen)
	return buf.Read(m.data)
}

func (m *MockDataItem) Validate() error {
	return m.validateErr
}

// MockMessageTypeItem implements DataItem for testing message types
type MockMessageTypeItem struct {
	typeValue uint8
}

func (m *MockMessageTypeItem) Encode(buf *bytes.Buffer) (int, error) {
	return buf.Write([]byte{m.typeValue})
}

func (m *MockMessageTypeItem) Decode(buf *bytes.Buffer) (int, error) {
	b, err := buf.ReadByte()
	if err != nil {
		return 0, err
	}
	m.typeValue = b
	return 1, nil
}

func (m *MockMessageTypeItem) Validate() error {
	return nil
}

func TestNewBaseUAP(t *testing.T) {
	// Valid case
	fields := []DataField{
		{FRN: 1, DataItem: "I021/010", Type: Fixed, Length: 2, Mandatory: true},
		{FRN: 2, DataItem: "I021/040", Type: Fixed, Length: 1, Mandatory: true},
		{FRN: 3, DataItem: "I021/030", Type: Fixed, Length: 3, Mandatory: false},
	}

	uap, err := NewBaseUAP(Cat021, "1.0", fields)
	if err != nil {
		t.Errorf("NewBaseUAP failed: %v", err)
	}
	if uap == nil {
		t.Fatal("NewBaseUAP returned nil")
	}

	// Invalid category
	_, err = NewBaseUAP(Category(0), "1.0", fields)
	if err == nil {
		t.Error("NewBaseUAP with invalid category should fail")
	}

	// No fields
	_, err = NewBaseUAP(Cat021, "1.0", nil)
	if err == nil {
		t.Error("NewBaseUAP with nil fields should fail")
	}

	// Duplicate FRN
	duplicateFields := []DataField{
		{FRN: 1, DataItem: "I021/010", Type: Fixed, Length: 2, Mandatory: true},
		{FRN: 1, DataItem: "I021/040", Type: Fixed, Length: 1, Mandatory: true}, // Duplicate FRN
	}
	_, err = NewBaseUAP(Cat021, "1.0", duplicateFields)
	if err == nil {
		t.Error("NewBaseUAP with duplicate FRN should fail")
	}

	// Invalid FRN
	invalidFields := []DataField{
		{FRN: 0, DataItem: "I021/010", Type: Fixed, Length: 2, Mandatory: true}, // Invalid FRN
	}
	_, err = NewBaseUAP(Cat021, "1.0", invalidFields)
	if err == nil {
		t.Error("NewBaseUAP with FRN=0 should fail")
	}
}

func TestBaseUAPGetters(t *testing.T) {
	fields := []DataField{
		{FRN: 1, DataItem: "I021/010", Type: Fixed, Length: 2, Mandatory: true},
		{FRN: 2, DataItem: "I021/040", Type: Fixed, Length: 1, Mandatory: true},
		{FRN: 3, DataItem: "I021/030", Type: Fixed, Length: 3, Mandatory: false},
	}

	uap, err := NewBaseUAP(Cat021, "1.0", fields)
	if err != nil {
		t.Fatalf("NewBaseUAP failed: %v", err)
	}

	// Test Category()
	if uap.Category() != Cat021 {
		t.Errorf("Category() = %v, want %v", uap.Category(), Cat021)
	}

	// Test Version()
	if uap.Version() != "1.0" {
		t.Errorf("Version() = %s, want 1.0", uap.Version())
	}

	// Test Fields() - should return a copy
	gotFields := uap.Fields()
	if len(gotFields) != len(fields) {
		t.Errorf("Fields() returned %d fields, want %d", len(gotFields), len(fields))
	}

	// Modify the returned fields - should not affect the original
	gotFields[0].Mandatory = false
	if !uap.Fields()[0].Mandatory {
		t.Error("Fields() should return a copy, not the original")
	}

	// Test GetFieldByDataItem()
	field, found := uap.GetFieldByDataItem("I021/010")
	if !found {
		t.Error("GetFieldByDataItem() should find I021/010")
	}
	if field.FRN != 1 {
		t.Errorf("GetFieldByDataItem() returned FRN=%d, want 1", field.FRN)
	}

	// Test GetFieldByDataItem() with unknown item
	_, found = uap.GetFieldByDataItem("I021/999")
	if found {
		t.Error("GetFieldByDataItem() should not find unknown item")
	}

	// Test GetFieldByFRN()
	field, found = uap.GetFieldByFRN(2)
	if !found {
		t.Error("GetFieldByFRN() should find FRN=2")
	}
	if field.DataItem != "I021/040" {
		t.Errorf("GetFieldByFRN() returned DataItem=%s, want I021/040", field.DataItem)
	}

	// Test GetFieldByFRN() with unknown FRN
	_, found = uap.GetFieldByFRN(99)
	if found {
		t.Error("GetFieldByFRN() should not find unknown FRN")
	}

	// Test MaxFRN()
	if uap.MaxFRN() != 3 {
		t.Errorf("MaxFRN() = %d, want 3", uap.MaxFRN())
	}
}

func TestBaseUAPIsMandatory(t *testing.T) {
	fields := []DataField{
		{FRN: 1, DataItem: "I021/010", Type: Fixed, Length: 2, Mandatory: true},
		{FRN: 2, DataItem: "I021/040", Type: Fixed, Length: 1, Mandatory: true},
		{FRN: 3, DataItem: "I021/030", Type: Fixed, Length: 3, Mandatory: false},
	}

	uap, err := NewBaseUAP(Cat021, "1.0", fields)
	if err != nil {
		t.Fatalf("NewBaseUAP failed: %v", err)
	}

	// Test IsMandatory() with mandatory item
	if !uap.IsMandatory("I021/010") {
		t.Error("IsMandatory() should return true for I021/010")
	}

	// Test IsMandatory() with non-mandatory item
	if uap.IsMandatory("I021/030") {
		t.Error("IsMandatory() should return false for I021/030")
	}

	// Test IsMandatory() with unknown item
	if uap.IsMandatory("I021/999") {
		t.Error("IsMandatory() should return false for unknown item")
	}
}

func TestBaseUAPValidate(t *testing.T) {
	fields := []DataField{
		{FRN: 1, DataItem: "I021/010", Type: Fixed, Length: 2, Mandatory: true},
		{FRN: 2, DataItem: "I021/040", Type: Fixed, Length: 1, Mandatory: true},
		{FRN: 3, DataItem: "I021/030", Type: Fixed, Length: 3, Mandatory: false},
	}

	uap, err := NewBaseUAP(Cat021, "1.0", fields)
	if err != nil {
		t.Fatalf("NewBaseUAP failed: %v", err)
	}

	// Valid case - all mandatory items present
	items := map[string]DataItem{
		"I021/010": &MockDataItem{id: "I021/010"},
		"I021/040": &MockDataItem{id: "I021/040"},
	}
	err = uap.Validate(items)
	if err != nil {
		t.Errorf("Validate() failed: %v", err)
	}

	// Missing mandatory item
	items = map[string]DataItem{
		"I021/010": &MockDataItem{id: "I021/010"},
		// I021/040 missing
	}
	err = uap.Validate(items)
	if err == nil {
		t.Error("Validate() should fail with missing mandatory item")
	}
	if !IsMandatoryFieldMissing(err) {
		t.Errorf("Error should be of mandatory field type, got: %v", err)
	}
}

func TestBaseUAPCreateDataItem(t *testing.T) {
	fields := []DataField{
		{FRN: 1, DataItem: "I021/010", Type: Fixed, Length: 2, Mandatory: true},
	}

	uap, err := NewBaseUAP(Cat021, "1.0", fields)
	if err != nil {
		t.Fatalf("NewBaseUAP failed: %v", err)
	}

	// CreateDataItem should return error in BaseUAP
	_, err = uap.CreateDataItem("I021/010")
	if err == nil {
		t.Error("CreateDataItem() should return error in BaseUAP")
	}
}

func TestTypedUAP(t *testing.T) {
	// Create base UAP
	fields := []DataField{
		{FRN: 1, DataItem: "I021/000", Type: Fixed, Length: 1, Mandatory: true},  // Message type
		{FRN: 2, DataItem: "I021/010", Type: Fixed, Length: 2, Mandatory: true},  // Data source
		{FRN: 3, DataItem: "I021/020", Type: Fixed, Length: 1, Mandatory: false}, // Optional for type 1
		{FRN: 4, DataItem: "I021/030", Type: Fixed, Length: 3, Mandatory: false}, // Mandatory for type 2
	}

	baseUAP, err := NewBaseUAP(Cat021, "1.0", fields)
	if err != nil {
		t.Fatalf("NewBaseUAP failed: %v", err)
	}

	// Create typed UAP
	typedUAP := NewTypedUAP(baseUAP, "I021/000")

	// Register message types
	// Type 1: I021/020 is optional
	typedUAP.RegisterMessageType(1, "Type 1", nil)

	// Type 2: I021/030 is mandatory
	type2Validator := func(messageType uint8, items map[string]DataItem) error {
		if _, exists := items["I021/030"]; !exists {
			return fmt.Errorf("%w: I021/030 is mandatory for message type 2", ErrMandatoryField)
		}
		return nil
	}
	typedUAP.RegisterMessageType(2, "Type 2", type2Validator)

	// Override extractMessageType
	typedUAP.extractMessageTypeFunc = func(item DataItem) (uint8, error) {
		if typeItem, ok := item.(*MockMessageTypeItem); ok {
			return typeItem.typeValue, nil
		}
		return 0, fmt.Errorf("invalid message type item")
	}

	// Test MessageTypes()
	types := typedUAP.MessageTypes()
	if len(types) != 2 {
		t.Errorf("MessageTypes() returned %d types, want 2", len(types))
	}
	if types[1] != "Type 1" {
		t.Errorf("MessageTypes()[1] = %s, want Type 1", types[1])
	}
	if types[2] != "Type 2" {
		t.Errorf("MessageTypes()[2] = %s, want Type 2", types[2])
	}

	// Test validation for type 1
	items := map[string]DataItem{
		"I021/000": &MockMessageTypeItem{typeValue: 1},
		"I021/010": &MockDataItem{id: "I021/010"},
		// I021/020 is optional for type 1
	}
	err = typedUAP.Validate(items)
	if err != nil {
		t.Errorf("Validate() for type 1 failed: %v", err)
	}

	// Test validation for type 2 without mandatory field
	items = map[string]DataItem{
		"I021/000": &MockMessageTypeItem{typeValue: 2},
		"I021/010": &MockDataItem{id: "I021/010"},
		// I021/030 is mandatory for type 2 but missing
	}
	err = typedUAP.Validate(items)
	if err == nil {
		t.Error("Validate() for type 2 without I021/030 should fail")
	}

	// Test validation for type 2 with mandatory field
	items = map[string]DataItem{
		"I021/000": &MockMessageTypeItem{typeValue: 2},
		"I021/010": &MockDataItem{id: "I021/010"},
		"I021/030": &MockDataItem{id: "I021/030"},
	}
	err = typedUAP.Validate(items)
	if err != nil {
		t.Errorf("Validate() for type 2 with I021/030 failed: %v", err)
	}

	// Test validation with unknown message type
	items = map[string]DataItem{
		"I021/000": &MockMessageTypeItem{typeValue: 99}, // Unknown type
		"I021/010": &MockDataItem{id: "I021/010"},
	}
	err = typedUAP.Validate(items)
	if err == nil {
		t.Error("Validate() with unknown message type should fail")
	}

	// Test validation with invalid message type item
	items = map[string]DataItem{
		"I021/000": &MockDataItem{id: "I021/000"}, // Not a MockMessageTypeItem
		"I021/010": &MockDataItem{id: "I021/010"},
	}
	err = typedUAP.Validate(items)
	if err == nil {
		t.Error("Validate() with invalid message type item should fail")
	}
}

func BenchmarkBaseUAPValidate(b *testing.B) {
	fields := []DataField{
		{FRN: 1, DataItem: "I021/010", Type: Fixed, Length: 2, Mandatory: true},
		{FRN: 2, DataItem: "I021/040", Type: Fixed, Length: 1, Mandatory: true},
		{FRN: 3, DataItem: "I021/030", Type: Fixed, Length: 3, Mandatory: false},
		{FRN: 4, DataItem: "I021/050", Type: Fixed, Length: 2, Mandatory: false},
		{FRN: 5, DataItem: "I021/060", Type: Fixed, Length: 4, Mandatory: false},
		{FRN: 6, DataItem: "I021/070", Type: Fixed, Length: 2, Mandatory: false},
		{FRN: 7, DataItem: "I021/080", Type: Fixed, Length: 3, Mandatory: false},
	}

	uap, _ := NewBaseUAP(Cat021, "1.0", fields)

	// Create some items
	items := map[string]DataItem{
		"I021/010": &MockDataItem{id: "I021/010"},
		"I021/040": &MockDataItem{id: "I021/040"},
		"I021/030": &MockDataItem{id: "I021/030"},
		"I021/050": &MockDataItem{id: "I021/050"},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		uap.Validate(items)
	}
}

func BenchmarkTypedUAPValidate(b *testing.B) {
	// Create base UAP
	fields := []DataField{
		{FRN: 1, DataItem: "I021/000", Type: Fixed, Length: 1, Mandatory: true},  // Message type
		{FRN: 2, DataItem: "I021/010", Type: Fixed, Length: 2, Mandatory: true},  // Data source
		{FRN: 3, DataItem: "I021/020", Type: Fixed, Length: 1, Mandatory: false}, // Optional for type 1
		{FRN: 4, DataItem: "I021/030", Type: Fixed, Length: 3, Mandatory: false}, // Mandatory for type 2
	}

	baseUAP, _ := NewBaseUAP(Cat021, "1.0", fields)

	// Create typed UAP
	typedUAP := NewTypedUAP(baseUAP, "I021/000")
	typedUAP.RegisterMessageType(1, "Type 1", nil)
	typedUAP.extractMessageTypeFunc = func(item DataItem) (uint8, error) {
		if typeItem, ok := item.(*MockMessageTypeItem); ok {
			return typeItem.typeValue, nil
		}
		return 0, fmt.Errorf("invalid message type item")
	}

	// Create items
	items := map[string]DataItem{
		"I021/000": &MockMessageTypeItem{typeValue: 1},
		"I021/010": &MockDataItem{id: "I021/010"},
		"I021/020": &MockDataItem{id: "I021/020"},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		typedUAP.Validate(items)
	}
}
